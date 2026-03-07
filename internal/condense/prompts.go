package condense

import "github.com/josh-ninjatrader/awesome-n8n-docs/internal/parser"

func PromptFor(sectionType parser.SectionType) string {
	switch sectionType {
	case parser.NodeDoc:
		return `Condense this n8n node documentation into terse, structured bullet points. Use ### sub-headers for operation groups. Standardize verbs (use "Retrieve" not "Get information about"). Remove redundant context (e.g., don't repeat the node name in descriptions). Keep all parameter names, types, and valid values. Preserve all code examples verbatim. Output markdown only, no commentary.`
	case parser.Credential:
		return `Condense this n8n credential setup guide into concise step-by-step instructions. Keep all required field names, URLs, and configuration values. Remove conversational prose. Preserve all code examples verbatim. Output markdown only, no commentary.`
	case parser.APIRef:
		return `Condense this API reference, keeping all endpoint paths, methods, parameters, and response schemas. Remove redundant descriptions. Preserve all code examples and JSON schemas verbatim. Output markdown only, no commentary.`
	default:
		return `Condense this n8n documentation page to its essential information. Preserve all code examples, configuration values, and API endpoints verbatim. Remove filler prose and obvious statements. Output markdown only, no commentary.`
	}
}
