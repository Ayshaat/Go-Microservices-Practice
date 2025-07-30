package server

import (
	"net/http"
	"time"

	"stocks/internal/log"
)

func loggingMiddleware(logger log.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		start := time.Now()

		next.ServeHTTP(rec, r)

		duration := time.Since(start)
		logger.Info("HTTP request",
			log.String("method", r.Method),
			log.String("path", r.URL.Path),
			log.Int("status", rec.status),
			log.String("duration", duration.String()),
		)
	})
}

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (rec *statusRecorder) WriteHeader(code int) {
	rec.status = code
	rec.ResponseWriter.WriteHeader(code)
}
