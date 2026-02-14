package auth

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"golang.org/x/oauth2"
)

func TestFileTokenStore_SaveAndLoad(t *testing.T) {
	dir := t.TempDir()
	store, err := NewFileTokenStore(dir)
	if err != nil {
		t.Fatalf("NewFileTokenStore: %v", err)
	}

	token := &oauth2.Token{
		AccessToken:  "access-123",
		RefreshToken: "refresh-456",
		TokenType:    "Bearer",
		Expiry:       time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
	}

	email := "test@example.com"

	if err := store.Save(email, token); err != nil {
		t.Fatalf("Save: %v", err)
	}

	loaded, err := store.Load(email)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	if loaded.AccessToken != token.AccessToken {
		t.Errorf("AccessToken: got %q, want %q", loaded.AccessToken, token.AccessToken)
	}
	if loaded.RefreshToken != token.RefreshToken {
		t.Errorf("RefreshToken: got %q, want %q", loaded.RefreshToken, token.RefreshToken)
	}
}

func TestFileTokenStore_LoadNonExistent(t *testing.T) {
	dir := t.TempDir()
	store, err := NewFileTokenStore(dir)
	if err != nil {
		t.Fatalf("NewFileTokenStore: %v", err)
	}

	_, err = store.Load("nobody@example.com")
	if err == nil {
		t.Fatal("expected error for non-existent token")
	}
}

func TestFileTokenStore_TokenPathUsesHash(t *testing.T) {
	dir := t.TempDir()
	store, err := NewFileTokenStore(dir)
	if err != nil {
		t.Fatalf("NewFileTokenStore: %v", err)
	}

	path := store.tokenPath("user@example.com")
	filename := filepath.Base(path)

	// Should be a hex-encoded SHA-256 hash (64 chars) + ".json"
	if len(filename) != 64+5 {
		t.Errorf("expected token filename of 69 chars, got %d: %s", len(filename), filename)
	}

	// Different emails should produce different paths
	path2 := store.tokenPath("other@example.com")
	if path == path2 {
		t.Error("expected different paths for different emails")
	}
}

func TestFileTokenStore_FilePermissions(t *testing.T) {
	dir := t.TempDir()
	store, err := NewFileTokenStore(dir)
	if err != nil {
		t.Fatalf("NewFileTokenStore: %v", err)
	}

	token := &oauth2.Token{AccessToken: "test", TokenType: "Bearer"}
	if err := store.Save("perm@test.com", token); err != nil {
		t.Fatalf("Save: %v", err)
	}

	path := store.tokenPath("perm@test.com")
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("Stat: %v", err)
	}

	perm := info.Mode().Perm()
	if perm != 0o600 {
		t.Errorf("expected file permission 0600, got %04o", perm)
	}
}

func TestPersistingTokenSource_PersistsOnChange(t *testing.T) {
	dir := t.TempDir()
	store, err := NewFileTokenStore(dir)
	if err != nil {
		t.Fatalf("NewFileTokenStore: %v", err)
	}

	email := "refresh@test.com"

	// Start with token "v1"
	initialToken := &oauth2.Token{
		AccessToken: "v1",
		TokenType:   "Bearer",
		Expiry:      time.Now().Add(time.Hour),
	}
	if err := store.Save(email, initialToken); err != nil {
		t.Fatalf("Save initial: %v", err)
	}

	// Create a PersistingTokenSource that will return "v2"
	newToken := &oauth2.Token{
		AccessToken: "v2",
		TokenType:   "Bearer",
		Expiry:      time.Now().Add(time.Hour),
	}
	pts := &PersistingTokenSource{
		Base:      oauth2.StaticTokenSource(newToken),
		Store:     store,
		UserEmail: email,
	}

	// First call should persist the changed token
	got, err := pts.Token()
	if err != nil {
		t.Fatalf("Token: %v", err)
	}
	if got.AccessToken != "v2" {
		t.Errorf("expected AccessToken v2, got %s", got.AccessToken)
	}

	// Verify it was persisted
	loaded, err := store.Load(email)
	if err != nil {
		t.Fatalf("Load after refresh: %v", err)
	}
	if loaded.AccessToken != "v2" {
		t.Errorf("persisted token should be v2, got %s", loaded.AccessToken)
	}

	// Second call with same token should NOT write again (we can verify by checking the function completes)
	got2, err := pts.Token()
	if err != nil {
		t.Fatalf("Token second call: %v", err)
	}
	if got2.AccessToken != "v2" {
		t.Errorf("expected AccessToken v2 on second call, got %s", got2.AccessToken)
	}
}

// ── InMemoryTokenStore tests ────────────────────────────────────────

