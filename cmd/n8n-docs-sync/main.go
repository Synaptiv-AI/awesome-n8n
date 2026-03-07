package main

import (
	"flag"
	"fmt"
	"os"
)

func main() {
	dryRun := flag.Bool("dry-run", false, "Show what would change without writing")
	noCondense := flag.Bool("no-condense", false, "Skip LLM condensation")
	output := flag.String("output", "n8n-docs-llms.txt", "Output file path")
	concurrency := flag.Int("concurrency", 5, "Number of concurrent claude -p workers")
	cacheDir := flag.String("cache-dir", ".cache", "Directory for condensation cache")
	flag.Parse()

	fmt.Printf("n8n-docs-sync: dry-run=%v no-condense=%v output=%s concurrency=%d cache-dir=%s\n",
		*dryRun, *noCondense, *output, *concurrency, *cacheDir)
	os.Exit(0)
}
