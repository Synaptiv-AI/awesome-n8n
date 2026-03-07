package parser

import (
	"regexp"
	"strings"
)

type SectionType int

const (
	General SectionType = iota
	NodeDoc
	Credential
	ReleaseNote
	APIRef
)

func (t SectionType) String() string {
	switch t {
	case NodeDoc:
		return "node-doc"
	case Credential:
		return "credential"
	case ReleaseNote:
		return "release-note"
	case APIRef:
		return "api-ref"
	default:
		return "general"
	}
}

type Section struct {
	Title   string
	Content string
	Type    SectionType
}

var versionRe = regexp.MustCompile(`^n8n@\d+\.\d+`)

func Parse(input string) []Section {
	lines := strings.Split(input, "\n")
	var sections []Section
	var current *Section
	var buf strings.Builder

	for _, line := range lines {
		if strings.HasPrefix(line, "# ") && !strings.HasPrefix(line, "## ") {
			// Flush previous section
			if current != nil {
				current.Content = buf.String()
				current.Type = classify(current.Title, current.Content)
				sections = append(sections, *current)
			}
			title := strings.TrimPrefix(line, "# ")
			current = &Section{Title: title}
			buf.Reset()
		}
		if current != nil {
			buf.WriteString(line)
			buf.WriteString("\n")
		}
	}

	// Flush last section
	if current != nil {
		current.Content = buf.String()
		current.Type = classify(current.Title, current.Content)
		sections = append(sections, *current)
	}

	return sections
}

func classify(title, content string) SectionType {
	lower := strings.ToLower(title)

	if versionRe.MatchString(title) {
		return ReleaseNote
	}
	if strings.Contains(lower, "release note") || strings.Contains(lower, "release version") {
		return ReleaseNote
	}
	if strings.HasSuffix(lower, " node") || strings.HasSuffix(lower, " node (version 2)") {
		return NodeDoc
	}
	if strings.HasSuffix(lower, " credentials") {
		return Credential
	}
	if strings.HasPrefix(lower, "api") || strings.Contains(content, "/api/v1/") {
		return APIRef
	}

	return General
}
