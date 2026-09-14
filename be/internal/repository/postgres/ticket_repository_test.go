package postgres_test

import (
	"context"
	"database/sql/driver"
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

func TestTicketRepository_BatchCreate(t *testing.T) {
	db, mock, err := sqlmock.New()

	assert.NoError(t, err)
	defer db.Close()

	logger, _ := testutil.SetupLogger(t)

	sqlxDB := sqlx.NewDb(db, "postgres")
	repo := postgres.NewTicketRepository(sqlxDB, logger)

	runID, _ := uuid.NewV7()
	cat1ID, _ := uuid.NewV7()
	cat2ID, _ := uuid.NewV7()

	payload := []domain.TicketCategories{
		{
			ID:        cat1ID,
			RunID:     runID,
			Name:      domain.CategoryCAT1.Name,
			Quota:     1,
			Price:     domain.CategoryCAT1.Price,
			CreatedAt: time.Now(),
		},
		{
			ID:        cat2ID,
			RunID:     runID,
			Name:      domain.CategoryCAT2.Name,
			Quota:     1,
			Price:     domain.CategoryCAT2.Price,
			CreatedAt: time.Now(),
		},
	}

	t.Run("should success create data", func(t *testing.T) {
		expectedArgs := []driver.Value{
			payload[0].ID, payload[0].RunID, payload[0].Name, payload[0].Quota, payload[0].Price, sqlmock.AnyArg(),
			payload[1].ID, payload[1].RunID, payload[1].Name, payload[1].Quota, payload[1].Price, sqlmock.AnyArg(),
		}

		expectedQuery := regexp.QuoteMeta(`INSERT INTO ticket_categories`)
		mock.ExpectExec(expectedQuery).
			WithArgs(expectedArgs...).
			WillReturnResult(sqlmock.NewResult(0, int64(len(payload))))

		err = repo.BatchCreate(context.Background(), payload)

		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
		assert.Len(t, payload, 2)
	})

	t.Run("should error failed insert ticket", func(t *testing.T) {
		expectedArgs := []driver.Value{
			payload[0].ID, payload[0].RunID, payload[0].Name, payload[0].Quota, payload[0].Price, sqlmock.AnyArg(),
			payload[1].ID, payload[1].RunID, payload[1].Name, payload[1].Quota, payload[1].Price, sqlmock.AnyArg(),
		}

		expectedQuery := regexp.QuoteMeta(`INSERT INTO ticket_categories`)
		mock.ExpectExec(expectedQuery).
			WithArgs(expectedArgs...).
			WillReturnError(errors.New("connection refused"))

		err = repo.BatchCreate(context.Background(), payload)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed insert ticket")
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestTicketRepository_GetByRunID(t *testing.T) {
	db, mock, err := sqlmock.New()

	assert.NoError(t, err)
	defer db.Close()

	logger, _ := testutil.SetupLogger(t)

	sqlxDB := sqlx.NewDb(db, "postgres")
	repo := postgres.NewTicketRepository(sqlxDB, logger)

	runID, _ := uuid.NewV7()
	cat1ID, _ := uuid.NewV7()
	var rows = []string{"id", "run_id", "name", "quota", "price"}

	t.Run("should success return rows data", func(t *testing.T) {
		rows := sqlmock.NewRows(rows).
			AddRow(cat1ID, runID, domain.CategoryCAT1.Name, 1, domain.CategoryCAT1.Price)

		mock.ExpectQuery(`^SELECT .+ FROM ticket_categories WHERE run_id = \$1`).
			WithArgs(runID).
			WillReturnRows(rows)

		res, err := repo.GetByRunID(context.Background(), runID)

		assert.NoError(t, err)
		assert.NotNil(t, res)
		assert.Len(t, res, 1)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("should error resource not found", func(t *testing.T) {
		rows := sqlmock.NewRows(rows)

		mock.ExpectQuery(`^SELECT .+ FROM ticket_categories WHERE run_id = \$1`).
			WithArgs(runID).
			WillReturnRows(rows)

		res, err := repo.GetByRunID(context.Background(), runID)

		assert.Error(t, err)
		assert.Nil(t, res)
		assert.ErrorIs(t, err, domain.ErrNotFound)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("should error failed query ticket", func(t *testing.T) {
		mock.ExpectQuery(`^SELECT .+ FROM ticket_categories WHERE run_id = \$1`).
			WithArgs(runID).
			WillReturnError(errors.New("connection refused"))

		_, err = repo.GetByRunID(context.Background(), runID)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed query ticket")
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}
