package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// User Entity
type User struct {
	ID        uuid.UUID `db:"id" json:"id"`
	Name      string    `db:"name" json:"name"`
	Email     string    `db:"email" json:"email"`
	Role      string    `db:"role" json:"role"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}

// Request DTOs
type CreateUserRequest struct {
	Name  string `json:"name" validate:"required,min=2,max=100"`
	Email string `json:"email" validate:"required,email"`
	Role  string `json:"role" validate:"required,oneof=admin user manager"`
}

type UpdateUserRequest struct {
	Name  string `json:"name" validate:"omitempty,min=2,max=100"`
	Email string `json:"email" validate:"omitempty,email"`
	Role  string `json:"role" validate:"omitempty,oneof=admin user manager"`
}

// Response DTOs
type UserResponse struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func ToUserResponse(user *User) *UserResponse {
	return &UserResponse{
		ID:        user.ID,
		Name:      user.Name,
		Email:     user.Email,
		Role:      user.Role,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}
}

func ToUserResponses(users []User) []*UserResponse {
	responses := make([]*UserResponse, len(users))
	for i, u := range users {
		responses[i] = ToUserResponse(&u)
	}
	return responses
}

// UserRepository defines contract for user persistence storage
type UserRepository interface {
	Create(ctx context.Context, user *User) error
	GetByID(ctx context.Context, id uuid.UUID) (*User, error)
	GetByEmail(ctx context.Context, email string) (*User, error)
	Fetch(ctx context.Context, limit, offset int) ([]User, int64, error)
	Update(ctx context.Context, user *User) error
	Delete(ctx context.Context, id uuid.UUID) error
}

// UserUsecase defines business logic contract for user domain
type UserUsecase interface {
	Create(ctx context.Context, req *CreateUserRequest) (*UserResponse, error)
	GetByID(ctx context.Context, id uuid.UUID) (*UserResponse, error)
	Fetch(ctx context.Context, page, limit int) ([]*UserResponse, int64, error)
	Update(ctx context.Context, id uuid.UUID, req *UpdateUserRequest) (*UserResponse, error)
	Delete(ctx context.Context, id uuid.UUID) error
}
