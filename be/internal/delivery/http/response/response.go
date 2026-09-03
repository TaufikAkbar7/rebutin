package response

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"rebutin/internal/domain"
)

type Meta struct {
	Page       int   `json:"page,omitempty"`
	Limit      int   `json:"limit,omitempty"`
	TotalItems int64 `json:"total_items,omitempty"`
	TotalPages int   `json:"total_pages,omitempty"`
}

type APIResponseArray[T any] struct {
	Message string `json:"message"`
	Data    []T    `json:"data,omitempty"`
	Meta    *Meta  `json:"meta,omitempty"`
}

type APIResponse[T any] struct {
	Message string `json:"message"`
	Data    T      `json:"data,omitempty"`
}

type APIErrorResponse[T any] struct {
	Message string `json:"message"`
	Errors  T      `json:"errors,omitempty"`
}

func Success[T any](c *gin.Context, statusCode int, message string, data T) {
	c.JSON(statusCode, APIResponse[T]{
		Message: message,
		Data:    data,
	})
}

func SuccessWithMeta[T any](c *gin.Context, statusCode int, message string, data []T, page, limit int, totalItems int64) {
	totalPages := 0
	if limit > 0 {
		totalPages = int((totalItems + int64(limit) - 1) / int64(limit))
	}

	c.JSON(statusCode, APIResponseArray[T]{
		Message: message,
		Data:    data,
		Meta: &Meta{
			Page:       page,
			Limit:      limit,
			TotalItems: totalItems,
			TotalPages: totalPages,
		},
	})
}

func Error(c *gin.Context, statusCode int, message string, errs error) {
	c.JSON(statusCode, APIErrorResponse[error]{
		Message: message,
		Errors:  errs,
	})
}

func ValidationError(c *gin.Context, validationErrors map[string]string) {
	c.JSON(http.StatusBadRequest, APIErrorResponse[map[string]string]{
		Message: "Validation failed",
		Errors:  validationErrors,
	})
}

func HandleDomainError(c *gin.Context, err error) {
	var domainValidationErr *domain.ValidationError
	if errors.As(err, &domainValidationErr) {
		c.JSON(http.StatusBadRequest, APIErrorResponse[map[string]string]{
			Message: "Business validation failed",
			Errors:  domainValidationErr.Errors,
		})
		return
	}

	switch {
	case errors.Is(err, domain.ErrNotFound):
		Error(c, http.StatusNotFound, "Resource not found", nil)
	case errors.Is(err, domain.ErrAlreadyExists):
		Error(c, http.StatusConflict, "Resource already exists", nil)
	case errors.Is(err, domain.ErrInvalidInput):
		Error(c, http.StatusBadRequest, "Invalid input data", nil)
	case errors.Is(err, domain.ErrUnauthorized):
		Error(c, http.StatusUnauthorized, "Unauthorized", nil)
	case errors.Is(err, domain.ErrForbidden):
		Error(c, http.StatusForbidden, "Forbidden action", nil)
	default:
		Error(c, http.StatusInternalServerError, "An internal server error occurred", nil)
	}
}
