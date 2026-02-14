package registry

import (
	"context"
	"fmt"
	"log/slog"
	"regexp"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/evert/google-workspace-mcp-go/internal/auth"
	"github.com/evert/google-workspace-mcp-go/internal/config"
	"github.com/evert/google-workspace-mcp-go/internal/services"
	"github.com/evert/google-workspace-mcp-go/internal/tools/appscript"
	authtools "github.com/evert/google-workspace-mcp-go/internal/tools/auth"
	"github.com/evert/google-workspace-mcp-go/internal/tools/calendar"
	"github.com/evert/google-workspace-mcp-go/internal/tools/chat"
	"github.com/evert/google-workspace-mcp-go/internal/tools/contacts"
	"github.com/evert/google-workspace-mcp-go/internal/tools/docs"
	"github.com/evert/google-workspace-mcp-go/internal/tools/drive"
	"github.com/evert/google-workspace-mcp-go/internal/tools/forms"
	"github.com/evert/google-workspace-mcp-go/internal/tools/gmail"
	"github.com/evert/google-workspace-mcp-go/internal/tools/search"
	"github.com/evert/google-workspace-mcp-go/internal/tools/sheets"
	"github.com/evert/google-workspace-mcp-go/internal/tools/slides"
	"github.com/evert/google-workspace-mcp-go/internal/tools/tasks"
)

// toolNameRE enforces SEP-986: tool names must match ^[a-zA-Z0-9_-]{1,64}$
var toolNameRE = regexp.MustCompile(`^[a-zA-Z0-9_-]{1,64}$`)

// ValidateToolName checks that a tool name complies with SEP-986.
func ValidateToolName(name string) error {
	if !toolNameRE.MatchString(name) {
		return fmt.Errorf("tool name %q does not match SEP-986 pattern ^[a-zA-Z0-9_-]{1,64}$", name)
	}
	return nil
}

// serviceEnabled returns true if the service is enabled (or no filter is set).
func serviceEnabled(cfg *config.Config, service string) bool {
	if len(cfg.EnabledServices) == 0 {
		return true
	}
	for _, s := range cfg.EnabledServices {
		if s == service {
			return true
		}
	}
	return false
}

// RegisterAll registers all tool packages with the server, applying tier, service, and mode filters.
// Each service package exposes Register(server, factory) which adds its tools.
// Tier and read-only filtering is enforced via middleware that intercepts tools/call
// requests, rejecting calls to tools excluded by the current config.
func RegisterAll(server *mcp.Server, factory *services.Factory, cfg *config.Config, tierMap map[string]config.ToolInfo, oauthMgr *auth.OAuthManager) {
	slog.Info("registering tools",
		"tier", cfg.ToolTier,
		"services", cfg.EnabledServices,
		"readOnly", cfg.ReadOnly,
	)

	// Install tier/read-only filtering middleware. This intercepts tools/call
	// requests and blocks calls to tools that are excluded by the current tier
	// or read-only config. tools/list responses are also filtered so excluded
	// tools never appear in the tool listing.
	if len(tierMap) > 0 {
		server.AddReceivingMiddleware(tierFilterMiddleware(cfg, tierMap))
	}

	// Phase 2: Core services (Gmail, Drive, Calendar, Sheets)
	if serviceEnabled(cfg, "gmail") {
		gmail.Register(server, factory)
		slog.Info("registered service", "service", "gmail")
	}
	if serviceEnabled(cfg, "drive") {
		drive.Register(server, factory)
		slog.Info("registered service", "service", "drive")
	}
	if serviceEnabled(cfg, "calendar") {
		calendar.Register(server, factory)
		slog.Info("registered service", "service", "calendar")
	}
	if serviceEnabled(cfg, "sheets") {
		sheets.Register(server, factory)
		slog.Info("registered service", "service", "sheets")
	}

	// Phase 3: Extended services (Docs, Tasks, Contacts, Chat)
	if serviceEnabled(cfg, "docs") {
		docs.Register(server, factory)
		slog.Info("registered service", "service", "docs")
	}
	if serviceEnabled(cfg, "tasks") {
		tasks.Register(server, factory)
		slog.Info("registered service", "service", "tasks")
	}
	if serviceEnabled(cfg, "contacts") {
		contacts.Register(server, factory)
		slog.Info("registered service", "service", "contacts")
	}
	if serviceEnabled(cfg, "chat") {
		chat.Register(server, factory)
		slog.Info("registered service", "service", "chat")
	}

	// Phase 4: Complete coverage (Forms, Slides, Search, Apps Script, Auth)
	if serviceEnabled(cfg, "forms") {
		forms.Register(server, factory)
		slog.Info("registered service", "service", "forms")
	}
	if serviceEnabled(cfg, "slides") {
		slides.Register(server, factory)
		slog.Info("registered service", "service", "slides")
	}
	if serviceEnabled(cfg, "search") {
		search.Register(server, factory, cfg.CSEID)
		slog.Info("registered service", "service", "search")
	}
	if serviceEnabled(cfg, "appscript") {
		appscript.Register(server, factory)
		slog.Info("registered service", "service", "appscript")
	}

	// Auth tool (filtered out when OAuth 2.1 is enabled)
	if !cfg.EnableOAuth21 {
		authtools.Register(server, oauthMgr)
		slog.Info("registered service", "service", "auth")
	}
}

