package gamma

import (
	"context"
	"net/http"
	"testing"
	"time"
)

func TestFetchSeries_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	client := NewClient(&http.Client{Timeout: 30 * time.Second})
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	active := true
	series, err := client.FetchSeries(ctx, &Filter{Active: &active, Limit: 5})
	if err != nil {
		t.Fatalf("FetchSeries failed: %v", err)
	}

	if len(series) == 0 {
		t.Log("Warning: no active series returned")
		return
	}

	t.Logf("Fetched %d series", len(series))
	for i, s := range series {
		t.Logf("  [%d] %s (slug=%s, active=%v)", i, s.Title, s.Slug, s.Active)
	}
}

func TestFetchEvents_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	client := NewClient(&http.Client{Timeout: 30 * time.Second})
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	active := true
	events, err := client.FetchEvents(ctx, &Filter{Active: &active, Limit: 5})
	if err != nil {
		t.Fatalf("FetchEvents failed: %v", err)
	}

	if len(events) == 0 {
		t.Log("Warning: no active events returned")
		return
	}

	t.Logf("Fetched %d events", len(events))
	for i, e := range events {
		t.Logf("  [%d] %s (slug=%s, markets=%d)", i, e.Title, e.Slug, len(e.Markets))
	}
}

func TestFetchMarkets_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	client := NewClient(&http.Client{Timeout: 30 * time.Second})
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	active := true
	markets, err := client.FetchMarkets(ctx, &Filter{Active: &active, Limit: 5})
	if err != nil {
		t.Fatalf("FetchMarkets failed: %v", err)
	}

	if len(markets) == 0 {
		t.Log("Warning: no active markets returned")
		return
	}

	t.Logf("Fetched %d markets", len(markets))
	for i, m := range markets {
		tokenIDs, _ := m.ParseTokenIDs()
		t.Logf("  [%d] %s (tokens=%d)", i, m.Question, len(tokenIDs))
	}
}

func TestParseTokenIDs(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    []string
		wantErr bool
	}{
		{
			name:  "valid tokens",
			input: `["token1", "token2"]`,
			want:  []string{"token1", "token2"},
		},
		{
			name:  "empty string",
			input: "",
			want:  nil,
		},
		{
			name:  "empty array",
			input: `[]`,
			want:  []string{},
		},
		{
			name:    "invalid json",
			input:   `[invalid`,
			wantErr: true,
		},
		{
			name:  "single token",
			input: `["83955612885151370769947492812886282601680164705864046042194488203730621200472"]`,
			want:  []string{"83955612885151370769947492812886282601680164705864046042194488203730621200472"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &Market{ClobTokenIds: tt.input}
			got, err := m.ParseTokenIDs()
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseTokenIDs() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if len(got) != len(tt.want) {
					t.Errorf("ParseTokenIDs() got %d tokens, want %d", len(got), len(tt.want))
					return
				}
				for i := range got {
					if got[i] != tt.want[i] {
						t.Errorf("ParseTokenIDs()[%d] = %v, want %v", i, got[i], tt.want[i])
					}
				}
			}
		})
	}
}
