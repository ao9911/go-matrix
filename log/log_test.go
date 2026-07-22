package log

import (
	"context"
	"testing"
)

func init() {
	cfg := &Config{
		// LogPath: "./logs/",
		AppName: "test",
		Debug:   false,
	}
	Init(cfg)
}

// go test -v -test.run TestDebug
func TestDebug(t *testing.T) {
	Debug("hello debug")
	Debugf("hello number=%d", 100)
}

// go test -v -test.run TestInfo
func TestInfo(t *testing.T) {
	Info("hello")
	Infof("hello number=%d", 100)
}

func TestWarn(t *testing.T) {
	Warn("hello")
	Warnf("hello  number=%d", 100)
}

// go test -v -test.run TestError
func TestError(t *testing.T) {
	Error("hello")
	Errorf("hello number=%d", 100)
}

// go test -v -test.run TestCtxInfof
func TestCtxInfof(t *testing.T) {
	ctx := WithTraceID(context.Background(), "abc-123")
	CtxInfo(ctx, "hello world")
}
