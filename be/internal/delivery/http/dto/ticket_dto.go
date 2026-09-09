package dto

import (
	"rebutin/internal/domain"

	"github.com/google/uuid"
)

type TicketResponse struct {
	ID    uuid.UUID `json:"id"`
	Name  string    `json:"name"`
	Quota int       `json:"quota"`
	Price float64   `json:"price"`
}

func ToTicketResponse(ticket domain.TicketCategories) TicketResponse {
	return TicketResponse{
		ID:    ticket.ID,
		Name:  ticket.Name,
		Quota: ticket.Quota,
		Price: ticket.Price,
	}
}

func ToTicketResponses(tickets []domain.TicketCategories) []TicketResponse {
	responses := make([]TicketResponse, len(tickets))
	for i, t := range tickets {
		responses[i] = ToTicketResponse(t)
	}
	return responses
}
