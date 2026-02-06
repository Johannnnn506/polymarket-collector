// Package types provides shared type definitions for the Polymarket collector.
package types

import "time"

// PriceLevel represents a single price level in an order book.
type PriceLevel struct {
	Price string `json:"price"`
	Size  string `json:"size"`
}

// TokenSpec contains the specification for a tradeable token.
type TokenSpec struct {
	TokenID  string    `json:"token_id"`
	MarketID string    `json:"market_id"`
	Question string    `json:"question"`
	Outcome  string    `json:"outcome"`
	EndDate  time.Time `json:"end_date"`
}
