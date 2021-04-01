package algorithm

import (
	"context"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
)

var (
	Registry []*RegInfo
)

type (
	Limit interface {
		GetRate() int64
		GetBurst() int64
		GetAlgorithm() string
		GetPeriod() time.Duration
	}

	Rediser interface {
		TxPipeline() redis.Pipeliner
		Incr(ctx context.Context, key string) *redis.IntCmd
		ZCard(ctx context.Context, key string) *redis.IntCmd
		Get(ctx context.Context, key string) *redis.StringCmd
		Del(ctx context.Context, keys ...string) *redis.IntCmd
		ScriptLoad(ctx context.Context, script string) *redis.StringCmd
		ScriptExists(ctx context.Context, hashes ...string) *redis.BoolSliceCmd
		ZAdd(ctx context.Context, key string, members ...*redis.Z) *redis.IntCmd
		Expire(ctx context.Context, key string, expiration time.Duration) *redis.BoolCmd
		ZRemRangeByScore(ctx context.Context, key string, min string, max string) *redis.IntCmd
		Eval(ctx context.Context, script string, keys []string, args ...interface{}) *redis.Cmd
		EvalSha(ctx context.Context, sha1 string, keys []string, args ...interface{}) *redis.Cmd
		TxPipelined(ctx context.Context, fn func(pipe redis.Pipeliner) error) ([]redis.Cmder, error)
		ZRangeByScoreWithScores(ctx context.Context, key string, opt *redis.ZRangeBy) *redis.ZSliceCmd
	}

	Algorithm interface {
		Allow(key string, limit Limit) (*Result, error)
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

	RegInfo struct {
		Name         string
		NewAlgorithm func(rdb Rediser) (Algorithm, error)
	}
)

func Register(info *RegInfo) {
	Registry = append(Registry, info)
}

func Find(name string) (*RegInfo, error) {
	for _, item := range Registry {
		if item.Name == name {
			return item, nil
		}
	}

	return nil, fmt.Errorf("didn't find algorithm called %q", name)
}
