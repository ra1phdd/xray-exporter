package version

import (
	"xray-exporter/internal/pkg/app"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func NewVersionCommand(build app.BuildInfo) *cobra.Command {
	return &cobra.Command{
		Use:     "version",
		Aliases: []string{"v"},
		Short:   "Show version information",
		Run: func(_ *cobra.Command, _ []string) {
			logrus.Printf("XRay Exporter %v-%v (built %v)\n", build.Version, build.Commit, build.Date)
		},
	}
}
