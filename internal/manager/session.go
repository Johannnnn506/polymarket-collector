package manager

import (
	"bufio"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"time"

	"github.com/johan/polymarket-collector/internal/gamma"
	"github.com/johan/polymarket-collector/internal/ws"
)

// MarketSession manages data collection for a single market instance.
type MarketSession struct {
	// Market identification
	SeriesSlug  string
	MarketID    string
	ConditionID string
	TokenIDs    []string
	EndDate     time.Time
	GracePeriod time.Duration

	// Output
	outputDir  string
	file       *os.File
	gzWriter   *gzip.Writer
	bufWriter  *bufio.Writer
	filePath   string
	useGzip    bool

	// WebSocket
	wsClient *ws.Client

	// State
	ctx          context.Context
	cancel       context.CancelFunc
	mu           sync.Mutex
	started      bool
	stopped      bool
	messageCount int64
	startTime    time.Time
}

// SessionMetadata is written at the start of each data file.
type SessionMetadata struct {
	Type        string    `json:"type"`
	SeriesSlug  string    `json:"series_slug"`
	MarketID    string    `json:"market_id"`
	ConditionID string    `json:"condition_id"`
	TokenIDs    []string  `json:"token_ids"`
	EndDate     time.Time `json:"end_date"`
	StartTime   time.Time `json:"start_time"`
}

// NewMarketSession creates a new session for collecting market data.
func NewMarketSession(market gamma.Market, seriesSlug, outputDir string, gracePeriod time.Duration, useGzip bool) (*MarketSession, error) {
	tokenIDs, err := market.ParseTokenIDs()
	if err != nil {
		return nil, fmt.Errorf("parsing token IDs: %w", err)
	}

	if len(tokenIDs) == 0 {
		return nil, fmt.Errorf("no token IDs found for market %s", market.ID)
	}

	return &MarketSession{
		SeriesSlug:  seriesSlug,
		MarketID:    market.ID,
		ConditionID: market.ConditionID,
		TokenIDs:    tokenIDs,
		EndDate:     market.EndDate,
		GracePeriod: gracePeriod,
		outputDir:   outputDir,
		useGzip:     useGzip,
	}, nil
}

// Start begins collecting data for this market.
func (s *MarketSession) Start(parentCtx context.Context) error {
	s.mu.Lock()
	if s.started {
		s.mu.Unlock()
		return nil
	}
	s.started = true
	s.startTime = time.Now()
	s.mu.Unlock()

	s.ctx, s.cancel = context.WithCancel(parentCtx)

	// Create output directory for this series
	seriesDir := filepath.Join(s.outputDir, s.shortSlug())
	if err := os.MkdirAll(seriesDir, 0755); err != nil {
		return fmt.Errorf("creating series directory: %w", err)
	}

	// Create output file named by date and end timestamp
	var filename string
	if s.useGzip {
		filename = fmt.Sprintf("%s_%d.jsonl.gz",
			s.EndDate.Format("2006-01-02"),
			s.EndDate.Unix())
	} else {
		filename = fmt.Sprintf("%s_%d.jsonl",
			s.EndDate.Format("2006-01-02"),
			s.EndDate.Unix())
	}
	s.filePath = filepath.Join(seriesDir, filename)

	f, err := os.Create(s.filePath)
	if err != nil {
		return fmt.Errorf("creating output file: %w", err)
	}
	s.file = f

	// Set up writers based on compression setting
	if s.useGzip {
		s.gzWriter = gzip.NewWriter(f)
		s.bufWriter = bufio.NewWriter(s.gzWriter)
	} else {
		s.bufWriter = bufio.NewWriter(f)
	}

	// Write metadata as first line
	meta := SessionMetadata{
		Type:        "metadata",
		SeriesSlug:  s.SeriesSlug,
		MarketID:    s.MarketID,
		ConditionID: s.ConditionID,
		TokenIDs:    s.TokenIDs,
		EndDate:     s.EndDate,
		StartTime:   s.startTime,
	}
	metaData, _ := json.Marshal(meta)
	s.bufWriter.Write(metaData)
	s.bufWriter.WriteString("\n")

	// Create WebSocket client
	s.wsClient = ws.NewWSClient(s.handleMessages)

	// Connect
	if err := s.wsClient.Connect(s.ctx); err != nil {
		s.file.Close()
		return fmt.Errorf("connecting WebSocket: %w", err)
	}

	// Subscribe to tokens
	if err := s.wsClient.Subscribe(s.TokenIDs); err != nil {
		s.wsClient.Close()
		s.file.Close()
		return fmt.Errorf("subscribing to tokens: %w", err)
	}

	log.Printf("[%s] Session started for market %s, ends at %s",
		s.shortSlug(), s.shortMarketID(), s.EndDate.Format("15:04:05"))

	return nil
}

// Stop gracefully stops the session.
func (s *MarketSession) Stop() error {
	s.mu.Lock()
	if s.stopped {
		s.mu.Unlock()
		return nil
	}
	s.stopped = true
	s.mu.Unlock()

	if s.cancel != nil {
		s.cancel()
	}

	if s.wsClient != nil {
		s.wsClient.Close()
	}

	// Close writers in correct order
	s.mu.Lock()
	if s.bufWriter != nil {
		s.bufWriter.Flush()
	}
	if s.gzWriter != nil {
		s.gzWriter.Close()
	}
	if s.file != nil {
		s.file.Close()
	}
	s.mu.Unlock()

	count := atomic.LoadInt64(&s.messageCount)
	log.Printf("[%s] Session stopped for market %s, collected %d messages",
		s.shortSlug(), s.shortMarketID(), count)

	return nil
}

// ShouldClose returns true if the session should be closed.
func (s *MarketSession) ShouldClose() bool {
	return time.Now().After(s.EndDate.Add(s.GracePeriod))
}

// MessageCount returns the number of messages collected.
func (s *MarketSession) MessageCount() int64 {
	return atomic.LoadInt64(&s.messageCount)
}

// FilePath returns the path to the output file.
func (s *MarketSession) FilePath() string {
	return s.filePath
}

// handleMessages processes incoming WebSocket messages.
func (s *MarketSession) handleMessages(messages []ws.WSMessage) {
	s.mu.Lock()
	writer := s.bufWriter
	stopped := s.stopped
	s.mu.Unlock()

	if writer == nil || stopped {
		return
	}

	for _, msg := range messages {
		data, err := json.Marshal(msg)
		if err != nil {
			log.Printf("[%s] Error marshaling message: %v", s.shortSlug(), err)
			continue
		}

		s.mu.Lock()
		if s.bufWriter != nil && !s.stopped {
			s.bufWriter.Write(data)
			s.bufWriter.WriteString("\n")
		}
		s.mu.Unlock()

		atomic.AddInt64(&s.messageCount, 1)
	}
}

// shortSlug returns a shortened version of the series slug for logging.
func (s *MarketSession) shortSlug() string {
	// Convert "eth-up-or-down-15m" to "eth-15m"
	slug := s.SeriesSlug
	if len(slug) > 20 && (slug[:3] == "eth" || slug[:3] == "btc") {
		crypto := slug[:3]
		// Find the timeframe at the end
		for _, tf := range []string{"15m", "hourly", "daily", "weekly", "monthly", "5m", "4h"} {
			if len(slug) > len(tf) && slug[len(slug)-len(tf):] == tf {
				return crypto + "-" + tf
			}
		}
	}
	return slug
}

// shortMarketID returns a shortened version of the market ID for logging.
func (s *MarketSession) shortMarketID() string {
	if len(s.MarketID) > 8 {
		return s.MarketID[:8]
	}
	return s.MarketID
}
