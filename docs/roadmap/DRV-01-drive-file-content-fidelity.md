# Epic: Drive — file content fidelity (PDF, binaries, attachments)

| Field | Value |
|-------|--------|
| **Status** | Proposed |
| **Horizon** | Next |
| **Priority** | P1 |
| **Upstream** | [v1.19.0 — PDF text extraction & image passthrough for `get_drive_file_content`](https://github.com/taylorwilsdon/google_workspace_mcp/releases/tag/v1.19.0), [#659](https://github.com/taylorwilsdon/google_workspace_mcp/pull/659); [v1.19.0 — inaccessible directory warning](https://github.com/taylorwilsdon/google_workspace_mcp/releases/tag/v1.19.0), [#691](https://github.com/taylorwilsdon/google_workspace_mcp/pull/691) |
| **Dependencies** | `./SEC-01-security-hygiene-upstream.md` (size/path bounds); `docs/security.md`; `./POL-01-tool-count-policy.md` only if new tools are required instead of extending `get_drive_file_content` |

## Problem

Agents expect `get_drive_file_content` to behave predictably on **PDFs** and **non-text binaries** (clear errors, extracted text where feasible, or structured passthrough). Gaps vs upstream create “works in Python MCP” friction without renaming tools.

## Outcome

Documented, bounded behavior for PDF and common binary types: **text extraction** where safe, **explicit** base64/size-cap policy or export path where not, and **warnings** when Drive returns items the principal cannot read (shortcuts / shared drive edge cases) where the API exposes detectable signals without excessive false positives.

## Scope

### In scope

- `get_drive_file_content` behavior, helpers, and tool description text.
- Size limits, timeouts, and security alignment per `docs/security.md`.
- Tests with fixtures (small PDF, plain text, unsupported binary).

### Out of scope

- Replacing `export_doc_to_pdf` or other export tools — this epic is **Drive file read** fidelity only.

## Dependencies

- `internal/tools/drive/handlers.go`, `internal/tools/drive/helpers.go`, `internal/pkg/office/extract.go` as applicable.
- `docs/tools-inventory.md` if parameters or semantics change.

## Acceptance criteria (Gherkin)

### AC1: Plain text and native Google formats

```gherkin
Given a small plain-text Drive file or supported native export path
When get_drive_file_content is invoked with valid file id and user context
Then the tool returns deterministic text or structured content per docs/tools-inventory.md
And response size stays within documented limits
```

### AC2: PDF — success or explicit failure

```gherkin
Given a PDF fixture representative of text-based PDFs
When get_drive_file_content is invoked
Then either extractable text is returned
Or the tool returns a structured error with a stable reason code string documented for agents
And the tool does not silently return empty content when extraction failed
```

### AC3: Unsupported binary

```gherkin
Given a non-text binary type outside the supported set
When get_drive_file_content is invoked
Then the tool refuses or returns a documented bounded representation (e.g. base64 chunk policy)
And memory usage stays bounded for a large file per documented cap
```

### AC4: Inaccessible or partial-access items

```gherkin
Given Drive returns conditions equivalent to upstream’s “inaccessible directory” class of issues where detectable
When get_drive_file_content is invoked
Then the tool surfaces a user-visible warning or error consistent with docs/tools-inventory.md
And does not claim success with empty misleading content
```

### AC5: Verification gates

```gherkin
Given the implementation is complete
When golangci-lint run ./... and go test -race ./... execute
Then both succeed
```

## Implementation notes

### MCP tool surface

- Primary tool: `get_drive_file_content` — preserve `snake_case` name; extend parameters only if `./POL-01-tool-count-policy.md` allows (prefer optional fields on existing tool).
- Annotations: preserve `ReadOnlyHint` / size-related hints per `docs/code-patterns.md`.

### Packages and registration

- `internal/tools/drive/handlers.go`, `internal/tools/drive/helpers.go`
- `internal/pkg/office/extract.go` — reuse before adding new PDF dependency; justify new deps in `go.mod` in PR.

### Auth and transport

- No OAuth scope expansion unless Google export APIs require it — document in `docs/auth-and-scopes.md` if changed.

### Verification

- Unit tests with committed small fixtures under `internal/tools/drive/` (or testdata).

## Component inventory

| Path | Change |
|------|--------|
| `internal/tools/drive/handlers.go` | Modified |
| `internal/tools/drive/helpers.go` | Modified |
| `internal/pkg/office/extract.go` | Modified (if reuse extended) |
| `go.mod` / `go.sum` | Modified only if justified |
| `docs/tools-inventory.md` | Modified |
| `docs/security.md` | Modified if new limits documented |

## Existing infrastructure to reuse

- `internal/pkg/response/builder.go` — dual text + structured output per `docs/code-patterns.md`.
- `internal/middleware/retry.go` — 429 backoff for Drive API reads.

## Risks

| Risk | Impact | Mitigation |
|------|--------|------------|
| New PDF parser dependency increases supply-chain surface | Security review overhead | Prefer stdlib + existing `internal/pkg/office` patterns; pin minimal version |
| False-positive “inaccessible” warnings | User trust drops | Gate warnings on API signals documented in PR; add regression tests |
