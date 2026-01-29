package testing

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"io"
	"sync"
)

type Conn struct{}

func (c *Conn) Prepare(_ string) (driver.Stmt, error) { return &Stmt{}, nil }
func (c *Conn) Close() error                          { return nil }
func (c *Conn) Begin() (driver.Tx, error)             { return &Tx{}, nil }

func (c *Conn) ExecContext(ctx context.Context, _ string, _ []driver.NamedValue) (driver.Result, error) {
	return driver.RowsAffected(1), nil
}

var _ driver.ExecerContext = (*Conn)(nil)
var _ driver.QueryerContext = (*Conn)(nil)

// QueryContext returns an empty result set to avoid fast-path ErrSkip propagation
// from the driver during tests.
func (c *Conn) QueryContext(ctx context.Context, _ string, _ []driver.NamedValue) (driver.Rows, error) {
	return &Rows{}, nil
}

// A fake SQL driver for testing that reports one row affected.

type Driver struct{}

func (d *Driver) Open(name string) (driver.Conn, error) { return &Conn{}, nil }

type Stmt struct{}

func (s *Stmt) Close() error  { return nil }
func (s *Stmt) NumInput() int { return -1 }
func (s *Stmt) Exec(_ []driver.Value) (driver.Result, error) {
	return driver.RowsAffected(1), nil
}
func (s *Stmt) Query(_ []driver.Value) (driver.Rows, error) { return &Rows{}, nil }

// Rows is a minimal implementation that represents an empty result set.
type Rows struct{}

func (r *Rows) Columns() []string { return []string{} }
func (r *Rows) Close() error      { return nil }
func (r *Rows) Next(_ []driver.Value) error {
	return io.EOF
}

type Tx struct{}

func (t *Tx) Commit() error   { return nil }
func (t *Tx) Rollback() error { return nil }

var (
	registerOnce sync.Once
	DriverName   = "testdb"
)

func NewDb() *sql.DB {
	registerOnce.Do(func() {
		sql.Register(DriverName, &Driver{})
	})
	db, _ := sql.Open(DriverName, "")
	return db
}
