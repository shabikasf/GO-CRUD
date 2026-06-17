package store

import (
	"context"
	"encoding/json"
	"time"

	"github.com/redis/go-redis/v9"
)


type RedisStore struct{
	rdb *redis.Client
}

func NewRedisStore(rdb *redis.Client) *RedisStore {
	return &RedisStore{
		rdb: rdb,
	}
}

func (r *RedisStore) Get(ctx context.Context, key string) (string, error) {

	value, err := r.rdb.Get(ctx, key).Result()

	if err != nil{
		return "", err
	}

	return value, nil
}

func (r *RedisStore) Set(ctx context.Context, key string, value any, ttl time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}

	return r.rdb.Set(ctx, key, data, ttl).Err()
}

func (r *RedisStore) Delete(ctx context.Context, key string) error{
	err := r.rdb.Del(ctx, key).Err()
	return err
}