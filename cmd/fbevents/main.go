package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	fbevents "github.com/sylwek/fb-event-scraper"
)

func main() {
	maxEvents := flag.Int("max_events", 0, "maximum number of events to scrape (0 = no limit)")
	includePast := flag.Bool("include_past_events", false, "include past events (default: upcoming only)")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: fbevents [flags] <url>\n\nFlags:\n")
		flag.PrintDefaults()
	}
	flag.Parse()

	if flag.NArg() != 1 {
		flag.Usage()
		os.Exit(1)
	}

	url := flag.Arg(0)

	opts := &fbevents.Options{
		MaxEvents:         *maxEvents,
		IncludePastEvents: *includePast,
	}

	events, err := fbevents.ScrapeFbEventListFull(url, opts)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	out, err := json.MarshalIndent(events, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "error encoding JSON: %v\n", err)
		os.Exit(1)
	}

	fmt.Println(string(out))
}
