# Polymarket Data Capture Implementation Plan (修正版)

> **状态**: 已根据 2026-02-06 探针测试结果修正
> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** 构建 Go 数据采集服务，持续采集 Polymarket 市场的 L2 订单簿数据。

**关键修正:**
1. ~~`/markets/series/<slug>`~~ → 使用 `/series` + `/events` 组合
2. WebSocket 响应是数组格式
3. `clobTokenIds` 需要 JSON 解析

---

## Phase 1: 项目初始化

### Task 1.1: Go 环境和项目结构

**Files:**
- Create: `go.mod`
- Create: `go.sum`
- Create: `.gitignore` (更新)

**Steps:**
1. 安装 Go 1.21+ (如未安装)
2. `go mod init github.com/user/polymarket-collector`
3. 添加依赖:
   - `nhooyr.io/websocket` (WebSocket)
   - `go.uber.org/zap` (日志)
   - `github.com/stretchr/testify` (测试)

---

## Phase 2: Gamma API 客户端

### Task 2.1: Series/Events 客户端

**Files:**
- Create: `internal/gamma/client.go`
- Create: `internal/gamma/types.go`
- Create: `internal/gamma/client_test.go`

**实际 API 结构:**
```go
// GET https://gamma-api.polymarket.com/series?limit=50
type Series struct {
    ID         string   `json:"id"`
    Slug       string   `json:"slug"`
    Title      string   `json:"title"`
    Active     bool     `json:"active"`
    Volume24hr float64  `json:"volume24hr"`
    Events     []Event  `json:"events"`
}

// GET https://gamma-api.polymarket.com/events?tag_slug=bitcoin&active=true
type Event struct {
    ID       string   `json:"id"`
    Title    string   `json:"title"`
    Slug     string   `json:"slug"`
    EndDate  string   `json:"endDate"`
    Markets  []Market `json:"markets"`
}

type Market struct {
    ID           string `json:"id"`
    Question     string `json:"question"`
    ClobTokenIds string `json:"clobTokenIds"` // JSON 字符串！需要解析
    OutcomePrices string `json:"outcomePrices"`
    Active       bool   `json:"active"`
    Closed       bool   `json:"closed"`
}
```

**测试用例:**
```go
func TestFetchActiveSeries(t *testing.T) {
    cli := gamma.NewClient(http.DefaultClient)
    series, err := cli.FetchSeries(ctx, gamma.SeriesFilter{Active: true, Limit: 10})
    require.NoError(t, err)
    assert.NotEmpty(t, series)
}

func TestFetchEventsByTag(t *testing.T) {
    cli := gamma.NewClient(http.DefaultClient)
    events, err := cli.FetchEvents(ctx, gamma.EventFilter{TagSlug: "bitcoin", Active: true})
    require.NoError(t, err)
    // 验证返回的事件包含 markets
}
```

---

## Phase 3: CLOB REST 客户端

### Task 3.1: 订单簿快照客户端

**Files:**
- Create: `internal/clob/rest_client.go`
- Create: `internal/clob/types.go`
- Create: `internal/clob/rest_client_test.go`

**实际 API 结构:**
```go
// GET https://clob.polymarket.com/book?token_id=xxx
type BookSnapshot struct {
    Market         string       `json:"market"`
    AssetID        string       `json:"asset_id"`
    Timestamp      string       `json:"timestamp"`
    Hash           string       `json:"hash"`
    Bids           []PriceLevel `json:"bids"`
    Asks           []PriceLevel `json:"asks"`
    MinOrderSize   string       `json:"min_order_size"`
    TickSize       string       `json:"tick_size"`
    NegRisk        bool         `json:"neg_risk"`
    LastTradePrice string       `json:"last_trade_price"`
}

type PriceLevel struct {
    Price string `json:"price"`
    Size  string `json:"size"`
}
```

---

## Phase 4: WebSocket 客户端

### Task 4.1: WebSocket 连接和解析

**Files:**
- Create: `internal/ws/client.go`
- Create: `internal/ws/types.go`
- Create: `internal/ws/parser.go`
- Create: `internal/ws/parser_test.go`

