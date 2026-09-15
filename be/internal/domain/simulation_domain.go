package domain

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
)

type SimulationStatus string

const (
	StatusRunning SimulationStatus = "running"
	StatusEnded   SimulationStatus = "ended"
)

var (
	ErrSimulationAlreadyStopped = errors.New("simulation run is already ended")
)

func newValidationError() *ValidationError {
	return &ValidationError{Errors: make(map[string]string)}
}

type SimulationRun struct {
	ID                 uuid.UUID
	TotalTickets       int
	BotCount           int
	BotThrottleSeconds int
	MaxConcurrent      int
	Status             SimulationStatus
	CreatedAt          time.Time
}

func (s *SimulationRun) Validate() error {
	vErr := newValidationError()

	if s.TotalTickets <= 0 {
		vErr.Add("total_tickets", "total tickets must be greater than 0")
	}

	if s.BotCount <= 0 {
		vErr.Add("bot_count", "bot count must be at least 1")
	} else if s.BotCount > 1000 {
		vErr.Add("bot_count", "bot count exceeds the maximum allowed value 1000")
	}

	if s.MaxConcurrent <= 0 {
		vErr.Add("max_concurrent", "max concurrent must be at least 1")
	} else if s.MaxConcurrent > s.BotCount {
		vErr.Add("max_concurrent", "max concurrent cannot exceed total bot count")
	} else if s.MaxConcurrent > 80 {
		vErr.Add("max_concurrent", "max concurrent exceeds the maximum allowed value 80")
	}

	if s.BotThrottleSeconds < 0 {
		vErr.Add("bot_throttle_seconds", "bot throttle seconds cannot be negative")
	}

	if s.BotThrottleSeconds < 0 {
		vErr.Add("bot_throttle_seconds", "bot throttle seconds cannot be negative")
	} else if s.BotThrottleSeconds > 10 {
		vErr.Add("bot_throttle_seconds", "bot throttle seconds exceeds the maximum allowed value 10")
	}

	if vErr.HasErrors() {
		return vErr
	}
	return nil
}

func (s *SimulationRun) Stop() error {
	if s.Status == StatusEnded {
		vErr := newValidationError()
		vErr.Add("status", ErrSimulationAlreadyStopped.Error())
		if vErr.HasErrors() {
			return vErr
		}
	}
	return nil
}

type SimulationRepository interface {
	Create(ctx context.Context, sim *SimulationRun) error
	GetByID(ctx context.Context, id uuid.UUID) (*SimulationRun, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status SimulationStatus) error
}
