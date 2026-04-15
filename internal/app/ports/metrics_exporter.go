package ports

import "github.com/prometheus/client_golang/prometheus"

type MetricsExporter interface {
	Gatherer() prometheus.Gatherer
	Close() error
}
