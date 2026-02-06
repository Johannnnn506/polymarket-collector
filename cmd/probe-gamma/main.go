// Command probe-gamma is a CLI tool for exploring the Polymarket Gamma API.
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

	"github.com/johan/polymarket-collector/internal/gamma"
)

func main() {
	listSeries := flag.Bool("list-series", false, "List active series")
	listEvents := flag.Bool("list-events", false, "List active events")
	listMarkets := flag.Bool("list-markets", false, "List active markets")
	tag := flag.String("tag", "", "Filter by tag slug")
	slug := flag.String("slug", "", "Filter by slug")
	limit := flag.Int("limit", 10, "Maximum number of results")
	output := flag.String("output", "table", "Output format: table or json")
	timeout := flag.Duration("timeout", 30*time.Second, "Request timeout")

	flag.Parse()

	if !*listSeries && !*listEvents && !*listMarkets && *tag == "" && *slug == "" {
		fmt.Println("Usage: probe-gamma [options]")
		fmt.Println()
		fmt.Println("Options:")
		flag.PrintDefaults()
		fmt.Println()
		fmt.Println("Examples:")
		fmt.Println("  probe-gamma --list-series")
		fmt.Println("  probe-gamma --tag bitcoin")
		fmt.Println("  probe-gamma --list-markets --output json")
		os.Exit(1)
	}

	client := gamma.NewClient(&http.Client{Timeout: *timeout})
	ctx, cancel := context.WithTimeout(context.Background(), *timeout)
	defer cancel()

	active := true
	filter := &gamma.Filter{
		Active:  &active,
		TagSlug: *tag,
		Slug:    *slug,
		Limit:   *limit,
	}

	switch {
	case *listSeries:
		series, err := client.FetchSeries(ctx, filter)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		outputSeries(series, *output)

	case *listEvents, *tag != "":
		events, err := client.FetchEvents(ctx, filter)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		outputEvents(events, *output)

	case *listMarkets:
		markets, err := client.FetchMarkets(ctx, filter)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		outputMarkets(markets, *output)
	}
}

func outputSeries(series []gamma.Series, format string) {
	if format == "json" {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		enc.Encode(series)
		return
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tSLUG\tTITLE\tTYPE\tACTIVE\tVOLUME24H")
	for _, s := range series {
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%v\t%.2f\n",
			s.ID, s.Slug, truncate(s.Title, 40), s.SeriesType, s.Active, s.Volume24hr)
	}
	w.Flush()
	fmt.Printf("\nTotal: %d series\n", len(series))
}

func outputEvents(events []gamma.Event, format string) {
	if format == "json" {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		enc.Encode(events)
		return
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tSLUG\tTITLE\tMARKETS\tACTIVE\tVOLUME24H")
	for _, e := range events {
		fmt.Fprintf(w, "%s\t%s\t%s\t%d\t%v\t%.2f\n",
			e.ID, e.Slug, truncate(e.Title, 40), len(e.Markets), e.Active, e.Volume24hr)
	}
	w.Flush()
	fmt.Printf("\nTotal: %d events\n", len(events))
}

func outputMarkets(markets []gamma.Market, format string) {
	if format == "json" {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		enc.Encode(markets)
		return
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tQUESTION\tTOKENS\tACTIVE\tVOLUME24H")
	for _, m := range markets {
		tokens, _ := m.ParseTokenIDs()
		fmt.Fprintf(w, "%s\t%s\t%d\t%v\t%.2f\n",
			m.ID, truncate(m.Question, 50), len(tokens), m.Active, m.Volume24hr)
	}
	w.Flush()
	fmt.Printf("\nTotal: %d markets\n", len(markets))
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
