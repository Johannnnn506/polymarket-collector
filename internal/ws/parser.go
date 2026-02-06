package ws

import (
	"encoding/json"
	"fmt"
)

// Parse parses a WebSocket message payload.
// The WebSocket returns messages either as JSON arrays or single objects.
func Parse(data []byte) ([]WSMessage, error) {
	if len(data) == 0 {
		return nil, nil
	}

	// Determine if it's an array or single object
	data = trimWhitespace(data)
	if len(data) == 0 {
		return nil, nil
	}

	if data[0] == '[' {
		// Array format
		var messages []WSMessage
		if err := json.Unmarshal(data, &messages); err != nil {
			return nil, fmt.Errorf("parsing websocket message array: %w (data: %s)", err, truncate(data, 100))
		}
		return messages, nil
	}

	// Single object format
	var msg WSMessage
	if err := json.Unmarshal(data, &msg); err != nil {
		return nil, fmt.Errorf("parsing websocket message: %w (data: %s)", err, truncate(data, 100))
	}
	return []WSMessage{msg}, nil
}

// trimWhitespace removes leading whitespace from a byte slice.
func trimWhitespace(data []byte) []byte {
	for len(data) > 0 && (data[0] == ' ' || data[0] == '\t' || data[0] == '\n' || data[0] == '\r') {
		data = data[1:]
	}
	return data
}

// truncate truncates a byte slice to a maximum length for error messages.
func truncate(data []byte, maxLen int) string {
	if len(data) <= maxLen {
		return string(data)
	}
	return string(data[:maxLen]) + "..."
}
