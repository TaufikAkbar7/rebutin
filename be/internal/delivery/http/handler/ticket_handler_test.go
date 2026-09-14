package handler_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"rebutin/internal/delivery/http/dto"
	"rebutin/internal/delivery/http/handler"
	"rebutin/internal/delivery/http/response"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockTicketUseCase struct{ mock.Mock }

func (m *MockTicketUseCase) GetTicketBySimulation(ctx context.Context, id uuid.UUID) ([]dto.TicketResponse, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]dto.TicketResponse), args.Error(1)
}

func TestTicketHandler_FindCategories(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("should return 200 with data tickets", func(t *testing.T) {
		mockUseCase := new(MockTicketUseCase)
		h := handler.NewTicketHandler(mockUseCase)
		r := gin.New()
		r.GET("/api/v1/categories/:id", h.FindCategories)

		targetID, _ := uuid.NewV7()

		datas := make([]dto.TicketResponse, 1)
		datas = append(datas, dto.TicketResponse{
			ID:    targetID,
			Name:  "CAT 1",
			Quota: 30,
			Price: 10000,
		})

		mockUseCase.On("GetTicketBySimulation", mock.Anything, targetID).Return(datas, nil)

		req, _ := http.NewRequest(http.MethodGet, "/api/v1/categories/"+targetID.String(), nil)
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		var response response.APIResponseArray[dto.TicketResponse]
		responseBody, err := io.ReadAll(w.Body)

		err = json.Unmarshal(responseBody, &response)
		assert.NoError(t, err)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, response.Message, "Get categories successfully")
		assert.Equal(t, response.Data, datas)
		assert.Len(t, datas, 1)
		mockUseCase.AssertExpectations(t)
	})

	t.Run("should return 400 bad request when UUID param invalid", func(t *testing.T) {
		mockUseCase := new(MockTicketUseCase)
		h := handler.NewTicketHandler(mockUseCase)
		r := gin.New()
		r.GET("/api/v1/categories/:id", h.FindCategories)

		req, _ := http.NewRequest(http.MethodGet, "/api/v1/categories/not-uuid", nil)
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "Invalid UUID")
		mockUseCase.AssertNotCalled(t, "GetTicketBySimulation", mock.Anything, mock.Anything)
	})
}
