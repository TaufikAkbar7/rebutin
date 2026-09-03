package domain

import "context"

type AtomicFunc func(ctx context.Context) error

type TransactionManager interface {
	WithTransaction(ctx context.Context, fn AtomicFunc) error
}
