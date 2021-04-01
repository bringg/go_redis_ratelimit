package main

import (
	"log"
	"time"

	"github.com/go-redis/redis/v8"

	limiter "github.com/bringg/go_redis_ratelimit"
	"github.com/bringg/go_redis_ratelimit/algorithm/cloudflare"
)

func main() {
	option, err := redis.ParseURL("redis://127.0.0.1:6379/0")
	if err != nil {
		log.Fatal(err)
	}

	client := redis.NewClient(option)
	l, err := limiter.NewLimiter(client)
	if err != nil {
		log.Fatal(err)
	}

	res, err := l.Allow("api_gateway:klu4ik", &limiter.Limit{
		Algorithm: cloudflare.AlgorithmName,
		Rate:      10,
		Period:    10 * time.Second,
	})

	if err != nil {
		log.Fatal(err)
	}

	log.Println("===> ", res.Allowed, res.Remaining, res)
	// Output: true 1
}
