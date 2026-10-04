package worker_test

import (
	"context"
	"rebutin/internal/delivery/worker"
	"rebutin/pkg/testutil"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestDrainerScheduler_Start_CallsDrainerRun(t *testing.T) {
	repo := new(MockSessionRedisRepo)
	logger, _ := testutil.SetupLogger(t)

	d := worker.NewDrainer(repo, 10*time.Millisecond, logger)
	s := worker.NewDrainerScheduler(context.Background(), d, logger)

	called := make(chan struct{}, 1)
	repo.On("Drain", mock.Anything, "run-1", 5, 2).
		Run(func(args mock.Arguments) {
			select {
			case called <- struct{}{}:
			default:
			}
		}).
		Return([]string{}, nil).Maybe()

	s.Start("run-1", 5, 2)
	defer s.Stop("run-1")

	select {
	case <-called:
	case <-time.After(200 * time.Millisecond):
		t.Fatal("scheduler did not start the drainer")
	}
}

func TestDrainerScheduler_Start_IsIdempotent(t *testing.T) {
	repo := new(MockSessionRedisRepo)
	logger, _ := testutil.SetupLogger(t)

	d := worker.NewDrainer(repo, 10*time.Millisecond, logger)
	s := worker.NewDrainerScheduler(context.Background(), d, logger)

	var callCount int32
	repo.On("Drain", mock.Anything, "run-1", 5, 2).
		Run(func(args mock.Arguments) { atomic.AddInt32(&callCount, 1) }).
		Return([]string{}, nil).Maybe()

	s.Start("run-1", 5, 2)
	s.Start("run-1", 5, 2)
	defer s.Stop("run-1")

	require.Eventually(t, func() bool {
		return atomic.LoadInt32(&callCount) >= 1
	}, 200*time.Millisecond, 5*time.Millisecond)

	s.Stop("run-1")
	count := atomic.LoadInt32(&callCount)
	time.Sleep(50 * time.Millisecond)
	assert.Equal(t, count, atomic.LoadInt32(&callCount), "drainer should have stopped, no more Drain calls")
}

func TestDrainerScheduler_Stop_CancelsRun(t *testing.T) {
	repo := new(MockSessionRedisRepo)
	logger, _ := testutil.SetupLogger(t)

	d := worker.NewDrainer(repo, 10*time.Millisecond, logger)
	s := worker.NewDrainerScheduler(context.Background(), d, logger)

	var callCount int32
	tickHappened := make(chan struct{}, 1)
	repo.On("Drain", mock.Anything, "run-1", 5, 2).
		Run(func(args mock.Arguments) {
			atomic.AddInt32(&callCount, 1)

			select {
			case tickHappened <- struct{}{}:
			default:
			}
		}).
		Return([]string{}, nil).Maybe()

	s.Start("run-1", 5, 2)

	// wait after 1 tick is working
	select {
	case <-tickHappened:
	case <-time.After(200 * time.Millisecond):
		t.Fatal("drainer never ticked before Stop")
	}

	s.Stop("run-1")
	countAtStop := atomic.LoadInt32(&callCount)

	// make sure no new calls after run Stop()
	time.Sleep(50 * time.Millisecond)
	assert.Equal(t, countAtStop, atomic.LoadInt32(&callCount), "Drain should not be called after Stop")
}

func TestDrainerScheduler_Stop_UnknownRunID_NoPanic(t *testing.T) {
	repo := new(MockSessionRedisRepo)
	logger, _ := testutil.SetupLogger(t)
	d := worker.NewDrainer(repo, 10*time.Millisecond, logger)
	s := worker.NewDrainerScheduler(context.Background(), d, logger)

	assert.NotPanics(t, func() {
		s.Stop("run-does-not-exist")
	})
}
