package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"rebutin/internal/delivery/http/dto"
	"rebutin/internal/delivery/http/response"
	"rebutin/internal/usecase"
)

type SimulationHandler struct {
	simulationUsecase usecase.SimulationUseCase
}

func NewSimulationHandler(simulationUsecase usecase.SimulationUseCase) *SimulationHandler {
	return &SimulationHandler{
		simulationUsecase: simulationUsecase,
	}
}

func (h *SimulationHandler) Start(c *gin.Context) {
	var req dto.CreateSimulationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	// get session from context middleware
	var sessionCookie string
	if val, exists := c.Get("session-cookie"); exists {
		userCookie := val.(string)
		sessionCookie = userCookie
	}

	user, err := h.simulationUsecase.StartSimulation(c.Request.Context(), &req, &sessionCookie)
	if err != nil {
		response.HandleDomainError(c, err)
		return
	}

	response.Success(c, http.StatusCreated, "Simulation created successfully", user)
}

func (h *SimulationHandler) End(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid UUID", nil)
		return
	}

	err = h.simulationUsecase.EndSimulation(c.Request.Context(), id)
	if err != nil {
		response.HandleDomainError(c, err)
		return
	}

	response.Success[any](c, http.StatusOK, "Simulation end successfully", nil)
}
