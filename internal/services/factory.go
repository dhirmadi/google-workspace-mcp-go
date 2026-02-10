package services

import (
	"context"
	"fmt"
	"net/http"
	"sync"

	"golang.org/x/oauth2"
	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/chat/v1"
	customsearch "google.golang.org/api/customsearch/v1"
	"google.golang.org/api/docs/v1"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/forms/v1"
	"google.golang.org/api/gmail/v1"
	"google.golang.org/api/option"
	"google.golang.org/api/people/v1"
	"google.golang.org/api/script/v1"
	"google.golang.org/api/sheets/v4"
	"google.golang.org/api/slides/v1"
	"google.golang.org/api/tasks/v1"

	"github.com/evert/google-workspace-mcp-go/internal/auth"
)

// Factory manages authenticated Google API service clients per user email.
// Clients are cached with ReuseTokenSource for concurrency-safe auto-refresh.
type Factory struct {
	oauthConfig *oauth2.Config
	tokenStore  auth.TokenStore
	mu          sync.RWMutex
	clients     map[string]*http.Client
}

// NewFactory creates a service factory backed by the given OAuth manager.
func NewFactory(oauthMgr *auth.OAuthManager) *Factory {
	return &Factory{
		oauthConfig: oauthMgr.Config(),
		tokenStore:  oauthMgr.TokenStore(),
		clients:     make(map[string]*http.Client),
	}
}

// clientFor returns a cached, auto-refreshing HTTP client for the user.
// IMPORTANT: Uses context.Background() for the cached HTTP client/token source
// so they outlive any single request context. Individual API calls pass their
// own request context via .Context(ctx) on each Google API call.
func (f *Factory) clientFor(ctx context.Context, userEmail string) (*http.Client, error) {
	// Fast path: check cache
	f.mu.RLock()
	client, ok := f.clients[userEmail]
	f.mu.RUnlock()
	if ok {
		return client, nil
	}

	// Slow path: create new client
	f.mu.Lock()
	defer f.mu.Unlock()

	// Double-check after acquiring write lock
	if client, ok := f.clients[userEmail]; ok {
		return client, nil
	}

	token, err := f.tokenStore.Load(userEmail)
	if err != nil {
		return nil, err
	}

	// Use context.Background() for the token source and HTTP client so they
	// outlive the originating request. Each Google API call passes its own
	// request-scoped context via .Context(ctx).Do(), which correctly controls
	// the lifetime of individual HTTP requests.
	bgCtx := context.Background()
	baseSource := f.oauthConfig.TokenSource(bgCtx, token)
	reuseSource := oauth2.ReuseTokenSource(token, &auth.PersistingTokenSource{
		Base:      baseSource,
		Store:     f.tokenStore,
		UserEmail: userEmail,
	})

	client = oauth2.NewClient(bgCtx, reuseSource)
	f.clients[userEmail] = client
	return client, nil
}

// InvalidateClient removes the cached HTTP client for a user, forcing the
// next API call to rebuild it from the latest persisted token. Call this
// after re-authentication to ensure fresh credentials are picked up.
func (f *Factory) InvalidateClient(userEmail string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	delete(f.clients, userEmail)
}

// Gmail returns a Gmail service client for the given user.
func (f *Factory) Gmail(ctx context.Context, userEmail string) (*gmail.Service, error) {
	client, err := f.clientFor(ctx, userEmail)
	if err != nil {
		return nil, fmt.Errorf("gmail client for %s: %w", userEmail, err)
	}
	return gmail.NewService(ctx, option.WithHTTPClient(client))
}

// Drive returns a Drive service client for the given user.
func (f *Factory) Drive(ctx context.Context, userEmail string) (*drive.Service, error) {
	client, err := f.clientFor(ctx, userEmail)
	if err != nil {
		return nil, fmt.Errorf("drive client for %s: %w", userEmail, err)
	}
	return drive.NewService(ctx, option.WithHTTPClient(client))
}

