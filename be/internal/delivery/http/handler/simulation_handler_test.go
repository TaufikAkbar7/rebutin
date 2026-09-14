package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"rebutin/internal/delivery/http/dto"
	"rebutin/internal/delivery/http/handler"
	"rebutin/internal/delivery/http/response"
	"rebutin/internal/domain"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockSimulationUseCase struct{ mock.Mock }

func (m *MockSimulationUseCase) StartSimulation(ctx context.Context, req *dto.CreateSimulationRequest) (*dto.SimulationResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.SimulationResponse), args.Error(1)
}

func (m *MockSimulationUseCase) EndSimulation(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)

	return args.Error(0)
}

func TestSimulationHandler_Start(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("should return 201 Created on valid input", func(t *testing.T) {
		mockUseCase := new(MockSimulationUseCase)
		h := handler.NewSimulationHandler(mockUseCase)
		r := gin.New()
		r.POST("/api/v1/simulation", h.Start)

		reqBody := dto.CreateSimulationRequest{
			TotalTickets:       100,
			BotCount:           10,
			BotThrottleSeconds: 0,
			MaxConcurrent:      1,
		}
		bodyBytes, _ := json.Marshal(reqBody)

		id, _ := uuid.NewV7()
		mockUseCase.On("StartSimulation", mock.Anything, &reqBody).Return(&dto.SimulationResponse{
			ID:     id,
			Status: "running",
		}, nil)

		req, _ := http.NewRequest(http.MethodPost, "/api/v1/simulation", bytes.NewBuffer(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
		assert.Contains(t, w.Body.String(), "Simulation created successfully")
		assert.Contains(t, w.Body.String(), domain.StatusRunning)
		mockUseCase.AssertExpectations(t)
	})

	t.Run("should return 400 bad request on validation failure", func(t *testing.T) {
		mockUseCase := new(MockSimulationUseCase)
		h := handler.NewSimulationHandler(mockUseCase)
		r := gin.New()
		r.POST("/api/v1/simulation", h.Start)

		invalidPayload := bytes.NewBufferString(`{"total_tickets": "number"}`)

		req, _ := http.NewRequest(http.MethodPost, "/api/v1/simulation", invalidPayload)
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "Validation failed")
		mockUseCase.AssertNotCalled(t, "StartSimulation", mock.Anything, mock.Anything)
	})

	t.Run("should return 400 bad request on business validation failure", func(t *testing.T) {
		mockUseCase := new(MockSimulationUseCase)
		h := handler.NewSimulationHandler(mockUseCase)
		r := gin.New()
		r.POST("/api/v1/simulation", h.Start)

		reqBody := dto.CreateSimulationRequest{
			TotalTickets:  100,
			MaxConcurrent: 20,
			BotCount:      10,
		}
		bodyBytes, _ := json.Marshal(reqBody)

		bizErr := &domain.ValidationError{Errors: make(map[string]string)}
		bizErr.Add("max_concurrent", "max concurrent cannot exceed total bot count")

		mockUseCase.On("StartSimulation", mock.Anything, mock.MatchedBy(func(req *dto.CreateSimulationRequest) bool {
			return req != nil && req.MaxConcurrent > req.BotCount
		})).Return(nil, bizErr)

		req, _ := http.NewRequest(http.MethodPost, "/api/v1/simulation", bytes.NewBuffer(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		var response response.APIErrorResponse[map[string]string]
		responseBody, err := io.ReadAll(w.Body)

		err = json.Unmarshal(responseBody, &response)
		assert.NoError(t, err)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, response.Message, "Business validation failed")
		assert.Len(t, response.Errors, 1)
		mockUseCase.AssertExpectations(t)
	})
}

func TestSimulationHandler_End(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("should return 200 Created on valid input", func(t *testing.T) {
		mockUseCase := new(MockSimulationUseCase)
		h := handler.NewSimulationHandler(mockUseCase)
		r := gin.New()
		r.PATCH("/api/v1/simulation/end/:id", h.End)

		targetID, _ := uuid.NewV7()

		mockUseCase.On("EndSimulation", mock.Anything, targetID).Return(nil)

		req, _ := http.NewRequest(http.MethodPatch, "/api/v1/simulation/end/"+targetID.String(), nil)
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "Simulation end successfully")
		mockUseCase.AssertExpectations(t)
	})

	t.Run("should return 400 bad request when UUID param invalid", func(t *testing.T) {
		mockUseCase := new(MockSimulationUseCase)
		h := handler.NewSimulationHandler(mockUseCase)
		r := gin.New()
		r.POST("/api/v1/simulation/end/:id", h.End)

		req, _ := http.NewRequest(http.MethodPost, "/api/v1/simulation/end/not-uuid", nil)
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "Invalid UUID")
		mockUseCase.AssertNotCalled(t, "EndSimulation", mock.Anything, mock.Anything)
	})
}
