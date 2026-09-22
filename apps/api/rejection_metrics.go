package main

import (
	"github.com/prometheus/client_golang/prometheus"
	"strings"
)

var rejectionCategories = prometheus.NewCounterVec(prometheus.CounterOpts{Name: "clipsense_rejections_total", Help: "API rejections by fixed safe category, never raw error text."}, []string{"category"})

func init() { metricRegistry.MustRegister(rejectionCategories) }
func recordRejection(status int, code string) {
	category := "invalid"
	switch {
	case status == 413:
		category = "too_large"
	case status == 401 || status == 403:
		category = "unauthenticated"
	case status == 429:
		category = "rate_limited"
	case status >= 500:
		category = "unavailable"
	case strings.Contains(code, "archive"):
		category = "archive"
	case strings.Contains(code, "multipart"):
		category = "multipart"
	}
	rejectionCategories.WithLabelValues(category).Inc()
}
