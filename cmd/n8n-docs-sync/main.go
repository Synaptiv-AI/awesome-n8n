package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/josh-ninjatrader/awesome-n8n-docs/internal/condense"
	"github.com/josh-ninjatrader/awesome-n8n-docs/internal/fetch"
	"github.com/josh-ninjatrader/awesome-n8n-docs/internal/filter"
	"github.com/josh-ninjatrader/awesome-n8n-docs/internal/formatter"
	"github.com/josh-ninjatrader/awesome-n8n-docs/internal/parser"
)

func main() {
	dryRun := flag.Bool("dry-run", false, "Show what would change without writing")
	noCondense := flag.Bool("no-condense", false, "Skip LLM condensation")
	output := flag.String("output", "n8n-docs-llms.txt", "Output file path")
	concurrency := flag.Int("concurrency", 5, "Number of concurrent claude -p workers")
	cacheDir := flag.String("cache-dir", ".cache", "Directory for condensation cache")
	sourceURL := flag.String("source-url", fetch.DefaultURL, "URL to fetch llms-full.txt from")
	flag.Parse()

	// Stage 1: Fetch
	fmt.Println("Fetching n8n documentation...")
	result, err := fetch.Fetch(*sourceURL, *cacheDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "fetch failed: %v\n", err)
		os.Exit(1)
	}
	if !result.Changed {
		fmt.Println("Documentation unchanged (ETag match). Nothing to do.")
		os.Exit(0)
	}
	fmt.Printf("Downloaded %d bytes.\n", len(result.Body))

	// Stage 2: Parse
	fmt.Println("Parsing sections...")
	sections := parser.Parse(result.Body)
	fmt.Printf("Parsed %d sections.\n", len(sections))

	// Stage 3: Filter
	fmt.Println("Filtering...")
	filtered := filter.Filter(sections)
	fmt.Printf("Filtered to %d sections (removed %d).\n", len(filtered), len(sections)-len(filtered))

	// Stage 4: Condense
	var condensed []parser.Section
	if *noCondense {
		fmt.Println("Skipping condensation (--no-condense).")
		condensed = filtered
	} else {
		fmt.Printf("Condensing %d sections with %d workers...\n", len(filtered), *concurrency)
		condensed, err = condense.Condense(filtered, *cacheDir, *concurrency, condense.ClaudeRunner)
		if err != nil {
			fmt.Fprintf(os.Stderr, "condense failed: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Condensation complete.")
	}

	// Stage 5: Format
	fmt.Println("Formatting output...")
	formatted := formatter.Format(condensed)

	// Stage 6: Output
	if *dryRun {
		fmt.Printf("Dry run: would write %d bytes to %s\n", len(formatted), *output)
		os.Exit(0)
	}

	if err := os.WriteFile(*output, []byte(formatted), 0644); err != nil {
		fmt.Fprintf(os.Stderr, "write failed: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Written %d bytes to %s\n", len(formatted), *output)
}
