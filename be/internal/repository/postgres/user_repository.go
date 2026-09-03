package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"github.com/sirupsen/logrus"

	"rebutin/internal/domain"
)

type userRepository struct {
	db  *sqlx.DB
	log *logrus.Logger
}

func NewUserRepository(db *sqlx.DB, log *logrus.Logger) domain.UserRepository {
	return &userRepository{
		db:  db,
		log: log,
	}
}

func (r *userRepository) Create(ctx context.Context, user *domain.User) error {
	query := `
		INSERT INTO users (id, name, email, role, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	_, err := r.db.ExecContext(ctx, query,
		user.ID,
		user.Name,
		user.Email,
		user.Role,
		user.CreatedAt,
		user.UpdatedAt,
	)

	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == "23505" { // Unique violation
			return domain.ErrAlreadyExists
		}
		r.log.Errorf("[UserRepository.Create] Error creating user: %v", err)
		return fmt.Errorf("%w: %v", domain.ErrDatabaseOpFailed, err)
	}

	return nil
}

func (r *userRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	var user domain.User
	query := `SELECT id, name, email, role, created_at, updated_at FROM users WHERE id = $1`

	err := r.db.GetContext(ctx, &user, query, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		r.log.Errorf("[UserRepository.GetByID] Error getting user %s: %v", id, err)
		return nil, fmt.Errorf("%w: %v", domain.ErrDatabaseOpFailed, err)
	}

	return &user, nil
}

func (r *userRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	var user domain.User
	query := `SELECT id, name, email, role, created_at, updated_at FROM users WHERE email = $1`

	err := r.db.GetContext(ctx, &user, query, strings.ToLower(email))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		r.log.Errorf("[UserRepository.GetByEmail] Error getting user by email %s: %v", email, err)
		return nil, fmt.Errorf("%w: %v", domain.ErrDatabaseOpFailed, err)
	}

	return &user, nil
}

func (r *userRepository) Fetch(ctx context.Context, limit, offset int) ([]domain.User, int64, error) {
	var users []domain.User
	var total int64

	countQuery := `SELECT COUNT(*) FROM users`
	if err := r.db.GetContext(ctx, &total, countQuery); err != nil {
		r.log.Errorf("[UserRepository.Fetch] Error counting users: %v", err)
		return nil, 0, fmt.Errorf("%w: %v", domain.ErrDatabaseOpFailed, err)
	}

	query := `
		SELECT id, name, email, role, created_at, updated_at 
		FROM users 
		ORDER BY created_at DESC 
		LIMIT $1 OFFSET $2
	`

	if err := r.db.SelectContext(ctx, &users, query, limit, offset); err != nil {
		r.log.Errorf("[UserRepository.Fetch] Error fetching users: %v", err)
		return nil, 0, fmt.Errorf("%w: %v", domain.ErrDatabaseOpFailed, err)
	}

	if users == nil {
		users = []domain.User{}
	}

	return users, total, nil
}

func (r *userRepository) Update(ctx context.Context, user *domain.User) error {
	query := `
		UPDATE users 
		SET name = $1, email = $2, role = $3, updated_at = $4 
		WHERE id = $5
	`
	res, err := r.db.ExecContext(ctx, query,
		user.Name,
		user.Email,
		user.Role,
		user.UpdatedAt,
		user.ID,
	)

	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == "23505" {
			return domain.ErrAlreadyExists
		}
		r.log.Errorf("[UserRepository.Update] Error updating user %s: %v", user.ID, err)
		return fmt.Errorf("%w: %v", domain.ErrDatabaseOpFailed, err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return domain.ErrNotFound
	}

	return nil
}

func (r *userRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM users WHERE id = $1`
	res, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		r.log.Errorf("[UserRepository.Delete] Error deleting user %s: %v", id, err)
		return fmt.Errorf("%w: %v", domain.ErrDatabaseOpFailed, err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return domain.ErrNotFound
	}

	return nil
}
