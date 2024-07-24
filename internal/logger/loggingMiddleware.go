package logger

import (
	"bytes"
	"net/http"
	"runtime/debug"
	"time"
)

// responseWriter is a minimal wrapper for http.ResponseWriter that allows the
// written HTTP status code to be captured for logging.
type responseWriter struct {
	http.ResponseWriter
	status      int
	wroteHeader bool
	body        bytes.Buffer
}

func wrapResponseWriter(w http.ResponseWriter) *responseWriter {
	return &responseWriter{ResponseWriter: w}
}

func (rw *responseWriter) Status() int {
	return rw.status
}

func (rw *responseWriter) WriteHeader(code int) {
	if rw.wroteHeader {
		return
	}

	rw.status = code
	rw.ResponseWriter.WriteHeader(code)
	rw.wroteHeader = true
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	if !rw.wroteHeader {
		rw.WriteHeader(http.StatusOK)
	}

	rw.body.Write(b)
	return rw.ResponseWriter.Write(b)
}

// LoggingMiddleware logs the request and response of each HTTP request.
func LoggingMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					Error("Panic occurred", String("error", string(debug.Stack())))
				}
			}()

			start := time.Now()
			wrapped := wrapResponseWriter(w)
			next.ServeHTTP(wrapped, r)
			Info("Request completed",
				Int("status", wrapped.status),
				// logger.String("body", wrapped.body.String()),
				String("method", r.Method),
				String("path", r.URL.EscapedPath()),
				String("duration", time.Since(start).String()),
			)
		}
		return http.HandlerFunc(fn)
	}
}
