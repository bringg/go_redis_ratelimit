package algorithm

import (
	"context"
	"time"

	"github.com/go-redis/redis/v8"
)

type (
	Limit interface {
		GetAlgorithm() string
		GetBurst() int64
		GetRate() int64
		GetPeriod() time.Duration
	}

	Rediser interface {
		TxPipeline() redis.Pipeliner
		TxPipelined(ctx context.Context, fn func(pipe redis.Pipeliner) error) ([]redis.Cmder, error)
		Get(ctx context.Context, key string) *redis.StringCmd
		Incr(ctx context.Context, key string) *redis.IntCmd
		Eval(ctx context.Context, script string, keys []string, args ...interface{}) *redis.Cmd
		EvalSha(ctx context.Context, sha1 string, keys []string, args ...interface{}) *redis.Cmd
		ScriptExists(ctx context.Context, hashes ...string) *redis.BoolSliceCmd
		ScriptLoad(ctx context.Context, script string) *redis.StringCmd
		ZRangeByScoreWithScores(ctx context.Context, key string, opt *redis.ZRangeBy) *redis.ZSliceCmd
		ZRemRangeByScore(ctx context.Context, key string, min string, max string) *redis.IntCmd
		ZCard(ctx context.Context, key string) *redis.IntCmd
		ZAdd(ctx context.Context, key string, members ...*redis.Z) *redis.IntCmd
		Expire(ctx context.Context, key string, expiration time.Duration) *redis.BoolCmd
	}

	Result struct {
		// Limit is the limit that was used to obtain this result.
		Limit Limit

		// Key is the key of limit
		Key string

		// Allowed reports whether event may happen at time now.
		Allowed bool

		// Remaining is the maximum number of requests that could be
		// permitted instantaneously for this key given the current
		// state. For example, if a rate limiter allows 10 requests per
		// second and has already received 6 requests for this key this
		// second, Remaining would be 4.
		Remaining int64

		// RetryAfter is the time until the next request will be permitted.
		// It should be -1 unless the rate limit has been exceeded.
		RetryAfter time.Duration

		// ResetAfter is the time until the RateLimiter returns to its
		// initial state for a given key. For example, if a rate limiter
		// manages requests per second and received one request 200ms ago,
		// Reset would return 800ms. You can also think of this as the time
		// until Limit and Remaining will be equal.
		ResetAfter time.Duration
	}
)
