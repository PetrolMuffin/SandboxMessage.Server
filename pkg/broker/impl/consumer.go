package broker

import (
	"context"
	"log/slog"

	"google.golang.org/protobuf/proto"
)

type Consumer[TReq proto.Message] interface {
	Subject() string
	Stream() string
	Handle(ctx context.Context, request TReq) error
	NewRequest() TReq
}

type anyConsumer interface {
	Subject() string
	Stream() string
	handle(ctx context.Context, data []byte, msgId string) error
}

type consumerAdapter[TReq proto.Message] struct {
	c Consumer[TReq]
}

func (i *consumerAdapter[TReq]) Subject() string {
	return i.c.Subject()
}

func (i *consumerAdapter[TReq]) Stream() string {
	return i.c.Stream()
}

func (i *consumerAdapter[TReq]) handle(ctx context.Context, data []byte, msgId string) error {
	slog.Info("request handling", "subject", i.Subject(), "msgId", msgId)
	req := i.c.NewRequest()
	if err := proto.Unmarshal(data, req); err != nil {
		slog.Error("failed to unmarshal request", "subject", i.Subject(), "msgId", msgId, "error", err)
		return ErrInvalidRequest
	}

	err := i.c.Handle(ctx, req)
	if err != nil {
		slog.Error("request handling failed", "subject", i.Subject(), "msgId", msgId, "request", req, "error", err)
		return err
	}

	slog.Info("request handled", "subject", i.Subject(), "msgId", msgId, "request", req)
	return nil
}

func AsAny[T proto.Message](c Consumer[T]) anyConsumer {
	return &consumerAdapter[T]{c: c}
}
