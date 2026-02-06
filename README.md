# Polymarket L2 数据采集服务

采集 Polymarket 预测市场的 L2 订单簿数据。

## 目录

- [快速开始](#快速开始)
- [循环采集器](#循环采集器)
- [CLI 工具](#cli-工具)
- [数据格式](#数据格式)
- [VPS 部署](#vps-部署)
- [配置说明](#配置说明)

---

## 快速开始

### 环境要求

- Go 1.21+

### 安装

```bash
# 克隆项目
git clone <repo-url>
cd Polymarket_data_collection

# 安装依赖
go mod tidy

# 编译所有工具
go build ./...
```

### 快速测试

```bash
# 测试 Gamma API
go run ./cmd/probe-gamma --list-series --limit 5

# 测试 REST API (需要 token ID)
go run ./cmd/probe-rest --token 83955612885151370769947492812886282601680164705864046042194488203730621200472

# 测试 WebSocket (采集 10 秒)
go run ./cmd/probe-ws --tokens <token_id> --duration 10s -v
```

---

## 循环采集器

自动采集 ETH/BTC 的循环市场 (15m, hourly, daily)。

### 特性

- 自动发现新开盘的市场
- 支持多个系列并发采集
- 处理市场重叠（下一个周期在上一个结束前开始）
- 按系列/市场组织数据文件
- 市场结束后自动关闭会话

### 快速开始

```bash
# 使用默认配置启动
go run ./cmd/cycle-collector --config config.cycle.yaml

# 指定输出目录
go run ./cmd/cycle-collector --config config.cycle.yaml --output /path/to/data
```

### 配置文件 (config.cycle.yaml)

```yaml
manager:
  scan_interval: 30s      # 扫描新市场的间隔
  grace_period: 60s       # 市场结束后的宽限期
  series:
    # ETH 市场
    - slug: eth-up-or-down-15m
      enabled: true
    - slug: eth-up-or-down-hourly
      enabled: true
    - slug: eth-up-or-down-daily
      enabled: true
    # BTC 市场
    - slug: btc-up-or-down-15m
      enabled: true
    - slug: btc-up-or-down-hourly
      enabled: true
    - slug: btc-up-or-down-daily
      enabled: true

storage:
  type: file
  output_dir: data
```

### 数据目录结构

```
data/
├── eth-up-or-down-15m/
│   ├── 2026-02-06_1770361200.jsonl   # 文件名: 日期_结束时间戳
│   └── 2026-02-06_1770362100.jsonl
├── eth-up-or-down-hourly/
│   └── ...
├── btc-up-or-down-15m/
│   └── ...
└── btc-up-or-down-daily/
    └── ...
```

### 数据文件格式

每个文件的第一行是元数据:

```json
{
  "type": "metadata",
  "series_slug": "eth-up-or-down-15m",
  "market_id": "1338378",
  "condition_id": "0x...",
  "token_ids": ["token1", "token2"],
  "end_date": "2026-02-06T08:15:00Z",
  "start_time": "2026-02-06T03:06:37Z"
}
```

后续行是 WebSocket 消息 (book, price_change, last_trade_price)。

### 运行示例

```
$ go run ./cmd/cycle-collector --config config.cycle.yaml

2026/02/06 03:17:37 Tracking series: eth-up-or-down-15m
2026/02/06 03:17:37 Tracking series: eth-up-or-down-hourly
2026/02/06 03:17:37 Tracking series: btc-up-or-down-15m
...
2026/02/06 03:17:37 Starting cycle collector with 6 series...
2026/02/06 03:18:14 [eth-up-or-down-15m] Session started for market 1338519, ends at 08:30:00
2026/02/06 03:18:34 [eth-hourly] Session started for market 1331753, ends at 09:00:00
...
2026/02/06 03:19:29 Active sessions: 6
2026/02/06 03:19:29   [eth-up-or-down-15m] market=1338519 msgs=11620 ends_in=10m30s
2026/02/06 03:19:29   [eth-hourly] market=1331753 msgs=5412 ends_in=40m30s
...
```

### systemd 服务配置

```ini
[Unit]
Description=Polymarket Cycle Collector
After=network.target

[Service]
Type=simple
User=polymarket
WorkingDirectory=/opt/polymarket-collector
ExecStart=/opt/polymarket-collector/cycle-collector --config /etc/polymarket/config.cycle.yaml
Restart=always
RestartSec=10

[Install]
WantedBy=multi-user.target
```

---

## CLI 工具

### 1. probe-gamma - Gamma API 探针

探索市场、事件、系列数据。

```bash
# 列出活跃系列
go run ./cmd/probe-gamma --list-series

# 按标签搜索事件
go run ./cmd/probe-gamma --tag ethereum --limit 20

# 列出活跃市场
go run ./cmd/probe-gamma --list-markets --limit 10

# JSON 输出
go run ./cmd/probe-gamma --list-series --output json
```

**参数:**
| 参数 | 说明 |
|------|------|
| `--list-series` | 列出系列 |
| `--list-events` | 列出事件 |
| `--list-markets` | 列出市场 |
| `--tag <slug>` | 按标签过滤 |
| `--limit <n>` | 结果数量限制 |
| `--output <format>` | 输出格式: table / json |

### 2. probe-rest - CLOB REST API 探针

获取订单簿快照。

```bash
# 获取订单簿
go run ./cmd/probe-rest --token <token_id>

# 持续轮询
go run ./cmd/probe-rest --token <token_id> --watch --interval 2s

# 只获取中间价
go run ./cmd/probe-rest --token <token_id> --midpoint

# 只获取价差
go run ./cmd/probe-rest --token <token_id> --spread
```

**参数:**
| 参数 | 说明 |
|------|------|
| `--token <id>` | Token ID (必需) |
| `--watch` | 持续轮询模式 |
| `--interval <duration>` | 轮询间隔 (默认 5s) |
| `--midpoint` | 只获取中间价 |
| `--spread` | 只获取价差 |
| `--output <format>` | 输出格式: table / json |

### 3. probe-ws - WebSocket 探针

实时订阅订单簿更新。

```bash
# 订阅单个 token
go run ./cmd/probe-ws --tokens <token_id> --duration 1m -v

# 订阅多个 token
go run ./cmd/probe-ws --tokens <id1>,<id2> --duration 5m

# 保存到文件
go run ./cmd/probe-ws --tokens <token_id> --duration 10m --output data/output.jsonl -v
```

**参数:**
| 参数 | 说明 |
|------|------|
| `--tokens <ids>` | Token ID 列表，逗号分隔 (必需) |
| `--duration <duration>` | 运行时长 (0 = 无限) |
| `--output <file>` | 输出文件路径 |
| `-v` | 详细输出模式 |

### 4. collector - 完整采集服务

自动发现市场并持续采集。

```bash
# 使用配置文件
cp config.example.yaml config.yaml
go run ./cmd/collector --config config.yaml

# 使用默认配置
go run ./cmd/collector
```

---

## 数据格式

采集的数据为 JSONL 格式 (每行一个 JSON 对象)。

### 消息类型

#### 1. book - 订单簿快照

```json
{
  "event_type": "book",
  "market": "0x...",
  "asset_id": "token_id",
  "timestamp": "1770361953444",
  "hash": "cb543a7fdb0dd29a8f227703806980cf14e48085",
  "bids": [{"price": "0.54", "size": "467.63"}, ...],
  "asks": [{"price": "0.55", "size": "236.12"}, ...],
  "last_trade_price": "0.450"
}
```

#### 2. price_change - 价格变动

```json
{
  "event_type": "price_change",
  "market": "0x...",
  "timestamp": "1770361955508",
  "price_changes": [
    {
      "asset_id": "token_id",
      "price": "0.54",
      "size": "100",
      "side": "BUY",
      "hash": "...",
      "best_bid": "0.54",
      "best_ask": "0.55"
    }
  ]
}
```

#### 3. last_trade_price - 最新成交价更新

```json
{
  "event_type": "last_trade_price",
  "market": "0x...",
  "asset_id": "token_id",
  "timestamp": "1770361964656"
}
```

### 字段说明

| 字段 | 说明 |
|------|------|
| `timestamp` | 毫秒时间戳 |
| `market` | 市场条件 ID (condition_id) |
| `asset_id` | Token ID |
| `bids` | 买盘档位 (价格降序) |
| `asks` | 卖盘档位 (价格升序) |
| `price` | 价格 |
| `size` | 数量 |
| `side` | 方向: BUY / SELL |
| `best_bid` | 最优买价 |
| `best_ask` | 最优卖价 |
| `hash` | 订单簿状态哈希 |

---

## VPS 部署

### 1. 准备工作

```bash
# 连接到 VPS
ssh user@your-vps-ip

# 安装 Go (Ubuntu/Debian)
wget https://go.dev/dl/go1.22.0.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.22.0.linux-amd64.tar.gz
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
source ~/.bashrc

# 验证安装
go version
```

### 2. 部署项目

```bash
# 克隆项目
git clone <repo-url> ~/polymarket-collector
cd ~/polymarket-collector

# 编译
go build -o bin/probe-gamma ./cmd/probe-gamma
go build -o bin/probe-rest ./cmd/probe-rest
go build -o bin/probe-ws ./cmd/probe-ws
go build -o bin/collector ./cmd/collector

# 创建数据目录
mkdir -p data
```

### 3. 配置

```bash
# 复制并编辑配置
cp config.example.yaml config.yaml
nano config.yaml
```

**推荐配置:**

```yaml
discovery:
  refresh_interval: 5m
  tags: ["ethereum", "bitcoin"]  # 按需修改
  active_only: true
  max_markets: 50

storage:
  type: file
  output_dir: /home/user/polymarket-collector/data
  rotation_interval: 1h

websocket:
  initial_backoff: 1s
  max_backoff: 30s
  backoff_factor: 2.0

logging:
  level: info
  format: text
```

### 4. 使用 systemd 管理服务

创建服务文件:

```bash
sudo nano /etc/systemd/system/polymarket-collector.service
```

内容:

```ini
[Unit]
Description=Polymarket L2 Data Collector
After=network.target

[Service]
Type=simple
User=your-username
WorkingDirectory=/home/your-username/polymarket-collector
ExecStart=/home/your-username/polymarket-collector/bin/collector --config config.yaml
Restart=always
RestartSec=10

# 日志
StandardOutput=append:/home/your-username/polymarket-collector/logs/collector.log
StandardError=append:/home/your-username/polymarket-collector/logs/collector.log

[Install]
WantedBy=multi-user.target
```

启动服务:

```bash
# 创建日志目录
mkdir -p ~/polymarket-collector/logs

# 重载 systemd
sudo systemctl daemon-reload

# 启动服务
sudo systemctl start polymarket-collector

# 设置开机自启
sudo systemctl enable polymarket-collector

# 查看状态
sudo systemctl status polymarket-collector

# 查看日志
tail -f ~/polymarket-collector/logs/collector.log
```

### 5. 使用 Screen/Tmux (简单方式)

如果不想用 systemd:

```bash
# 使用 screen
screen -S collector
cd ~/polymarket-collector
./bin/collector --config config.yaml
# Ctrl+A, D 分离

# 重新连接
screen -r collector
```

或使用 tmux:

```bash
tmux new -s collector
cd ~/polymarket-collector
./bin/collector --config config.yaml
# Ctrl+B, D 分离

# 重新连接
tmux attach -t collector
```

### 6. 手动采集特定市场

```bash
# 1. 查找市场
./bin/probe-gamma --tag ethereum --output json | jq '.[0]'

# 2. 获取 token IDs
curl -s "https://gamma-api.polymarket.com/events?slug=eth-updown-15m-XXXXXXXXXX" | \
  jq '.[0].markets[0].clobTokenIds'

# 3. 开始采集
./bin/probe-ws \
  --tokens "token_id_1,token_id_2" \
  --duration 1h \
  --output data/eth-updown-$(date +%Y%m%d-%H%M%S).jsonl \
  -v
```

### 7. 定时任务 (Cron)

每小时采集特定市场:

```bash
crontab -e
```

添加:

```cron
# 每小时采集 ETH 15分钟市场，持续 55 分钟
0 * * * * /home/user/polymarket-collector/scripts/collect-eth.sh >> /home/user/polymarket-collector/logs/cron.log 2>&1
```

创建脚本 `scripts/collect-eth.sh`:

```bash
#!/bin/bash
cd /home/user/polymarket-collector

# 获取当前活跃的 ETH 15m 市场
TOKENS=$(curl -s "https://gamma-api.polymarket.com/events?active=true&tag_slug=ethereum&_limit=1" | \
  jq -r '.[0].markets[0].clobTokenIds' | \
  jq -r 'fromjson | join(",")')

if [ -n "$TOKENS" ]; then
  ./bin/probe-ws \
    --tokens "$TOKENS" \
    --duration 55m \
    --output "data/eth-$(date +%Y%m%d-%H%M%S).jsonl"
fi
```

```bash
chmod +x scripts/collect-eth.sh
```

### 8. 数据管理

```bash
# 查看数据文件
ls -lh data/

# 压缩旧数据
gzip data/orderbook_2026-02-05*.jsonl

# 统计数据量
wc -l data/*.jsonl

# 清理 7 天前的数据
find data/ -name "*.jsonl" -mtime +7 -delete
```

---

## 配置说明

`config.yaml` 完整配置:

```yaml
# 市场发现设置
discovery:
  refresh_interval: 5m      # 刷新市场列表间隔
  tags: []                  # 标签过滤 (空 = 所有市场)
  active_only: true         # 只包含活跃市场
  max_markets: 100          # 最大跟踪市场数

# 存储设置
storage:
  type: file                # file 或 none
  output_dir: data          # 输出目录
  rotation_interval: 1h     # 文件轮转间隔

# WebSocket 设置
websocket:
  url: ""                   # 自定义 URL (空 = 默认)
  initial_backoff: 1s       # 初始重连等待
  max_backoff: 30s          # 最大重连等待
  backoff_factor: 2.0       # 退避倍数

# 日志设置
logging:
  level: info               # debug, info, warn, error
  format: text              # text 或 json
```

---

## 常见问题

### Q: 如何找到特定市场的 Token ID?

```bash
# 方法 1: 通过 Gamma API
curl -s "https://gamma-api.polymarket.com/events?slug=<event-slug>" | \
  jq '.[0].markets[0].clobTokenIds'

# 方法 2: 使用 probe-gamma
go run ./cmd/probe-gamma --tag bitcoin --output json | \
  jq '.[].markets[].clobTokenIds'
```

### Q: WebSocket 断开怎么办?

服务会自动重连，使用指数退避策略。如果持续断开，检查网络连接。

### Q: 数据文件太大怎么办?

1. 减小 `rotation_interval` (如 30m)
2. 定期压缩: `gzip data/*.jsonl`
3. 设置 cron 清理旧文件

### Q: 如何只采集特定类型的市场?

在 `config.yaml` 中设置 `tags`:

```yaml
discovery:
  tags: ["ethereum", "bitcoin"]
```

或手动指定 token IDs 使用 `probe-ws`。

---

## API 参考

- Gamma API: `https://gamma-api.polymarket.com`
- CLOB REST: `https://clob.polymarket.com`
- WebSocket: `wss://ws-subscriptions-clob.polymarket.com/ws/market`

## License

MIT
