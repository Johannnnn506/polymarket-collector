// Command collector is the main data collection service.
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/johan/polymarket-collector/internal/collector"
	"github.com/johan/polymarket-collector/internal/config"
)

func main() {
	configPath := flag.String("config", "config.yaml", "Path to configuration file")
	flag.Parse()

	// Load configuration
	cfg, err := config.Load(*configPath)
	if err != nil {
		// If config file doesn't exist, use defaults
		if os.IsNotExist(err) {
			log.Printf("Config file not found, using defaults")
			cfg = config.DefaultConfig()
		} else {
			log.Fatalf("Error loading config: %v", err)
		}
	}

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		log.Fatalf("Invalid configuration: %v", err)
	}

	// Create service
	svc, err := collector.NewService(cfg)
	if err != nil {
		log.Fatalf("Error creating service: %v", err)
	}
	defer svc.Close()

	// Setup context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle shutdown signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-sigChan
		fmt.Printf("\nReceived signal %v, shutting down...\n", sig)
		cancel()
	}()

	// Run the service
	if err := svc.Run(ctx); err != nil {
		if ctx.Err() != context.Canceled {
			log.Fatalf("Service error: %v", err)
		}
	}

	log.Println("Collector shutdown complete")
}
