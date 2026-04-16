---
name: mcp-security-review
description: >-
  Security and safety review for MCP tools and OAuth in this repo. Use when
  reviewing changes, before merge, or with /review-roadmapitem. Focuses on
  secrets, scopes, destructive hints, PII in output, and input validation.
---

# MCP + OAuth security review (Google Workspace Go server)

## Checklist

- [ ] **Secrets**: No `CLIENT_SECRET`, refresh tokens, or bearer tokens in logs, `TextContent`, or error strings.
- [ ] **Token store**: File permissions 0700/0600 paths respected; no world-readable credential paths introduced.
- [ ] **OAuth state**: CSRF protections unchanged or improved when touching `internal/auth`.
- [ ] **Scopes**: Every new/changed Google call justified in `internal/auth/scopes.go`; documented in `docs/auth-and-scopes.md` if user-facing.
- [ ] **Tool annotations**: `DestructiveHint` / `ReadOnlyHint` / `IdempotentHint` / `OpenWorldHint` match real behavior.
- [ ] **Data exfiltration**: Dual-output structs contain no unexpected raw PII; size limits for large payloads where APIs allow huge responses.
- [ ] **Injection**: User-controlled strings not interpolated into shell, file paths, or URLs without validation.
- [ ] **Dependencies**: `go.sum` changes reviewed; no unnecessary broad replacements.

## Output

For each failed check: **severity**, **file**, **evidence**, **remediation**. Prefer minimal fixes that preserve MCP contract stability.
