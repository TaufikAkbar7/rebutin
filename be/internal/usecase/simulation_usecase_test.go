package usecase_test

import (
	"context"
	"errors"
	"rebutin/internal/delivery/http/dto"
	"rebutin/internal/domain"
	"rebutin/internal/usecase"
	"rebutin/pkg/testutil"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type MockSimulationRepo struct{ mock.Mock }

func (m *MockSimulationRepo) Create(ctx context.Context, sim *domain.SimulationRun) error {
	args := m.Called(ctx, sim)
	return args.Error(0)
}

func (m *MockSimulationRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.SimulationRun, error) {
	args := m.Called(ctx, id)

	val := args.Get(0)
	if val == nil {
		return nil, args.Error(1)
	}

	return val.(*domain.SimulationRun), args.Error(1)
}

func (m *MockSimulationRepo) UpdateStatus(ctx context.Context, id uuid.UUID, status domain.SimulationStatus) error {
	args := m.Called(ctx, id, status)
	return args.Error(0)
}

func TestSimulationUseCase_StartSimulation_Success(t *testing.T) {
	mockTxManager := new(MockTxManager)
	mockSimRepo := new(MockSimulationRepo)
	mockTicketRepo := new(MockTicketRepo)
	logger, _ := testutil.SetupLogger(t)

	uc := usecase.NewSimulationUseCase(mockTxManager, mockSimRepo, logger, mockTicketRepo)

	req := dto.CreateSimulationRequest{
		TotalTickets:       100,
		BotCount:           10,
		BotThrottleSeconds: 1,
		MaxConcurrent:      5,
	}

	mockSimRepo.On("Create", mock.Anything, mock.AnythingOfType("*domain.SimulationRun")).Return(nil)
	mockTicketRepo.On("BatchCreate", mock.Anything, mock.AnythingOfType("[]domain.TicketCategories")).Return(nil)
	mockTxManager.On("WithTransaction", mock.Anything, mock.Anything).Return(nil)

	res, err := uc.StartSimulation(context.Background(), &req)

	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.Equal(t, string(domain.StatusRunning), res.Status)

	mockSimRepo.AssertExpectations(t)
	mockTicketRepo.AssertExpectations(t)
	mockTxManager.AssertExpectations(t)
}

func TestSimulationUseCase_QoutaDistribution(t *testing.T) {
	mockTxManager := new(MockTxManager)
	mockSimRepo := new(MockSimulationRepo)
	mockTicketRepo := new(MockTicketRepo)
	logger, _ := testutil.SetupLogger(t)

	uc := usecase.NewSimulationUseCase(mockTxManager, mockSimRepo, logger, mockTicketRepo)

	req := dto.CreateSimulationRequest{
		TotalTickets:       100,
		BotCount:           10,
		BotThrottleSeconds: 1,
		MaxConcurrent:      5,
	}

	mockSimRepo.On("Create", mock.Anything, mock.AnythingOfType("*domain.SimulationRun")).Return(nil)
	mockTicketRepo.On("BatchCreate", mock.Anything, mock.MatchedBy(func(tickets []domain.TicketCategories) bool {
		if len(tickets) != 3 {
			return false
		}

		// expected qouta distribution
		// CAT 1 34 | CAT 2 33 | CAT 3 33
		return tickets[0].Quota == 34 && tickets[1].Quota == 33 && tickets[2].Quota == 33
	})).Return(nil)
	mockTxManager.On("WithTransaction", mock.Anything, mock.Anything).Return(nil)

	res, err := uc.StartSimulation(context.Background(), &req)

	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.Equal(t, string(domain.StatusRunning), res.Status)

	mockSimRepo.AssertExpectations(t)
	mockTicketRepo.AssertExpectations(t)
	mockTxManager.AssertExpectations(t)
}

func TestSimulationUseCase_StartSimulation_Failures(t *testing.T) {
	t.Run("should fail fast and not call DB when domain validation fails", func(t *testing.T) {
		mockTx := new(MockTxManager)
		mockSimRepo := new(MockSimulationRepo)
		mockTicketRepo := new(MockTicketRepo)
		logger, _ := testutil.SetupLogger(t)

		uc := usecase.NewSimulationUseCase(mockTx, mockSimRepo, logger, mockTicketRepo)

		req := dto.CreateSimulationRequest{
			TotalTickets: 0,
			BotCount:     10,
		}

		res, err := uc.StartSimulation(context.Background(), &req)

		assert.Error(t, err)
		assert.Nil(t, res)

		// extract struct error
		var valErr *domain.ValidationError
		ok := errors.As(err, &valErr)
		require.True(t, ok, "error should be of type *domain.ValidationErrors")

		// expected custom error
		assert.Len(t, valErr.Errors, 2)
		assert.Equal(t, "total tickets must be greater than 0", valErr.Errors["total_tickets"])
		assert.Equal(t, "max concurrent must be at least 1", valErr.Errors["max_concurrent"])

		// expected layer repo not called
		mockTx.AssertNotCalled(t, "WithTransaction", mock.Anything, mock.Anything)
		mockSimRepo.AssertNotCalled(t, "Create", mock.Anything, mock.Anything)
		mockTicketRepo.AssertNotCalled(t, "BatchCreate", mock.Anything, mock.Anything)
	})

	t.Run("should abort and rollback when simRepo Create fails", func(t *testing.T) {
		mockTx := new(MockTxManager)
		mockSimRepo := new(MockSimulationRepo)
		mockTicketRepo := new(MockTicketRepo)
		logger, _ := testutil.SetupLogger(t)

		uc := usecase.NewSimulationUseCase(mockTx, mockSimRepo, logger, mockTicketRepo)

		req := dto.CreateSimulationRequest{
			TotalTickets:       100,
			BotCount:           10,
			BotThrottleSeconds: 0,
			MaxConcurrent:      1,
		}

		dbErr := errors.New("unique constraint violation")
		mockSimRepo.On("Create", mock.Anything, mock.Anything).Return(dbErr)
		mockTx.On("WithTransaction", mock.Anything, mock.Anything).Return(nil)

		res, err := uc.StartSimulation(context.Background(), &req)

		assert.Error(t, err)
		assert.Nil(t, res)
		assert.ErrorContains(t, err, dbErr.Error())

		mockTicketRepo.AssertNotCalled(t, "BatchCreate", mock.Anything, mock.Anything)
	})

	t.Run("should abort and rollback when ticketRepo BatchCreate fails", func(t *testing.T) {
		mockTx := new(MockTxManager)
		mockSimRepo := new(MockSimulationRepo)
		mockTicketRepo := new(MockTicketRepo)
		logger, _ := testutil.SetupLogger(t)

		uc := usecase.NewSimulationUseCase(mockTx, mockSimRepo, logger, mockTicketRepo)

		req := dto.CreateSimulationRequest{
			TotalTickets:       100,
			BotCount:           10,
			BotThrottleSeconds: 0,
			MaxConcurrent:      1,
		}

		dbErr := errors.New("unique constraint violation")
		mockSimRepo.On("Create", mock.Anything, mock.Anything).Return(nil)
		mockTicketRepo.On("BatchCreate", mock.Anything, mock.Anything).Return(dbErr)
		mockTx.On("WithTransaction", mock.Anything, mock.Anything).Return(nil)

		res, err := uc.StartSimulation(context.Background(), &req)

		assert.Error(t, err)
		assert.Nil(t, res)
		assert.ErrorContains(t, err, dbErr.Error())
	})

	t.Run("should return error when txManager fail to begin trx", func(t *testing.T) {
		mockTx := new(MockTxManager)
		mockSimRepo := new(MockSimulationRepo)
		mockTicketRepo := new(MockTicketRepo)
		logger, _ := testutil.SetupLogger(t)

		uc := usecase.NewSimulationUseCase(mockTx, mockSimRepo, logger, mockTicketRepo)

		req := dto.CreateSimulationRequest{
			TotalTickets:       100,
			BotCount:           10,
			BotThrottleSeconds: 0,
			MaxConcurrent:      1,
		}

		dbErr := errors.New("db not found")
		mockTx.On("WithTransaction", mock.Anything, mock.Anything).Return(dbErr)

		res, err := uc.StartSimulation(context.Background(), &req)

		assert.Error(t, err)
		assert.Nil(t, res)
		assert.ErrorContains(t, err, dbErr.Error())

		mockSimRepo.AssertNotCalled(t, "Create", mock.Anything, mock.Anything)
		mockTicketRepo.AssertNotCalled(t, "BatchCreate", mock.Anything, mock.Anything)
	})

	t.Run("should abort and rollback when txManager fail to commit trx", func(t *testing.T) {
		mockTx := new(MockTxManager)
		mockSimRepo := new(MockSimulationRepo)
		mockTicketRepo := new(MockTicketRepo)
		logger, _ := testutil.SetupLogger(t)

		uc := usecase.NewSimulationUseCase(mockTx, mockSimRepo, logger, mockTicketRepo)

		req := dto.CreateSimulationRequest{
			TotalTickets:       100,
			BotCount:           10,
			BotThrottleSeconds: 0,
			MaxConcurrent:      1,
		}

		dbErr := errors.New("commit fail cause deadlock")
		mockSimRepo.On("Create", mock.Anything, mock.Anything).Return(nil)
		mockTicketRepo.On("BatchCreate", mock.Anything, mock.Anything).Return(nil)
		mockTx.On("WithTransaction", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
			// get closure from ticketRepo
			fn := args.Get(1).(domain.AtomicFunc)
			_ = fn(context.Background())
		}).Return(dbErr)

		res, err := uc.StartSimulation(context.Background(), &req)

		assert.Error(t, err)
		assert.Nil(t, res)
		assert.ErrorContains(t, err, dbErr.Error())

		mockSimRepo.AssertCalled(t, "Create", mock.Anything, mock.Anything)
		mockTicketRepo.AssertCalled(t, "BatchCreate", mock.Anything, mock.Anything)
	})
}

