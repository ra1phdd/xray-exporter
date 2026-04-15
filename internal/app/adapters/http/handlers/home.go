package handlers

import (
	"fmt"
	stdhttp "net/http"

	"github.com/sirupsen/logrus"
)

func Home(buildVersion string, metricsPath string) stdhttp.HandlerFunc {
	return func(w stdhttp.ResponseWriter, _ *stdhttp.Request) {
		body := fmt.Sprintf(`<html>
<head><title>XRay Exporter</title></head>
<body>
<h1>XRay Exporter %s</h1>
<p><a href='/metrics'>Exporter Metrics</a></p>
<p><a href='%s'>Scrape XRay Metrics</a></p>
</body>
</html>
`, buildVersion, metricsPath)
		if _, err := w.Write([]byte(body)); err != nil {
			logrus.Debugf("Write() err: %s", err)
		}
	}
}
