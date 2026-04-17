package config

import "testing"

// TestAUTH01_ConfidentialClientMode verifies AC1: ID + secret → success, PublicClient=false.
func TestAUTH01_ConfidentialClientMode(t *testing.T) {
	cfg := &Config{}
	cfg.OAuth.ClientID = "test-id"
	cfg.OAuth.ClientSecret = "test-secret"
	cfg.OAuth.PublicClient = false

	if err := validateOAuthMode(cfg); err != nil {
		t.Fatalf("expected success for confidential mode: %v", err)
	}
}

// TestAUTH01_PublicClientMode verifies AC2: public client mode works without a secret.
func TestAUTH01_PublicClientMode(t *testing.T) {
	cfg := &Config{}
	cfg.OAuth.ClientID = "test-id"
	cfg.OAuth.PublicClient = true
	// ClientSecret intentionally empty

	if err := validateOAuthMode(cfg); err != nil {
		t.Fatalf("expected success for public client mode: %v", err)
	}
}

// TestAUTH01_MissingSecret_ConfidentialMode verifies AC1: missing secret in confidential mode → error.
func TestAUTH01_MissingSecret_ConfidentialMode(t *testing.T) {
	cfg := &Config{}
	cfg.OAuth.ClientID = "test-id"
	cfg.OAuth.PublicClient = false
	// ClientSecret intentionally empty

	if err := validateOAuthMode(cfg); err == nil {
		t.Error("expected error when GOOGLE_OAUTH_CLIENT_SECRET is missing in confidential mode")
	}
}

// TestAUTH01_PublicClient_SecretIgnored verifies that a secret can coexist with public client mode
// without causing an error (operators may set both during migration).
func TestAUTH01_PublicClient_SecretIgnored(t *testing.T) {
	cfg := &Config{}
	cfg.OAuth.ClientID = "test-id"
	cfg.OAuth.ClientSecret = "test-secret"
	cfg.OAuth.PublicClient = true

	if err := validateOAuthMode(cfg); err != nil {
		t.Fatalf("expected success when both secret and public client flag are set: %v", err)
	}
}
