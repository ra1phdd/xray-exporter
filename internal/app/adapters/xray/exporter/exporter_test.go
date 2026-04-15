package exporter

import (
	"context"
	"errors"
	"net"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
	"github.com/xtls/xray-core/app/stats/command"
	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/credentials/insecure"
)

type fakeStatsClient struct {
	queryResp *command.QueryStatsResponse
	queryErr  error
	sysResp   *command.SysStatsResponse
	sysErr    error
}

func (f fakeStatsClient) GetStats(context.Context, *command.GetStatsRequest, ...grpc.CallOption) (*command.GetStatsResponse, error) {
	return nil, errors.New("not implemented")
}
func (f fakeStatsClient) GetStatsOnline(context.Context, *command.GetStatsRequest, ...grpc.CallOption) (*command.GetStatsResponse, error) {
	return nil, errors.New("not implemented")
}
func (f fakeStatsClient) QueryStats(context.Context, *command.QueryStatsRequest, ...grpc.CallOption) (*command.QueryStatsResponse, error) {
	if f.queryErr != nil {
		return nil, f.queryErr
	}
	if f.queryResp == nil {
		return &command.QueryStatsResponse{}, nil
	}
	return f.queryResp, nil
}
func (f fakeStatsClient) GetSysStats(context.Context, *command.SysStatsRequest, ...grpc.CallOption) (*command.SysStatsResponse, error) {
	if f.sysErr != nil {
		return nil, f.sysErr
	}
	if f.sysResp == nil {
		return &command.SysStatsResponse{}, nil
	}
	return f.sysResp, nil
}
func (f fakeStatsClient) GetStatsOnlineIpList(context.Context, *command.GetStatsRequest, ...grpc.CallOption) (*command.GetStatsOnlineIpListResponse, error) {
	return nil, errors.New("not implemented")
}
func (f fakeStatsClient) GetAllOnlineUsers(context.Context, *command.GetAllOnlineUsersRequest, ...grpc.CallOption) (*command.GetAllOnlineUsersResponse, error) {
	return nil, errors.New("not implemented")
}

type testStatsServer struct {
	command.UnimplementedStatsServiceServer
	queryResp *command.QueryStatsResponse
	queryErr  error
	sysResp   *command.SysStatsResponse
	sysErr    error
}

func (s testStatsServer) QueryStats(context.Context, *command.QueryStatsRequest) (*command.QueryStatsResponse, error) {
	if s.queryErr != nil {
		return nil, s.queryErr
	}
	if s.queryResp == nil {
		return &command.QueryStatsResponse{}, nil
	}
	return s.queryResp, nil
}

func (s testStatsServer) GetSysStats(context.Context, *command.SysStatsRequest) (*command.SysStatsResponse, error) {
	if s.sysErr != nil {
		return nil, s.sysErr
	}
	if s.sysResp == nil {
		return &command.SysStatsResponse{}, nil
	}
	return s.sysResp, nil
}

func newTestExporter() *Exporter {
	e := &Exporter{
		registry: prometheus.NewRegistry(),
		totalScrapes: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: "xray",
			Name:      "scrapes_total",
			Help:      "Total number of scrapes performed",
		}),
		metricDescriptions: map[string]*prometheus.Desc{},
	}
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
	return e
}

func metricFamilyByName(mfs []*dto.MetricFamily, name string) *dto.MetricFamily {
	for _, mf := range mfs {
		if mf.GetName() == name {
			return mf
		}
	}
	return nil
}

func startStatsServer(t *testing.T, impl command.StatsServiceServer) (addr string, stop func()) {
	t.Helper()

	lis, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("net.Listen() err = %v", err)
	}

	srv := grpc.NewServer()
	command.RegisterStatsServiceServer(srv, impl)
	go func() {
		_ = srv.Serve(lis)
	}()

	return lis.Addr().String(), func() {
		srv.Stop()
		_ = lis.Close()
	}
}

