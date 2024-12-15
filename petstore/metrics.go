package petstore

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	httpRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"path", "method", "status"},
	)
)

func init() {
	prometheus.MustRegister(httpRequestsTotal)
}

func MetricsHandler() http.Handler {
	return promhttp.Handler()
}

func RecordMetrics(path, method, status string) {
	httpRequestsTotal.WithLabelValues(path, method, status).Inc()
}
