// sql.go
package sql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/go-sql-driver/mysql"
)

var (
	ErrNoMaster = errors.New("sql: no master instance")
)

// DB database.
type DB struct {
	write  *conn
	read   []*conn
	idx    int64
	master *DB
}

// Tx transaction.
type Tx struct {
	*sql.Tx
	cancel context.CancelFunc
}

// conn database connection
type conn struct {
	*sql.DB
	conf *Config
	addr string
}

// Open create a mysql databse .
func Open(c *Config) (*DB, error) {
	if c == nil {
		return nil, errors.New("sql: nil mysql config")
	}
	if c.DSN == "" {
		return nil, errors.New("sql: empty mysql DSN")
	}

	db := new(DB)
	d, err := connect(c, c.DSN)
	if err != nil {
		return nil, err
	}

	addr := parseDSN(c.DSN)
	db.write = &conn{DB: d, conf: c, addr: addr}

	// 初始化读库
	rs := make([]*conn, 0, len(c.ReadDSN))
	for _, rd := range c.ReadDSN {
		d, err := connect(c, rd)
		if err != nil {
			_ = db.Close()
			return nil, err
		}
		rs = append(rs, &conn{DB: d, conf: c, addr: parseDSN(rd)})
	}
	db.read = rs
	db.master = &DB{write: db.write}

	timeout := time.Duration(c.QueryTimeout)
	if timeout <= 0 {
		timeout = 20 * time.Second
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, err
	}

	return db, nil
}

func connect(c *Config, dataSourceName string) (*sql.DB, error) {
	d, err := sql.Open("mysql", dataSourceName)
	if err != nil {
		return nil, fmt.Errorf("open mysql driver: %w", err)
	}
	d.SetMaxOpenConns(c.Active)
	d.SetMaxIdleConns(c.Idle)
	d.SetConnMaxLifetime(time.Duration(c.IdleTimeout))
	return d, nil
}

// parseDSN parse dsn name and return addr.
func parseDSN(dsn string) (addr string) {
	cfg, err := mysql.ParseDSN(dsn)
	if err != nil {
		return ""
	}
	return cfg.Addr
}

func (db *DB) readIndex() int {
	if len(db.read) == 0 {
		return 0
	}
	v := atomic.AddInt64(&db.idx, 1)
	return int(v) % len(db.read)
}

// Close closes the write and read database, releasing any open resources.
func (db *DB) Close() (err error) {
	if db == nil {
		return nil
	}
	var errs []error
	if db.write != nil {
		errs = appendCloseError(errs, db.write.Close())
	}
	for _, rd := range db.read {
		errs = appendCloseError(errs, rd.Close())
	}
	return errors.Join(errs...)
}

func appendCloseError(errs []error, err error) []error {
	if err != nil {
		return append(errs, fmt.Errorf("close mysql connection: %w", err))
	}
	return errs
}

// PingContext verifies the write connection and every configured read connection.
func (db *DB) PingContext(ctx context.Context) error {
	if db == nil || db.write == nil {
		return errors.New("sql: mysql is not initialized")
	}
	if err := db.write.PingContext(ctx); err != nil {
		return fmt.Errorf("ping mysql writer: %w", err)
	}
	for i, rd := range db.read {
		if err := rd.PingContext(ctx); err != nil {
			return fmt.Errorf("ping mysql reader %d: %w", i, err)
		}
	}
	return nil
}

// Master is mysql master instance
func (db *DB) Master() *DB {
	if db.master == nil {
		panic(ErrNoMaster)
	}
	return db.master
}
