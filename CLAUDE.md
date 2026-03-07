# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Repository Overview

This is a fork of [Synaptiv-AI/awesome-n8n](https://github.com/Synaptiv-AI/awesome-n8n) — a community-driven knowledge base for n8n workflow automation resources. The primary asset is `n8n-docs-llms.txt`, a comprehensive LLM-optimized copy of the official n8n documentation (~42k lines, ~1.4MB).

**Remotes:**
- `origin`: `josh-ninjatrader/awesome-n8n-docs` (fork)
- `upstream`: `Synaptiv-AI/awesome-n8n` (source)

## Repository Structure

```
n8n-docs-llms.txt   # Complete n8n docs formatted for LLM consumption (RAG, custom GPTs, etc.)
README.md           # Project overview and contribution guidelines
.gitignore          # Excludes /docs and .DS_Store
cmd/
  n8n-docs-sync/
    main.go              # CLI entry point (flag parsing, pipeline orchestration)
internal/
  fetch/                 # HTTP download with ETag caching
  parser/                # Splits llms-full.txt into typed sections
  filter/                # Removes release notes and boilerplate
  condense/              # Worker pool for claude -p with content-hash cache
  formatter/             # Adds section delimiters and intro summary
docs/plans/              # Design and implementation plans
n8n-docs-llms.txt        # Output: LLM-optimized n8n docs
README.md
go.mod
```

## Key File: n8n-docs-llms.txt

Contains the full official n8n documentation, condensed and optimized for LLM consumption. Organized into sections separated by `----------------------------------------` dividers, covering:
- Core concepts and best practices
- Node/integration documentation (Google Sheets, Slack, Postgres, etc.)
- Example workflow definitions (JSON-like structured data)
- n8n API endpoint documentation and schemas

Source: `https://docs.n8n.io/llms-full.txt` (~4MB), processed down to ~1.4MB.

## Go Tool: n8n-docs-sync

Pipeline that downloads the official n8n docs and processes them: fetch -> parse -> filter -> condense (via `claude -p`) -> format -> output.

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

# Custom output and concurrency
./n8n-docs-sync --output custom.txt --concurrency 3
```

### Pipeline Stages

1. **Fetch** — Downloads `llms-full.txt` with ETag caching (skips if unchanged)
2. **Parse** — Splits into ~1,350 sections, classifies as node-doc/credential/release-note/api-ref/general
3. **Filter** — Removes release notes, strips 4 boilerplate patterns (templates, parameter resolution, unsupported ops, human-in-the-loop)
4. **Condense** — Concurrent `claude -p --model haiku` calls with content-hash cache and quality guardrails
5. **Format** — Adds `----------------------------------------` delimiters and intro summary
6. **Output** — Writes `n8n-docs-llms.txt`

## Working With This Repo

- Sync with upstream: `git fetch upstream && git merge upstream/main`
- The `/docs` directory is gitignored (except `docs/plans/`)
- The n8n docs MCP tool (`mcp__n8n-docs__search_n8n_knowledge_sources`) can be used to cross-reference or validate content against current n8n documentation
