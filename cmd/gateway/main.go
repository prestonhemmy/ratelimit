package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/prestonhemmy/ratelimit/internal/config"
	"github.com/prestonhemmy/ratelimit/internal/middleware"
	"github.com/prestonhemmy/ratelimit/internal/proxy"
	"github.com/prestonhemmy/ratelimit/internal/ratelimiter"
	"github.com/redis/go-redis/v9"
)

// Entry point that loads config, creates a handler for incoming HTTP requests
// and starts the server.

func main() {
	// load config
	cfg, err := config.Load("configs/config.yaml")
	if err != nil {
		log.Fatal(err)
	}

	// create reverse proxy
	revProxy := proxy.NewProxy(cfg.Backend.Url)

	redisClient := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})

	// create fixed-window rate limiter
	limiter := ratelimiter.NewFixedWindowLimiter(redisClient, 10, time.Minute)

	// pass limiter to middleware then reverse proxy (Why?)
	handler := middleware.RateLimitMiddleware(limiter)(revProxy)

	http.Handle("/", handler)

	// initialize server
	_, err = fmt.Printf("Starting HTTP server on port %d\n", cfg.Server.Port)
	if err != nil {
		log.Fatal(err)
	}

	addr := fmt.Sprintf(":%d", cfg.Server.Port)
	if err = http.ListenAndServe(addr, nil); err != nil {
		log.Fatal(err)
	}
}
