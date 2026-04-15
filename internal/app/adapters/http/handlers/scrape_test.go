package handlers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
)

type testExporter struct {
	gatherer prometheus.Gatherer
}

func (t testExporter) Gatherer() prometheus.Gatherer { return t.gatherer }
func (t testExporter) Close() error                  { return nil }

func TestScrape(t *testing.T) {
	t.Parallel()

	reg := prometheus.NewRegistry()
	g := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "test_metric",
		Help: "test metric",
	})
	g.Set(42)
	reg.MustRegister(g)

	req := httptest.NewRequest(http.MethodGet, "/scrape", nil)
	rec := httptest.NewRecorder()

	Scrape(testExporter{gatherer: reg})(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if !strings.Contains(rec.Body.String(), "test_metric 42") {
		t.Fatalf("scrape output missing metric: %q", rec.Body.String())
	}
}
