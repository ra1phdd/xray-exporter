package http

import (
	stdhttp "net/http"
	handlers2 "xray-exporter/internal/app/adapters/http/handlers"
	"xray-exporter/internal/app/adapters/http/middlewares"
	"xray-exporter/internal/app/ports"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type RouterConfig struct {
	ScrapePath        string
	BuildVersion      string
	BasicAuthUsername string
	BasicAuthPassword string
}

func NewRouter(cfg RouterConfig, exporter ports.MetricsExporter) stdhttp.Handler {
	mux := stdhttp.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())
	mux.HandleFunc(cfg.ScrapePath, handlers2.Scrape(exporter))
	mux.HandleFunc("/", handlers2.Home(cfg.BuildVersion, cfg.ScrapePath))

	var handler stdhttp.Handler = mux
	if cfg.BasicAuthUsername != "" || cfg.BasicAuthPassword != "" {
		handler = middlewares.BasicAuth(cfg.BasicAuthUsername, cfg.BasicAuthPassword)(handler)
	}

	return handler
}
