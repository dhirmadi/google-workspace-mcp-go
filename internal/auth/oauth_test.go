package auth

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"golang.org/x/oauth2"
)

func TestSignAndVerifyState(t *testing.T) {
	mgr := NewOAuthManager("client-id", "client-secret", "http://localhost/callback", false, []string{"scope"}, nil)

	tests := []struct {
		name  string
		email string
	}{
		{"simple email", "user@example.com"},
		{"email with dots", "first.last@example.com"},
		{"email with plus", "user+tag@example.com"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := mgr.GetAuthURL(tt.email)
			if url == "" {
				t.Fatal("expected non-empty auth URL")
			}

			state := mgr.signState(tt.email)
			email, ok := mgr.VerifyAndExtractEmail(state)
			if !ok {
				t.Fatal("expected valid state verification")
			}
			if email != tt.email {
				t.Errorf("expected email %q, got %q", tt.email, email)
			}
		})
	}
}

func TestVerifyAndExtractEmail_Invalid(t *testing.T) {
	mgr := NewOAuthManager("client-id", "client-secret", "http://localhost/callback", false, []string{"scope"}, nil)

	tests := []struct {
		name  string
		state string
	}{
		{"empty string", ""},
		{"no colon", "nocolonhere"},
		{"wrong signature", "user@example.com:deadbeef"},
		{"tampered email", "evil@attacker.com:" + mgr.hmacSign("user@example.com")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, ok := mgr.VerifyAndExtractEmail(tt.state)
			if ok {
				t.Error("expected verification to fail")
			}
		})
	}
}

func TestDifferentSecrets(t *testing.T) {
	mgr1 := NewOAuthManager("id", "secret1", "http://localhost/callback", false, nil, nil)
	mgr2 := NewOAuthManager("id", "secret2", "http://localhost/callback", false, nil, nil)

	state := mgr1.signState("user@example.com")

	// Same manager should verify
	if _, ok := mgr1.VerifyAndExtractEmail(state); !ok {
		t.Error("expected mgr1 to verify its own state")
	}

	// Different secret should fail
	if _, ok := mgr2.VerifyAndExtractEmail(state); ok {
		t.Error("expected mgr2 to reject mgr1's state")
	}
}

// TestAUTH01_PKCE_AuthURL_ContainsChallenge verifies AC3: public client auth URL includes S256 PKCE params.
func TestAUTH01_PKCE_AuthURL_ContainsChallenge(t *testing.T) {
	mgr := NewOAuthManager("client-id", "", "http://localhost/callback", true, []string{"scope"}, nil)
	url := mgr.GetAuthURL("user@example.com")

	if url == "" {
		t.Fatal("expected non-empty auth URL")
	}
	if !strings.Contains(url, "code_challenge=") {
		t.Errorf("expected code_challenge in PKCE auth URL, got: %s", url)
	}
	if !strings.Contains(url, "code_challenge_method=S256") {
		t.Errorf("expected code_challenge_method=S256 in PKCE auth URL, got: %s", url)
	}
}

// TestAUTH01_ConfidentialMode_NoPKCE verifies AC3: confidential mode auth URL has no PKCE params.
func TestAUTH01_ConfidentialMode_NoPKCE(t *testing.T) {
	mgr := NewOAuthManager("client-id", "secret", "http://localhost/callback", false, []string{"scope"}, nil)
	url := mgr.GetAuthURL("user@example.com")

	if strings.Contains(url, "code_challenge=") {
		t.Errorf("expected no code_challenge in confidential client auth URL, got: %s", url)
	}
	if strings.Contains(url, "code_verifier=") {
		t.Errorf("expected no code_verifier in confidential client auth URL, got: %s", url)
	}
}

// TestAUTH01_PKCE_VerifierStoredAfterAuthURL verifies that a PKCE verifier is stored per-user after GetAuthURL.
func TestAUTH01_PKCE_VerifierStoredAfterAuthURL(t *testing.T) {
	mgr := NewOAuthManager("client-id", "", "http://localhost/callback", true, []string{"scope"}, nil)
	mgr.GetAuthURL("user@example.com")

	if _, ok := mgr.pkceVerifiers.Load("user@example.com"); !ok {
		t.Error("expected PKCE verifier to be stored after GetAuthURL")
	}
}

