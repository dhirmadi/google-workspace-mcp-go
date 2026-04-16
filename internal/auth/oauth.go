package auth

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"log/slog"
	"strings"
	"sync"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"

	"github.com/evert/google-workspace-mcp-go/internal/pkg/validate"
)

// OAuthManager handles OAuth2 configuration and token exchange.
type OAuthManager struct {
	config        *oauth2.Config
	tokenStore    TokenStore
	stateKey      []byte   // HMAC key for signing OAuth state
	publicClient  bool     // true = PKCE mode, no client secret required
	pkceVerifiers sync.Map // email (string) → PKCE verifier (string); one-use per auth attempt
}

// NewOAuthManager creates an OAuth manager with the given credentials.
// Set publicClient=true to enable PKCE-based flow without a client secret.
// When publicClient is false (confidential mode), clientSecret is used as the HMAC key for state signing.
// When publicClient is true, a random 32-byte key is generated for state signing.
func NewOAuthManager(clientID, clientSecret, redirectURL string, publicClient bool, scopes []string, store TokenStore) *OAuthManager {
	stateKey := []byte(clientSecret)
	if len(stateKey) == 0 {
		// Public client has no secret; generate an ephemeral random HMAC key for state signing.
		stateKey = mustGenerateRandomKey()
	}
	return &OAuthManager{
		config: &oauth2.Config{
			ClientID:     clientID,
			ClientSecret: clientSecret,
			RedirectURL:  redirectURL,
			Scopes:       scopes,
			Endpoint:     google.Endpoint,
		},
		tokenStore:   store,
		stateKey:     stateKey,
		publicClient: publicClient,
	}
}

// mustGenerateRandomKey creates a cryptographically random 32-byte key for HMAC state signing.
// Panics if the system PRNG is unavailable (unrecoverable startup condition).
func mustGenerateRandomKey() []byte {
	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		panic("auth: cannot generate random OAuth state key: " + err.Error())
	}
	return key
}

// GetAuthURL returns the URL for the user to authenticate.
// The state parameter is the user email signed with HMAC to prevent CSRF.
// In public client mode, PKCE S256 parameters are appended and the verifier is stored for the exchange step.
func (m *OAuthManager) GetAuthURL(userEmail string) string {
	if err := validate.Email(userEmail); err != nil {
		return ""
	}
	state := m.signState(userEmail)
	opts := []oauth2.AuthCodeOption{oauth2.AccessTypeOffline, oauth2.ApprovalForce}

	if m.publicClient {
		verifier, err := generatePKCEVerifier()
		if err != nil {
			slog.Error("PKCE verifier generation failed — PRNG unavailable", "error", err)
			return ""
		}
		if _, overwritten := m.pkceVerifiers.Swap(userEmail, verifier); overwritten {
			slog.Warn("PKCE verifier overwritten — previous auth flow for this user was abandoned",
				"user_google_email", userEmail,
			)
		}
		challenge := pkceS256Challenge(verifier)
		opts = append(opts,
			oauth2.SetAuthURLParam("code_challenge", challenge),
			oauth2.SetAuthURLParam("code_challenge_method", "S256"),
		)
	}

	return m.config.AuthCodeURL(state, opts...)
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
// In public client mode, the stored PKCE verifier is consumed (one-use).
func (m *OAuthManager) ExchangeCode(ctx context.Context, code, userEmail string) (*oauth2.Token, error) {
	var opts []oauth2.AuthCodeOption

	if m.publicClient {
		v, ok := m.pkceVerifiers.LoadAndDelete(userEmail)
		if !ok {
			return nil, fmt.Errorf(
				"no PKCE verifier found for %s — restart authentication from the MCP client", userEmail,
			)
		}
		opts = append(opts, oauth2.SetAuthURLParam("code_verifier", v.(string)))
	}

	token, err := m.config.Exchange(ctx, code, opts...)
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

// generatePKCEVerifier creates a RFC 7636-compliant code verifier:
// 32 random bytes encoded as base64url without padding → 43 characters.
func generatePKCEVerifier() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generating PKCE verifier: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

// pkceS256Challenge computes the PKCE S256 code challenge from a verifier:
// BASE64URL(SHA256(verifier)) without padding.
func pkceS256Challenge(verifier string) string {
	sum := sha256.Sum256([]byte(verifier))
	return base64.RawURLEncoding.EncodeToString(sum[:])
}
