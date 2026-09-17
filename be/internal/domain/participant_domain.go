package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type Participant struct {
	ID         uuid.UUID
	RunID      uuid.UUID
	IsBot      bool
	Identifier string
	CreatedAt  time.Time
}

const (
	IdentifierUser string = "real-user"
	IdentifierBot  string = "bot"
)

type ParticipantRepository interface {
	BatchCreate(ctx context.Context, participants []Participant) error
}
