package handlers

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHome(t *testing.T) {
	t.Parallel()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	Home("1.2.3", "/scrape")(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	body := rec.Body.String()
	if !strings.Contains(body, "XRay Exporter 1.2.3") {
		t.Fatalf("body does not contain build version: %q", body)
	}
	if !strings.Contains(body, "href='/scrape'") {
		t.Fatalf("body does not contain scrape path: %q", body)
	}
}

type failingWriter struct {
	header http.Header
}

func (f *failingWriter) Header() http.Header {
	if f.header == nil {
		f.header = make(http.Header)
	}
	return f.header
}

func (f *failingWriter) Write([]byte) (int, error) {
	return 0, errors.New("write failed")
}

func (f *failingWriter) WriteHeader(int) {}

func TestHomeWriteError(t *testing.T) {
	t.Parallel()

	Home("1.2.3", "/scrape")(&failingWriter{}, httptest.NewRequest(http.MethodGet, "/", nil))
}
