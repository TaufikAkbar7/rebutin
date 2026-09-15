package postgres

import (
	"context"
	"fmt"
	"rebutin/internal/domain"
	"rebutin/internal/model"

	"github.com/jmoiron/sqlx"
	"github.com/sirupsen/logrus"
)

type participantRepository struct {
	db  *sqlx.DB
	log *logrus.Logger
}

func NewParticipantRepository(db *sqlx.DB, log *logrus.Logger) domain.ParticipantRepository {
	return &participantRepository{db: db, log: log}
}

func (r *participantRepository) BatchCreate(ctx context.Context, participants []domain.Participant) error {
	if len(participants) == 0 {
		return nil
	}

	const query = `
		INSERT INTO participants (
			id, run_id, is_bot, identifier, created_at
		) VALUES (
			:id, :run_id, :is_bot, :identifier, :created_at
		)`

	payload := make([]model.ParticipantModel, 0, len(participants))
	for _, item := range participants {
		payload = append(payload, model.ToParticipantDBModel(&item))
	}

	executor := getExecutor(ctx, r.db)

	_, err := executor.NamedExecContext(ctx, query, payload)
	if err != nil {
		r.log.Errorf("[ParticipantRepository.BatchCreate] Error insert participant: %v", err)
		return fmt.Errorf("failed insert participant")
	}

	return nil
}
