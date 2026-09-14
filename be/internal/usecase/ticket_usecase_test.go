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
)

type MockTicketRepo struct{ mock.Mock }

func (m *MockTicketRepo) BatchCreate(ctx context.Context, payload []domain.TicketCategories) error {
	args := m.Called(ctx, payload)
	return args.Error(0)
}

func (m *MockTicketRepo) GetByRunID(ctx context.Context, runID uuid.UUID) ([]domain.TicketCategories, error) {
	args := m.Called(ctx, runID)

	val := args.Get(0)
	if val == nil {
		return nil, args.Error(1)
	}

	return val.([]domain.TicketCategories), args.Error(1)
}

func TestTicketUseCase_GetTicketBySimulation_Success(t *testing.T) {
	mockTicketRepo := new(MockTicketRepo)
	logger, _ := testutil.SetupLogger(t)

	uc := usecase.NewTicketUseCase(logger, mockTicketRepo)

	targetID, _ := uuid.NewV7()

	datas := make([]domain.TicketCategories, 0, len(domain.DefaultCategories))
	expectedRes := make([]dto.TicketResponse, 0, len(domain.DefaultCategories))
	for _, item := range domain.DefaultCategories {
		datas = append(datas, domain.TicketCategories{
			ID:    targetID,
			RunID: targetID,
			Name:  item.Name,
			Quota: 1,
			Price: item.Price,
		})
		expectedRes = append(expectedRes, dto.TicketResponse{
			ID:    targetID,
			Name:  item.Name,
			Quota: 1,
			Price: item.Price,
		})
	}

	mockTicketRepo.On("GetByRunID", mock.Anything, targetID).Return(datas, nil)

	res, err := uc.GetTicketBySimulation(context.Background(), targetID)
	assert.NoError(t, err)

	mockTicketRepo.AssertExpectations(t)
	mockTicketRepo.AssertCalled(t, "GetByRunID", mock.Anything, targetID)
	assert.Len(t, res, 3)
	assert.Equal(t, expectedRes, res)
}

func TestTicketUseCase_GetTicketBySimulation_Failures(t *testing.T) {
	mockTicketRepo := new(MockTicketRepo)
	logger, _ := testutil.SetupLogger(t)

	uc := usecase.NewTicketUseCase(logger, mockTicketRepo)

	targetID, _ := uuid.NewV7()

	dbErr := errors.New("unique constraint violation")
	mockTicketRepo.On("GetByRunID", mock.Anything, targetID).Return(nil, dbErr)

	res, err := uc.GetTicketBySimulation(context.Background(), targetID)

	assert.Error(t, err)
	assert.Nil(t, res)
	assert.ErrorContains(t, err, dbErr.Error())
	mockTicketRepo.AssertCalled(t, "GetByRunID", mock.Anything, targetID)
}
