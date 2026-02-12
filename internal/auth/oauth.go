package auth

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

// OAuthManager handles OAuth2 configuration and token exchange.
type OAuthManager struct {
	config     *oauth2.Config
	tokenStore TokenStore
	stateKey   []byte // HMAC key for signing OAuth state
}

// NewOAuthManager creates an OAuth manager with the given credentials.
// The client secret is reused as the HMAC key for OAuth state signing.
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
		stateKey:   []byte(clientSecret),
	}
}

// GetAuthURL returns the URL for the user to authenticate.
// The state parameter is the user email signed with HMAC to prevent CSRF.
func (m *OAuthManager) GetAuthURL(userEmail string) string {
	state := m.signState(userEmail)
	return m.config.AuthCodeURL(state, oauth2.AccessTypeOffline, oauth2.ApprovalForce)
}

// VerifyAndExtractEmail verifies the HMAC-signed state parameter and extracts
// the user email. Returns the email and true if valid, or ("", false) if not.
func (m *OAuthManager) VerifyAndExtractEmail(state string) (string, bool) {
	// State format: "email:hex-signature"
	parts := strings.SplitN(state, ":", 2)
	if len(parts) != 2 {
		return "", false
	}
	email := parts[0]
	providedSig := parts[1]

	expectedSig := m.hmacSign(email)
	if !hmac.Equal([]byte(providedSig), []byte(expectedSig)) {
		return "", false
	}
	return email, true
}

// signState creates an HMAC-signed state string: "email:hex-signature"
func (m *OAuthManager) signState(email string) string {
	return email + ":" + m.hmacSign(email)
}

// hmacSign returns the hex-encoded HMAC-SHA256 of the given data.
func (m *OAuthManager) hmacSign(data string) string {
	mac := hmac.New(sha256.New, m.stateKey)
	mac.Write([]byte(data))
	return hex.EncodeToString(mac.Sum(nil))
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
