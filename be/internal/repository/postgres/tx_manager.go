package postgres

import (
	"context"
	"fmt"

	"rebutin/internal/domain"

	"github.com/jmoiron/sqlx"
)

type txKey struct{}

type TxManager struct {
	db *sqlx.DB
}

func NewTxManager(db *sqlx.DB) *TxManager {
	return &TxManager{db: db}
}

func injectTx(ctx context.Context, tx *sqlx.Tx) context.Context {
	return context.WithValue(ctx, txKey{}, tx)
}

func extractTx(ctx context.Context) (*sqlx.Tx, bool) {
	tx, ok := ctx.Value(txKey{}).(*sqlx.Tx)
	return tx, ok
}

func (tm *TxManager) WithTransaction(ctx context.Context, fn domain.AtomicFunc) error {
	tx, err := tm.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("fail start db tx: %w", err)
	}

	// inject *sqlx.Tx to context
	txCtx := injectTx(ctx, tx)

	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
			panic(p)
		}
	}()

	// run business flow
	if err := fn(txCtx); err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("err: %v, rollback err: %w", err, rbErr)
		}
		return err
	}

	return tx.Commit()
}
