package evateamclient

import (
	"errors"
	"strconv"

	"github.com/prometheus/client_golang/prometheus"
)

type Metrics interface {
	RecordRequestDuration(status int, method, host, function string, duration float64)
}

// PrometheusMetrics holds Prometheus metrics for the eva.team client
type PrometheusMetrics struct {
	RequestDuration prometheus.HistogramVec
}

func NewPrometheusMetrics() *PrometheusMetrics {
	return &PrometheusMetrics{
		RequestDuration: *prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "eva_client_request_duration_seconds",
				Help:    "Duration of eva.team API requests in seconds",
				Buckets: []float64{0.01, 0.05, 0.1, 0.5, 1, 2, 5, 10},
			},
			[]string{"status", "method", "host", "function"},
		),
	}
}

// Register registers metrics with Prometheus registry
func (m *PrometheusMetrics) Register(registerer prometheus.Registerer) error {
	if err := registerer.Register(&m.RequestDuration); err != nil {
		// If already registered, that's fine
		var alreadyRegisteredError prometheus.AlreadyRegisteredError
		if errors.As(err, &alreadyRegisteredError) {
			return nil
		}

		return err
	}

	return nil
}

// Unregister unregisters metrics from Prometheus registry
func (m *PrometheusMetrics) Unregister(registerer prometheus.Registerer) bool {
	return registerer.Unregister(&m.RequestDuration)
}

// RecordRequestDuration writes duration request with labels
func (m *PrometheusMetrics) RecordRequestDuration(status int, method, host, function string, duration float64) {
	m.RequestDuration.WithLabelValues(strconv.Itoa(status), method, host, function).Observe(duration)
}
