package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type TicketCategories struct {
	ID        uuid.UUID
	RunID     uuid.UUID
	Name      string
	Quota     int
	Price     float64
	CreatedAt time.Time
}

type TicketItem struct {
	Name  string
	Price float64
}

var (
	CategoryCAT1 = TicketItem{Name: "CAT 1", Price: 1500000}
	CategoryCAT2 = TicketItem{Name: "CAT 2", Price: 1000000}
	CategoryCAT3 = TicketItem{Name: "CAT 3", Price: 500000}
)

var DefaultCategories = []TicketItem{
	CategoryCAT1,
	CategoryCAT2,
	CategoryCAT3,
}

type TicketRepository interface {
	BatchCreate(ctx context.Context, payload []TicketCategories) error
	GetByRunID(ctx context.Context, runID uuid.UUID) ([]TicketCategories, error)
}
