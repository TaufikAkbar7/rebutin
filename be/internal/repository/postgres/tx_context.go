package postgres

import (
	"context"
	"database/sql"

	"github.com/jmoiron/sqlx"
)

type queryer interface {
	NamedExecContext(ctx context.Context, query string, arg any) (sql.Result, error)
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	GetContext(ctx context.Context, dest any, query string, args ...any) error
	SelectContext(ctx context.Context, dest any, query string, args ...any) error
	QueryxContext(ctx context.Context, query string, args ...any) (*sqlx.Rows, error)
	QueryRowxContext(ctx context.Context, query string, args ...any) *sqlx.Row
}

func getExecutor(ctx context.Context, db *sqlx.DB) queryer {
	if tx, ok := extractTx(ctx); ok && tx != nil {
		return tx
	}
	return db
}
