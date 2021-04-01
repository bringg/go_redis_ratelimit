package sliding_window

import (
	"context"
	"strconv"
	"time"

	"github.com/bringg/go_redis_ratelimit/algorithm"
)

const AlgorithmName = "sliding_window"

type SlidingWindow struct {
	RDB algorithm.Rediser
}

func init() {
	algorithm.Register(&algorithm.RegInfo{
		Name:         AlgorithmName,
		NewAlgorithm: NewAlgorithm,
	})
}

func NewAlgorithm(rdb algorithm.Rediser) (algorithm.Algorithm, error) {
	return &SlidingWindow{
		RDB: rdb,
	}, nil
}

func (c *SlidingWindow) Allow(key string, limit algorithm.Limit) (r *algorithm.Result, err error) {
	values := []interface{}{limit.GetRate(), limit.GetPeriod().Seconds()}

	v, err := script.Run(context.Background(), c.RDB, []string{key}, values...).Result()
	if err != nil {
		return nil, err
	}

	values = v.([]interface{})

	retryAfter, err := strconv.ParseFloat(values[2].(string), 64)
	if err != nil {
		return nil, err
	}

	return &algorithm.Result{
		Limit:      limit,
		Key:        key,
		Allowed:    values[0].(int64) == 1,
		Remaining:  values[1].(int64),
		RetryAfter: dur(retryAfter),
		ResetAfter: limit.GetPeriod(),
	}, nil
}

func dur(f float64) time.Duration {
	if f == -1 {
		return -1
	}

	return time.Duration(f * float64(time.Second))
}
