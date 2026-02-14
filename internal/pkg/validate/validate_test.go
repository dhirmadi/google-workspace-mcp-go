package validate

import "testing"

func TestDriveID(t *testing.T) {
	tests := []struct {
		name    string
		id      string
		wantErr bool
	}{
		{"root literal", "root", false},
		{"typical drive ID", "1BxiMVs0XRA5nFMdKvBdBZjgmUUqptlbs74OgVE2upms", false},
		{"short ID", "abc123", false},
		{"with hyphens", "abc-123_def", false},
		{"empty", "", true},
		{"single quote injection", "root' or name contains 'secret", true},
		{"spaces", "has spaces", true},
		{"special chars", "file/../../../etc", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := DriveID(tt.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("DriveID(%q) error = %v, wantErr %v", tt.id, err, tt.wantErr)
			}
		})
	}
}

func TestEmail(t *testing.T) {
	tests := []struct {
		name    string
		email   string
		wantErr bool
	}{
		{"valid email", "user@example.com", false},
		{"with dots", "first.last@example.com", false},
		{"with plus", "user+tag@example.com", false},
		{"with subdomain", "user@sub.example.com", false},
		{"empty", "", true},
		{"no at sign", "userexample.com", true},
		{"no domain", "user@", true},
		{"no TLD", "user@example", true},
		{"spaces", "user @example.com", true},
		{"arbitrary string", "not-an-email", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Email(tt.email)
			if (err != nil) != tt.wantErr {
				t.Errorf("Email(%q) error = %v, wantErr %v", tt.email, err, tt.wantErr)
			}
		})
	}
}
