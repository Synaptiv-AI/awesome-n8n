package parser

import (
	"testing"
)

func TestParse_SplitsSections(t *testing.T) {
	input := `# Welcome to n8n

Some intro text.

## Getting Started

More text here.

# Slack node

Slack node documentation.

## Operations

- Send message
- Get channel

# Slack credentials

How to set up Slack credentials.

# n8n@2.11.1

Release notes for 2.11.1.

# API Reference

## GET /workflows

Returns all workflows.
`

	sections := Parse(input)

	if len(sections) != 5 {
		t.Fatalf("expected 5 sections, got %d", len(sections))
	}

	tests := []struct {
		idx      int
		title    string
		sectType SectionType
	}{
		{0, "Welcome to n8n", General},
		{1, "Slack node", NodeDoc},
		{2, "Slack credentials", Credential},
		{3, "n8n@2.11.1", ReleaseNote},
		{4, "API Reference", APIRef},
	}

	for _, tt := range tests {
		s := sections[tt.idx]
		if s.Title != tt.title {
			t.Errorf("section %d: expected title %q, got %q", tt.idx, tt.title, s.Title)
		}
		if s.Type != tt.sectType {
			t.Errorf("section %d (%s): expected type %v, got %v", tt.idx, tt.title, tt.sectType, s.Type)
		}
	}
}

func TestParse_PreservesContent(t *testing.T) {
	input := "# Test node\n\nLine 1.\n\nLine 2.\n\n## Sub-heading\n\nMore content.\n"

	sections := Parse(input)
	if len(sections) != 1 {
		t.Fatalf("expected 1 section, got %d", len(sections))
	}

	expected := "# Test node\n\nLine 1.\n\nLine 2.\n\n## Sub-heading\n\nMore content.\n\n"
	if sections[0].Content != expected {
		t.Errorf("content mismatch:\nexpected: %q\ngot:      %q", expected, sections[0].Content)
	}
}

func TestClassify(t *testing.T) {
	tests := []struct {
		title    string
		content  string
		expected SectionType
	}{
		{"Slack node", "", NodeDoc},
		{"HTTP Request node", "", NodeDoc},
		{"Slack credentials", "", Credential},
		{"n8n@2.11.1", "", ReleaseNote},
		{"n8n@1.0.0", "", ReleaseNote},
		{"API Reference", "## GET /api/v1/workflows", APIRef},
		{"Getting Started", "Some intro text.", General},
		{"Release notes", "version history", ReleaseNote},
	}

	for _, tt := range tests {
		got := classify(tt.title, tt.content)
		if got != tt.expected {
			t.Errorf("classify(%q): expected %v, got %v", tt.title, tt.expected, got)
		}
	}
}
