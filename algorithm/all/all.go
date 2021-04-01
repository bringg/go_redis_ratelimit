package all

import (
	_ "github.com/bringg/go_redis_ratelimit/algorithm/cloudflare"
	_ "github.com/bringg/go_redis_ratelimit/algorithm/gcra"
	_ "github.com/bringg/go_redis_ratelimit/algorithm/sliding_window"
)
