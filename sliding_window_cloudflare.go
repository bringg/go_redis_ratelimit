package go_redis_ratelimit

import "time"

const SlidingWindowCloudflareAlgorithmName = "sliding_window_cloudflare"

type slidingWindoCloudflare struct {
	key   string
	limit *Limit
	rdb   rediser
}

func (c *slidingWindoCloudflare) SetKey(key string) {
	c.key = key
}

func (c *slidingWindoCloudflare) Allow() (r *Result, err error) {
	limit := c.limit

	v, err := script3.Run(c.rdb, []string{c.key}).Result()
	if err != nil {
		return nil, err
	}

	values := v.([]interface{})

	allowed := true
	rate := values[0].(int64)
	if rate > limit.Rate {
		allowed = false
	}

	return &Result{
		Limit:      limit,
		Key:        c.key,
		Allowed:    allowed,
		Remaining:  rate,
		RetryAfter: time.Minute,
		ResetAfter: limit.Period,
	}, nil
}
