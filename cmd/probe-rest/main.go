// Command probe-rest is a CLI tool for exploring the Polymarket CLOB REST API.
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"text/tabwriter"
	"time"

	"github.com/johan/polymarket-collector/internal/clob"
)

func main() {
	token := flag.String("token", "", "Token ID to fetch order book")
	watch := flag.Bool("watch", false, "Continuously poll for updates")
	interval := flag.Duration("interval", 5*time.Second, "Poll interval (with --watch)")
	midpoint := flag.Bool("midpoint", false, "Fetch midpoint price only")
	spread := flag.Bool("spread", false, "Fetch spread only")
	output := flag.String("output", "table", "Output format: table or json")
	timeout := flag.Duration("timeout", 30*time.Second, "Request timeout")

	flag.Parse()

	if *token == "" {
		fmt.Println("Usage: probe-rest --token <token_id> [options]")
		fmt.Println()
		fmt.Println("Options:")
		flag.PrintDefaults()
		fmt.Println()
		fmt.Println("Examples:")
		fmt.Println("  probe-rest --token 83955612... ")
		fmt.Println("  probe-rest --token 83955612... --watch --interval 2s")
		fmt.Println("  probe-rest --token 83955612... --midpoint")
		os.Exit(1)
	}

	client := clob.NewClient(&http.Client{Timeout: *timeout})

	if *midpoint {
		fetchMidpoint(client, *token, *timeout)
		return
	}

	if *spread {
		fetchSpread(client, *token, *timeout)
		return
	}

	if *watch {
		watchBook(client, *token, *interval, *output, *timeout)
		return
	}

	fetchBook(client, *token, *output, *timeout)
}

func fetchBook(client *clob.Client, tokenID, format string, timeout time.Duration) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	book, err := client.FetchBook(ctx, tokenID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	outputBook(book, format)
}

func fetchMidpoint(client *clob.Client, tokenID string, timeout time.Duration) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	mid, err := client.FetchMidpoint(ctx, tokenID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Midpoint: %s\n", mid)
}

func fetchSpread(client *clob.Client, tokenID string, timeout time.Duration) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	spread, err := client.FetchSpread(ctx, tokenID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Spread: %s\n", spread)
}

func watchBook(client *clob.Client, tokenID string, interval time.Duration, format string, timeout time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	fmt.Printf("Watching token %s... (Ctrl+C to stop)\n\n", tokenID[:20]+"...")

	for {
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		book, err := client.FetchBook(ctx, tokenID)
		cancel()

		if err != nil {
			fmt.Fprintf(os.Stderr, "[%s] Error: %v\n", time.Now().Format("15:04:05"), err)
		} else {
			fmt.Printf("\n[%s]\n", time.Now().Format("15:04:05"))
			outputBook(book, format)
		}

		<-ticker.C
	}
}

func outputBook(book *clob.BookSnapshot, format string) {
	if format == "json" {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		enc.Encode(book)
		return
	}

	fmt.Printf("Market: %s\n", book.Market)
	fmt.Printf("Asset:  %s\n", book.AssetID[:20]+"...")
	fmt.Printf("Time:   %s\n", book.Timestamp)
	fmt.Printf("Hash:   %s\n", book.Hash)
	fmt.Printf("Last:   %s\n", book.LastTradePrice)
	fmt.Println()

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "BIDS\t\tASKS")
	fmt.Fprintln(w, "PRICE\tSIZE\tPRICE\tSIZE")

	maxRows := len(book.Bids)
	if len(book.Asks) > maxRows {
		maxRows = len(book.Asks)
	}
	if maxRows > 10 {
		maxRows = 10
	}

	for i := 0; i < maxRows; i++ {
		var bidPrice, bidSize, askPrice, askSize string
		if i < len(book.Bids) {
			bidPrice = book.Bids[i].Price
			bidSize = book.Bids[i].Size
		}
		if i < len(book.Asks) {
			askPrice = book.Asks[i].Price
			askSize = book.Asks[i].Size
		}
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", bidPrice, bidSize, askPrice, askSize)
	}
	w.Flush()

	if len(book.Bids) > 10 || len(book.Asks) > 10 {
		fmt.Printf("\n... showing top 10 of %d bids and %d asks\n", len(book.Bids), len(book.Asks))
	}
}
