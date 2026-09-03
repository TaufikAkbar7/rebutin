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
	vErr := &ValidationError{Errors: make(map[string]string)}

	if s.TotalTickets <= 0 {
		vErr.Add("total_tickets", "total tickets must be greater than 0")
	}
	if s.BotCount <= 0 {
		vErr.Add("bot_count", "bot count must be at least 1")
	}
	if s.MaxConcurrent <= 0 {
		vErr.Add("max_concurrent", "max concurrent must be at least 1")
	} else if s.MaxConcurrent > s.BotCount {
		vErr.Add("max_concurrent", "max concurrent cannot exceed total bot count")
	}
	if s.BotThrottleSeconds < 0 {
		vErr.Add("bot_throttle_seconds", "bot throttle seconds cannot be negative")
	}

	if vErr.HasErrors() {
		return vErr
	}
	return nil
}

func (s *SimulationRun) Stop() error {
	if s.Status == StatusEnded {
		return ErrSimulationAlreadyStopped
	}
	s.Status = StatusEnded
	return nil
}

type SimulationRepository interface {
	Create(ctx context.Context, sim *SimulationRun) error
	GetByID(ctx context.Context, id uuid.UUID) (*SimulationRun, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status SimulationStatus) error
}
