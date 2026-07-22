package postgre

import (
	"context"
	"database/sql"
	"os"
	"testing"
	"time"

	"github.com/ao9911/go-matrix/util/xtime"
)

var client *sql.DB

func init() {
	cfg := &Config{
		DSN:             "postgres://postgres:postgres@127.0.0.1:5432/postgres?sslmode=disable",
		MaxOpenConns:    4,
		MaxIdleConns:    2,
		ConnMaxLifetime: xtime.Duration(time.Minute),
		ConnMaxIdleTime: xtime.Duration(30 * time.Second),
		ConnTimeout:     xtime.Duration(3 * time.Second),
	}
	client = NewPostgre(cfg)
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
	var got int
	if err := client.QueryRowContext(ctx, "SELECT $1::integer", 1).Scan(&got); err != nil {
		t.Fatal(err)
	}
	if got != 1 {
		t.Fatalf("SELECT result = %d, want 1", got)
	}
}
