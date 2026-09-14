package postgres_test

import (
	"context"
	"errors"
	"rebutin/internal/domain"
	"rebutin/internal/repository/postgres"
	"rebutin/pkg/testutil"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
)

func TestSimulationRepository_Create(t *testing.T) {
	db, mock, err := sqlmock.New()

	assert.NoError(t, err)
	defer db.Close()

	logger, _ := testutil.SetupLogger(t)

	sqlxDB := sqlx.NewDb(db, "postgres")
	repo := postgres.NewSimulationRepository(sqlxDB, logger)

	t.Run("should success simulation", func(t *testing.T) {
		id, _ := uuid.NewV7()
		sim := domain.SimulationRun{
			ID:                 id,
			TotalTickets:       100,
			BotCount:           10,
			BotThrottleSeconds: 2,
			MaxConcurrent:      5,
			Status:             domain.StatusRunning,
			CreatedAt:          time.Now(),
		}

		expectedQuery := regexp.QuoteMeta(`INSERT INTO simulation_runs`)
		mock.ExpectExec(expectedQuery).
			WithArgs(
				sim.ID,
				sim.TotalTickets,
				sim.BotCount,
				sim.BotThrottleSeconds,
				sim.MaxConcurrent,
				sim.Status,
				sqlmock.AnyArg(),
			).
			WillReturnResult(sqlmock.NewResult(1, 1))

		err = repo.Create(context.Background(), &sim)

		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("should error failed insert simulation run", func(t *testing.T) {
		sim := &domain.SimulationRun{
			ID:           uuid.New(),
			TotalTickets: 100,
			BotCount:     10,
		}

		mock.ExpectExec("INSERT INTO simulation_runs").
			WillReturnError(errors.New("connection refused"))

		err = repo.Create(context.Background(), sim)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed insert simulation run")
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestSimulationRepository_GetByID(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	logger, _ := testutil.SetupLogger(t)

	sqlxDB := sqlx.NewDb(db, "postgres")
	repo := postgres.NewSimulationRepository(sqlxDB, logger)

	targetID, _ := uuid.NewV7()
	now := time.Now()

	t.Run("should success return rows data", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id", "total_tickets", "status", "created_at"}).
			AddRow(targetID, 100, domain.StatusRunning, now)

		mock.ExpectQuery(`^SELECT .+ FROM simulation_runs WHERE id = \$1`).
			WithArgs(targetID).
			WillReturnRows(rows)

		res, err := repo.GetByID(context.Background(), targetID)

		assert.NoError(t, err)
		assert.NotNil(t, res)
		assert.Equal(t, targetID, res.ID)
		assert.Equal(t, 100, res.TotalTickets)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("should error resource not found", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id", "total_tickets", "status", "created_at"})

		mock.ExpectQuery(`^SELECT .+ FROM simulation_runs WHERE id = \$1`).
			WithArgs(targetID).
			WillReturnRows(rows)

		res, err := repo.GetByID(context.Background(), targetID)

		assert.Error(t, err)
		assert.Nil(t, res)
		assert.ErrorIs(t, err, domain.ErrNotFound)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("should error failed query simulation run", func(t *testing.T) {
		mock.ExpectQuery(`^SELECT .+ FROM simulation_runs WHERE id = \$1`).
			WithArgs(targetID).
			WillReturnError(errors.New("connection refused"))

		_, err = repo.GetByID(context.Background(), targetID)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed query simulation run")
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestSimulationRepository_UpdateStatus(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	logger, _ := testutil.SetupLogger(t)

	sqlxDB := sqlx.NewDb(db, "postgres")
	repo := postgres.NewSimulationRepository(sqlxDB, logger)

	targetID, _ := uuid.NewV7()

	t.Run("should success return rows data", func(t *testing.T) {
		mock.ExpectExec(`^UPDATE simulation_runs SET status = \$1 WHERE id = \$2`).
			WithArgs(domain.StatusEnded, targetID).
			WillReturnResult(sqlmock.NewResult(0, 1))

		err := repo.UpdateStatus(context.Background(), targetID, domain.StatusEnded)

		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("should error resource not found", func(t *testing.T) {
		mock.ExpectExec(`^UPDATE simulation_runs SET status = \$1 WHERE id = \$2`).
			WithArgs(domain.StatusEnded, targetID).
			WillReturnResult(sqlmock.NewResult(0, 0))

		err := repo.UpdateStatus(context.Background(), targetID, domain.StatusEnded)

		assert.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrNotFound)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("should error failed update status simulation run", func(t *testing.T) {
		mock.ExpectExec(`^UPDATE simulation_runs SET status = \$1 WHERE id = \$2`).
			WithArgs(domain.StatusEnded, targetID).
			WillReturnError(errors.New("connection refused"))

		err = repo.UpdateStatus(context.Background(), targetID, domain.StatusEnded)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed update status simulation run")
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}
