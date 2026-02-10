package middleware

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/evert/google-workspace-mcp-go/internal/auth"
)

// authErrorMarkers are substrings that identify auth-related tool errors.
var authErrorMarkers = []string{
	"start_google_auth",
	"no credentials found",
	"authentication expired",
}

// AuthEnhancerMiddleware returns MCP SDK middleware that detects auth-related
// tool errors and appends the OAuth authentication URL so the user can
// authenticate without an extra round-trip.
func AuthEnhancerMiddleware(oauthMgr *auth.OAuthManager) mcp.Middleware {
	return func(next mcp.MethodHandler) mcp.MethodHandler {
		return func(ctx context.Context, method string, req mcp.Request) (mcp.Result, error) {
			result, err := next(ctx, method, req)

			// Only enhance tools/call responses.
			if method != "tools/call" {
				return result, err
			}

			// Check whether the result is a tool-error CallToolResult.
			toolResult, ok := result.(*mcp.CallToolResult)
			if !ok || !toolResult.IsError || len(toolResult.Content) == 0 {
				return result, err
			}

			textContent, ok := toolResult.Content[0].(*mcp.TextContent)
			if !ok {
				return result, err
			}

			if !isAuthRelatedError(textContent.Text) {
				return result, err
			}

			// Extract user_google_email from the raw tool arguments.
			userEmail := extractUserEmail(req)
			if userEmail == "" {
				return result, err
			}

			// Append the auth URL to the existing error message.
			authURL := oauthMgr.GetAuthURL(userEmail)
			textContent.Text = fmt.Sprintf(
				"%s\n\nPlease authenticate by visiting this URL:\n%s",
				textContent.Text, authURL,
			)

			return result, err
		}
	}
}

// isAuthRelatedError returns true if the text contains any auth-error marker.
func isAuthRelatedError(text string) bool {
	lower := strings.ToLower(text)
	for _, marker := range authErrorMarkers {
		if strings.Contains(lower, marker) {
			return true
		}
	}
	return false
}

// extractUserEmail tries to read user_google_email from the raw tool arguments.
func extractUserEmail(req mcp.Request) string {
	params, ok := req.GetParams().(*mcp.CallToolParamsRaw)
	if !ok {
		return ""
	}

	var args struct {
		UserEmail string `json:"user_google_email"`
	}
	if err := json.Unmarshal(params.Arguments, &args); err != nil {
		return ""
	}
	return args.UserEmail
}
