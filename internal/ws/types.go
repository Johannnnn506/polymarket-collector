// Package ws provides a WebSocket client for the Polymarket CLOB L2 feed.
package ws

import (
	"github.com/johan/polymarket-collector/internal/types"
)

// SubscribeMessage is the message sent to subscribe to token updates.
type SubscribeMessage struct {
	AssetsIDs []string `json:"assets_ids"`
	Type      string   `json:"type,omitempty"`
}

// WSMessage represents a message received from the WebSocket.
type WSMessage struct {
	EventType      string             `json:"event_type"`
	Market         string             `json:"market"`
	AssetID        string             `json:"asset_id,omitempty"`
	Timestamp      string             `json:"timestamp"`
	Hash           string             `json:"hash,omitempty"`
	Bids           []types.PriceLevel `json:"bids,omitempty"`
	Asks           []types.PriceLevel `json:"asks,omitempty"`
	LastTradePrice string             `json:"last_trade_price,omitempty"`
	PriceChanges   []PriceChange      `json:"price_changes,omitempty"`
}

// PriceChange represents a single price level change.
type PriceChange struct {
	AssetID string `json:"asset_id"`
	Price   string `json:"price"`
	Size    string `json:"size"`
	Side    string `json:"side"` // "BUY" or "SELL"
	Hash    string `json:"hash"`
	BestBid string `json:"best_bid"`
	BestAsk string `json:"best_ask"`
}

// EventTypeBook is the event type for a full order book snapshot.
const EventTypeBook = "book"

// EventTypePriceChange is the event type for price level changes.
const EventTypePriceChange = "price_change"
