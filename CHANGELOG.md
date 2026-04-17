# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [1.4.0] — 2026-04-17

### Changed

- **README**: Restructured for 2026-style operator documentation — badges (CI, Go, release), role-oriented tables, TOC, prerequisites, GHCR pull instructions, expanded env reference with links to `docs/configuration.md`, troubleshooting table, contributing strip, clearer MCP transport defaults and spec compliance summary.

### Container images

Pushing git tag **`v1.4.0`** triggers [`.github/workflows/publish.yml`](.github/workflows/publish.yml), which builds and pushes **multi-arch** (`linux/amd64`, `linux/arm64`) images to **GitHub Container Registry** (`ghcr.io`) with semver tags (for example `v1.4.0`, `1.4`, `1`).

## [1.3.0] — 2026-04-16

### Added

- `docs/README.md` — documentation hub (by-role entry points, canonical table, 136 vs 137 MCP tool note).
- `docs/templates/roadmap_item.md` — SDD-ready epic template (Gherkin ACs, implementation notes, risks).
- `CHANGELOG.md` — changelog for roadmap hardening and releases.
- `.github/ISSUE_TEMPLATE/` — bug report and feature request forms; `dependabot.yml`; `pull_request_template.md`; `.github/SECURITY.md`.
- `docs/cursor-mcp.json.example` — sample Cursor MCP config.
- `docs/google-workspace-api-documentation.md` — index of official Google API documentation URLs.
- `.vscode/extensions.json` and `.vscode/settings.json` — recommended editor setup.

### Changed

- **README**: Hero links **`docs/README.md`**; `start.sh` default behavior clarifies **137** MCP tools vs **136** Workspace tools and OAuth 2.1; notes **plain JSON** token files with a pointer to `docs/security.md`.
- **Docs**: `docs/configuration.md` tier counts aligned with `docs/tools-inventory.md` (47 / 49 / 40 exclusive tiers, 96 / 136 cumulative). `docs/implementation-plan.md` adds a **Document status** banner (phases 1–4 delivered; roadmap is live planning). `docs/security.md` points operators to README + `docs/README.md` for token-at-rest context. `docs/roadmap/README.md` lists `docs/README.md` under repo goals.
- Roadmap: items use **`XXX-NN-itemname.md`** under `docs/roadmap/`; epics hardened to **`docs/templates/roadmap_item.md`** (Gherkin ACs, scope, implementation notes, component inventory, risks, **Priority**). **`docs/roadmap/README.md`** adds **Seq** (AUTH-01 → SEC-01 → DRV-01 → CON-01), a **POL-01** gate before **PAR-01**, and a Priority column in the items index. **SEC-01** upstream link corrected to PR **#682**. **MCP-01** uses **Horizon: Watchlist** per template.
- Completed roadmap items are **archived** under `docs/roadmap/archive/` via Product Manager **`/archive-roadmap-item`** (Status **Closed**, README + `archive/README.md` updated).
- **`.gitignore`**: ignore local Cursor paths (`.cursor/`, `AGENTS.md`, `.cursorignore`).

### Container images

Pushing git tag **`v1.3.0`** triggers [`.github/workflows/publish.yml`](.github/workflows/publish.yml), which builds and pushes **multi-arch** (`linux/amd64`, `linux/arm64`) images to **GitHub Container Registry** (`ghcr.io`) with semver tags (for example `v1.3.0`, `1.3`, `1`).
