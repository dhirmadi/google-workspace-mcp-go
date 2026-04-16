# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- `docs/templates/roadmap_item.md` — SDD-ready epic template (Gherkin ACs, implementation notes, risks).
- `CHANGELOG.md` — changelog for roadmap hardening and releases.

### Changed

- Roadmap: items use **`XXX-NN-itemname.md`** under `docs/roadmap/`; epics hardened to **`docs/templates/roadmap_item.md`** (Gherkin ACs, scope, implementation notes, component inventory, risks, **Priority**). **`docs/roadmap/README.md`** adds **Seq** (AUTH-01 → SEC-01 → DRV-01 → CON-01), a **POL-01** gate before **PAR-01**, and a Priority column in the items index. **SEC-01** upstream link corrected to PR **#682**. **MCP-01** uses **Horizon: Watchlist** per template.
- Completed roadmap items are **archived** under `docs/roadmap/archive/` via Product Manager **`/archive-roadmap-item`** (Status **Closed**, README + `archive/README.md` updated).