func TestInMemoryTokenStore_SaveAndLoad(t *testing.T) {
	store := NewInMemoryTokenStore()

	token := &oauth2.Token{
		AccessToken:  "mem-access-123",
		RefreshToken: "mem-refresh-456",
		TokenType:    "Bearer",
		Expiry:       time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC),
	}

	email := "inmem@example.com"

	if err := store.Save(email, token); err != nil {
		t.Fatalf("Save: %v", err)
	}

	loaded, err := store.Load(email)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	if loaded.AccessToken != token.AccessToken {
		t.Errorf("AccessToken: got %q, want %q", loaded.AccessToken, token.AccessToken)
	}
	if loaded.RefreshToken != token.RefreshToken {
		t.Errorf("RefreshToken: got %q, want %q", loaded.RefreshToken, token.RefreshToken)
	}
}

func TestInMemoryTokenStore_LoadNonExistent(t *testing.T) {
	store := NewInMemoryTokenStore()

	_, err := store.Load("nobody@example.com")
	if err == nil {
		t.Fatal("expected error for non-existent token")
	}
}

func TestInMemoryTokenStore_Overwrite(t *testing.T) {
	store := NewInMemoryTokenStore()
	email := "overwrite@example.com"

	token1 := &oauth2.Token{AccessToken: "v1", TokenType: "Bearer"}
	token2 := &oauth2.Token{AccessToken: "v2", TokenType: "Bearer"}

	if err := store.Save(email, token1); err != nil {
		t.Fatalf("Save v1: %v", err)
	}
	if err := store.Save(email, token2); err != nil {
		t.Fatalf("Save v2: %v", err)
	}

	loaded, err := store.Load(email)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if loaded.AccessToken != "v2" {
		t.Errorf("expected v2 after overwrite, got %s", loaded.AccessToken)
	}
}

func TestInMemoryTokenStore_MultipleUsers(t *testing.T) {
	store := NewInMemoryTokenStore()

	users := map[string]*oauth2.Token{
		"alice@example.com": {AccessToken: "alice-token", TokenType: "Bearer"},
		"bob@example.com":   {AccessToken: "bob-token", TokenType: "Bearer"},
		"carol@example.com": {AccessToken: "carol-token", TokenType: "Bearer"},
	}

	for email, token := range users {
		if err := store.Save(email, token); err != nil {
			t.Fatalf("Save %s: %v", email, err)
		}
	}

	for email, want := range users {
		got, err := store.Load(email)
		if err != nil {
			t.Fatalf("Load %s: %v", email, err)
		}
		if got.AccessToken != want.AccessToken {
			t.Errorf("user %s: got %q, want %q", email, got.AccessToken, want.AccessToken)
		}
	}
}

func TestInMemoryTokenStore_ConcurrentAccess(t *testing.T) {
	store := NewInMemoryTokenStore()

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			email := "concurrent@example.com"
			token := &oauth2.Token{
				AccessToken: fmt.Sprintf("token-%d", n),
				TokenType:   "Bearer",
			}
			if err := store.Save(email, token); err != nil {
				t.Errorf("concurrent Save: %v", err)
			}
			_, _ = store.Load(email)
		}(i)
	}
	wg.Wait()

	// After all goroutines, we should be able to load a valid token
	loaded, err := store.Load("concurrent@example.com")
	if err != nil {
		t.Fatalf("Load after concurrent writes: %v", err)
	}
	if loaded.AccessToken == "" {
		t.Error("expected non-empty access token")
	}
}

func TestPersistingTokenSource_WithInMemoryStore(t *testing.T) {
	store := NewInMemoryTokenStore()
	email := "pts-inmem@example.com"

	initial := &oauth2.Token{AccessToken: "old", TokenType: "Bearer"}
	if err := store.Save(email, initial); err != nil {
		t.Fatalf("Save initial: %v", err)
	}

	refreshed := &oauth2.Token{
		AccessToken: "refreshed",
		TokenType:   "Bearer",
		Expiry:      time.Now().Add(time.Hour),
	}
	pts := &PersistingTokenSource{
		Base:      oauth2.StaticTokenSource(refreshed),
		Store:     store,
		UserEmail: email,
	}

	got, err := pts.Token()
	if err != nil {
		t.Fatalf("Token: %v", err)
	}
	if got.AccessToken != "refreshed" {
		t.Errorf("expected 'refreshed', got %s", got.AccessToken)
	}

	// Verify the refreshed token was persisted in the in-memory store
	loaded, err := store.Load(email)
	if err != nil {
		t.Fatalf("Load after refresh: %v", err)
	}
	if loaded.AccessToken != "refreshed" {
		t.Errorf("persisted token should be 'refreshed', got %s", loaded.AccessToken)
	}
}