func TestWaitForReadyShutdown(t *testing.T) {
	t.Parallel()

	conn, err := grpc.NewClient("127.0.0.1:1", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("grpc.NewClient() err = %v", err)
	}
	_ = conn.Close()

	err = waitForReady(context.Background(), conn)
	if err == nil {
		t.Fatal("waitForReady() expected shutdown error")
	}
}

func TestWaitForReadyContextDeadline(t *testing.T) {
	t.Parallel()

	conn, err := grpc.NewClient("127.0.0.1:1", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("grpc.NewClient() err = %v", err)
	}
	defer conn.Close()
	conn.Connect()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	err = waitForReady(ctx, conn)
	if err == nil {
		t.Fatal("waitForReady() expected context deadline error")
	}
}

func TestNewAndClose(t *testing.T) {
	t.Parallel()

	addr, stop := startStatsServer(t, testStatsServer{})
	defer stop()

	e, err := New(addr, time.Second)
	if err != nil {
		t.Fatalf("New() err = %v", err)
	}
	if e.conn == nil {
		t.Fatal("New() returned exporter with nil conn")
	}
	if got := e.conn.GetState(); got != connectivity.Ready && got != connectivity.Idle && got != connectivity.Connecting {
		t.Fatalf("unexpected gRPC connection state: %v", got)
	}
	if err := e.Close(); err != nil {
		t.Fatalf("Close() err = %v", err)
	}
}

func TestCloseNilConnection(t *testing.T) {
	t.Parallel()

	if err := (&Exporter{}).Close(); err != nil {
		t.Fatalf("Close() err = %v", err)
	}
}

func TestNewDialError(t *testing.T) {
	t.Parallel()

	_, err := New("bad::endpoint", 50*time.Millisecond)
	if err == nil {
		t.Fatal("New() expected dial error")
	}
}

func TestScrapeXRayMetrics(t *testing.T) {
	t.Parallel()

	e := newTestExporter()
	ch := make(chan prometheus.Metric, 8)

	client := fakeStatsClient{
		queryResp: &command.QueryStatsResponse{
			Stat: []*command.Stat{
				{Name: "user>>>alice>>>traffic>>>uplink", Value: 10},
				{Name: "user>>>alice>>>traffic>>>downlink", Value: 20},
				{Name: "bad-format", Value: 99},
			},
		},
	}

	if err := e.scrapeXRayMetrics(context.Background(), ch, client); err != nil {
		t.Fatalf("scrapeXRayMetrics() err = %v", err)
	}
	close(ch)

	count := 0
	for range ch {
		count++
	}
	if count != 2 {
		t.Fatalf("metrics count = %d, want 2", count)
	}
}

func TestScrapeXRayMetricsError(t *testing.T) {
	t.Parallel()

	e := newTestExporter()
	err := e.scrapeXRayMetrics(context.Background(), make(chan prometheus.Metric, 1), fakeStatsClient{queryErr: errors.New("boom")})
	if err == nil {
		t.Fatal("scrapeXRayMetrics() expected error")
	}
}

func TestScrapeXRaySysMetrics(t *testing.T) {
	t.Parallel()

	e := newTestExporter()
	ch := make(chan prometheus.Metric, 16)
	client := fakeStatsClient{
		sysResp: &command.SysStatsResponse{
			Uptime:       1,
			NumGoroutine: 2,
			Alloc:        3,
			TotalAlloc:   4,
			Sys:          5,
			Mallocs:      6,
			Frees:        7,
			NumGC:        8,
			PauseTotalNs: 9,
		},
	}

	if err := e.scrapeXRaySysMetrics(context.Background(), ch, client); err != nil {
		t.Fatalf("scrapeXRaySysMetrics() err = %v", err)
	}
	close(ch)

	count := 0
	for range ch {
		count++
	}
	if count != 9 {
		t.Fatalf("metrics count = %d, want 9", count)
	}
}

func TestScrapeXRaySysMetricsError(t *testing.T) {
	t.Parallel()

	e := newTestExporter()
	err := e.scrapeXRaySysMetrics(context.Background(), make(chan prometheus.Metric, 1), fakeStatsClient{sysErr: errors.New("boom")})
	if err == nil {
		t.Fatal("scrapeXRaySysMetrics() expected error")
	}
}

