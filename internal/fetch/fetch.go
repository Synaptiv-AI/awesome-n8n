package fetch

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

const DefaultURL = "https://docs.n8n.io/llms-full.txt"

type Result struct {
	Body    string
	Changed bool
}

func Fetch(url string, cacheDir string) (*Result, error) {
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return nil, fmt.Errorf("creating cache dir: %w", err)
	}

	etagPath := filepath.Join(cacheDir, "etag")

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	// Send If-None-Match if we have a cached ETag
	if cachedETag, err := os.ReadFile(etagPath); err == nil {
		req.Header.Set("If-None-Match", string(cachedETag))
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetching %s: %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotModified {
		return &Result{Changed: false}, nil
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading body: %w", err)
	}

	// Cache the ETag for next run
	if etag := resp.Header.Get("ETag"); etag != "" {
		os.WriteFile(etagPath, []byte(etag), 0644)
	}

	return &Result{Body: string(body), Changed: true}, nil
}
