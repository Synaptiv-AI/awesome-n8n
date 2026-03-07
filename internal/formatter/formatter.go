package formatter

import (
	"strings"

	"github.com/josh-ninjatrader/awesome-n8n-docs/internal/parser"
)

const delimiter = "----------------------------------------"

const introSummary = `# n8n Documentation

## Core Concepts

n8n is a fair-code licensed workflow automation tool that enables the connection of any app with an API to manipulate data with minimal coding effort. It supports highly flexible workflows, customizable nodes, and can be used via npm, Docker, or self-hosting.

## Node & Parameter Details

Standard nodes include integrations with services such as Google Sheets, Telegram, MySQL, Slack, Discord, Postgres, Notion, Gmail, Airtable, and Google Drive, among hundreds of others.

## Best Practices

- Use n8n's quickstart guides to rapidly understand and implement workflows.
- Select the appropriate hosting method (Cloud, npm, self-host) based on your needs.
- Explore n8n's integration library to find suitable applications to automate.

## Workflow Patterns

Key workflow patterns include:
- Creating an API endpoint
- AI agent chat
- Web scraping and summarization with AI
- Joining different datasets
- Backing up workflows to GitHub
- JSON to Excel conversion

## API Documentation

For specific API endpoints, methods, authentication, request, and response details, refer to the [API Documentation](https://docs.n8n.io/api/).`

func Format(sections []parser.Section) string {
	var b strings.Builder

	// Intro
	b.WriteString(introSummary)
	b.WriteString("\n\n")
	b.WriteString(delimiter)
	b.WriteString("\n\n")

	// Sections with delimiters
	for i, s := range sections {
		b.WriteString(strings.TrimSpace(s.Content))
		b.WriteString("\n\n")
		b.WriteString(delimiter)
		if i < len(sections)-1 {
			b.WriteString("\n\n")
		} else {
			b.WriteString("\n")
		}
	}

	return b.String()
}
