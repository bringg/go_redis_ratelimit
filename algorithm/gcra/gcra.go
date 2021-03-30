// https://github.com/go-redis/redis_rate
package gcra

import (
	"context"

	"github.com/go-redis/redis_rate/v9"

	"github.com/bringg/go_redis_ratelimit/algorithm"
)

const AlgorithmName = "gcra"

type GCRA struct {
	Limit algorithm.Limit
	RDB   algorithm.Rediser

	key string
}

func (c *GCRA) Allow() (*algorithm.Result, error) {
	res, err := redis_rate.NewLimiter(c.RDB).Allow(context.Background(), c.key, redis_rate.Limit{
		Rate:   int(c.Limit.GetRate()),
		Period: c.Limit.GetPeriod(),
		Burst:  int(c.Limit.GetBurst()),
	})
	if err != nil {
		return nil, err
	}

	return &algorithm.Result{
		Limit:      c.Limit,
		Key:        c.key,
		Allowed:    res.Allowed == 1,
		Remaining:  int64(res.Remaining),
		RetryAfter: res.RetryAfter,
		ResetAfter: res.ResetAfter,
	}, nil
}

// SetKey _
func (c *GCRA) SetKey(key string) {
	c.key = key
}
