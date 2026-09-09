package response

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"

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
	Data    []T    `json:"data"`
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

func SuccessWithList[T any](c *gin.Context, statusCode int, message string, data []T) {
	c.JSON(statusCode, APIResponseArray[T]{
		Message: message,
		Data:    data,
	})
}

func Error(c *gin.Context, statusCode int, message string, errs error) {
	c.JSON(statusCode, APIErrorResponse[error]{
		Message: message,
		Errors:  errs,
	})
}

func ValidationError(c *gin.Context, err error) {
	fieldErrors := make(map[string]string)
	if validationErrors, ok := err.(validator.ValidationErrors); ok {
		for _, fieldError := range validationErrors {
			fieldErrors[fieldError.Field()] = formatErrorMessage(fieldError)
		}
	}

	c.JSON(http.StatusBadRequest, APIErrorResponse[map[string]string]{
		Message: "Validation failed",
		Errors:  fieldErrors,
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

func formatErrorMessage(err validator.FieldError) string {
	switch err.Tag() {
	case "required":
		return fmt.Sprintf("%s field is required", err.Field())
	case "email":
		return fmt.Sprintf("%s must be a valid email address", err.Field())
	case "min":
		return fmt.Sprintf("%s must be at least %s characters long", err.Field(), err.Param())
	case "max":
		return fmt.Sprintf("%s must be at most %s characters long", err.Field(), err.Param())
	case "uuid":
		return fmt.Sprintf("%s must be a valid UUID", err.Field())
	default:
		return fmt.Sprintf("%s is invalid", err.Field())
	}
}
