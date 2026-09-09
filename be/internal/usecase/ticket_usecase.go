package usecase

import (
	"context"
	"rebutin/internal/delivery/http/dto"
	"rebutin/internal/domain"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

type TicketUseCase struct {
	log  *logrus.Logger
	repo domain.TicketRepository
}

func NewTicketUseCase(log *logrus.Logger, repo domain.TicketRepository) *TicketUseCase {
	return &TicketUseCase{log: log, repo: repo}
}

func (r *TicketUseCase) GetTicketBySimulation(ctx context.Context, id uuid.UUID) ([]dto.TicketResponse, error) {
	tickets, err := r.repo.GetByRunID(ctx, id)
	if err != nil {
		return nil, err
	}

	return dto.ToTicketResponses(tickets), nil
}
