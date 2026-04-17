---
name: google-workspace-mcp-go
description: >-
  Go 1.24 MCP server for Google Workspace using modelcontextprotocol/go-sdk.
  Use when adding or changing tools, auth, registry, middleware, config, or
  tests in this repository. Enforces dual output, ToolAnnotations, SEP-986
  names, factory pattern, and agent-actionable errors.
---

# Google Workspace MCP (Go) — implementation standard

## Stack

- Go **1.24**, MCP spec **2025-11-25**, `github.com/modelcontextprotocol/go-sdk` **v1.2.0**
- Google APIs: `google.golang.org/api`; OAuth: `golang.org/x/oauth2`
- Canonical docs: `CLAUDE.md`, `docs/architecture.md`, `docs/code-patterns.md`, `docs/auth-and-scopes.md`, `docs/security.md`

## Tool handler checklist

1. **Registration** in `internal/tools/<service>/` — `Register` adds tools with full **`ToolAnnotations`**.
2. **Handler** as closure over `*services.Factory`; no global mutable state.
3. **Inputs**: struct tags `json` + `jsonschema`; `user_google_email` first when required.
4. **Outputs**: search/list/get-style → **text + typed struct**; mutations → text only, `nil` typed output (`any` return type for suppression of OutputSchema per code-patterns).
5. **Errors**: use shared middleware translation — messages must help the **agent** recover (re-auth, backoff, fix args).
6. **Tiers**: update `configs/tool_tiers.yaml` when visibility changes; validate `internal/registry` filtering.

## Commands to run before claiming done

```bash
golangci-lint run ./...
go test -race ./...
```

## Anti-patterns

- Logging or returning OAuth client **secrets** or refresh tokens.
- Tools without annotations or with misleading read-only hints.
- Placeholder imports or `<owner>` module paths — only `github.com/evert/google-workspace-mcp-go`.
