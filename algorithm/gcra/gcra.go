// https://github.com/go-redis/redis_rate
package gcra

import (
	"context"

	"github.com/go-redis/redis_rate/v10"

	"github.com/bringg/go_redis_ratelimit/algorithm"
)

const AlgorithmName = "gcra"

type GCRA struct {
	limiter *redis_rate.Limiter
}

func init() {
	algorithm.Register(&algorithm.RegInfo{
		Name:         AlgorithmName,
		NewAlgorithm: NewAlgorithm,
	})
}

func NewAlgorithm(rdb algorithm.Rediser) (algorithm.Algorithm, error) {
	return &GCRA{
		limiter: redis_rate.NewLimiter(rdb),
	}, nil
}

func (c *GCRA) Allow(key string, limit algorithm.Limit) (*algorithm.Result, error) {
	res, err := c.limiter.Allow(context.Background(), key, redis_rate.Limit{
		Rate:   int(limit.GetRate()),
		Period: limit.GetPeriod(),
		Burst:  int(limit.GetBurst()),
	})
	if err != nil {
		return nil, err
	}

	return &algorithm.Result{
		Limit:      limit,
		Key:        key,
		Allowed:    res.Allowed == 1,
		Remaining:  int64(res.Remaining),
		RetryAfter: res.RetryAfter,
		ResetAfter: res.ResetAfter,
	}, nil
}
