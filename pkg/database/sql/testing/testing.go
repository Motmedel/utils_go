package testing

import (
	"context"
	"database/sql"
	"database/sql/driver"
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

// A fake SQL driver for testing that reports one row affected.

type Driver struct{}

func (d *Driver) Open(name string) (driver.Conn, error) { return &Conn{}, nil }

type Stmt struct{}

func (s *Stmt) Close() error  { return nil }
func (s *Stmt) NumInput() int { return -1 }
func (s *Stmt) Exec(_ []driver.Value) (driver.Result, error) {
	return driver.RowsAffected(1), nil
}
func (s *Stmt) Query(_ []driver.Value) (driver.Rows, error) { return nil, driver.ErrSkip }

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
