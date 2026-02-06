# Authentication & OAuth Scopes

## Auth Architecture

The server supports two OAuth modes:

1. **Legacy OAuth 2.0** (default): Server manages tokens locally in a credentials directory. Users authenticate via the `start_google_auth` tool which launches a browser-based OAuth flow.
2. **OAuth 2.1** (`MCP_ENABLE_OAUTH21=true`): The MCP client handles authentication. The `start_google_auth` tool is disabled.

### MCP 2025-11-25 Authorization Updates

The MCP spec 2025-11-25 overhauled the authorization model significantly:

- **Client ID Metadata Documents (CIMD)**: Now the default client registration method, replacing Dynamic Client Registration (DCR). When OAuth 2.1 is enabled, the server should support CIMD-based discovery.
- **OpenID Connect Discovery 1.0**: Authorization server discovery via `/.well-known/openid-configuration`.
- **Incremental scope consent**: The server can request additional scopes via `WWW-Authenticate` headers when a tool requires a scope the user hasn't yet granted.

**v1 implementation**: The legacy OAuth 2.0 mode handles auth internally. When `MCP_ENABLE_OAUTH21=true`, the server delegates to the MCP client but does not yet implement CIMD or OpenID Connect Discovery. Full spec-compliant OAuth 2.1 (CIMD + Discovery) is planned for v1.1.

### Credential Storage

- Default directory: `~/.google_workspace_mcp/credentials` (created with `0700` permissions)
- Override: `WORKSPACE_MCP_CREDENTIALS_DIR` env var
- Tokens stored per user email as JSON files (`0600` permissions)
- Automatic token refresh via `oauth2.ReuseTokenSource` (concurrency-safe)
- Refreshed tokens persisted to disk automatically (see `code-patterns.md`)
- See `security.md` for token storage security considerations

### OAuth Callback (stdio mode)

When running in stdio mode, the server starts a temporary local HTTP server to handle the OAuth callback redirect. Implemented in `internal/auth/callback.go`.

## Base Scopes (Always Required)

```go
var BaseScopes = []string{
    "https://www.googleapis.com/auth/userinfo.email",
    "https://www.googleapis.com/auth/userinfo.profile",
    "openid",
}
```

## Service Scopes (Full Access — Minimum Required)

Request only the **minimum scopes** needed. Broader scopes already imply narrower ones. Over-requesting makes the consent screen intimidating and increases the chance of rejection by Google's OAuth verification team.

### Gmail
```
https://www.googleapis.com/auth/gmail.modify
https://www.googleapis.com/auth/gmail.send
https://www.googleapis.com/auth/gmail.labels
https://www.googleapis.com/auth/gmail.settings.basic
```
> `gmail.modify` already implies `gmail.readonly`. `gmail.compose` is implied by `gmail.send` + `gmail.modify`.

### Drive
```
https://www.googleapis.com/auth/drive
```
> `drive` already implies `drive.readonly` and `drive.file`. No need to request all three.

### Calendar
```
https://www.googleapis.com/auth/calendar
```
> `calendar` already implies `calendar.readonly` and `calendar.events`.

### Docs
```
https://www.googleapis.com/auth/documents
```
> `documents` implies `documents.readonly`.

### Sheets
```
https://www.googleapis.com/auth/spreadsheets
```
> `spreadsheets` implies `spreadsheets.readonly`.

### Chat
```
https://www.googleapis.com/auth/chat.messages
https://www.googleapis.com/auth/chat.spaces
```
> `chat.messages` implies `chat.messages.readonly`.

### Forms
```
https://www.googleapis.com/auth/forms.body
https://www.googleapis.com/auth/forms.responses.readonly
```
> `forms.body` implies `forms.body.readonly`.

### Slides
```
https://www.googleapis.com/auth/presentations
```
> `presentations` implies `presentations.readonly`.

### Tasks
```
https://www.googleapis.com/auth/tasks
```
> `tasks` implies `tasks.readonly`.

### Contacts (People API)
```
https://www.googleapis.com/auth/contacts
```
> `contacts` implies `contacts.readonly`.

### Search (CSE)
```
https://www.googleapis.com/auth/cse
```

### Apps Script
```
https://www.googleapis.com/auth/script.projects
https://www.googleapis.com/auth/script.deployments
https://www.googleapis.com/auth/script.processes
https://www.googleapis.com/auth/script.metrics
https://www.googleapis.com/auth/drive.file
```
> `script.projects` implies `script.projects.readonly`. `script.deployments` implies `script.deployments.readonly`.

## Read-Only Scopes

When `--read-only` is set, each service requests only these scopes:

| Service | Read-Only Scopes |
|---------|-----------------|
| Gmail | `gmail.readonly` |
| Drive | `drive.readonly` |
| Calendar | `calendar.readonly` |
| Docs | `documents.readonly` |
| Sheets | `spreadsheets.readonly` |
| Chat | `chat.messages.readonly`, `chat.spaces.readonly` |
| Forms | `forms.body.readonly`, `forms.responses.readonly` |
| Slides | `presentations.readonly` |
| Tasks | `tasks.readonly` |
| Contacts | `contacts.readonly` |
| Search | `cse` |
| Apps Script | `script.projects.readonly`, `script.deployments.readonly`, `script.processes`, `script.metrics`, `drive.readonly` |
