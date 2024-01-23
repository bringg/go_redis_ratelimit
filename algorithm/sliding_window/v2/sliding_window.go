package sliding_window

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/bringg/go_redis_ratelimit/algorithm"
)

const AlgorithmName = "sliding_window_v2"

type SlidingWindow struct {
	Limit algorithm.Limit
	RDB   algorithm.Rediser

	key string
}

func (c *SlidingWindow) SetKey(key string) {
	c.key = key
}

func (c *SlidingWindow) Allow() (r *algorithm.Result, err error) {
	limit := c.Limit
	now := time.Now()
	nowNanos := now.UnixNano()
	clearBefore := now.Add(-limit.GetPeriod()).UnixNano()
	ctx := context.Background()

	count, err := c.allowCheckCard(ctx, c.key, clearBefore)
	if err != nil {
		return nil, fmt.Errorf("Failed to get redis cardinality: %v", err)
	}

	resetAfter := limit.GetPeriod()
	rate := limit.GetRate() - 1

	// we increase later with ZAdd, so max-1
	if err == nil && count >= rate {
		return &algorithm.Result{
			Limit:      limit,
			Allowed:    false,
			Key:        c.key,
			ResetAfter: resetAfter,
			RetryAfter: c.retryAfter(ctx),
			Remaining:  0,
		}, nil
	}

	pipe := c.RDB.TxPipeline()
	if err := pipe.ZAdd(
		ctx,
		c.key,
		redis.Z{
			Member: nowNanos,
			Score:  float64(nowNanos),
		},
	).Err(); err != nil {
		return nil, fmt.Errorf("Failed to ZAdd proceeding with Expire: %v", err)
	}

	if err := pipe.Expire(ctx, c.key, limit.GetPeriod()+time.Second).Err(); err != nil {
		return nil, fmt.Errorf("Failed to Expire: %v", err)
	}

	if _, err := pipe.Exec(ctx); err != nil {
		return nil, fmt.Errorf("Failed to Expire: %v", err)
	}

	return &algorithm.Result{
		Limit:      limit,
		Allowed:    true,
		Key:        c.key,
		ResetAfter: resetAfter,
		RetryAfter: 0,
		Remaining:  rate - count,
	}, nil
}

// It will use ZRANGEBYSCORE with offset 0 and count 1 to get the
// oldest item stored in redis.
func (c *SlidingWindow) oldest(ctx context.Context) (time.Duration, error) {
	now := time.Now()

	res := c.RDB.ZRangeByScoreWithScores(ctx, c.key, &redis.ZRangeBy{
		Min:    "0.0",
		Max:    fmt.Sprint(float64(now.UnixNano())),
		Offset: 0,
		Count:  1,
	})

	zs, err := res.Result()
	if err != nil {
		return 0, err
	}

	if len(zs) == 0 {
		log.Printf("Oldest() got no valid data: %v", res)

		return 0, nil
	}

	z := zs[0]
	s, ok := z.Member.(string)
	if !ok {
		return 0, errors.New("failed to evaluate redis data")
	}

	oldest, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("failed to convert value to float64: %w", err)
	}

	return time.Duration(oldest), nil
}

func (c *SlidingWindow) allowCheckCard(ctx context.Context, key string, clearBefore int64) (int64, error) {
	// drop all elements of the set which occurred before one interval ago.
	pipe := c.RDB.TxPipeline()
	if err := pipe.ZRemRangeByScore(ctx, key, "0.0", fmt.Sprint(float64(clearBefore))).Err(); err != nil {
		return 0, fmt.Errorf("zremrangebyscore: %w", err)
	}

	// get cardinality
	res := pipe.ZCard(ctx, key)
	if err := res.Err(); err != nil {
		return 0, fmt.Errorf("zcard: %w", err)
	}

	if _, err := pipe.Exec(ctx); err != nil {
		return 0, fmt.Errorf("exec: %w", err)
	}

	return res.Val(), nil
}

func (c *SlidingWindow) deltaFrom(ctx context.Context, from time.Time) (time.Duration, error) {
	oldest, err := c.oldest(ctx)
	if err != nil {
		return 0, err
	}

	gap := from.Sub(time.Unix(0, int64(oldest)))
	return c.Limit.GetPeriod() - gap, nil
}

func (c *SlidingWindow) retryAfter(ctx context.Context) time.Duration {
	// If less than 1s to wait -> so set to 1
	const minWait = 1

	now := time.Now()

	d, err := c.deltaFrom(ctx, now)
	if err != nil {
		log.Printf("Failed to get the duration until the next call is allowed: %v", err)

		return minWait
	}

	res := d / time.Second
	if res > 0 {
		return res + 1
	}

	return minWait
}
