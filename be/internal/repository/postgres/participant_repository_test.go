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

func TestParticipantRepository_BatchCreate(t *testing.T) {
	db, mock, err := sqlmock.New()

	assert.NoError(t, err)
	defer db.Close()

	logger, _ := testutil.SetupLogger(t)

	sqlxDB := sqlx.NewDb(db, "postgres")
	repo := postgres.NewParticipantRepository(sqlxDB, logger)

	runID, _ := uuid.NewV7()
	userID, _ := uuid.NewV7()
	botID, _ := uuid.NewV7()

	payload := []domain.Participant{
		{
			ID:         userID,
			RunID:      runID,
			IsBot:      false,
			Identifier: "real-user",
			CreatedAt:  time.Now(),
		},
		{
			ID:         botID,
			RunID:      runID,
			IsBot:      true,
			Identifier: "bot-1",
			CreatedAt:  time.Now(),
		},
	}

	t.Run("should success create data", func(t *testing.T) {
		expectedArgs := []driver.Value{
			payload[0].ID, payload[0].RunID, payload[0].IsBot, payload[0].Identifier, sqlmock.AnyArg(),
			payload[1].ID, payload[1].RunID, payload[1].IsBot, payload[1].Identifier, sqlmock.AnyArg(),
		}

		expectedQuery := regexp.QuoteMeta(`INSERT INTO participants`)
		mock.ExpectExec(expectedQuery).
			WithArgs(expectedArgs...).
			WillReturnResult(sqlmock.NewResult(0, int64(len(payload))))

		err = repo.BatchCreate(context.Background(), payload)

		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
		assert.Len(t, payload, 2)
	})

	t.Run("should error failed insert participant", func(t *testing.T) {
		expectedArgs := []driver.Value{
			payload[0].ID, payload[0].RunID, payload[0].IsBot, payload[0].Identifier, sqlmock.AnyArg(),
			payload[1].ID, payload[1].RunID, payload[1].IsBot, payload[1].Identifier, sqlmock.AnyArg(),
		}

		expectedQuery := regexp.QuoteMeta(`INSERT INTO participants`)
		mock.ExpectExec(expectedQuery).
			WithArgs(expectedArgs...).
			WillReturnError(errors.New("connection refused"))

		err = repo.BatchCreate(context.Background(), payload)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed insert participant")
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}
