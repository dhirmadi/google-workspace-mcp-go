# Architecture

## Overview

The Google Workspace MCP server is a Go 1.24 application that exposes 136 tools across 12 Google Workspace services via the Model Context Protocol (targeting spec **2025-11-25**). It runs as a container and communicates over stdio (default) or streamable HTTP.

```
┌─────────────────────────────────────────────────────┐
│                   Tech Stack                         │
├─────────────────────────────────────────────────────┤
│  Go:        1.24 (stable, Feb 2025)                 │
│  MCP Spec:  2025-11-25                              │
│  Framework: github.com/modelcontextprotocol/go-sdk  │
│  OAuth:     go-sdk/auth + golang.org/x/oauth2       │
│  Google:    google.golang.org/api/...               │
│  Config:    gopkg.in/yaml.v3                        │
│  Logging:   log/slog (std lib)                      │
└─────────────────────────────────────────────────────┘
```

> Input validation is handled by the MCP SDK's built-in `jsonschema` support (via `github.com/google/jsonschema-go` internally). No external validation library needed.

## MCP Spec Compliance (2025-11-25)

### Primitives

| Primitive | Status | Notes |
|-----------|--------|-------|
| **Tools** | Implemented | 136 tools across 12 services |
| **Resources** | Deferred to v2 | See [Resources & Prompts](#resources--prompts-deferred-to-v2) |
| **Prompts** | Deferred to v2 | See [Resources & Prompts](#resources--prompts-deferred-to-v2) |

### Spec Features

| Feature | Status | Notes |
|---------|--------|-------|
| Tool annotations (`ReadOnlyHint`, `DestructiveHint`, `IdempotentHint`, `OpenWorldHint`) | Implemented | All 136 tools annotated |
| Structured output (`OutputSchema` / `structuredContent`) | Partial | Data-returning tools provide dual output (text + typed) |
| Elicitation (Form + URL mode) | Planned v1.1 | See [Elicitation](#elicitation) |
| Progress notifications | Implemented | For batch/long-running tools |
| Tool icons | Implemented | Per-service icons for MCP client UX |
| SDK middleware hooks | Implemented | `AddSendingMiddleware` / `AddReceivingMiddleware` |
| OAuth 2.1 / CIMD | Planned | See `auth-and-scopes.md` |

### Tool Naming

All tool names comply with MCP SEP-986: `^[a-zA-Z0-9_-]{1,64}$`, using `snake_case`. This regex is enforced as a validation rule in the registry.

---

## Project Structure

```
google-workspace-mcp-go/
├── cmd/server/main.go              # Entry point (signal handling, transport selection)
├── internal/
│   ├── auth/                       # OAuth2 flow, credentials, scopes, callback
│   ├── config/                     # Env var loading, tier config
│   ├── registry/registry.go        # Tool filtering by tier, annotations, services
│   ├── services/factory.go         # Google service client factory (12 APIs)
│   ├── tools/                      # One sub-package per Google Workspace service
│   │   ├── comments/comments.go    # SHARED comment tools (Docs, Sheets, Slides via Drive)
│   │   ├── auth/auth.go            # start_google_auth tool (legacy OAuth 2.0)
│   │   ├── gmail/ drive/ calendar/ docs/ sheets/ chat/ forms/ slides/ tasks/ contacts/ search/ appscript/
│   ├── middleware/
│   │   ├── logging.go              # SDK middleware: AddSendingMiddleware/AddReceivingMiddleware
│   │   ├── errors.go               # Agent-actionable error translation
│   │   └── retry.go                # Exponential backoff for 429s
│   └── pkg/
│       ├── response/builder.go     # Response string builder (DRY)
│       ├── format/format.go        # Common formatting utilities
│       ├── office/extract.go       # Office XML text extraction
│       └── htmlutil/htmlutil.go    # HTML to plain text
├── configs/tool_tiers.yaml
├── docs/
├── Dockerfile
└── README.md
```

## Core Components

### 1. Service Factory

Manages authenticated Google API clients per user email. Clients are cached with `oauth2.ReuseTokenSource` for concurrency-safe auto-refresh. Refreshed tokens are persisted to disk via `persistingTokenSource`.

### 2. Tool Registration

Every service package exposes `Register(server *mcp.Server, factory *services.Factory)`. Tools declare full `ToolAnnotations` (read-only, destructive, idempotent, open-world hints). Handlers are closures over the factory.

### 3. Dual Output Strategy

- **Text output** (default): Human-readable via response builder, returned as `mcp.TextContent`
- **Structured output** (data-returning tools): Typed output struct auto-serialized as `structuredContent` alongside text. Used for search results, event lists, file metadata — any data AI agents parse programmatically.

### 4. SDK Middleware Integration

Logging and error handling use the SDK's built-in middleware hooks:

```go
server.AddSendingMiddleware(middleware.LoggingMiddleware)
server.AddReceivingMiddleware(middleware.ErrorMiddleware)
```

This integrates with the MCP protocol layer rather than wrapping handlers individually.

### 5. Progress Notifications

Long-running operations (batch tools, large searches, attachment processing) report progress via the SDK:

```go
req.Session.NotifyProgress(ctx, &mcp.ProgressNotificationParams{
    ProgressToken: req.Params.Meta.ProgressToken,
    Progress:      float64(i),
    Total:         float64(total),
    Message:       fmt.Sprintf("Processing %d/%d", i, total),
})
```

### 6. Tool Registry

Filters registered tools based on: tier, enabled services, `ReadOnlyHint`, and OAuth mode.

### 7. Shared Comment Tools

Comments on Docs, Sheets, and Slides all use the Drive API. A single `comments` package registers 4 tools parameterized by resource type.

### Elicitation

The MCP 2025-11-25 spec introduced **Elicitation** — servers can request structured input from users during tool execution. The Go SDK v1.2.0 supports this via `ElicitRequest`, `ElicitResult`, and `URLElicitationRequiredError`.

Two modes:
- **Form Mode**: Direct structured data requests with JSON schema validation
- **URL Mode**: Out-of-band interaction for sensitive data and OAuth flows

Use cases for this server:
- Account selection during multi-user OAuth flow
- Confirmation before destructive operations (delete events, contacts, etc.)
- OAuth consent URL presentation

**Status**: Planned for v1.1. The architecture supports it — Elicitation will be added to destructive tools and the auth flow after core tool implementation is complete.

### Resources & Prompts (Deferred to v2)

The MCP spec defines three server primitives: **Tools** (model-controlled), **Resources** (application-controlled context), and **Prompts** (user-controlled templates). This server implements Tools only for v1.

Compelling future use cases:
- **Resources**: Expose Drive files, calendar events, or contacts as MCP resources that clients can attach to context
- **Prompts**: Pre-built templates like "summarize this email thread" or "draft a reply to this message"

These are deferred because the tool surface alone (136 tools) provides full Google Workspace coverage, and Resources/Prompts would require additional state management and caching patterns. They will be considered for v2 based on user feedback.

## Transport Modes

| Transport | Description | Flag |
|-----------|-------------|------|
| `stdio` | Standard I/O (default, for MCP client integration) | `--transport stdio` |
| `streamable-http` | HTTP with streamable responses (`mcp.NewStreamableHTTPHandler`) | `--transport streamable-http` |

## Dependencies

```
Go                                      1.24
github.com/modelcontextprotocol/go-sdk  v1.2.0
golang.org/x/oauth2                     v0.34.0
google.golang.org/api                   v0.262.0
gopkg.in/yaml.v3                        v3.0.1
```

> Pin exact versions in `go.mod`. Run `go get -u` to check for newer releases at project init time. The `google.golang.org/api` module releases frequently — use the latest available.

---

## Known Limitations & Service Notes

### Google Chat API — Workspace Only

The Chat API **requires the app to be configured as a Chat app** in the Google Workspace Admin Console. It does **NOT** work with consumer Gmail accounts. Users on consumer accounts will receive 403 errors.

### Custom Search (CSE) — Requires Engine ID

CSE requires a **Custom Search Engine ID** (`cx`) via `GOOGLE_CSE_ID` env var, created at [programmablesearchengine.google.com](https://programmablesearchengine.google.com).

### Apps Script — `run_script_function` Constraints

- Only works if the script is **deployed as an API executable**
- User must have **edit access** to the script project
- Rate limits: ~30 calls/min

### People API (not "Contacts API")

The Go client is `google.golang.org/api/people/v1`. The legacy Contacts API is fully deprecated. Tool names use "contacts" for user-facing clarity.
