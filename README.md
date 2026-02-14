# Google Workspace MCP Server (Go)

A [Model Context Protocol](https://modelcontextprotocol.io/) server that gives AI agents full access to Google Workspace — Gmail, Drive, Calendar, Docs, Sheets, Slides, Chat, Forms, Tasks, Contacts, Search, and Apps Script.

**136 tools. 12 services. One container. ~33 MB image.**

---

## Getting Started

### Step 1: Get Google OAuth Credentials

1. Go to the [Google Cloud Console](https://console.cloud.google.com/apis/credentials)
2. Create a new project (or select an existing one)
3. Go to **APIs & Services > Credentials**
4. Click **Create Credentials > OAuth 2.0 Client ID**
5. Choose **Web application** as the application type
6. Under **Authorized redirect URIs**, add: `http://localhost:8000/oauth/callback`
   (if you plan to use a different port with `--port`, add that URI too, e.g., `http://localhost:9000/oauth/callback`)
7. Copy the **Client ID** and **Client Secret**
8. Go to **APIs & Services > Library** and enable the APIs you need:
   - Gmail API, Google Drive API, Google Calendar API, etc.

### Step 2: Build the Container

```bash
git clone https://github.com/evert/google-workspace-mcp-go.git
cd google-workspace-mcp-go
docker build -t google-workspace-mcp .
```

That's it. The image is ~33 MB (distroless, non-root).

### Step 3: Start the Server

Use the `start.sh` script:

```bash
./start.sh "YOUR_CLIENT_ID.apps.googleusercontent.com" "YOUR_CLIENT_SECRET"
```

The server starts on `http://localhost:8000/mcp` with all 136 tools enabled.

### Step 4: Connect Your MCP Client

Add to your **Cursor** settings (`.cursor/mcp.json`):

```json
{
  "mcpServers": {
    "google-workspace": {
      "url": "http://localhost:8000/mcp"
    }
  }
}
```

Or for **Claude Desktop** (`claude_desktop_config.json`):

```json
{
  "mcpServers": {
    "google-workspace": {
      "url": "http://localhost:8000/mcp"
    }
  }
}
```

> If you used `--port 9000`, change the URL to `http://localhost:9000/mcp`.

### Step 5: Authenticate

The first time you (or the AI agent) use a Google tool, you'll be prompted to authenticate. The agent calls `start_google_auth` with your email, and you open the returned URL in your browser. After granting access, you're all set — tokens are stored and auto-refresh transparently.

---

## start.sh Reference

```bash
./start.sh <CLIENT_ID> <CLIENT_SECRET> [OPTIONS]
```

The two required arguments are your Google OAuth Client ID and Client Secret. Everything else is optional.

### What Happens with No Options

```bash
./start.sh "YOUR_CLIENT_ID.apps.googleusercontent.com" "YOUR_SECRET"
```

This starts the server with:

- **All 12 services enabled** (137 tools total)
- **Port 8000** — MCP endpoint at `http://localhost:8000/mcp`
- **OAuth callback** at `http://localhost:8000/oauth/callback`
- **In-memory auth** — tokens are stored in memory only (lost on restart). Use `--persistent-auth` to persist tokens to a Docker volume.
- **Container auto-restarts** if it crashes or Docker Desktop restarts

The image is built automatically on first run. Subsequent runs reuse the cached image.

### Options

| Flag | Default | Description |
|------|---------|-------------|
| `--services SVCS` | all 12 services | Comma-separated list of services to enable |
| `--port PORT` | `8000` | HTTP port to expose (OAuth callback adjusts automatically) |
| `--persistent-auth` | off | Persist OAuth tokens to a Docker volume (survives restarts) |
| `--email EMAIL` | — | Your Google email for authentication |
| `--cse-id ID` | — | Google Custom Search Engine ID |
| `--log-level LEVEL` | `info` | `debug`, `info`, `warn`, or `error` |
| `--rebuild` | — | Force rebuild the Docker image |
| `--stop` | — | Stop and remove the running container |

### Examples

**All services on default port** — the simplest way to get everything running:

```bash
./start.sh "YOUR_CLIENT_ID.apps.googleusercontent.com" "GOCSPX-yourSecret"
# → 137 tools on http://localhost:8000/mcp
```

**Only Gmail, Drive, and Calendar** — fewer tools means less noise for the AI agent:

```bash
./start.sh "YOUR_CLIENT_ID" "YOUR_SECRET" --services gmail,drive,calendar
# → 38 tools (15 Gmail + 16 Drive + 6 Calendar + 1 auth)
# → Only requests OAuth scopes for those 3 services
```

**Just Gmail for email automation:**

```bash
./start.sh "YOUR_CLIENT_ID" "YOUR_SECRET" --services gmail --email user@company.com
# → 16 tools (15 Gmail + 1 auth)
# → Single-user mode: no need to pass email in every tool call
```

**Productivity suite** — Gmail, Calendar, Docs, and Sheets:

```bash
./start.sh "YOUR_CLIENT_ID" "YOUR_SECRET" --services gmail,calendar,docs,sheets
# → 55 tools for everyday office work
```

**Run on a different port** — useful if port 8000 is taken:

```bash
./start.sh "YOUR_CLIENT_ID" "YOUR_SECRET" --port 9000
# → MCP endpoint at http://localhost:9000/mcp
# → OAuth callback automatically adjusts to http://localhost:9000/oauth/callback
# → Remember to add http://localhost:9000/oauth/callback to your
#   Google Cloud Console redirect URIs
```

**Persistent authentication** — tokens survive container restarts:

```bash
./start.sh "YOUR_CLIENT_ID" "YOUR_SECRET" --persistent-auth
# → Tokens stored in Docker volume 'mcp-credentials'
# → Users don't need to re-authenticate after restarts
```

**Debug logging** — see every request and response in the container logs:

```bash
./start.sh "YOUR_CLIENT_ID" "YOUR_SECRET" --log-level debug
# Then: docker logs -f google-workspace-mcp
```

**Force rebuild** — after pulling new code or if the image seems stale:

```bash
./start.sh "YOUR_CLIENT_ID" "YOUR_SECRET" --rebuild
```

**Stop the server:**

```bash
./start.sh --stop
# Stops and removes the container. Credentials volume is preserved.
```

**Combine options:**

```bash
./start.sh "YOUR_CLIENT_ID" "YOUR_SECRET" \
  --services gmail,drive,calendar \
  --port 9000 \
  --email user@company.com \
  --log-level debug
```

### Available Services

| Service | Flag value | Tools |
|---------|-----------|-------|
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
| Custom Search | `search` | 3 |
| Apps Script | `appscript` | 17 |

> **Tip**: When you limit services with `--services`, the server only requests OAuth scopes for those services. This means users grant fewer permissions during authentication.

---

## What's Included

### Services & Tools

| Service | Tools | What you can do |
|---------|-------|-----------------|
| **Gmail** | 15 | Search, read, send, draft, labels, filters, attachments, batch ops |
| **Drive** | 16 | Search, read, create, upload, share, permissions, batch share |
| **Calendar** | 6 | List calendars, events, create/modify/delete events, freebusy |
| **Docs** | 19 | Read/create/edit documents, tables, images, comments, find & replace, PDF export |
| **Sheets** | 14 | Create/read/modify spreadsheets, formatting, conditional formatting, comments |
| **Chat** | 4 | List spaces, read/search/send messages |
| **Forms** | 6 | Create forms, read responses, conditional formatting |
| **Slides** | 9 | Create/read presentations, pages, thumbnails, comments |
| **Tasks** | 12 | Full CRUD on tasks and task lists, move, clear completed |
| **Contacts** | 15 | Search/CRUD contacts and groups, batch create/update/delete |
| **Search** | 3 | Google Custom Search Engine queries |
| **Apps Script** | 17 | Manage projects, deployments, versions, run functions, view metrics |
| **Total** | **136** | + 1 auth tool = **137** |

### Tool Annotations

Every tool has MCP annotations that tell AI agents what the tool does:

- **ReadOnlyHint** — safe to call, no side effects (e.g., `search_gmail_messages`)
- **DestructiveHint** — irreversible action (e.g., `delete_event`, `delete_contact`)
- **IdempotentHint** — safe to retry (e.g., `modify_event`, `update_task`)
- **OpenWorldHint** — interacts with external systems (all tools)

---

## Configuration Reference

### Environment Variables

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `GOOGLE_OAUTH_CLIENT_ID` | **Yes** | — | OAuth 2.0 client ID |
| `GOOGLE_OAUTH_CLIENT_SECRET` | **Yes** | — | OAuth 2.0 client secret |
| `ENABLED_SERVICES` | No | all | Comma-separated services to enable |
| `MCP_TRANSPORT` | No | `stdio` | Transport: `stdio` or `streamable-http` |
| `MCP_PORT` | No | `8000` | HTTP server port |
| `WORKSPACE_MCP_HOST` | No | `0.0.0.0` | HTTP bind address |
| `WORKSPACE_MCP_BASE_URI` | No | `http://localhost` | Base URI for OAuth callbacks |
| `WORKSPACE_MCP_PERSISTENT_AUTH` | No | `false` | Persist OAuth tokens to disk (survives restarts) |
| `WORKSPACE_MCP_CREDENTIALS_DIR` | No | `~/.google_workspace_mcp/credentials` | Token storage directory (only used with persistent auth) |
| `WORKSPACE_MCP_READ_ONLY` | No | `false` | Read-only mode (only read scopes) |
| `TOOL_TIER` | No | `complete` | Tool tier: `core`, `extended`, or `complete` |
| `GOOGLE_CSE_ID` | No | — | Custom Search Engine ID |
| `LOG_LEVEL` | No | `info` | `debug`, `info`, `warn`, `error` |

### Docker Compose

```bash
cp .env.example .env
# Edit .env with your credentials
docker compose up --build
```

To enable persistent auth in Docker Compose, add `WORKSPACE_MCP_PERSISTENT_AUTH=true` and mount a volume for `/data/credentials`. See `docker-compose.yml` for the full configuration.

---

## How Authentication Works

```
1. AI agent calls start_google_auth with the user's email
2. Server returns a Google OAuth consent URL
3. User opens the URL in their browser and grants access
4. Google redirects to http://localhost:<port>/oauth/callback
5. Server exchanges the code for a token and stores it
6. All subsequent tool calls use the stored token (auto-refreshes)
```

The OAuth callback URL is built from the port you pass to `start.sh`. If you use `--port 9000`, the callback goes to `http://localhost:9000/oauth/callback`. Make sure the redirect URI in your [Google Cloud Console](https://console.cloud.google.com/apis/credentials) matches the port you're using.

By default, tokens are stored **in memory only** and lost on restart. With `--persistent-auth`, tokens are stored per-user at `/data/credentials/` (in Docker) or `~/.google_workspace_mcp/credentials/` (local). Token files have `0600` permissions. The credentials directory has `0700` permissions. In Docker, a named volume (`mcp-credentials`) is mounted automatically to persist tokens across restarts.

---

## Service-Specific Notes

### Google Chat

Requires a Google Workspace account. The Chat API does **not** work with consumer Gmail accounts. The app must be configured as a Chat app in the Workspace Admin Console.

### Custom Search

Requires a Custom Search Engine ID (`GOOGLE_CSE_ID`), created at [programmablesearchengine.google.com](https://programmablesearchengine.google.com). Pass it with `--cse-id` or the `GOOGLE_CSE_ID` env var.

### Apps Script

`run_script_function` only works with scripts deployed as an **API executable**. The user must have **edit access** to the script project. Rate limit: ~30 calls/min.

### Contacts

Uses the Google People API. The legacy Contacts API is deprecated. Tool names use "contacts" for clarity.

---

## Development

### Build from Source

```bash
go build -o server ./cmd/server

# Run locally in stdio mode
export GOOGLE_OAUTH_CLIENT_ID="your-client-id"
export GOOGLE_OAUTH_CLIENT_SECRET="your-secret"
./server

# Run locally in HTTP mode
./server --transport streamable-http
```

### Run Tests

```bash
# Unit tests
go test ./...

# Integration tests (verifies all 136 tools register correctly)
GOOGLE_OAUTH_CLIENT_ID=test GOOGLE_OAUTH_CLIENT_SECRET=test \
  go test -tags=integration ./internal/integration/

# Race detection
go test -race ./...
```

### Lint

```bash
golangci-lint run
```

### Project Structure

```
cmd/server/main.go              Entry point, signal handling, transport routing
internal/
  auth/                         OAuth2 flow, token storage, callback handler, scopes
  config/                       Env var + CLI flag loading, tier config
  registry/registry.go          Tool registration with service/tier/mode filtering
  services/factory.go           Google API client factory (12 APIs, per-user caching)
  tools/                        One package per Google Workspace service
    gmail/  drive/  calendar/   ... (handlers, helpers, registration)
    comments/                   Shared comment tools for Docs/Sheets/Slides
  middleware/                   SDK middleware (logging, error translation, retry)
  pkg/                          Shared utilities (response builder, HTML-to-text, Office extraction)
configs/tool_tiers.yaml         Tier assignments for all 136 tools
```

---

## MCP Spec Compliance

Targeting MCP spec **2025-11-25**:

| Feature | Status |
|---------|--------|
| Tools (136 + auth) | Implemented |
| Tool Annotations | Implemented |
| Structured Output | Implemented |
| Progress Notifications | Implemented (batch tools) |
| Tool Icons | Implemented (per-service) |
| SDK Middleware | Implemented |
| Resources / Prompts | Deferred to v2 |

## License

See [LICENSE](LICENSE) for details.
