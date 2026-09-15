package model

import (
	"rebutin/internal/domain"
	"time"

	"github.com/google/uuid"
)

type ParticipantModel struct {
	ID         uuid.UUID `db:"id"`
	RunID      uuid.UUID `db:"run_id"`
	IsBot      bool      `db:"is_bot"`
	Identifier string    `db:"identifier"`
	CreatedAt  time.Time `db:"created_at"`
}

// Mapper: DB Model -> Domain Entity
func (m *ParticipantModel) ToDomain() *domain.Participant {
	if m == nil {
		return nil
	}
	return &domain.Participant{
		ID:         m.ID,
		RunID:      m.RunID,
		IsBot:      m.IsBot,
		Identifier: m.Identifier,
		CreatedAt:  m.CreatedAt,
	}
}

// Mapper: Domain Entity -> DB Model
func ToParticipantDBModel(d *domain.Participant) ParticipantModel {
	return ParticipantModel{
		ID:         d.ID,
		RunID:      d.RunID,
		IsBot:      d.IsBot,
		Identifier: d.Identifier,
		CreatedAt:  d.CreatedAt,
	}
}
