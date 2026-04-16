# Security

## Reporting a vulnerability

Please **do not** open a public issue for security-sensitive problems.

Instead, contact the repository maintainers through GitHub **private vulnerability reporting** for this repository (Security → Advisories), or another channel your organization uses for this project.

Include: affected version or commit, reproduction steps, and impact (confidentiality / integrity / availability). Redact secrets, tokens, and personal data from reports.

## Operational reminders

- OAuth client secrets must never be committed. Use `.env` (gitignored) or your secret manager.
- Local MCP client config (e.g. Cursor’s `.cursor/mcp.json`) should stay on your machine; use `docs/cursor-mcp.json.example` in this repo as a template.
