package version

import (
	"bytes"
	"strings"
	"testing"
	"xray-exporter/internal/pkg/app"

	"github.com/sirupsen/logrus"
)

func TestNewVersionCommand(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	original := logrus.StandardLogger().Out
	logrus.SetOutput(&buf)
	t.Cleanup(func() { logrus.SetOutput(original) })

	build := app.BuildInfo{Version: "1.2.3", Commit: "abc", Date: "2026-01-01"}
	cmd := NewVersionCommand(build)

	if cmd.Use != "version" {
		t.Fatalf("Use = %q, want version", cmd.Use)
	}
	if len(cmd.Aliases) != 1 || cmd.Aliases[0] != "v" {
		t.Fatalf("Aliases = %v, want [v]", cmd.Aliases)
	}

	cmd.Run(cmd, nil)

	got := buf.String()
	if !strings.Contains(got, "XRay Exporter 1.2.3-abc (built 2026-01-01)") {
		t.Fatalf("unexpected log output: %q", got)
	}
}
