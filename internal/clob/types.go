// Package clob provides a client for the Polymarket CLOB REST API.
package clob

import (
	"github.com/johan/polymarket-collector/internal/types"
)

// BookSnapshot represents an order book snapshot from the CLOB API.
type BookSnapshot struct {
	Market         string             `json:"market"`
	AssetID        string             `json:"asset_id"`
	Timestamp      string             `json:"timestamp"`
	Hash           string             `json:"hash"`
	Bids           []types.PriceLevel `json:"bids"`
	Asks           []types.PriceLevel `json:"asks"`
	MinOrderSize   string             `json:"min_order_size"`
	TickSize       string             `json:"tick_size"`
	NegRisk        bool               `json:"neg_risk"`
	LastTradePrice string             `json:"last_trade_price"`
}

// MidpointResponse represents the response from the midpoint endpoint.
type MidpointResponse struct {
	Mid string `json:"mid"`
}

// SpreadResponse represents the response from the spread endpoint.
type SpreadResponse struct {
	Spread string `json:"spread"`
}

// CLOBMarket represents a market from the CLOB API.
type CLOBMarket struct {
	ConditionID      string       `json:"condition_id"`
	Question         string       `json:"question"`
	MarketSlug       string       `json:"market_slug"`
	MinimumOrderSize float64      `json:"minimum_order_size"`
	MinimumTickSize  float64      `json:"minimum_tick_size"`
	Tokens           []CLOBToken  `json:"tokens"`
	Active           bool         `json:"active"`
	Closed           bool         `json:"closed"`
	NegRisk          bool         `json:"neg_risk"`
}

// CLOBToken represents a token in a CLOB market.
type CLOBToken struct {
	TokenID string  `json:"token_id"`
	Outcome string  `json:"outcome"`
	Price   float64 `json:"price"`
	Winner  bool    `json:"winner"`
}

// MarketsResponse represents the paginated response from the markets endpoint.
type MarketsResponse struct {
	Data       []CLOBMarket `json:"data"`
	NextCursor string       `json:"next_cursor"`
	Limit      int          `json:"limit"`
	Count      int          `json:"count"`
}
