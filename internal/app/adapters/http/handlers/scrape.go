package handlers

import (
	stdhttp "net/http"
	"xray-exporter/internal/app/ports"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func Scrape(exporter ports.MetricsExporter) stdhttp.HandlerFunc {
	return func(w stdhttp.ResponseWriter, r *stdhttp.Request) {
		promhttp.HandlerFor(
			exporter.Gatherer(),
			promhttp.HandlerOpts{ErrorHandling: promhttp.ContinueOnError},
		).ServeHTTP(w, r)
	}
}
