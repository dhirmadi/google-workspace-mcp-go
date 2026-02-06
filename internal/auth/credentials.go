package auth

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"golang.org/x/oauth2"
)

// TokenStore handles persisting and loading OAuth tokens per user.
type TokenStore interface {
	Save(userEmail string, token *oauth2.Token) error
	Load(userEmail string) (*oauth2.Token, error)
}

// FileTokenStore stores tokens as JSON files on disk.
// Directory permissions: 0700. File permissions: 0600.
type FileTokenStore struct {
	dir string
}

// NewFileTokenStore creates a token store at the given directory path.
// The directory is created with 0700 permissions if it doesn't exist.
func NewFileTokenStore(dir string) (*FileTokenStore, error) {
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return nil, fmt.Errorf("creating credentials directory %s: %w", dir, err)
	}

	// Verify permissions are correct
	info, err := os.Stat(dir)
	if err != nil {
		return nil, fmt.Errorf("checking credentials directory %s: %w", dir, err)
	}
	if perm := info.Mode().Perm(); perm != 0o700 {
		slog.Warn("credentials directory has open permissions — should be 0700",
			"dir", dir,
			"perm", fmt.Sprintf("%04o", perm),
		)
	}

	return &FileTokenStore{dir: dir}, nil
}

// Save persists a token for the given user email.
func (s *FileTokenStore) Save(userEmail string, token *oauth2.Token) error {
	data, err := json.Marshal(token)
	if err != nil {
		return fmt.Errorf("marshaling token: %w", err)
	}
	path := s.tokenPath(userEmail)
	if err := os.WriteFile(path, data, 0o600); err != nil {
		return fmt.Errorf("writing token to %s: %w", path, err)
	}
	return nil
}

// Load reads a token for the given user email.
func (s *FileTokenStore) Load(userEmail string) (*oauth2.Token, error) {
	path := s.tokenPath(userEmail)
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("no credentials found for %s — call start_google_auth to authenticate", userEmail)
		}
		return nil, fmt.Errorf("reading token from %s: %w", path, err)
	}

	var token oauth2.Token
	if err := json.Unmarshal(data, &token); err != nil {
		return nil, fmt.Errorf("parsing token for %s: %w", userEmail, err)
	}
	return &token, nil
}

func (s *FileTokenStore) tokenPath(userEmail string) string {
	return filepath.Join(s.dir, userEmail+".json")
}

// PersistingTokenSource wraps an oauth2.TokenSource to persist refreshed tokens to disk.
type PersistingTokenSource struct {
	Base      oauth2.TokenSource
	Store     TokenStore
	UserEmail string
}

// Token returns a token, persisting any newly refreshed token to the store.
func (p *PersistingTokenSource) Token() (*oauth2.Token, error) {
	token, err := p.Base.Token()
	if err != nil {
		return nil, err
	}
	// Persist on refresh — best effort, don't fail the request
	if err := p.Store.Save(p.UserEmail, token); err != nil {
		slog.Warn("failed to persist refreshed token",
			"email", p.UserEmail,
			"error", err,
		)
	}
	return token, nil
}
