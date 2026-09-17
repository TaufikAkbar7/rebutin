package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type SessionStatus string

const (
	StatusQueue     SessionStatus = "queue"
	StatusActive    SessionStatus = "active"
	StatusExpired   SessionStatus = "expired"
	StatusCompleted SessionStatus = "completed"
)

type UserSession struct {
	ID            uuid.UUID
	ParticipantID uuid.UUID
	RunID         uuid.UUID
	Status        SessionStatus
	CreatedAt     time.Time
	ExpiresAt     time.Time
	QueueAt       time.Time
	TTLStartedAt  time.Time
}

type UserSessionRedis struct {
	SessionID         *string
	ParticipantID     uuid.UUID
	RunID             uuid.UUID
	Status            SessionStatus
	CurrentCategoryID *uuid.UUID
}

type SessionCacheRepository interface {
	SetBatch(ctx context.Context, data []UserSessionRedis) error
	GetValueAllField(ctx context.Context, key string) (*UserSessionRedis, error)
	GetValueByField(ctx context.Context, key string, field string) (*string, error)
}