func TestSimulationUseCase_EndSimulation_Success(t *testing.T) {
	mockTxManager := new(MockTxManager)
	mockSimRepo := new(MockSimulationRepo)
	mockTicketRepo := new(MockTicketRepo)
	logger, _ := testutil.SetupLogger(t)

	uc := usecase.NewSimulationUseCase(mockTxManager, mockSimRepo, logger, mockTicketRepo)

	targetID, _ := uuid.NewV7()

	data := domain.SimulationRun{
		ID:     targetID,
		Status: domain.StatusRunning,
	}

	mockSimRepo.On("GetByID", mock.Anything, targetID).Return(&data, nil)
	mockSimRepo.On("UpdateStatus", mock.Anything, targetID, domain.StatusEnded).Return(nil)

	err := uc.EndSimulation(context.Background(), targetID)
	assert.NoError(t, err)

	mockSimRepo.AssertExpectations(t)
}

func TestSimulationUseCase_EndSimulation_Failures(t *testing.T) {
	t.Run("should error when status already ended", func(t *testing.T) {
		mockTxManager := new(MockTxManager)
		mockSimRepo := new(MockSimulationRepo)
		mockTicketRepo := new(MockTicketRepo)
		logger, _ := testutil.SetupLogger(t)

		uc := usecase.NewSimulationUseCase(mockTxManager, mockSimRepo, logger, mockTicketRepo)

		targetID, _ := uuid.NewV7()

		data := domain.SimulationRun{
			ID:     targetID,
			Status: domain.StatusEnded,
		}
		mockSimRepo.On("GetByID", mock.Anything, targetID).Return(&data, nil)

		err := uc.EndSimulation(context.Background(), targetID)

		// extract struct error
		var valErr *domain.ValidationError
		ok := errors.As(err, &valErr)
		require.True(t, ok, "error should be of type *domain.ValidationErrors")

		assert.Len(t, valErr.Errors, 1)
		assert.Equal(t, domain.ErrSimulationAlreadyStopped.Error(), valErr.Errors["status"])

		mockSimRepo.AssertNotCalled(t, "UpdateStatus")
		mockSimRepo.AssertExpectations(t)
	})

	t.Run("should error when UpdateStatus fails", func(t *testing.T) {
		mockTxManager := new(MockTxManager)
		mockSimRepo := new(MockSimulationRepo)
		mockTicketRepo := new(MockTicketRepo)
		logger, _ := testutil.SetupLogger(t)

		uc := usecase.NewSimulationUseCase(mockTxManager, mockSimRepo, logger, mockTicketRepo)

		targetID, _ := uuid.NewV7()

		dbErr := errors.New("unique constraint violation")
		data := domain.SimulationRun{
			ID:     targetID,
			Status: domain.StatusRunning,
		}
		mockSimRepo.On("GetByID", mock.Anything, targetID).Return(&data, nil)
		mockSimRepo.On("UpdateStatus", mock.Anything, targetID, domain.StatusEnded).Return(dbErr)

		err := uc.EndSimulation(context.Background(), targetID)

		assert.ErrorContains(t, err, dbErr.Error())
		mockSimRepo.AssertCalled(t, "GetByID", mock.Anything, targetID)
		mockSimRepo.AssertCalled(t, "UpdateStatus", mock.Anything, targetID, domain.StatusEnded)
	})
}
