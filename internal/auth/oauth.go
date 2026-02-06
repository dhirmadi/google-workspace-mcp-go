package auth

import (
	"context"
	"fmt"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

// OAuthManager handles OAuth2 configuration and token exchange.
type OAuthManager struct {
	config     *oauth2.Config
	tokenStore TokenStore
}

// NewOAuthManager creates an OAuth manager with the given credentials.
func NewOAuthManager(clientID, clientSecret, redirectURL string, scopes []string, store TokenStore) *OAuthManager {
	return &OAuthManager{
		config: &oauth2.Config{
			ClientID:     clientID,
			ClientSecret: clientSecret,
			RedirectURL:  redirectURL,
			Scopes:       scopes,
			Endpoint:     google.Endpoint,
		},
		tokenStore: store,
	}
}

// GetAuthURL returns the URL for the user to authenticate.
func (m *OAuthManager) GetAuthURL(state string) string {
	return m.config.AuthCodeURL(state, oauth2.AccessTypeOffline, oauth2.ApprovalForce)
}

// ExchangeCode exchanges an authorization code for a token and persists it.
func (m *OAuthManager) ExchangeCode(ctx context.Context, code, userEmail string) (*oauth2.Token, error) {
	token, err := m.config.Exchange(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("exchanging auth code: %w", err)
	}
	if err := m.tokenStore.Save(userEmail, token); err != nil {
		return nil, fmt.Errorf("saving token for %s: %w", userEmail, err)
	}
	return token, nil
}

// Config returns the underlying oauth2.Config for building token sources.
func (m *OAuthManager) Config() *oauth2.Config {
	return m.config
}

// TokenStore returns the underlying token store.
func (m *OAuthManager) TokenStore() TokenStore {
	return m.tokenStore
}
