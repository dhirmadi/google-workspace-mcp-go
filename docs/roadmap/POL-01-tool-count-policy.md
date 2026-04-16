# Decision: Tool count vs upstream parity

| Field | Value |
|-------|--------|
| **Status** | **Decision needed** |
| **Horizon** | Policy |
| **Priority** | P0 |
| **Upstream** | New tools such as `resize_sheet_dimensions`, `create_calendar`, `rsvp_event`, `manage_focus_time`, etc. ([releases](https://github.com/taylorwilsdon/google_workspace_mcp/releases)) |
| **Dependencies** | Blocks `./PAR-01-calendar-sheets-gmail-parity.md` and constrains `./DRV-01-drive-file-content-fidelity.md`, `./CON-01-contacts-chat-reliability.md`, `./DOC-01-docs-quality-output.md` when tool splits vs parameters are ambiguous |

## Problem

Without an explicit policy, parity work oscillates between **strict tool-count stability** and **upstream-shaped tool names**, creating rework in `internal/registry/registry.go`, `docs/tools-inventory.md`, and MCP client integrations.

## Outcome

One **recorded decision** (option A, B, or C below) with sign-off in a PR, reflected in `docs/tools-inventory.md` and, if applicable, `docs/implementation-plan.md` Phase 4 exit criteria. Engineers can implement PAR/DOC/CON/DRV changes without reopening the debate each sprint.

## Scope

### In scope

- Choose A, B, or C and document consequences for registry validation and docs.
- Update header/policy statements in `docs/tools-inventory.md`.

### Out of scope

- Implementing PAR-01 itself (this epic is **decision + doc gates** only).

## Dependencies

- Current inventory: `docs/tools-inventory.md` (136 tools baseline).
- Registry: `internal/registry/registry.go` (SEP-986 naming).

## Acceptance criteria (Gherkin)

### AC1: Decision is singular and stored

```gherkin
Given stakeholders review options A B and C
When the decision PR merges
Then this file’s metadata records exactly one chosen option by letter
And the choice is restated in docs/tools-inventory.md header or policy subsection
```

### AC2: Implementation plan alignment

```gherkin
Given the chosen option changes Phase 4 assumptions about fixed tool count
When the decision PR merges
Then docs/implementation-plan.md Phase 4 exit criteria are updated to match
Or explicitly states tool count is no longer fixed at 136 with the new rule
```

### AC3: Downstream epics can cite the policy

```gherkin
Given PAR-01 or other parity epics are opened after the decision
When reviewers read the epic or PR description
Then each references POL-01 option letter and does not contradict it
```

### AC4: No code requirement for closure

```gherkin
Given POL-01 is a policy epic
When only documentation files change
Then golangci-lint and go test may still be required by repo CI for the PR
And the epic can close without production code changes
```

## Implementation notes

### MCP tool surface

- Decision drives whether new upstream tools become new `snake_case` registrations vs optional JSON on existing tools.

### Packages and registration

- No code changes required to **close** POL-01; subsequent epics modify `internal/tools/*` and `internal/registry/registry.go`.

### Verification

- Doc-only PR: `markdown` / human review; follow normal CI if any touched files trigger builds.

## Component inventory

| Path | Change |
|------|--------|
| `docs/roadmap/POL-01-tool-count-policy.md` | Modified |
| `docs/tools-inventory.md` | Modified |
| `docs/implementation-plan.md` | Modified (if exit criteria change) |

## Existing infrastructure to reuse

- N/A — policy artifact.

## Risks

| Risk | Impact | Mitigation |
|------|--------|------------|
| Hybrid option C without guardrails | endless judgment calls | Add a short “when to add a new tool” checklist in docs/tools-inventory.md |
| Decision delayed | PAR-01 blocked | Time-box decision; default to documented interim rule |

## Options (reference)

| Option | Pros | Cons |
|--------|------|------|
| **A — Strict 136** | Stable inventory, simpler docs, predictable MCP clients | Awkward when upstream’s tool split is clearer; richer single-tool schemas |
| **B — Track upstream** | Easier mental migration from Python MCP; clearer tool names | More registry/tier work; docs chase releases |
| **C — Hybrid** | Pragmatic balance | Requires ongoing judgment calls |

### Recommendation (PM)

Default to **C — Hybrid** unless the team commits to full parity automation: add tools when they **reduce ambiguity** or **match Google API resource boundaries**; otherwise extend existing tools with optional JSON fields.
