# Epic: Contacts & Chat — reliability and structural parity

| Field | Value |
|-------|--------|
| **Status** | Proposed |
| **Horizon** | Next |
| **Priority** | P1 |
| **Upstream** | Contacts: [v1.19.0 — multi-phone/email, merge modes, batch fix](https://github.com/taylorwilsdon/google_workspace_mcp/releases/tag/v1.19.0), [#688](https://github.com/taylorwilsdon/google_workspace_mcp/pull/688); Chat: [v1.19.0 — message search stability](https://github.com/taylorwilsdon/google_workspace_mcp/releases/tag/v1.19.0), [#686](https://github.com/taylorwilsdon/google_workspace_mcp/pull/686) |
| **Dependencies** | `./SEC-01-security-hygiene-upstream.md`; `docs/auth-and-scopes.md` (People API / Chat scopes); `./POL-01-tool-count-policy.md` if tool split vs parameters changes |

## Problem

Real-world Contacts data uses **multiple phones/emails** and merge semantics; upstream fixed batch and mode gaps. Chat **message search** had stability issues upstream addressed (pagination, empty results, API quirks).

## Outcome

Contacts tools accept and return **multi-value** fields consistent with Google People API semantics; batch operations behave deterministically. Chat search tools do not fail inconsistently on pagination, empty results, or known API edge cases covered upstream.

## Scope

### In scope

- `internal/tools/contacts/` — schemas, handlers, batch paths.
- `internal/tools/chat/` — search handlers and pagination.
- Tests and, if useful, a short parity table in PR description or `docs/` appendix.

### Out of scope

- New Chat products beyond the current tool surface unless justified by an explicit parity row and `./POL-01-tool-count-policy.md` decision.

## Dependencies

- Diff upstream tool contracts vs `internal/tools/contacts/`, `internal/tools/chat/`.
- Scope alignment in `internal/auth/scopes.go` and `docs/auth-and-scopes.md` if new methods require scopes.

## Acceptance criteria (Gherkin)

### AC1: Contacts multi-value fields round-trip or validate

```gherkin
Given contact fixtures or test doubles include multiple emails and phones
When create or update contacts tools run with multi-value payloads
Then persisted or returned structures preserve cardinality per People API field semantics
And invalid combinations produce stable validation errors (no partial silent drops)
```

### AC2: Contacts batch operations are deterministic

```gherkin
Given a batch contacts request with merge mode explicitly set per tool contract
When the batch operation completes
Then outcomes match documented merge semantics for each input row
And partial failures report per-item status without corrupting unrelated rows
```

### AC3: Chat search — empty and single-page results

```gherkin
Given a Chat space with no matching messages
When message search runs with valid parameters
Then the tool returns an empty result set with success semantics
And does not error solely due to zero results
```

### AC4: Chat search — pagination stability

```gherkin
Given search results span multiple pages per Google API
When the tool aggregates or exposes pagination tokens
Then repeated calls with the same cursor semantics do not skip or duplicate messages
And the tool terminates under documented max result or page limits
```

### AC5: Verification gates

```gherkin
Given the implementation is complete
When go test -race ./... runs
Then it succeeds
```

## Implementation notes

### MCP tool surface

- Inventory exact tool names from `internal/tools/contacts/contacts.go` and `internal/tools/chat/chat.go`; preserve names unless `./POL-01-tool-count-policy.md` chooses upstream-aligned splits.
- Destructive contacts operations keep `DestructiveHint` per `docs/code-patterns.md`.

### Packages and registration

- `internal/tools/contacts/handlers.go`, `helpers.go`, `contacts.go`
- `internal/tools/chat/handlers.go`, `chat.go`
- `configs/tool_tiers.yaml` — only if tier membership changes.

### Auth and transport

- No change to `cmd/server/main.go` unless new middleware is required for logging PII — prefer redaction over removal of useful errors.

### Verification

- Unit tests with httptest doubles or golden files; document any manual Chat repro steps in PR if API cannot be mocked.

## Component inventory

| Path | Change |
|------|--------|
| `internal/tools/contacts/handlers.go` | Modified |
| `internal/tools/contacts/helpers.go` | Modified |
| `internal/tools/contacts/contacts.go` | Modified (registration if needed) |
| `internal/tools/chat/handlers.go` | Modified |
| `internal/tools/chat/chat.go` | Modified (registration if needed) |
| `docs/tools-inventory.md` | Modified if contracts change |
| `docs/auth-and-scopes.md` | Modified if scopes change |

## Existing infrastructure to reuse

- `internal/middleware/errors.go` — Google API error shaping for People and Chat.
- `internal/middleware/retry.go` — 429 handling for search-heavy calls.

## Risks

| Risk | Impact | Mitigation |
|------|--------|------------|
| People API batch limits differ from assumptions | Intermittent 400s | Match Google quotas in code; document limits in tool descriptions |
| Chat API behavior changes without notice | Flaky integration | Pin test doubles; narrow client surface |
