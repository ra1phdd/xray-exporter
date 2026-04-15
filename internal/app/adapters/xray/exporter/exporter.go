package exporter

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
	"github.com/xtls/xray-core/app/stats/command"
	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/credentials/insecure"
)

type Exporter struct {
	sync.Mutex
	endpoint           string
	scrapeTimeout      time.Duration
	registry           *prometheus.Registry
	totalScrapes       prometheus.Counter
	metricDescriptions map[string]*prometheus.Desc
	conn               *grpc.ClientConn
}

func New(endpoint string, scrapeTimeout time.Duration) (*Exporter, error) {
	e := Exporter{
		endpoint:      endpoint,
		scrapeTimeout: scrapeTimeout,
		registry:      prometheus.NewRegistry(),
		totalScrapes: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: "xray",
			Name:      "scrapes_total",
			Help:      "Total number of scrapes performed",
		}),
	}

	e.metricDescriptions = map[string]*prometheus.Desc{}
	for k, desc := range map[string]struct {
		txt  string
		lbls []string
	}{
		"up":                           {txt: "Indicate scrape succeeded or not"},
		"scrape_duration_seconds":      {txt: "Scrape duration in seconds"},
		"uptime_seconds":               {txt: "XRay uptime in seconds"},
		"traffic_uplink_bytes_total":   {txt: "Number of transmitted bytes", lbls: []string{"dimension", "target"}},
		"traffic_downlink_bytes_total": {txt: "Number of received bytes", lbls: []string{"dimension", "target"}},
	} {
		e.metricDescriptions[k] = e.newMetricDescr(k, desc.txt, desc.lbls)
	}

	e.registry.MustRegister(&e)

	ctx, cancel := context.WithTimeout(context.Background(), scrapeTimeout)
	defer cancel()

	conn, err := grpc.NewClient(
		endpoint,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to dial: %w, timeout: %v", err, e.scrapeTimeout)
	}

	conn.Connect()
	if err := waitForReady(ctx, conn); err != nil {
		_ = conn.Close()
		return nil, fmt.Errorf("failed to dial: %w, timeout: %v", err, e.scrapeTimeout)
	}

	e.conn = conn
	return &e, nil
}

func waitForReady(ctx context.Context, conn *grpc.ClientConn) error {
	for {
		state := conn.GetState()
		if state == connectivity.Ready {
			return nil
		}
		if state == connectivity.Shutdown {
			return errors.New("gRPC connection is shut down")
		}
		if !conn.WaitForStateChange(ctx, state) {
			return ctx.Err()
		}
	}
}

func (e *Exporter) Gatherer() prometheus.Gatherer {
	return e.registry
}

func (e *Exporter) Close() error {
	if e.conn == nil {
		return nil
	}
	return e.conn.Close()
}

func (e *Exporter) Collect(ch chan<- prometheus.Metric) {
	e.Lock()
	defer e.Unlock()
	e.totalScrapes.Inc()

	start := time.Now().UnixNano()
	up := float64(1)
	if err := e.scrapeXRay(ch); err != nil {
		up = 0
		logrus.Warnf("Scrape failed: %s", err)
	}

	e.registerConstMetricGauge(ch, "up", up)
	e.registerConstMetricGauge(ch, "scrape_duration_seconds", float64(time.Now().UnixNano()-start)/1000000000)
	ch <- e.totalScrapes
}

func (e *Exporter) Describe(ch chan<- *prometheus.Desc) {
	for _, desc := range e.metricDescriptions {
		ch <- desc
	}
	ch <- e.totalScrapes.Desc()
}

func (e *Exporter) scrapeXRay(ch chan<- prometheus.Metric) error {
	client := command.NewStatsServiceClient(e.conn)

	if err := e.scrapeXRaySysMetrics(context.Background(), ch, client); err != nil {
		return err
	}

	if err := e.scrapeXRayMetrics(context.Background(), ch, client); err != nil {
		return err
	}

	return nil
}

func (e *Exporter) scrapeXRayMetrics(ctx context.Context, ch chan<- prometheus.Metric, client command.StatsServiceClient) error {
	resp, err := client.QueryStats(ctx, &command.QueryStatsRequest{Reset_: false})
	if err != nil {
		return fmt.Errorf("failed to get stats: %w", err)
	}

	for _, s := range resp.GetStat() {
		p := strings.Split(s.GetName(), ">>>")
		if len(p) < 4 {
			logrus.Debugf("unexpected stat name format: %s", s.GetName())
			continue
		}

		metric := p[2] + "_" + p[3] + "_bytes_total"
		dimension := p[0]
		target := p[1]
		e.registerConstMetricCounter(ch, metric, float64(s.GetValue()), dimension, target)
	}

	return nil
}

func (e *Exporter) scrapeXRaySysMetrics(ctx context.Context, ch chan<- prometheus.Metric, client command.StatsServiceClient) error {
	resp, err := client.GetSysStats(ctx, &command.SysStatsRequest{})
	if err != nil {
		return fmt.Errorf("failed to get sys stats: %w", err)
	}

	e.registerConstMetricGauge(ch, "uptime_seconds", float64(resp.GetUptime()))
	e.registerConstMetricGauge(ch, "goroutines", float64(resp.GetNumGoroutine()))
	e.registerConstMetricGauge(ch, "memstats_alloc_bytes", float64(resp.GetAlloc()))
	e.registerConstMetricGauge(ch, "memstats_alloc_bytes_total", float64(resp.GetTotalAlloc()))
	e.registerConstMetricGauge(ch, "memstats_sys_bytes", float64(resp.GetSys()))
	e.registerConstMetricGauge(ch, "memstats_mallocs_total", float64(resp.GetMallocs()))
	e.registerConstMetricGauge(ch, "memstats_frees_total", float64(resp.GetFrees()))
	e.registerConstMetricGauge(ch, "memstats_num_gc", float64(resp.GetNumGC()))
	e.registerConstMetricGauge(ch, "memstats_pause_total_ns", float64(resp.GetPauseTotalNs()))

	return nil
}

func (e *Exporter) registerConstMetricGauge(ch chan<- prometheus.Metric, metric string, val float64, labels ...string) {
	e.registerConstMetric(ch, metric, val, prometheus.GaugeValue, labels...)
}

func (e *Exporter) registerConstMetricCounter(ch chan<- prometheus.Metric, metric string, val float64, labels ...string) {
	e.registerConstMetric(ch, metric, val, prometheus.CounterValue, labels...)
}

func (e *Exporter) registerConstMetric(
	ch chan<- prometheus.Metric,
	metric string,
	val float64,
	valType prometheus.ValueType,
	labelValues ...string,
) {
	descr := e.metricDescriptions[metric]
	if descr == nil {
		descr = e.newMetricDescr(metric, metric+" metric", nil)
	}

	if m, err := prometheus.NewConstMetric(descr, valType, val, labelValues...); err == nil {
		ch <- m
	} else {
		logrus.Debugf("NewConstMetric() err: %s", err)
	}
}

func (e *Exporter) newMetricDescr(metricName string, docString string, labels []string) *prometheus.Desc {
	return prometheus.NewDesc(prometheus.BuildFQName("xray", "", metricName), docString, labels, nil)
}
