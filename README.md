# Ratelimit

A lightweight rate limiting API gateway built in Go with Redis.

![reverse-proxy.png](reverse-proxy.png)

Sits as a reverse proxy in front of a backend service and enforces per-IP and 
per-endpoint rate limits using a sliding-window counter stored in Redis.
Returns HTTP 429 when a client exceeds the configured threshold.
Includes an admin endpoint for viewing live rate limit statistics.


## Tech Stack

- **Go** — HTTP server, reverse proxy, middleware chain
- **Redis** — shared rate limit counter store
- **YAML** — configuration


## Quick Start

### Prerequisites

- Go 1.25+
- Redis 7+
- Git

### Run

```bash
# Start Redis
docker run -d --name redis -p 6379:6379 redis:7-alpine

# Clone and run
git clone https://github.com/prestonhemmy/ratelimit.git
cd ratelimit
go run ./cmd/gateway
```

### Test

```bash
# Send requests through the gateway
curl http://localhost:8080/get
```


## Configuration

Edit `configs/config.yaml` to set the backend URL, server port and rate limit rules.


## Author

**Preston Hemmy**

GitHub: [@prestonhemmy](https://github.com/prestonhemmy)

LinkedIn: [Preston Hemmy](https://linkedin.com/in/prestonhemmy)