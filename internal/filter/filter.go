package filter

import (
	"regexp"
	"strings"

	"github.com/josh-ninjatrader/awesome-n8n-docs/internal/parser"
)

var (
	// Templates and examples block: starts with "### Templates and examples"
	// and ends with the "search all templates" link
	templatesRe = regexp.MustCompile(
		`(?ms)### Templates and examples\n\n.*?\[search all templates\]\(https://n8n\.io/workflows/\)\n*`,
	)

	// Parameter resolution in sub-nodes boilerplate
	paramResolutionRe = regexp.MustCompile(
		`(?ms)##? Parameter resolution in sub-nodes\n\nSub-nodes behave differently.*?always resolves to the first name\.\n*`,
	)

	// What to do if your operation isn't supported
	unsupportedOpRe = regexp.MustCompile(
		`(?ms)### What to do if your operation isn't supported\n\nIf this node doesn't support.*?Refer to \[Custom API operations\]\(https://docs\.n8n\.io/integrations/custom-operations/\) for more information\.\n*`,
	)

	// Human-in-the-loop for AI tool calls
	humanLoopRe = regexp.MustCompile(
		`(?ms)##? Human-in-the-loop for AI tool calls\n\n.*?human-in-the-loop-tools/\)\.\n*`,
	)
)

func Filter(sections []parser.Section) []parser.Section {
	var result []parser.Section
	for _, s := range sections {
		if s.Type == parser.ReleaseNote {
			continue
		}
		s.Content = stripBoilerplate(s.Content)
		result = append(result, s)
	}
	return result
}

func stripBoilerplate(content string) string {
	content = templatesRe.ReplaceAllString(content, "")
	content = paramResolutionRe.ReplaceAllString(content, "")
	content = unsupportedOpRe.ReplaceAllString(content, "")
	content = humanLoopRe.ReplaceAllString(content, "")

	// Clean up excessive blank lines left after removals
	for strings.Contains(content, "\n\n\n") {
		content = strings.ReplaceAll(content, "\n\n\n", "\n\n")
	}

	return content
}
