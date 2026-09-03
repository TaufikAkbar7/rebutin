package usecase

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"

	"rebutin/internal/domain"
)

type userUsecase struct {
	userRepo domain.UserRepository
	log      *logrus.Logger
}

func NewUserUsecase(userRepo domain.UserRepository, log *logrus.Logger) domain.UserUsecase {
	return &userUsecase{
		userRepo: userRepo,
		log:      log,
	}
}

func (u *userUsecase) Create(ctx context.Context, req *domain.CreateUserRequest) (*domain.UserResponse, error) {
	req.Email = strings.ToLower(strings.TrimSpace(req.Email))
	req.Name = strings.TrimSpace(req.Name)

	// Check if user with same email exists
	existingUser, err := u.userRepo.GetByEmail(ctx, req.Email)
	if err != nil && !errors.Is(err, domain.ErrNotFound) {
		return nil, err
	}
	if existingUser != nil {
		return nil, domain.ErrAlreadyExists
	}

	now := time.Now().UTC()
	user := &domain.User{
		ID:        uuid.New(),
		Name:      req.Name,
		Email:     req.Email,
		Role:      req.Role,
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := u.userRepo.Create(ctx, user); err != nil {
		return nil, err
	}

	u.log.Infof("[UserUsecase.Create] User created successfully with ID: %s", user.ID)
	return domain.ToUserResponse(user), nil
}

func (u *userUsecase) GetByID(ctx context.Context, id uuid.UUID) (*domain.UserResponse, error) {
	user, err := u.userRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return domain.ToUserResponse(user), nil
}

func (u *userUsecase) Fetch(ctx context.Context, page, limit int) ([]*domain.UserResponse, int64, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	offset := (page - 1) * limit
	users, total, err := u.userRepo.Fetch(ctx, limit, offset)
	if err != nil {
		return nil, 0, err
	}

	return domain.ToUserResponses(users), total, nil
}

func (u *userUsecase) Update(ctx context.Context, id uuid.UUID, req *domain.UpdateUserRequest) (*domain.UserResponse, error) {
	user, err := u.userRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if req.Name != "" {
		user.Name = strings.TrimSpace(req.Name)
	}

	if req.Email != "" {
		newEmail := strings.ToLower(strings.TrimSpace(req.Email))
		if newEmail != user.Email {
			existing, err := u.userRepo.GetByEmail(ctx, newEmail)
			if err != nil && !errors.Is(err, domain.ErrNotFound) {
				return nil, err
			}
			if existing != nil {
				return nil, domain.ErrAlreadyExists
			}
			user.Email = newEmail
		}
	}

	if req.Role != "" {
		user.Role = req.Role
	}

	user.UpdatedAt = time.Now().UTC()

	if err := u.userRepo.Update(ctx, user); err != nil {
		return nil, err
	}

	u.log.Infof("[UserUsecase.Update] User updated successfully with ID: %s", id)
	return domain.ToUserResponse(user), nil
}

func (u *userUsecase) Delete(ctx context.Context, id uuid.UUID) error {
	if _, err := u.userRepo.GetByID(ctx, id); err != nil {
		return err
	}

	if err := u.userRepo.Delete(ctx, id); err != nil {
		return err
	}

	u.log.Infof("[UserUsecase.Delete] User deleted successfully with ID: %s", id)
	return nil
}