// tierFilterMiddleware returns MCP middleware that enforces per-tool tier and
// read-only filtering. It blocks tools/call requests for tools that are above
// the configured tier or are write tools in read-only mode.
func tierFilterMiddleware(cfg *config.Config, tierMap map[string]config.ToolInfo) mcp.Middleware {
	// Pre-build the set of excluded tool names for fast lookup.
	excluded := make(map[string]bool)
	for toolName, info := range tierMap {
		if config.TierLevel(info.Tier) > config.TierLevel(cfg.ToolTier) {
			excluded[toolName] = true
		}
	}

	// readOnlyAllowed tracks which tools are safe to call in read-only mode.
	// Built lazily on first tools/list response (when annotations are available).
	readOnlyAllowed := make(map[string]bool)
	readOnlyBuilt := false

	return func(next mcp.MethodHandler) mcp.MethodHandler {
		return func(ctx context.Context, method string, req mcp.Request) (mcp.Result, error) {
			if method != "tools/call" {
				result, err := next(ctx, method, req)

				// Filter tools/list responses to hide excluded tools.
				if method == "tools/list" && err == nil {
					if listResult, ok := result.(*mcp.ListToolsResult); ok {
						// Build the read-only allowed set from actual tool annotations.
						if !readOnlyBuilt {
							for _, tool := range listResult.Tools {
								if tool.Annotations != nil && tool.Annotations.ReadOnlyHint {
									readOnlyAllowed[tool.Name] = true
								}
							}
							readOnlyBuilt = true
						}
						listResult.Tools = filterToolPtrList(listResult.Tools, excluded, cfg)
					}
				}

				return result, err
			}

			// Extract tool name from the request.
			params, ok := req.GetParams().(*mcp.CallToolParamsRaw)
			if !ok {
				return next(ctx, method, req)
			}

			toolName := params.Name

			// Check tier exclusion.
			if excluded[toolName] {
				return &mcp.CallToolResult{
					IsError: true,
					Content: []mcp.Content{&mcp.TextContent{
						Text: fmt.Sprintf("tool %q is not available at tier %q — upgrade to a higher tier or change TOOL_TIER config", toolName, cfg.ToolTier),
					}},
				}, nil
			}

			// Enforce read-only mode at call time: reject write tools.
			if cfg.ReadOnly && readOnlyBuilt && !readOnlyAllowed[toolName] {
				return &mcp.CallToolResult{
					IsError: true,
					Content: []mcp.Content{&mcp.TextContent{
						Text: fmt.Sprintf("tool %q is a write operation and cannot be called in read-only mode", toolName),
					}},
				}, nil
			}

			return next(ctx, method, req)
		}
	}
}

// filterToolPtrList removes tools from the list that are excluded by tier or
// read-only config.
func filterToolPtrList(tools []*mcp.Tool, excluded map[string]bool, cfg *config.Config) []*mcp.Tool {
	filtered := make([]*mcp.Tool, 0, len(tools))
	for _, tool := range tools {
		if excluded[tool.Name] {
			continue
		}
		// In read-only mode, exclude tools that are not marked as read-only.
		if cfg.ReadOnly && (tool.Annotations == nil || !tool.Annotations.ReadOnlyHint) {
			continue
		}
		filtered = append(filtered, tool)
	}
	return filtered
}

// ShouldIncludeTool checks whether a tool should be registered based on the current config.
func ShouldIncludeTool(toolName string, cfg *config.Config, tierMap map[string]config.ToolInfo, annotations *mcp.ToolAnnotations) bool {
	info, ok := tierMap[toolName]
	if !ok {
		slog.Warn("tool not found in tier config, skipping", "tool", toolName)
		return false
	}

	// Filter by tier level
	if config.TierLevel(info.Tier) > config.TierLevel(cfg.ToolTier) {
		return false
	}

	// Filter by enabled services
	if len(cfg.EnabledServices) > 0 {
		found := false
		for _, svc := range cfg.EnabledServices {
			if svc == info.Service {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	// Filter by read-only mode: exclude tools that are not read-only
	if cfg.ReadOnly && annotations != nil && !annotations.ReadOnlyHint {
		return false
	}

	// Filter out legacy auth tool when OAuth 2.1 is enabled
	if cfg.EnableOAuth21 && toolName == "start_google_auth" {
		return false
	}

	return true
}
