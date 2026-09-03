package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"rebutin/internal/domain"

	"rebutin/internal/model"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/sirupsen/logrus"
)

type simulationRepository struct {
	db  *sqlx.DB
	log *logrus.Logger
}

func NewSimulationRepository(db *sqlx.DB, log *logrus.Logger) domain.SimulationRepository {
	return &simulationRepository{db: db, log: log}
}

func (r *simulationRepository) Create(ctx context.Context, sim *domain.SimulationRun) error {
	const query = `
		INSERT INTO simulation_runs (
			id, total_tickets, bot_count, bot_throttle_seconds, max_concurrent, status, created_at
		) VALUES (
			:id, :total_tickets, :bot_count, :bot_throttle_seconds, :max_concurrent, :status, :created_at
		)`

	model := model.ToDBModel(sim)

	executor := getExecutor(ctx, r.db)

	_, err := executor.NamedExecContext(ctx, query, model)
	if err != nil {
		r.log.Errorf("[SimulasitionRepository.Create] Error insert simulation run: %v", err)
		return fmt.Errorf("failed insert simulation run")
	}

	return nil
}

func (r *simulationRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.SimulationRun, error) {
	const query = `
		SELECT 
			id, total_tickets, bot_count, bot_throttle_seconds, max_concurrent, status, created_at 
		FROM simulation_runs 
		WHERE id = $1`

	var model model.SimulationRunModel

	executor := getExecutor(ctx, r.db)
	if err := executor.GetContext(ctx, &model, query, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		r.log.Errorf("[SimulasitionRepository.GetByID] Error query simulation run: %v", err)
		return nil, fmt.Errorf("failed query simulation run")
	}

	return model.ToDomain(), nil
}

func (r *simulationRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status domain.SimulationStatus) error {
	const query = `
		UPDATE simulation_runs 
		SET status = $1 
		WHERE id = $2`

	executor := getExecutor(ctx, r.db)
	res, err := executor.ExecContext(ctx, query, string(status), id)
	if err != nil {
		r.log.Errorf("[SimulasitionRepository.UpdateStatus] Error failed update status simulation run: %v", err)
		return fmt.Errorf("failed update status simulation run")
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		r.log.Errorf("[SimulasitionRepository.UpdateStatus] Error read rows affected simulation run: %v", err)
		return fmt.Errorf("failed read rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return domain.ErrNotFound
	}

	return nil
}
