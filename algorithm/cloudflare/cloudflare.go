package cloudflare

import (
	"time"

	"github.com/coinpaprika/ratelimiter"

	"github.com/bringg/go_redis_ratelimit/algorithm"
)

const AlgorithmName = "cloudflare"

type Cloudflare struct {
	RDB algorithm.Rediser
}

func init() {
	algorithm.Register(&algorithm.RegInfo{
		Name:         AlgorithmName,
		NewAlgorithm: NewAlgorithm,
	})
}

func NewAlgorithm(rdb algorithm.Rediser) (algorithm.Algorithm, error) {
	return &Cloudflare{
		RDB: rdb,
	}, nil
}

func (c *Cloudflare) Allow(key string, limit algorithm.Limit) (*algorithm.Result, error) {
	rate := limit.GetRate() - 1
	rateLimiter := ratelimiter.New(&RedisDataStore{
		RDB:            c.RDB,
		ExpirationTime: 2 * limit.GetPeriod(),
	}, rate, limit.GetPeriod())

	limitStatus, err := rateLimiter.Check(key)
	if err != nil {
		return nil, err
	}

	rateKey := mapKey(key, time.Now().UTC().Truncate(limit.GetPeriod()))
	currentRate := int64(limitStatus.CurrentRate)

	if limitStatus.IsLimited {
		return &algorithm.Result{
			Limit:      limit,
			Key:        rateKey,
			Allowed:    false,
			Remaining:  0,
			RetryAfter: *limitStatus.LimitDuration,
			ResetAfter: limit.GetPeriod(),
		}, nil
	}

	if err := rateLimiter.Inc(key); err != nil {
		return nil, err
	}

	return &algorithm.Result{
		Limit:      limit,
		Key:        rateKey,
		Allowed:    true,
		Remaining:  rate - currentRate,
		RetryAfter: 0,
		ResetAfter: limit.GetPeriod(),
	}, nil
}
