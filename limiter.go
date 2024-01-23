package go_redis_ratelimit

import (
	"errors"

	"github.com/redis/go-redis/v9"

	"github.com/bringg/go_redis_ratelimit/algorithm"
	_ "github.com/bringg/go_redis_ratelimit/algorithm/all"
)

const DefaultPrefix = "limiter"

// Limiter controls how frequently events are allowed to happen.
type Limiter struct {
	Prefix     string
	algorithms map[string]algorithm.Algorithm
}

// NewLimiter returns a new Limiter.
func NewLimiter(rdb *redis.Client) (*Limiter, error) {
	algorithms := make(map[string]algorithm.Algorithm, len(algorithm.Registry))

	var err error
	for _, info := range algorithm.Registry {
		if algorithms[info.Name], err = info.NewAlgorithm(rdb); err != nil {
			return nil, err
		}
	}

	return &Limiter{
		Prefix:     DefaultPrefix,
		algorithms: algorithms,
	}, nil
}

func (l *Limiter) Allow(key string, limit *Limit) (*algorithm.Result, error) {
	algo, err := l.findAlgorithm(limit.Algorithm)
	if err != nil {
		return nil, err
	}

	return algo.Allow(
		l.Prefix+":"+limit.Algorithm+":"+key,
		limit,
	)
}

func (l *Limiter) findAlgorithm(name string) (algorithm.Algorithm, error) {
	if algo, ok := l.algorithms[name]; ok {
		return algo, nil
	}

	return nil, errors.New("algorithm is not supported")
}
