// Package gamma provides a client for the Polymarket Gamma API.
package gamma

import (
	"encoding/json"
	"time"
)

// Series represents a series of related events.
type Series struct {
	ID         string  `json:"id"`
	Slug       string  `json:"slug"`
	Title      string  `json:"title"`
	SeriesType string  `json:"seriesType"`
	Recurrence string  `json:"recurrence"`
	Active     bool    `json:"active"`
	Volume24hr float64 `json:"volume24hr"`
	Liquidity  float64 `json:"liquidity"`
	Events     []Event `json:"events,omitempty"`
}

// Event represents a prediction market event.
type Event struct {
	ID         string    `json:"id"`
	Slug       string    `json:"slug"`
	Title      string    `json:"title"`
	Active     bool      `json:"active"`
	Closed     bool      `json:"closed"`
	StartDate  time.Time `json:"startDate,omitempty"`
	EndDate    time.Time `json:"endDate,omitempty"`
	StartTime  time.Time `json:"startTime,omitempty"` // When trading starts
	Volume24hr float64   `json:"volume24hr"`
	Liquidity  float64   `json:"liquidity"`
	Markets    []Market  `json:"markets,omitempty"`
	Series     []Series  `json:"series,omitempty"`
	Tags       []Tag     `json:"tags,omitempty"`
}

// Tag represents a tag on an event or market.
type Tag struct {
	ID    string `json:"id"`
	Label string `json:"label"`
	Slug  string `json:"slug"`
}

// Market represents a prediction market.
type Market struct {
	ID              string    `json:"id"`
	Question        string    `json:"question"`
	ConditionID     string    `json:"conditionId"`
	Slug            string    `json:"slug"`
	Active          bool      `json:"active"`
	Closed          bool      `json:"closed"`
	LiquidityNum    float64   `json:"liquidityNum"`
	Volume24hr      float64   `json:"volume24hr"`
	EndDate         time.Time `json:"endDate,omitempty"`

	// These fields are JSON strings that need secondary parsing
	ClobTokenIds  string `json:"clobTokenIds"`  // JSON array as string
	OutcomePrices string `json:"outcomePrices"` // JSON array as string
	Outcomes      string `json:"outcomes"`      // JSON array as string

	Events []Event `json:"events,omitempty"`
}

// ParseTokenIDs parses the ClobTokenIds JSON string into a slice of token IDs.
func (m *Market) ParseTokenIDs() ([]string, error) {
	if m.ClobTokenIds == "" {
		return nil, nil
	}
	var ids []string
	if err := json.Unmarshal([]byte(m.ClobTokenIds), &ids); err != nil {
		return nil, err
	}
	return ids, nil
}

// ParseOutcomes parses the Outcomes JSON string into a slice of outcome names.
func (m *Market) ParseOutcomes() ([]string, error) {
	if m.Outcomes == "" {
		return nil, nil
	}
	var outcomes []string
	if err := json.Unmarshal([]byte(m.Outcomes), &outcomes); err != nil {
		return nil, err
	}
	return outcomes, nil
}

// ParseOutcomePrices parses the OutcomePrices JSON string into a slice of prices.
func (m *Market) ParseOutcomePrices() ([]string, error) {
	if m.OutcomePrices == "" {
		return nil, nil
	}
	var prices []string
	if err := json.Unmarshal([]byte(m.OutcomePrices), &prices); err != nil {
		return nil, err
	}
	return prices, nil
}

// Filter contains query parameters for API requests.
type Filter struct {
	Active   *bool  `url:"active,omitempty"`
	Closed   *bool  `url:"closed,omitempty"`
	TagSlug  string `url:"tag_slug,omitempty"`
	Slug     string `url:"slug,omitempty"`
	Limit    int    `url:"_limit,omitempty"`
	Offset   int    `url:"_offset,omitempty"`
}
