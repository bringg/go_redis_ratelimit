package sliding_window

import (
	"context"
	"strconv"
	"time"

	"github.com/bringg/go_redis_ratelimit/algorithm"
)

const AlgorithmName = "sliding_window"

type SlidingWindow struct {
	Limit algorithm.Limit
	RDB   algorithm.Rediser

	key string
}

func (c *SlidingWindow) SetKey(key string) {
	c.key = key
}

func (c *SlidingWindow) Allow() (r *algorithm.Result, err error) {
	limit := c.Limit
	values := []interface{}{limit.GetRate(), limit.GetPeriod().Seconds()}

	v, err := script2.Run(context.Background(), c.RDB, []string{c.key}, values...).Result()
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
		Key:        c.key,
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