// TestAUTH01_PKCE_Exchange_SendsVerifier verifies AC3: ExchangeCode sends code_verifier for public clients,
// and that the verifier matches the challenge from the auth URL (challenge = BASE64URL(SHA256(verifier))).
func TestAUTH01_PKCE_Exchange_SendsVerifier(t *testing.T) {
	var gotVerifier string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			t.Errorf("parse form: %v", err)
			return
		}
		gotVerifier = r.FormValue("code_verifier")
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"access_token":"tok","token_type":"Bearer","expires_in":3600}`)
	}))
	defer ts.Close()

	mgr := NewOAuthManager("client-id", "", "http://localhost/callback", true, []string{"scope"}, NewInMemoryTokenStore())
	// Override token endpoint to point at test server.
	mgr.config.Endpoint = oauth2.Endpoint{TokenURL: ts.URL}

	// GetAuthURL stores the PKCE verifier and embeds the challenge in the URL.
	authURL := mgr.GetAuthURL("user@example.com")

	// Extract code_challenge from the auth URL for later verification.
	parsed, err := url.Parse(authURL)
	if err != nil {
		t.Fatalf("parsing auth URL: %v", err)
	}
	wantChallenge := parsed.Query().Get("code_challenge")
	if wantChallenge == "" {
		t.Fatal("expected code_challenge in auth URL")
	}

	_, err = mgr.ExchangeCode(context.Background(), "test-code", "user@example.com")
	if err != nil {
		t.Fatalf("ExchangeCode error: %v", err)
	}
	if gotVerifier == "" {
		t.Fatal("expected code_verifier in token request for public client mode")
	}

	// Verify challenge/verifier consistency: challenge == BASE64URL(SHA256(verifier))
	sum := sha256.Sum256([]byte(gotVerifier))
	gotChallenge := base64.RawURLEncoding.EncodeToString(sum[:])
	if gotChallenge != wantChallenge {
		t.Errorf("challenge/verifier mismatch: recomputed challenge %q, auth URL had %q", gotChallenge, wantChallenge)
	}
}

// TestAUTH01_PKCE_ExchangeWithoutPriorAuth_Fails verifies that Exchange fails if no verifier was stored.
func TestAUTH01_PKCE_ExchangeWithoutPriorAuth_Fails(t *testing.T) {
	mgr := NewOAuthManager("client-id", "", "http://localhost/callback", true, []string{"scope"}, NewInMemoryTokenStore())

	_, err := mgr.ExchangeCode(context.Background(), "test-code", "never-authed@example.com")
	if err == nil {
		t.Error("expected error when no PKCE verifier is stored for user")
	}
}

// TestAUTH01_PKCE_VerifierClearedAfterExchange verifies verifier is consumed (one-use) after exchange.
func TestAUTH01_PKCE_VerifierClearedAfterExchange(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"access_token":"tok","token_type":"Bearer","expires_in":3600}`)
	}))
	defer ts.Close()

	mgr := NewOAuthManager("client-id", "", "http://localhost/callback", true, []string{"scope"}, NewInMemoryTokenStore())
	mgr.config.Endpoint = oauth2.Endpoint{TokenURL: ts.URL}

	mgr.GetAuthURL("user@example.com")
	_, _ = mgr.ExchangeCode(context.Background(), "test-code", "user@example.com")

	// Second exchange for same user should fail (verifier consumed).
	_, err := mgr.ExchangeCode(context.Background(), "test-code-2", "user@example.com")
	if err == nil {
		t.Error("expected error on second exchange attempt — verifier should be consumed")
	}
}

// TestAUTH01_PublicClient_StateVerifiesCorrectly verifies AC2: state HMAC works even without a client secret.
func TestAUTH01_PublicClient_StateVerifiesCorrectly(t *testing.T) {
	mgr := NewOAuthManager("client-id", "", "http://localhost/callback", true, []string{"scope"}, nil)
	email := "pub@example.com"

	state := mgr.signState(email)
	got, ok := mgr.VerifyAndExtractEmail(state)
	if !ok {
		t.Fatal("expected state verification to succeed for public client (random HMAC key)")
	}
	if got != email {
		t.Errorf("expected email %q, got %q", email, got)
	}
}
