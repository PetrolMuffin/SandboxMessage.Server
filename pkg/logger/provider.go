package logger

import (
	"context"
	"log/slog"
)

type loggerKeyType struct{}

var loggerKey = loggerKeyType{}

func GetLogger(ctx context.Context) Logger {
	logger, ok := ctx.Value(loggerKey).(Logger)
	if !ok {
		logger = NewLog(slog.Default())
		ctx = withLogger(ctx, logger)
	}

	return logger
}

func RegisterLogger(ctx context.Context, log *slog.Logger) context.Context {
	return withLogger(ctx, NewLog(log))
}

func withLogger(ctx context.Context, logger Logger) context.Context {
	return context.WithValue(ctx, loggerKey, logger)
}
