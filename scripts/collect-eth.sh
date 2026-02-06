#!/bin/bash
# 采集 ETH 15分钟市场数据
# 用法: ./collect-eth.sh [duration]
# 示例: ./collect-eth.sh 55m

set -e

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"
DURATION="${1:-55m}"

cd "$PROJECT_DIR"

# 获取当前活跃的 ETH 15m 市场
echo "[$(date)] Fetching active ETH 15m market..."
RESPONSE=$(curl -s "https://gamma-api.polymarket.com/events?active=true&tag_slug=ethereum&_limit=5")

# 找到 eth-updown-15m 系列的市场
TOKENS=$(echo "$RESPONSE" | jq -r '
  [.[] | select(.series[]?.slug == "eth-up-or-down-15m")] |
  .[0].markets[0].clobTokenIds // empty' | \
  jq -r 'fromjson | join(",")')

if [ -z "$TOKENS" ]; then
  echo "[$(date)] No active ETH 15m market found"
  exit 1
fi

OUTPUT_FILE="data/eth-15m-$(date +%Y%m%d-%H%M%S).jsonl"
echo "[$(date)] Starting collection for $DURATION"
echo "[$(date)] Tokens: ${TOKENS:0:50}..."
echo "[$(date)] Output: $OUTPUT_FILE"

./bin/probe-ws \
  --tokens "$TOKENS" \
  --duration "$DURATION" \
  --output "$OUTPUT_FILE" \
  -v

echo "[$(date)] Collection complete: $OUTPUT_FILE"
