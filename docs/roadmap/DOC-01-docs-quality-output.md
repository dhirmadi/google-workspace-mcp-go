# Epic: Docs — quality of markdown / tables / batch updates

| Field | Value |
|-------|--------|
| **Status** | Proposed |
| **Horizon** | Then |
| **Priority** | P2 |
| **Upstream** | Table batch operations [#656](https://github.com/taylorwilsdon/google_workspace_mcp/pull/656); `create_table_with_data` fix [#653](https://github.com/taylorwilsdon/google_workspace_mcp/pull/653); smart chips / paragraph rendering [#649](https://github.com/taylorwilsdon/google_workspace_mcp/pull/649); sophisticated Docs creation [#628](https://github.com/taylorwilsdon/google_workspace_mcp/pull/628) |
| **Dependencies** | `./POL-01-tool-count-policy.md`; `internal/tools/docs/`; `./SEC-01-security-hygiene-upstream.md` (large-doc memory bounds) |

## Problem

Agents judge quality by **readable structure** (markdown, tables, smart chips) when reading or round-tripping Docs — not only API success. Upstream invested heavily in fidelity and batch table operations.

## Outcome

Docs read paths produce **more faithful** structured text for agents; table create/update batch tools **fail loudly** or succeed atomically on covered cases (no silent empty tables). Documented limits prevent unbounded memory use on large documents.

## Scope

### In scope

- `internal/tools/docs/` read and write paths affecting markdown/table fidelity.
- Fixtures or integration-tagged tests for table create/populate.
- Release-note style summary for users migrating from Python MCP.

### Out of scope

- Full WYSIWYG parity with the Google Docs web UI.

## Dependencies

- `docs/code-patterns.md` for structured vs text responses.
- Google Docs API batch limits — document in tool descriptions.

## Acceptance criteria (Gherkin)

### AC1: Table create or populate — success path

```gherkin
Given a representative Docs table fixture or test document id
When create_table_with_data or equivalent batch table tool runs with valid parameters
Then the document contains the expected rows and columns
And the tool response confirms success with structured summary
```

### AC2: Table batch — failure is explicit

```gherkin
Given parameters that violate Docs API or documented size limits
When the batch table tool runs
Then the tool returns a structured error without partial silent success
And the document is not left in an undeclared inconsistent state when atomicity is promised in docs
```

### AC3: Read path markdown fidelity

```gherkin
Given a document fixture with headings lists and a simple table
When the read Doc content tool runs
Then the returned markdown preserves logical structure per golden file expectations
```

### AC4: Large document bounds

```gherkin
Given a document at the documented size limit
When the read path executes
Then peak memory stays within documented bounds or the tool errors with a clear limit message
```

### AC5: Verification gates

```gherkin
Given the implementation merges
When golangci-lint run ./... and go test -race ./... execute
Then both succeed
```

## Implementation notes

### MCP tool surface

- Exact tool names from `internal/tools/docs/docs.go`; extend parameters vs new tools per `./POL-01-tool-count-policy.md`.

### Packages and registration

- `internal/tools/docs/handlers.go`, helpers, `docs.go`
- Shared utilities only if duplication is proven — prefer small focused helpers under `internal/pkg/`.

### Auth and transport

- Scopes in `internal/auth/scopes.go` — Docs read/write already present; extend only if new methods require additional scopes.

### Verification

- Golden markdown files under `internal/tools/docs/testdata/` or integration tests with `-tags=integration`.

## Component inventory

| Path | Change |
|------|--------|
| `internal/tools/docs/handlers.go` | Modified |
| `internal/tools/docs/docs.go` | Modified |
| `internal/pkg/*` | Modified only if shared helper justified |
| `docs/tools-inventory.md` | Modified |
| `CHANGELOG.md` | Modified under [Unreleased] for user-visible behavior |

## Existing infrastructure to reuse

- `internal/pkg/response/builder.go` — dual output.
- Comment tools under `internal/tools/comments/` if table operations share list/update patterns.

## Risks

| Risk | Impact | Mitigation |
|------|--------|------------|
| Docs API batch partial failures | Corrupt user documents | Map to batch response semantics; add tests for failure injection |
| Markdown export diverges from Google rendering | Agent confusion | Version golden files when API output shifts |
