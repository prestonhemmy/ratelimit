package ratelimiter

import (
	"context"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

// Defines the RateLimiter interface and provides two Redis-backed
// implementations: fixed window and sliding window counters.

var slidingWindowScript = redis.NewScript(`
	-- GET previous window count
	local prev = redis.call('GET', KEYS[2])
	if prev == false then
		prev = 0
	else
		prev = tonumber(prev)
	end
	
	-- INCR current window counter
	local curr = redis.call('INCR', KEYS[1])
	
	-- set expiration if first request in window
	if curr == 1 then
		redis.call('EXPIRE', KEYS[1], ARGV[1])
	end
	
	return {prev, curr}
`)

type RateLimiter interface {
	Allow(key string, limit int, window time.Duration) (bool, error)
}

// Fixed Window

type FixedWindowLimiter struct {
	client *redis.Client
}

func NewFixedWindowLimiter(client *redis.Client) *FixedWindowLimiter {
	return &FixedWindowLimiter{client: client}
}

func (l *FixedWindowLimiter) Allow(
	key string,
	limit int,
	window time.Duration,
) (bool, error) {
	ctx := context.Background()

	windowID := time.Now().Unix() / int64(window.Seconds())
	key = "ratelimit:" + key + ":" + strconv.Itoa(int(windowID))

	// increment the key
	cnt, err := l.client.Incr(ctx, key).Result()
	if err != nil {
		return false, err
	}

	// check if previous window expired
	if cnt == 1 {
		if err := l.client.Expire(ctx, key, window).Err(); err != nil {
			return false, err
		}
	}

	// check if at capacity
	return cnt <= int64(limit), nil
}

// Sliding Window

type SlidingWindowLimiter struct {
	client *redis.Client
}

func NewSlidingWindowLimiter(client *redis.Client) *SlidingWindowLimiter {
	return &SlidingWindowLimiter{client: client}
}

func (l *SlidingWindowLimiter) Allow(
	key string,
	limit int,
	window time.Duration,
) (bool, error) {
	ctx := context.Background()
	now := time.Now().Unix()
	windowDuration := int64(window.Seconds())

	currWindowID := now / windowDuration
	prevWindowID := currWindowID - 1
	elapsedTime := now % windowDuration
	elapsedFrac := float64(elapsedTime) / float64(windowDuration)

	// build two Redis keys from curr window ID and prev window ID
	currKey := "ratelimit:" + key + ":" + strconv.Itoa(int(currWindowID))
	prevKey := "ratelimit:" + key + ":" + strconv.Itoa(int(prevWindowID))

	// pipeline read-increment-expire sequence to ensure atomicity
	res, err := slidingWindowScript.Run(
		ctx, l.client, []string{currKey, prevKey}, int(2*window.Seconds()),
	).Int64Slice()

	if err != nil {
		return false, err
	}

	prevCount := res[0]
	currCount := res[1]

	weightedCount := float64(prevCount)*(1-elapsedFrac) + float64(currCount)

	// check if at capacity
	return weightedCount <= float64(limit), nil
}
