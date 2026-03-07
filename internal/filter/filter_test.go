package filter

import (
	"strings"
	"testing"

	"github.com/josh-ninjatrader/awesome-n8n-docs/internal/parser"
)

func TestFilter_RemovesReleaseNotes(t *testing.T) {
	sections := []parser.Section{
		{Title: "Slack node", Type: parser.NodeDoc, Content: "# Slack node\n\nSlack docs."},
		{Title: "n8n@2.11.1", Type: parser.ReleaseNote, Content: "# n8n@2.11.1\n\nRelease notes."},
		{Title: "Getting Started", Type: parser.General, Content: "# Getting Started\n\nIntro."},
	}

	result := Filter(sections)
	if len(result) != 2 {
		t.Fatalf("expected 2 sections, got %d", len(result))
	}
	for _, s := range result {
		if s.Type == parser.ReleaseNote {
			t.Error("release note section should have been removed")
		}
	}
}

func TestFilter_StripsTemplatesBlock(t *testing.T) {
	content := `# Slack node

Some useful docs.

### Templates and examples

**Cool Workflow**by Author

[View template details](https://n8n.io/workflows/1234-cool-workflow/)

[Browse Slack integration templates](https://n8n.io/integrations/slack/), or [search all templates](https://n8n.io/workflows/)

## Operations

- Send message
`

	sections := []parser.Section{
		{Title: "Slack node", Type: parser.NodeDoc, Content: content},
	}

	result := Filter(sections)
	if len(result) != 1 {
		t.Fatalf("expected 1 section, got %d", len(result))
	}

	if strings.Contains(result[0].Content, "Templates and examples") {
		t.Error("templates block should have been stripped")
	}
	if !strings.Contains(result[0].Content, "Send message") {
		t.Error("operations content should be preserved")
	}
}

func TestFilter_StripsParameterResolutionBoilerplate(t *testing.T) {
	content := "# Some AI node\n\nUseful AI docs.\n\n## Parameter resolution in sub-nodes\n\nSub-nodes behave differently to other nodes when processing multiple items using an expression.\n\nMost nodes, including root nodes, take any number of items as input, process these items, and output the results. You can use expressions to refer to input items, and the node resolves the expression for each item in turn. For example, given an input of five `name` values, the expression `{{ $json.name }}` resolves to each name in turn.\n\nIn sub-nodes, the expression always resolves to the first item. For example, given an input of five `name` values, the expression `{{ $json.name }}` always resolves to the first name.\n\n## Next section\n\nKeep this.\n"

	sections := []parser.Section{
		{Title: "Some AI node", Type: parser.NodeDoc, Content: content},
	}

	result := Filter(sections)
	if strings.Contains(result[0].Content, "Parameter resolution in sub-nodes") {
		t.Error("parameter resolution boilerplate should have been stripped")
	}
	if !strings.Contains(result[0].Content, "Keep this.") {
		t.Error("next section content should be preserved")
	}
}

func TestFilter_StripsUnsupportedOpBlock(t *testing.T) {
	content := `# Slack node

Some docs.

### What to do if your operation isn't supported

If this node doesn't support the operation you want to do, you can use the [HTTP Request node](https://docs.n8n.io/integrations/builtin/core-nodes/n8n-nodes-base.httprequest/) to call the service's API.

You can use the credential you created for this service in the HTTP Request node:

1. In the HTTP Request node, select **Authentication** > **Predefined Credential Type**.
1. Select the service you want to connect to.
1. Select your credential.

Refer to [Custom API operations](https://docs.n8n.io/integrations/custom-operations/) for more information.

## Next part

More content.
`

	sections := []parser.Section{
		{Title: "Slack node", Type: parser.NodeDoc, Content: content},
	}

	result := Filter(sections)
	if strings.Contains(result[0].Content, "What to do if your operation isn't supported") {
		t.Error("unsupported op boilerplate should have been stripped")
	}
	if !strings.Contains(result[0].Content, "More content.") {
		t.Error("next part content should be preserved")
	}
}

func TestFilter_StripsHumanInTheLoopBlock(t *testing.T) {
	content := `# Slack node

Some docs.

## Human-in-the-loop for AI tool calls

This node can be used as a human review step for AI Agent tool calls. When configured this way, the AI Agent will pause and request human approval through this service before executing tools that require oversight. Learn more in [Human-in-the-loop for AI tool calls](https://docs.n8n.io/advanced-ai/human-in-the-loop-tools/).

## Next part

More content.
`

	sections := []parser.Section{
		{Title: "Slack node", Type: parser.NodeDoc, Content: content},
	}

	result := Filter(sections)
	if strings.Contains(result[0].Content, "Human-in-the-loop for AI tool calls") {
		t.Error("human-in-the-loop boilerplate should have been stripped")
	}
	if !strings.Contains(result[0].Content, "More content.") {
		t.Error("next part content should be preserved")
	}
}
