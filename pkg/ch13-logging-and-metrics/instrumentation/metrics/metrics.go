package metrics

// metrics package defines some metrics that can be useful
// when instrumenting a typical web-service

import (
	"flag"

	"github.com/go-kit/kit/metrics"
	"github.com/go-kit/kit/metrics/prometheus"
	prom "github.com/prometheus/client_golang/prometheus"
)

// metrics related variables
var (
	// Namespace: metrics namespace
	Namespace = flag.String("namespace", "web", "metrics namespace")
	// Subsystem: metrics subsystem
	Subsystem = flag.String("subsystem", "server1", "metrics subsystem")

	// create and globally register metrics based on prometheus implementation
	//
	// Requests: counter of requests
	Requests metrics.Counter = prometheus.NewCounterFrom(
		prom.CounterOpts{
			Namespace: *Namespace,
			Subsystem: *Subsystem,
			Name:      "request_count",
			Help:      "Total requests",
		},
		[]string{},
	)

	// WriteErrors: counter of errors
	WriteErrors metrics.Counter = prometheus.NewCounterFrom(
		prom.CounterOpts{
			Namespace: *Namespace,
			Subsystem: *Subsystem,
			Name:      "write_error_count",
			Help:      "Total write errors",
		},
		[]string{},
	)

	// OpenConnections: gauge for currently open connections
	OpenConnections metrics.Gauge = prometheus.NewGaugeFrom(
		prom.GaugeOpts{
			Namespace: *Namespace,
			Subsystem: *Subsystem,
			Name:      "open_connections",
			Help:      "Current open connections",
		},
		[]string{},
	)

	// RequestDuration: histogram for request processing duration
	RequestDuration metrics.Histogram = prometheus.NewHistogramFrom(
		prom.HistogramOpts{
			Namespace: *Namespace,
			Subsystem: *Subsystem,
			Name:      "request_duration_histogram_seconds",
			Help:      "Total duration of all requests",
			Buckets: []float64{
				0.0000001, 0.0000002, 0.0000003, 0.0000004, 0.0000005,
				0.000001, 0.0000025, 0.000005, 0.0000075, 0.00001,
				0.0001, 0.001, 0.01,
			},
		},
		[]string{},
	)
)
