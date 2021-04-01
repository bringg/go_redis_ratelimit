package go_redis_ratelimit

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/go-redis/redis/v8"
	"github.com/stretchr/testify/assert"

	"github.com/bringg/go_redis_ratelimit/algorithm/cloudflare"
	"github.com/bringg/go_redis_ratelimit/algorithm/gcra"
	"github.com/bringg/go_redis_ratelimit/algorithm/sliding_window"
)

var limiter = rateLimiter()

func rateLimiter() *Limiter {
	mr, err := miniredis.Run()
	if err != nil {
		panic(err)
	}

	client := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})

	if err := client.FlushDB(context.Background()).Err(); err != nil {
		panic(err)
	}

	l, err := NewLimiter(client)
	if err != nil {
		panic(err)
	}

	return l
}

func TestLimiter_Allow(t *testing.T) {
	l := rateLimiter()

	limit := &Limit{
		Algorithm: sliding_window.AlgorithmName,
		Rate:      10,
		Period:    time.Minute,
		Burst:     10,
	}

	t.Run("sliding_window", func(t *testing.T) {
		res, err := l.Allow("test_me"+t.Name(), limit)

		assert.Nil(t, err)
		assert.True(t, res.Allowed)
		assert.Equal(t, int64(9), res.Remaining)
	})

	t.Run("gcra", func(t *testing.T) {
		limit.Algorithm = gcra.AlgorithmName

		res, err := l.Allow("test_me"+t.Name(), limit)

		assert.Nil(t, err)
		assert.True(t, res.Allowed)
		assert.Equal(t, int64(9), res.Remaining)
		assert.Equal(t, res.RetryAfter, time.Duration(-1))
	})

	t.Run("cloudflare", func(t *testing.T) {
		limit.Algorithm = cloudflare.AlgorithmName

		res, err := l.Allow("test_me"+t.Name(), limit)

		assert.Nil(t, err)
		assert.True(t, res.Allowed)
		assert.Equal(t, int64(9), res.Remaining)
		assert.Equal(t, res.RetryAfter, time.Duration(0))
	})
}

func Benchmark_CloudflareAlgorithm(b *testing.B) {
	for i := 0; i < b.N; i++ {
		limiter.Allow("cloudflare", &Limit{
			Algorithm: cloudflare.AlgorithmName,
			Rate:      2000,
			Period:    60 * time.Second,
		})
	}
}

func Benchmark_GcraAlgorithm(b *testing.B) {
	for i := 0; i < b.N; i++ {
		limiter.Allow("gcra", &Limit{
			Algorithm: gcra.AlgorithmName,
			Rate:      2000,
			Burst:     2000,
			Period:    10 * time.Second,
		})
	}
}

func Benchmark_SlidingWindowAlgorithm(b *testing.B) {
	for i := 0; i < b.N; i++ {
		limiter.Allow("sliding_window", &Limit{
			Algorithm: sliding_window.AlgorithmName,
			Rate:      2000,
			Period:    10 * time.Second,
		})
	}
}
