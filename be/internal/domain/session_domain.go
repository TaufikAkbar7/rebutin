package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type SessionStatus string

const (
	StatusQueued    SessionStatus = "queued"
	StatusActive    SessionStatus = "active"
	StatusExpired   SessionStatus = "expired"
	StatusCompleted SessionStatus = "completed"
	MaxLeaky        int           = 10
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

type WaitingRoomQueueSession struct {
	Key   string
	Value int64
}

type SessionCacheRepository interface {
	RunState(ctx context.Context, runID string, max int, data []UserSessionRedis) error
	GetValueAllField(ctx context.Context, key string) (*UserSessionRedis, error)
	GetValueByField(ctx context.Context, key string, field string) (*string, error)
	Drain(ctx context.Context, runID string, max, batch int) ([]string, error)
}
