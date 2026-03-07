# n8n-docs-sync Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Build a Go CLI tool that downloads n8n's official `llms-full.txt`, processes it through a pipeline (parse, filter, condense via `claude -p`, format), and outputs an optimized `n8n-docs-llms.txt` — automated weekly via GitHub Actions.

**Architecture:** Single Go binary with a six-stage pipeline: Fetch -> Parse -> Filter -> Condense -> Format -> Output. Each stage is an independent package under `internal/`. The condense stage shells out to `claude -p` with a worker pool and content-hash cache. Standard library only, no external deps.

**Tech Stack:** Go (stdlib only), `claude -p` CLI for LLM condensation, GitHub Actions for CI

**Design Doc:** `docs/plans/2026-03-07-n8n-docs-sync-design.md`

---

### Task 1: Go Module & Project Scaffolding

**Files:**
- Create: `go.mod`
- Create: `cmd/n8n-docs-sync/main.go`

**Step 1: Initialize Go module**

Run: `cd /Users/josh.delsman/Projects/awesome-n8n-docs && go mod init github.com/josh-ninjatrader/awesome-n8n-docs`
Expected: `go.mod` created

**Step 2: Write minimal main.go with flag parsing**

Create `cmd/n8n-docs-sync/main.go`:

```go
package main

import (
	"flag"
	"fmt"
	"os"
)

func main() {
	dryRun := flag.Bool("dry-run", false, "Show what would change without writing")
	noCondense := flag.Bool("no-condense", false, "Skip LLM condensation")
	output := flag.String("output", "n8n-docs-llms.txt", "Output file path")
	concurrency := flag.Int("concurrency", 5, "Number of concurrent claude -p workers")
	cacheDir := flag.String("cache-dir", ".cache", "Directory for condensation cache")
	flag.Parse()

	fmt.Printf("n8n-docs-sync: dry-run=%v no-condense=%v output=%s concurrency=%d cache-dir=%s\n",
		*dryRun, *noCondense, *output, *concurrency, *cacheDir)
	os.Exit(0)
}
```

**Step 3: Verify it builds and runs**

Run: `go build -o n8n-docs-sync ./cmd/n8n-docs-sync && ./n8n-docs-sync --help`
Expected: Flag usage output showing all five flags

**Step 4: Add binary and cache to .gitignore**

Append to `.gitignore`:
```
n8n-docs-sync
.cache/
```

**Step 5: Commit**

```bash
git add go.mod cmd/n8n-docs-sync/main.go .gitignore
git commit -m "feat: scaffold Go module and CLI entry point"
```

---

### Task 2: Fetch Stage

**Files:**
- Create: `internal/fetch/fetch.go`
- Create: `internal/fetch/fetch_test.go`

**Step 1: Write the failing test**

Create `internal/fetch/fetch_test.go`:

```go
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
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/fetch/ -v`
Expected: FAIL — package/functions don't exist yet

**Step 3: Write implementation**

Create `internal/fetch/fetch.go`:

```go
package fetch

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

const DefaultURL = "https://docs.n8n.io/llms-full.txt"

type Result struct {
	Body    string
	Changed bool
}

func Fetch(url string, cacheDir string) (*Result, error) {
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return nil, fmt.Errorf("creating cache dir: %w", err)
	}

	etagPath := filepath.Join(cacheDir, "etag")

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	// Send If-None-Match if we have a cached ETag
	if cachedETag, err := os.ReadFile(etagPath); err == nil {
		req.Header.Set("If-None-Match", string(cachedETag))
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetching %s: %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotModified {
		return &Result{Changed: false}, nil
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading body: %w", err)
	}

	// Cache the ETag for next run
	if etag := resp.Header.Get("ETag"); etag != "" {
		os.WriteFile(etagPath, []byte(etag), 0644)
	}

	return &Result{Body: string(body), Changed: true}, nil
}
```

**Step 4: Run tests to verify they pass**

Run: `go test ./internal/fetch/ -v`
Expected: PASS (both tests)

**Step 5: Commit**

```bash
git add internal/fetch/
git commit -m "feat(fetch): add HTTP fetcher with ETag caching"
```

---

### Task 3: Parser Stage

**Files:**
- Create: `internal/parser/parser.go`
- Create: `internal/parser/parser_test.go`

**Step 1: Write the failing test**

Create `internal/parser/parser_test.go`:

