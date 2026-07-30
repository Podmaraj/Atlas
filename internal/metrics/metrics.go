package metrics

import (
	"net/http"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	RequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "edgecore_requests_total",
			Help: "Total number of HTTP requests processed by EdgeCore API Gateway",
		},
		[]string{"service", "route", "method", "status_code"},
	)

	RequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "edgecore_request_duration_seconds",
			Help:    "HTTP request latency distributions in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"service", "route"},
	)

	ActiveConnections = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "edgecore_active_connections",
			Help: "Current number of active client connections",
		},
	)

	UpstreamHealth = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "edgecore_upstream_health_status",
			Help: "Health status of upstream service instances (1 = healthy, 0 = unhealthy)",
		},
		[]string{"service", "instance"},
	)
)

func init() {
	prometheus.MustRegister(RequestsTotal)
	prometheus.MustRegister(RequestDuration)
	prometheus.MustRegister(ActiveConnections)
	prometheus.MustRegister(UpstreamHealth)
}

// RecordRequest metrics helper
func RecordRequest(service, route, method string, statusCode int, duration time.Duration) {
	statusStr := strconv.Itoa(statusCode)
	RequestsTotal.WithLabelValues(service, route, method, statusStr).Inc()
	RequestDuration.WithLabelValues(service, route).Observe(duration.Seconds())
}

// Handler returns standard Prometheus HTTP scraping handler
func Handler() http.Handler {
	return promhttp.Handler()
}
