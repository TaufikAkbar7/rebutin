package dto

import (
	"github.com/google/uuid"
)

type CreateSimulationRequest struct {
	TotalTickets       int `json:"total_tickets" validate:"required,min=1"`
	BotCount           int `json:"bot_count" validate:"required,min=1"`
	BotThrottleSeconds int `json:"bot_throttle_seconds" validate:"gte=0"`
	MaxConcurrent      int `json:"max_concurrent" validate:"required,min=1"`
}

type SimulationResponse struct {
	ID     uuid.UUID `json:"id"`
	Status string    `json:"status"`
}
