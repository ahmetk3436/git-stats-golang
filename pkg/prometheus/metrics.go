package prometheus

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	APICallsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "gits_api_calls_total",
			Help: "Total number of API calls.",
		},
		[]string{"api_type", "endpoint", "status"}, // api_type (e.g., github, gitlab), endpoint, status (e.g., success, failure)
	)

	RepositoryFetchesTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "gits_repository_fetches_total",
			Help: "Total number of repository fetch attempts.",
		},
		[]string{"provider", "status"}, // provider (e.g., github, gitlab), status (e.g., success, failure)
	)

	APICallDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "gits_api_call_duration_seconds",
			Help:    "Duration of API calls.",
			Buckets: prometheus.LinearBuckets(0.1, 0.1, 10), // 0.1s, 0.2s, ..., 1s
		},
		[]string{"api_type", "endpoint"},
	)
)

// InitMetrics can be called to ensure metrics are registered.
// With promauto, registration is automatic on declaration, so this function
// primarily serves as an explicit initialization point if needed.
func InitMetrics() {
	// This function can be expanded if more complex setup is needed.
	// For now, its existence ensures the package is imported and promauto executes.
}
