# Epic: Calendar, Sheets, Gmail — upstream feature drift

| Field | Value |
|-------|--------|
| **Status** | Proposed |
| **Horizon** | Then |
| **Priority** | P1 |
| **Upstream** | Examples: `resize_sheet_dimensions` [v1.18.0 #662](https://github.com/taylorwilsdon/google_workspace_mcp/pull/662); `manage_focus_time`, `rsvp_event`, OOO, `create_calendar` + scope fixes ([releases](https://github.com/taylorwilsdon/google_workspace_mcp/releases)); Gmail `body_format` [e.g. #571](https://github.com/taylorwilsdon/google_workspace_mcp/pull/571); drafts overhaul [#631](https://github.com/taylorwilsdon/google_workspace_mcp/pull/631); filters + list headers [v1.15.1](https://github.com/taylorwilsdon/google_workspace_mcp/releases/tag/v1.15.1) |
| **Dependencies** | **`./POL-01-tool-count-policy.md` must be decided first**; `docs/auth-and-scopes.md`; `docs/tools-inventory.md`; `configs/tool_tiers.yaml` |

## Problem

Upstream has added **tools and parameters** after baseline `docs/tools-inventory.md` (136 tools). Agents migrating from the Python server expect behaviors (RSVP, focus time, sheet dimension resize, raw HTML mail bodies, richer drafts).

## Outcome

A **prioritized backlog** (P0/P1/P2) of upstream deltas with one of: (a) implement in Go, (b) document intentional omission, (c) extend existing tools via optional parameters per `./POL-01-tool-count-policy.md`. Execution splits into small PRs per service.

## Scope

### In scope

- Parity analysis tables and phased implementation PRs for Calendar, Sheets, and Gmail.
- Registry and tier updates when tools change.
- Documentation updates for intentional gaps.

### Out of scope

- Rewriting entire Gmail or Calendar surfaces in one PR.
- Drive/Docs/Chat coverage (handled by sibling epics).

## Dependencies

- `./POL-01-tool-count-policy.md` — blocks implementation choices.
- `internal/tools/calendar/`, `internal/tools/sheets/`, `internal/tools/gmail/`.
- `internal/registry/registry.go` name validation (SEP-986).

## Acceptance criteria (Gherkin)

### AC1: Policy alignment before merge-heavy work

```gherkin
Given POL-01 records a single chosen option A B or C with owner sign-off
When PAR-01 implementation PRs are opened
Then each PR states how new upstream behavior maps to that policy
```

### AC2: Backlog artifact exists

```gherkin
Given PAR-01 is in progress or complete for a slice
When a reviewer opens the linked backlog table in-repo or in a referenced PR
Then each upstream delta row includes priority P0 P1 P2 mapping to our tool or omission rationale and T-shirt effort
```

### AC3: P0 handling

```gherkin
Given a backlog row is marked P0
When the epic slice closes
Then that row is implemented in Go or explicitly deferred with reason and follow-up issue link
```

### AC4: Inventory and tiers stay truthful

```gherkin
Given tools or parameters change
When the slice merges
Then docs/tools-inventory.md and configs/tool_tiers.yaml reflect the new surface
```

### AC5: Verification gates

```gherkin
Given code changes merge
When golangci-lint run ./... and go test -race ./... execute
Then both succeed
```

## Implementation notes

### MCP tool surface

- Follow `docs/code-patterns.md` for dual output, hints (`ReadOnlyHint`, `DestructiveHint`, `IdempotentHint`, `OpenWorldHint`).
- Exact `snake_case` names validated in `internal/registry/registry.go`.

### Packages and registration

- `internal/tools/calendar/*.go`, `internal/tools/sheets/*.go`, `internal/tools/gmail/*.go`
- `internal/services/factory.go` — new Google API clients only if scopes/services expand.

### Auth and transport

- `internal/auth/scopes.go` — extend minimal scope sets; document in `docs/auth-and-scopes.md`.

### Verification

- Handler `_test.go` per changed package; integration tests where `-tags=integration` already applies.

## Component inventory

| Path | Change |
|------|--------|
| `internal/tools/calendar/*` | Modified |
| `internal/tools/sheets/*` | Modified |
| `internal/tools/gmail/*` | Modified |
| `internal/registry/registry.go` | Modified if registration rules change |
| `configs/tool_tiers.yaml` | Modified |
| `docs/tools-inventory.md` | Modified |
| `docs/auth-and-scopes.md` | Modified |
| `docs/implementation-plan.md` | Modified if Phase 4 exit criteria change |

## Existing infrastructure to reuse

- `internal/tools/comments/` pattern for shared comment-related flows where Calendar/Docs overlap.
- `internal/pkg/htmlutil` / Gmail helpers for body format handling.

## Risks

| Risk | Impact | Mitigation |
|------|--------|------------|
| POL-01 slips while engineers start PAR work | Rework from wrong tool-split strategy | README sequencing — treat POL-01 as gate |
| Scope creep on Gmail drafts | Large PRs | Cap each merge to one upstream theme |
