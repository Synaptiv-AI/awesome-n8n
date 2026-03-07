package condense

import (
	"testing"

	"github.com/josh-ninjatrader/awesome-n8n-docs/internal/parser"
)

func TestCondense_WithMockRunner(t *testing.T) {
	sections := []parser.Section{
		{Title: "Slack node", Type: parser.NodeDoc, Content: "# Slack node\n\nVerbose docs about Slack."},
		{Title: "Gmail node", Type: parser.NodeDoc, Content: "# Gmail node\n\nVerbose docs about Gmail."},
	}

	// Mock runner that returns a short condensed version
	runner := func(content, prompt string) (string, error) {
		return "CONDENSED: " + content[:20], nil
	}

	cacheDir := t.TempDir()
	result, err := Condense(sections, cacheDir, 2, runner)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 2 {
		t.Fatalf("expected 2 sections, got %d", len(result))
	}
	for _, s := range result {
		if len(s.Content) < 10 || s.Content[:10] != "CONDENSED:" {
			t.Errorf("expected condensed content, got: %q", s.Content[:20])
		}
	}
}

func TestCondense_UsesCacheOnSecondRun(t *testing.T) {
	sections := []parser.Section{
		{Title: "Test node", Type: parser.NodeDoc, Content: "# Test node\n\nSome docs."},
	}

	callCount := 0
	runner := func(content, prompt string) (string, error) {
		callCount++
		return "condensed output", nil
	}

	cacheDir := t.TempDir()

	// First run — should call runner
	_, err := Condense(sections, cacheDir, 1, runner)
	if err != nil {
		t.Fatal(err)
	}
	if callCount != 1 {
		t.Errorf("expected 1 call, got %d", callCount)
	}

	// Second run — should use cache
	_, err = Condense(sections, cacheDir, 1, runner)
	if err != nil {
		t.Fatal(err)
	}
	if callCount != 1 {
		t.Errorf("expected still 1 call (cached), got %d", callCount)
	}
}

func TestCondense_QualityGuardrail_TooLarge(t *testing.T) {
	sections := []parser.Section{
		{Title: "Test node", Type: parser.NodeDoc, Content: "# Test node\n\nShort."},
	}

	// Runner returns output larger than 80% of input — should retry
	callCount := 0
	runner := func(content, prompt string) (string, error) {
		callCount++
		if callCount == 1 {
			return content, nil // returns 100% of input
		}
		return "short", nil // retry succeeds
	}

	cacheDir := t.TempDir()
	result, err := Condense(sections, cacheDir, 1, runner)
	if err != nil {
		t.Fatal(err)
	}
	if callCount != 2 {
		t.Errorf("expected 2 calls (retry), got %d", callCount)
	}
	if result[0].Content != "short" {
		t.Errorf("expected retried content, got: %q", result[0].Content)
	}
}
