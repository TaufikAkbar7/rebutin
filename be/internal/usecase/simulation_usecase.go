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

type SimulationUseCase interface {
	StartSimulation(ctx context.Context, req *dto.CreateSimulationRequest) (*dto.SimulationResponse, error)
	EndSimulation(ctx context.Context, id uuid.UUID) error
}

type simulationUseCase struct {
	txManager  domain.TransactionManager
	repo       domain.SimulationRepository
	log        *logrus.Logger
	ticketRepo domain.TicketRepository
}

func NewSimulationUseCase(txManager domain.TransactionManager, repo domain.SimulationRepository, log *logrus.Logger, ticketRepo domain.TicketRepository) SimulationUseCase {
	return &simulationUseCase{txManager: txManager, repo: repo, log: log, ticketRepo: ticketRepo}
}

func (u *simulationUseCase) StartSimulation(ctx context.Context, req *dto.CreateSimulationRequest) (*dto.SimulationResponse, error) {
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

func (u *simulationUseCase) EndSimulation(ctx context.Context, id uuid.UUID) error {
	data, err := u.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// check if status already stop
	if err := data.Stop(); err != nil {
		return err
	}

	// update status to ended
	if err := u.repo.UpdateStatus(ctx, id, domain.StatusEnded); err != nil {
		return err
	}

	u.log.Infof("[SimulationUseCase.EndSimulation] End simulation successfully with ID: %s", id)
	return nil
}

func (u *simulationUseCase) setupCategories(totalTickets int, simID uuid.UUID) []domain.TicketCategories {
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
