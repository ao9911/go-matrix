// query.go
package sql

import (
	"context"
	"database/sql"
	"time"
)

// Query executes a query against a read connection when one is configured.
func (db *DB) Query(c context.Context, query string, args ...interface{}) (rows *sql.Rows, err error) {
	idx := db.readIndex()
	if len(db.read) > idx {
		if rows, err = db.read[idx].QueryContext(c, query, args...); err != nil {
			return nil, err
		}
		return rows, nil
	}
	return db.write.QueryContext(c, query, args...)
}

// QueryRow executes a query expected to return at most one row.
func (db *DB) QueryRow(c context.Context, query string, args ...interface{}) (row *sql.Row) {
	idx := db.readIndex()
	if len(db.read) > idx {
		return db.read[idx].QueryRowContext(c, query, args...)
	}
	return db.write.QueryRowContext(c, query, args...)
}

// Qurey is kept for backward compatibility. Deprecated: use Query.
func (db *DB) Qurey(c context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	return db.Query(c, query, args...)
}

// QureyRow is kept for backward compatibility. Deprecated: use QueryRow.
func (db *DB) QureyRow(c context.Context, query string, args ...interface{}) *sql.Row {
	return db.QueryRow(c, query, args...)
}

// Exec executes a query against the write connection.
func (db *DB) Exec(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	return db.write.ExecContext(ctx, query, args...)
}

// BeginTx starts a transaction using the master connection.
func (db *DB) BeginTx(ctx context.Context, opts *sql.TxOptions) (*Tx, error) {
	var (
		txCtx  context.Context
		cancel context.CancelFunc
	)
	if timeout := time.Duration(db.write.conf.TranTimeout); timeout > 0 {
		txCtx, cancel = context.WithTimeout(ctx, timeout)
	} else {
		txCtx, cancel = context.WithCancel(ctx)
	}

	tx, err := db.write.BeginTx(txCtx, opts)
	if err != nil {
		cancel()
		return nil, err
	}
	return &Tx{Tx: tx, cancel: cancel}, nil
}

// Commit commits the transaction and releases its context.
func (t *Tx) Commit() error {
	defer t.cancel()
	return t.Tx.Commit()
}

// Rollback rolls back the transaction and releases its context.
func (t *Tx) Rollback() error {
	defer t.cancel()
	return t.Tx.Rollback()
}
