package broker

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	"github.com/petrolmuffin/sandboxmessage-server/pkg/log"
)

const (
	dlqOriginalSubjectHeader  = "X-Original-Subject"
	dlqErrorReasonHeader      = "X-Error-Reason"
	dlqDeliveryAttemptsHeader = "X-Delivery-Attempts"
)

func (b *NatsBroker) Listen(ctx context.Context, consumers ...anyConsumer) error {
	for _, h := range consumers {
		logger := log.CtxLogger(ctx)
		subject := h.Subject()

		stream := h.Stream()
		if err := b.registerSubject(subject, stream); err != nil {
			logger.Error(ctx, "subject already registered", "subject", subject, "stream", stream, "error", err)
			return err
		}

		backOff := b.config.BackOff(subject)
		maxDeliver := len(backOff) + 1
		cons, err := b.js.CreateOrUpdateConsumer(ctx, stream, jetstream.ConsumerConfig{
			Durable:       b.config.stream.workerName + "." + stream,
			FilterSubject: subject,
			AckPolicy:     jetstream.AckExplicitPolicy,
			AckWait:       b.config.Timeout(subject),
			MaxDeliver:    maxDeliver,
			BackOff:       backOff,
		})
		if err != nil {
			logger.Error(ctx, "failed to create consumer", "subject", subject, "stream", stream, "error", err)
			return ErrConsumerRunFailed
		}

		handler := h
		consumeContext, err := cons.Consume(func(msg jetstream.Msg) {
			hCtx, hCtxCancel := context.WithTimeout(ctx, b.config.Timeout(subject))
			defer hCtxCancel()

			headers := msg.Headers()
			msgId := headers.Get(nats.MsgIdHdr)
			err := handler.handle(hCtx, msg.Data(), msgId)
			if err == nil {
				msg.Ack()
				return
			}

			meta, metaErr := msg.Metadata()
			var numDelivered uint64
			if metaErr == nil {
				numDelivered = meta.NumDelivered
			} else {
				logger.Warn(hCtx, "failed to get message metadata", "error", metaErr)
				numDelivered = uint64(maxDeliver)
			}

			if errors.Is(err, ErrInvalidRequest) || int(numDelivered) >= maxDeliver {
				logger.Error(hCtx, "moving message to DLQ", "subject", subject, "msgId", msgId, "attempts", numDelivered, "error", err)

				if dlqErr := b.sentToDlq(hCtx, subject, msg, err); dlqErr != nil {
					logger.Error(hCtx, "failed to send to DLQ", "error", dlqErr)
				}

				msg.Term()
				return
			}

			logger.Warn(hCtx, "handler failed, nacking", "subject", subject, "msgId", msgId, "error", err)
			msg.Nak()
		})

		if err != nil {
			return err
		}

		b.waitGroup.Add(1)
		logger.Info(ctx, "listening events", "subject", subject, "stream", stream)

		go func() {
			defer b.waitGroup.Done()
			<-ctx.Done()
			consumeContext.Drain()
			logger.Info(context.Background(), "consumer drained", "subject", subject, "stream", stream)
		}()
	}

	return nil
}

func (b *NatsBroker) registerSubject(subject string, stream string) error {
	key := stream + ":" + subject

	if !b.subjects.Add(key) {
		return fmt.Errorf("subject already registered: %s:%s", stream, subject)
	}

	return nil
}

func (b *NatsBroker) sentToDlq(ctx context.Context, originalSubject string, msg jetstream.Msg, processingErr error) error {
	dlqCtx := context.WithValue(ctx, "originalSubject", originalSubject)
	dlqCtx = context.WithValue(dlqCtx, "processingErr", processingErr)

	headers := msg.Headers()
	if headers == nil {
		headers = nats.Header{
			nats.MsgIdHdr: {uuid.NewString()},
		}
	}

	attempts := "unknown"
	if metadata, err := msg.Metadata(); err == nil {
		attempts = fmt.Sprintf("%d", metadata.NumDelivered)
	}

	headers.Set(dlqOriginalSubjectHeader, originalSubject)
	headers.Set(dlqErrorReasonHeader, processingErr.Error())
	headers.Set(dlqDeliveryAttemptsHeader, attempts)

	msgId := headers.Get(nats.MsgIdHdr)

	err := b.emit(ctx, msgId, headers, originalSubject+dlqSuffix, msg.Data())
	if err != nil {
		return fmt.Errorf("publish to dlq failed: %w", err)
	}

	return nil
}
