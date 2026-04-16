# Roadmap

Outcome-oriented work derived from **repo goals** (`docs/implementation-plan.md`, `docs/tools-inventory.md`) and drift vs the upstream **[google_workspace_mcp](https://github.com/taylorwilsdon/google_workspace_mcp)** project (releases and community PRs). This folder is the **planning source**; implementation detail stays in `docs/code-patterns.md`, `docs/auth-and-scopes.md`, and service-specific docs.

**This `README.md` is the roadmap overview** — keep the **horizon summary**, **sequencing**, and **items index** current whenever you add, rename, split, merge, or change status/priority of roadmap files.

## Roadmap item filenames

Each item is a single markdown file in this directory, named:

`XXX-NN-itemname.md`

| Part | Meaning |
|------|---------|
| **`XXX`** | Short **topic** tag (3 letters, uppercase). Examples: `AUTH` (OAuth/auth flows), `SEC` (security hygiene), `DRV` (Drive), `DOC` (Docs output), `POL` (policy decisions), `PAR` (multi-service parity), `CON` (Contacts/Chat), `ENT` (enterprise), `MCP` (MCP protocol/client). |
| **`NN`** | Two-digit sequence **per topic** (`01`, `02`, …). Renumber only when splitting/merging items; otherwise use the next free number for that topic. |
| **`itemname`** | Lowercase **kebab-case** slug (no spaces). |

Examples: `AUTH-01-pkce-public-client.md`, `SEC-09-oauth-serverless-version.md`.

Cross-link other items with the **filename** (e.g. `` `POL-01-tool-count-policy.md` ``), not free-text-only references.

## How to use these items

Each file is one **epic** (shippable theme). Prefer small PRs that close a subset of **Gherkin acceptance criteria** in the item. Link PRs to the roadmap **ID** (e.g. `AUTH-01`) and filename in the PR description when relevant.

**Hardening for development:** Use the Product Manager **`/harden`** slash command in Cursor (see `.cursor/commands/product-manager/harden.md` locally) with the item path. Canonical structure and AC format live in **`docs/templates/roadmap_item.md`**. Hardening updates should add a line under **`CHANGELOG.md`** → `[Unreleased]`.

**Archiving completed work:** Use **`/archive-roadmap-item`** (see `.cursor/commands/product-manager/archive-roadmap-item.md` locally). That sets **Status** to **Closed**, moves the file to **`docs/roadmap/archive/`**, removes it from the **horizon** and **items index** tables below, and appends a row to the **Archived items** table in **`docs/roadmap/archive/README.md`**.

## Upstream reference

- Repository: [taylorwilsdon/google_workspace_mcp](https://github.com/taylorwilsdon/google_workspace_mcp)
- Recent release themes (PKCE, Drive PDF/attachments, Contacts, Chat stability, security): see [releases](https://github.com/taylorwilsdon/google_workspace_mcp/releases).

## Horizon summary

| Horizon | Seq | ID | Priority | Epic | One-line outcome |
|---------|-----|-----|----------|------|------------------|
| **Now** | 1 | AUTH-01 | P0 | [PKCE & public client](./AUTH-01-pkce-public-client.md) | Optional client secret; PKCE where Google allows; docs + console steps |
| **Next** | 1 | SEC-01 | P0 | [Security hygiene (upstream-informed)](./SEC-01-security-hygiene-upstream.md) | Credential/path/input invariants audited before expanding read surfaces |
| **Next** | 2 | DRV-01 | P1 | [Drive: content fidelity](./DRV-01-drive-file-content-fidelity.md) | `get_drive_file_content` matches PDF/binary expectations with bounded behavior |
| **Next** | 3 | CON-01 | P1 | [Contacts & Chat reliability](./CON-01-contacts-chat-reliability.md) | Multi-value contacts + deterministic Chat search vs upstream |
| **Policy** | — | POL-01 | P0 | [Tool count vs upstream](./POL-01-tool-count-policy.md) | **Gate before PAR-01 merges:** choose A/B/C; update inventory + plan |
| **Then** | 1 | PAR-01 | P1 | [Calendar, Sheets, Gmail parity](./PAR-01-calendar-sheets-gmail-parity.md) | Close feature drift per POL-01 decision |
| **Then** | 2 | DOC-01 | P2 | [Docs: quality of output](./DOC-01-docs-quality-output.md) | Markdown/table fidelity for agent-consumed Doc content |
| **Deferred** | — | ENT-01 | P3 | [Enterprise: DWD / service accounts](./ENT-01-domain-wide-delegation-deferred.md) | Only if explicitly required; high risk |
| **Watchlist** | — | MCP-01 | P2 | [MCP client compat shims](./MCP-01-client-compat-shims.md) | Shims only with named client + removal condition |

**Seq** is the recommended merge order within a horizon (lower first). `—` means not sequenced against siblings (single item or non-executable).

### Sequencing rules

1. **AUTH-01** ships first — everything depends on auth/config correctness.
2. **SEC-01** immediately follows AUTH-01 — credential and path invariants before expanding high-risk read paths (**DRV-01**) and multi-value APIs (**CON-01**).
3. **POL-01** must record **one** option (A/B/C) before **PAR-01** implementation PRs merge; backlog research for PAR-01 may run in parallel, but merges that add or split tools must cite the signed policy (**PAR-01** AC1).
4. **DOC-01** can proceed independently of PAR-01 after **POL-01** if tool-split choices are already satisfied; if in doubt, sequence DOC-01 after PAR-01’s first merged slice.

## Items index

| ID | Priority | File | Status |
|----|----------|------|--------|
| AUTH-01 | P0 | [AUTH-01-pkce-public-client.md](./AUTH-01-pkce-public-client.md) | Proposed |
| SEC-01 | P0 | [SEC-01-security-hygiene-upstream.md](./SEC-01-security-hygiene-upstream.md) | Proposed |
| DRV-01 | P1 | [DRV-01-drive-file-content-fidelity.md](./DRV-01-drive-file-content-fidelity.md) | Proposed |
| CON-01 | P1 | [CON-01-contacts-chat-reliability.md](./CON-01-contacts-chat-reliability.md) | Proposed |
| POL-01 | P0 | [POL-01-tool-count-policy.md](./POL-01-tool-count-policy.md) | **Decision needed** |
| PAR-01 | P1 | [PAR-01-calendar-sheets-gmail-parity.md](./PAR-01-calendar-sheets-gmail-parity.md) | Proposed |
| DOC-01 | P2 | [DOC-01-docs-quality-output.md](./DOC-01-docs-quality-output.md) | Proposed |
| ENT-01 | P3 | [ENT-01-domain-wide-delegation-deferred.md](./ENT-01-domain-wide-delegation-deferred.md) | Deferred |
| MCP-01 | P2 | [MCP-01-client-compat-shims.md](./MCP-01-client-compat-shims.md) | Watchlist |

## Archived items

Completed items are listed in **[`archive/README.md`](./archive/README.md)** (table of **Closed** items with links under `docs/roadmap/archive/`). Do not keep closed items in the horizon or items index above.
