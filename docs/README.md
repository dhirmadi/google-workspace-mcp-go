# Documentation index

Start here, then open the files that match your role.

## By role

| You are… | Start with | Then |
|----------|------------|------|
| **Running the server** (Docker, MCP URL, auth) | [`README.md`](../README.md) | [`configuration.md`](configuration.md), [`cursor-mcp.json.example`](cursor-mcp.json.example) |
| **Implementing or changing tools** | [`architecture.md`](architecture.md) | [`code-patterns.md`](code-patterns.md), [`tools-inventory.md`](tools-inventory.md), [`configs/tool_tiers.yaml`](../configs/tool_tiers.yaml) |
| **OAuth, scopes, or identity** | [`auth-and-scopes.md`](auth-and-scopes.md) | [`security.md`](security.md) |
| **Security / compliance review** | [`security.md`](security.md) | [`auth-and-scopes.md`](auth-and-scopes.md), [`README.md`](../README.md) (authentication section) |
| **Planning or sequencing work** | [`roadmap/README.md`](roadmap/README.md) | [`implementation-plan.md`](implementation-plan.md) (historical bootstrap), [`templates/roadmap_item.md`](templates/roadmap_item.md) |
| **Delivering roadmap items (Claude Code)** | [`../CLAUDE.md`](../CLAUDE.md) | Project **`.claude/commands/`** — `/implement-roadmapitem`, `/review-roadmapitem`, `/podman-mcp-local-verify`; **`.claude/agents/`** (e.g. Podman local test); skills under **`.claude/skills/`** |
| **Looking up official Google API URLs** | [`google-workspace-api-documentation.md`](google-workspace-api-documentation.md) | Product hubs linked from that page |

## Canonical references

| Topic | Document |
|-------|----------|
| System layout, MCP compliance, deferred features | [`architecture.md`](architecture.md) |
| Handler patterns, dual output, imports | [`code-patterns.md`](code-patterns.md) |
| Env vars, CLI flags, tiers (cumulative counts) | [`configuration.md`](configuration.md) |
| **Tool names, tiers, read-only flags** (contract for agents) | [`tools-inventory.md`](tools-inventory.md) |
| Scopes, OAuth 2.0 vs 2.1, callback behavior | [`auth-and-scopes.md`](auth-and-scopes.md) |
| Credentials, logging, transport, abuse limits | [`security.md`](security.md) |
| Tool count nuance | **136** Workspace tools in [`tools-inventory.md`](tools-inventory.md); default MCP registration adds **`start_google_auth`** → **137** tools unless OAuth 2.1 removes it (see [`auth-and-scopes.md`](auth-and-scopes.md)). |

## Roadmap and epics

- Overview and sequencing: [`roadmap/README.md`](roadmap/README.md)
- Closed items: [`roadmap/archive/README.md`](roadmap/archive/README.md)
- New epic template: [`templates/roadmap_item.md`](templates/roadmap_item.md)
- **Claude Code**: [`../CLAUDE.md`](../CLAUDE.md) and [`../.claude/`](../.claude/) — roadmap TDD + review commands and skills

## Counts and drift

Authoritative tool list: **`docs/tools-inventory.md`** (aligned with `configs/tool_tiers.yaml`). When you add or rename tools, update the inventory and tier YAML in the same change.
