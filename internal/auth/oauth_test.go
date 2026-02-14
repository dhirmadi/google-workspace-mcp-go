package auth

import (
	"testing"
)

func TestSignAndVerifyState(t *testing.T) {
	mgr := NewOAuthManager("client-id", "client-secret", "http://localhost/callback", []string{"scope"}, nil)

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
	mgr := NewOAuthManager("client-id", "client-secret", "http://localhost/callback", []string{"scope"}, nil)

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
	mgr1 := NewOAuthManager("id", "secret1", "http://localhost/callback", nil, nil)
	mgr2 := NewOAuthManager("id", "secret2", "http://localhost/callback", nil, nil)

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
