package ratelimit

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
)

const luaScript = `
local key = KEYS[1]
local limit = tonumber(ARGV[1])
local val = redis.call('incr', key)
if val == 1 then
  redis.call('expire', key, 60)
end
if val > limit then
  return {0, val - 1}
end
return {1, limit - val}
`

func CheckLimit(ctx context.Context, redisClient *redis.Client, userID string, limit int) (allowed bool, remaining int, err error) {
	key := fmt.Sprintf("ratelimit:%s:executions", userID)
	script := redis.NewScript(luaScript)

	result, err := script.Run(ctx, redisClient, []string{key}, limit).Result()
	if err != nil {
		return false, 0, err
	}

	values := result.([]interface{})
	allowed = values[0].(int64) == 1
	remaining = int(values[1].(int64))

	return allowed, remaining, nil
}
