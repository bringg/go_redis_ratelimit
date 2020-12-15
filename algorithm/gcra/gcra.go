// https://github.com/go-redis/redis_rate
package gcra

import (
	"context"
	"strconv"
	"time"

	"github.com/bringg/go_redis_ratelimit/algorithm"
)

const AlgorithmName = "gcra"

type GCRA struct {
	Limit algorithm.Limit
	RDB   algorithm.Rediser

	key string
}

// Allow is shorthand for AllowN(key, 1).
func (c *GCRA) Allow() (*algorithm.Result, error) {
	return c.AllowN(1)
}

// SetKey _
func (c *GCRA) SetKey(key string) {
	c.key = key
}

// AllowN reports whether n events may happen at time now.
func (c *GCRA) AllowN(n int) (*algorithm.Result, error) {
	limit := c.Limit
	values := []interface{}{limit.GetBurst(), limit.GetRate(), limit.GetPeriod().Seconds(), n}

	v, err := script.Run(context.Background(), c.RDB, []string{c.key}, values...).Result()
	if err != nil {
		return nil, err
	}

	values = v.([]interface{})

	retryAfter, err := strconv.ParseFloat(values[2].(string), 64)
	if err != nil {
		return nil, err
	}

	resetAfter, err := strconv.ParseFloat(values[3].(string), 64)
	if err != nil {
		return nil, err
	}

	res := &algorithm.Result{
		Limit:      limit,
		Key:        c.key,
		Allowed:    values[0].(int64) == 0,
		Remaining:  values[1].(int64),
		RetryAfter: dur(retryAfter),
		ResetAfter: dur(resetAfter),
	}
	return res, nil
}

func dur(f float64) time.Duration {
	if f == -1 {
		return -1
	}

	return time.Duration(f * float64(time.Second))
}
