# Epic: Security hygiene (upstream-informed audit)

| Field | Value |
|-------|--------|
| **Status** | Proposed |
| **Horizon** | Next |
| **Priority** | P0 |
| **Upstream** | [v1.19.0 — security / patch themes](https://github.com/taylorwilsdon/google_workspace_mcp/releases/tag/v1.19.0); credential permissions + path handling [#657](https://github.com/taylorwilsdon/google_workspace_mcp/pull/657); related hardening [#682](https://github.com/taylorwilsdon/google_workspace_mcp/pull/682) |
| **Dependencies** | `./AUTH-01-pkce-public-client.md` (optional secret) may change credential files layout — coordinate ordering or re-audit after AUTH-01 merges; `docs/security.md` |

## Problem

Upstream fixed classes of issues (credential file permissions, path traversal / sanitization, related hardening). The Go server must enforce the same **invariants** without copying Python line-by-line.

## Outcome

A completed **audit checklist** (checked items in `docs/security.md` or a single linked audit file under `docs/roadmap/`) covering credential paths, file permissions, user-influenced paths, log redaction, and input bounds on high-risk tools. Every **High** finding is fixed or tracked with owner and timeline.

## Scope

### In scope

- Read-only engineering audit with concrete file/path references.
- Checklist artifacts and fixes for issues rated High or Medium.
- Alignment with `docs/implementation-plan.md` Phase 1 credential expectations (0700/0600).

### Out of scope

- Full third-party penetration test.
- New product features disguised as security work.

## Dependencies

- `docs/security.md`, `internal/auth/credentials.go`, tools accepting paths or writing local data.
- Optional traceability: map findings to upstream PRs above.

## Acceptance criteria (Gherkin)

### AC1: Audit checklist exists and is objective

```gherkin
Given the epic is complete
When a reviewer opens docs/security.md (or the linked audit subsection)
Then a checklist exists with pass/fail rows for credential paths, permissions, path sanitization, log redaction, and high-risk tool input bounds
And each row references at least one repo path or tool name to verify
```

### AC2: Credential directory and token file permissions

```gherkin
Given a fresh credentials directory is created by the server
When token material is written to disk
Then directory and file permissions match the values documented in docs/security.md (0700/0600 as applicable)
And no regression introduces world-readable token files
```

### AC3: High-severity gaps are closed or tracked

```gherkin
Given the audit identifies a High severity issue
When the epic closes
Then the issue is fixed in code with tests or documentation
Or a GitHub issue exists with severity, mitigation, and target date
```

### AC4: Path-influenced tools

```gherkin
Given a tool accepts a user-controlled file path or document identifier that maps to local or API resources
When inputs include traversal sequences or out-of-range values
Then the server rejects or normalizes them per documented rules without leaking internal paths in tool output
```

### AC5: Verification gates

```gherkin
Given changes land on main
When CI runs golangci-lint and go test -race on ./...
Then both complete successfully
```

## Implementation notes

### MCP tool surface

- Focus on tools that accept paths, IDs with filesystem semantics, or export large payloads — inventory from `internal/tools/` (Drive, Docs, local-adjacent helpers).

### Packages and registration

- `internal/auth/credentials.go` — chmod, mkdir, symlink/hardlink considerations if any.
- `internal/config/config.go` — paths from env vars.
- `cmd/server/main.go` — logging configuration; ensure secrets not logged.

### Auth and transport

- Callback server `internal/auth/callback.go` — bind address, CSRF/state if applicable.

### Verification

- Add or extend `_test.go` for permission bits where OS allows; document macOS vs Linux differences in checklist if needed.

## Component inventory

| Path | Change |
|------|--------|
| `docs/security.md` | Modified (checklist + outcomes) |
| `internal/auth/credentials.go` | Modified (if gaps) |
| `internal/auth/callback.go` | Modified (if gaps) |
| Target tool handlers under `internal/tools/**` | Modified as findings dictate |
| `docs/roadmap/audit-YYYY-MM.md` | New (optional; only if team prefers append-only audit file) |

## Existing infrastructure to reuse

- `internal/middleware/logging.go` — redaction patterns if present.
- `internal/middleware/errors.go` — ensure error messages do not echo raw tokens.

## Risks

| Risk | Impact | Mitigation |
|------|--------|------------|
| AUTH-01 changes credential storage concurrently | Duplicate or conflicting audits | Run SEC checklist after AUTH-01 or explicitly re-run delta audit |
| Platform-specific permission tests flaky in CI | False green | Gate strict chmod tests behind build tags or document manual verification steps |
