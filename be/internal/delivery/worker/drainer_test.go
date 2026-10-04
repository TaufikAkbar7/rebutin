package worker_test

import (
	"context"
	"errors"
	"rebutin/internal/delivery/worker"
	"rebutin/internal/domain"
	"rebutin/pkg/testutil"
	"sync/atomic"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type MockDrainer struct{ mock.Mock }

func (m *MockDrainer) Run(ctx context.Context, runID string, max, batch int) error {
	args := m.Called(ctx, runID, max, batch)

	return args.Error(0)
}

type MockSessionRedisRepo struct{ mock.Mock }

func (m *MockSessionRedisRepo) RunState(ctx context.Context, runID string, max int, data []domain.UserSessionRedis) error {
	args := m.Called(ctx, runID, max, data)
	return args.Error(0)
}

func (m *MockSessionRedisRepo) GetValueAllField(ctx context.Context, key string) (*domain.UserSessionRedis, error) {
	args := m.Called(ctx, key)

	val := args.Get(0)
	if val == nil {
		return nil, args.Error(1)
	}

	return val.(*domain.UserSessionRedis), args.Error(1)
}

func (m *MockSessionRedisRepo) GetValueByField(ctx context.Context, key string, field string) (*string, error) {
	args := m.Called(ctx, key)

	val := args.Get(0)
	if val == nil {
		return nil, args.Error(1)
	}

	return val.(*string), args.Error(1)
}

func (m *MockSessionRedisRepo) Drain(ctx context.Context, runID string, max, batch int) ([]string, error) {
	args := m.Called(ctx, runID, max, batch)

	val := args.Get(0)
	if val == nil {
		return nil, args.Error(1)
	}

	return val.([]string), args.Error(1)
}

func TestDrainer_Run_InvalidBatch(t *testing.T) {
	repo := new(MockSessionRedisRepo)
	logger, _ := testutil.SetupLogger(t)

	d := worker.NewDrainer(repo, 10*time.Millisecond, logger)

	err := d.Run(context.Background(), "run-1", 5, 0)

	assert.Error(t, err)
	repo.AssertNotCalled(t, "Drain", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
}

func TestDrainer_Run_ClampBatchToMax(t *testing.T) {
	repo := new(MockSessionRedisRepo)
	logger, _ := testutil.SetupLogger(t)

	d := worker.NewDrainer(repo, 10*time.Millisecond, logger)

	called := make(chan struct{}, 1)
	repo.On("Drain", mock.Anything, "run-1", 3, 3).Run(func(args mock.Arguments) {
		called <- struct{}{}
	}).Return([]string{}, nil).Maybe()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go d.Run(ctx, "run-1", 3, 10)

	select {
	case <-called:
	case <-time.After(200 * time.Millisecond):
		t.Fatal("func Drain not called with clamped batch")
	}
}

func TestDrainer_Run_StopsOnContextCancel(t *testing.T) {
	repo := new(MockSessionRedisRepo)
	logger, _ := testutil.SetupLogger(t)
	d := worker.NewDrainer(repo, 5*time.Millisecond, logger)

	tickHappened := make(chan struct{}, 1)
	repo.On("Drain", mock.Anything, "run-1", 5, 2).
		Run(func(args mock.Arguments) {
			select {
			case tickHappened <- struct{}{}:
			default:
			}
		}).
		Return([]string{}, nil).Maybe()

	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan error, 1)
	go func() { done <- d.Run(ctx, "run-1", 5, 2) }()

	// wait after 1 tick is working
	<-tickHappened
	cancel()

	select {
	case err := <-done:
		assert.NoError(t, err)
	case <-time.After(200 * time.Millisecond):
		t.Fatal("Run did not return after context cancel")
	}
}

func TestDrainer_Run_LogsAdmittedUsers(t *testing.T) {
	repo := new(MockSessionRedisRepo)
	logger, hook := testutil.SetupLogger(t)
	d := worker.NewDrainer(repo, 10*time.Millisecond, logger)

	var callCount int32

	repo.On("Drain", mock.Anything, "run-1", 2, 2).
		Run(func(args mock.Arguments) { atomic.AddInt32(&callCount, 1) }).
		Return([]string{"session:one"}, nil).Once()
	repo.On("Drain", mock.Anything, "run-1", 2, 2).Return([]string{}, nil).Maybe()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go d.Run(ctx, "run-1", 2, 2)

	require.Eventually(t, func() bool {
		return atomic.LoadInt32(&callCount) >= 1
	}, 200*time.Millisecond, 5*time.Millisecond)

	assert.Equal(t, 1, len(hook.Entries))
	assert.Equal(t, logrus.InfoLevel, hook.LastEntry().Level)
	assert.Contains(t, hook.LastEntry().Message, "participant left the waiting room")
}

func TestDrainer_Run_LogsErrorButKeepingRun(t *testing.T) {
	repo := new(MockSessionRedisRepo)
	logger, hook := testutil.SetupLogger(t)
	d := worker.NewDrainer(repo, 10*time.Millisecond, logger)

	var callCount int32

	repo.On("Drain", mock.Anything, "run-1", 2, 1).
		Run(func(args mock.Arguments) { atomic.AddInt32(&callCount, 1) }).
		Return([]string{}, errors.New("redis error"))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go d.Run(ctx, "run-1", 2, 1)

	require.Eventually(t, func() bool {
		return atomic.LoadInt32(&callCount) >= 2
	}, 200*time.Millisecond, 5*time.Millisecond)

	assert.Equal(t, 2, len(hook.Entries))
	assert.Equal(t, logrus.ErrorLevel, hook.LastEntry().Level)
	assert.Contains(t, hook.LastEntry().Message, "[Drainer.Run] Failed drain run")
}
