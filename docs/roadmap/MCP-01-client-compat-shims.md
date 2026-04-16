# Watchlist: MCP client compatibility shims

| Field | Value |
|-------|--------|
| **Status** | Proposed |
| **Horizon** | Watchlist |
| **Priority** | P2 |
| **Upstream** | [v1.15.1 — Claude Cowork / Desktop JSON workarounds](https://github.com/taylorwilsdon/google_workspace_mcp/releases/tag/v1.15.1); [v1.19.0 — “patch shims”](https://github.com/taylorwilsdon/google_workspace_mcp/releases/tag/v1.19.0); DXT / stdio fixes [e.g. #626](https://github.com/taylorwilsdon/google_workspace_mcp/pull/626) |
| **Dependencies** | MCP spec (`github.com/modelcontextprotocol/go-sdk`); `cmd/server/main.go` transports; optional `.github/ISSUE_TEMPLATE/` for repro fields |

## Problem

Some MCP clients send **invalid JSON** or otherwise violate the spec for tool arguments. Upstream added **temporary shims** for specific clients.

## Outcome

**No server-side shim** until: a **named client + version**, minimal repro payload, MCP spec reference, narrow scope, and documented **removal condition**. Until then this item remains watchlist-only (no scheduled engineering).

## Scope

### In scope

- Issue template or docs note for capturing client version + repro when shims are proposed.
- Future focused epic per client if justified.

### Out of scope

- Broad spec violations “for compatibility” without named client and removal plan.
- Duplicating upstream shims preemptively.

## Dependencies

- `cmd/server/main.go` — stdio vs streamable HTTP entrypoints.
- `.github/ISSUE_TEMPLATE/` — optional field for MCP client name and version.

## Acceptance criteria (Gherkin)

### AC1: Default — no undocumented shims

```gherkin
Given MCP-01 remains on the watchlist
When reviewers inspect argument parsing paths
Then no client-specific shim exists without a linked issue naming client version and removal condition
```

### AC2: Issue intake — client identity

```gherkin
Given a contributor files a client-compat bug
When they use the bug template updated for this epic
Then the issue includes MCP client name version transport stdio or HTTP and minimal JSON repro
```

### AC3: If a shim ships — documentation and removal

```gherkin
Given a shim is approved and merged
When a reader opens the shim source
Then a comment cites MCP spec section client versions covered and the removal condition
And docs mention the temporary nature under docs/ or CHANGELOG
```

## Implementation notes

### MCP tool surface

- Shims must not weaken validation for conforming clients.

### Packages and registration

- Likely `cmd/server/main.go`, `internal/middleware/*`, or SDK wiring — only when AC3 triggers.

### Verification

- Regression tests that conforming payloads are unchanged; shim covered by targeted test with fixture from repro.

## Component inventory

| Path | Change |
|------|--------|
| `.github/ISSUE_TEMPLATE/bug_report.yml` | Modified (optional fields) |
| `cmd/server/main.go` | Modified only if shim approved |
| `internal/middleware/*` | Modified only if shim approved |

## Existing infrastructure to reuse

- `internal/middleware/errors.go` — prefer correct spec errors over silent fixes when rejecting shims.

## Risks

| Risk | Impact | Mitigation |
|------|--------|------------|
| Shim sprawl | unmaintainable server | Watchlist principle — one epic per client |
| Shims hide client bugs forever | spec ecosystem degrades | removal condition in code comment + issue |
