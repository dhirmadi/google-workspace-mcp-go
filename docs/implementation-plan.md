# Implementation Plan

## Principles

- **Test alongside code, not after**: Write `_test.go` files alongside every handler and helper. Phase 5 is for integration tests only.
- **Front-load infrastructure**: Auth, factory, and registry bugs block everything. Get these solid first.
- **Shared patterns early**: Build the shared comment tools in Phase 2 to validate the DRY pattern before scaling out.
- **Lint from day one**: Set up `golangci-lint` in Phase 1 to catch issues immediately.

---

## Phase 1: Foundation (Week 1)

Core infrastructure that everything else depends on.

- [ ] Initialize Go module (`go.mod`) — Go 1.24, pin `go-sdk` v1.2.0, `google.golang.org/api` v0.262.0, `oauth2` v0.34.0, `yaml.v3` v3.0.1
- [ ] `.golangci.yml` — linter configuration (set up early, not Phase 5)
- [ ] `internal/config/config.go` — env var loading, CLI flag parsing, `GOOGLE_OAUTH_CLIENT_ID` naming
- [ ] `internal/config/tiers.go` — YAML tier config loading
- [ ] `internal/auth/scopes.go` — minimal scope sets (full + read-only, no redundant scopes)
- [ ] `internal/auth/oauth.go` — OAuth2 flow handling
- [ ] `internal/auth/credentials.go` — local credential storage with `persistingTokenSource` (0700/0600 perms)
- [ ] `internal/auth/callback.go` — OAuth callback HTTP server for stdio mode
- [ ] `internal/services/factory.go` — service client factory with per-user caching + `ReuseTokenSource`
- [ ] `internal/pkg/response/builder.go` — response string builder + tests
- [ ] `internal/middleware/logging.go` — SDK middleware: `AddSendingMiddleware`/`AddReceivingMiddleware`
- [ ] `internal/middleware/errors.go` — Agent-actionable Google API error translation
- [ ] `internal/middleware/retry.go` — exponential backoff for 429s
- [ ] `internal/registry/registry.go` — tool registry with tier/annotations/service filtering, SEP-986 name validation
- [ ] `cmd/server/main.go` — signal handling, stdio + streamable-http transport, SDK middleware wiring
- [ ] `configs/tool_tiers.yaml` — tier config (updated with tier promotions)
- [ ] `Dockerfile` — Go 1.24, multi-stage build, BuildKit cache mounts, distroless runtime
- [ ] `.dockerignore`

**Exit criteria**: Server starts, loads config, initializes factory, runs over both stdio and HTTP with zero tools registered. Graceful shutdown works. Linter passes.

## Phase 2: Core Services + Shared Patterns (Week 1–2)

First batch of tools — the `core` tier across the four most-used services, plus the shared comment pattern.

- [ ] `internal/tools/comments/comments.go` — shared comment tools (read/create/reply/resolve) with full annotations
- [ ] `internal/tools/comments/comments_test.go` — tests for comment handlers
- [ ] `internal/tools/gmail/gmail.go` — registration (4 core tools) with `ReadOnlyHint`, `OpenWorldHint`
- [ ] `internal/tools/gmail/handlers.go` — search (dual output), get, batch get (progress notifications), send
- [ ] `internal/tools/gmail/helpers.go` — HTML-to-text, MIME parsing, email building
- [ ] `internal/tools/gmail/helpers_test.go` — tests for helpers
- [ ] `internal/pkg/htmlutil/htmlutil.go` — HTML to plain text conversion + tests
- [ ] `internal/tools/drive/drive.go` — registration (7 core tools) with full annotations
- [ ] `internal/tools/drive/handlers.go` — search (dual output), get content, download URL, create, import, share, shareable link
- [ ] `internal/tools/drive/helpers.go` — permission formatting, Office XML extraction
- [ ] `internal/tools/drive/helpers_test.go`
- [ ] `internal/pkg/office/extract.go` — .docx/.xlsx/.pptx text extraction + tests
- [ ] `internal/tools/calendar/calendar.go` — registration (5 core tools, includes delete_event)
- [ ] `internal/tools/calendar/handlers.go` — list (dual output), get events (dual output), create, modify (`IdempotentHint`), delete (`DestructiveHint`)
- [ ] `internal/tools/calendar/helpers.go` — reminder parsing, attendee formatting + tests
- [ ] `internal/tools/sheets/sheets.go` — registration (3 core tools)
- [ ] `internal/tools/sheets/handlers.go` — create spreadsheet, read values (dual output), modify values

