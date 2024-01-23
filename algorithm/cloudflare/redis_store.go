package cloudflare

import (
	"context"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"

	"github.com/bringg/go_redis_ratelimit/algorithm"
)

type RedisDataStore struct {
	RDB            algorithm.Rediser
	ExpirationTime time.Duration
}

func (s *RedisDataStore) Inc(key string, window time.Time) error {
	ctx := context.Background()
	key = mapKey(key, window)

	if _, err := s.RDB.TxPipelined(ctx, func(pipe redis.Pipeliner) error {
		pipe.Incr(ctx, key)
		pipe.Expire(ctx, key, s.ExpirationTime)

		return nil
	}); err != nil {
		return err
	}

	return nil
}

func (s *RedisDataStore) Get(key string, previousWindow, currentWindow time.Time) (prevValue int64, currValue int64, err error) {
	ctx := context.Background()
	pipe := s.RDB.TxPipeline()
	prevRes := pipe.Get(ctx, mapKey(key, previousWindow))
	currRes := pipe.Get(ctx, mapKey(key, currentWindow))

	_, err = pipe.Exec(ctx)
	if err != nil && err != redis.Nil {
		return
	}

	prevValue, _ = prevRes.Int64()
	currValue, _ = currRes.Int64()

	return prevValue, currValue, nil
}

func mapKey(key string, window time.Time) string {
	return fmt.Sprintf("%s_%s", key, window.Format(time.RFC3339))
}
