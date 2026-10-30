package repos

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// DBConnection is the subset of pgx.Conn used by repositories and handlers.
// Keeping the dependency structural lets unit tests use an in-memory SQL mock
// while production continues to pass a regular *pgx.Conn.
type DBConnection interface {
	Begin(context.Context) (pgx.Tx, error)
	Exec(context.Context, string, ...any) (pgconn.CommandTag, error)
	Query(context.Context, string, ...any) (pgx.Rows, error)
	QueryRow(context.Context, string, ...any) pgx.Row
	Close(context.Context) error
	Ping(context.Context) error
}
