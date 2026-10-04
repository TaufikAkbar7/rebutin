package redis

import (
	"context"
	"encoding/json"
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
			redis.call('ZADD', KEYS[3], 'NX', ARGV[3], ARGV[4])
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
		now := time.Now().UnixMicro()

		// check and enter after reshuffle
		participantKeySession := fmt.Sprintf("session:%s", *data[i].SessionID)
		keys := []string{"active_sessions_count", participantKeySession, waitingRoom}

		cmds[i] = script.Run(ctx, pipe, keys, runID, max, now+int64(i), *data[i].SessionID)
	}

	_, err := pipe.Exec(ctx)

	for i, cmd := range cmds {
		val, cmdErr := cmd.Result()
		r.log.Infof("[Session.RunState] User %s | Result: %v | Error: %v\n", *data[i].SessionID, val, cmdErr)
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

func (r *sessionCacheRepository) Drain(ctx context.Context, runID string, max, batch int) ([]string, error) {
	leakyScript := redis.NewScript(`
		local current = tonumber(redis.call('HGET', KEYS[1], ARGV[1]) or "0")
		local n = math.min(tonumber(ARGV[2]) - current, tonumber(ARGV[3]))
		if n <= 0 then return {} end

		local popped = redis.call('ZPOPMIN', KEYS[2], n)
		local admitted = {}

		-- looping odd number to get session
		for i = 1, #popped, 2 do
			-- concat str
		    local key = 'session:' .. popped[i]
		    redis.call('HSET', key, 'status', ARGV[4])
		    redis.call('EXPIRE', key, 600)
		    redis.call('HINCRBY', KEYS[1], ARGV[1], 1)
		    admitted[#admitted + 1] = key
		end
		return admitted
	`)

	waitingRoom := fmt.Sprintf("waiting_room_queue:%s", runID)
	values, err := leakyScript.Run(ctx, r.rdb, []string{"active_sessions_count", waitingRoom}, runID, max, batch, string(domain.StatusActive)).StringSlice()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, nil
		}
		return nil, err
	}

	valuesJSON, _ := json.MarshalIndent(values, "", "  ")
	r.log.Infof("[Session.Drainer] Participant left the waiting room %s\n", valuesJSON)

	return values, nil
}
