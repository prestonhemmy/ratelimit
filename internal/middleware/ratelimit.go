package middleware

import (
	"log"
	"net"
	"net/http"

	"github.com/prestonhemmy/ratelimit/internal/ratelimiter"
)

// Middleware that calls rate limiter.

func RateLimitMiddleware(limiter ratelimiter.RateLimiter) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// extract client IP
			host, _, err := net.SplitHostPort(r.RemoteAddr)
			if err != nil {
				host = r.RemoteAddr
			}

			// invoke rate limiter
			allowed, err := limiter.Allow(host)

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
