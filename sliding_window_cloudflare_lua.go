package go_redis_ratelimit

import "github.com/go-redis/redis/v7"

var script3 = redis.NewScript(`
-- this script has side-effects, so it requires replicate commands mode
redis.replicate_commands()

local key = KEYS[1]
local math_floor = math.floor
local now = redis.call("TIME")[1]

local function redis_rate ()
  local current_time = math_floor(now)
  local current_second = current_time % 60
  local current_minute = math_floor(current_time / 60) % 60
  local past_minute = (current_minute + 59) % 60
  local current_key = key .. "_" .. current_minute
  local past_key = key .. "_" .. past_minute

  local first_resp = redis.call("GET", past_key)
  local second_resp = redis.call("INCR", current_key)

  redis.call("EXPIRE", current_key, 2 * 60 - current_second)

  if not first_resp then
    first_resp  = "0"
  end
  local past_counter = tonumber(first_resp)
  local current_counter = tonumber(second_resp) - 1

  -- strongly inspired by https://blog.cloudflare.com/counting-things-a-lot-of-different-things/
  local current_rate = past_counter * ((60 - (current_time % 60)) / 60) + current_counter
  return current_rate
end

local rate = redis_rate()

return {rate+1}
`)
