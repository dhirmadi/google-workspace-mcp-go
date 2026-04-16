# Epic: Auth — PKCE & public (secret-less) client support

| Field | Value |
|-------|--------|
| **Status** | Done |
| **Horizon** | Now |
| **Priority** | P0 |
| **Upstream** | [v1.19.0 — SecretLess PKCE](https://github.com/taylorwilsdon/google_workspace_mcp/releases/tag/v1.19.0), [#677](https://github.com/taylorwilsdon/google_workspace_mcp/pull/677) |
| **Dependencies** | `docs/auth-and-scopes.md`; `./SEC-01-security-hygiene-upstream.md` may re-audit `internal/auth/credentials.go` after public-client paths land |

## Problem

Today the server requires `GOOGLE_OAUTH_CLIENT_SECRET` at config time (`internal/config/config.go`). Public / native-style OAuth clients use **PKCE** without a confidential secret. That blocks stricter secret hygiene, simpler local onboarding, and alignment with **OAuth 2.1 / MCP authorization** direction in `docs/auth-and-scopes.md`.

## Outcome

Operators can run the server with **only a client ID** (or empty secret where the stack allows), using **PKCE** for the authorization code flow, with **documented** Google Cloud Console redirect URI steps and limitations. Legacy mode (ID + secret) remains supported and tested.

## Scope

### In scope

- Config validation for confidential vs public client modes and PKCE verifier lifecycle.
- Auth URL generation and code exchange paths in `internal/auth/oauth.go` (and call sites).
- `start_google_auth` tool (`internal/tools/auth/auth.go`) end-to-end for public mode with local HTTP callback (`internal/auth/callback.go`).
- Operator docs: `README.md`, `docs/auth-and-scopes.md`, `docs/configuration.md`, `start.sh`, Docker env examples.

### Out of scope

- Full MCP OAuth 2.1 CIMD + OIDC discovery (see `docs/auth-and-scopes.md` v1.1 narrative).
- Domain-wide delegation (`./ENT-01-domain-wide-delegation-deferred.md`).

## Dependencies

- Google OAuth client type (Web vs Desktop) and redirect URI rules (documented, not guessed).
- `golang.org/x/oauth2` — PKCE via `oauth2.SetAuthURLParam` and verifier on `Exchange`.
- Registry: `internal/registry/registry.go` — `start_google_auth` remains hidden when `MCP_ENABLE_OAUTH21` disables legacy auth per existing pattern.

## Acceptance criteria (Gherkin)

### AC1: Confidential client still works ✅

```gherkin
Given GOOGLE_OAUTH_CLIENT_ID and GOOGLE_OAUTH_CLIENT_SECRET are set to non-empty values
And MCP_ENABLE_OAUTH21 is false
When the server loads configuration
Then configuration succeeds without error
And the OAuth config is treated as confidential (no PKCE requirement for mode selection)
```

_Covered by `TestAUTH01_ConfidentialClientMode` in `internal/config/config_test.go`._

### AC2: Public client mode is explicitly selectable ✅

```gherkin
Given GOOGLE_OAUTH_CLIENT_ID is set
And GOOGLE_OAUTH_PUBLIC_CLIENT=true
When the server loads configuration
Then configuration succeeds without requiring GOOGLE_OAUTH_CLIENT_SECRET
And Config.OAuth.PublicClient is true
```

_Covered by `TestAUTH01_PublicClientMode` in `internal/config/config_test.go`.
Env var: `GOOGLE_OAUTH_PUBLIC_CLIENT=true`._

### AC3: PKCE on authorization and exchange ✅

```gherkin
Given the server is in public client mode
When the legacy OAuth flow builds the authorization URL
Then the URL includes code_challenge and code_challenge_method=S256
When the callback receives an authorization code
Then the token exchange includes the PKCE code verifier associated with that auth request
```

_Covered by `TestAUTH01_PKCE_AuthURL_ContainsChallenge`, `TestAUTH01_PKCE_Exchange_SendsVerifier`,
`TestAUTH01_PKCE_VerifierClearedAfterExchange`, and `TestAUTH01_ConfidentialMode_NoPKCE`
in `internal/auth/oauth_test.go`._

### AC4: start_google_auth in public mode ✅

```gherkin
Given MCP_ENABLE_OAUTH21 is false
And the server is in public client mode with valid redirect configuration
When the MCP client invokes start_google_auth for a user email
Then the tool returns an authorization URL with PKCE parameters
```

_`start_google_auth` calls `OAuthManager.GetAuthURL`, which now appends PKCE params in public mode.
No change to the tool's handler or registration — behaviour flows through the OAuthManager._

### AC5: OAuth 2.1 mode unchanged ✅

```gherkin
Given MCP_ENABLE_OAUTH21 is true
When the server registers tools
Then start_google_auth is not registered
And no new code path forces legacy PKCE parameters when OAuth 2.1 is enabled
```

_Registry guard (`!cfg.EnableOAuth21`) in `internal/registry/registry.go` is unchanged.
PKCE is only applied inside `OAuthManager.GetAuthURL`, which is never called when OAuth 2.1 is active._

### AC6: Verification gates ✅

```gherkin
Given the implementation is complete
When CI runs golangci-lint and go test -race on ./...
Then both complete successfully
```

_`go test -race ./...` — all packages pass. `golangci-lint run` — clean on changed packages._

## Implementation notes

### MCP tool surface

- Tool: `start_google_auth` — existing name; annotations unchanged unless spec requires new hints.
- No new tools unless a separate explicitly named `*_pkce` helper is unavoidable (prefer extending existing flow).

### Packages and registration

- `internal/config/config.go` — validate OAuth mode; document env vars in `docs/configuration.md`.
- `internal/auth/oauth.go` — build auth URL and `Exchange` with conditional PKCE.
- `internal/auth/callback.go` — ensure callback server compatible with public redirect URIs used in docs.
- `internal/tools/auth/auth.go` — wire to factory/oauth entrypoints.
- `internal/registry/registry.go` — preserve OAuth21 filtering for `start_google_auth`.

### Auth and transport

- `cmd/server/main.go` — no regression to stdio / streamable HTTP startup when secret optional.
- Scopes: `internal/auth/scopes.go` unchanged unless Google public-client rules require different scope strings (unlikely).

### Verification

- `golangci-lint run ./...`, `go test -race ./...`; add `_test.go` for config matrix (public vs confidential) where pure logic allows.

## Component inventory

| Path | Change |
|------|--------|
| `internal/config/config.go` | Modified |
| `internal/auth/oauth.go` | Modified |
| `internal/auth/callback.go` | Modified (if redirect/callback assumptions change) |
| `internal/tools/auth/auth.go` | Modified |
| `README.md` | Modified |
| `docs/auth-and-scopes.md` | Modified |
| `docs/configuration.md` | Modified |
| `start.sh` | Modified |
| `docker-compose.yml` / `Dockerfile` docs | Modified if env contract changes |
| `.env.example` | Modified |

## Existing infrastructure to reuse

- `internal/auth/credentials.go` — `persistingTokenSource`, directory permissions (0700/0600) patterns; align with `./SEC-01-security-hygiene-upstream.md`.
- `internal/middleware/errors.go` — messages referencing `start_google_auth`; keep wording accurate for both modes.
- `internal/integration/registration_test.go` — extend env fixtures for new config branches where feasible.

## Risks

| Risk | Impact | Mitigation |
|------|--------|------------|
| Google rejects certain redirect URI shapes for public clients | Users cannot complete OAuth | Document one supported console setup; integration-test or script the happy path |
| Accidental regression disabling OAuth 2.1 path | Enterprise MCP clients break | AC5 + existing `internal/registry/registry.go` tests |
| PKCE verifier stored incorrectly across concurrent auth attempts | Intermittent exchange failures | Single-flight or per-user verifier map with clear lifecycle in `internal/auth/oauth.go` |
