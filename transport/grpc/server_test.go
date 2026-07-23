package grpc

import (
	"context"
	"errors"
	"net"
	"testing"
	"time"

	"github.com/ao9911/go-matrix/util/xtime"
	grpcgo "google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"
)

var (
	serverCfg *ServerConfig
	server    *Server
)

func init() {
	serverCfg = &ServerConfig{
		Network:           "tcp",
		Addr:              "127.0.0.1:0",
		Timeout:           xtime.Duration(time.Second),
		IdleTimeout:       xtime.Duration(time.Minute),
		MaxLifeTime:       xtime.Duration(2 * time.Hour),
		ForceCloseWait:    xtime.Duration(20 * time.Second),
		KeepAliveInterval: xtime.Duration(time.Minute),
		KeepAliveTimeout:  xtime.Duration(20 * time.Second),
	}
	server = NewServer(serverCfg)
}

// go test -v -test.run TestNewServer
func TestNewServer(t *testing.T) {
	if server == nil {
		t.Fatal("NewServer() returned nil")
	}
	if server.conf != serverCfg {
		t.Fatal("server did not retain its configuration")
	}
}

// go test -v -test.run TestNewServerDefaults
func TestNewServerDefaults(t *testing.T) {
	cfg := &ServerConfig{}
	s := NewServer(cfg)
	if s == nil {
		t.Fatal("NewServer() returned nil")
	}
	if cfg.Network != "tcp" || cfg.Addr != "0.0.0.0:9000" || cfg.Timeout != xtime.Duration(time.Second) {
		t.Fatalf("server defaults were not applied: %+v", cfg)
	}
	if cfg.IdleTimeout != xtime.Duration(time.Minute) || cfg.MaxLifeTime != xtime.Duration(2*time.Hour) ||
		cfg.ForceCloseWait != xtime.Duration(20*time.Second) ||
		cfg.KeepAliveInterval != xtime.Duration(time.Minute) || cfg.KeepAliveTimeout != xtime.Duration(20*time.Second) {
		t.Fatalf("server duration defaults were not applied: %+v", cfg)
	}
}

// go test -v -test.run TestServerServeAndStop
func TestServerServeAndStop(t *testing.T) {
	lis := bufconn.Listen(1024 * 1024)

	s := NewServer(&ServerConfig{})
	s.buildServer()
	serveErr := make(chan error, 1)
	go func() { serveErr <- s.Serve(lis) }()

	conn, err := grpcgo.NewClient("passthrough:///bufnet",
		grpcgo.WithTransportCredentials(insecure.NewCredentials()),
		grpcgo.WithContextDialer(func(context.Context, string) (net.Conn, error) { return lis.Dial() }),
	)
	if err != nil {
		t.Fatalf("NewRPCClient() error = %v", err)
	}
	defer func() {
		if err := conn.Close(); err != nil {
			t.Errorf("close connection: %v", err)
		}
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	resp, err := healthpb.NewHealthClient(conn).Check(ctx, &healthpb.HealthCheckRequest{})
	if err != nil {
		t.Fatalf("health check: %v", err)
	}
	if resp.Status != healthpb.HealthCheckResponse_SERVING {
		t.Fatalf("health status = %s, want SERVING", resp.Status)
	}

	if err := s.Stop(ctx); err != nil {
		t.Fatalf("Stop() error = %v", err)
	}
	if err := <-serveErr; err != nil && !errors.Is(err, grpcgo.ErrServerStopped) {
		t.Fatalf("Serve() error = %v", err)
	}
}

// go test -v -test.run TestServerStopTimeout
func TestTracingInterceptor(t *testing.T) {
	wantErr := status.Error(codes.InvalidArgument, "bad request")
	interceptor := TracingInterceptor()
	resp, err := interceptor(context.Background(), "request", &grpcgo.UnaryServerInfo{FullMethod: "/test.Service/Call"},
		func(context.Context, any) (any, error) { return nil, wantErr },
	)
	if resp != nil || !errors.Is(err, wantErr) {
		t.Fatalf("TracingInterceptor() = (%v, %v), want (nil, %v)", resp, err, wantErr)
	}
}
