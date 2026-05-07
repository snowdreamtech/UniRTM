package sqlite

import (
	"context"
	"database/sql"
)

// DBExecutor is a common interface for *sql.DB and *sql.Tx
// This allows repositories to work with both regular database connections
// and transactions transparently
type DBExecutor interface {
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
	Prepare(query string) (*sql.Stmt, error)
}
