package cloudflare

import (
	"time"

	"github.com/coinpaprika/ratelimiter"

	"github.com/bringg/go_redis_ratelimit/algorithm"
)

const AlgorithmName = "cloudflare"

type Cloudflare struct {
	Limit algorithm.Limit
	RDB   algorithm.Rediser

	key string
}

func (c *Cloudflare) Allow() (*algorithm.Result, error) {
	rate := c.Limit.GetRate() - 1
	rateLimiter := ratelimiter.New(&RedisDataStore{
		RDB:            c.RDB,
		ExpirationTime: 2 * c.Limit.GetPeriod(),
	}, rate, c.Limit.GetPeriod())

	limitStatus, err := rateLimiter.Check(c.key)
	if err != nil {
		return nil, err
	}

	rateKey := mapKey(c.key, time.Now().UTC().Truncate(c.Limit.GetPeriod()))
	currentRate := int64(limitStatus.CurrentRate)

	if limitStatus.IsLimited {
		return &algorithm.Result{
			Limit:      c.Limit,
			Key:        rateKey,
			Allowed:    false,
			Remaining:  0,
			RetryAfter: *limitStatus.LimitDuration,
			ResetAfter: c.Limit.GetPeriod(),
		}, nil
	}

	if err := rateLimiter.Inc(c.key); err != nil {
		return nil, err
	}

	return &algorithm.Result{
		Limit:      c.Limit,
		Key:        rateKey,
		Allowed:    true,
		Remaining:  rate - currentRate,
		RetryAfter: 0,
		ResetAfter: c.Limit.GetPeriod(),
	}, nil
}

func (c *Cloudflare) SetKey(key string) {
	c.key = key
}
