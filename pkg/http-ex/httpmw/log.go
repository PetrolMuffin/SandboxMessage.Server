package httpmw

import (
	"net/http"

	"github.com/petrolmuffin/sandboxmessage-server/pkg/log"
)

type responseWriter struct {
	http.ResponseWriter
	status int
}

func (w *responseWriter) WriteHeader(status int) {
	w.status = status
	w.ResponseWriter.WriteHeader(status)
}

func Log(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		logger := log.CtxLogger(ctx)
		logger.Debug(ctx, "request", "method", r.Method, "url", r.URL.Path)
		writer := &responseWriter{ResponseWriter: w}
		next.ServeHTTP(writer, r)
		logger.Debug(ctx, "response", "status", writer.status)
	})
}
