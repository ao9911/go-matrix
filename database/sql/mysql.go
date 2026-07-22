// mysql.go
package sql

import (
	"fmt"
	"time"

	"github.com/ao9911/go-matrix/log"
	"github.com/ao9911/go-matrix/util/xtime"
)

type Config struct {
	DSN          string         // write data source name.
	ReadDSN      []string       // read data source name.
	Active       int            // pool
	Idle         int            // pool
	IdleTimeout  xtime.Duration // connect max life time.
	QueryTimeout xtime.Duration // query sql timeout
	ExecTimeout  xtime.Duration // execute sql timeout
	TranTimeout  xtime.Duration // transaction sql timeout
}

// NewMySQL new db instance .
func NewMySQL(c *Config) (db *DB) {
	if c == nil {
		panic("sql: nil mysql config")
	}

	if c.QueryTimeout == 0 {
		c.QueryTimeout = xtime.Duration(20 * time.Second)
		log.Warnf("NewMySQL QueryTimeout reset to default: %v", c.QueryTimeout)
	}
	if c.ExecTimeout == 0 {
		c.ExecTimeout = xtime.Duration(20 * time.Second)
		log.Warnf("NewMySQL ExecTimeout reset to default: %v", c.ExecTimeout)
	}
	if c.TranTimeout == 0 {
		c.TranTimeout = xtime.Duration(2 * time.Second)
		log.Warnf("NewMySQL TranTimeout reset to default: %v", c.TranTimeout)
	}

	db, err := Open(c)
	if err != nil {
		panic(fmt.Errorf("open mysql: %w", err))
	}
	return
}
