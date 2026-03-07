package fetch

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestFetch_Downloads(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("ETag", `"abc123"`)
		w.Write([]byte("# n8n Docs\n\nSome content here."))
	}))
	defer server.Close()

	cacheDir := t.TempDir()
	result, err := Fetch(server.URL, cacheDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Body != "# n8n Docs\n\nSome content here." {
		t.Errorf("unexpected body: %q", result.Body)
	}
	if result.Changed != true {
		t.Error("expected Changed=true on first fetch")
	}
}

func TestFetch_SkipsWhenUnchanged(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("If-None-Match") == `"abc123"` {
			w.WriteHeader(http.StatusNotModified)
			return
		}
		w.Header().Set("ETag", `"abc123"`)
		w.Write([]byte("# n8n Docs\n\nSome content here."))
	}))
	defer server.Close()

	cacheDir := t.TempDir()
	// Write cached ETag
	os.WriteFile(filepath.Join(cacheDir, "etag"), []byte(`"abc123"`), 0644)

	result, err := Fetch(server.URL, cacheDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Changed != false {
		t.Error("expected Changed=false when ETag matches")
	}
}
