# Code Patterns

## Module Path

```
github.com/evert/google-workspace-mcp-go
```

Use this consistently across all imports. Never use `<owner>` placeholders.

---

## Output Strategy: Dual Output

### Default: Text + Structured for Data-Returning Tools

In 2026, MCP clients increasingly rely on structured output (`structuredContent`) for reliable data extraction. The SDK auto-generates an `OutputSchema` for typed output structs. Use **dual output** for tools that return parseable data (search results, event lists, file metadata):

```go
func createSearchMessagesHandler(factory *services.Factory) func(
    context.Context, *mcp.CallToolRequest, SearchMessagesInput,
) (*mcp.CallToolResult, any, error) {
    return func(ctx context.Context, req *mcp.CallToolRequest, input SearchMessagesInput) (
        *mcp.CallToolResult, any, error,
    ) {
        // ... API call ...

        // Build text output for display
        rb := response.New()
        rb.Header("Gmail Search Results")
        rb.KeyValue("Query", input.Query)
        rb.KeyValue("Results", len(messages))
        rb.Blank()
        for _, msg := range messages {
            rb.Item("Subject: %s", msg.Subject)
            rb.Line("  From: %s | Date: %s", msg.From, msg.Date)
        }

        // Return BOTH text content and typed output
        return &mcp.CallToolResult{
            Content: []mcp.Content{&mcp.TextContent{Text: rb.Build()}},
        }, SearchMessagesOutput{Messages: summaries, Query: input.Query}, nil
    }
}
```

### Text-Only: For Action Tools

Tools that perform actions (send, create, delete, modify) return text-only confirmation. Use `any` as the output type to suppress `OutputSchema` generation:

```go
func createSendMessageHandler(factory *services.Factory) func(
    context.Context, *mcp.CallToolRequest, SendMessageInput,
) (*mcp.CallToolResult, any, error) {
    return func(ctx context.Context, req *mcp.CallToolRequest, input SendMessageInput) (
        *mcp.CallToolResult, any, error,
    ) {
        // ... send email ...
        rb := response.New()
        rb.Header("Message Sent")
        rb.KeyValue("To", input.To)
        rb.KeyValue("Message ID", sentMsg.Id)
        return &mcp.CallToolResult{
            Content: []mcp.Content{&mcp.TextContent{Text: rb.Build()}},
        }, nil, nil  // nil typed output, `any` type = no OutputSchema
    }
}
```

> When the output type is `any`, the SDK does not generate an `OutputSchema`. Use this for text-only tools.

---

## Tool Handler Template

