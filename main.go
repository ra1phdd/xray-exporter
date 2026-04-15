package main

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/jessevdk/go-flags"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
)

var opts struct {
	Listen                 string `short:"l" long:"listen" description:"Listen address" value-name:"[ADDR]:PORT" default:":9550"`
	MetricsPath            string `short:"m" long:"metrics-path" description:"Metrics path" value-name:"PATH" default:"/scrape"`
	V2RayEndpoint          string `short:"e" long:"v2ray-endpoint" description:"V2Ray API endpoint" value-name:"HOST:PORT" default:"127.0.0.1:8080"`
	ScrapeTimeoutInSeconds int64  `short:"t" long:"scrape-timeout" description:"The timeout in seconds for every individual scrape" value-name:"N" default:"3"`
	BasicAuthUsername      string `short:"u" long:"basic-auth-username" description:"Username for HTTP Basic Auth protection"`
	BasicAuthPassword      string `short:"p" long:"basic-auth-password" description:"Password for HTTP Basic Auth protection"`
	Version                bool   `short:"v" long:"version" description:"Display the version and exit"`
}

var (
	buildVersion = "dev"
	buildCommit  = "none"
	buildDate    = "unknown"
)

var exporter *Exporter

func scrapeHandler(w http.ResponseWriter, r *http.Request) {
	promhttp.HandlerFor(
		exporter.registry, promhttp.HandlerOpts{ErrorHandling: promhttp.ContinueOnError},
	).ServeHTTP(w, r)
}

func basicAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		username, password, ok := r.BasicAuth()
		if !ok || username != opts.BasicAuthUsername || password != opts.BasicAuthPassword {
			w.Header().Set("WWW-Authenticate", `Basic realm="v2ray-exporter"`)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func main() {
	var err error
	if _, err = flags.Parse(&opts); err != nil {
		fmt.Println(err)
		os.Exit(0)
	}

	fmt.Printf("V2Ray Exporter %v-%v (built %v)\n", buildVersion, buildCommit, buildDate)
	if opts.Version {
		os.Exit(0)
	}

	scrapeTimeout := time.Duration(opts.ScrapeTimeoutInSeconds) * time.Second
	exporter, err = NewExporter(opts.V2RayEndpoint, scrapeTimeout)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())
	mux.HandleFunc(opts.MetricsPath, scrapeHandler)
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		_, err := w.Write([]byte(`<html>
<head><title>V2Ray Exporter</title></head>
<body>
<h1>V2Ray Exporter ` + buildVersion + `</h1>
<p><a href='/metrics'>Exporter Metrics</a></p>
<p><a href='` + opts.MetricsPath + `'>Scrape V2Ray Metrics</a></p>
</body>
</html>
`))
		if err != nil {
			logrus.Debugf("Write() err: %s", err)
		}
	})

	var handler http.Handler = mux
	authEnabled := opts.BasicAuthUsername != "" || opts.BasicAuthPassword != ""
	if authEnabled {
		handler = basicAuthMiddleware(handler)
	}

	logrus.Infof("Server is ready to handle incoming scrape requests.")
	if authEnabled {
		logrus.Info("HTTP Basic Auth is enabled")
	}
	logrus.Fatal(http.ListenAndServe(opts.Listen, handler))

	defer exporter.conn.Close()
}
