package grpc

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/ao9911/go-matrix/util/xtime"
	"github.com/mercari/go-circuitbreaker"
	grpcgo "google.golang.org/grpc"
)

var (
	clientCfg *ClientConfig
	client    *grpcgo.ClientConn
)

func init() {
	clientCfg = &ClientConfig{
		Addr:           "127.0.0.1:8091",
		RequestTimeout: xtime.Duration(time.Second),
		Circuitbreaker: Circuitbreaker{
			CounterResetInterval: xtime.Duration(time.Minute),
			Threshold:            3,
			OpenTimeout:          xtime.Duration(20 * time.Second),
			HalfOpenMaxSuccesses: 10,
		},
	}
	client = NewRPCClient(clientCfg)
}

// go test -v -test.run TestNewRPCClient
func TestNewRPCClient(t *testing.T) {
	if client == nil {
		t.Fatal("NewRPCClient() returned a nil connection")
	}
	t.Cleanup(func() {
		if err := client.Close(); err != nil {
			t.Errorf("close client: %v", err)
		}
	})
	if client.Target() != clientCfg.Addr {
		t.Fatalf("Target() = %q, want %q", client.Target(), clientCfg.Addr)
	}
}

// go test -v -test.run TestNewRPCClientDefaults
func TestNewRPCClientDefaults(t *testing.T) {
	cfg := &ClientConfig{Addr: "passthrough:///127.0.0.1:8092"}
	conn := NewRPCClient(cfg)
	t.Cleanup(func() { _ = conn.Close() })

	if cfg.LoadBalancing == "" || cfg.RequestTimeout != xtime.Duration(time.Second) {
		t.Fatalf("client defaults were not applied: %+v", cfg)
	}
	if cfg.Circuitbreaker.CounterResetInterval != xtime.Duration(time.Minute) ||
		cfg.Circuitbreaker.Threshold != 3 ||
		cfg.Circuitbreaker.OpenTimeout != xtime.Duration(20*time.Second) ||
		cfg.Circuitbreaker.HalfOpenMaxSuccesses != 10 {
		t.Fatalf("circuit-breaker defaults were not applied: %+v", cfg.Circuitbreaker)
	}
}

// go test -v -test.run TestUnaryClientInterceptor
func TestUnaryClientInterceptor(t *testing.T) {
	cb := circuitbreaker.New()
	wantErr := errors.New("invoke failed")
	handlerCalled := false
	interceptor := UnaryClientInterceptor(cb, func(context.Context, string, any) {
		handlerCalled = true
	})

	err := interceptor(context.Background(), "/test.Service/Call", nil, nil, nil,
		func(context.Context, string, any, any, *grpcgo.ClientConn, ...grpcgo.CallOption) error {
			return wantErr
		},
	)
	if !errors.Is(err, wantErr) {
		t.Fatalf("interceptor error = %v, want %v", err, wantErr)
	}
	if handlerCalled {
		t.Fatal("open-state handler was called while the circuit breaker was closed")
	}
}
