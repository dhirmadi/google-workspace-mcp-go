# Deferred: Enterprise — domain-wide delegation (DWD) / service accounts

| Field | Value |
|-------|--------|
| **Status** | Deferred |
| **Horizon** | Deferred |
| **Priority** | P3 |
| **Upstream** | [v1.18.0 — DWD / service account mode](https://github.com/taylorwilsdon/google_workspace_mcp/releases/tag/v1.18.0), [#665](https://github.com/taylorwilsdon/google_workspace_mcp/pull/665) — upstream **warns against casual use** |
| **Dependencies** | Conflicts with interactive OAuth assumptions documented in `./AUTH-01-pkce-public-client.md` and `docs/auth-and-scopes.md`; requires enterprise legal/security path |

## Problem

Domain-wide delegation is powerful and **dangerous** if misconfigured. This project emphasizes production safety. DWD also **conflicts** with interactive OAuth and single-user assumptions unless very carefully scoped.

## Outcome

While deferred: no implementation. If un-deferred: documented, **off-by-default** capability with hard startup validation, impersonation audit story, and explicit incompatibility matrix vs OAuth 2.0 / OAuth 2.1 modes.

## Scope

### In scope (only when un-deferred)

- Threat model, admin controls, key storage, rotation, audit logging.
- Startup validation matrix vs `MCP_ENABLE_OAUTH21` and legacy OAuth.

### Out of scope (until gating criteria met)

- Any DWD code paths, flags, or tools in the default build.

## Dependencies

- Enterprise customer demand and administrator-owned Google Workspace controls.
- Legal / security review sign-off.

## Acceptance criteria (Gherkin)

### AC1: While deferred — no DWD surface

```gherkin
Given the epic status is Deferred
When a reviewer searches the codebase for domain-wide delegation or service-account impersonation flags
Then no supported operator path enables DWD without an explicit future epic and ADR
```

### AC2: When un-deferred — gating documented first

```gherkin
Given stakeholders approve un-deferring ENT-01
When planning completes
Then docs include gating criteria met threat model admin controls and incompatibility matrix
And a new implementation epic references this file and supersedes Deferred status
```

### AC3: When un-deferred — startup validation

```gherkin
Given DWD mode is implemented behind an explicit env flag
When forbidden combinations are configured
Then the server fails fast with an actionable error before registering tools
```

## Implementation notes

### MCP tool surface

- N/A while deferred; future work must annotate high-risk tools per `docs/security.md`.

### Packages and registration

- Would touch `internal/auth/*`, `internal/config/*`, `cmd/server/main.go` — **no changes** until un-deferred.

### Verification

- N/A while deferred.

## Component inventory

| Path | Change |
|------|--------|
| — | No changes while Deferred |

## Existing infrastructure to reuse

- N/A until un-deferred; future reuse of `internal/config/config.go` validation patterns from `./AUTH-01-pkce-public-client.md`.

## Risks

| Risk | Impact | Mitigation |
|------|--------|------------|
| Casual DWD enablement | Workspace-wide data exfiltration | Keep deferred; require ADR + security review before code |
| Pressure to “match upstream” quickly | Shipping foot-guns | README positions ENT-01 as explicit opt-in only |
