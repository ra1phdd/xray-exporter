package app

import (
	stdhttp "net/http"
	"time"
	internalhttp "xray-exporter/internal/app/adapters/http"
	xrayexporter "xray-exporter/internal/app/adapters/xray/exporter"

	"github.com/sirupsen/logrus"
)

type BuildInfo struct {
	Version string
	Commit  string
	Date    string
}

type Config struct {
	Listen                 string
	ScrapePath             string
	XRayEndpoint           string
	ScrapeTimeoutInSeconds int64
	BasicAuthUsername      string
	BasicAuthPassword      string
}

func Run(build BuildInfo, cfg Config) error {
	metricsExporter, err := xrayexporter.New(cfg.XRayEndpoint, cfg.scrapeTimeout())
	if err != nil {
		return err
	}
	defer metricsExporter.Close()

	handler := internalhttp.NewRouter(cfg.toRouterConfig(build), metricsExporter)

	authEnabled := cfg.authEnabled()
	logrus.Info("Server is ready to handle incoming scrape requests.")
	if authEnabled {
		logrus.Info("HTTP Basic Auth is enabled")
	}

	return stdhttp.ListenAndServe(cfg.Listen, handler)
}

func (o Config) authEnabled() bool {
	return o.BasicAuthUsername != "" || o.BasicAuthPassword != ""
}

func (o Config) toRouterConfig(build BuildInfo) internalhttp.RouterConfig {
	return internalhttp.RouterConfig{
		ScrapePath:        o.ScrapePath,
		BuildVersion:      build.Version,
		BasicAuthUsername: o.BasicAuthUsername,
		BasicAuthPassword: o.BasicAuthPassword,
	}
}

func (o Config) scrapeTimeout() time.Duration {
	return time.Duration(o.ScrapeTimeoutInSeconds) * time.Second
}
