// Package collector provides the main data collection service.
package collector

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/johan/polymarket-collector/internal/config"
	"github.com/johan/polymarket-collector/internal/gamma"
	"github.com/johan/polymarket-collector/internal/storage"
	"github.com/johan/polymarket-collector/internal/ws"
)

// Service is the main data collection service.
type Service struct {
	config  *config.Config
	gamma   *gamma.Client
	storage storage.Storage
	ws      *ws.Client

	mu       sync.Mutex
	tokenIDs []string
}

// NewService creates a new collector service.
func NewService(cfg *config.Config) (*Service, error) {
	// Create HTTP client
	httpClient := &http.Client{Timeout: 30 * time.Second}

	// Create Gamma client
	gammaClient := gamma.NewClient(httpClient)

	// Create storage
	var stor storage.Storage
	var err error
	switch cfg.Storage.Type {
	case "file":
		stor, err = storage.NewFileStorage(cfg.Storage.OutputDir, cfg.Storage.RotationInterval)
		if err != nil {
			return nil, fmt.Errorf("creating file storage: %w", err)
		}
	case "none":
		stor = storage.NewNullStorage()
	default:
		return nil, fmt.Errorf("unknown storage type: %s", cfg.Storage.Type)
	}

	s := &Service{
		config:  cfg,
		gamma:   gammaClient,
		storage: stor,
	}

	// Create WebSocket client
	s.ws = ws.NewWSClient(s.handleMessages)
	if cfg.WebSocket.URL != "" {
		s.ws.WithURL(cfg.WebSocket.URL)
	}
	s.ws.WithReconnectConfig(ws.ReconnectConfig{
		InitialBackoff: cfg.WebSocket.InitialBackoff,
		MaxBackoff:     cfg.WebSocket.MaxBackoff,
		BackoffFactor:  cfg.WebSocket.BackoffFactor,
	})

	return s, nil
}

// Run starts the collector service.
func (s *Service) Run(ctx context.Context) error {
	log.Println("Starting collector service...")

	// Initial market discovery
	if err := s.discoverMarkets(ctx); err != nil {
		return fmt.Errorf("initial market discovery: %w", err)
	}

	if len(s.tokenIDs) == 0 {
		return fmt.Errorf("no markets discovered")
	}

	log.Printf("Discovered %d tokens to track", len(s.tokenIDs))

	// Connect to WebSocket
	if err := s.ws.Connect(ctx); err != nil {
		return fmt.Errorf("connecting to websocket: %w", err)
	}
	defer s.ws.Close()

	// Subscribe to discovered tokens
	if err := s.ws.Subscribe(s.tokenIDs); err != nil {
		return fmt.Errorf("subscribing to tokens: %w", err)
	}

	log.Println("Subscribed to WebSocket feed. Collecting data...")

	// Start market refresh ticker
	refreshTicker := time.NewTicker(s.config.Discovery.RefreshInterval)
	defer refreshTicker.Stop()

	// Main loop
	for {
		select {
		case <-ctx.Done():
			log.Println("Shutting down collector service...")
			return s.storage.Close()

		case <-refreshTicker.C:
			log.Println("Refreshing market list...")
			if err := s.discoverMarkets(ctx); err != nil {
				log.Printf("Warning: market refresh failed: %v", err)
				continue
			}

			// Re-subscribe with updated token list
			if err := s.ws.Subscribe(s.tokenIDs); err != nil {
				log.Printf("Warning: resubscription failed: %v", err)
			} else {
				log.Printf("Updated subscription with %d tokens", len(s.tokenIDs))
			}
		}
	}
}

// discoverMarkets fetches active markets and extracts token IDs.
func (s *Service) discoverMarkets(ctx context.Context) error {
	var allTokenIDs []string
	active := s.config.Discovery.ActiveOnly

	if len(s.config.Discovery.Tags) > 0 {
		// Fetch events by tag
		for _, tag := range s.config.Discovery.Tags {
			events, err := s.gamma.FetchEvents(ctx, &gamma.Filter{
				Active:  &active,
				TagSlug: tag,
				Limit:   s.config.Discovery.MaxMarkets,
			})
			if err != nil {
				log.Printf("Warning: failed to fetch events for tag %s: %v", tag, err)
				continue
			}

			for _, event := range events {
				for _, market := range event.Markets {
					tokenIDs, err := market.ParseTokenIDs()
					if err != nil {
						log.Printf("Warning: failed to parse token IDs for market %s: %v", market.ID, err)
						continue
					}
					allTokenIDs = append(allTokenIDs, tokenIDs...)
				}
			}
		}
	} else {
		// Fetch all active markets
		markets, err := s.gamma.FetchMarkets(ctx, &gamma.Filter{
			Active: &active,
			Limit:  s.config.Discovery.MaxMarkets,
		})
		if err != nil {
			return fmt.Errorf("fetching markets: %w", err)
		}

		for _, market := range markets {
			tokenIDs, err := market.ParseTokenIDs()
			if err != nil {
				log.Printf("Warning: failed to parse token IDs for market %s: %v", market.ID, err)
				continue
			}
			allTokenIDs = append(allTokenIDs, tokenIDs...)
		}
	}

	// Limit total tokens
	if s.config.Discovery.MaxMarkets > 0 && len(allTokenIDs) > s.config.Discovery.MaxMarkets*2 {
		allTokenIDs = allTokenIDs[:s.config.Discovery.MaxMarkets*2]
	}

	s.mu.Lock()
	s.tokenIDs = allTokenIDs
	s.mu.Unlock()

	return nil
}

// handleMessages processes incoming WebSocket messages.
func (s *Service) handleMessages(messages []ws.WSMessage) {
	for i := range messages {
		if err := s.storage.Write(&messages[i]); err != nil {
			log.Printf("Error writing message: %v", err)
		}
	}
}

// Close shuts down the service.
func (s *Service) Close() error {
	if s.ws != nil {
		s.ws.Close()
	}
	if s.storage != nil {
		return s.storage.Close()
	}
	return nil
}
