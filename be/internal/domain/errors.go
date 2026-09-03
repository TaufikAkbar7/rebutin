package domain

import (
	"errors"
	"fmt"
	"strings"
)

var (
	ErrNotFound         = errors.New("resource not found")
	ErrAlreadyExists    = errors.New("resource already exists")
	ErrInvalidInput     = errors.New("invalid request input")
	ErrUnauthorized     = errors.New("unauthorized access")
	ErrForbidden        = errors.New("action forbidden")
	ErrInternalServer   = errors.New("internal server error")
	ErrDatabaseOpFailed = errors.New("database operation failed")
)

type ValidationError struct {
	Errors map[string]string
}

func (v *ValidationError) Error() string {
	var b strings.Builder
	b.WriteString("validation failed: ")
	for field, msg := range v.Errors {
		b.WriteString(fmt.Sprintf("[%s: %s] ", field, msg))
	}
	return strings.TrimSpace(b.String())
}

func (v *ValidationError) HasErrors() bool {
	return len(v.Errors) > 0
}

func (v *ValidationError) Add(field, msg string) {
	v.Errors[field] = msg
}
