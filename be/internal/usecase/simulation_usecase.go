package usecase

import (
	"context"
	"time"

	"rebutin/internal/delivery/http/dto"
	"rebutin/internal/domain"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

type SimulationUseCase struct {
	txManager domain.TransactionManager
	repo      domain.SimulationRepository
	log       *logrus.Logger
}

func NewSimulationUseCase(txManager domain.TransactionManager, repo domain.SimulationRepository, log *logrus.Logger) *SimulationUseCase {
	return &SimulationUseCase{txManager: txManager, repo: repo, log: log}
}

func (u *SimulationUseCase) StartSimulation(ctx context.Context, req *dto.CreateSimulationRequest) (*dto.SimulationResponse, error) {
	uuid, _ := uuid.NewV7()
	sim := &domain.SimulationRun{
		ID:                 uuid,
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

	if err := u.repo.Create(ctx, sim); err != nil {
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