```go
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

	expected := "# Test node\n\nLine 1.\n\nLine 2.\n\n## Sub-heading\n\nMore content.\n"
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
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/parser/ -v`
Expected: FAIL — package doesn't exist

**Step 3: Write implementation**

Create `internal/parser/parser.go`:

```go
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
```

**Step 4: Run tests to verify they pass**

Run: `go test ./internal/parser/ -v`
Expected: PASS (all tests)

**Step 5: Commit**

```bash
git add internal/parser/
git commit -m "feat(parser): add section parser with type classification"
```

---

### Task 4: Filter Stage

**Files:**
- Create: `internal/filter/filter.go`
- Create: `internal/filter/filter_test.go`

**Step 1: Write the failing test**

Create `internal/filter/filter_test.go`:

```go
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
	content := `# Some AI node

Useful AI docs.

## Parameter resolution in sub-nodes

Sub-nodes behave differently to other nodes when processing multiple items using an expression.

Most nodes, including root nodes, take any number of items as input, process these items, and output the results. You can use expressions to refer to input items, and the node resolves the expression for each item in turn. For example, given an input of five ` + "`name`" + ` values, the expression ` + "`{{ $json.name }}`" + ` resolves to each name in turn.

In sub-nodes, the expression always resolves to the first item. For example, given an input of five ` + "`name`" + ` values, the expression ` + "`{{ $json.name }}`" + ` always resolves to the first name.

## Next section

Keep this.
`

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
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/filter/ -v`
Expected: FAIL — package doesn't exist

**Step 3: Write implementation**

Create `internal/filter/filter.go`:

```go
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
```

**Step 4: Run tests to verify they pass**

Run: `go test ./internal/filter/ -v`
Expected: PASS (all 5 tests)

**Step 5: Commit**

```bash
git add internal/filter/
git commit -m "feat(filter): remove release notes and strip boilerplate patterns"
```

---

### Task 5: Condense Stage — Prompts & Cache

**Files:**
- Create: `internal/condense/prompts.go`
- Create: `internal/condense/cache.go`
- Create: `internal/condense/cache_test.go`

**Step 1: Write prompts.go (no test needed — pure data)**

Create `internal/condense/prompts.go`:

```go
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
```

**Step 2: Write the failing cache test**

Create `internal/condense/cache_test.go`:

```go
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
```

**Step 3: Run test to verify it fails**

Run: `go test ./internal/condense/ -v`
Expected: FAIL — package doesn't exist

**Step 4: Write cache implementation**

Create `internal/condense/cache.go`:

```go
package condense

import (
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
)

type Cache struct {
	dir string
}

func NewCache(dir string) *Cache {
	os.MkdirAll(dir, 0755)
	return &Cache{dir: dir}
}

func (c *Cache) Get(content string) (string, bool) {
	path := c.pathFor(content)
	data, err := os.ReadFile(path)
	if err != nil {
		return "", false
	}
	return string(data), true
}

func (c *Cache) Put(content, condensed string) {
	path := c.pathFor(content)
	os.WriteFile(path, []byte(condensed), 0644)
}

func (c *Cache) pathFor(content string) string {
	hash := sha256.Sum256([]byte(content))
	return filepath.Join(c.dir, fmt.Sprintf("%x.txt", hash))
}
```

**Step 5: Run tests to verify they pass**

Run: `go test ./internal/condense/ -v`
Expected: PASS (both tests)

**Step 6: Commit**

```bash
git add internal/condense/
git commit -m "feat(condense): add section prompts and content-hash cache"
```

---

### Task 6: Condense Stage — Worker Pool

**Files:**
- Create: `internal/condense/condense.go`
- Create: `internal/condense/condense_test.go`

**Step 1: Write the failing test**

Create `internal/condense/condense_test.go`:

```go
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

	// Mock runner that uppercases content
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
		if s.Content[:10] != "CONDENSED:" {
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
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/condense/ -v -run TestCondense`
Expected: FAIL — `Condense` function doesn't exist

**Step 3: Write implementation**

Create `internal/condense/condense.go`:

