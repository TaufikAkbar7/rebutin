package validator

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
)

type CustomValidator struct {
	validate *validator.Validate
}

func NewCustomValidator() *CustomValidator {
	v := validator.New()

	v.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})

	return &CustomValidator{validate: v}
}

func (cv *CustomValidator) Validate(i interface{}) map[string]string {
	fieldErrors := make(map[string]string)

	err := cv.validate.Struct(i)
	if err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			for _, fieldError := range validationErrors {
				fieldErrors[fieldError.Field()] = formatErrorMessage(fieldError)
			}
		}
	}

	return fieldErrors
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
