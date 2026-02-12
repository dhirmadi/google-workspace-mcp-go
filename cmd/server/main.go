package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/evert/google-workspace-mcp-go/internal/auth"
	"github.com/evert/google-workspace-mcp-go/internal/config"
	"github.com/evert/google-workspace-mcp-go/internal/middleware"
	"github.com/evert/google-workspace-mcp-go/internal/registry"
	"github.com/evert/google-workspace-mcp-go/internal/services"
)

func main() {
	// Structured logging to stderr (stdout is reserved for MCP stdio transport)
	logger := slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)

	if err := run(ctx, logger); err != nil {
		cancel()
		slog.Error("fatal error", "error", err)
		os.Exit(1)
	}
	cancel()
}

func run(ctx context.Context, logger *slog.Logger) error {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	// Set log level from config
	switch cfg.LogLevel {
	case "debug":
		slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelDebug})))
	case "warn":
		slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelWarn})))
	case "error":
		slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError})))
	}

	// Initialize token store
	tokenStore, err := auth.NewFileTokenStore(cfg.CredentialsDir)
	if err != nil {
		return fmt.Errorf("initializing token store: %w", err)
	}

	// Determine scopes
	scopes := auth.AllScopes(cfg.EnabledServices, cfg.ReadOnly)

	// Create OAuth manager
	oauthMgr := auth.NewOAuthManager(
		cfg.OAuth.ClientID,
		cfg.OAuth.ClientSecret,
		cfg.OAuth.RedirectURL,
		scopes,
		tokenStore,
	)

	// Create service factory
	factory := services.NewFactory(oauthMgr)

	// Load tier config — try absolute path (container) then relative (local dev)
	tierConfigPath := "/configs/tool_tiers.yaml"
	if _, statErr := os.Stat(tierConfigPath); statErr != nil {
		tierConfigPath = filepath.Join("configs", "tool_tiers.yaml")
	}
	tierMap, err := config.LoadTiers(tierConfigPath)
	if err != nil {
		slog.Warn("could not load tier config — all tools will be registered unfiltered",
			"path", tierConfigPath,
			"error", err,
		)
		tierMap = make(map[string]config.ToolInfo)
	}

	// Create MCP server
	server := mcp.NewServer(&mcp.Implementation{
		Name:    "google-workspace-mcp",
		Version: "1.0.0",
	}, nil)

	// Wire SDK middleware
	server.AddReceivingMiddleware(
		middleware.LoggingMiddleware(logger),
		middleware.AuthEnhancerMiddleware(oauthMgr),
	)

	// Register all tools through the registry
	registry.RegisterAll(server, factory, cfg, tierMap, oauthMgr)

	slog.Info("starting Google Workspace MCP server",
		"transport", cfg.Server.Transport,
		"tier", cfg.ToolTier,
		"readOnly", cfg.ReadOnly,
	)

	// Start server on selected transport
	switch cfg.Server.Transport {
	case "stdio":
		if err := server.Run(ctx, &mcp.StdioTransport{}); err != nil {
			return fmt.Errorf("stdio server error: %w", err)
		}

	case "streamable-http":
		mcpHandler := mcp.NewStreamableHTTPHandler(
			func(r *http.Request) *mcp.Server { return server },
			nil,
		)

		// Use a mux to route /oauth/callback separately from MCP
		mux := http.NewServeMux()
		mux.Handle("/mcp", mcpHandler)
		mux.HandleFunc("/oauth/callback", auth.OAuthCallbackHandler(oauthMgr, factory))

		addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
		httpServer := &http.Server{
			Addr:    addr,
			Handler: mux,
		}

		// Graceful shutdown
		go func() {
			<-ctx.Done()
			slog.Info("shutting down HTTP server")
			shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer shutdownCancel()
			if err := httpServer.Shutdown(shutdownCtx); err != nil {
				slog.Error("HTTP server shutdown error", "error", err)
			}
		}()

		slog.Info("listening", "addr", addr)
		if err := httpServer.ListenAndServe(); err != http.ErrServerClosed {
			return fmt.Errorf("HTTP server error: %w", err)
		}

	default:
		return fmt.Errorf("unknown transport %q — use 'stdio' or 'streamable-http'", cfg.Server.Transport)
	}

	return nil
}
