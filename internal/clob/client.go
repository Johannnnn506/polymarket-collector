package clob

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

const (
	// DefaultBaseURL is the base URL for the CLOB API.
	DefaultBaseURL = "https://clob.polymarket.com"
)

// Client is an HTTP client for the CLOB API.
type Client struct {
	httpClient *http.Client
	baseURL    string
}

// NewClient creates a new CLOB API client.
func NewClient(httpClient *http.Client) *Client {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	return &Client{
		httpClient: httpClient,
		baseURL:    DefaultBaseURL,
	}
}

// WithBaseURL sets a custom base URL for the client.
func (c *Client) WithBaseURL(baseURL string) *Client {
	c.baseURL = baseURL
	return c
}

// FetchBook fetches the order book for a given token ID.
func (c *Client) FetchBook(ctx context.Context, tokenID string) (*BookSnapshot, error) {
	u := fmt.Sprintf("%s/book?token_id=%s", c.baseURL, url.QueryEscape(tokenID))

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("executing request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("token not found: %s", tokenID)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	var book BookSnapshot
	if err := json.NewDecoder(resp.Body).Decode(&book); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	return &book, nil
}

// FetchMidpoint fetches the midpoint price for a given token ID.
func (c *Client) FetchMidpoint(ctx context.Context, tokenID string) (string, error) {
	u := fmt.Sprintf("%s/midpoint?token_id=%s", c.baseURL, url.QueryEscape(tokenID))

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return "", fmt.Errorf("creating request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("executing request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	var midResp MidpointResponse
	if err := json.NewDecoder(resp.Body).Decode(&midResp); err != nil {
		return "", fmt.Errorf("decoding response: %w", err)
	}

	return midResp.Mid, nil
}

// FetchSpread fetches the spread for a given token ID.
func (c *Client) FetchSpread(ctx context.Context, tokenID string) (string, error) {
	u := fmt.Sprintf("%s/spread?token_id=%s", c.baseURL, url.QueryEscape(tokenID))

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return "", fmt.Errorf("creating request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("executing request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	var spreadResp SpreadResponse
	if err := json.NewDecoder(resp.Body).Decode(&spreadResp); err != nil {
		return "", fmt.Errorf("decoding response: %w", err)
	}

	return spreadResp.Spread, nil
}

// FetchMarkets fetches markets from the CLOB API with optional pagination cursor.
func (c *Client) FetchMarkets(ctx context.Context, cursor string) (*MarketsResponse, error) {
	u := c.baseURL + "/markets"
	if cursor != "" {
		u += "?next_cursor=" + url.QueryEscape(cursor)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("executing request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	var marketsResp MarketsResponse
	if err := json.NewDecoder(resp.Body).Decode(&marketsResp); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	return &marketsResp, nil
}
