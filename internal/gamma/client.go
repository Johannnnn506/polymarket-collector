package gamma

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

const (
	// DefaultBaseURL is the base URL for the Gamma API.
	DefaultBaseURL = "https://gamma-api.polymarket.com"
)

// Client is an HTTP client for the Gamma API.
type Client struct {
	httpClient *http.Client
	baseURL    string
}

// NewClient creates a new Gamma API client.
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

// FetchSeries fetches series from the Gamma API.
func (c *Client) FetchSeries(ctx context.Context, filter *Filter) ([]Series, error) {
	u := c.baseURL + "/series"
	if filter != nil {
		u += "?" + buildQuery(filter)
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

	var series []Series
	if err := json.NewDecoder(resp.Body).Decode(&series); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	return series, nil
}

// FetchEvents fetches events from the Gamma API.
func (c *Client) FetchEvents(ctx context.Context, filter *Filter) ([]Event, error) {
	u := c.baseURL + "/events"
	if filter != nil {
		u += "?" + buildQuery(filter)
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

	var events []Event
	if err := json.NewDecoder(resp.Body).Decode(&events); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	return events, nil
}

// FetchMarkets fetches markets from the Gamma API.
func (c *Client) FetchMarkets(ctx context.Context, filter *Filter) ([]Market, error) {
	u := c.baseURL + "/markets"
	if filter != nil {
		u += "?" + buildQuery(filter)
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

	var markets []Market
	if err := json.NewDecoder(resp.Body).Decode(&markets); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	return markets, nil
}

// FetchSeriesBySlug fetches a series by its slug, including its events.
func (c *Client) FetchSeriesBySlug(ctx context.Context, slug string) (*Series, error) {
	u := c.baseURL + "/series?slug=" + url.QueryEscape(slug)

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

	var series []Series
	if err := json.NewDecoder(resp.Body).Decode(&series); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	if len(series) == 0 {
		return nil, fmt.Errorf("series not found: %s", slug)
	}

	return &series[0], nil
}

// FetchActiveMarketsForSeries fetches active (not closed) markets for a series.
// Only returns markets that are currently tradeable (startTime <= now < endDate).
// For markets without startTime, we estimate based on the series recurrence.
func (c *Client) FetchActiveMarketsForSeries(ctx context.Context, seriesSlug string) ([]Market, error) {
	series, err := c.FetchSeriesBySlug(ctx, seriesSlug)
	if err != nil {
		return nil, err
	}

	// Determine the trading window based on recurrence
	var tradingWindow time.Duration
	switch series.Recurrence {
	case "5m":
		tradingWindow = 5 * time.Minute
	case "15m":
		tradingWindow = 15 * time.Minute
	case "hourly":
		tradingWindow = 1 * time.Hour
	case "4h":
		tradingWindow = 4 * time.Hour
	case "daily":
		tradingWindow = 24 * time.Hour
	case "weekly":
		tradingWindow = 7 * 24 * time.Hour
	case "monthly":
		tradingWindow = 30 * 24 * time.Hour
	default:
		tradingWindow = 1 * time.Hour // Default to 1 hour
	}

	now := time.Now()
	var activeMarkets []Market

	for _, event := range series.Events {
		if event.Closed {
			continue
		}

		// Skip events that have already ended
		if event.EndDate.Before(now) {
			continue
		}

		// The series API doesn't include nested markets, so we need to fetch each event separately
		events, err := c.FetchEvents(ctx, &Filter{Slug: event.Slug})
		if err != nil {
			continue
		}
		if len(events) == 0 {
			continue
		}

		fullEvent := events[0]

		// Determine if trading has started
		// Note: Actual trading starts before the official startTime, so we start collecting 5 minutes early
		var tradingStarted bool
		earlyStart := 5 * time.Minute
		if !fullEvent.StartTime.IsZero() {
			// Use explicit startTime minus early start buffer
			tradingStarted = !fullEvent.StartTime.Add(-earlyStart).After(now)
		} else {
			// Estimate: trading starts tradingWindow before endDate
			estimatedStart := fullEvent.EndDate.Add(-tradingWindow).Add(-earlyStart)
			tradingStarted = !estimatedStart.After(now)
		}

		if !tradingStarted {
			continue
		}

		for _, market := range fullEvent.Markets {
			if !market.Closed && market.EndDate.After(now) {
				activeMarkets = append(activeMarkets, market)
			}
		}
	}

	return activeMarkets, nil
}

// buildQuery builds URL query parameters from a Filter.
func buildQuery(f *Filter) string {
	v := url.Values{}
	if f.Active != nil {
		v.Set("active", strconv.FormatBool(*f.Active))
	}
	if f.Closed != nil {
		v.Set("closed", strconv.FormatBool(*f.Closed))
	}
	if f.TagSlug != "" {
		v.Set("tag_slug", f.TagSlug)
	}
	if f.Slug != "" {
		v.Set("slug", f.Slug)
	}
	if f.Limit > 0 {
		v.Set("_limit", strconv.Itoa(f.Limit))
	}
	if f.Offset > 0 {
		v.Set("_offset", strconv.Itoa(f.Offset))
	}
	return v.Encode()
}
