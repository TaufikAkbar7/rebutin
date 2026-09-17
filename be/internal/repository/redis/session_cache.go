package redis

import (
	"context"
	"errors"
	"fmt"

	"rebutin/internal/domain"
	"rebutin/internal/model"
	"rebutin/internal/pkg/token"

	"github.com/redis/go-redis/v9"
)

type sessionCacheRepository struct {
	rdb *redis.Client
}

func NewSessionCacheRepository(rdb *redis.Client) domain.SessionCacheRepository {
	return &sessionCacheRepository{rdb: rdb}
}

func (r *sessionCacheRepository) SetBatch(ctx context.Context, data []domain.UserSessionRedis) error {
	if len(data) == 0 {
		return nil
	}

	pipe := r.rdb.Pipeline()

	for i := range data {
		if data[i].SessionID == nil {
			token, _ := token.GenerateRandomToken()
			data[i].SessionID = &token
		}

		key := fmt.Sprintf("session:%s", *data[i].SessionID)
		data := model.ToUserSessionRedisModel(&data[i])

		pipe.HSet(ctx, key, data)
	}

	_, err := pipe.Exec(ctx)

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
