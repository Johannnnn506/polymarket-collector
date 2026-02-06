package manager

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/johan/polymarket-collector/internal/config"
	"github.com/johan/polymarket-collector/internal/gamma"
)

// MarketManager orchestrates data collection across multiple market sessions.
type MarketManager struct {
	gamma   *gamma.Client
	config  *config.ManagerConfig
	storage config.StorageConfig
	useGzip bool

	mu       sync.RWMutex
	sessions map[string]*MarketSession // key: marketID
}

// NewMarketManager creates a new market manager.
func NewMarketManager(gammaClient *gamma.Client, cfg *config.ManagerConfig, storageCfg config.StorageConfig, useGzip bool) *MarketManager {
	return &MarketManager{
		gamma:    gammaClient,
		config:   cfg,
		storage:  storageCfg,
		useGzip:  useGzip,
		sessions: make(map[string]*MarketSession),
	}
}

// Run starts the manager and runs until the context is cancelled.
func (m *MarketManager) Run(ctx context.Context) error {
	log.Println("Starting market manager...")

	// Initial scan
	if err := m.discoverMarkets(ctx); err != nil {
		log.Printf("Warning: initial market discovery failed: %v", err)
	}

	// Print initial status
	m.printStatus()

	ticker := time.NewTicker(m.config.ScanInterval)
	defer ticker.Stop()

	cleanupTicker := time.NewTicker(10 * time.Second)
	defer cleanupTicker.Stop()

	statusTicker := time.NewTicker(60 * time.Second)
	defer statusTicker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("Shutting down market manager...")
			m.stopAllSessions()
			return ctx.Err()

		case <-ticker.C:
			if err := m.discoverMarkets(ctx); err != nil {
				log.Printf("Warning: market discovery failed: %v", err)
			}

		case <-cleanupTicker.C:
			m.cleanupExpiredSessions()

		case <-statusTicker.C:
			m.printStatus()
		}
	}
}

// discoverMarkets scans for new markets in configured series.
func (m *MarketManager) discoverMarkets(ctx context.Context) error {
	for _, seriesCfg := range m.config.Series {
		if !seriesCfg.Enabled {
			continue
		}

		markets, err := m.gamma.FetchActiveMarketsForSeries(ctx, seriesCfg.Slug)
		if err != nil {
			log.Printf("[%s] Error fetching markets: %v", seriesCfg.Slug, err)
			continue
		}

		for _, market := range markets {
			m.mu.RLock()
			_, exists := m.sessions[market.ID]
			m.mu.RUnlock()

			if exists {
				continue
			}

			// Start new session
			if err := m.startSession(ctx, market, seriesCfg.Slug); err != nil {
				log.Printf("[%s] Error starting session for market %s: %v",
					seriesCfg.Slug, market.ID, err)
			}
		}
	}

	return nil
}

// startSession creates and starts a new market session.
func (m *MarketManager) startSession(ctx context.Context, market gamma.Market, seriesSlug string) error {
	session, err := NewMarketSession(market, seriesSlug, m.storage.OutputDir, m.config.GracePeriod, m.useGzip)
	if err != nil {
		return err
	}

	if err := session.Start(ctx); err != nil {
		return err
	}

	m.mu.Lock()
	m.sessions[market.ID] = session
	m.mu.Unlock()

	return nil
}

// cleanupExpiredSessions stops and removes sessions that have expired.
func (m *MarketManager) cleanupExpiredSessions() {
	m.mu.Lock()
	defer m.mu.Unlock()

	for id, session := range m.sessions {
		if session.ShouldClose() {
			session.Stop()
			delete(m.sessions, id)
		}
	}
}

// stopAllSessions stops all active sessions.
func (m *MarketManager) stopAllSessions() {
	m.mu.Lock()
	defer m.mu.Unlock()

	for id, session := range m.sessions {
		session.Stop()
		delete(m.sessions, id)
	}
}

// printStatus logs the current status of all sessions.
func (m *MarketManager) printStatus() {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if len(m.sessions) == 0 {
		log.Println("No active sessions")
		return
	}

	log.Printf("Active sessions: %d", len(m.sessions))
	for _, session := range m.sessions {
		remaining := time.Until(session.EndDate)
		if remaining < 0 {
			remaining = 0
		}
		log.Printf("  [%s] market=%s msgs=%d ends_in=%v",
			session.shortSlug(),
			session.shortMarketID(),
			session.MessageCount(),
			remaining.Round(time.Second))
	}
}

// SessionCount returns the number of active sessions.
func (m *MarketManager) SessionCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.sessions)
}

// GetSessions returns a copy of all active sessions.
func (m *MarketManager) GetSessions() []*MarketSession {
	m.mu.RLock()
	defer m.mu.RUnlock()

	sessions := make([]*MarketSession, 0, len(m.sessions))
	for _, s := range m.sessions {
		sessions = append(sessions, s)
	}
	return sessions
}
