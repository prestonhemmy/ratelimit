package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/prestonhemmy/ratelimit/internal/admin"
	"github.com/prestonhemmy/ratelimit/internal/config"
	"github.com/prestonhemmy/ratelimit/internal/middleware"
	"github.com/prestonhemmy/ratelimit/internal/proxy"
	"github.com/prestonhemmy/ratelimit/internal/ratelimiter"
	"github.com/redis/go-redis/v9"
)

// Entry point for the rate limiting API gateway.
// Loads config, connects to Redis and starts the HTTP server with middleware
// chain logging -> rate limiting -> reverse proxy.

func main() {
	// load config
	cfg, err := config.Load("configs/config.yaml")
	if err != nil {
		log.Fatal(err)
	}

	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = cfg.Redis.Addr
	}

	redisClient := redis.NewClient(&redis.Options{
		Addr:     redisAddr,
		Password: "",
		DB:       0,
	})

	// create fixed window rate limiter
	//limiter := ratelimiter.NewFixedWindowLimiter(redisClient)

	// create sliding window rate limiter
	limiter := ratelimiter.NewSlidingWindowLimiter(redisClient)

	// create reverse proxy
	revProxy := proxy.NewProxy(cfg.Backend.Url)

	// middleware chain (logging -> rate limiter -> proxy)
	rateLimiterHandler := middleware.RateLimitMiddleware(limiter, cfg)(revProxy)
	handler := middleware.LoggingMiddleware(rateLimiterHandler)

	// admin stats endpoint (logging -> admin
	adminHandler := admin.NewAdminHandler(redisClient, cfg)
	http.Handle("/admin/stats", middleware.LoggingMiddleware(adminHandler))

	// o.w. handled by rate limiting proxy
	http.Handle("/", handler)

	// initialize server
	fmt.Printf("Starting HTTP server on port %d\n", cfg.Server.Port)

	addr := fmt.Sprintf(":%d", cfg.Server.Port)

	// column titles for log output
	fmt.Printf(" %-20s %-4s %-14s %-21s %s\n",
		"TIMESTAMP", "CODE", "LATENCY", "CLIENT", "REQUEST",
	)

	if err = http.ListenAndServe(addr, nil); err != nil {
		log.Fatal(err)
	}
}
