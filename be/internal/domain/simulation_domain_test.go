package domain_test

import (
	"errors"
	"rebutin/internal/domain"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSimulationRun_Validate(t *testing.T) {
	t.Run("should pass when data is valid", func(t *testing.T) {
		sim := domain.SimulationRun{
			TotalTickets:       100,
			BotCount:           5,
			BotThrottleSeconds: 0,
			MaxConcurrent:      2,
		}

		err := sim.Validate()
		assert.NoError(t, err)
	})

	t.Run("should fail with multiple field errors", func(t *testing.T) {
		sim := domain.SimulationRun{
			Status:    domain.StatusRunning,
			CreatedAt: time.Now(),
		}

		err := sim.Validate()
		require.Error(t, err)

		// extract struct error
		var valErr *domain.ValidationError
		ok := errors.As(err, &valErr)
		require.True(t, ok, "error should be of type *domain.ValidationErrors")

		assert.Len(t, valErr.Errors, 3)
		assert.Equal(t, "total tickets must be greater than 0", valErr.Errors["total_tickets"])
		assert.Equal(t, "bot count must be at least 1", valErr.Errors["bot_count"])
		assert.Equal(t, "max concurrent must be at least 1", valErr.Errors["max_concurrent"])
	})

	t.Run("should fail when bot throttle is negative", func(t *testing.T) {
		sim := domain.SimulationRun{
			BotThrottleSeconds: -1,
		}

		err := sim.Validate()
		assert.ErrorContains(t, err, "bot throttle seconds cannot be negative")
	})

	t.Run("should fail when total tickets is zero or negative", func(t *testing.T) {
		sim := domain.SimulationRun{
			TotalTickets:       0,
			BotCount:           10,
			BotThrottleSeconds: 0,
			MaxConcurrent:      1,
		}

		err := sim.Validate()
		assert.ErrorContains(t, err, "total tickets must be greater than 0")
	})

	t.Run("should fail when max concurrent exceed bot count", func(t *testing.T) {
		sim := domain.SimulationRun{
			MaxConcurrent: 10,
			BotCount:      2,
		}

		err := sim.Validate()
		assert.ErrorContains(t, err, "max concurrent cannot exceed total bot count")
	})
}

func TestSimulationRun_Stop(t *testing.T) {
	t.Run("should fail when status already ended", func(t *testing.T) {
		sim := domain.SimulationRun{
			Status: domain.StatusEnded,
		}

		err := sim.Stop()
		assert.ErrorContains(t, err, "simulation run is already ended")
	})

	t.Run("should pass when status set to running", func(t *testing.T) {
		sim := domain.SimulationRun{
			Status: domain.StatusRunning,
		}

		err := sim.Stop()
		assert.NoError(t, err)
	})
}
