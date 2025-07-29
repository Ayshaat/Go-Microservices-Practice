package server

import (
	"net/http"
	"time"

	"cart/internal/log"
	"cart/internal/log/zap"

	"go.opentelemetry.io/otel/trace"
)

func loggingMiddleware(logger *zap.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}

		start := time.Now()
		next.ServeHTTP(rec, r)
		duration := time.Since(start)

		span := trace.SpanFromContext(r.Context())
		sc := span.SpanContext()
		traceID := ""
		if sc.IsValid() {
			traceID = sc.TraceID().String()
		}

		logger.Info("HTTP request",
			log.String("method", r.Method),
			log.String("path", r.URL.Path),
			log.Int("status", rec.status),
			log.String("duration", duration.String()),
			log.String("trace_id", traceID),
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
