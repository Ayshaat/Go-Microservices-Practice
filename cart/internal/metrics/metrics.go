package metrics

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

type MetricsInterface interface {
	IncRequest(path, method string)
	ObserveDuration(path, method string, duration float64)
	IncError(path, method string)
	MetricsHandler() http.Handler
}

type Metrics struct {
	RequestsTotal   *prometheus.CounterVec
	RequestDuration *prometheus.HistogramVec
	RequestErrors   *prometheus.CounterVec
}

func (m *Metrics) IncRequest(path, method string) {
	m.RequestsTotal.WithLabelValues(path, method).Inc()
}

func (m *Metrics) ObserveDuration(path, method string, duration float64) {
	m.RequestDuration.WithLabelValues(path, method).Observe(duration)
}

func (m *Metrics) IncError(path, method string) {
	m.RequestErrors.WithLabelValues(path, method).Inc()
}

func (m *Metrics) MetricsHandler() http.Handler {
	return promhttp.Handler()
}

func StartMetricsServer(addr string) {
	go func() {
		http.Handle("/metrics", promhttp.Handler())
		log.Printf("Starting Prometheus metrics server at %s/metrics", addr)
		if err := http.ListenAndServe(addr, nil); err != nil {
			log.Fatalf("Failed to start metrics server: %v", err)
		}
	}()
}

func RegisterMetrics() *Metrics {
	m := &Metrics{
		RequestsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "http_requests_total",
				Help: "Total number of HTTP requests",
			},
			[]string{"path", "method"},
		),
		RequestDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "http_response_duration_seconds",
				Help:    "Duration of HTTP requests in seconds",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"path", "method"},
		),
		RequestErrors: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "http_errors_total",
				Help: "Total number of HTTP request errors",
			},
			[]string{"path", "method"},
		),
	}

	prometheus.MustRegister(m.RequestsTotal, m.RequestDuration, m.RequestErrors)

	return m
}

func WithMetrics(metrics *Metrics, path string, handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		handler(w, r)

		duration := time.Since(start).Seconds()
		metrics.RequestsTotal.WithLabelValues(path, r.Method).Inc()
		metrics.RequestDuration.WithLabelValues(path, r.Method).Observe(duration)
	}
}

func ErrorMetrics(metrics *Metrics, path string, handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rw := &responseWriter{ResponseWriter: w, statusCode: 200}

		start := time.Now()
		handler(rw, r)
		duration := time.Since(start).Seconds()

		metrics.RequestsTotal.WithLabelValues(path, r.Method).Inc()
		metrics.RequestDuration.WithLabelValues(path, r.Method).Observe(duration)

		if rw.statusCode >= 400 {
			metrics.RequestErrors.WithLabelValues(path, r.Method).Inc()
		}
	}
}

type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func UnaryServerInterceptor(m *Metrics) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		start := time.Now()

		resp, err := handler(ctx, req)

		duration := time.Since(start).Seconds()
		m.RequestsTotal.WithLabelValues(info.FullMethod, "GRPC").Inc() // method + fixed "GRPC" label for method type
		m.RequestDuration.WithLabelValues(info.FullMethod, "GRPC").Observe(duration)

		if err != nil {
			st, _ := status.FromError(err)
			if st.Code() != 0 {
				m.RequestErrors.WithLabelValues(info.FullMethod, "GRPC").Inc()
			}
		}

		return resp, err
	}
}
