package http

import (
	stdhttp "net/http"
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

func newTestGatherer() prometheus.Gatherer {
	reg := prometheus.NewRegistry()
	g := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "router_test_metric",
		Help: "router test metric",
	})
	g.Set(1)
	reg.MustRegister(g)
	return reg
}

func TestNewRouterWithoutAuth(t *testing.T) {
	t.Parallel()

	router := NewRouter(
		RouterConfig{
			ScrapePath:   "/scrape",
			BuildVersion: "1.0.0",
		},
		testExporter{gatherer: newTestGatherer()},
	)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(stdhttp.MethodGet, "/scrape", nil)
	router.ServeHTTP(rec, req)
	if rec.Code != stdhttp.StatusOK {
		t.Fatalf("/scrape status = %d, want %d", rec.Code, stdhttp.StatusOK)
	}
	if !strings.Contains(rec.Body.String(), "router_test_metric 1") {
		t.Fatalf("/scrape body missing metric: %q", rec.Body.String())
	}

	rec = httptest.NewRecorder()
	req = httptest.NewRequest(stdhttp.MethodGet, "/", nil)
	router.ServeHTTP(rec, req)
	if rec.Code != stdhttp.StatusOK {
		t.Fatalf("/ status = %d, want %d", rec.Code, stdhttp.StatusOK)
	}
	if !strings.Contains(rec.Body.String(), "XRay Exporter 1.0.0") {
		t.Fatalf("/ body missing version: %q", rec.Body.String())
	}
}

func TestNewRouterWithAuth(t *testing.T) {
	t.Parallel()

	router := NewRouter(
		RouterConfig{
			ScrapePath:        "/scrape",
			BuildVersion:      "1.0.0",
			BasicAuthUsername: "user",
			BasicAuthPassword: "pass",
		},
		testExporter{gatherer: newTestGatherer()},
	)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(stdhttp.MethodGet, "/scrape", nil)
	router.ServeHTTP(rec, req)
	if rec.Code != stdhttp.StatusUnauthorized {
		t.Fatalf("unauthorized status = %d, want %d", rec.Code, stdhttp.StatusUnauthorized)
	}

	rec = httptest.NewRecorder()
	req = httptest.NewRequest(stdhttp.MethodGet, "/scrape", nil)
	req.SetBasicAuth("user", "pass")
	router.ServeHTTP(rec, req)
	if rec.Code != stdhttp.StatusOK {
		t.Fatalf("authorized status = %d, want %d", rec.Code, stdhttp.StatusOK)
	}
}
