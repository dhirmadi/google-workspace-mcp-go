# Epic: [SHORT TITLE]

| Field | Value |
|-------|-------|
| **Status** | Proposed \| Not Started \| In Progress \| Done \| Closed |
| **Horizon** | Now \| Next \| Then \| Policy \| Deferred \| Watchlist |
| **Priority** | P0 \| P1 \| P2 (optional; align with `docs/roadmap/README.md`) |
| **Upstream** | Links to upstream release/PR when parity-related (or “N/A”) |
| **Dependencies** | Roadmap file links (e.g. `./POL-01-tool-count-policy.md`) or `docs/` anchors only — avoid vague “team X” |

**Closed** is reserved for items in **`docs/roadmap/archive/`** after **`/archive-roadmap-item`**.

## Problem

[User-visible or operator-visible gap; one tight paragraph.]

## Outcome

[Measurable theme; what “done” means for users/agents.]

## Scope

### In scope

- [Bullet list]

### Out of scope

- [Explicit non-goals]

## Dependencies

- [Technical or sequencing deps; use roadmap IDs or doc paths, e.g. `docs/auth-and-scopes.md`]

## Acceptance criteria (Gherkin)

Every AC must be **testable** (implementable as `_test.go`, integration test, or verifiable doc/checklist). No vague language (“fast”, “robust”, “handles errors well”).

### AC1: [Name]

```gherkin
Given [precondition]
When [action]
Then [observable outcome]
```

### AC2: [Name]

```gherkin
Given [precondition]
When [action]
Then [observable outcome]
```

[Add ACs for error paths, destructive vs read-only tools, OAuth/scopes boundaries, and concurrency if relevant.]

## Implementation notes

Ground every claim in the **current** repo (file paths, symbols). This server has **no SQL migrations**; persistence is OAuth tokens and runtime config.

### MCP tool surface

- Tools to add/change/remove (exact `snake_case` names; SEP-986).
- `ReadOnlyHint` / `DestructiveHint` / `IdempotentHint` / `OpenWorldHint` expectations.
- Dual-output (`structuredContent`) vs text-only per `docs/code-patterns.md`.

### Packages and registration

- `internal/tools/[service]/` handlers and `Register` patterns.
- `internal/registry/registry.go` — tier/service filters, name validation.
- `internal/services/factory.go` — client construction and caching.
- `configs/tool_tiers.yaml` — tier membership if applicable.

### Auth and transport

- `internal/auth/*`, `internal/config/*` — env vars, callback URLs, scopes in `internal/auth/scopes.go`.
- `cmd/server/main.go` — transport, middleware wiring.

### Verification

- Commands: `golangci-lint run ./...`, `go test -race ./...`, and `-tags=integration` when OAuth/API mocks apply.

## Component inventory

| Path | Change |
|------|--------|
| `path/to/file.go` | New \| Modified \| Removed |

## Existing infrastructure to reuse

- Response builder, middleware (`internal/middleware/*`), shared helpers (`internal/pkg/*`), comment tools pattern (`internal/tools/comments/`), etc. — list **concrete** symbols or files to extend.

## Risks

| Risk | Impact | Mitigation |
|------|--------|------------|
| [Realistic, codebase-informed] | [Who/what breaks] | [Concrete] |
