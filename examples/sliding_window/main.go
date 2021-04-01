package main

import (
	"log"
	"time"

	"github.com/go-redis/redis/v8"

	"github.com/bringg/go_redis_ratelimit"
	"github.com/bringg/go_redis_ratelimit/algorithm/sliding_window"
)

func main() {
	option, err := redis.ParseURL("redis://127.0.0.1:6379/0")
	if err != nil {
		log.Fatal(err)
	}
	client := redis.NewClient(option)

	limiter, err := go_redis_ratelimit.NewLimiter(client)
	if err != nil {
		log.Fatal(err)
	}

	res, err := limiter.Allow("api_gateway:klu4ik", &go_redis_ratelimit.Limit{
		Algorithm: sliding_window.AlgorithmName,
		Rate:      10,
		Period:    2 * time.Minute,
	})

	if err != nil {
		log.Fatal(err)
	}

	log.Println("===> ", res.Allowed, res.Remaining, res)
	// Output: true 1
}