```go
package condense

import (
	"fmt"
	"sync"

	"github.com/josh-ninjatrader/awesome-n8n-docs/internal/parser"
)

// Runner executes an LLM condensation. Takes section content and prompt, returns condensed output.
type Runner func(content, prompt string) (string, error)

func Condense(sections []parser.Section, cacheDir string, concurrency int, run Runner) ([]parser.Section, error) {
	cache := NewCache(cacheDir)
	result := make([]parser.Section, len(sections))
	errs := make([]error, len(sections))

	sem := make(chan struct{}, concurrency)
	var wg sync.WaitGroup

	for i, s := range sections {
		wg.Add(1)
		go func(idx int, sect parser.Section) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			condensed, err := condenseOne(sect, cache, run)
			if err != nil {
				errs[idx] = fmt.Errorf("condensing %q: %w", sect.Title, err)
				return
			}
			sect.Content = condensed
			result[idx] = sect
		}(i, s)
	}

	wg.Wait()

	// Check for errors
	for _, err := range errs {
		if err != nil {
			return nil, err
		}
	}

	return result, nil
}

func condenseOne(s parser.Section, cache *Cache, run Runner) (string, error) {
	// Check cache
	if cached, ok := cache.Get(s.Content); ok {
		return cached, nil
	}

	prompt := PromptFor(s.Type)
	condensed, err := run(s.Content, prompt)
	if err != nil {
		return "", err
	}

	// Quality guardrail: if output is >80% of input, retry with stricter prompt
	if len(condensed) > 0 && float64(len(condensed))/float64(len(s.Content)) > 0.8 {
		stricter := prompt + " Be much more aggressive in condensing. Remove all unnecessary words."
		condensed, err = run(s.Content, stricter)
		if err != nil {
			return "", err
		}
	}

	// Quality guardrail: if output is <10% of input, return with warning (don't cache)
	if len(condensed) > 0 && float64(len(condensed))/float64(len(s.Content)) < 0.1 {
		fmt.Printf("WARNING: section %q condensed to <10%% of original (%d -> %d bytes)\n",
			s.Title, len(s.Content), len(condensed))
	}

	cache.Put(s.Content, condensed)
	return condensed, nil
}
```

**Step 4: Run tests to verify they pass**

Run: `go test ./internal/condense/ -v`
Expected: PASS (all tests including cache tests)

**Step 5: Commit**

```bash
git add internal/condense/condense.go internal/condense/condense_test.go
git commit -m "feat(condense): add worker pool with quality guardrails"
```

---

### Task 7: Claude Runner (the `claude -p` integration)

**Files:**
- Create: `internal/condense/claude.go`
- Create: `internal/condense/claude_test.go`

**Step 1: Write the failing test**

Create `internal/condense/claude_test.go`:

```go
package condense

import (
	"os/exec"
	"testing"
)

func TestClaudeRunner_Integration(t *testing.T) {
	// Skip if claude CLI not available
	if _, err := exec.LookPath("claude"); err != nil {
		t.Skip("claude CLI not found, skipping integration test")
	}

	result, err := ClaudeRunner("Say only the word 'hello' with no other text.", "Respond with exactly one word.")
	if err != nil {
		t.Fatalf("claude runner error: %v", err)
	}
	if len(result) == 0 {
		t.Error("expected non-empty result from claude")
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/condense/ -v -run TestClaudeRunner`
Expected: FAIL — `ClaudeRunner` doesn't exist

**Step 3: Write implementation**

Create `internal/condense/claude.go`:

```go
package condense

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

// ClaudeRunner shells out to `claude -p` for LLM condensation.
func ClaudeRunner(content, prompt string) (string, error) {
	cmd := exec.Command("claude", "-p", prompt)
	cmd.Stdin = strings.NewReader(content)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("claude -p failed: %w\nstderr: %s", err, stderr.String())
	}

	return strings.TrimSpace(stdout.String()), nil
}
```

**Step 4: Run tests to verify they pass**

Run: `go test ./internal/condense/ -v -run TestClaudeRunner`
Expected: PASS (or SKIP if claude CLI not installed)

**Step 5: Commit**

```bash
git add internal/condense/claude.go internal/condense/claude_test.go
git commit -m "feat(condense): add claude -p runner integration"
```

---

### Task 8: Formatter Stage

**Files:**
- Create: `internal/formatter/formatter.go`
- Create: `internal/formatter/formatter_test.go`

**Step 1: Write the failing test**

Create `internal/formatter/formatter_test.go`:

```go
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
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/formatter/ -v`
Expected: FAIL — package doesn't exist

**Step 3: Write implementation**

Create `internal/formatter/formatter.go`:

```go
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
```

**Step 4: Run tests to verify they pass**

