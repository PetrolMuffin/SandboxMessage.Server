package httpmw

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/google/uuid"
	"github.com/petrolmuffin/sandboxmessage-server/pkg/log"
)

const (
	traceIdKey = "trace_id"
	urlPathKey = "url_path"
	userIdKey  = "user_id"
)

const (
	traceIdHeader = "X-Trace-Id"
	userIdHeader  = "X-User-Id"
)

func Trace(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), urlPathKey, r.URL.Path)
		logger := log.CtxLogger(ctx)

		userId := r.Header.Get(userIdHeader)
		if userId == "" {
			userId = "unknown"
		}
		ctx = context.WithValue(ctx, userIdKey, userId)

		traceId := r.Header.Get(traceIdHeader)
		if traceId == "" {
			if id, err := uuid.NewV7(); err == nil {
				traceId = id.String()
			} else {
				logger.Error(ctx, "failed to generate new trace id", "error", err, "request_url", r.URL.String())
			}
		}
		ctx = context.WithValue(ctx, traceIdKey, traceId)

		clog := slog.Default().With(slog.String(userIdKey, userId), slog.String(urlPathKey, r.URL.Path), slog.String(traceIdKey, traceId))
		ctx = log.WithLogger(ctx, clog)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
