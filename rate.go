package go_redis_ratelimit

import (
	"errors"
	"time"

	"github.com/go-redis/redis/v8"

	"github.com/bringg/go_redis_ratelimit/algorithm"
	"github.com/bringg/go_redis_ratelimit/algorithm/cloudflare"
	"github.com/bringg/go_redis_ratelimit/algorithm/gcra"
	"github.com/bringg/go_redis_ratelimit/algorithm/sliding_window"
	sliding_windowV2 "github.com/bringg/go_redis_ratelimit/algorithm/sliding_window/v2"
)

const (
	DefaultPrefix = "limiter"
)

type (
	Algorithm interface {
		Allow() (*algorithm.Result, error)
		SetKey(string)
	}

	Limit struct {
		Algorithm string
		Burst     int64
		Rate      int64
		Period    time.Duration
	}

	// Limiter controls how frequently events are allowed to happen.
	Limiter struct {
		rdb    *redis.Client
		Prefix string
	}
)

// NewLimiter returns a new Limiter.
func NewLimiter(rdb *redis.Client) *Limiter {
	return &Limiter{
		rdb:    rdb,
		Prefix: DefaultPrefix,
	}
}

func (l *Limiter) Allow(key string, limit *Limit) (*algorithm.Result, error) {
	var algo Algorithm

	switch limit.Algorithm {
	case sliding_windowV2.AlgorithmName:
		algo = &sliding_windowV2.SlidingWindow{Limit: limit, RDB: l.rdb}
	case sliding_window.AlgorithmName:
		algo = &sliding_window.SlidingWindow{Limit: limit, RDB: l.rdb}
	case cloudflare.AlgorithmName:
		algo = &cloudflare.Cloudflare{Limit: limit, RDB: l.rdb}
	case gcra.AlgorithmName:
		algo = &gcra.GCRA{Limit: limit, RDB: l.rdb}
	default:
		return nil, errors.New("algorithm is not supported")
	}

	algo.SetKey(l.Prefix + ":" + limit.Algorithm + ":" + key)

	return algo.Allow()
}

func (l *Limit) GetAlgorithm() string {
	return l.Algorithm
}

func (l *Limit) GetBurst() int64 {
	return l.Burst
}

func (l *Limit) GetRate() int64 {
	return l.Rate
}

func (l *Limit) GetPeriod() time.Duration {
	return l.Period
}
