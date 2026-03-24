#!/bin/bash

GATEWAY="http://localhost:8080"
BOLD="\033[1m"
RESET="\033[0m"
GREEN="\033[32m"
RED="\033[31m"
CYAN="\033[36m"

echo -e "${BOLD}=== Rate Limiting API Gateway Demo ===${RESET}\n"

# Flush Redis

echo -e "${CYAN}[Step 0] Flushing Redis...${RESET}"
redis-cli FLUSHALL > /dev/null 2>&1
echo ""

# Send requests to /get (10 req/min default)

echo -e "${CYAN}[Step 1] Sending 12 GET requests to /get (limit: 10 req/min)${RESET}"
for i in $(seq 1 12); do
    status=$(curl -s -o /dev/null -w "%{http_code}" "$GATEWAY/get")
    if [ "$status" == "200" ]; then
        echo -e "  Request $i: ${GREEN}${status}${RESET}"
    else
        echo -e "  Request $i: ${RED}${status}${RESET}"
    fi
done
echo ""

# Send requests to /post (5 req/min per-endpoint limit)

echo -e "${CYAN}[Step 2] Sending 7 POST requests to /post (limit: 5 req/min)${RESET}"
for i in $(seq 1 7); do
    status=$(curl -s -o /dev/null -w "%{http_code}" -X POST "$GATEWAY/post")
    if [ "$status" == "200" ]; then
        echo -e "  Request $i: ${GREEN}${status}${RESET}"
    else
        echo -e "  Request $i: ${RED}${status}${RESET}"
    fi
done
echo ""

# Verify /get and /post have separate counters

echo -e "${CYAN}[Step 3] Sending 3 GET requests to /get (should still be blocked)${RESET}"
for i in $(seq 1 3); do
    status=$(curl -s -o /dev/null -w "%{http_code}" "$GATEWAY/get")
    if [ "$status" == "200" ]; then
        echo -e "  Request $i: ${GREEN}${status}${RESET}"
    else
        echo -e "  Request $i: ${RED}${status}${RESET}"
    fi
done
echo ""

# Check admin stats

echo -e "${CYAN}[Step 4] Querying /admin/stats${RESET}"
curl -s "$GATEWAY/admin/stats" | jq .
echo ""

# Confirm admin endpoint is not rate limited

echo -e "${CYAN}[Step 5] Sending 20 requests to /admin/stats (should never be blocked)${RESET}"
blocked=0
for i in $(seq 1 20); do
    status=$(curl -s -o /dev/null -w "%{http_code}" "$GATEWAY/admin/stats")
    if [ "$status" != "200" ]; then
        blocked=$((blocked + 1))
    fi
done
if [ "$blocked" -eq 0 ]; then
    echo -e "  ${GREEN}All 20 requests returned 200 — admin endpoint is not rate-limited${RESET}"
else
    echo -e "  ${RED}${blocked} requests were blocked — admin endpoint should not be rate-limited!${RESET}"
fi
echo ""

echo -e "${BOLD}=== Demo Complete ===${RESET}"
