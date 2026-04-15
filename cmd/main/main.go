package main

import (
	"fmt"
	"os"
	versioncmd "xray-exporter/cmd/version"
	"xray-exporter/internal/pkg/app"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	buildVersion = "dev"
	buildCommit  = "none"
	buildDate    = "unknown"
)

func main() {
	build := app.BuildInfo{
		Version: buildVersion,
		Commit:  buildCommit,
		Date:    buildDate,
	}

	if err := newRootCommand(build).Execute(); err != nil {
		logrus.Fatal(err)
	}
}

func newRootCommand(build app.BuildInfo) *cobra.Command {
	cfg := app.Config{
		Listen:                 ":9550",
		ScrapePath:             "/scrape",
		XRayEndpoint:           "127.0.0.1:8080",
		ScrapeTimeoutInSeconds: 3,
	}

	cmd := &cobra.Command{
		Use:   "xray-exporter",
		Short: "Export XRay metrics for Prometheus",
		RunE: func(_ *cobra.Command, _ []string) error {
			logrus.Printf("XRay Exporter %v-%v (built %v)\n", build.Version, build.Commit, build.Date)
			return app.Run(build, cfg)
		},
	}

	cmd.Flags().StringVar(&cfg.Listen, "listen", cfg.Listen, "Listen address")
	cmd.Flags().StringVar(&cfg.ScrapePath, "scrapepath", cfg.ScrapePath, "Metrics scrape path")
	cmd.Flags().StringVar(&cfg.XRayEndpoint, "endpoint", cfg.XRayEndpoint, "XRay API endpoint")
	cmd.Flags().Int64Var(&cfg.ScrapeTimeoutInSeconds, "scrapetimeout", cfg.ScrapeTimeoutInSeconds, "Timeout in seconds for each scrape")
	cmd.Flags().String("basicauth", "", "HTTP basic auth in username:password format")

	cmd.AddCommand(versioncmd.NewVersionCommand(build))
	cmd.PreRunE = func(c *cobra.Command, _ []string) error {
		value, err := c.Flags().GetString("basicauth")
		if err != nil {
			return err
		}

		user, pass, err := parseBasicAuth(value)
		if err != nil {
			return err
		}

		cfg.BasicAuthUsername = user
		cfg.BasicAuthPassword = pass
		return nil
	}

	cmd.SetOut(os.Stdout)
	cmd.SetErr(os.Stderr)
	return cmd
}

func parseBasicAuth(raw string) (string, string, error) {
	if raw == "" {
		return "", "", nil
	}

	for i := range raw {
		if raw[i] == ':' {
			return raw[:i], raw[i+1:], nil
		}
	}

	return "", "", fmt.Errorf("invalid --basicauth value %q: expected username:password", raw)
}
