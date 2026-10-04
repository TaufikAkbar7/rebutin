package redis_test

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"rebutin/internal/domain"
	"rebutin/internal/pkg/token"
	"rebutin/pkg/testutil"

	repoRedis "rebutin/internal/repository/redis"

	"github.com/alicebob/miniredis/v2"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSessionCacheRepository_RunState(t *testing.T) {
	t.Run("should all participant avoid waiting room and status active", func(t *testing.T) {
		mr, err := miniredis.Run()
		assert.NoError(t, err)
		defer mr.Close()

		client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
		logger, _ := testutil.SetupLogger(t)
		repo := repoRedis.NewSessionCacheRepository(client, logger)

		ctx := context.Background()

		sessionID, _ := token.GenerateRandomToken()
		sessionKey := "session:" + sessionID
		userParticipantID, _ := uuid.NewV7()
		botParticipantID, _ := uuid.NewV7()
		runID, _ := uuid.NewV7()
		runStr := runID.String()
		activeStr := string(domain.StatusActive)
		max := 3

		sessionData := []domain.UserSessionRedis{
			{
				SessionID:     &sessionID,
				ParticipantID: userParticipantID,
				RunID:         runID,
				Status:        "",
			},
			{
				ParticipantID: botParticipantID,
				RunID:         runID,
				Status:        "",
			},
		}

		err = repo.RunState(ctx, runID.String(), max, sessionData)
		assert.NoError(t, err)

		keys := mr.Keys()
		assert.Len(t, keys, 3)

		// get session boy key
		var sessionBotKey string
		for _, k := range keys {
			if k != sessionKey && k != "active_sessions_count" {
				sessionBotKey = k
				break
			}
		}

		// assert active_sessions_count
		activeSession := mr.HGet("active_sessions_count", runStr)
		assert.NotNil(t, activeSession)
		assert.Equal(t, activeSession, "2")

		// assert waiting_room_queue
		waitingRoomKey := fmt.Sprintf("waiting_room_queue:%s", runID)
		assert.False(t, mr.Exists(waitingRoomKey))

		resultUserParticipantID := mr.HGet(sessionKey, "participant_id")
		resultUserRunID := mr.HGet(sessionKey, "run_id")
		resultUserStatus := mr.HGet(sessionKey, "status")
		resultUserCategoryID := mr.HGet(sessionKey, "current_category_id")
		resultUserTTL := mr.TTL(sessionKey)
		assert.True(t, resultUserTTL > 0)
		assert.Equal(t, 600*time.Second, resultUserTTL)
		assert.Equal(t, userParticipantID.String(), resultUserParticipantID)
		assert.Equal(t, runStr, resultUserRunID)
		assert.Equal(t, activeStr, resultUserStatus)
		assert.Equal(t, "", resultUserCategoryID)

		resultBotParticipantID := mr.HGet(sessionBotKey, "participant_id")
		resultBotRunID := mr.HGet(sessionBotKey, "run_id")
		resultBotStatus := mr.HGet(sessionBotKey, "status")
		resultBotCategoryID := mr.HGet(sessionBotKey, "current_category_id")
		resultBotTTL := mr.TTL(sessionKey)
		assert.True(t, resultBotTTL > 0)
		assert.Equal(t, 600*time.Second, resultBotTTL)
		assert.Equal(t, botParticipantID.String(), resultBotParticipantID)
		assert.Equal(t, runStr, resultBotRunID)
		assert.Equal(t, activeStr, resultBotStatus)
		assert.Equal(t, "", resultBotCategoryID)
	})

	t.Run("should some participant join waiting room and status queued", func(t *testing.T) {
		mr, err := miniredis.Run()
		assert.NoError(t, err)
		defer mr.Close()

		client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
		logger, _ := testutil.SetupLogger(t)
		repo := repoRedis.NewSessionCacheRepository(client, logger)

		ctx := context.Background()

		sessionID, _ := token.GenerateRandomToken()
		sessionKey := "session:" + sessionID
		userParticipantID, _ := uuid.NewV7()
		botParticipantID, _ := uuid.NewV7()
		runID, _ := uuid.NewV7()
		runStr := runID.String()
		activeStr := string(domain.StatusActive)
		max := 1

		sessionData := []domain.UserSessionRedis{
			{
				SessionID:     &sessionID,
				ParticipantID: userParticipantID,
				RunID:         runID,
				Status:        "",
			},
			{
				ParticipantID: botParticipantID,
				RunID:         runID,
				Status:        "",
			},
		}

		err = repo.RunState(ctx, runID.String(), max, sessionData)
		assert.NoError(t, err)

		keys := mr.Keys()
		assert.Len(t, keys, 4)

		// get session boy key
		waitingRoomKey := fmt.Sprintf("waiting_room_queue:%s", runID)
		var sessionBotKey string
		for _, k := range keys {
			if k != sessionKey && k != "active_sessions_count" && k != waitingRoomKey {
				sessionBotKey = k
				break
			}
		}

		// assert active_sessions_count
		activeSession := mr.HGet("active_sessions_count", runStr)
		assert.NotNil(t, activeSession)
		assert.Equal(t, activeSession, "1")

		// assert waiting_room_queue
		queueRoom, err := mr.ZMembers(waitingRoomKey)
		assert.NoError(t, err)
		assert.Len(t, queueRoom, 1)
		botKey := strings.Split(sessionBotKey, ":")
		assert.Equal(t, queueRoom[0], botKey[1])

		resultUserParticipantID := mr.HGet(sessionKey, "participant_id")
		resultUserRunID := mr.HGet(sessionKey, "run_id")
		resultUserStatus := mr.HGet(sessionKey, "status")
		resultUserCategoryID := mr.HGet(sessionKey, "current_category_id")
		resultUserTTL := mr.TTL(sessionKey)
		assert.True(t, resultUserTTL > 0)
		assert.Equal(t, 600*time.Second, resultUserTTL)
		assert.Equal(t, userParticipantID.String(), resultUserParticipantID)
		assert.Equal(t, runStr, resultUserRunID)
		assert.Equal(t, activeStr, resultUserStatus)
		assert.Equal(t, "", resultUserCategoryID)

		resultBotParticipantID := mr.HGet(sessionBotKey, "participant_id")
		resultBotRunID := mr.HGet(sessionBotKey, "run_id")
		resultBotStatus := mr.HGet(sessionBotKey, "status")
		resultBotCategoryID := mr.HGet(sessionBotKey, "current_category_id")
		resultBotTTL := mr.TTL(sessionBotKey)
		assert.Equal(t, resultBotTTL, time.Duration(0))
		assert.Equal(t, botParticipantID.String(), resultBotParticipantID)
		assert.Equal(t, runStr, resultBotRunID)
		assert.Equal(t, string(domain.StatusQueued), resultBotStatus)
		assert.Equal(t, "", resultBotCategoryID)
	})

	t.Run("should participant session deleted after TTL expired", func(t *testing.T) {
		mr, err := miniredis.Run()
		assert.NoError(t, err)
		defer mr.Close()

		client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
		logger, _ := testutil.SetupLogger(t)
		repo := repoRedis.NewSessionCacheRepository(client, logger)

		ctx := context.Background()

		sessionID, _ := token.GenerateRandomToken()
		sessionKey := "session:" + sessionID
		userParticipantID, _ := uuid.NewV7()
		runID, _ := uuid.NewV7()
		activeStr := string(domain.StatusActive)

		sessionData := []domain.UserSessionRedis{
			{
				SessionID:     &sessionID,
				ParticipantID: userParticipantID,
				RunID:         runID,
				Status:        "",
			},
		}

		err = repo.RunState(ctx, runID.String(), 1, sessionData)
		assert.NoError(t, err)

		keys := mr.Keys()
		assert.Len(t, keys, 2)

		resultUserStatus := mr.HGet(sessionKey, "status")
		resultUserTTL := mr.TTL(sessionKey)
		assert.True(t, resultUserTTL > 0)
		assert.Equal(t, 600*time.Second, resultUserTTL)
		assert.Equal(t, activeStr, resultUserStatus)

		// fast forward
		mr.FastForward(600 * time.Second)
		assert.False(t, mr.Exists(sessionKey))
	})

	t.Run("should not saving data when data/payload is empty", func(t *testing.T) {
		mr, err := miniredis.Run()
		assert.NoError(t, err)
		defer mr.Close()

		client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
		logger, _ := testutil.SetupLogger(t)
		repo := repoRedis.NewSessionCacheRepository(client, logger)

		ctx := context.Background()

		sessionData := []domain.UserSessionRedis{}
		runID, _ := uuid.NewV7()

		err = repo.RunState(ctx, runID.String(), 1, sessionData)

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
	logger, _ := testutil.SetupLogger(t)
	repo := repoRedis.NewSessionCacheRepository(client, logger)
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
	logger, _ := testutil.SetupLogger(t)
	repo := repoRedis.NewSessionCacheRepository(client, logger)
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

func TestSessionCacheRepository_Drain(t *testing.T) {
	t.Run("should drain batch = 2", func(t *testing.T) {
		mr, err := miniredis.Run()
		assert.NoError(t, err)
		defer mr.Close()

		client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
		logger, _ := testutil.SetupLogger(t)
		repo := repoRedis.NewSessionCacheRepository(client, logger)

		ctx := context.Background()
		runID := "run-1"
		waitingRoom := fmt.Sprintf("waiting_room_queue:%s", runID)
		activeStr := string(domain.StatusActive)

		mr.HSet("active_sessions_count", runID, "0")
		_, _ = mr.ZAdd(waitingRoom, 1, "one")
		_, _ = mr.ZAdd(waitingRoom, 2, "two")
		_, _ = mr.ZAdd(waitingRoom, 3, "three")

		result, err := repo.Drain(ctx, runID, 5, 2)
		assert.NoError(t, err)

		assert.Len(t, result, 2)
		require.Equal(t, []string{"session:one", "session:two"}, result)

		sessionOne := result[0]
		sessionOneTTL := mr.TTL(sessionOne)
		sessionOneStatus := mr.HGet(sessionOne, "status")
		assert.Equal(t, 600*time.Second, sessionOneTTL)
		assert.Equal(t, activeStr, sessionOneStatus)

		sessionTwo := result[0]
		sessionTwoTTL := mr.TTL(sessionTwo)
		sessionTwoStatus := mr.HGet(sessionTwo, "status")
		assert.Equal(t, 600*time.Second, sessionTwoTTL)
		assert.Equal(t, activeStr, sessionTwoStatus)

		incr := mr.HGet("active_sessions_count", runID)
		assert.Equal(t, "2", incr)

		members, _ := mr.ZMembers(waitingRoom)
		require.Equal(t, []string{"three"}, members)
	})

	t.Run("should release one session", func(t *testing.T) {
		mr, err := miniredis.Run()
		assert.NoError(t, err)
		defer mr.Close()

		client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
		logger, _ := testutil.SetupLogger(t)
		repo := repoRedis.NewSessionCacheRepository(client, logger)

		ctx := context.Background()
		runID := "run-1"
		waitingRoom := fmt.Sprintf("waiting_room_queue:%s", runID)

		mr.HSet("active_sessions_count", runID, "2")
		_, _ = mr.ZAdd(waitingRoom, 1, "one")
		_, _ = mr.ZAdd(waitingRoom, 2, "two")
		_, _ = mr.ZAdd(waitingRoom, 3, "three")

		result, err := repo.Drain(ctx, runID, 3, 5)
		assert.NoError(t, err)

		assert.Len(t, result, 1)
		require.Equal(t, []string{"session:one"}, result)

		count := mr.HGet("active_sessions_count", runID)
		require.Equal(t, "3", count)

		members, _ := mr.ZMembers(waitingRoom)
		require.Equal(t, []string{"two", "three"}, members)
	})

	t.Run("should no slot avail", func(t *testing.T) {
		mr, err := miniredis.Run()
		assert.NoError(t, err)
		defer mr.Close()

		client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
		logger, _ := testutil.SetupLogger(t)
		repo := repoRedis.NewSessionCacheRepository(client, logger)

		ctx := context.Background()
		runID := "run-1"
		waitingRoom := fmt.Sprintf("waiting_room_queue:%s", runID)

		mr.HSet("active_sessions_count", runID, "2")
		_, _ = mr.ZAdd(waitingRoom, 1, "one")

		result, err := repo.Drain(ctx, runID, 2, 5)
		assert.NoError(t, err)

		assert.Len(t, result, 0)
		require.Equal(t, []string{}, result)

		members, _ := mr.ZMembers(waitingRoom)
		require.Equal(t, []string{"one"}, members)
	})

	t.Run("should queue is empty", func(t *testing.T) {
		mr, err := miniredis.Run()
		assert.NoError(t, err)
		defer mr.Close()

		client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
		logger, _ := testutil.SetupLogger(t)
		repo := repoRedis.NewSessionCacheRepository(client, logger)

		ctx := context.Background()
		runID := "run-1"

		mr.HSet("active_sessions_count", runID, "0")

		result, err := repo.Drain(ctx, runID, 2, 5)
		assert.NoError(t, err)

		assert.Empty(t, result)
	})

	t.Run("should active_sessions_count start 0 when active_sessions_count is missing", func(t *testing.T) {
		mr, err := miniredis.Run()
		assert.NoError(t, err)
		defer mr.Close()

		client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
		logger, _ := testutil.SetupLogger(t)
		repo := repoRedis.NewSessionCacheRepository(client, logger)

		ctx := context.Background()
		runID := "run-1"
		waitingRoom := fmt.Sprintf("waiting_room_queue:%s", runID)

		_, _ = mr.ZAdd(waitingRoom, 1, "one")

		result, err := repo.Drain(ctx, runID, 1, 4)
		assert.NoError(t, err)

		assert.Len(t, result, 1)
		require.Equal(t, []string{"session:one"}, result)

		count := mr.HGet("active_sessions_count", runID)
		assert.Equal(t, "1", count)
	})

	t.Run("should lowest score first waiting_room_queue", func(t *testing.T) {
		mr, err := miniredis.Run()
		assert.NoError(t, err)
		defer mr.Close()

		client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
		logger, _ := testutil.SetupLogger(t)
		repo := repoRedis.NewSessionCacheRepository(client, logger)

		ctx := context.Background()
		runID := "run-1"
		waitingRoom := fmt.Sprintf("waiting_room_queue:%s", runID)

		mr.HSet("active_sessions_count", runID, "0")
		_, _ = mr.ZAdd(waitingRoom, 300, "three")
		_, _ = mr.ZAdd(waitingRoom, 100, "one")
		_, _ = mr.ZAdd(waitingRoom, 200, "two")

		result, err := repo.Drain(ctx, runID, 5, 2)
		assert.NoError(t, err)

		assert.Len(t, result, 2)
		require.Equal(t, []string{"session:one", "session:two"}, result)

		count := mr.HGet("active_sessions_count", runID)
		assert.Equal(t, "2", count)
	})

	t.Run("should isolated per runID", func(t *testing.T) {
		mr, err := miniredis.Run()
		assert.NoError(t, err)
		defer mr.Close()

		client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
		logger, _ := testutil.SetupLogger(t)
		repo := repoRedis.NewSessionCacheRepository(client, logger)

		ctx := context.Background()

		mr.HSet("active_sessions_count", "run-1", "0")
		mr.HSet("active_sessions_count", "run-2", "0")
		_, _ = mr.ZAdd("waiting_room_queue:run-1", 1, "a1")
		_, _ = mr.ZAdd("waiting_room_queue:run-2", 1, "b1")

		got, err := repo.Drain(ctx, "run-1", 5, 5)
		require.NoError(t, err)
		require.Equal(t, []string{"session:a1"}, got)

		got, err = repo.Drain(ctx, "run-2", 5, 5)
		require.NoError(t, err)
		require.Equal(t, []string{"session:b1"}, got)
	})
}
