package model

import (
	"rebutin/internal/domain"

	"github.com/google/uuid"
)

type SessionModel struct {
	ParticipantID     string `redis:"participant_id"`
	RunID             string `redis:"run_id"`
	Status            string `redis:"status"`
	CurrentCategoryID string `redis:"current_category_id"`
}

func ToUserSessionRedisModel(d *domain.UserSessionRedis) *SessionModel {
	var categoryId string = ""
	if d.CurrentCategoryID != nil {
		categoryId = d.CurrentCategoryID.String()
	}

	return &SessionModel{
		ParticipantID:     d.ParticipantID.String(),
		RunID:             d.RunID.String(),
		CurrentCategoryID: categoryId,
		Status:            string(d.Status),
	}
}

func parseUUID(val string) uuid.UUID {
	id, _ := uuid.Parse(val)
	return id
}

func parseUUIDPtr(val string) *uuid.UUID {
	if val == "" {
		return nil
	}
	id, err := uuid.Parse(val)
	if err != nil || id == uuid.Nil {
		return nil
	}
	return &id
}

func ToDomainUserSession(d map[string]string) *domain.UserSessionRedis {
	return &domain.UserSessionRedis{
		RunID:             parseUUID(d["run_id"]),
		ParticipantID:     parseUUID(d["participant_id"]),
		CurrentCategoryID: parseUUIDPtr(d["current_category_id"]),
		Status:            domain.SessionStatus(d["status"]),
	}
}
