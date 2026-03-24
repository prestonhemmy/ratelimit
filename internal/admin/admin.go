package admin

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/prestonhemmy/ratelimit/internal/config"
	"github.com/redis/go-redis/v9"
)

// Serves the /admin/stats endpoint which reads rate limit counters from Redis
// and returns a JSON summary of active clients.

type StatusSummary struct {
	ActiveClients int           `json:"active_clients"`
	Entries       []StatusEntry `json:"entries"`
}

type StatusEntry struct {
	IP                string `json:"ip"`
	Path              string `json:"path"`
	CurrentCount      int64  `json:"current_count"`
	Limit             int    `json:"limit"`
	WindowResetsInSec int64  `json:"window_resets_in_sec"`
}

type AdminHandler struct {
	client *redis.Client
	cfg    *config.Config
}

func NewAdminHandler(client *redis.Client, cfg *config.Config) *AdminHandler {
	return &AdminHandler{client: client, cfg: cfg}
}

func (h *AdminHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	// extract all rate limit keys from Redis
	var keys []string
	iter := h.client.Scan(ctx, 0, "ratelimit:*", 0).Iterator()
	for iter.Next(ctx) {
		keys = append(keys, iter.Val())
	}

	if err := iter.Err(); err != nil {
		log.Printf("admin stats scan error: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	if len(keys) == 0 {
		writeJSON(w, &StatusSummary{ActiveClients: 0, Entries: []StatusEntry{}})
		return
	}

	// pipelined GET and TTL reads for each key
	pipe := h.client.Pipeline()

	type pendingResult struct {
		key    string
		getCmd *redis.StringCmd
		ttlCmd *redis.DurationCmd
	}

	pending := make([]pendingResult, len(keys))
	for i, key := range keys {
		pending[i] = pendingResult{
			key:    key,
			getCmd: pipe.Get(ctx, key),
			ttlCmd: pipe.TTL(ctx, key),
		}
	}

	_, err := pipe.Exec(ctx)
	if err != nil && err != redis.Nil {
		log.Printf("admin stats pipeline error: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	// parse keys and build entries
	uniqueIPs := make(map[string]bool)
	var entries []StatusEntry
	for _, p := range pending {
		ip, path, ok := parseKey(p.key)
		if !ok {
			continue
		}

		count, err := p.getCmd.Int64()
		if err != nil {
			continue // key expired
		}

		// validate key exists (!= -2) and has expiration (!= -1)
		ttl := int64(p.ttlCmd.Val().Seconds())
		if ttl < 0 {
			ttl = 0
		}

		limit, _ := h.cfg.RuleForPath(path)

		uniqueIPs[ip] = true

		entries = append(entries, StatusEntry{
			IP:                ip,
			Path:              path,
			CurrentCount:      count,
			Limit:             limit,
			WindowResetsInSec: ttl,
		})
	}

	// build summary and write JSON
	summary := &StatusSummary{
		ActiveClients: len(uniqueIPs),
		Entries:       entries,
	}

	if summary.Entries == nil {
		summary.Entries = []StatusEntry{}
	}

	writeJSON(w, summary)
}

// key = "ratelimit:<ip>:<path>:<windowID>"
// Ex.   "ratelimit:192.168.1.42:/get:16"
// Ex.	 "ratelimit:::1:54321:post/:17"
func parseKey(key string) (ip string, path string, ok bool) {
	// remove prefix
	trimmed := strings.TrimPrefix(key, "ratelimit:") // Ex. "192.168.1.42:/get:16"
	if trimmed == key {
		return "", "", false
	}

	lastColonIndex := strings.LastIndex(trimmed, ":")
	if lastColonIndex < 0 {
		return "", "", false
	}

	// validate windowID is int
	if _, err := strconv.ParseInt(trimmed[lastColonIndex+1:], 10, 64); err != nil {
		return "", "", false
	}

	remaining := trimmed[:lastColonIndex] // Ex. "192.168.1.42:/get"

	// extract IP
	sep := strings.Index(remaining, ":/")
	if sep < 0 {
		return "", "", false
	}

	ip = remaining[:sep]     // Ex. "192.168.1.24"
	path = remaining[sep+1:] // Ex. "/get"

	return ip, path, true
}

func writeJSON(w http.ResponseWriter, data any) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Printf("admin stats json encoding error: %v", err)
	}
}