**实际消息格式:**
```go
// 订阅消息
type SubscribeMessage struct {
    AssetsIDs []string `json:"assets_ids"`
}

// 响应是数组！
// 解析时需要: var messages []WSMessage; json.Unmarshal(data, &messages)

type WSMessage struct {
    EventType string `json:"event_type"` // "book" 或 "price_change"
    // book 事件
    Market         string       `json:"market,omitempty"`
    AssetID        string       `json:"asset_id,omitempty"`
    Timestamp      string       `json:"timestamp"`
    Hash           string       `json:"hash,omitempty"`
    Bids           []PriceLevel `json:"bids,omitempty"`
    Asks           []PriceLevel `json:"asks,omitempty"`
    LastTradePrice string       `json:"last_trade_price,omitempty"`
    // price_change 事件
    PriceChanges []PriceChange `json:"price_changes,omitempty"`
}

type PriceChange struct {
    AssetID  string `json:"asset_id"`
    Price    string `json:"price"`
    Size     string `json:"size"`
    Side     string `json:"side"` // "BUY" 或 "SELL"
    Hash     string `json:"hash"`
    BestBid  string `json:"best_bid"`
    BestAsk  string `json:"best_ask"`
}
```

**测试用例:**
```go
func TestParseBookMessage(t *testing.T) {
    raw := `[{"event_type":"book","asset_id":"123","bids":[{"price":"0.5","size":"100"}],"asks":[]}]`
    msgs, err := ws.Parse([]byte(raw))
    require.NoError(t, err)
    assert.Len(t, msgs, 1)
    assert.Equal(t, "book", msgs[0].EventType)
}

func TestParsePriceChangeMessage(t *testing.T) {
    raw := `[{"event_type":"price_change","price_changes":[{"asset_id":"123","price":"0.5","side":"BUY"}]}]`
    msgs, err := ws.Parse([]byte(raw))
    require.NoError(t, err)
    assert.Equal(t, "price_change", msgs[0].EventType)
}
```

---

## Phase 5: CLI 探针工具

### Task 5.1: Gamma 探针

**Files:**
- Create: `cmd/probe-gamma/main.go`

**功能:**
- `--list-series`: 列出所有活跃系列
- `--tag <slug>`: 按标签获取事件
- `--output json|table`: 输出格式

### Task 5.2: REST 探针

**Files:**
- Create: `cmd/probe-rest/main.go`

**功能:**
- `--token <id>`: 获取订单簿
- `--watch`: 持续轮询
- `--interval <seconds>`: 轮询间隔

### Task 5.3: WebSocket 探针

**Files:**
- Create: `cmd/probe-ws/main.go`

**功能:**
- `--tokens <id1,id2>`: 订阅多个 token
- `--duration <seconds>`: 运行时长
- `--output <file>`: 保存消息到文件

---

## Phase 6: 数据采集服务

### Task 6.1: 服务主体

**Files:**
- Create: `cmd/collector/main.go`
- Create: `internal/collector/service.go`
- Create: `internal/collector/config.go`

**配置:**
```yaml
gamma:
  base_url: https://gamma-api.polymarket.com
  poll_interval: 60s
  tags:
    - bitcoin
    - ethereum

clob:
  ws_url: wss://ws-subscriptions-clob.polymarket.com/ws/market
  rest_url: https://clob.polymarket.com
  reconnect_backoff:
    base: 2s
    max: 60s

storage:
  type: file  # file | postgres | kafka
  file:
    dir: ./data/raw
    format: jsonl  # jsonl | parquet
```

---

## 测试 Token IDs (已验证可用)

```
# Lady Gaga Grammy (有订单簿)
83955612885151370769947492812886282601680164705864046042194488203730621200472

# Seahawks vs Patriots (高活跃度)
46434110155841033529384949983718980438706543876953886750286883506638610790525
```

---

## 依赖清单

```
go 1.21

require (
    nhooyr.io/websocket v1.8.10
    go.uber.org/zap v1.26.0
    github.com/stretchr/testify v1.8.4
    gopkg.in/yaml.v3 v3.0.1
)
```
