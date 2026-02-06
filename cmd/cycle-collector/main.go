package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/johan/polymarket-collector/internal/config"
	"github.com/johan/polymarket-collector/internal/gamma"
	"github.com/johan/polymarket-collector/internal/manager"
)

func main() {
	configPath := flag.String("config", "config.cycle.yaml", "Path to configuration file")
	outputDir := flag.String("output", "", "Override output directory")
	noGzip := flag.Bool("no-gzip", false, "Disable gzip compression (enabled by default)")
	flag.Parse()

	useGzip := !*noGzip

	// Load configuration
	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatalf("Error loading config: %v", err)
	}

	// Override output directory if specified
	if *outputDir != "" {
		cfg.Storage.OutputDir = *outputDir
	}

	// Validate configuration
	if len(cfg.Manager.Series) == 0 {
		log.Fatal("No series configured in manager.series")
	}

	enabledCount := 0
	for _, s := range cfg.Manager.Series {
		if s.Enabled {
			enabledCount++
			log.Printf("Tracking series: %s", s.Slug)
		}
	}
	if enabledCount == 0 {
		log.Fatal("No series enabled in configuration")
	}

	// Create HTTP client with timeout
	httpClient := &http.Client{
		Timeout: 30 * time.Second,
	}

	// Create Gamma client
	gammaClient := gamma.NewClient(httpClient)

	// Create market manager
	mgr := manager.NewMarketManager(gammaClient, &cfg.Manager, cfg.Storage, useGzip)

	// Setup signal handling
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-sigCh
		log.Printf("Received signal %v, shutting down...", sig)
		cancel()
	}()

	// Run the manager
	log.Printf("Starting cycle collector with %d series...", enabledCount)
	log.Printf("Output directory: %s", cfg.Storage.OutputDir)
	log.Printf("Gzip compression: %v", useGzip)
	log.Printf("Scan interval: %v", cfg.Manager.ScanInterval)
	log.Printf("Grace period: %v", cfg.Manager.GracePeriod)

	if err := mgr.Run(ctx); err != nil && err != context.Canceled {
		log.Fatalf("Manager error: %v", err)
	}

	log.Println("Cycle collector stopped")
}
