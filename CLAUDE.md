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
```

There is no build system, test suite, or application code. This is a documentation-only repository.

## Key File: n8n-docs-llms.txt

Contains the full official n8n documentation scraped as of June 24, 2025. Organized into sections separated by `----------------------------------------` dividers, covering:
- Core concepts and best practices
- Node/integration documentation (Google Sheets, Slack, Postgres, etc.)
- Example workflow definitions (JSON-like structured data)
- n8n API endpoint documentation and schemas

## Working With This Repo

- Sync with upstream: `git fetch upstream && git merge upstream/main`
- The `/docs` directory is gitignored — any local docs processing output goes there
- Changes to `n8n-docs-llms.txt` should preserve the existing section divider format (`----------------------------------------`)
- The n8n docs MCP tool (`mcp__n8n-docs__search_n8n_knowledge_sources`) can be used to cross-reference or validate content against current n8n documentation
