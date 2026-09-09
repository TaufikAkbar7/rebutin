package usecase

import (
	"context"
	"fmt"
	"time"

	"rebutin/internal/delivery/http/dto"
	"rebutin/internal/domain"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

type SimulationUseCase struct {
	txManager  domain.TransactionManager
	repo       domain.SimulationRepository
	log        *logrus.Logger
	ticketRepo domain.TicketRepository
}

func NewSimulationUseCase(txManager domain.TransactionManager, repo domain.SimulationRepository, log *logrus.Logger, ticketRepo domain.TicketRepository) *SimulationUseCase {
	return &SimulationUseCase{txManager: txManager, repo: repo, log: log, ticketRepo: ticketRepo}
}

func (u *SimulationUseCase) StartSimulation(ctx context.Context, req *dto.CreateSimulationRequest) (*dto.SimulationResponse, error) {
	id, _ := uuid.NewV7()
	sim := &domain.SimulationRun{
		ID:                 id,
		TotalTickets:       req.TotalTickets,
		BotCount:           req.BotCount,
		BotThrottleSeconds: req.BotThrottleSeconds,
		MaxConcurrent:      req.MaxConcurrent,
		Status:             domain.StatusRunning,
		CreatedAt:          time.Now(),
	}

	if err := sim.Validate(); err != nil {
		return nil, err
	}

	categories := u.setupCategories(sim.TotalTickets, sim.ID)

	err := u.txManager.WithTransaction(ctx, func(ctx context.Context) error {
		if err := u.repo.Create(ctx, sim); err != nil {
			return fmt.Errorf("error insert simulation %f", err)
		}
		if err := u.ticketRepo.BatchCreate(ctx, categories); err != nil {
			return fmt.Errorf("error batch categories %f", err)
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	u.log.Infof("[SimulationUseCase.StartSimulation] Simulation created successfully with ID: %s", sim.ID)
	return &dto.SimulationResponse{
		ID:     sim.ID,
		Status: string(sim.Status),
	}, nil
}

func (u *SimulationUseCase) EndSimulation(ctx context.Context, id uuid.UUID) error {
	err := u.txManager.WithTransaction(ctx, func(txCtx context.Context) error {
		if _, err := u.repo.GetByID(txCtx, id); err != nil {
			return err
		}

		// update status to ended
		if err := u.repo.UpdateStatus(txCtx, id, domain.StatusEnded); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return err
	}

	u.log.Infof("[SimulationUseCase.EndSimulation] End simulation successfully with ID: %s", id)
	return nil
}

func (u *SimulationUseCase) setupCategories(totalTickets int, simID uuid.UUID) []domain.TicketCategories {
	quotaPerCat := totalTickets / len(domain.DefaultCategories)
	remaining := totalTickets % len(domain.DefaultCategories)
	categories := make([]domain.TicketCategories, 0, len(domain.DefaultCategories))

	for i, item := range domain.DefaultCategories {
		quota := quotaPerCat
		// if remaining ticket exists
		// then add to another CAT (+1) based on total/count remaining
		if i < remaining {
			quota++
		}
		categoryID, _ := uuid.NewV7()
		categories = append(categories, domain.TicketCategories{
			ID:        categoryID,
			RunID:     simID,
			Name:      item.Name,
			Price:     item.Price,
			Quota:     quota,
			CreatedAt: time.Now(),
		})
	}

	return categories
}
