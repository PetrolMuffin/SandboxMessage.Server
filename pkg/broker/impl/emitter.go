package broker

import (
	"context"

	"github.com/google/uuid"
	"github.com/nats-io/nats.go"
	"github.com/petrolmuffin/sandboxmessage-server/pkg/log"
	"google.golang.org/protobuf/proto"
)

type EmitFormat int8

const (
	EmitFormatProto EmitFormat = iota
	EmitFormatJson
)

func (b *NatsBroker) Emit(ctx context.Context, command string, data proto.Message) error {
	logger := log.CtxLogger(ctx)

	msgId := uuid.NewString()
	logger.Info(ctx, "emitting", "subject", command, "msgId", msgId)

	requestData, err := proto.Marshal(data)
	if err != nil {
		logger.Error(ctx, "failed to get request data", "error", err, "msgId", msgId)
		return err
	}

	headers := nats.Header{
		nats.MsgIdHdr: {msgId},
	}

	return b.emit(ctx, msgId, headers, command, requestData)
}

func (b *NatsBroker) emit(ctx context.Context, msgId string, headers nats.Header, subject string, data []byte) error {
	logger := log.CtxLogger(ctx)

	msg := nats.NewMsg(subject)
	msg.Data = data
	msg.Header = headers

	_, err := b.js.PublishMsg(ctx, msg)
	if err != nil {
		logger.Error(ctx, "publish failed", "subject", subject, "error", err, "msgId", msgId)
		return ErrPublishFailed
	}

	logger.Info(ctx, "emitted", "subject", subject, "msgId", msgId)
	return nil
}
