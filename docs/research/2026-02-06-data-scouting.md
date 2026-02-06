# 2026-02-06 Polymarket 数据探针记录

## 概览
- **目标**: 验证 Gamma API 结构、CLOB WebSocket L2 消息、REST 订单簿快照，收集真实返回字段
- **测试日期**: 2026-02-06
- **测试环境**: Python 3.12 + httpx + websockets
- **结论**: API 结构与原计划假设有显著差异，需要修正

---

## Gamma API

### 端点发现
| 端点 | 状态 | 说明 |
|------|------|------|
| `/markets` | ✅ 200 | 市场列表，支持分页和过滤 |
| `/events` | ✅ 200 | 事件列表，包含关联市场 |
| `/series` | ✅ 200 | 系列列表 |
| `/tags` | ✅ 200 | 标签列表 |
| `/categories` | ✅ 200 | 分类列表 |
| `/markets/series/<slug>` | ❌ 404 | **不存在！原计划假设错误** |

### 关键发现

1. **没有 `/markets/series/<slug>` 端点**
   - 原计划假设的端点不存在
   - 需要通过 `/events?tag_slug=xxx` 或遍历 `/markets` 来获取系列市场

2. **市场结构** (`/markets` 返回)
```json
{
  "id": "36",
  "question": "What will the price of Bitcoin be on November 4th, 2020?",
  "conditionId": "0x...",
  "slug": "what-will-the-price-of-bitcoin-be-on-november-4th-2020",
  "clobTokenIds": "[\"token1\", \"token2\"]",  // 注意：是 JSON 字符串！
  "outcomePrices": "[\"0\", \"0\"]",           // 也是 JSON 字符串
  "outcomes": "[\"Yes\", \"No\"]",
  "active": true,
  "closed": false,
  "liquidityNum": 50000,
  "volume24hr": 1234,
  "endDate": "2026-02-10T00:00:00Z",
  "events": [...]  // 关联的事件
}
```

3. **系列结构** (`/series` 返回)
```json
{
  "id": "45",
  "slug": "btc-multi-strikes-weekly",
  "title": "BTC Multi Strikes Weekly",
  "seriesType": "single",
  "recurrence": "weekly",
  "active": true,
  "volume24hr": 6186480.89,
  "liquidity": 2036796.25,
  "events": [...]  // 关联的事件
}
```

4. **获取特定系列市场的方法**
   - 方法1: `/events?tag_slug=bitcoin&active=true`
   - 方法2: 遍历 `/markets` 检查 `events[].series[].slug`
   - 方法3: `/series` 获取系列，然后通过 `events` 字段获取关联事件

---

## CLOB REST API

### 端点发现
| 端点 | 状态 | 参数 | 说明 |
|------|------|------|------|
| `/markets` | ✅ 200 | `next_cursor` | 市场列表，分页 |
| `/book` | ✅ 200 | `token_id` | 订单簿快照 |
| `/midpoint` | ✅ 200 | `token_id` | 中间价 |
| `/spread` | ✅ 200 | `token_id` | 买卖价差 |
| `/prices-history` | ✅ 200 | `market`, `interval`, `fidelity` | 历史价格 |
| `/price` | ⚠️ 400 | 需要 `side` 参数 | 单边价格 |
| `/sampling-markets` | ✅ 200 | - | 采样市场 |

### `/book` 响应结构
```json
{
  "market": "0x0d880d85cadbe01cf69b30215a8f7304f0bc3e31f6f92218b0b02c9f145e9780",
  "asset_id": "83955612885151370769947492812886282601680164705864046042194488203730621200472",
  "timestamp": "1770358551108",
  "hash": "0aa74f335d82c958cffe22fef5c04b8f8541278f",
  "bids": [{"price": "0.68", "size": "1000"}],
  "asks": [{"price": "0.69", "size": "500"}],
  "min_order_size": "5",
  "tick_size": "0.001",
  "neg_risk": true,
  "last_trade_price": "0.685"
}
```

### `/markets` 响应结构 (CLOB)
```json
{
  "data": [...],
  "next_cursor": "MTAwMA==",
  "limit": 1000,
  "count": 1000
}
```

单个市场:
```json
{
  "condition_id": "0x...",
  "question": "...",
  "market_slug": "...",
  "minimum_order_size": 15,
  "minimum_tick_size": 0.01,
  "tokens": [
    {"token_id": "...", "outcome": "Yes", "price": 0.68, "winner": false},
    {"token_id": "...", "outcome": "No", "price": 0.32, "winner": false}
  ],
  "active": true,
  "closed": false,
  "neg_risk": false
}
```

---

## WebSocket L2

### 连接信息
- **URL**: `wss://ws-subscriptions-clob.polymarket.com/ws/market`
- **订阅格式**: `{"assets_ids": ["token_id1", "token_id2"]}`
- **响应格式**: 数组 `[{...}, {...}]`

### 事件类型

#### 1. `book` - 订单簿快照
首次订阅时发送完整订单簿:
```json
{
  "market": "0x...",
  "asset_id": "token_id",
  "timestamp": "1770358715148",
  "hash": "85689a7a09cab2edbfe5785f9a418bdd71451877",
  "bids": [{"price": "0.68", "size": "1000"}, ...],
  "asks": [{"price": "0.69", "size": "500"}, ...],
  "event_type": "book",
  "last_trade_price": "0.310"
}
```

#### 2. `price_change` - 价格变动
订单簿变化时推送:
```json
{
  "market": "0x...",
  "price_changes": [
    {
      "asset_id": "token_id",
      "price": "0.31",
      "size": "2589581.43",
      "side": "BUY",
      "hash": "e533a8fbeaa3fbb55211f1c2e1664c5b86a219a2",
      "best_bid": "0.31",
      "best_ask": "0.32"
    }
  ],
  "timestamp": "1770358730471",
  "event_type": "price_change"
}
```

### 关键观察
1. **响应是数组**: 每条消息是 `[{...}]` 格式
2. **双边推送**: `price_change` 同时包含两个 outcome 的变化
3. **Hash 校验**: 每条消息都有 `hash` 字段，可用于一致性校验
4. **高频更新**: 活跃市场每秒多条 `price_change`
5. **无心跳**: 测试期间未观察到显式心跳消息

---

## 问题与修正

### 原计划错误
1. ❌ `/markets/series/<slug>` 端点不存在
2. ❌ WebSocket 消息不是单个对象，而是数组
3. ❌ `clobTokenIds` 是 JSON 字符串，需要解析

### 需要修正的架构
1. **Gamma Watcher**: 改用 `/events?tag_slug=xxx` 或遍历 `/series`
2. **WebSocket Parser**: 处理数组响应
3. **Token ID 解析**: 从 Gamma API 获取时需要 `json.loads()`

---

## 下一步
1. 修正实施计划中的 API 假设
2. 确定 BTC 系列的正确获取方式
3. 设计增量订单簿维护逻辑（基于 `price_change`）
4. 确定 hash 校验策略