```go
package gmail

import (
    "context"
    "fmt"

    "github.com/modelcontextprotocol/go-sdk/mcp"
    "google.golang.org/api/gmail/v1"

    "github.com/evert/google-workspace-mcp-go/internal/middleware"
    "github.com/evert/google-workspace-mcp-go/internal/pkg/response"
    "github.com/evert/google-workspace-mcp-go/internal/services"
)

// 1. Input struct — json + jsonschema tags on every field
type SearchMessagesInput struct {
    UserEmail string `json:"user_google_email" jsonschema:"required,description=The user's Google email address"`
    Query     string `json:"query" jsonschema:"required,description=Gmail search query"`
    PageSize  int    `json:"page_size,omitempty" jsonschema:"description=Max results (default 10)"`
    PageToken string `json:"page_token,omitempty" jsonschema:"description=Token for next page of results"`
}

// 2. Output struct for data-returning tools (typed output = OutputSchema generated)
type SearchMessagesOutput struct {
    Messages []MessageSummary `json:"messages"`
    Query    string           `json:"query"`
}

// 3. Handler factory — closure over services.Factory
//    Use `any` output type for text-only tools, typed struct for data-returning tools
func createSearchMessagesHandler(factory *services.Factory) func(
    context.Context, *mcp.CallToolRequest, SearchMessagesInput,
) (*mcp.CallToolResult, any, error) {
    return func(ctx context.Context, req *mcp.CallToolRequest, input SearchMessagesInput) (
        *mcp.CallToolResult, any, error,
    ) {
        if input.PageSize == 0 {
            input.PageSize = 10
        }

        srv, err := factory.Gmail(ctx, input.UserEmail)
        if err != nil {
            return nil, nil, middleware.HandleGoogleAPIError(err)
        }

        result, err := srv.Users.Messages.List(input.UserEmail).
            Q(input.Query).
            MaxResults(int64(input.PageSize)).
            PageToken(input.PageToken).
            Context(ctx).
            Do()
        if err != nil {
            return nil, nil, middleware.HandleGoogleAPIError(err)
        }

        // Build text output
        rb := response.New()
        rb.Header("Search Results for: %s", input.Query)
        rb.KeyValue("Total results", len(result.Messages))
        if result.NextPageToken != "" {
            rb.KeyValue("Next page token", result.NextPageToken)
        }
        rb.Blank()
        for _, msg := range result.Messages {
            rb.Item("ID: %s (Thread: %s)", msg.Id, msg.ThreadId)
        }

        return &mcp.CallToolResult{
            Content: []mcp.Content{&mcp.TextContent{Text: rb.Build()}},
        }, SearchMessagesOutput{Messages: summaries, Query: input.Query}, nil
    }
}

// 4. Register tools — full ToolAnnotations + optional Icon
func Register(server *mcp.Server, factory *services.Factory) {
    mcp.AddTool(server, &mcp.Tool{
        Name:        "search_gmail_messages",
        Description: "Search Gmail messages using query syntax",
        Annotations: &mcp.ToolAnnotations{
            ReadOnlyHint: ptrBool(true),
            OpenWorldHint: ptrBool(true),
        },
    }, createSearchMessagesHandler(factory))

    mcp.AddTool(server, &mcp.Tool{
        Name:        "send_gmail_message",
        Description: "Send an email message",
        Annotations: &mcp.ToolAnnotations{
            ReadOnlyHint:  ptrBool(false),
            OpenWorldHint: ptrBool(true),
        },
    }, createSendMessageHandler(factory))

    mcp.AddTool(server, &mcp.Tool{
        Name:        "delete_gmail_filter",
        Description: "Delete an email filter",
        Annotations: &mcp.ToolAnnotations{
            ReadOnlyHint:    ptrBool(false),
            DestructiveHint: ptrBool(true),
        },
    }, createDeleteFilterHandler(factory))
}

func ptrBool(b bool) *bool { return &b }
```

### Key Rules

- **Input structs**: Every field has `json` + `jsonschema` tags. Required fields use `jsonschema:"required,..."`.
- **`user_google_email`**: Always the first required field in every input struct.
- **Handler factory**: Returns a closure that captures `factory`. No global state.
- **Output type**: Use `any` for text-only tools, typed struct for data-returning tools.
- **Errors**: Always pass through `middleware.HandleGoogleAPIError(err)`.
- **Context**: Always pass `ctx` to API calls: `.Context(ctx).Do()`.
- **Annotations**: Set all relevant hints — `ReadOnlyHint`, `DestructiveHint`, `IdempotentHint`, `OpenWorldHint`.

### Annotation Guide

| Hint | When to set `true` | Examples |
|------|-------------------|----------|
| `ReadOnlyHint` | Tool only reads data | `search_*`, `get_*`, `list_*` |
| `DestructiveHint` | Tool deletes or permanently removes data | `delete_*`, `remove_*`, `clear_*` |
| `IdempotentHint` | Calling the tool twice with same input has same effect | `modify_*`, `update_*`, `share_*` |
| `OpenWorldHint` | Tool interacts with external services (all Google API tools) | All tools in this server |

---

## Pagination Pattern

```go
type SearchInput struct {
    UserEmail string `json:"user_google_email" jsonschema:"required,description=The user's Google email address"`
    Query     string `json:"query" jsonschema:"required,description=Search query"`
    PageSize  int    `json:"page_size,omitempty" jsonschema:"description=Max results per page (default 10)"`
    PageToken string `json:"page_token,omitempty" jsonschema:"description=Token for next page of results"`
}
```

