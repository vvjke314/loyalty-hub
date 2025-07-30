package metrics

import "github.com/prometheus/client_golang/prometheus"

var (
	HttpRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Общее количество HTTP-запросов",
		},
		[]string{"method", "path", "status"},
	)

	HttpRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "Время обработки HTTP-запросов",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path", "status"},
	)
)

func RegisterMetrics() {
	prometheus.MustRegister(HttpRequestsTotal, HttpRequestDuration)
}
