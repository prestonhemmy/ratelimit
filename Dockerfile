# Build Stage
FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -o /gateway ./cmd/gateway


# Runtime Stage
FROM alpine:3.20

COPY --from=builder /gateway /gateway
COPY --from=builder /app/configs /configs

EXPOSE 8080

ENTRYPOINT ["/gateway"]
