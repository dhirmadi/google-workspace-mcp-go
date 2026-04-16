---
description: Spec-first TDD — implement a roadmap item from docs/roadmap/
argument-hint: "[path|ID] e.g. AUTH-01 or docs/roadmap/AUTH-01-pkce-public-client.md"
model: sonnet
---

# /implement-roadmapitem

You are implementing a **roadmap item** for this Go MCP server. Follow **spec-first TDD**: acceptance criteria drive tests; tests drive implementation.

## Input

The user provides either:

- Full path: `docs/roadmap/XXX-NN-itemname.md`, or
- ID only (e.g. `AUTH-01`) — resolve to the matching file under `docs/roadmap/`.

Refuse if the file is under `docs/roadmap/archive/` (closed items). Offer `/review-roadmapitem` or archive workflow instead.

## Before coding

1. Read the roadmap item end-to-end plus **`docs/templates/roadmap_item.md`** if the spec looks thin.
2. Read **`CLAUDE.md`**, **`docs/code-patterns.md`**, and any paths cited in the item (e.g. `docs/auth-and-scopes.md`).
3. Optionally load **`.claude/skills/roadmap-spec-tdd/SKILL.md`** and **`.claude/skills/google-workspace-mcp-go/SKILL.md`** for the full loop.

## TDD loop (non-negotiable)

For each acceptance criterion (Gherkin `Given/When/Then` or explicit checkbox):

1. **Red**: Add a **failing** test first (`*_test.go`, table-driven, `testing` only — no external assertion libs). The test name should reference the AC (e.g. `TestAUTH01_PKCE_Exchange_SendsVerifier`).
2. **Green**: Write the **minimal** production code to pass that test only.
3. **Refactor**: Clean up; keep handlers small; extract helpers; preserve existing patterns in the same package.

Order ACs by **dependency** (config/auth before tools that call APIs). Use **fakes/mocks** at Google client boundaries where integration is impractical; prefer **real** `httptest` or small integration tests with `-tags=integration` when env vars are already documented.

## Implementation standards (this repo)

- **Module**: `github.com/evert/google-workspace-mcp-go` imports only; `goimports` local prefix.
- **Tools**: `mcp.AddTool` + full **`ToolAnnotations`**; SEP-986 `snake_case` names; register in `internal/tools/<svc>/`, wire **`internal/registry`**, **`configs/tool_tiers.yaml`** when applicable.
- **Output**: Data tools → **dual output** (text + typed struct); action tools → text-only + `nil` typed output per **`docs/code-patterns.md`**.
- **Errors**: Route Google API failures through **`internal/middleware`** patterns — agent-actionable messages, no secret leakage.
- **Security**: No new secrets in code; scope changes require **`internal/auth/scopes.go`** + doc updates.

## Verification (run and fix until clean)

```bash
golangci-lint run ./...
go test -race ./...
# when applicable:
GOOGLE_OAUTH_CLIENT_ID=test GOOGLE_OAUTH_CLIENT_SECRET=test go test -race -tags=integration ./...
```

## After implementation

1. Update the roadmap item: check off completed acceptance signals; set **Status** toward **Done** when all in-scope ACs are met (do not set **Closed** without `/archive-roadmap-item` flow).
2. Update **`docs/roadmap/README.md`** only if horizon/status text changed.
3. Add a line to **`CHANGELOG.md`** → `[Unreleased]` → **Added** or **Changed** with the roadmap **ID**.
4. Summarize: files touched, tests added, any follow-ups for `/review-roadmapitem`.
