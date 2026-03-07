package condense

import (
	"fmt"
	"sync"

	"github.com/josh-ninjatrader/awesome-n8n-docs/internal/parser"
)

// Runner executes an LLM condensation. Takes section content and prompt, returns condensed output.
type Runner func(content, prompt string) (string, error)

func Condense(sections []parser.Section, cacheDir string, concurrency int, run Runner) ([]parser.Section, error) {
	cache := NewCache(cacheDir)
	result := make([]parser.Section, len(sections))
	errs := make([]error, len(sections))

	sem := make(chan struct{}, concurrency)
	var wg sync.WaitGroup

	for i, s := range sections {
		wg.Add(1)
		go func(idx int, sect parser.Section) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			condensed, err := condenseOne(sect, cache, run)
			if err != nil {
				errs[idx] = fmt.Errorf("condensing %q: %w", sect.Title, err)
				return
			}
			sect.Content = condensed
			result[idx] = sect
		}(i, s)
	}

	wg.Wait()

	// Check for errors
	for _, err := range errs {
		if err != nil {
			return nil, err
		}
	}

	return result, nil
}

func condenseOne(s parser.Section, cache *Cache, run Runner) (string, error) {
	// Check cache
	if cached, ok := cache.Get(s.Content); ok {
		return cached, nil
	}

	prompt := PromptFor(s.Type)
	condensed, err := run(s.Content, prompt)
	if err != nil {
		return "", err
	}

	// Quality guardrail: if output is >80% of input, retry with stricter prompt
	if len(condensed) > 0 && float64(len(condensed))/float64(len(s.Content)) > 0.8 {
		stricter := prompt + " Be much more aggressive in condensing. Remove all unnecessary words."
		condensed, err = run(s.Content, stricter)
		if err != nil {
			return "", err
		}
	}

	// Quality guardrail: if output is <10% of input, log warning (don't cache)
	if len(condensed) > 0 && float64(len(condensed))/float64(len(s.Content)) < 0.1 {
		fmt.Printf("WARNING: section %q condensed to <10%% of original (%d -> %d bytes)\n",
			s.Title, len(s.Content), len(condensed))
	}

	cache.Put(s.Content, condensed)
	return condensed, nil
}
