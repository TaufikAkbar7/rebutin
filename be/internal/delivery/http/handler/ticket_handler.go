package handler

import (
	"encoding/json"
	"net/http"

	"github.com/sirupsen/logrus"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"rebutin/internal/delivery/http/response"
	"rebutin/internal/usecase"
)

type TicketHandler struct {
	ticketUsecase usecase.TicketUseCase
}

func NewTicketHandler(ticketUsecase usecase.TicketUseCase) *TicketHandler {
	return &TicketHandler{
		ticketUsecase: ticketUsecase,
	}
}

func (h *TicketHandler) FindCategories(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid UUID", nil)
		return
	}

	tickets, err := h.ticketUsecase.GetTicketBySimulation(c.Request.Context(), id)
	if err != nil {
		response.HandleDomainError(c, err)
		return
	}

	groupDataJSON, err := json.MarshalIndent(tickets, "", "  ")
	if err != nil {
		logrus.Printf("[ERROR] Failed to marshal groupData to JSON: %v", err)
	} else {
		logrus.Printf("[DEBUG] Aggregated groupData:\n%s", string(groupDataJSON))
	}

	response.SuccessWithList(c, http.StatusOK, "Get categories successfully", tickets)
}
