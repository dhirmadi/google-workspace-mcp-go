---
name: roadmap-spec-tdd
description: >-
  Spec-first TDD for docs/roadmap/XXX-NN-itemname.md items. Use when the user
  or /implement-roadmapitem asks to implement, continue, or verify work against
  acceptance criteria. Red-green-refactor; tests before production code.
---

# Roadmap item — spec-first TDD

## Locate the spec

- Active items: `docs/roadmap/XXX-NN-itemname.md`
- Archived (read-only review): `docs/roadmap/archive/XXX-NN-itemname.md`
- Overview: `docs/roadmap/README.md`
- Hardening template: `docs/templates/roadmap_item.md`

## Cycle

1. Parse **Acceptance criteria** (Gherkin or checklist). Order by dependency.
2. For each AC: **failing test** → minimal **implementation** → **refactor**.
3. Prefer **table-driven** tests in the same package as handlers; `//go:build integration` only when documented env is required.
4. Map Gherkin **Then** clauses to assertions with stable, observable behavior (HTTP status, tool result text, structured field values).

## Done criteria

- All in-scope ACs have passing tests or documented objective verification.
- `golangci-lint run ./...` and `go test -race ./...` pass.
- Roadmap file updated (checked items / status **Done** when complete — not **Closed** until archive command).
