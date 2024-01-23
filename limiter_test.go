package go_redis_ratelimit

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/bringg/go_redis_ratelimit/algorithm/cloudflare"
	"github.com/bringg/go_redis_ratelimit/algorithm/gcra"
	"github.com/bringg/go_redis_ratelimit/algorithm/sliding_window"
)

func rateLimiter(addr string) *Limiter {
	client := redis.NewClient(&redis.Options{
		Addr: addr,
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
	mr, err := miniredis.Run()
	require.NoError(t, err)

	defer mr.Close()

	l := rateLimiter(mr.Addr())

	limit := &Limit{
		Algorithm: sliding_window.AlgorithmName,
		Rate:      10,
		Period:    time.Minute,
		Burst:     10,
	}

	t.Run("sliding_window", func(t *testing.T) {
		res, err := l.Allow("test_me"+t.Name(), limit)

		require.NoError(t, err)
		assert.True(t, res.Allowed)
		assert.Equal(t, int64(9), res.Remaining)
	})

	t.Run("gcra", func(t *testing.T) {
		limit.Algorithm = gcra.AlgorithmName

		res, err := l.Allow("test_me"+t.Name(), limit)

		require.NoError(t, err)
		assert.True(t, res.Allowed)
		assert.Equal(t, int64(9), res.Remaining)
		assert.Equal(t, res.RetryAfter, time.Duration(-1))
	})

	t.Run("cloudflare", func(t *testing.T) {
		limit.Algorithm = cloudflare.AlgorithmName

		res, err := l.Allow("test_me"+t.Name(), limit)

		if err != nil {
			t.Fatalf("err: %+v", err)
		}

		require.NoError(t, err)
		assert.True(t, res.Allowed)
		assert.Equal(t, int64(9), res.Remaining)
		assert.Equal(t, res.RetryAfter, time.Duration(0))
	})
}

// In benchmarks we will use real redis-server, since mini-redis impacts the benchmarks.
const redisAddress = "localhost:6379"

func Benchmark_CloudflareAlgorithm(b *testing.B) {
	limiter := rateLimiter(redisAddress)

	for i := 0; i < b.N; i++ {
		if _, err := limiter.Allow("cloudflare", &Limit{
			Algorithm: cloudflare.AlgorithmName,
			Rate:      2000,
			Period:    60 * time.Second,
		}); err != nil {
			b.Fatalf("Failed to check ratelimit: %+v", err)
		}
	}
}

func Benchmark_GcraAlgorithm(b *testing.B) {
	limiter := rateLimiter(redisAddress)

	for i := 0; i < b.N; i++ {
		if _, err := limiter.Allow("gcra", &Limit{
			Algorithm: gcra.AlgorithmName,
			Rate:      2000,
			Burst:     2000,
			Period:    10 * time.Second,
		}); err != nil {
			b.Fatalf("Failed to check ratelimit: %+v", err)
		}
	}
}

func Benchmark_SlidingWindowAlgorithm(b *testing.B) {
	limiter := rateLimiter(redisAddress)

	for i := 0; i < b.N; i++ {
		if _, err := limiter.Allow("sliding_window", &Limit{
			Algorithm: sliding_window.AlgorithmName,
			Rate:      2000,
			Period:    10 * time.Second,
		}); err != nil {
			b.Fatalf("Failed to check ratelimit: %+v", err)
		}
	}
}
