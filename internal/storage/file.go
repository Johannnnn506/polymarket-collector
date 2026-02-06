package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/johan/polymarket-collector/internal/ws"
)

// FileStorage writes messages to JSONL files with rotation.
type FileStorage struct {
	outputDir        string
	rotationInterval time.Duration

	mu           sync.Mutex
	currentFile  *os.File
	currentPath  string
	lastRotation time.Time
	messageCount int64
}

// NewFileStorage creates a new file storage.
func NewFileStorage(outputDir string, rotationInterval time.Duration) (*FileStorage, error) {
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return nil, fmt.Errorf("creating output directory: %w", err)
	}

	s := &FileStorage{
		outputDir:        outputDir,
		rotationInterval: rotationInterval,
	}

	if err := s.rotate(); err != nil {
		return nil, err
	}

	return s, nil
}

// Write writes a message to the current file.
func (s *FileStorage) Write(msg *ws.WSMessage) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check if rotation is needed
	if s.rotationInterval > 0 && time.Since(s.lastRotation) > s.rotationInterval {
		if err := s.rotate(); err != nil {
			return err
		}
	}

	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("marshaling message: %w", err)
	}

	if _, err := s.currentFile.Write(data); err != nil {
		return fmt.Errorf("writing message: %w", err)
	}
	if _, err := s.currentFile.WriteString("\n"); err != nil {
		return fmt.Errorf("writing newline: %w", err)
	}

	s.messageCount++
	return nil
}

// Close closes the current file.
func (s *FileStorage) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.currentFile != nil {
		return s.currentFile.Close()
	}
	return nil
}

// rotate creates a new output file.
func (s *FileStorage) rotate() error {
	if s.currentFile != nil {
		s.currentFile.Close()
	}

	filename := fmt.Sprintf("orderbook_%s.jsonl", time.Now().UTC().Format("2006-01-02_15-04-05"))
	path := filepath.Join(s.outputDir, filename)

	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("creating output file: %w", err)
	}

	s.currentFile = f
	s.currentPath = path
	s.lastRotation = time.Now()
	s.messageCount = 0

	return nil
}

// CurrentPath returns the path to the current output file.
func (s *FileStorage) CurrentPath() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.currentPath
}

// MessageCount returns the number of messages written to the current file.
func (s *FileStorage) MessageCount() int64 {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.messageCount
}
