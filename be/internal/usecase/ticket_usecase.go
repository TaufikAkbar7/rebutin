package usecase

import (
	"context"
	"rebutin/internal/delivery/http/dto"
	"rebutin/internal/domain"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

type TicketUseCase interface {
	GetTicketBySimulation(ctx context.Context, id uuid.UUID) ([]dto.TicketResponse, error)
}

type ticketUseCase struct {
	log  *logrus.Logger
	repo domain.TicketRepository
}

func NewTicketUseCase(log *logrus.Logger, repo domain.TicketRepository) TicketUseCase {
	return &ticketUseCase{log: log, repo: repo}
}

func (r *ticketUseCase) GetTicketBySimulation(ctx context.Context, id uuid.UUID) ([]dto.TicketResponse, error) {
	tickets, err := r.repo.GetByRunID(ctx, id)
	if err != nil {
		return nil, err
	}

	return dto.ToTicketResponses(tickets), nil
}
