# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- `docs/README.md` — documentation hub (by-role entry points, canonical table, 136 vs 137 MCP tool note).
- `docs/templates/roadmap_item.md` — SDD-ready epic template (Gherkin ACs, implementation notes, risks).
- `CHANGELOG.md` — changelog for roadmap hardening and releases.

### Changed

- **README**: Hero links **`docs/README.md`**; `start.sh` default behavior clarifies **137** MCP tools vs **136** Workspace tools and OAuth 2.1; notes **plain JSON** token files with a pointer to `docs/security.md`.
- **Docs**: `docs/configuration.md` tier counts aligned with `docs/tools-inventory.md` (47 / 49 / 40 exclusive tiers, 96 / 136 cumulative). `docs/implementation-plan.md` adds a **Document status** banner (phases 1–4 delivered; roadmap is live planning). `docs/security.md` points operators to README + `docs/README.md` for token-at-rest context. `docs/roadmap/README.md` lists `docs/README.md` under repo goals.
- Roadmap: items use **`XXX-NN-itemname.md`** under `docs/roadmap/`; epics hardened to **`docs/templates/roadmap_item.md`** (Gherkin ACs, scope, implementation notes, component inventory, risks, **Priority**). **`docs/roadmap/README.md`** adds **Seq** (AUTH-01 → SEC-01 → DRV-01 → CON-01), a **POL-01** gate before **PAR-01**, and a Priority column in the items index. **SEC-01** upstream link corrected to PR **#682**. **MCP-01** uses **Horizon: Watchlist** per template.
- Completed roadmap items are **archived** under `docs/roadmap/archive/` via Product Manager **`/archive-roadmap-item`** (Status **Closed**, README + `archive/README.md` updated).
