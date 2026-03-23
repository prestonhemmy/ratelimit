package ratelimiter

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

// File housing rate limiter interface definitions and Redis implementation.

type RateLimiter interface {
	Allow(key string) (bool, error)
}

type FixedWindowLimiter struct {
	client         *redis.Client
	maxRequests    int
	windowDuration time.Duration
}

func NewFixedWindowLimiter(
	client *redis.Client,
	limit int,
	window time.Duration,
) *FixedWindowLimiter {
	return &FixedWindowLimiter{
		client:         client,
		maxRequests:    limit,
		windowDuration: window,
	}
}

func (limiter *FixedWindowLimiter) Allow(key string) (bool, error) {
	// add prefix
	key = "ratelimit:" + key

	// increment the key
	ctx := context.Background()
	cnt, err := limiter.client.Incr(ctx, key).Result()
	if err != nil {
		return false, err
	}

	// check if previous window expired
	if cnt == 1 {
		res := limiter.client.Expire(ctx, key, limiter.windowDuration)
		if res.Err() != nil {
			return false, res.Err()
		}
	}

	// check if at capacity
	if cnt > int64(limiter.maxRequests) {
		return false, nil
	}

	return true, nil
}
