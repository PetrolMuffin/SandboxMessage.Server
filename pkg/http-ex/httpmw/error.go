package httpmw

import (
	"net/http"
	"runtime/debug"

	"github.com/petrolmuffin/sandboxmessage-server/pkg/log"
)

func Error(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				ctx := r.Context()
				logger := log.CtxLogger(ctx)
				logger.Error(ctx, "panic", "error", err, "request_url", r.URL.String(), "stack", string(debug.Stack()))
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
		}()

		next.ServeHTTP(w, r)
	})
}
