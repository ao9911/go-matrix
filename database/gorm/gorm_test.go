package gorm

import (
	"os"
	"testing"
	"time"

	"github.com/ao9911/go-matrix/util/xtime"
	"gorm.io/gorm"
)

var client *gorm.DB

func init() {
	cfg := &Config{
		DSN:          "root:root@tcp(127.0.0.1:3306)/mysql?parseTime=true",
		ReadDSN:      []string{"root:root@tcp(127.0.0.1:3306)/mysql?parseTime=true"},
		MaxIdleConns: 2,
		MaxOpenConns: 4,
		MaxLifetime:  xtime.Duration(time.Minute),
		MaxIdleTime:  xtime.Duration(30 * time.Second),
	}
	client = NewORM(cfg)
}

func TestMain(m *testing.M) {
	code := m.Run()
	db, err := client.DB()
	if err == nil {
		_ = db.Close()
	}
	os.Exit(code)
}

func TestDatabase(t *testing.T) {
	db, err := client.DB()
	if err != nil {
		t.Fatal(err)
	}
	if err := db.Ping(); err != nil {
		t.Fatal(err)
	}
	var got int
	if err := client.Raw("SELECT 1").Scan(&got).Error; err != nil {
		t.Fatal(err)
	}
	if got != 1 {
		t.Fatalf("SELECT 1 = %d, want 1", got)
	}
}
