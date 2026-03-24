package middleware

import (
	"log"
	"net"
	"net/http"
	"time"

	"github.com/prestonhemmy/ratelimit/internal/config"
	"github.com/prestonhemmy/ratelimit/internal/ratelimiter"
)

// Middleware that calls rate limiter.

func RateLimitMiddleware(
	limiter ratelimiter.RateLimiter,
	cfg *config.Config,
) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// early exit if rate limiting disabled
			if !cfg.RateLimit.Enabled {
				next.ServeHTTP(w, r)
				return
			}

			// extract client IP
			host, _, err := net.SplitHostPort(r.RemoteAddr)
			if err != nil {
				host = r.RemoteAddr
			}

			path := r.URL.Path
			key := host + ":" + path
			requests, windowSeconds := cfg.RuleForPath(path)
			window := time.Duration(windowSeconds) * time.Second

			// invoke fixed or sliding (default) rate limiter
			allowed, err := limiter.Allow(key, requests, window)

			// if error from Redis then fail open (let through and log error)
			if err != nil {
				next.ServeHTTP(w, r)
				log.Printf("rate limiter error: %v", err)
				return
			}

			// if not allowed, send 429 HTTP error message
			if !allowed {
				http.Error(w, "rate limit exceeded", http.StatusTooManyRequests)
				return
			}

			// o.w. forward to next handler
			next.ServeHTTP(w, r)
		})
	}
}