func TestRegisterConstMetricInvalidLabels(t *testing.T) {
	t.Parallel()

	e := newTestExporter()
	ch := make(chan prometheus.Metric, 1)
	e.registerConstMetricCounter(ch, "traffic_uplink_bytes_total", 1)
	close(ch)

	for range ch {
		t.Fatal("expected no metrics because label count is invalid")
	}
}

func TestRegisterConstMetricCreatesDynamicDescription(t *testing.T) {
	t.Parallel()

	e := newTestExporter()
	ch := make(chan prometheus.Metric, 1)
	e.registerConstMetricGauge(ch, "dynamic_metric", 12)
	close(ch)

	m, ok := <-ch
	if !ok {
		t.Fatal("expected metric in channel")
	}
	var dtoMetric dto.Metric
	if err := m.Write(&dtoMetric); err != nil {
		t.Fatalf("metric.Write() err = %v", err)
	}
	if got := dtoMetric.GetGauge().GetValue(); got != 12 {
		t.Fatalf("gauge value = %v, want 12", got)
	}
}

func TestCollectSuccessAndFailure(t *testing.T) {
	t.Parallel()

	successAddr, successStop := startStatsServer(t, testStatsServer{
		sysResp: &command.SysStatsResponse{
			Uptime:       11,
			NumGoroutine: 12,
			Alloc:        13,
			TotalAlloc:   14,
			Sys:          15,
			Mallocs:      16,
			Frees:        17,
			NumGC:        18,
			PauseTotalNs: 19,
		},
		queryResp: &command.QueryStatsResponse{
			Stat: []*command.Stat{
				{Name: "user>>>bob>>>traffic>>>uplink", Value: 21},
			},
		},
	})
	defer successStop()

	e, err := New(successAddr, time.Second)
	if err != nil {
		t.Fatalf("New() err = %v", err)
	}
	defer e.Close()

	mfs, err := e.Gatherer().Gather()
	if err != nil {
		t.Fatalf("Gather() err = %v", err)
	}
	if mf := metricFamilyByName(mfs, "xray_up"); mf == nil || len(mf.GetMetric()) == 0 || mf.GetMetric()[0].GetGauge().GetValue() != 1 {
		t.Fatal("xray_up metric is missing or has unexpected value")
	}
	if mf := metricFamilyByName(mfs, "xray_scrapes_total"); mf == nil || len(mf.GetMetric()) == 0 || mf.GetMetric()[0].GetCounter().GetValue() != 1 {
		t.Fatal("xray_scrapes_total metric is missing or has unexpected value")
	}

	failAddr, failStop := startStatsServer(t, testStatsServer{sysErr: errors.New("sys boom")})
	defer failStop()

	failing, err := New(failAddr, time.Second)
	if err != nil {
		t.Fatalf("New() err = %v", err)
	}
	defer failing.Close()

	mfs, err = failing.Gatherer().Gather()
	if err != nil {
		t.Fatalf("Gather() err = %v", err)
	}
	if mf := metricFamilyByName(mfs, "xray_up"); mf == nil || len(mf.GetMetric()) == 0 || mf.GetMetric()[0].GetGauge().GetValue() != 0 {
		t.Fatal("xray_up metric for failing scrape is missing or has unexpected value")
	}
}

func TestCollectFailureOnQueryStats(t *testing.T) {
	t.Parallel()

	addr, stop := startStatsServer(t, testStatsServer{
		sysResp:  &command.SysStatsResponse{},
		queryErr: errors.New("query boom"),
	})
	defer stop()

	e, err := New(addr, time.Second)
	if err != nil {
		t.Fatalf("New() err = %v", err)
	}
	defer e.Close()

	mfs, err := e.Gatherer().Gather()
	if err != nil {
		t.Fatalf("Gather() err = %v", err)
	}
	if mf := metricFamilyByName(mfs, "xray_up"); mf == nil || len(mf.GetMetric()) == 0 || mf.GetMetric()[0].GetGauge().GetValue() != 0 {
		t.Fatal("xray_up metric for query failure is missing or has unexpected value")
	}
}
