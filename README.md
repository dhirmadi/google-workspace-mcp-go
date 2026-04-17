# Google Workspace MCP Server (Go)

[![CI](https://github.com/evert/google-workspace-mcp-go/actions/workflows/ci.yml/badge.svg)](https://github.com/evert/google-workspace-mcp-go/actions/workflows/ci.yml)
[![Go](https://img.shields.io/github/go-mod/go-version/evert/google-workspace-mcp-go?label=Go)](go.mod)
[![Release](https://img.shields.io/github/v/release/evert/google-workspace-mcp-go?label=release)](https://github.com/evert/google-workspace-mcp-go/releases)

A **[Model Context Protocol](https://modelcontextprotocol.io/)** server in **Go 1.24** that exposes **Google Workspace** to AI agents: Gmail, Drive, Calendar, Docs, Sheets, Slides, Chat, Forms, Tasks, Contacts, Programmable Search, and Apps Script. Implements **tools** targeting MCP spec **2025-11-25** (via [`github.com/modelcontextprotocol/go-sdk`](https://github.com/modelcontextprotocol/go-sdk)).

| | |
| :--- | :--- |
| **Workspace tools** | **136** ([`docs/tools-inventory.md`](docs/tools-inventory.md)) |
| **Default MCP tools** | **137** (includes `start_google_auth`; OAuth 2.1 mode → **136** — [`docs/auth-and-scopes.md`](docs/auth-and-scopes.md)) |
| **Image size** | **~33 MB** (multi-stage build, distroless, non-root) |
| **Doc hub (by role)** | **[`docs/README.md`](docs/README.md)** |
| **Changelog** | [`CHANGELOG.md`](CHANGELOG.md) |
| **Security** | [`docs/security.md`](docs/security.md) · [vulnerability reporting](.github/SECURITY.md) |

## Contents

- [Prerequisites](#prerequisites)
- [Quick start](#quick-start)
- [Prebuilt container images](#prebuilt-container-images)
- [Getting started (detailed)](#getting-started-detailed)
- [`start.sh` reference](#startsh-reference)
- [What's included](#whats-included)
- [Configuration](#configuration)
- [Authentication](#authentication)
- [Service-specific notes](#service-specific-notes)
- [Troubleshooting](#troubleshooting)
- [Development](#development)
- [MCP spec compliance](#mcp-spec-compliance)
- [Contributing](#contributing)
- [License](#license)

---

## Prerequisites

- **Docker** (recommended) *or* **Go 1.24+** to build from source
- A **Google Cloud** project with **OAuth 2.0 Client ID** (type: **Web application**) and **Authorized redirect URIs** matching your server port (default `http://localhost:8000/oauth/callback`)
- **APIs enabled** in Google Cloud for the products you use (Gmail, Drive, Calendar, and so on)

---

## Quick start

1. **OAuth client** — [Google Cloud Console → Credentials](https://console.cloud.google.com/apis/credentials): Web client, redirect URI as above, enable needed APIs.
2. **Run** — from a clone of this repository:

   ```bash
   ./start.sh "YOUR_CLIENT_ID.apps.googleusercontent.com" "YOUR_CLIENT_SECRET"
   ```

3. **Connect** — MCP URL **`http://localhost:8000/mcp`** (adjust host/port if needed). Copy **[`docs/cursor-mcp.json.example`](docs/cursor-mcp.json.example)** into **`.cursor/mcp.json`** (or your client’s equivalent).

4. **Authenticate** — first Google tool use triggers OAuth; with legacy mode the agent calls **`start_google_auth`** and you complete consent in the browser.

For **Cursor**: **`.cursor/`** is local-only (not committed); use **`docs/cursor-mcp.json.example`** as the template for **`mcp.json`**.

---

## Prebuilt container images

Tagged releases build **multi-arch** images (**`linux/amd64`**, **`linux/arm64`**) and push to **GitHub Container Registry** via [`.github/workflows/publish.yml`](.github/workflows/publish.yml).

Replace **`OWNER/REPO`** with your GitHub **`owner/repo`** (for this module’s home, **`evert/google-workspace-mcp-go`**):

```bash
docker pull ghcr.io/OWNER/REPO:v1.3.0
```

Tags typically include the **semver** (`v1.3.0`), **major.minor** (`1.3`), **major** (`1`), and a **git SHA**. See [**Releases**](https://github.com/evert/google-workspace-mcp-go/releases) for the current version.

The [`Dockerfile`](Dockerfile) defaults **`MCP_TRANSPORT=streamable-http`** and **`MCP_PORT=8000`**, which matches HTTP-based MCP clients (Cursor, Claude Desktop) using **`http://…/mcp`**.

---

## Getting started (detailed)

### Step 1: Get Google OAuth credentials

1. Open [Google Cloud Console → Credentials](https://console.cloud.google.com/apis/credentials).
2. Create or select a project.
3. **APIs & Services → Credentials → Create credentials → OAuth client ID → Web application**.
4. Under **Authorized redirect URIs**, add **`http://localhost:8000/oauth/callback`** (and the same host with another port if you use `--port`).
5. Copy **Client ID** and **Client Secret**.
6. **APIs & Services → Library** — enable each API you need (Gmail, Drive, Calendar, …).

### Step 2: Build the container (optional if using `start.sh` alone)

```bash
git clone https://github.com/evert/google-workspace-mcp-go.git
cd google-workspace-mcp-go
docker build -t google-workspace-mcp .
```

Image is ~33 MB (distroless, non-root).

### Step 3: Start the server

```bash
./start.sh "YOUR_CLIENT_ID.apps.googleusercontent.com" "YOUR_CLIENT_SECRET"
```

Default: **HTTP** MCP at **`http://localhost:8000/mcp`**, **137** MCP tools (136 Workspace + `start_google_auth`).

### Step 4: Connect your MCP client

**Cursor** — **`.cursor/mcp.json`** (project or user config):

```json
{
  "mcpServers": {
    "google-workspace": {
      "url": "http://localhost:8000/mcp"
    }
  }
}
```

**Claude Desktop** — `claude_desktop_config.json` with the same `url` shape.

> If you use **`--port 9000`**, set the URL to **`http://localhost:9000/mcp`** and add the matching redirect URI in Google Cloud Console.

### Step 5: Authenticate

On first use of a Google tool, complete OAuth (legacy flow: agent calls **`start_google_auth`** with your email, you open the URL, grant access). Tokens refresh automatically when persisted.

---

## `start.sh` reference

```bash
./start.sh <CLIENT_ID> <CLIENT_SECRET> [OPTIONS]
```

`CLIENT_ID` and `CLIENT_SECRET` are required; everything else is optional.

### Default behavior (no extra flags)

```bash
./start.sh "YOUR_CLIENT_ID.apps.googleusercontent.com" "YOUR_SECRET"
```

- **All 12 services** — **137** MCP tools by default (**136** Workspace tools per [`docs/tools-inventory.md`](docs/tools-inventory.md) plus **`start_google_auth`**; OAuth 2.1 omits the auth tool → **136** — [`docs/auth-and-scopes.md`](docs/auth-and-scopes.md))
- **Port `8000`** — MCP **`http://localhost:8000/mcp`**, OAuth callback **`http://localhost:8000/oauth/callback`**
- **In-memory auth** unless **`--persistent-auth`** (tokens lost on container restart)
- **Auto-restart** container on failure (when managed by `start.sh` / Docker as documented)

The image is built on first run if missing; later runs reuse the cached image.

### Options

| Flag | Default | Description |
|------|---------|-------------|
| `--services SVCS` | all 12 services | Comma-separated: `gmail`, `drive`, `calendar`, … |
| `--port PORT` | `8000` | HTTP port (OAuth callback follows this port) |
| `--persistent-auth` | off | Persist OAuth tokens in a Docker volume |
| `--email EMAIL` | — | Default Google account (single-user convenience) |
| `--cse-id ID` | — | Programmable Search Engine ID |
| `--log-level LEVEL` | `info` | `debug`, `info`, `warn`, `error` |
| `--rebuild` | — | Force image rebuild |
| `--stop` | — | Stop and remove the container (volume preserved) |

### Examples

**All services, default port**

```bash
./start.sh "YOUR_CLIENT_ID.apps.googleusercontent.com" "GOCSPX-yourSecret"
# → 137 MCP tools on http://localhost:8000/mcp
```

**Subset: Gmail, Drive, Calendar**

```bash
./start.sh "YOUR_CLIENT_ID" "YOUR_SECRET" --services gmail,drive,calendar
# → 38 tools (15 Gmail + 16 Drive + 6 Calendar + 1 auth); OAuth scopes limited to those services
```

**Gmail only + default email**

```bash
./start.sh "YOUR_CLIENT_ID" "YOUR_SECRET" --services gmail --email user@company.com
# → 16 tools (15 Gmail + 1 auth)
```

**Gmail, Calendar, Docs, Sheets**

```bash
./start.sh "YOUR_CLIENT_ID" "YOUR_SECRET" --services gmail,calendar,docs,sheets
```

**Different port**

```bash
./start.sh "YOUR_CLIENT_ID" "YOUR_SECRET" --port 9000
# Add http://localhost:9000/oauth/callback to the OAuth client redirect URIs
```

**Persistent tokens**

```bash
./start.sh "YOUR_CLIENT_ID" "YOUR_SECRET" --persistent-auth
```

**Debug logs**

```bash
./start.sh "YOUR_CLIENT_ID" "YOUR_SECRET" --log-level debug
# docker logs -f google-workspace-mcp
```

**Rebuild image**

```bash
./start.sh "YOUR_CLIENT_ID" "YOUR_SECRET" --rebuild
```

**Stop**

```bash
./start.sh --stop
```

**Combine flags**

```bash
./start.sh "YOUR_CLIENT_ID" "YOUR_SECRET" \
  --services gmail,drive,calendar \
  --port 9000 \
  --email user@company.com \
  --log-level debug
```

### Available services (`--services`)

| Service | Flag | Tools |
|---------|------|-------|
| Gmail | `gmail` | 15 |
| Google Drive | `drive` | 16 |
| Google Calendar | `calendar` | 6 |
| Google Docs | `docs` | 19 |
| Google Sheets | `sheets` | 14 |
| Google Chat | `chat` | 4 |
| Google Forms | `forms` | 6 |
| Google Slides | `slides` | 9 |
| Google Tasks | `tasks` | 12 |
| Google Contacts | `contacts` | 15 |
| Programmable Search | `search` | 3 |
| Apps Script | `appscript` | 17 |

Limiting **`--services`** reduces both **tool surface** and **OAuth scope** requests at consent time.

---

## What's included

### Services and tools

| Service | Tools | Capabilities (summary) |
|---------|-------|-------------------------|
| **Gmail** | 15 | Search, read, send, drafts, labels, filters, attachments, batch |
| **Drive** | 16 | Search, read, create, share, permissions, batch |
| **Calendar** | 6 | Calendars, events, create/update/delete, free/busy |
| **Docs** | 19 | Read/write, tables, images, comments, find/replace, PDF export |
| **Sheets** | 14 | Read/write, formatting, conditional formatting, comments |
| **Chat** | 4 | Spaces, read/search/send |
| **Forms** | 6 | Forms, responses, layout |
| **Slides** | 9 | Decks, pages, thumbnails, comments |
| **Tasks** | 12 | Tasks and lists, move, clear completed |
| **Contacts** | 15 | People API, groups, batch |
| **Search** | 3 | Custom Search Engine queries |
| **Apps Script** | 17 | Projects, deployments, versions, execute, metrics |
| **Total** | **136** | **+1** auth tool **`start_google_auth`** = **137** MCP tools (default legacy OAuth) |

### Tool annotations

Every tool declares MCP **annotations** so clients can reason about safety and retries:

- **ReadOnlyHint** — no writes (e.g. `search_gmail_messages`)
- **DestructiveHint** — irreversible (e.g. `delete_event`)
- **IdempotentHint** — safe to retry (e.g. `modify_event`)
- **OpenWorldHint** — external side effects (essentially all integration tools)

---

## Configuration

### Environment variables (common)

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `GOOGLE_OAUTH_CLIENT_ID` | **Yes** | — | OAuth 2.0 client ID |
| `GOOGLE_OAUTH_CLIENT_SECRET` | **Cond.** | — | OAuth 2.0 client secret (required unless public client mode) |
| `GOOGLE_OAUTH_PUBLIC_CLIENT` | No | `false` | PKCE public client — no secret needed ([details](docs/auth-and-scopes.md#confidential-vs-public-client-pkce)) |
| `ENABLED_SERVICES` | No | all | Comma-separated service list (same names as `--services`) |
| `MCP_TRANSPORT` | No | `stdio` * | Transport: `stdio` or `streamable-http` (*`Dockerfile` defaults to `streamable-http`*) |
| `MCP_PORT` / `PORT` | No | `8000` | HTTP port |
| `WORKSPACE_MCP_HOST` | No | `0.0.0.0` | Bind address |
| `WORKSPACE_MCP_BASE_URI` | No | `http://localhost` | Base URL for OAuth callback construction |
| `WORKSPACE_MCP_PERSISTENT_AUTH` | No | `false` | Persist tokens under `WORKSPACE_MCP_CREDENTIALS_DIR` |
| `WORKSPACE_MCP_CREDENTIALS_DIR` | No | `~/.google_workspace_mcp/credentials` | Token directory (with persistent auth) |
| `WORKSPACE_MCP_READ_ONLY` | No | `false` | Read-only scopes; write tools filtered out |
| `TOOL_TIER` | No | `complete` | `core`, `extended`, or `complete` (cumulative) |
| `GOOGLE_CSE_ID` | No | — | Required for Search tools |
| `LOG_LEVEL` | No | `info` | `debug`, `info`, `warn`, `error` |
| `MCP_SINGLE_USER_MODE` | No | `false` | Single-user session behavior (see `docker-compose.yml` / `.env.example`) |
| `MCP_ENABLE_OAUTH21` | No | `false` | OAuth 2.1 / client-mediated auth ([`docs/auth-and-scopes.md`](docs/auth-and-scopes.md)) |

Full tables (including stateless mode): **[`docs/configuration.md`](docs/configuration.md)**.

### Docker Compose

```bash
cp .env.example .env
# Edit .env with your credentials
docker compose up --build
```

For persistent auth, set **`WORKSPACE_MCP_PERSISTENT_AUTH=true`** and mount **`/data/credentials`** as in [`docker-compose.yml`](docker-compose.yml).

---

## Authentication

```text
1. Agent calls start_google_auth with the user's email (legacy OAuth)
2. Server returns Google consent URL
3. User completes consent in the browser
4. Google redirects to http://localhost:<port>/oauth/callback
5. Server stores tokens (and refreshes them automatically)
6. Subsequent tool calls use the stored session
```

Match **redirect URIs** in Google Cloud Console to **`WORKSPACE_MCP_BASE_URI`** and **`MCP_PORT`** (or `start.sh --port`).

**Persistence:** without **`--persistent-auth`** / **`WORKSPACE_MCP_PERSISTENT_AUTH`**, tokens live **in memory** and are lost on restart. With persistence, files are **`0600`**, directory **`0700`**.

**At-rest format:** persisted tokens are **plain JSON** in v1; see **[`docs/security.md`](docs/security.md)** for threat model and future encryption/keyring notes.

---

## Service-specific notes

### Google Chat

Requires **Google Workspace** (not consumer Gmail alone). The Chat app may need configuration in the **Workspace Admin** console.

### Programmable Search

Requires **`GOOGLE_CSE_ID`** from [Programmable Search Engine](https://programmablesearchengine.google.com).

### Apps Script

**`run_script_function`** requires deployment as an **API executable** and **edit** access to the project (~30 calls/min typical quota behavior).

### Contacts

Uses the **Google People API** (legacy Contacts API is deprecated). Tool names say **contacts** for clarity.

---

## Troubleshooting

| Symptom | What to check |
|--------|----------------|
| **MCP client cannot connect** | Server running; URL **`http://HOST:PORT/mcp`**; firewall; same transport (**`streamable-http`** for HTTP clients). |
| **OAuth redirect mismatch** | Redirect URI in Google Cloud **exactly** matches **`http://localhost:<port>/oauth/callback`** (scheme, host, port, path). |
| **“No tools” / empty tool list** | **`ENABLED_SERVICES`** / **`--services`** not overly narrow; **`TOOL_TIER`** not `core` unless intended. |
| **Search tools fail** | **`GOOGLE_CSE_ID`** / **`--cse-id`** set. |
| **Chat always errors** | Workspace account and Chat API / app configuration. |
| **429 / rate limit** | Back off; see Google quotas; batch tools may emit progress ([`docs/architecture.md`](docs/architecture.md)). |

Agent-facing errors are mapped to actionable messages in middleware — see **`internal/middleware/errors.go`** and **[`docs/architecture.md`](docs/architecture.md)**.

---

## Development

### Build from source

```bash
go build -o server ./cmd/server

export GOOGLE_OAUTH_CLIENT_ID="your-client-id"
export GOOGLE_OAUTH_CLIENT_SECRET="your-secret"

./server                              # stdio (default from env)
./server --transport streamable-http  # HTTP MCP
```

### Tests and lint

```bash
go test ./...
go test -race ./...
GOOGLE_OAUTH_CLIENT_ID=test GOOGLE_OAUTH_CLIENT_SECRET=test \
  go test -tags=integration ./internal/integration/
golangci-lint run ./...
```

Architecture, registry behavior, and patterns: **[`docs/architecture.md`](docs/architecture.md)**, **[`docs/code-patterns.md`](docs/code-patterns.md)**.

### Project structure

```text
cmd/server/main.go           Entry point, transports, wiring
internal/
  auth/                      OAuth2, scopes, callback, token persistence
  config/                    Env, flags, tier YAML
  registry/registry.go       Tool registration, tier/service/read-only filters
  services/factory.go        Google API client factory
  tools/<service>/           Per-product tools + handlers
  comments/                  Shared Drive-backed comments (Docs/Sheets/Slides)
  middleware/                Logging, Google error mapping, retry
  pkg/                       Response builder, HTML/Office helpers
configs/tool_tiers.yaml      Tier assignments
```

---

## MCP spec compliance

Targeting **MCP 2025-11-25**:

| Feature | Status |
|---------|--------|
| Tools (Workspace + optional auth tool) | Implemented |
| Tool annotations | Implemented |
| Structured output (dual text + typed where applicable) | Implemented |
| Progress notifications | Implemented (batch / long-running tools) |
| Tool icons (per service) | Implemented |
| SDK middleware | Implemented |
| Resources / Prompts | Deferred (see [`docs/architecture.md`](docs/architecture.md)) |

---

## Contributing

Use **[`docs/code-patterns.md`](docs/code-patterns.md)** for tool handlers and output conventions. Before opening a PR, run **`golangci-lint run ./...`** and **`go test -race ./...`** (see [`.github/workflows/ci.yml`](.github/workflows/ci.yml)). Use [`.github/pull_request_template.md`](.github/pull_request_template.md) when filing changes.

---

## License

See the **`LICENSE`** file in the repository root when your checkout includes it.
