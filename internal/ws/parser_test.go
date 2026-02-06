package ws

import (
	"testing"
)

func TestParse_BookMessage(t *testing.T) {
	data := []byte(`[{
		"market": "0x0d880d85cadbe01cf69b30215a8f7304f0bc3e31f6f92218b0b02c9f145e9780",
		"asset_id": "83955612885151370769947492812886282601680164705864046042194488203730621200472",
		"timestamp": "1770358715148",
		"hash": "85689a7a09cab2edbfe5785f9a418bdd71451877",
		"bids": [{"price": "0.68", "size": "1000"}],
		"asks": [{"price": "0.69", "size": "500"}],
		"event_type": "book",
		"last_trade_price": "0.310"
	}]`)

	messages, err := Parse(data)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if len(messages) != 1 {
		t.Fatalf("Expected 1 message, got %d", len(messages))
	}

	msg := messages[0]
	if msg.EventType != EventTypeBook {
		t.Errorf("EventType = %q, want %q", msg.EventType, EventTypeBook)
	}
	if len(msg.Bids) != 1 {
		t.Errorf("Bids count = %d, want 1", len(msg.Bids))
	}
	if len(msg.Asks) != 1 {
		t.Errorf("Asks count = %d, want 1", len(msg.Asks))
	}
	if msg.Bids[0].Price != "0.68" {
		t.Errorf("Bids[0].Price = %q, want %q", msg.Bids[0].Price, "0.68")
	}
	if msg.LastTradePrice != "0.310" {
		t.Errorf("LastTradePrice = %q, want %q", msg.LastTradePrice, "0.310")
	}
}

func TestParse_PriceChangeMessage(t *testing.T) {
	data := []byte(`[{
		"market": "0x0d880d85cadbe01cf69b30215a8f7304f0bc3e31f6f92218b0b02c9f145e9780",
		"price_changes": [
			{
				"asset_id": "token1",
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
	}]`)

	messages, err := Parse(data)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if len(messages) != 1 {
		t.Fatalf("Expected 1 message, got %d", len(messages))
	}

	msg := messages[0]
	if msg.EventType != EventTypePriceChange {
		t.Errorf("EventType = %q, want %q", msg.EventType, EventTypePriceChange)
	}
	if len(msg.PriceChanges) != 1 {
		t.Fatalf("PriceChanges count = %d, want 1", len(msg.PriceChanges))
	}

	pc := msg.PriceChanges[0]
	if pc.Side != "BUY" {
		t.Errorf("PriceChanges[0].Side = %q, want %q", pc.Side, "BUY")
	}
	if pc.Price != "0.31" {
		t.Errorf("PriceChanges[0].Price = %q, want %q", pc.Price, "0.31")
	}
	if pc.BestBid != "0.31" {
		t.Errorf("PriceChanges[0].BestBid = %q, want %q", pc.BestBid, "0.31")
	}
}

func TestParse_EmptyArray(t *testing.T) {
	data := []byte(`[]`)

	messages, err := Parse(data)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if len(messages) != 0 {
		t.Errorf("Expected 0 messages, got %d", len(messages))
	}
}

func TestParse_EmptyData(t *testing.T) {
	messages, err := Parse(nil)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if messages != nil {
		t.Errorf("Expected nil, got %v", messages)
	}
}

func TestParse_InvalidJSON(t *testing.T) {
	data := []byte(`[{invalid json`)

	_, err := Parse(data)
	if err == nil {
		t.Error("Expected error for invalid JSON, got nil")
	}
}

func TestParse_MultipleMessages(t *testing.T) {
	data := []byte(`[
		{"event_type": "book", "timestamp": "1"},
		{"event_type": "price_change", "timestamp": "2"}
	]`)

	messages, err := Parse(data)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if len(messages) != 2 {
		t.Fatalf("Expected 2 messages, got %d", len(messages))
	}

	if messages[0].EventType != EventTypeBook {
		t.Errorf("messages[0].EventType = %q, want %q", messages[0].EventType, EventTypeBook)
	}
	if messages[1].EventType != EventTypePriceChange {
		t.Errorf("messages[1].EventType = %q, want %q", messages[1].EventType, EventTypePriceChange)
	}
}

func TestParse_SingleObject(t *testing.T) {
	data := []byte(`{
		"market": "0x204d24f3a0f5dd5fca825292bdeab6a97af3978b2caa2b21bb37e610eddfff5d",
		"price_changes": [
			{
				"asset_id": "token1",
				"price": "0.50",
				"size": "100",
				"side": "BUY"
			}
		],
		"timestamp": "1770358730471",
		"event_type": "price_change"
	}`)

	messages, err := Parse(data)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if len(messages) != 1 {
		t.Fatalf("Expected 1 message, got %d", len(messages))
	}

	msg := messages[0]
	if msg.EventType != EventTypePriceChange {
		t.Errorf("EventType = %q, want %q", msg.EventType, EventTypePriceChange)
	}
	if len(msg.PriceChanges) != 1 {
		t.Errorf("PriceChanges count = %d, want 1", len(msg.PriceChanges))
	}
}

func TestParse_SingleObjectBook(t *testing.T) {
	data := []byte(`{
		"market": "0x...",
		"asset_id": "12345",
		"timestamp": "1770358715148",
		"bids": [{"price": "0.68", "size": "1000"}],
		"asks": [],
		"event_type": "book"
	}`)

	messages, err := Parse(data)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if len(messages) != 1 {
		t.Fatalf("Expected 1 message, got %d", len(messages))
	}

	if messages[0].EventType != EventTypeBook {
		t.Errorf("EventType = %q, want %q", messages[0].EventType, EventTypeBook)
	}
}