Run: `go test ./internal/formatter/ -v`
Expected: PASS (all tests)

**Step 5: Commit**

```bash
git add internal/formatter/
git commit -m "feat(formatter): add section delimiter formatting with intro summary"
```

---

### Task 9: Wire Pipeline in main.go

**Files:**
- Modify: `cmd/n8n-docs-sync/main.go`

**Step 1: Write the full pipeline orchestration**

Replace `cmd/n8n-docs-sync/main.go` with:

```go
package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/josh-ninjatrader/awesome-n8n-docs/internal/condense"
	"github.com/josh-ninjatrader/awesome-n8n-docs/internal/fetch"
	"github.com/josh-ninjatrader/awesome-n8n-docs/internal/filter"
	"github.com/josh-ninjatrader/awesome-n8n-docs/internal/formatter"
	"github.com/josh-ninjatrader/awesome-n8n-docs/internal/parser"
)

func main() {
	dryRun := flag.Bool("dry-run", false, "Show what would change without writing")
	noCondense := flag.Bool("no-condense", false, "Skip LLM condensation")
	output := flag.String("output", "n8n-docs-llms.txt", "Output file path")
	concurrency := flag.Int("concurrency", 5, "Number of concurrent claude -p workers")
	cacheDir := flag.String("cache-dir", ".cache", "Directory for condensation cache")
	sourceURL := flag.String("source-url", fetch.DefaultURL, "URL to fetch llms-full.txt from")
	flag.Parse()

	// Stage 1: Fetch
	fmt.Println("Fetching n8n documentation...")
	result, err := fetch.Fetch(*sourceURL, *cacheDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "fetch failed: %v\n", err)
		os.Exit(1)
	}
	if !result.Changed {
		fmt.Println("Documentation unchanged (ETag match). Nothing to do.")
		os.Exit(0)
	}
	fmt.Printf("Downloaded %d bytes.\n", len(result.Body))

	// Stage 2: Parse
	fmt.Println("Parsing sections...")
	sections := parser.Parse(result.Body)
	fmt.Printf("Parsed %d sections.\n", len(sections))

	// Stage 3: Filter
	fmt.Println("Filtering...")
	filtered := filter.Filter(sections)
	fmt.Printf("Filtered to %d sections (removed %d).\n", len(filtered), len(sections)-len(filtered))

	// Stage 4: Condense
	var condensed []parser.Section
	if *noCondense {
		fmt.Println("Skipping condensation (--no-condense).")
		condensed = filtered
	} else {
		fmt.Printf("Condensing %d sections with %d workers...\n", len(filtered), *concurrency)
		condensed, err = condense.Condense(filtered, *cacheDir, *concurrency, condense.ClaudeRunner)
		if err != nil {
			fmt.Fprintf(os.Stderr, "condense failed: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Condensation complete.")
	}

	// Stage 5: Format
	fmt.Println("Formatting output...")
	formatted := formatter.Format(condensed)

	// Stage 6: Output
	if *dryRun {
		fmt.Printf("Dry run: would write %d bytes to %s\n", len(formatted), *output)
		os.Exit(0)
	}

	if err := os.WriteFile(*output, []byte(formatted), 0644); err != nil {
		fmt.Fprintf(os.Stderr, "write failed: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Written %d bytes to %s\n", len(formatted), *output)
}
```

**Step 2: Build and verify**

Run: `go build -o n8n-docs-sync ./cmd/n8n-docs-sync`
Expected: Compiles successfully

**Step 3: Test with --no-condense and --dry-run**

Run: `./n8n-docs-sync --no-condense --dry-run`
Expected: Downloads docs, parses, filters, and reports what would be written (no file created)

**Step 4: Commit**

```bash
git add cmd/n8n-docs-sync/main.go
git commit -m "feat: wire complete pipeline in main.go"
```

---

### Task 10: GitHub Actions Workflow

**Files:**
- Create: `.github/workflows/sync-n8n-docs.yml`

**Step 1: Write the workflow**

Create `.github/workflows/sync-n8n-docs.yml`:

