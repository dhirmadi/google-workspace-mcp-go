# Security

## Credential Storage

### File Permissions

The credentials directory (`~/.google_workspace_mcp/credentials`) stores OAuth tokens as plain JSON files. The directory and its contents must be restricted:

```go
// internal/auth/credentials.go
os.MkdirAll(credDir, 0700)       // directory: owner only
os.WriteFile(path, data, 0600)   // files: owner read/write only
```

Enforce `0700` on the directory and `0600` on token files at creation time. Log a warning at startup if permissions are too open.

### Token Encryption at Rest

v1 stores tokens as plain JSON (matching the Python version's behavior). For production deployments handling sensitive Workspace data, consider:

- **v1.1**: AES-256-GCM encryption with a machine-derived key
- **v2**: Integration with OS keyring (macOS Keychain, Linux Secret Service) or cloud secret managers (GCP Secret Manager, Vault)

Document this limitation in the README.

## Input Sanitization

### Google API Query Injection

Gmail search queries, Drive search queries, and other string parameters are passed directly to Google APIs. While Google's APIs handle their own input validation, the server should:

1. **Limit query length**: Reject excessively long queries (>1024 chars)
2. **Log suspicious patterns**: Queries containing unusual operators or encoded content
3. **Never interpolate user input into API URLs** — always use SDK method parameters

### Parameter Validation

The MCP SDK validates input against the `jsonschema` tags automatically. Additional server-side validation:

- Email addresses: Basic format check before API calls
- File IDs: Alphanumeric + hyphen/underscore, reject obviously invalid IDs
- Page sizes: Enforce upper bounds (Google APIs have their own limits, but catch abuse early)

## Secret Protection

### Logging

OAuth client secrets and tokens must **never** appear in logs:

```go
// GOOD
slog.Info("authenticated user", "email", userEmail)

// BAD — token value in log
slog.Info("got token", "token", token.AccessToken)
```

The structured logging middleware should redact fields named `token`, `secret`, `password`, `access_token`, `refresh_token`.

### Environment Variables

- `GOOGLE_OAUTH_CLIENT_SECRET` should be treated as sensitive
- When running in containers, prefer secret mounts over env vars where possible
- Never log the full environment at startup

## Rate Limiting (Multi-User Mode)

In multi-user mode, the server handles requests for multiple Google accounts. Without rate limiting, a single user could exhaust Google API quotas that affect all users.

Implement per-user rate limiting:

- Track API call counts per user email per minute
- Apply Google's per-API rate limits (Gmail: 250 quota units/sec, Drive: 12,000 requests/min, etc.)
- Return a clear error when rate-limited: "Rate limit reached for this account — try again in N seconds"

## Transport Security

### stdio Mode

stdio communicates over process pipes — inherently local and not exposed to the network. No TLS needed.

### streamable-http Mode

For HTTP transport:
- **Always use TLS** in production (terminate at a reverse proxy if not handling directly)
- Set appropriate CORS headers
- Consider authentication on the HTTP endpoint itself (the MCP SDK's OAuth 2.1 support handles this when enabled)
