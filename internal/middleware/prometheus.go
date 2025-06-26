package middleware

import (
	"net/http"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// HTTP request metrics
	httpRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "path", "status"},
	)

	httpRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP request duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path", "status"},
	)

	httpRequestSize = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_size_bytes",
			Help:    "HTTP request size in bytes",
			Buckets: prometheus.ExponentialBuckets(100, 10, 8),
		},
		[]string{"method", "path"},
	)

	httpResponseSize = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_response_size_bytes",
			Help:    "HTTP response size in bytes",
			Buckets: prometheus.ExponentialBuckets(100, 10, 8),
		},
		[]string{"method", "path", "status"},
	)

	// Application-specific metrics
	aircraftSearchesTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "aircraft_searches_total",
			Help: "Total number of aircraft searches performed",
		},
		[]string{"query_type"}, // "search" or "browse"
	)

	aircraftSearchDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "aircraft_search_duration_seconds",
			Help:    "Aircraft search duration in seconds",
			Buckets: []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1.0, 2.5, 5.0},
		},
		[]string{"query_type"},
	)

	aircraftDetailsViews = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "aircraft_details_views_total",
			Help: "Total number of aircraft detail views",
		},
	)

	databaseQueriesTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "database_queries_total",
			Help: "Total number of database queries",
		},
		[]string{"query_type", "status"},
	)

	databaseQueryDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "database_query_duration_seconds",
			Help:    "Database query duration in seconds",
			Buckets: []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1.0},
		},
		[]string{"query_type"},
	)

	// System metrics
	activeConnections = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "active_connections",
			Help: "Number of active HTTP connections",
		},
	)

	totalAircraftCount = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "total_aircraft_count",
			Help: "Total number of aircraft in the database",
		},
	)
)

// PrometheusMiddleware returns an Echo middleware that tracks HTTP metrics
func PrometheusMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			start := time.Now()
			
			// Track active connections
			activeConnections.Inc()
			defer activeConnections.Dec()

			// Get request size
			reqSize := computeApproximateRequestSize(c.Request())

			// Process request
			err := next(c)

			// Calculate metrics
			status := c.Response().Status
			if err != nil {
				// Handle errors that might not set status
				if he, ok := err.(*echo.HTTPError); ok {
					status = he.Code
				} else {
					status = 500
				}
			}

			duration := time.Since(start).Seconds()
			method := c.Request().Method
			path := c.Path() // Use route pattern, not actual path
			statusStr := strconv.Itoa(status)

			// Record metrics
			httpRequestsTotal.WithLabelValues(method, path, statusStr).Inc()
			httpRequestDuration.WithLabelValues(method, path, statusStr).Observe(duration)
			httpRequestSize.WithLabelValues(method, path).Observe(float64(reqSize))
			httpResponseSize.WithLabelValues(method, path, statusStr).Observe(float64(c.Response().Size))

			return err
		}
	}
}

// Helper functions for application metrics
func RecordAircraftSearch(queryType string, duration time.Duration) {
	aircraftSearchesTotal.WithLabelValues(queryType).Inc()
	aircraftSearchDuration.WithLabelValues(queryType).Observe(duration.Seconds())
}

func RecordAircraftDetailView() {
	aircraftDetailsViews.Inc()
}

func RecordDatabaseQuery(queryType string, duration time.Duration, success bool) {
	status := "success"
	if !success {
		status = "error"
	}
	databaseQueriesTotal.WithLabelValues(queryType, status).Inc()
	databaseQueryDuration.WithLabelValues(queryType).Observe(duration.Seconds())
}

func UpdateTotalAircraftCount(count float64) {
	totalAircraftCount.Set(count)
}

// computeApproximateRequestSize computes the approximate size of the request
func computeApproximateRequestSize(r *http.Request) int {
	s := 0
	if r.URL != nil {
		s = len(r.URL.Path)
	}

	s += len(r.Method)
	s += len(r.Proto)
	for name, values := range r.Header {
		s += len(name)
		for _, value := range values {
			s += len(value)
		}
	}
	s += len(r.Host)

	// N.B. r.Form and r.MultipartForm are assumed to be included in r.URL.

	if r.ContentLength != -1 {
		s += int(r.ContentLength)
	}
	return s
}