# Polymarket L2 数据采集 Go 服务架构（修正版）

> **状态**: 已根据 2026-02-06 探针测试结果修正

## 目标
持续、稳定地采集 Polymarket 市场 L2 数据，支持多市场并发、断线自动恢复、数据落盘、监控告警。

## API 端点总结

### Gamma API (市场元数据)
| 端点 | 用途 |
|------|------|
| `GET /series` | 获取系列列表 |
| `GET /events?tag_slug=xxx` | 按标签获取事件 |
| `GET /markets?active=true` | 获取活跃市场 |

### CLOB REST API (订单簿)
| 端点 | 用途 |
|------|------|
| `GET /book?token_id=xxx` | 订单簿快照 |
| `GET /midpoint?token_id=xxx` | 中间价 |
| `GET /spread?token_id=xxx` | 买卖价差 |
| `GET /prices-history` | 历史价格 |

### WebSocket
| URL | 用途 |
|-----|------|
| `wss://ws-subscriptions-clob.polymarket.com/ws/market` | L2 实时流 |

---

## 组件划分

### 1. Market Discovery (市场发现)
- 定期轮询 Gamma API `/events?tag_slug=xxx` 或 `/series`
- 解析 `clobTokenIds` (JSON 字符串)
- 输出: `[]TokenSpec{token_id, market_id, question, end_date}`

```go
type MarketDiscovery struct {
    gammaClient *gamma.Client
    tags        []string        // ["bitcoin", "ethereum"]
    interval    time.Duration   // 60s
}

func (d *MarketDiscovery) Discover(ctx context.Context) ([]TokenSpec, error)
```

### 2. Subscription Manager (订阅管理)
- 维护 `map[token_id]*MarketSession`
- 处理 token 增删
- 暴露 Prometheus 指标

```go
type SubscriptionManager struct {
    sessions map[string]*MarketSession
    mu       sync.RWMutex
}

func (m *SubscriptionManager) Subscribe(token TokenSpec) error
func (m *SubscriptionManager) Unsubscribe(tokenID string) error
func (m *SubscriptionManager) ActiveCount() int
```

### 3. WebSocket Session (WebSocket 会话)
- 连接 `wss://ws-subscriptions-clob.polymarket.com/ws/market`
- 发送订阅: `{"assets_ids": ["token1", "token2"]}`
- **处理数组响应**: `[{event_type: "book", ...}]`
- 事件类型:
  - `book`: 完整订单簿快照
  - `price_change`: 增量更新

```go
type WSSession struct {
    conn      *websocket.Conn
    tokens    []string
    handler   func([]WSMessage)
    reconnect BackoffConfig
}

// 关键: 响应是数组
func (s *WSSession) readLoop(ctx context.Context) {
    for {
        _, data, err := s.conn.Read(ctx)
        var messages []WSMessage
        json.Unmarshal(data, &messages)  // 数组！
        s.handler(messages)
    }
}
```

### 4. REST Verifier (REST 校验器)
- 周期性调用 `/book` 获取快照
- 比对 WebSocket 维护的订单簿
- Hash 不一致时触发重建

```go
type RESTVerifier struct {
    client   *clob.Client
    interval time.Duration  // 120s
}

func (v *RESTVerifier) Verify(tokenID string, localBook *OrderBook) (bool, error)
```

### 5. Event Pipeline (事件管道)
- 接收 WebSocket 事件
- 添加元数据: `received_at`, `latency`
- 写入存储

```go
type EventPipeline struct {
    input   chan WSMessage
    storage Storage
    buffer  int  // 批量写入大小
}
```

### 6. Storage (存储)
- 接口抽象，支持多种后端

```go
type Storage interface {
    Write(ctx context.Context, events []Event) error
    Close() error
}

// 实现
type FileStorage struct { ... }      // JSONL 文件
type PostgresStorage struct { ... }  // TimescaleDB
type KafkaStorage struct { ... }     // Kafka/Redpanda
```

---

## 数据流

```
┌─────────────────┐
│ Market Discovery│ ──(60s)──> 发现新 token
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│ Subscription Mgr│ ──> 创建/销毁 Session
└────────┬────────┘
         │
         ▼
┌─────────────────┐     ┌──────────────┐
│  WebSocket Sess │ ◄───│ REST Verifier│ (120s 校验)
└────────┬────────┘     └──────────────┘
         │
         ▼
┌─────────────────┐
│ Event Pipeline  │ ──> 批量写入
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│    Storage      │ (File/Postgres/Kafka)
└─────────────────┘
```

---

## 配置

```yaml
# config.yaml
discovery:
  gamma_url: https://gamma-api.polymarket.com
  interval: 60s
  tags:
    - bitcoin
    - ethereum
  # 或者指定 series
  series:
    - btc-multi-strikes-weekly

websocket:
  url: wss://ws-subscriptions-clob.polymarket.com/ws/market
  reconnect:
    base: 2s
    max: 60s
    jitter: 0.1

rest:
  url: https://clob.polymarket.com
  verify_interval: 120s
  timeout: 10s

storage:
  type: file  # file | postgres | kafka
  file:
    dir: ./data/raw
    rotate: daily
    format: jsonl

observability:
  metrics_port: 9090
  log_level: info
  log_format: json
```

---

## 监控指标

```
# WebSocket
polymarket_ws_messages_total{event_type="book|price_change"}
polymarket_ws_reconnects_total
polymarket_ws_latency_seconds

# REST
polymarket_rest_requests_total{endpoint="book|midpoint"}
polymarket_rest_latency_seconds
polymarket_rest_hash_mismatch_total

# Discovery
polymarket_discovery_tokens_active
polymarket_discovery_tokens_added_total
polymarket_discovery_tokens_removed_total

# Storage
polymarket_storage_writes_total
polymarket_storage_errors_total
polymarket_storage_buffer_size
```

---

## 部署

### 单机部署
```bash
# 构建
go build -o polymarket-collector ./cmd/collector

# 运行
./polymarket-collector --config config.yaml

# systemd service
[Unit]
Description=Polymarket Data Collector
After=network.target

[Service]
ExecStart=/usr/local/bin/polymarket-collector --config /etc/polymarket/config.yaml
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
```

### Docker
```dockerfile
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o collector ./cmd/collector

FROM alpine:3.19
COPY --from=builder /app/collector /usr/local/bin/
ENTRYPOINT ["collector"]
```

---

## 已验证的测试数据

```
# 有订单簿的 Token (可用于测试)
83955612885151370769947492812886282601680164705864046042194488203730621200472

# 高活跃度 Token (Seahawks vs Patriots)
46434110155841033529384949983718980438706543876953886750286883506638610790525
```

---

## 未决问题

1. **存储选型**: 初期用 JSONL 文件，后续根据数据量决定是否迁移到 TimescaleDB
2. **多地区冗余**: 暂不需要，单点部署即可
3. **L3 数据**: 官方不提供，暂不实现
4. **Hash 校验策略**: 每 120s REST 校验，不一致时重建订单簿
