// Package config provides configuration loading for the collector service.
package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// Config represents the collector configuration.
type Config struct {
	// Discovery settings
	Discovery DiscoveryConfig `yaml:"discovery"`

	// Storage settings
	Storage StorageConfig `yaml:"storage"`

	// WebSocket settings
	WebSocket WebSocketConfig `yaml:"websocket"`

	// Logging settings
	Logging LoggingConfig `yaml:"logging"`

	// Manager settings for cycle collector
	Manager ManagerConfig `yaml:"manager"`
}

// ManagerConfig contains settings for the cycle collector manager.
type ManagerConfig struct {
	// How often to scan for new markets
	ScanInterval time.Duration `yaml:"scan_interval"`

	// Grace period after market ends before closing session
	GracePeriod time.Duration `yaml:"grace_period"`

	// Series to track
	Series []SeriesConfig `yaml:"series"`
}

// SeriesConfig contains settings for a single series.
type SeriesConfig struct {
	// Series slug (e.g., "eth-up-or-down-15m")
	Slug string `yaml:"slug"`

	// Whether this series is enabled
	Enabled bool `yaml:"enabled"`
}

// DiscoveryConfig contains market discovery settings.
type DiscoveryConfig struct {
	// How often to refresh market list
	RefreshInterval time.Duration `yaml:"refresh_interval"`

	// Tag slugs to filter events
	Tags []string `yaml:"tags"`

	// Only include active markets
	ActiveOnly bool `yaml:"active_only"`

	// Maximum markets to track
	MaxMarkets int `yaml:"max_markets"`
}

// StorageConfig contains storage settings.
type StorageConfig struct {
	// Storage type: "file" or "none"
	Type string `yaml:"type"`

	// Output directory for file storage
	OutputDir string `yaml:"output_dir"`

	// File rotation interval
	RotationInterval time.Duration `yaml:"rotation_interval"`
}

// WebSocketConfig contains WebSocket settings.
type WebSocketConfig struct {
	// Custom WebSocket URL (optional)
	URL string `yaml:"url"`

	// Initial reconnection backoff
	InitialBackoff time.Duration `yaml:"initial_backoff"`

	// Maximum reconnection backoff
	MaxBackoff time.Duration `yaml:"max_backoff"`

	// Backoff multiplier
	BackoffFactor float64 `yaml:"backoff_factor"`
}

// LoggingConfig contains logging settings.
type LoggingConfig struct {
	// Log level: debug, info, warn, error
	Level string `yaml:"level"`

	// Log format: text or json
	Format string `yaml:"format"`
}

// DefaultConfig returns the default configuration.
func DefaultConfig() *Config {
	return &Config{
		Discovery: DiscoveryConfig{
			RefreshInterval: 5 * time.Minute,
			ActiveOnly:      true,
			MaxMarkets:      100,
		},
		Storage: StorageConfig{
			Type:             "file",
			OutputDir:        "data",
			RotationInterval: 1 * time.Hour,
		},
		WebSocket: WebSocketConfig{
			InitialBackoff: 1 * time.Second,
			MaxBackoff:     30 * time.Second,
			BackoffFactor:  2.0,
		},
		Logging: LoggingConfig{
			Level:  "info",
			Format: "text",
		},
		Manager: ManagerConfig{
			ScanInterval: 30 * time.Second,
			GracePeriod:  60 * time.Second,
		},
	}
}

// Load loads configuration from a YAML file.
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading config file: %w", err)
	}

	config := DefaultConfig()
	if err := yaml.Unmarshal(data, config); err != nil {
		return nil, fmt.Errorf("parsing config file: %w", err)
	}

	return config, nil
}

// Validate checks the configuration for errors.
func (c *Config) Validate() error {
	if c.Storage.Type != "file" && c.Storage.Type != "none" {
		return fmt.Errorf("invalid storage type: %s", c.Storage.Type)
	}
	if c.Storage.Type == "file" && c.Storage.OutputDir == "" {
		return fmt.Errorf("output_dir required for file storage")
	}
	return nil
}
