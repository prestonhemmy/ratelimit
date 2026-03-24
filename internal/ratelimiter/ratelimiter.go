package ratelimiter

import (
	"context"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

// File housing rate limiter interface definitions and Redis implementations
// for fixed and sliding window rate limit strategies

type RateLimiter interface {
	Allow(key string, limit int, window time.Duration) (bool, error)
}

// Fixed Window Strategy

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

// Sliding Window Strategy

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

	// pipeline redis commands to reduce latency
	pipe := l.client.Pipeline()
	prevCmd := pipe.Get(ctx, prevKey)
	currCmd := pipe.Incr(ctx, currKey)
	_, err := pipe.Exec(ctx)

	// ignore 'redis.Nil' errors (GET fails on nonexistent prev window)
	// o.w. validate INCR command
	if err != nil && err != redis.Nil {
		if currCmd.Err() != nil {
			return false, currCmd.Err()
		}
	}

	// extract prev window count if key exists (o.w. defaults to zero)
	var prevCount int64
	prevVal, err := prevCmd.Result()
	if err == nil {
		prevCount, _ = strconv.ParseInt(prevVal, 10, 64)
	}

	currCount := currCmd.Val()

	// if new window then double expiration time so next window has access
	// to previous window across the entire window duration
	if currCount == 1 {
		if err := l.client.Expire(ctx, currKey, 2*window).Err(); err != nil {
			return false, err
		}
	}

	weightedCount := float64(prevCount)*(1-elapsedFrac) + float64(currCount)

	// check if at capacity
	return weightedCount <= float64(limit), nil
}