Always include `NextPageToken` in the output when present.

---

## Error Handling — Agent-Actionable Messages

Error messages must be **actionable for AI agents**, not end users. The agent needs to know what to do next:

```go
// internal/middleware/errors.go
func HandleGoogleAPIError(err error) error {
    if err == nil {
        return nil
    }

    var googleErr *googleapi.Error
    if errors.As(err, &googleErr) {
        switch googleErr.Code {
        case 401:
            return fmt.Errorf(
                "authentication expired for this user - call start_google_auth tool to re-authenticate, "+
                "or verify MCP_ENABLE_OAUTH21 is configured correctly")
        case 403:
            return fmt.Errorf(
                "permission denied - the required OAuth scope may not be granted. "+
                "Suggest user re-authenticate with broader scopes. Detail: %s", googleErr.Message)
        case 404:
            return fmt.Errorf("resource not found - verify the ID is correct and the user has access")
        case 429:
            return fmt.Errorf(
                "rate limit exceeded for this Google API - wait 30-60 seconds before retrying this tool call")
        default:
            return fmt.Errorf("Google API error (%d): %s", googleErr.Code, googleErr.Message)
        }
    }
    return err
}
```

### Rate Limiting & Retry

```go
// internal/middleware/retry.go
func WithRetry(ctx context.Context, maxAttempts int, fn func() error) error {
    var err error
    for attempt := 0; attempt < maxAttempts; attempt++ {
        err = fn()
        if err == nil {
            return nil
        }
        var googleErr *googleapi.Error
        if !errors.As(err, &googleErr) || googleErr.Code != 429 {
            return err
        }
        backoff := time.Duration(1<<uint(attempt)) * time.Second
        select {
        case <-time.After(backoff):
        case <-ctx.Done():
            return ctx.Err()
        }
    }
    return err
}
```

---

## SDK Middleware Integration

Use the SDK's built-in middleware hooks instead of bespoke wrappers:

```go
// cmd/server/main.go (after creating server)
server.AddSendingMiddleware(middleware.LoggingSendMiddleware(logger))
server.AddReceivingMiddleware(middleware.LoggingReceiveMiddleware(logger))
```

```go
// internal/middleware/logging.go
func LoggingSendMiddleware(logger *slog.Logger) mcp.SendingMiddleware {
    return func(ctx context.Context, msg *mcp.JSONRPCMessage, next mcp.SendFunc) error {
        logger.Info("sending", "method", msg.Method)
        return next(ctx, msg)
    }
}

func LoggingReceiveMiddleware(logger *slog.Logger) mcp.ReceivingMiddleware {
    return func(ctx context.Context, msg *mcp.JSONRPCMessage, next mcp.ReceiveFunc) error {
        logger.Info("received", "method", msg.Method)
        return next(ctx, msg)
    }
}
```

---

## Progress Notifications

For batch and long-running tools:

```go
func createBatchGetMessagesHandler(factory *services.Factory) func(
    context.Context, *mcp.CallToolRequest, BatchGetInput,
) (*mcp.CallToolResult, any, error) {
    return func(ctx context.Context, req *mcp.CallToolRequest, input BatchGetInput) (
        *mcp.CallToolResult, any, error,
    ) {
        total := len(input.MessageIDs)
        for i, id := range input.MessageIDs {
            // Report progress if client supports it
            if req.Params.Meta.ProgressToken != nil {
                req.Session.NotifyProgress(ctx, &mcp.ProgressNotificationParams{
                    ProgressToken: req.Params.Meta.ProgressToken,
                    Progress:      float64(i),
                    Total:         float64(total),
                    Message:       fmt.Sprintf("Fetching message %d/%d", i+1, total),
                })
            }
            // ... fetch message ...
        }
        // ... build response ...
    }
}
```

---

## Token Refresh & Persistence

