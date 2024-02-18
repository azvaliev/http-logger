package http_logger

import (
	"go.uber.org/zap"
	"net/http"
	"time"
)

type ResponseWriterWithLogging struct {
	http.ResponseWriter
	status int
}

// WriteHeader wraps original http.ResponseWriter WriteHeader method to store status code
func (w *ResponseWriterWithLogging) WriteHeader(status int) {
	w.status = status
	w.ResponseWriter.WriteHeader(status)
}

func (w *ResponseWriterWithLogging) Write(b []byte) (int, error) {
	// if status code was not set, assume 200, as per http.ResponseWriter.Write contract
	if w.status == 0 {
		w.status = 200
	}

	return w.ResponseWriter.Write(b)
}

// WithLogging wraps http.Handler with logging middleware
// It logs request method, path, remote address, response status and latency in nanoseconds
// See ExampleWithLogging for usage
func WithLogging(handler http.Handler, logger *zap.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		startedAt := time.Now()
		wl := &ResponseWriterWithLogging{w, 0}
		handler.ServeHTTP(wl, r)

		logger.Info("Request",
			zap.String("method", r.Method),
			zap.String("path", r.URL.Path),
			zap.String("proto", r.Proto),
			zap.String("remoteAddr", r.RemoteAddr),
			zap.Int("status", wl.status),
			zap.Duration("latency", time.Since(startedAt)),
		)
	})
}
