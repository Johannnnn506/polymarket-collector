package clob

import (
	"context"
	"net/http"
	"testing"
	"time"
)

const (
	// Known active token ID for testing
	testTokenID = "83955612885151370769947492812886282601680164705864046042194488203730621200472"
)

func TestFetchBook_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	client := NewClient(&http.Client{Timeout: 30 * time.Second})
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	book, err := client.FetchBook(ctx, testTokenID)
	if err != nil {
		t.Fatalf("FetchBook failed: %v", err)
	}

	t.Logf("Book for token %s:", testTokenID[:20]+"...")
	t.Logf("  Market: %s", book.Market)
	t.Logf("  Timestamp: %s", book.Timestamp)
	t.Logf("  Hash: %s", book.Hash)
	t.Logf("  Bids: %d levels", len(book.Bids))
	t.Logf("  Asks: %d levels", len(book.Asks))
	t.Logf("  LastTradePrice: %s", book.LastTradePrice)

	if len(book.Bids) > 0 {
		t.Logf("  Best bid: %s @ %s", book.Bids[0].Size, book.Bids[0].Price)
	}
	if len(book.Asks) > 0 {
		t.Logf("  Best ask: %s @ %s", book.Asks[0].Size, book.Asks[0].Price)
	}
}

func TestFetchBook_NotFound(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	client := NewClient(&http.Client{Timeout: 30 * time.Second})
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	_, err := client.FetchBook(ctx, "invalid_token_id_12345")
	if err == nil {
		t.Error("Expected error for invalid token ID, got nil")
	}
	t.Logf("Got expected error: %v", err)
}

func TestFetchMidpoint_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	client := NewClient(&http.Client{Timeout: 30 * time.Second})
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	mid, err := client.FetchMidpoint(ctx, testTokenID)
	if err != nil {
		t.Fatalf("FetchMidpoint failed: %v", err)
	}

	t.Logf("Midpoint for token %s: %s", testTokenID[:20]+"...", mid)
}

func TestFetchSpread_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	client := NewClient(&http.Client{Timeout: 30 * time.Second})
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	spread, err := client.FetchSpread(ctx, testTokenID)
	if err != nil {
		t.Fatalf("FetchSpread failed: %v", err)
	}

	t.Logf("Spread for token %s: %s", testTokenID[:20]+"...", spread)
}
