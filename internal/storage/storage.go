// Package storage provides storage backends for collected data.
package storage

import (
	"github.com/johan/polymarket-collector/internal/ws"
)

// Storage defines the interface for storing collected data.
type Storage interface {
	// Write writes a WebSocket message to storage.
	Write(msg *ws.WSMessage) error

	// Close closes the storage backend.
	Close() error
}

// NullStorage is a no-op storage that discards all data.
type NullStorage struct{}

// NewNullStorage creates a new null storage.
func NewNullStorage() *NullStorage {
	return &NullStorage{}
}

// Write does nothing.
func (s *NullStorage) Write(msg *ws.WSMessage) error {
	return nil
}

// Close does nothing.
func (s *NullStorage) Close() error {
	return nil
}
