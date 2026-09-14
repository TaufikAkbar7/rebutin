package handler

import (
	"net/http"

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

	response.SuccessWithList(c, http.StatusOK, "Get categories successfully", tickets)
}
