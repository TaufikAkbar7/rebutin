package model

import (
	"time"

	"rebutin/internal/domain"

	"github.com/google/uuid"
)

type TicketModel struct {
	ID        uuid.UUID `db:"id"`
	RunID     uuid.UUID `db:"run_id"`
	Name      string    `db:"name"`
	Quota     int       `db:"quota"`
	Price     float64   `db:"price"`
	CreatedAt time.Time `db:"created_at"`
}

// Mapper: DB Model -> Domain Entity
func (m *TicketModel) ToDomain() *domain.TicketCategories {
	if m == nil {
		return nil
	}
	return &domain.TicketCategories{
		ID:        m.ID,
		RunID:     m.RunID,
		Name:      m.Name,
		Quota:     m.Quota,
		Price:     m.Price,
		CreatedAt: m.CreatedAt,
	}
}

// Mapper: Domain Entity -> DB Model
func ToTicketDBModel(d *domain.TicketCategories) TicketModel {
	return TicketModel{
		ID:        d.ID,
		RunID:     d.RunID,
		Name:      d.Name,
		Quota:     d.Quota,
		Price:     d.Price,
		CreatedAt: d.CreatedAt,
	}
}
