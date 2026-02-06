// Command probe-ws is a CLI tool for exploring the Polymarket CLOB WebSocket feed.
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/johan/polymarket-collector/internal/ws"
)

func main() {
	tokens := flag.String("tokens", "", "Comma-separated list of token IDs to subscribe")
	duration := flag.Duration("duration", 0, "How long to run (0 = until Ctrl+C)")
	outputFile := flag.String("output", "", "Output file path (empty = stdout)")
	verbose := flag.Bool("v", false, "Verbose output")

	flag.Parse()

	if *tokens == "" {
		fmt.Println("Usage: probe-ws --tokens <id1,id2,...> [options]")
		fmt.Println()
		fmt.Println("Options:")
		flag.PrintDefaults()
		fmt.Println()
		fmt.Println("Examples:")
		fmt.Println("  probe-ws --tokens 83955612...,46434110...")
		fmt.Println("  probe-ws --tokens 83955612... --duration 30s")
		fmt.Println("  probe-ws --tokens 83955612... --output data.jsonl")
		os.Exit(1)
	}

	tokenList := strings.Split(*tokens, ",")
	for i := range tokenList {
		tokenList[i] = strings.TrimSpace(tokenList[i])
	}

	var out *os.File
	var err error
	if *outputFile != "" {
		out, err = os.Create(*outputFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error creating output file: %v\n", err)
			os.Exit(1)
		}
		defer out.Close()
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if *duration > 0 {
		ctx, cancel = context.WithTimeout(ctx, *duration)
		defer cancel()
	}

	// Handle Ctrl+C
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		fmt.Fprintln(os.Stderr, "\nShutting down...")
		cancel()
	}()

	messageCount := 0
	bookCount := 0
	priceChangeCount := 0

	handler := func(messages []ws.WSMessage) {
		for _, msg := range messages {
			messageCount++

			switch msg.EventType {
			case ws.EventTypeBook:
				bookCount++
				if *verbose {
					fmt.Fprintf(os.Stderr, "[%s] book: asset=%s bids=%d asks=%d\n",
						time.Now().Format("15:04:05"),
						truncateID(msg.AssetID),
						len(msg.Bids),
						len(msg.Asks))
				}
			case ws.EventTypePriceChange:
				priceChangeCount++
				if *verbose {
					fmt.Fprintf(os.Stderr, "[%s] price_change: market=%s changes=%d\n",
						time.Now().Format("15:04:05"),
						truncateID(msg.Market),
						len(msg.PriceChanges))
				}
			}

			// Output JSON
			data, _ := json.Marshal(msg)
			if out != nil {
				fmt.Fprintln(out, string(data))
			} else if !*verbose {
				fmt.Println(string(data))
			}
		}
	}

	client := ws.NewWSClient(handler)

	fmt.Fprintf(os.Stderr, "Connecting to WebSocket...\n")
	if err := client.Connect(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "Error connecting: %v\n", err)
		os.Exit(1)
	}
	defer client.Close()

	fmt.Fprintf(os.Stderr, "Subscribing to %d tokens...\n", len(tokenList))
	if err := client.Subscribe(tokenList); err != nil {
		fmt.Fprintf(os.Stderr, "Error subscribing: %v\n", err)
		os.Exit(1)
	}

	fmt.Fprintf(os.Stderr, "Listening... (Ctrl+C to stop)\n\n")

	<-ctx.Done()

	fmt.Fprintf(os.Stderr, "\n--- Summary ---\n")
	fmt.Fprintf(os.Stderr, "Total messages:  %d\n", messageCount)
	fmt.Fprintf(os.Stderr, "Book snapshots:  %d\n", bookCount)
	fmt.Fprintf(os.Stderr, "Price changes:   %d\n", priceChangeCount)

	if *outputFile != "" {
		fmt.Fprintf(os.Stderr, "Output written to: %s\n", *outputFile)
	}
}

func truncateID(id string) string {
	if len(id) > 20 {
		return id[:20] + "..."
	}
	return id
}
