package formatter

import (
	"strings"
	"testing"

	"github.com/josh-ninjatrader/awesome-n8n-docs/internal/parser"
)

func TestFormat_AddsDelimiters(t *testing.T) {
	sections := []parser.Section{
		{Title: "Slack node", Type: parser.NodeDoc, Content: "# Slack node\n\nSlack docs."},
		{Title: "Gmail node", Type: parser.NodeDoc, Content: "# Gmail node\n\nGmail docs."},
	}

	result := Format(sections)

	delimiter := "----------------------------------------"
	count := strings.Count(result, delimiter)
	// Delimiter between intro and first section, between sections, and after last section
	if count < 2 {
		t.Errorf("expected at least 2 delimiters, got %d", count)
	}
}

func TestFormat_HasIntroSummary(t *testing.T) {
	sections := []parser.Section{
		{Title: "Test node", Type: parser.NodeDoc, Content: "# Test node\n\nDocs."},
	}

	result := Format(sections)

	if !strings.HasPrefix(result, "# n8n Documentation") {
		t.Error("expected intro summary at start")
	}
	if !strings.Contains(result, "## Core Concepts") {
		t.Error("expected Core Concepts in intro")
	}
}

func TestFormat_PreservesContent(t *testing.T) {
	sections := []parser.Section{
		{Title: "Slack node", Type: parser.NodeDoc, Content: "# Slack node\n\nSlack docs.\n"},
	}

	result := Format(sections)

	if !strings.Contains(result, "Slack docs.") {
		t.Error("section content should be preserved")
	}
}
