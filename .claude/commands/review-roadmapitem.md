---
description: Critical roadmap implementation review — security, quality, docs, perf
argument-hint: "[path|ID] optional commit or package scope"
model: opus
---

# /review-roadmapitem

You are a **senior reviewer** for this Go MCP Google Workspace server. Review the **implementation that satisfies** the referenced roadmap item (diff against `main`, or current branch, plus the spec file). Be **direct and critical**; prioritize **Blocker** findings.

## Input

Roadmap path `docs/roadmap/XXX-NN-itemname.md` or **ID**. If the item was archived, use `docs/roadmap/archive/XXX-NN-itemname.md` and note that scope is historical.

Optionally the user names a PR, commit range, or package scope — respect that boundary.

## Review dimensions

### 1. Security (highest priority)

- **OAuth / tokens**: No secrets in logs, errors, or tool output; token file permissions; state/HMAC usage in auth flow.
- **Scopes**: Least privilege; changes aligned with **`docs/auth-and-scopes.md`** and **`internal/auth/scopes.go`**.
- **Destructive tools**: `DestructiveHint` accurate; no accidental data exfiltration via dual output or oversized payloads.
- **Input validation**: JSON schema tags on inputs; reject ambiguous IDs; path/safe handling for Drive/Docs paths where relevant.
- **Dependencies**: No risky upgrades without justification.

### 2. Quality & correctness

- Error handling: wrapped errors, context propagation, **`middleware.HandleGoogleAPIError`** (or successor) consistency.
- **Registry / tiers**: Tool name uniqueness; tier YAML matches registration; no dead tools.
- **Concurrency**: Factory/cache behavior; `context.Context` on API calls; no goroutine leaks.
- **Tests**: AC coverage — gaps between Gherkin/checklist and `_test.go`; flaky patterns; missing race-sensitive tests.

### 3. Code documentation

- **Exported** symbols that are non-obvious need concise GoDoc.
- Tool **descriptions** and parameter docs (MCP-facing) must match behavior.
- Repo docs: if behavior changed, **`docs/*.md`** / **`README.md`** updates or explicit “doc debt” items.

### 4. Performance

- Unnecessary **serial** Google calls where batch APIs exist.
- Large allocations in hot paths; streaming/progress where SDK supports it for batch tools.
- **Retry storms**: `internal/middleware/retry.go` interaction; 429 handling.
- Context deadlines for long operations.

Load **`.claude/skills/mcp-security-review/SKILL.md`** for a condensed checklist to cross-check.

## Output format

1. **Verdict**: Ship / Ship with fixes / Do not ship (with one-line why).
2. **Findings table**: `Severity` (Blocker / Should-fix / Nit) | `Area` | `File:line` or path | `Issue` | `Suggested fix`.
3. **Spec drift**: List roadmap ACs **not** clearly satisfied in code/tests.
4. **Claude follow-ups**: Ordered list of concrete edits (imperative bullets) the implementer should apply; reference patterns from **`docs/code-patterns.md`**.

Do **not** rewrite the whole codebase in one pass — scope feedback to the roadmap item’s surface area unless Blockers require broader fixes.
