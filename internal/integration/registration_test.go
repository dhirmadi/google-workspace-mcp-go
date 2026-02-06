//go:build integration

// Package integration contains integration tests that verify full system behavior
// without requiring real Google API credentials.
package integration

import (
	"os"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/evert/google-workspace-mcp-go/internal/auth"
	"github.com/evert/google-workspace-mcp-go/internal/config"
	"github.com/evert/google-workspace-mcp-go/internal/registry"
	"github.com/evert/google-workspace-mcp-go/internal/services"
)

// Shared state loaded once in TestMain.
var (
	sharedCfg     *config.Config
	sharedTierMap map[string]config.ToolInfo
)

func TestMain(m *testing.M) {
	// Set required env for all tests
	os.Setenv("GOOGLE_OAUTH_CLIENT_ID", "test-client-id")
	os.Setenv("GOOGLE_OAUTH_CLIENT_SECRET", "test-client-secret")
	os.Setenv("MCP_TRANSPORT", "stdio")
	os.Setenv("TOOL_TIER", "complete")

	tmpDir, err := os.MkdirTemp("", "mcp-integration-*")
	if err != nil {
		panic("creating temp dir: " + err.Error())
	}
	os.Setenv("WORKSPACE_MCP_CREDENTIALS_DIR", tmpDir)
	defer os.RemoveAll(tmpDir)

	// Load config once (calls flag.Parse)
	cfg, err := config.Load()
	if err != nil {
		panic("loading config: " + err.Error())
	}
	sharedCfg = cfg

	tierMap, err := config.LoadTiers("../../configs/tool_tiers.yaml")
	if err != nil {
		panic("loading tier config: " + err.Error())
	}
	sharedTierMap = tierMap

	os.Exit(m.Run())
}

// createTestServer creates a fully wired MCP server for testing.
func createTestServer(t *testing.T) *mcp.Server {
	t.Helper()

	tokenStore, err := auth.NewFileTokenStore(sharedCfg.CredentialsDir)
	if err != nil {
		t.Fatalf("creating token store: %v", err)
	}

	scopes := auth.AllScopes(sharedCfg.EnabledServices, sharedCfg.ReadOnly)
	oauthMgr := auth.NewOAuthManager(
		sharedCfg.OAuth.ClientID,
		sharedCfg.OAuth.ClientSecret,
		sharedCfg.OAuth.RedirectURL,
		scopes,
		tokenStore,
	)

	factory := services.NewFactory(oauthMgr)

	server := mcp.NewServer(&mcp.Implementation{
		Name:    "google-workspace-mcp",
		Version: "1.0.0-test",
	}, nil)

	registry.RegisterAll(server, factory, sharedCfg, sharedTierMap, oauthMgr)
	return server
}

func TestFullToolRegistration(t *testing.T) {
	server := createTestServer(t)

	if server == nil {
		t.Fatal("server is nil after registration")
	}

	// Verify tier map has the expected number of tools
	toolCount := 0
	for range sharedTierMap {
		toolCount++
	}

	expectedTotal := 136
	if toolCount != expectedTotal {
		t.Errorf("tier config has %d tools, expected %d", toolCount, expectedTotal)
	}
}

func TestConfigValues(t *testing.T) {
	if sharedCfg.OAuth.ClientID != "test-client-id" {
		t.Errorf("client ID = %q, want %q", sharedCfg.OAuth.ClientID, "test-client-id")
	}
	if sharedCfg.Server.Transport != "stdio" {
		t.Errorf("transport = %q, want %q", sharedCfg.Server.Transport, "stdio")
	}
	if sharedCfg.ToolTier != "complete" {
		t.Errorf("tool tier = %q, want %q", sharedCfg.ToolTier, "complete")
	}
}

func TestTierFiltering(t *testing.T) {
	tests := []struct {
		name     string
		tier     string
		minTools int
	}{
		{"core tier", "core", 40},
		{"extended tier", "extended", 80},
		{"complete tier", "complete", 130},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			count := 0
			for _, info := range sharedTierMap {
				if config.TierLevel(info.Tier) <= config.TierLevel(tt.tier) {
					count++
				}
			}
			if count < tt.minTools {
				t.Errorf("tier %q has %d tools, expected at least %d", tt.tier, count, tt.minTools)
			}
		})
	}
}

func TestToolNameValidation(t *testing.T) {
	for name := range sharedTierMap {
		if err := registry.ValidateToolName(name); err != nil {
			t.Errorf("tool name %q failed SEP-986 validation: %v", name, err)
		}
	}
}

func TestReadOnlyModeFiltering(t *testing.T) {
	cfg := &config.Config{
		ToolTier: "complete",
		ReadOnly: true,
	}

	readOnlyTools := []string{
		"search_gmail_messages",
		"get_gmail_message_content",
		"list_calendars",
		"get_events",
		"search_drive_files",
		"read_sheet_values",
	}

	writeTools := []string{
		"send_gmail_message",
		"create_event",
		"create_drive_file",
		"modify_sheet_values",
	}

	for _, name := range readOnlyTools {
		annotations := &mcp.ToolAnnotations{ReadOnlyHint: true}
		if !registry.ShouldIncludeTool(name, cfg, sharedTierMap, annotations) {
			t.Errorf("read-only tool %q should be included in read-only mode", name)
		}
	}

	for _, name := range writeTools {
		annotations := &mcp.ToolAnnotations{ReadOnlyHint: false}
		if registry.ShouldIncludeTool(name, cfg, sharedTierMap, annotations) {
			t.Errorf("write tool %q should be excluded in read-only mode", name)
		}
	}
}

func TestServiceFiltering(t *testing.T) {
	cfg := &config.Config{
		ToolTier:        "complete",
		EnabledServices: []string{"gmail"},
	}

	// Gmail tools should be included
	annotations := &mcp.ToolAnnotations{ReadOnlyHint: true}
	if !registry.ShouldIncludeTool("search_gmail_messages", cfg, sharedTierMap, annotations) {
		t.Error("search_gmail_messages should be included when gmail is enabled")
	}

	// Drive tools should be excluded
	if registry.ShouldIncludeTool("search_drive_files", cfg, sharedTierMap, annotations) {
		t.Error("search_drive_files should be excluded when only gmail is enabled")
	}
}
