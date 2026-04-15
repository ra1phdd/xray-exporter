package app

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/xtls/xray-core/app/stats/command"
	"google.golang.org/grpc"
)

func TestConfigAuthEnabled(t *testing.T) {
	t.Parallel()

	if (Config{}).authEnabled() {
		t.Fatal("authEnabled() = true, want false")
	}
	if !((Config{BasicAuthUsername: "u"}).authEnabled()) {
		t.Fatal("authEnabled() = false with username, want true")
	}
	if !((Config{BasicAuthPassword: "p"}).authEnabled()) {
		t.Fatal("authEnabled() = false with password, want true")
	}
}

func TestConfigToRouterConfig(t *testing.T) {
	t.Parallel()

	cfg := Config{
		ScrapePath:        "/x",
		BasicAuthUsername: "u",
		BasicAuthPassword: "p",
	}
	build := BuildInfo{Version: "1.0.0"}

	got := cfg.toRouterConfig(build)
	if got.ScrapePath != "/x" || got.BuildVersion != "1.0.0" || got.BasicAuthUsername != "u" || got.BasicAuthPassword != "p" {
		t.Fatalf("toRouterConfig() = %+v", got)
	}
}

func TestConfigScrapeTimeout(t *testing.T) {
	t.Parallel()

	cfg := Config{ScrapeTimeoutInSeconds: 7}
	if got := cfg.scrapeTimeout(); got != 7*time.Second {
		t.Fatalf("scrapeTimeout() = %v, want 7s", got)
	}
}

func TestRunReturnsErrorOnInvalidEndpoint(t *testing.T) {
	t.Parallel()

	err := Run(
		BuildInfo{Version: "test"},
		Config{
			Listen:                 "127.0.0.1:0",
			ScrapePath:             "/scrape",
			XRayEndpoint:           "bad::endpoint",
			ScrapeTimeoutInSeconds: 1,
		},
	)
	if err == nil {
		t.Fatal("Run() expected error for invalid endpoint")
	}
}

type testStatsServer struct {
	command.UnimplementedStatsServiceServer
}

func (testStatsServer) QueryStats(context.Context, *command.QueryStatsRequest) (*command.QueryStatsResponse, error) {
	return &command.QueryStatsResponse{}, nil
}

func (testStatsServer) GetSysStats(context.Context, *command.SysStatsRequest) (*command.SysStatsResponse, error) {
	return &command.SysStatsResponse{}, nil
}

func startStatsServer(t *testing.T) (addr string, stop func()) {
	t.Helper()

	lis, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("net.Listen() err = %v", err)
	}
	srv := grpc.NewServer()
	command.RegisterStatsServiceServer(srv, testStatsServer{})
	go func() {
		_ = srv.Serve(lis)
	}()

	return lis.Addr().String(), func() {
		srv.Stop()
		_ = lis.Close()
	}
}

func TestRunReturnsListenErrorAfterSuccessfulSetup(t *testing.T) {
	t.Parallel()

	addr, stop := startStatsServer(t)
	defer stop()

	err := Run(
		BuildInfo{Version: "test"},
		Config{
			Listen:                 "bad-listen-address",
			ScrapePath:             "/scrape",
			XRayEndpoint:           addr,
			ScrapeTimeoutInSeconds: 1,
			BasicAuthUsername:      "u",
		},
	)
	if err == nil {
		t.Fatal("Run() expected listen error")
	}
}
