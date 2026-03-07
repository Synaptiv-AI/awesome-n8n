package condense

import (
	"testing"
)

func TestCache_MissAndHit(t *testing.T) {
	dir := t.TempDir()
	c := NewCache(dir)

	content := "# Slack node\n\nSome documentation."

	// Miss
	result, ok := c.Get(content)
	if ok {
		t.Error("expected cache miss")
	}
	if result != "" {
		t.Error("expected empty result on miss")
	}

	// Store
	c.Put(content, "# Slack node\n\n- Condensed docs.")

	// Hit
	result, ok = c.Get(content)
	if !ok {
		t.Error("expected cache hit")
	}
	if result != "# Slack node\n\n- Condensed docs." {
		t.Errorf("unexpected cached result: %q", result)
	}
}

func TestCache_DifferentContentMisses(t *testing.T) {
	dir := t.TempDir()
	c := NewCache(dir)

	c.Put("content A", "condensed A")

	_, ok := c.Get("content B")
	if ok {
		t.Error("expected cache miss for different content")
	}
}
