package go_redis_ratelimit

import "time"

type (
	Limit struct {
		Algorithm string
		Burst     int64
		Rate      int64
		Period    time.Duration
	}
)

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
