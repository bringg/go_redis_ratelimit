package cloudflare

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/bringg/go_redis_ratelimit/algorithm"
)

type RedisDataStore struct {
	RDB            algorithm.Rediser
	ExpirationTime time.Duration
}

func (s *RedisDataStore) Inc(key string, window time.Time) error {
	ctx := context.Background()
	key = mapKey(key, window)

	pipe := s.RDB.TxPipeline()
	pipe.Incr(ctx, key)
	pipe.Expire(ctx, key, s.ExpirationTime)

	if _, err := pipe.Exec(ctx); err != nil {
		return err
	}

	return nil
}

func (s *RedisDataStore) Get(key string, previousWindow, currentWindow time.Time) (prevValue int64, currValue int64, err error) {
	ctx := context.Background()

	res, err := s.RDB.MGet(ctx, mapKey(key, previousWindow), mapKey(key, currentWindow)).Result()

	if err != nil {
		return
	}

	if res[0] != nil {
		prevValue, err = strconv.ParseInt(res[0].(string), 10, 64)
		if err != nil {
			err = fmt.Errorf("failed parsing previous value: %q", res[0])
			return
		}
	}

	if res[1] != nil {
		currValue, err = strconv.ParseInt(res[1].(string), 10, 64)
		if err != nil {
			err = fmt.Errorf("failed parsing current value: %q", res[0])
			return
		}
	}

	return prevValue, currValue, nil
}

func mapKey(key string, window time.Time) string {
	return key + "_" + window.Format(time.RFC3339)
}
