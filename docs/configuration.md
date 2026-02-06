# Configuration

## Environment Variables

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `GOOGLE_OAUTH_CLIENT_ID` | Yes | — | OAuth client ID |
| `GOOGLE_OAUTH_CLIENT_SECRET` | Yes | — | OAuth client secret |
| `GOOGLE_CSE_ID` | No* | — | Custom Search Engine ID (required for search tools) |
| `USER_GOOGLE_EMAIL` | No | — | Default email for single-user mode |
| `WORKSPACE_MCP_CREDENTIALS_DIR` | No | `~/.google_workspace_mcp/credentials` | Credential storage directory |
| `MCP_TRANSPORT` | No | `stdio` | Transport mode |
| `MCP_PORT` / `PORT` | No | `8000` | HTTP server port |
| `WORKSPACE_MCP_HOST` | No | `0.0.0.0` | HTTP bind address |
| `WORKSPACE_MCP_BASE_URI` | No | `http://localhost` | Base URI for OAuth callbacks |
| `MCP_SINGLE_USER_MODE` | No | `false` | Enable single-user mode |
| `MCP_ENABLE_OAUTH21` | No | `false` | Enable OAuth 2.1 mode |
| `WORKSPACE_MCP_STATELESS_MODE` | No | `false` | Stateless mode (requires OAuth 2.1) |
| `LOG_LEVEL` | No | `info` | Log verbosity |
| `TOOL_TIER` | No | `complete` | Default tool tier |

> **Naming**: Always use `GOOGLE_OAUTH_CLIENT_ID` / `GOOGLE_OAUTH_CLIENT_SECRET` — not `GOOGLE_CLIENT_ID` variants.

## CLI Flags

```
google-workspace-mcp-go [flags]

Flags:
  --transport string     Transport mode: stdio (default) or streamable-http
  --tools strings        Services to enable: gmail,drive,calendar,docs,sheets,
                         chat,forms,slides,tasks,contacts,search,appscript
  --tool-tier string     Load tools by tier: core, extended, or complete
  --single-user          Bypass session mapping, use any credentials
  --read-only            Request only read-only scopes, disable write tools
  --cli [command]        Direct tool invocation mode (no server)
```

CLI flags take precedence over environment variables.

## Transport Modes

| Transport | Description | Flag |
|-----------|-------------|------|
| `stdio` | Standard input/output (default) | `--transport stdio` |
| `streamable-http` | HTTP with streamable responses | `--transport streamable-http` |

## Tool Tiers

Tools are organized into tiers via `configs/tool_tiers.yaml`:

- **core** (44 tools): Essential tools for each service — the minimum for a useful integration
- **extended** (~51 tools): Additional commonly-used tools for power users
- **complete** (~41 tools): All tools including debug, batch, and administrative operations

The tier system is **cumulative**: `extended` includes all `core` tools; `complete` includes all `extended` and `core` tools.

### Tier Filtering Logic

The registry applies filters in this order:

1. Load tier config from `configs/tool_tiers.yaml`
2. Filter by `--tool-tier` (keep only tools at or below the selected tier)
3. Filter by `--tools` (keep only tools belonging to the listed services)
4. If `--read-only`, remove tools where `ToolAnnotations.ReadOnlyHint` is `false`
5. If OAuth 2.1 is enabled, remove `start_google_auth` tool

## Read-Only Mode

When `--read-only` is set:

1. Only request read-only OAuth scopes (from `ReadOnlyScopes` map in `internal/auth/scopes.go`)
2. Filter out tools with `ReadOnlyHint: false` in their `ToolAnnotations`
3. The registry uses the SDK's built-in `ToolAnnotations` — no separate read-only map needed

## Config Struct

```go
// internal/config/config.go
type Config struct {
    OAuth struct {
        ClientID     string // GOOGLE_OAUTH_CLIENT_ID
        ClientSecret string // GOOGLE_OAUTH_CLIENT_SECRET
        RedirectURL  string
    }
    Server struct {
        Transport string
        Port      int
        Host      string
        BaseURI   string
    }
    ToolTier        string
    EnabledServices []string
    ReadOnly        bool
    SingleUser      bool
    EnableOAuth21   bool
    StatelessMode   bool
    LogLevel        string
    CredentialsDir  string
    CSEID           string // GOOGLE_CSE_ID
    GoVersion       string // Build-time: Go 1.24
}
```
