# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

MCP server in Go 1.24 exposing 136 tools across 12 Google Workspace services (Gmail, Drive, Calendar, Docs, Sheets, Chat, Forms, Slides, Tasks, Contacts, Custom Search, Apps Script). Uses the official MCP Go SDK (`github.com/modelcontextprotocol/go-sdk v1.2.0`, spec 2025-11-25).

Module path: `github.com/evert/google-workspace-mcp-go`

## Commands

```bash
# Build
go build -o server ./cmd/server

# Run locally (stdio mode, default)
GOOGLE_OAUTH_CLIENT_ID=id GOOGLE_OAUTH_CLIENT_SECRET=secret ./server

# Run locally (HTTP mode)
GOOGLE_OAUTH_CLIENT_ID=id GOOGLE_OAUTH_CLIENT_SECRET=secret ./server --transport streamable-http

# Test
go test ./...                        # all unit tests
go test -race ./...                  # with race detector
go test ./internal/tools/gmail/...   # single package
GOOGLE_OAUTH_CLIENT_ID=test GOOGLE_OAUTH_CLIENT_SECRET=test \
  go test -tags=integration ./internal/integration/   # integration tests

# Lint
golangci-lint run

# Docker
docker build -t google-workspace-mcp .
./start.sh "CLIENT_ID" "CLIENT_SECRET"   # build + run via Docker
./start.sh --stop                         # stop container
```

## Architecture

### Entry Point & Transport

`cmd/server/main.go` — loads config, creates OAuth manager, service factory, MCP server, registers middleware and tools, then starts either `stdio` or `streamable-http` transport. Stdout is reserved for MCP stdio; all logging goes to stderr via `log/slog` JSON handler.

### Core Components

- **`internal/config/`** — Config loaded from env vars + CLI flags. `tiers.go` parses `configs/tool_tiers.yaml` for tool tier assignments (core/extended/complete).
- **`internal/auth/`** — OAuth2 flow with HMAC-SHA256-signed state for CSRF protection. `FileTokenStore` persists per-user tokens with 0700/0600 permissions. Token auto-refresh via `ReuseTokenSource` + `PersistingTokenSource`.
- **`internal/services/factory.go`** — Per-user Google API client cache. Clients use background context and outlive individual requests; each API call passes its own request-scoped context.
- **`internal/registry/registry.go`** — Registers all tools, applies tier/service/read-only filtering via SDK middleware. Validates tool names against SEP-986 pattern.
- **`internal/middleware/`** — SDK middleware hooks (`AddReceivingMiddleware`): request logging, auth enhancement (injects user email), tier filtering. `errors.go` translates Google API errors into agent-actionable messages.

### Tool Packages (`internal/tools/`)

Each service has its own package (e.g., `internal/tools/gmail/`) with:
- `<service>.go` — `Register(server, factory)` function with tool definitions and full `ToolAnnotations`
- `handlers.go` — handler functions as closures over the factory (no global state)
- `helpers.go` — shared utilities (parsing, formatting, validation)
- `handlers_complete.go` — extended/complete tier tools (optional)
- `*_test.go` — tests next to implementation

The `comments` package is shared across Docs, Sheets, and Slides (all use Drive API for comments).

### Key Design Patterns

**Dual output**: Data-returning tools (search, list, get) return BOTH text + typed struct for machine parsing. Action tools (send, create, delete) return text only.

**Tool annotations**: Every tool declares `ReadOnlyHint`, `DestructiveHint`, `IdempotentHint`, `OpenWorldHint`. Used for tier filtering, client UI hints, and retry logic.

**Agent-actionable errors**: Google API errors are translated to guidance (e.g., 401 → "call start_google_auth tool to re-authenticate", 429 → "wait 30-60 seconds before retrying").

**Input structs**: Use `json` + `jsonschema` tags on every field. `user_google_email` is always the first field. The MCP SDK validates input against jsonschema tags automatically.

## Adding a New Tool

1. Create handler in the appropriate `internal/tools/<service>/` package
2. Add tool definition with `mcp.AddTool()` in the `Register` function with complete `ToolAnnotations`
3. Add the tool name to the appropriate tier in `configs/tool_tiers.yaml`
4. Route errors through `middleware.HandleGoogleAPIError()`
5. For data-returning tools, return both text and typed struct; for action tools, return text only

## Conventions

- Conventional Commits for messages: `feat(gmail):`, `fix(calendar):`, etc.
- Error wrapping: `fmt.Errorf("doing thing: %w", err)`
- Use `any` instead of `interface{}`
- Max function length ~50 lines; extract helpers beyond that
- Run `go mod tidy` after dependency changes
- `goimports` local prefix: `github.com/evert/google-workspace-mcp-go`
- Integration tests use build tag `//go:build integration`
- Table-driven tests with stdlib `testing` package (no assertion libraries)
