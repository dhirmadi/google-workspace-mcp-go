---
description: Podman — build image, run MCP container, verify /mcp, teardown image
argument-hint: "[optional HOST_PORT] default 18000; set GOOGLE_OAUTH_* for real auth"
model: sonnet
---

Run the **Podman MCP local smoke test** for this repository.

Follow the full procedure in:

@.claude/agents/podman-mcp-local-test.md

Use the repository root as the working directory. After teardown, give a short pass/fail summary.
