package model

import (
	"time"

	"rebutin/internal/domain"

	"github.com/google/uuid"
)

type SimulationRunModel struct {
	ID                 uuid.UUID `db:"id"`
	TotalTickets       int       `db:"total_tickets"`
	BotCount           int       `db:"bot_count"`
	BotThrottleSeconds int       `db:"bot_throttle_seconds"`
	MaxConcurrent      int       `db:"max_concurrent"`
	Status             string    `db:"status"`
	CreatedAt          time.Time `db:"created_at"`
}

// Mapper: DB Model -> Domain Entity
func (m *SimulationRunModel) ToDomain() *domain.SimulationRun {
	return &domain.SimulationRun{
		ID:                 m.ID,
		TotalTickets:       m.TotalTickets,
		BotCount:           m.BotCount,
		BotThrottleSeconds: m.BotThrottleSeconds,
		MaxConcurrent:      m.MaxConcurrent,
		Status:             domain.SimulationStatus(m.Status),
		CreatedAt:          m.CreatedAt,
	}
}

// Mapper: Domain Entity -> DB Model
func ToDBModel(d *domain.SimulationRun) *SimulationRunModel {
	return &SimulationRunModel{
		ID:                 d.ID,
		TotalTickets:       d.TotalTickets,
		BotCount:           d.BotCount,
		BotThrottleSeconds: d.BotThrottleSeconds,
		MaxConcurrent:      d.MaxConcurrent,
		Status:             string(d.Status),
		CreatedAt:          d.CreatedAt,
	}
}
