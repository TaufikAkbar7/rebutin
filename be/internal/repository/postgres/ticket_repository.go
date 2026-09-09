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

type ticketRepository struct {
	db  *sqlx.DB
	log *logrus.Logger
}

func NewTicketRepository(db *sqlx.DB, log *logrus.Logger) domain.TicketRepository {
	return &ticketRepository{db: db, log: log}
}

func (r *ticketRepository) BatchCreate(ctx context.Context, categories []domain.TicketCategories) error {
	if len(categories) == 0 {
		return nil
	}

	const query = `
		INSERT INTO ticket_categories (
			id, run_id, name, quota, price, created_at
		) VALUES (
			:id, :run_id, :name, :quota, :price, :created_at
		)`

	payload := make([]model.TicketModel, 0, len(categories))
	for _, item := range categories {
		payload = append(payload, model.ToTicketDBModel(&item))
	}

	executor := getExecutor(ctx, r.db)

	_, err := executor.NamedExecContext(ctx, query, payload)
	if err != nil {
		r.log.Errorf("[TicketRepository.Create] Error insert ticket: %v", err)
		return fmt.Errorf("failed insert ticket")
	}

	return nil
}

func (r *ticketRepository) GetByRunID(ctx context.Context, runID uuid.UUID) ([]domain.TicketCategories, error) {
	const query = `
		SELECT 
			id, run_id, name, quota, price
		FROM ticket_categories 
		WHERE run_id = $1`

	var model []model.TicketModel

	executor := getExecutor(ctx, r.db)
	if err := executor.SelectContext(ctx, &model, query, runID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		r.log.Errorf("[TicketRepository.GetByRunID] Error query ticket: %v", err)
		return nil, fmt.Errorf("failed query ticket")
	}

	var results []domain.TicketCategories
	for _, item := range model {
		results = append(results, *item.ToDomain())
	}

	return results, nil
}
