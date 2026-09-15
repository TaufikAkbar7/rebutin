package dto

import (
	"github.com/google/uuid"
)

type CreateSimulationRequest struct {
	TotalTickets       int `json:"total_tickets" binding:"required,min=1"`
	BotCount           int `json:"bot_count" binding:"required,min=1,max=1000"`
	BotThrottleSeconds int `json:"bot_throttle_seconds" binding:"gte=0,max=10"`
	MaxConcurrent      int `json:"max_concurrent" binding:"required,min=1,max=80"`
}

type SimulationResponse struct {
	ID     uuid.UUID `json:"id"`
	Status string    `json:"status"`
}
