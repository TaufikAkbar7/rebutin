package redis_test

import (
	"context"
	"fmt"
	"testing"

	"rebutin/internal/domain"
	"rebutin/internal/pkg/token"

	repoRedis "rebutin/internal/repository/redis"

	"github.com/alicebob/miniredis/v2"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
)

func TestSessionCacheRepository_SetBatch(t *testing.T) {
	mr, err := miniredis.Run()
	assert.NoError(t, err)
	defer mr.Close()

	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	repo := repoRedis.NewSessionCacheRepository(client)

	ctx := context.Background()

	t.Run("should success store data", func(t *testing.T) {
		sessionID, _ := token.GenerateRandomToken()
		sessionKey := "session:" + sessionID
		userParticipantID, _ := uuid.NewV7()
		botParticipantID, _ := uuid.NewV7()
		runID, _ := uuid.NewV7()
		runStr := runID.String()
		activeStr := string(domain.StatusActive)

		sessionData := []domain.UserSessionRedis{
			{
				SessionID:     &sessionID,
				ParticipantID: userParticipantID,
				RunID:         runID,
				Status:        domain.StatusActive,
			},
			{
				ParticipantID: botParticipantID,
				RunID:         runID,
				Status:        domain.StatusActive,
			},
		}

		err := repo.SetBatch(ctx, sessionData)
		assert.NoError(t, err)

		keys := mr.Keys()
		assert.Len(t, keys, 2)

		// get session boy key
		var sessionBotKey string
		for _, k := range keys {
			if k != sessionKey {
				sessionBotKey = k
				break
			}
		}

		resultUserParticipantID := mr.HGet(sessionKey, "participant_id")
		resultUserRunID := mr.HGet(sessionKey, "run_id")
		resultUserStatus := mr.HGet(sessionKey, "status")
		resultUserCategoryID := mr.HGet(sessionKey, "current_category_id")
		assert.Equal(t, userParticipantID.String(), resultUserParticipantID)
		assert.Equal(t, runStr, resultUserRunID)
		assert.Equal(t, activeStr, resultUserStatus)
		assert.Equal(t, "", resultUserCategoryID)

		resultBotParticipantID := mr.HGet(sessionBotKey, "participant_id")
		resultBotRunID := mr.HGet(sessionBotKey, "run_id")
		resultBotStatus := mr.HGet(sessionBotKey, "status")
		resultBotCategoryID := mr.HGet(sessionBotKey, "current_category_id")
		assert.Equal(t, botParticipantID.String(), resultBotParticipantID)
		assert.Equal(t, runStr, resultBotRunID)
		assert.Equal(t, activeStr, resultBotStatus)
		assert.Equal(t, "", resultBotCategoryID)
	})

	t.Run("should not saving data when data/payload is empty", func(t *testing.T) {
		sessionData := []domain.UserSessionRedis{}

		err := repo.SetBatch(ctx, sessionData)

		assert.NoError(t, err)
		assert.Empty(t, mr.Keys())
		assert.Equal(t, 0, len(mr.Keys()))
	})
}

func TestSessionCacheRepository_GetValueAllField(t *testing.T) {
	mr, err := miniredis.Run()
	assert.NoError(t, err)
	defer mr.Close()

	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	repo := repoRedis.NewSessionCacheRepository(client)
	ctx := context.Background()

	sessionID, _ := token.GenerateRandomToken()
	participantID, _ := uuid.NewV7()
	runID, _ := uuid.NewV7()
	status := string(domain.StatusActive)

	key := fmt.Sprintf("session:%s", sessionID)

	// store data
	mr.HSet(key, "current_category_id", "")
	mr.HSet(key, "participant_id", participantID.String())
	mr.HSet(key, "run_id", runID.String())
	mr.HSet(key, "status", status)

	t.Run("should success retrieve full hash data", func(t *testing.T) {
		result, err := repo.GetValueAllField(ctx, key)
		assert.NoError(t, err)

		keys := mr.Keys()
		assert.Len(t, keys, 1)

		assert.NotNil(t, result)
		assert.Equal(t, runID, result.RunID)
		assert.Equal(t, participantID, result.ParticipantID)
		assert.Equal(t, domain.StatusActive, result.Status)
		assert.Nil(t, nil, result.CurrentCategoryID)
	})

	t.Run("should return error not found when key not exist", func(t *testing.T) {
		nonExistentKey := "session:xxx"
		result, err := repo.GetValueAllField(ctx, nonExistentKey)

		assert.ErrorIs(t, err, domain.ErrNotFound)
		assert.Nil(t, result)
	})
}

func TestSessionCacheRepository_GetValueByField(t *testing.T) {
	mr, err := miniredis.Run()
	assert.NoError(t, err)
	defer mr.Close()

	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	repo := repoRedis.NewSessionCacheRepository(client)
	ctx := context.Background()

	sessionID, _ := token.GenerateRandomToken()
	participantID, _ := uuid.NewV7()
	runID, _ := uuid.NewV7()
	categoryID, _ := uuid.NewV7()
	status := string(domain.StatusActive)

	key := fmt.Sprintf("session:%s", sessionID)

	// store data
	mr.HSet(key, "current_category_id", categoryID.String())
	mr.HSet(key, "participant_id", participantID.String())
	mr.HSet(key, "run_id", runID.String())
	mr.HSet(key, "status", status)

	t.Run("should success retrieve data current_category_id", func(t *testing.T) {
		result, err := repo.GetValueByField(ctx, key, "current_category_id")
		assert.NoError(t, err)

		keys := mr.Keys()
		assert.Len(t, keys, 1)

		assert.NotNil(t, result)
		assert.Equal(t, categoryID.String(), result)
	})

	t.Run("should return ErrNotFound when field does not exist", func(t *testing.T) {
		val, err := repo.GetValueByField(ctx, key, "xxx")

		assert.ErrorIs(t, err, domain.ErrNotFound)
		assert.Nil(t, val)
	})

	t.Run("should return ErrNotFound when key doesnt exist", func(t *testing.T) {
		val, err := repo.GetValueByField(ctx, "session:xxx", "status")

		assert.ErrorIs(t, err, domain.ErrNotFound)
		assert.Nil(t, val)
	})
}
