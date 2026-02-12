package config

import (
	"flag"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// Config holds all server configuration loaded from environment variables and CLI flags.
type Config struct {
	OAuth struct {
		ClientID     string
		ClientSecret string
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
	CSEID           string
	DefaultEmail    string
}

// Load reads configuration from environment variables and CLI flags.
// CLI flags take precedence over environment variables.
func Load() (*Config, error) {
	cfg := &Config{}

	// Environment variables
	cfg.OAuth.ClientID = os.Getenv("GOOGLE_OAUTH_CLIENT_ID")
	cfg.OAuth.ClientSecret = os.Getenv("GOOGLE_OAUTH_CLIENT_SECRET")
	cfg.CSEID = os.Getenv("GOOGLE_CSE_ID")
	cfg.DefaultEmail = os.Getenv("USER_GOOGLE_EMAIL")

	cfg.CredentialsDir = os.Getenv("WORKSPACE_MCP_CREDENTIALS_DIR")
	if cfg.CredentialsDir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("cannot determine home directory: %w", err)
		}
		cfg.CredentialsDir = filepath.Join(home, ".google_workspace_mcp", "credentials")
	}

	// Enabled services (comma-separated, empty = all)
	if svcEnv := os.Getenv("ENABLED_SERVICES"); svcEnv != "" {
		for _, s := range strings.Split(svcEnv, ",") {
			s = strings.TrimSpace(s)
			if s != "" {
				cfg.EnabledServices = append(cfg.EnabledServices, s)
			}
		}
	}

	cfg.Server.Host = envOrDefault("WORKSPACE_MCP_HOST", "0.0.0.0")
	cfg.Server.BaseURI = envOrDefault("WORKSPACE_MCP_BASE_URI", "http://localhost")
	cfg.Server.Transport = envOrDefault("MCP_TRANSPORT", "stdio")
	cfg.LogLevel = envOrDefault("LOG_LEVEL", "info")
	cfg.ToolTier = envOrDefault("TOOL_TIER", "complete")
	cfg.SingleUser = envBool("MCP_SINGLE_USER_MODE")
	cfg.EnableOAuth21 = envBool("MCP_ENABLE_OAUTH21")
	cfg.StatelessMode = envBool("WORKSPACE_MCP_STATELESS_MODE")
	cfg.ReadOnly = envBool("WORKSPACE_MCP_READ_ONLY")

	// Port
	portStr := os.Getenv("MCP_PORT")
	if portStr == "" {
		portStr = os.Getenv("PORT")
	}
	if portStr == "" {
		portStr = "8000"
	}
	port, err := strconv.Atoi(portStr)
	if err != nil {
		return nil, fmt.Errorf("invalid port %q: %w", portStr, err)
	}
	cfg.Server.Port = port

	// CLI flags override env vars
	flag.StringVar(&cfg.Server.Transport, "transport", cfg.Server.Transport, "Transport mode: stdio or streamable-http")
	var toolsFlag string
	flag.StringVar(&toolsFlag, "tools", "", "Services to enable (comma-separated): gmail,drive,calendar,docs,sheets,chat,forms,slides,tasks,contacts,search,appscript")
	flag.StringVar(&cfg.ToolTier, "tool-tier", cfg.ToolTier, "Load tools by tier: core, extended, or complete")
	flag.BoolVar(&cfg.SingleUser, "single-user", cfg.SingleUser, "Bypass session mapping, use any credentials")
	flag.BoolVar(&cfg.ReadOnly, "read-only", cfg.ReadOnly, "Request only read-only scopes, disable write tools")
	flag.Parse()

	// CLI --tools flag overrides (not appends to) the ENABLED_SERVICES env var.
	if toolsFlag != "" {
		cfg.EnabledServices = nil
		for _, s := range strings.Split(toolsFlag, ",") {
			s = strings.TrimSpace(s)
			if s != "" {
				cfg.EnabledServices = append(cfg.EnabledServices, s)
			}
		}
	}

	// Validate required fields
	if cfg.OAuth.ClientID == "" {
		return nil, fmt.Errorf("GOOGLE_OAUTH_CLIENT_ID environment variable is required")
	}
	if cfg.OAuth.ClientSecret == "" {
		return nil, fmt.Errorf("GOOGLE_OAUTH_CLIENT_SECRET environment variable is required")
	}

	// Build OAuth redirect URL
	// If the base URI already includes a port, use it as-is; otherwise append the server port.
	parsedURI, parseErr := url.Parse(cfg.Server.BaseURI)
	if parseErr == nil && parsedURI.Port() != "" {
		cfg.OAuth.RedirectURL = cfg.Server.BaseURI + "/oauth/callback"
	} else {
		cfg.OAuth.RedirectURL = fmt.Sprintf("%s:%d/oauth/callback", cfg.Server.BaseURI, cfg.Server.Port)
	}

	return cfg, nil
}

func envOrDefault(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func envBool(key string) bool {
	v := strings.ToLower(os.Getenv(key))
	return v == "true" || v == "1" || v == "yes"
}
