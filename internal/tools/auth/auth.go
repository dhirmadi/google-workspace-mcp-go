// Package auth implements the start_google_auth MCP tool for legacy OAuth 2.0 authentication.
// This tool is filtered out when MCP_ENABLE_OAUTH21 is true.
package auth

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	iauth "github.com/evert/google-workspace-mcp-go/internal/auth"
	"github.com/evert/google-workspace-mcp-go/internal/pkg/ptr"
	"github.com/evert/google-workspace-mcp-go/internal/pkg/response"
)

var serviceIcons = []mcp.Icon{{
	Source:   "https://www.gstatic.com/images/branding/product/1x/googleg_48dp.png",
	MIMEType: "image/png",
	Sizes:    []string{"48x48"},
}}

// Register registers the start_google_auth tool with the MCP server.
func Register(server *mcp.Server, oauthMgr *iauth.OAuthManager) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "start_google_auth",
		Icons:       serviceIcons,
		Description: "Start the Google OAuth 2.0 authentication flow. Returns an authorization URL for the user to visit and grant access to their Google Workspace account. After the user completes authentication, their credentials are stored for subsequent API calls.",
		Annotations: &mcp.ToolAnnotations{
			Title:         "Authenticate with Google",
			OpenWorldHint: ptr.Bool(true),
		},
	}, createStartAuthHandler(oauthMgr))
}

type StartAuthInput struct {
	UserEmail string `json:"user_google_email" jsonschema:"required" jsonschema_description:"The user's Google email address to authenticate"`
}

// StartAuthOutput exposes the authorization URL for MCP clients that surface structuredContent
// more reliably than plain tool text (some hosts hide long URLs in chat).
type StartAuthOutput struct {
	AuthURL   string `json:"auth_url"`
	UserEmail string `json:"user_google_email"`
}

func createStartAuthHandler(oauthMgr *iauth.OAuthManager) mcp.ToolHandlerFor[StartAuthInput, StartAuthOutput] {
	return func(ctx context.Context, req *mcp.CallToolRequest, input StartAuthInput) (*mcp.CallToolResult, StartAuthOutput, error) {
		// Generate auth URL with user email as state
		authURL := oauthMgr.GetAuthURL(input.UserEmail)

		// Duplicate on stderr: stdio MCP uses stdout for the protocol; Cursor and similar clients
		// often show MCP server stderr in logs while hiding tool output in the chat UI.
		slog.Warn("google oauth authorization URL issued — copy from stderr/MCP logs if the client hides tool text",
			"user_google_email", input.UserEmail,
			"auth_url", authURL,
		)
		fmt.Fprintf(os.Stderr, "\n[google-workspace-mcp] OAuth URL (open in browser). If your client hides tool results, copy from here or MCP server logs:\n%s\n\n", authURL)

		rb := response.New()
		rb.Header("Google Authentication")
		rb.Line("Please visit the following URL to authenticate:")
		rb.Blank()
		rb.Raw(authURL)
		rb.Blank()
		rb.Line("After granting access, the OAuth callback will automatically capture the authorization code.")
		rb.Line("Authenticating as: %s", input.UserEmail)
		rb.Blank()
		rb.Line("If you do not see a clickable link above, open the MCP server output / stderr for the same URL.")

		return rb.TextResult(), StartAuthOutput{AuthURL: authURL, UserEmail: input.UserEmail}, nil
	}
}
