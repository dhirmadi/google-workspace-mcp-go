package registry

import (
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
func RegisterAll(server *mcp.Server, factory *services.Factory, cfg *config.Config, tierMap map[string]config.ToolInfo, oauthMgr *auth.OAuthManager) {
	slog.Info("registering tools",
		"tier", cfg.ToolTier,
		"services", cfg.EnabledServices,
		"readOnly", cfg.ReadOnly,
	)

	_ = tierMap // TODO: per-tool tier filtering will be added when we have more tools per service

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
