package sql

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/ao9911/go-matrix/util/xtime"
)

var client *DB

func init() {
	cfg := &Config{
		DSN:          "root:root@tcp(127.0.0.1:3306)/mysql?parseTime=true",
		ReadDSN:      []string{"root:root@tcp(127.0.0.1:3306)/mysql?parseTime=true"},
		Active:       4,
		Idle:         2,
		IdleTimeout:  xtime.Duration(time.Minute),
		QueryTimeout: xtime.Duration(3 * time.Second),
		ExecTimeout:  xtime.Duration(3 * time.Second),
		TranTimeout:  xtime.Duration(3 * time.Second),
	}
	client = NewMySQL(cfg)
}

func TestMain(m *testing.M) {
	code := m.Run()
	_ = client.Close()
	os.Exit(code)
}

func TestDatabase(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if err := client.PingContext(ctx); err != nil {
		t.Fatal(err)
	}
}