// Calendar returns a Calendar service client for the given user.
func (f *Factory) Calendar(ctx context.Context, userEmail string) (*calendar.Service, error) {
	client, err := f.clientFor(ctx, userEmail)
	if err != nil {
		return nil, fmt.Errorf("calendar client for %s: %w", userEmail, err)
	}
	return calendar.NewService(ctx, option.WithHTTPClient(client))
}

// Docs returns a Docs service client for the given user.
func (f *Factory) Docs(ctx context.Context, userEmail string) (*docs.Service, error) {
	client, err := f.clientFor(ctx, userEmail)
	if err != nil {
		return nil, fmt.Errorf("docs client for %s: %w", userEmail, err)
	}
	return docs.NewService(ctx, option.WithHTTPClient(client))
}

// Sheets returns a Sheets service client for the given user.
func (f *Factory) Sheets(ctx context.Context, userEmail string) (*sheets.Service, error) {
	client, err := f.clientFor(ctx, userEmail)
	if err != nil {
		return nil, fmt.Errorf("sheets client for %s: %w", userEmail, err)
	}
	return sheets.NewService(ctx, option.WithHTTPClient(client))
}

// Slides returns a Slides service client for the given user.
func (f *Factory) Slides(ctx context.Context, userEmail string) (*slides.Service, error) {
	client, err := f.clientFor(ctx, userEmail)
	if err != nil {
		return nil, fmt.Errorf("slides client for %s: %w", userEmail, err)
	}
	return slides.NewService(ctx, option.WithHTTPClient(client))
}

// Chat returns a Chat service client for the given user.
func (f *Factory) Chat(ctx context.Context, userEmail string) (*chat.Service, error) {
	client, err := f.clientFor(ctx, userEmail)
	if err != nil {
		return nil, fmt.Errorf("chat client for %s: %w", userEmail, err)
	}
	return chat.NewService(ctx, option.WithHTTPClient(client))
}

// Forms returns a Forms service client for the given user.
func (f *Factory) Forms(ctx context.Context, userEmail string) (*forms.Service, error) {
	client, err := f.clientFor(ctx, userEmail)
	if err != nil {
		return nil, fmt.Errorf("forms client for %s: %w", userEmail, err)
	}
	return forms.NewService(ctx, option.WithHTTPClient(client))
}

// Tasks returns a Tasks service client for the given user.
func (f *Factory) Tasks(ctx context.Context, userEmail string) (*tasks.Service, error) {
	client, err := f.clientFor(ctx, userEmail)
	if err != nil {
		return nil, fmt.Errorf("tasks client for %s: %w", userEmail, err)
	}
	return tasks.NewService(ctx, option.WithHTTPClient(client))
}

// People returns a People service client for the given user (Contacts).
func (f *Factory) People(ctx context.Context, userEmail string) (*people.Service, error) {
	client, err := f.clientFor(ctx, userEmail)
	if err != nil {
		return nil, fmt.Errorf("people client for %s: %w", userEmail, err)
	}
	return people.NewService(ctx, option.WithHTTPClient(client))
}

// CustomSearch returns a Custom Search service client for the given user.
func (f *Factory) CustomSearch(ctx context.Context, userEmail string) (*customsearch.Service, error) {
	client, err := f.clientFor(ctx, userEmail)
	if err != nil {
		return nil, fmt.Errorf("customsearch client for %s: %w", userEmail, err)
	}
	return customsearch.NewService(ctx, option.WithHTTPClient(client))
}

// Script returns an Apps Script service client for the given user.
func (f *Factory) Script(ctx context.Context, userEmail string) (*script.Service, error) {
	client, err := f.clientFor(ctx, userEmail)
	if err != nil {
		return nil, fmt.Errorf("script client for %s: %w", userEmail, err)
	}
	return script.NewService(ctx, option.WithHTTPClient(client))
}
