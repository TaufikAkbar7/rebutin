package redis

import (
	"context"
	"errors"
	"fmt"
	"time"

	"rebutin/internal/domain"
	"rebutin/internal/model"
	"rebutin/internal/pkg/token"

	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
)

type sessionCacheRepository struct {
	rdb *redis.Client
	log *logrus.Logger
}

func NewSessionCacheRepository(rdb *redis.Client, log *logrus.Logger) domain.SessionCacheRepository {
	return &sessionCacheRepository{rdb: rdb, log: log}
}

func (r *sessionCacheRepository) RunState(ctx context.Context, runID string, max int, data []domain.UserSessionRedis) error {
	if len(data) == 0 {
		return nil
	}

	pipe := r.rdb.TxPipeline()

	script := redis.NewScript(`
		local runID = ARGV[1]
		local current = tonumber(redis.call('HGET', KEYS[1], runID) or "0")
		local max = tonumber(ARGV[2])
		local queued = "queued"
		local active = "active"

		-- check if current not exceed max 
		-- then go to ticket page
		if current < max then
			redis.call('HINCRBY', KEYS[1], runID, 1)
			redis.call('HSET', KEYS[2], "status", active)
			redis.call('EXPIRE', KEYS[2], 600)
			return 1
		else
		-- join to waiting room
			redis.call('HSET', KEYS[2], "status", queued)
			redis.call('ZADD', KEYS[3], ARGV[3], ARGV[4])
			return 0
		end
	`)

	if err := script.Load(ctx, r.rdb).Err(); err != nil {
		return err
	}

	waitingRoom := fmt.Sprintf("waiting_room_queue:%s", runID)
	cmds := make([]*redis.Cmd, len(data))
	for i := range data {
		// save session every participant
		if data[i].SessionID == nil {
			token, _ := token.GenerateRandomToken()
			data[i].SessionID = &token
		}

		key := fmt.Sprintf("session:%s", *data[i].SessionID)
		modelData := model.ToUserSessionRedisModel(&data[i])

		pipe.HSet(ctx, key, modelData)

		// init counter
		pipe.HSetNX(ctx, "active_sessions_count", runID, 0)

		// setup value unix timestamp for waiting_room_queue
		now := time.Now().Unix()

		// check and enter after reshuffle
		participantKeySession := fmt.Sprintf("session:%s", *data[i].SessionID)
		keys := []string{"active_sessions_count", participantKeySession, waitingRoom}

		cmds[i] = script.Run(ctx, pipe, keys, runID, max, now, *data[i].SessionID)
	}

	_, err := pipe.Exec(ctx)

	for i, cmd := range cmds {
		val, cmdErr := cmd.Result()
		r.log.Infof("[Simulation.RunState] User %s | Result: %v | Error: %v\n", *data[i].SessionID, val, cmdErr)
	}

	return err
}

func (r *sessionCacheRepository) GetValueAllField(ctx context.Context, key string) (*domain.UserSessionRedis, error) {
	val, err := r.rdb.HGetAll(ctx, key).Result()
	if err != nil {
		return nil, err
	}

	if len(val) == 0 {
		return nil, domain.ErrNotFound
	}

	return model.ToDomainUserSession(val), nil
}

func (r *sessionCacheRepository) GetValueByField(ctx context.Context, key string, field string) (*string, error) {
	val, err := r.rdb.HGet(ctx, key, field).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}

	return &val, nil
}
