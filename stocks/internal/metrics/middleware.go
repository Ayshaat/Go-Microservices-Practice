package metrics

import (
	"net/http"
	"time"

	"stocks/internal/log"
)

func MetricsMiddleware(m MetricsInterface, logger log.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			rw := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

			next.ServeHTTP(rw, r)

			duration := time.Since(start).Seconds()
			path := r.URL.Path
			method := r.Method

			m.IncRequest(path, method)
			m.ObserveDuration(path, method, duration)

			if rw.statusCode >= 400 {
				logger.Error("HTTP request error",
					log.String("method", method),
					log.String("path", path),
					log.Int("status", rw.statusCode),
					log.Float64("duration_seconds", duration),
				)
				m.IncError(path, method)
			} else {
				logger.Info("HTTP request",
					log.String("method", method),
					log.String("path", path),
					log.Int("status", rw.statusCode),
					log.Float64("duration_seconds", duration),
				)
			}
		})
	}
}
