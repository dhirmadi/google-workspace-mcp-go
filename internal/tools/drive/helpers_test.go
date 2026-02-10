package drive

import (
	"testing"

	gdrive "google.golang.org/api/drive/v3"
)

func TestFormatFileType(t *testing.T) {
	tests := []struct {
		mime string
		want string
	}{
		{"application/vnd.google-apps.document", "Google Doc"},
		{"application/vnd.google-apps.spreadsheet", "Google Sheet"},
		{"application/vnd.google-apps.presentation", "Google Slides"},
		{"application/vnd.google-apps.folder", "Folder"},
		{"application/pdf", "PDF"},
		{"image/png", "Image"},
		{"video/mp4", "Video"},
		{"audio/mp3", "Audio"},
		{"text/plain", "text/plain"},
	}

	for _, tt := range tests {
		got := formatFileType(tt.mime)
		if got != tt.want {
			t.Errorf("formatFileType(%q) = %q, want %q", tt.mime, got, tt.want)
		}
	}
}

func TestFormatSize(t *testing.T) {
	tests := []struct {
		bytes int64
		want  string
	}{
		{0, ""},
		{500, "500 B"},
		{1024, "1.0 KB"},
		{1536, "1.5 KB"},
		{1048576, "1.0 MB"},
	}

	for _, tt := range tests {
		got := formatSize(tt.bytes)
		if got != tt.want {
			t.Errorf("formatSize(%d) = %q, want %q", tt.bytes, got, tt.want)
		}
	}
}

func TestFileToSummary(t *testing.T) {
	f := &gdrive.File{
		Id:           "file123",
		Name:         "test.doc",
		MimeType:     "application/vnd.google-apps.document",
		Size:         1024,
		ModifiedTime: "2025-01-01T00:00:00Z",
		WebViewLink:  "https://docs.google.com/document/d/file123",
	}

	s := fileToSummary(f)
	if s.ID != "file123" {
		t.Errorf("ID = %q, want %q", s.ID, "file123")
	}
	if s.Name != "test.doc" {
		t.Errorf("Name = %q, want %q", s.Name, "test.doc")
	}
}

func TestFormatPermission(t *testing.T) {
	tests := []struct {
		perm *gdrive.Permission
		want string
	}{
		{
			perm: &gdrive.Permission{Type: "user", Role: "writer", DisplayName: "Alice", EmailAddress: "alice@example.com"},
			want: "Alice (alice@example.com) — writer",
		},
		{
			perm: &gdrive.Permission{Type: "anyone", Role: "reader"},
			want: "Anyone with the link — reader",
		},
		{
			perm: &gdrive.Permission{Type: "domain", Role: "reader", Domain: "example.com"},
			want: "Domain: example.com — reader",
		},
	}

	for _, tt := range tests {
		got := formatPermission(tt.perm)
		if got != tt.want {
			t.Errorf("formatPermission() = %q, want %q", got, tt.want)
		}
	}
}

func TestIsGoogleNativeType(t *testing.T) {
	if !isGoogleNativeType("application/vnd.google-apps.document") {
		t.Error("expected Google Doc to be native type")
	}
	if isGoogleNativeType("application/pdf") {
		t.Error("expected PDF to NOT be native type")
	}
}

func TestMimeTypeForExport(t *testing.T) {
	got := mimeTypeForExport("application/vnd.google-apps.document")
	if got != "text/plain" {
		t.Errorf("got %q, want %q", got, "text/plain")
	}
	got = mimeTypeForExport("text/plain")
	if got != "" {
		t.Errorf("got %q, want empty for non-google type", got)
	}
}
