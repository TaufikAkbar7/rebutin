package usecase_test

import (
	"context"
	"rebutin/internal/domain"

	"github.com/stretchr/testify/mock"
)

type MockTxManager struct{ mock.Mock }

func (m *MockTxManager) WithTransaction(ctx context.Context, fn domain.AtomicFunc) error {
	args := m.Called(ctx, fn)

	err := args.Error(0)

	if err != nil {
		return err
	}

	return fn(ctx)
}
