package middleware

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/evert/google-workspace-mcp-go/internal/auth"
)

// testOAuthMgr returns an OAuthManager with dummy credentials for testing.
// The generated auth URL will contain the client ID and redirect URL.
func testOAuthMgr() *auth.OAuthManager {
	return auth.NewOAuthManager(
		"test-client-id",
		"test-client-secret",
		"http://localhost:8000/oauth/callback",
		[]string{"https://www.googleapis.com/auth/gmail.readonly"},
		nil, // token store not used for URL generation
	)
}

// fakeToolRequest builds a CallToolRequest with the given arguments JSON.
func fakeToolRequest(argsJSON string) mcp.Request {
	return &mcp.CallToolRequest{
		Params: &mcp.CallToolParamsRaw{
			Name:      "search_gmail_messages",
			Arguments: json.RawMessage(argsJSON),
		},
	}
}

func TestAuthEnhancer_NoCredentials(t *testing.T) {
	oauthMgr := testOAuthMgr()
	mw := AuthEnhancerMiddleware(oauthMgr)

	errText := "no credentials found for user@test.com — call start_google_auth to authenticate"
	next := func(_ context.Context, _ string, _ mcp.Request) (mcp.Result, error) {
		return &mcp.CallToolResult{
			IsError: true,
			Content: []mcp.Content{&mcp.TextContent{Text: errText}},
		}, nil
	}

	handler := mw(next)
	req := fakeToolRequest(`{"user_google_email":"user@test.com","query":"test"}`)
	result, err := handler(context.Background(), "tools/call", req)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	toolResult := result.(*mcp.CallToolResult)
	text := toolResult.Content[0].(*mcp.TextContent).Text

	if !strings.Contains(text, errText) {
		t.Errorf("original error text missing, got: %s", text)
	}
	if !strings.Contains(text, "Please authenticate by visiting this URL:") {
		t.Errorf("auth prompt missing, got: %s", text)
	}
	if !strings.Contains(text, "accounts.google.com") {
		t.Errorf("auth URL missing, got: %s", text)
	}
	if !strings.Contains(text, "test-client-id") {
		t.Errorf("expected client ID in auth URL, got: %s", text)
	}
}

func TestAuthEnhancer_AuthExpired(t *testing.T) {
	oauthMgr := testOAuthMgr()
	mw := AuthEnhancerMiddleware(oauthMgr)

	errText := "authentication expired for this user — call start_google_auth tool to re-authenticate"
	next := func(_ context.Context, _ string, _ mcp.Request) (mcp.Result, error) {
		return &mcp.CallToolResult{
			IsError: true,
			Content: []mcp.Content{&mcp.TextContent{Text: errText}},
		}, nil
	}

	handler := mw(next)
	req := fakeToolRequest(`{"user_google_email":"user@test.com"}`)
	result, _ := handler(context.Background(), "tools/call", req)

	toolResult := result.(*mcp.CallToolResult)
	text := toolResult.Content[0].(*mcp.TextContent).Text

	if !strings.Contains(text, "Please authenticate by visiting this URL:") {
		t.Errorf("auth prompt missing for expired auth, got: %s", text)
	}
}

func TestAuthEnhancer_NonAuthError_Unchanged(t *testing.T) {
	oauthMgr := testOAuthMgr()
	mw := AuthEnhancerMiddleware(oauthMgr)

	errText := "resource not found — verify the ID is correct"
	next := func(_ context.Context, _ string, _ mcp.Request) (mcp.Result, error) {
		return &mcp.CallToolResult{
			IsError: true,
			Content: []mcp.Content{&mcp.TextContent{Text: errText}},
		}, nil
	}

	handler := mw(next)
	req := fakeToolRequest(`{"user_google_email":"user@test.com","query":"test"}`)
	result, _ := handler(context.Background(), "tools/call", req)

	toolResult := result.(*mcp.CallToolResult)
	text := toolResult.Content[0].(*mcp.TextContent).Text

	if text != errText {
		t.Errorf("non-auth error should be unchanged, got: %s", text)
	}
}

