package testing

import (
	"database/sql"
	"database/sql/driver"
	"errors"
	"io"
	"testing"
)

var (
	_ driver.Driver         = (*Driver)(nil)
	_ driver.Conn           = (*Conn)(nil)
	_ driver.ExecerContext  = (*Conn)(nil)
	_ driver.QueryerContext = (*Conn)(nil)
	_ driver.Stmt           = (*Stmt)(nil)
	_ driver.Rows           = (*Rows)(nil)
	_ driver.Tx             = (*Tx)(nil)
)

func TestDriverOpen(t *testing.T) {
	t.Parallel()

	conn, err := (&Driver{}).Open("dsn")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := conn.(*Conn); !ok {
		t.Fatalf("expected *Conn, got %T", conn)
	}
}

func TestConn(t *testing.T) {
	t.Parallel()

	conn := &Conn{}

	stmt, err := conn.Prepare("query")
	if err != nil {
		t.Fatalf("prepare: %v", err)
	}
	if _, ok := stmt.(*Stmt); !ok {
		t.Fatalf("expected *Stmt, got %T", stmt)
	}

	if err := conn.Close(); err != nil {
		t.Fatalf("close: %v", err)
	}

	tx, err := conn.Begin()
	if err != nil {
		t.Fatalf("begin: %v", err)
	}
	if _, ok := tx.(*Tx); !ok {
		t.Fatalf("expected *Tx, got %T", tx)
	}

	result, err := conn.ExecContext(t.Context(), "query", nil)
	if err != nil {
		t.Fatalf("exec context: %v", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		t.Fatalf("rows affected: %v", err)
	}
	if affected != 1 {
		t.Fatalf("expected 1 row affected, got %d", affected)
	}

	rows, err := conn.QueryContext(t.Context(), "query", nil)
	if err != nil {
		t.Fatalf("query context: %v", err)
	}
	if _, ok := rows.(*Rows); !ok {
		t.Fatalf("expected *Rows, got %T", rows)
	}
	if err := rows.Close(); err != nil {
		t.Fatalf("rows close: %v", err)
	}
}

func TestStmt(t *testing.T) {
	t.Parallel()

	stmt := &Stmt{}

	if got := stmt.NumInput(); got != -1 {
		t.Fatalf("expected NumInput -1, got %d", got)
	}

	result, err := stmt.Exec(nil)
	if err != nil {
		t.Fatalf("exec: %v", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		t.Fatalf("rows affected: %v", err)
	}
	if affected != 1 {
		t.Fatalf("expected 1 row affected, got %d", affected)
	}

	rows, err := stmt.Query(nil)
	if err != nil {
		t.Fatalf("query: %v", err)
	}
	if _, ok := rows.(*Rows); !ok {
		t.Fatalf("expected *Rows, got %T", rows)
	}

	if err := stmt.Close(); err != nil {
		t.Fatalf("close: %v", err)
	}
}

func TestRows(t *testing.T) {
	t.Parallel()

	rows := &Rows{}

	if cols := rows.Columns(); len(cols) != 0 {
		t.Fatalf("expected no columns, got %v", cols)
	}

	if err := rows.Next(nil); !errors.Is(err, io.EOF) {
		t.Fatalf("expected io.EOF, got %v", err)
	}

	if err := rows.Close(); err != nil {
		t.Fatalf("close: %v", err)
	}
}

func TestTx(t *testing.T) {
	t.Parallel()

	tx := &Tx{}

	if err := tx.Commit(); err != nil {
		t.Fatalf("commit: %v", err)
	}
	if err := tx.Rollback(); err != nil {
		t.Fatalf("rollback: %v", err)
	}
}

func TestNewDb(t *testing.T) {
	t.Parallel()

	if DriverName != "testdb" {
		t.Fatalf("expected DriverName %q, got %q", "testdb", DriverName)
	}

	db := NewDb()
	if db == nil {
		t.Fatal("expected non-nil db")
	}
	t.Cleanup(func() { _ = db.Close() })

	ctx := t.Context()

	result, err := db.ExecContext(ctx, "INSERT INTO t VALUES (1)")
	if err != nil {
		t.Fatalf("exec context: %v", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		t.Fatalf("rows affected: %v", err)
	}
	if affected != 1 {
		t.Fatalf("expected 1 row affected, got %d", affected)
	}

	rows, err := db.QueryContext(ctx, "SELECT 1")
	if err != nil {
		t.Fatalf("query context: %v", err)
	}
	if rows.Next() {
		t.Fatal("expected an empty result set")
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("rows err: %v", err)
	}
	if err := rows.Close(); err != nil {
		t.Fatalf("rows close: %v", err)
	}

	var value int
	if err := db.QueryRowContext(ctx, "SELECT 1").Scan(&value); !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("expected sql.ErrNoRows, got %v", err)
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("begin tx: %v", err)
	}
	if err := tx.Commit(); err != nil {
		t.Fatalf("commit: %v", err)
	}
}

func TestNewDbRepeatable(t *testing.T) {
	t.Parallel()

	first := NewDb()
	if first == nil {
		t.Fatal("expected non-nil db on first call")
	}
	t.Cleanup(func() { _ = first.Close() })

	second := NewDb()
	if second == nil {
		t.Fatal("expected non-nil db on second call")
	}
	t.Cleanup(func() { _ = second.Close() })
}