```yaml
name: Sync n8n Documentation

on:
  schedule:
    - cron: '0 8 * * 1'  # Every Monday at 08:00 UTC
  workflow_dispatch:

permissions:
  contents: write
  pull-requests: write

jobs:
  sync:
    runs-on: ubuntu-latest
    steps:
      - name: Checkout
        uses: actions/checkout@v4

      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version-file: go.mod

      - name: Install Claude CLI
        run: |
          npm install -g @anthropic-ai/claude-code

      - name: Restore cache
        uses: actions/cache@v4
        with:
          path: .cache
          key: n8n-docs-condense-${{ github.sha }}
          restore-keys: |
            n8n-docs-condense-

      - name: Build
        run: go build -o n8n-docs-sync ./cmd/n8n-docs-sync

      - name: Run sync
        env:
          ANTHROPIC_API_KEY: ${{ secrets.ANTHROPIC_API_KEY }}
        run: ./n8n-docs-sync

      - name: Check for changes
        id: changes
        run: |
          if git diff --quiet n8n-docs-llms.txt; then
            echo "changed=false" >> "$GITHUB_OUTPUT"
          else
            echo "changed=true" >> "$GITHUB_OUTPUT"
          fi

      - name: Create Pull Request
        if: steps.changes.outputs.changed == 'true'
        env:
          GH_TOKEN: ${{ secrets.GITHUB_TOKEN }}
        run: |
          BRANCH="docs/sync-n8n-docs-$(date +%Y-%m-%d)"
          git checkout -b "$BRANCH"
          git add n8n-docs-llms.txt
          git commit -m "docs: update n8n-docs-llms.txt from latest official docs"
          git push -u origin "$BRANCH"

          # Check if PR already exists for a docs/sync branch
          EXISTING=$(gh pr list --head "docs/sync-n8n-docs" --state open --json number --jq '.[0].number // empty')
          if [ -n "$EXISTING" ]; then
            echo "Updating existing PR #$EXISTING"
          else
            gh pr create \
              --title "docs: update n8n-docs-llms.txt $(date +%Y-%m-%d)" \
              --body "$(cat <<'EOF'
          ## Summary

          Automated weekly sync of n8n documentation from the official `llms-full.txt`.

          - Source: https://docs.n8n.io/llms-full.txt
          - Pipeline: fetch -> parse -> filter -> condense (claude -p) -> format

          Please review the diff to verify condensation quality.
          EOF
              )"
          fi
```

**Step 2: Verify YAML is valid**

Run: `python3 -c "import yaml; yaml.safe_load(open('.github/workflows/sync-n8n-docs.yml'))"`
Expected: No errors

**Step 3: Commit**

```bash
git add .github/workflows/sync-n8n-docs.yml
git commit -m "ci: add weekly n8n docs sync GitHub Actions workflow"
```

---

### Task 11: Run All Tests

**Step 1: Run full test suite**

Run: `go test ./... -v`
Expected: All tests PASS

**Step 2: Run build**

Run: `go build -o n8n-docs-sync ./cmd/n8n-docs-sync`
Expected: Clean compile

**Step 3: End-to-end smoke test (no-condense mode)**

Run: `./n8n-docs-sync --no-condense --output test-output.txt && wc -l test-output.txt && rm test-output.txt`
Expected: Downloads, processes, and writes output file. Line count should be in the tens of thousands.

---

### Task 12: Update CLAUDE.md

**Files:**
- Modify: `CLAUDE.md`

**Step 1: Add build/test commands and Go project info to CLAUDE.md**

Add after the "Working With This Repo" section:

```markdown
## Go Tool: n8n-docs-sync

The `cmd/n8n-docs-sync/` directory contains a Go CLI tool that automates updating `n8n-docs-llms.txt`.

### Build & Test

```bash
# Build
go build -o n8n-docs-sync ./cmd/n8n-docs-sync

# Run all tests
go test ./...

# Run a single package's tests
go test ./internal/parser/ -v

# Run a specific test
go test ./internal/filter/ -v -run TestFilter_StripsTemplatesBlock
```

### Usage

```bash
# Full run (requires claude CLI)
./n8n-docs-sync

# Without LLM condensation (for testing)
./n8n-docs-sync --no-condense

# Dry run
./n8n-docs-sync --dry-run
```

### Package Structure

- `internal/fetch/` — HTTP download with ETag caching
- `internal/parser/` — Splits docs into typed sections
- `internal/filter/` — Removes release notes and boilerplate
- `internal/condense/` — Worker pool for `claude -p` with content-hash cache
- `internal/formatter/` — Adds section delimiters and intro summary
```

**Step 2: Commit**

```bash
git add CLAUDE.md
git commit -m "docs: add Go tool build/test commands to CLAUDE.md"
```