func TestAuthEnhancer_NonToolCall_Unchanged(t *testing.T) {
	oauthMgr := testOAuthMgr()
	mw := AuthEnhancerMiddleware(oauthMgr)

	// Simulate a non-tool-call method (e.g. tools/list) that returns a tool list result.
	// The middleware should pass through any non-"tools/call" method unchanged.
	next := func(_ context.Context, _ string, _ mcp.Request) (mcp.Result, error) {
		return &mcp.ListToolsResult{}, nil
	}

	handler := mw(next)
	req := &mcp.ServerRequest[*mcp.ListToolsParams]{Params: &mcp.ListToolsParams{}}
	result, err := handler(context.Background(), "tools/list", req)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := result.(*mcp.ListToolsResult); !ok {
		t.Errorf("expected ListToolsResult, got %T", result)
	}
}

func TestAuthEnhancer_MissingEmail_Unchanged(t *testing.T) {
	oauthMgr := testOAuthMgr()
	mw := AuthEnhancerMiddleware(oauthMgr)

	errText := "no credentials found for unknown — call start_google_auth to authenticate"
	next := func(_ context.Context, _ string, _ mcp.Request) (mcp.Result, error) {
		return &mcp.CallToolResult{
			IsError: true,
			Content: []mcp.Content{&mcp.TextContent{Text: errText}},
		}, nil
	}

	handler := mw(next)
	// Arguments with no user_google_email field.
	req := fakeToolRequest(`{"query":"test"}`)
	result, _ := handler(context.Background(), "tools/call", req)

	toolResult := result.(*mcp.CallToolResult)
	text := toolResult.Content[0].(*mcp.TextContent).Text

	if text != errText {
		t.Errorf("error with missing email should be unchanged, got: %s", text)
	}
}

func TestAuthEnhancer_SuccessResult_Unchanged(t *testing.T) {
	oauthMgr := testOAuthMgr()
	mw := AuthEnhancerMiddleware(oauthMgr)

	next := func(_ context.Context, _ string, _ mcp.Request) (mcp.Result, error) {
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: "Search complete: 5 results"}},
		}, nil
	}

	handler := mw(next)
	req := fakeToolRequest(`{"user_google_email":"user@test.com","query":"test"}`)
	result, _ := handler(context.Background(), "tools/call", req)

	toolResult := result.(*mcp.CallToolResult)
	text := toolResult.Content[0].(*mcp.TextContent).Text

	if text != "Search complete: 5 results" {
		t.Errorf("successful result should be unchanged, got: %s", text)
	}
}

func TestAuthEnhancer_NilResult_NoPanic(t *testing.T) {
	oauthMgr := testOAuthMgr()
	mw := AuthEnhancerMiddleware(oauthMgr)

	// Simulate the SDK returning a typed-nil *CallToolResult with an error,
	// which is what happens when input validation fails before the handler runs.
	next := func(_ context.Context, _ string, _ mcp.Request) (mcp.Result, error) {
		var r *mcp.CallToolResult // nil
		return r, fmt.Errorf("validation failed: missing required field")
	}

	handler := mw(next)
	req := fakeToolRequest(`{"user_google_email":"user@test.com"}`)

	// Must not panic.
	result, err := handler(context.Background(), "tools/call", req)

	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if err.Error() != "validation failed: missing required field" {
		t.Errorf("unexpected error text: %v", err)
	}
	// result is a typed-nil *CallToolResult wrapped in mcp.Result interface.
	if toolResult, ok := result.(*mcp.CallToolResult); ok && toolResult != nil {
		t.Errorf("expected nil *CallToolResult, got %+v", toolResult)
	}
}

func TestIsAuthRelatedError(t *testing.T) {
	tests := []struct {
		text string
		want bool
	}{
		{"no credentials found for user@test.com — call start_google_auth to authenticate", true},
		{"authentication expired for this user — call start_google_auth tool to re-authenticate", true},
		{"gmail client for user@test.com: no credentials found for user@test.com", true},
		{"resource not found — verify the ID is correct", false},
		{"rate limit exceeded", false},
		{"", false},
	}

	for _, tt := range tests {
		got := isAuthRelatedError(tt.text)
		if got != tt.want {
			t.Errorf("isAuthRelatedError(%q) = %v, want %v", tt.text, got, tt.want)
		}
	}
}