**Exit criteria**: 19 core tools + shared comment pattern working end-to-end with correct annotations.

## Phase 3: Extended Services (Week 2)

Expand coverage: core tools for remaining services + extended tools for Phase 2 services.

- [ ] `internal/tools/docs/docs.go` — registration (3 core + 6 extended + comments)
- [ ] `internal/tools/docs/handlers.go`
- [ ] `internal/tools/tasks/tasks.go` — registration (5 core incl. list_task_lists + 1 extended)
- [ ] `internal/tools/tasks/handlers.go` — delete (`DestructiveHint`), update (`IdempotentHint`)
- [ ] `internal/tools/tasks/helpers.go` + tests
- [ ] `internal/tools/contacts/contacts.go` — registration (4 core + 4 extended)
- [ ] `internal/tools/contacts/handlers.go` — delete (`DestructiveHint`)
- [ ] `internal/tools/contacts/helpers.go` + tests
- [ ] `internal/tools/chat/chat.go` — registration (4 core incl. list_chat_spaces)
- [ ] `internal/tools/chat/handlers.go`
- [ ] Gmail extended tools (9 more handlers, batch tools with progress notifications)
- [ ] Drive extended tools (7 more handlers)
- [ ] Calendar extended tools (1: query_freebusy)
- [ ] Sheets extended tools (6 more handlers + comments)

**Exit criteria**: All core and extended tools implemented. ~95 tools operational. All annotations set correctly.

## Phase 4: Complete Coverage (Week 3)

All remaining tools across all services.

- [ ] `internal/tools/forms/forms.go` — registration (2 core + 1 extended + 3 complete)
- [ ] `internal/tools/forms/handlers.go`
- [ ] `internal/tools/slides/slides.go` — registration (2 core + 3 extended + comments)
- [ ] `internal/tools/slides/handlers.go`
- [ ] `internal/tools/search/search.go` — registration (1 core + 1 extended + 1 complete), wire `GOOGLE_CSE_ID`
- [ ] `internal/tools/search/handlers.go`
- [ ] `internal/tools/appscript/appscript.go` — registration (7 core + 10 extended)
- [ ] `internal/tools/appscript/handlers.go`
- [ ] `internal/tools/appscript/helpers.go` — trigger code generation + tests
- [ ] `internal/tools/auth/auth.go` — start_google_auth tool (legacy OAuth 2.0)
- [ ] All remaining complete-tier tools across Gmail, Drive, Docs, Sheets, Tasks, Contacts

**Exit criteria**: All 136 tools implemented. Feature parity with Python version.

## Phase 5: Production Readiness (Week 3)

Polish, integration testing, and CI/CD.

- [ ] Docker configuration finalized and tested end-to-end
- [ ] `docker-compose.yml` for local development
- [ ] `README.md` with setup, usage, configuration, security notes, and service-specific limitations
- [ ] Integration tests (build-tagged `//go:build integration`) for core tool flows
- [ ] CI/CD setup (GitHub Actions: lint, test, build, docker push)
- [ ] Tool icon metadata per service for MCP client UX
- [ ] Verify all `ToolAnnotations` values (ReadOnly, Destructive, Idempotent, OpenWorld)
- [ ] Verify tier assignments in `tool_tiers.yaml` match `tools-inventory.md`
- [ ] Security review: credential permissions, log redaction, input bounds
- [ ] Reconcile final tool count and document definitively
