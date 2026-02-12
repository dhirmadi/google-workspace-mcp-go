// Package auth implements the start_google_auth MCP tool for legacy OAuth 2.0 authentication.
// This tool is filtered out when MCP_ENABLE_OAUTH21 is true.
package auth

import (
	"context"

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

func createStartAuthHandler(oauthMgr *iauth.OAuthManager) mcp.ToolHandlerFor[StartAuthInput, any] {
	return func(ctx context.Context, req *mcp.CallToolRequest, input StartAuthInput) (*mcp.CallToolResult, any, error) {
		// Generate auth URL with user email as state
		authURL := oauthMgr.GetAuthURL(input.UserEmail)

		rb := response.New()
		rb.Header("Google Authentication")
		rb.Line("Please visit the following URL to authenticate:")
		rb.Blank()
		rb.Raw(authURL)
		rb.Blank()
		rb.Line("After granting access, the OAuth callback will automatically capture the authorization code.")
		rb.Line("Authenticating as: %s", input.UserEmail)

		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: rb.Build()}},
		}, nil, nil
	}
}
