package postgre

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/ao9911/go-matrix/util/xtime"
)

// Config configures a PostgreSQL database/sql connection pool.
type Config struct {
	DSN             string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime xtime.Duration
	ConnMaxIdleTime xtime.Duration
	ConnTimeout     xtime.Duration
}

// NewPostgre creates a PostgreSQL client and panics when it cannot connect.
func NewPostgre(c *Config) *sql.DB {
	db, err := Open(c)
	if err != nil {
		panic(fmt.Errorf("open postgre: %w", err))
	}
	return db
}

// Open creates and verifies a PostgreSQL database/sql connection pool.
func Open(c *Config) (*sql.DB, error) {
	if c == nil {
		return nil, errors.New("postgre: nil config")
	}
	if c.DSN == "" {
		return nil, errors.New("postgre: empty DSN")
	}

	db, err := sql.Open("pgx", c.DSN)
	if err != nil {
		return nil, fmt.Errorf("open pgx driver: %w", err)
	}
	db.SetMaxOpenConns(c.MaxOpenConns)
	db.SetMaxIdleConns(c.MaxIdleConns)
	db.SetConnMaxLifetime(time.Duration(c.ConnMaxLifetime))
	db.SetConnMaxIdleTime(time.Duration(c.ConnMaxIdleTime))

	timeout := time.Duration(c.ConnTimeout)
	if timeout <= 0 {
		timeout = 5 * time.Second
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("ping postgre: %w", err)
	}
	return db, nil
}