```go
// internal/auth/credentials.go

type persistingTokenSource struct {
    base      oauth2.TokenSource
    store     TokenStore
    userEmail string
}

func (p *persistingTokenSource) Token() (*oauth2.Token, error) {
    token, err := p.base.Token()
    if err != nil {
        return nil, err
    }
    if err := p.store.Save(p.userEmail, token); err != nil {
        slog.Warn("failed to persist refreshed token", "email", p.userEmail, "error", err)
    }
    return token, nil
}

func (f *Factory) clientFor(ctx context.Context, userEmail string) (*http.Client, error) {
    token, err := f.tokenStore.Load(userEmail)
    if err != nil {
        return nil, fmt.Errorf("no credentials for %s: %w", userEmail, err)
    }
    baseSource := f.oauth2Config.TokenSource(ctx, token)
    reuseSource := oauth2.ReuseTokenSource(token, &persistingTokenSource{
        base:      baseSource,
        store:     f.tokenStore,
        userEmail: userEmail,
    })
    return oauth2.NewClient(ctx, reuseSource), nil
}
```

---

## Shared Comment Tools (DRY Pattern)

```go
// internal/tools/comments/comments.go
func RegisterCommentTools(server *mcp.Server, factory *services.Factory, resourceType, fileIDParam string) {
    prefix := resourceType

    mcp.AddTool(server, &mcp.Tool{
        Name:        fmt.Sprintf("read_%s_comments", prefix),
        Description: fmt.Sprintf("Read all comments from a Google %s", capitalize(resourceType)),
        Annotations: &mcp.ToolAnnotations{ReadOnlyHint: ptrBool(true), OpenWorldHint: ptrBool(true)},
    }, createReadCommentsHandler(factory, resourceType, fileIDParam))

    mcp.AddTool(server, &mcp.Tool{
        Name:        fmt.Sprintf("create_%s_comment", prefix),
        Description: fmt.Sprintf("Create a comment on a Google %s", capitalize(resourceType)),
        Annotations: &mcp.ToolAnnotations{OpenWorldHint: ptrBool(true)},
    }, createCreateCommentHandler(factory, resourceType, fileIDParam))

    // ... reply and resolve similarly
}
```

---

## Main Entry Point

```go
// cmd/server/main.go
func main() {
    logger := slog.New(slog.NewJSONHandler(os.Stderr, nil))
    slog.SetDefault(logger)

    ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
    defer cancel()

    cfg, err := config.Load()
    if err != nil {
        slog.Error("failed to load config", "error", err)
        os.Exit(1)
    }

    factory, err := services.NewFactory(cfg.OAuth.ClientID, cfg.OAuth.ClientSecret)
    if err != nil {
        slog.Error("failed to create service factory", "error", err)
        os.Exit(1)
    }

    server := mcp.NewServer(&mcp.Implementation{
        Name:    "google-workspace-mcp",
        Version: "1.0.0",
    }, nil)

    // SDK middleware hooks
    server.AddSendingMiddleware(middleware.LoggingSendMiddleware(logger))
    server.AddReceivingMiddleware(middleware.LoggingReceiveMiddleware(logger))

    // Register all tools through the registry
    registry.RegisterAll(server, factory, cfg)

    slog.Info("starting Google Workspace MCP server", "transport", cfg.Server.Transport)

    switch cfg.Server.Transport {
    case "stdio":
        if err := server.Run(ctx, &mcp.StdioTransport{}); err != nil {
            slog.Error("server error", "error", err)
            os.Exit(1)
        }
    case "streamable-http":
        handler := mcp.NewStreamableHTTPHandler(
            func(r *http.Request) *mcp.Server { return server }, nil,
        )
        addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
        httpServer := &http.Server{Addr: addr, Handler: handler}
        go func() { <-ctx.Done(); httpServer.Shutdown(context.Background()) }()
        slog.Info("listening", "addr", addr)
        if err := httpServer.ListenAndServe(); err != http.ErrServerClosed {
            slog.Error("HTTP server error", "error", err)
            os.Exit(1)
        }
    default:
        slog.Error("unknown transport", "transport", cfg.Server.Transport)
        os.Exit(1)
    }
}
```
