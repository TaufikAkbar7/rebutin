package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"rebutin/internal/delivery/http/dto"
	"rebutin/internal/delivery/http/response"
	"rebutin/internal/usecase"
	customValidator "rebutin/pkg/validator"
)

type SimulationHandler struct {
	simulationUsecase usecase.SimulationUseCase
	validator         *customValidator.CustomValidator
}

func NewSimulationHandler(simulationUsecase usecase.SimulationUseCase, validator *customValidator.CustomValidator) *SimulationHandler {
	return &SimulationHandler{
		simulationUsecase: simulationUsecase,
		validator:         validator,
	}
}

func (h *SimulationHandler) Start(c *gin.Context) {
	var req dto.CreateSimulationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid JSON payload", nil)
		return
	}

	if valErrs := h.validator.Validate(&req); len(valErrs) > 0 {
		response.ValidationError(c, valErrs)
		return
	}

	user, err := h.simulationUsecase.StartSimulation(c.Request.Context(), &req)
	if err != nil {
		response.HandleDomainError(c, err)
		return
	}

	response.Success(c, http.StatusCreated, "Simulation created successfully", user)
}

func (h *SimulationHandler) End(c *gin.Context) {
	var req dto.CreateSimulationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid JSON payload", nil)
		return
	}

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

	response.Success[any](c, http.StatusCreated, "Simulation end successfully", nil)
}
