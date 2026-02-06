#!/bin/bash
# 清理旧数据文件
# 用法: ./cleanup-data.sh [days]
# 示例: ./cleanup-data.sh 7

set -e

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"
DAYS="${1:-7}"

cd "$PROJECT_DIR"

echo "Cleaning up data files older than $DAYS days..."

# 压缩超过 1 天的 jsonl 文件
find data/ -name "*.jsonl" -mtime +1 -exec gzip {} \;

# 删除超过指定天数的压缩文件
find data/ -name "*.jsonl.gz" -mtime +$DAYS -delete

# 显示当前数据目录状态
echo ""
echo "Current data directory:"
ls -lh data/ 2>/dev/null || echo "No data files"
echo ""
echo "Total size:"
du -sh data/ 2>/dev/null || echo "0"
