package gmail

import (
	"encoding/base64"
	"strings"
	"testing"

	"google.golang.org/api/gmail/v1"
)

func TestExtractHeader(t *testing.T) {
	msg := &gmail.Message{
		Payload: &gmail.MessagePart{
			Headers: []*gmail.MessagePartHeader{
				{Name: "Subject", Value: "Test Subject"},
				{Name: "From", Value: "alice@example.com"},
				{Name: "To", Value: "bob@example.com"},
			},
		},
	}

	tests := []struct {
		name string
		want string
	}{
		{"Subject", "Test Subject"},
		{"From", "alice@example.com"},
		{"subject", "Test Subject"}, // case-insensitive
		{"Missing", ""},
	}

	for _, tt := range tests {
		got := extractHeader(msg, tt.name)
		if got != tt.want {
			t.Errorf("extractHeader(%q) = %q, want %q", tt.name, got, tt.want)
		}
	}
}

func TestExtractHeaderNilPayload(t *testing.T) {
	msg := &gmail.Message{}
	if got := extractHeader(msg, "Subject"); got != "" {
		t.Errorf("expected empty string for nil payload, got %q", got)
	}
}

func TestExtractBody(t *testing.T) {
	plainText := base64.URLEncoding.EncodeToString([]byte("Hello, plain text!"))
	msg := &gmail.Message{
		Payload: &gmail.MessagePart{
			MimeType: "multipart/alternative",
			Parts: []*gmail.MessagePart{
				{
					MimeType: "text/plain",
					Body:     &gmail.MessagePartBody{Data: plainText},
				},
				{
					MimeType: "text/html",
					Body:     &gmail.MessagePartBody{Data: base64.URLEncoding.EncodeToString([]byte("<b>HTML</b>"))},
				},
			},
		},
	}

	got := extractBody(msg)
	if got != "Hello, plain text!" {
		t.Errorf("extractBody() = %q, want %q", got, "Hello, plain text!")
	}
}

func TestExtractBodyHTMLFallback(t *testing.T) {
	htmlContent := base64.URLEncoding.EncodeToString([]byte("<p>Hello from <b>HTML</b></p>"))
	msg := &gmail.Message{
		Payload: &gmail.MessagePart{
			MimeType: "text/html",
			Body:     &gmail.MessagePartBody{Data: htmlContent},
		},
	}

	got := extractBody(msg)
	if !strings.Contains(got, "Hello from HTML") {
		t.Errorf("extractBody() = %q, expected it to contain %q", got, "Hello from HTML")
	}
}

func TestMessageToSummary(t *testing.T) {
	msg := &gmail.Message{
		Id:       "msg123",
		ThreadId: "thread456",
		Snippet:  "Preview text...",
		LabelIds: []string{"INBOX", "UNREAD"},
		Payload: &gmail.MessagePart{
			Headers: []*gmail.MessagePartHeader{
				{Name: "Subject", Value: "Test"},
				{Name: "From", Value: "alice@example.com"},
			},
		},
	}

	summary := messageToSummary(msg)
	if summary.ID != "msg123" {
		t.Errorf("ID = %q, want %q", summary.ID, "msg123")
	}
	if summary.Subject != "Test" {
		t.Errorf("Subject = %q, want %q", summary.Subject, "Test")
	}
	if len(summary.LabelIDs) != 2 {
		t.Errorf("LabelIDs length = %d, want 2", len(summary.LabelIDs))
	}
}

func TestBuildRawMessage(t *testing.T) {
	raw := buildRawMessage(
		"bob@example.com",
		"Test Subject",
		"Hello Bob!",
		"cc@example.com",
		"",
		"",
		"<original@gmail.com>",
		"<original@gmail.com>",
	)

	decoded, err := base64.URLEncoding.DecodeString(raw)
	if err != nil {
		t.Fatalf("decoding raw message: %v", err)
	}

	msg := string(decoded)
	if !strings.Contains(msg, "To: bob@example.com") {
		t.Error("missing To header")
	}
	if !strings.Contains(msg, "Subject: Test Subject") {
		t.Error("missing Subject header")
	}
	if !strings.Contains(msg, "Cc: cc@example.com") {
		t.Error("missing Cc header")
	}
	if !strings.Contains(msg, "In-Reply-To: <original@gmail.com>") {
		t.Error("missing In-Reply-To header")
	}
	if !strings.Contains(msg, "Content-Transfer-Encoding: 8bit") {
		t.Error("missing Content-Transfer-Encoding header")
	}
	if !strings.Contains(msg, "Hello Bob!") {
		t.Error("missing body content")
	}
}

func TestBuildRawMessageMinimal(t *testing.T) {
	raw := buildRawMessage("bob@example.com", "Hi", "Body", "", "", "", "", "")
	decoded, err := base64.URLEncoding.DecodeString(raw)
	if err != nil {
		t.Fatalf("decoding raw message: %v", err)
	}

	msg := string(decoded)
	if strings.Contains(msg, "Cc:") {
		t.Error("should not contain Cc header when empty")
	}
	if strings.Contains(msg, "Bcc:") {
		t.Error("should not contain Bcc header when empty")
	}
	if strings.Contains(msg, "In-Reply-To:") {
		t.Error("should not contain In-Reply-To header when empty")
	}
}

func TestBuildRawMessageNonASCIISubject(t *testing.T) {
	raw := buildRawMessage("bob@example.com", "Héllo Wörld 日本語", "Body text", "", "", "", "", "")
	decoded, err := base64.URLEncoding.DecodeString(raw)
	if err != nil {
		t.Fatalf("decoding raw message: %v", err)
	}

	msg := string(decoded)
	// Subject should be RFC 2047 encoded, not raw UTF-8
	if strings.Contains(msg, "Subject: H\u00e9llo") {
		t.Error("non-ASCII subject should be RFC 2047 encoded, not raw UTF-8")
	}
	if !strings.Contains(msg, "Subject: =?UTF-8?q?") {
		t.Errorf("expected RFC 2047 Q-encoded subject, got: %s", msg)
	}
}

func TestBuildRawMessageASCIISubjectPassthrough(t *testing.T) {
	raw := buildRawMessage("bob@example.com", "Plain ASCII", "Body", "", "", "", "", "")
	decoded, err := base64.URLEncoding.DecodeString(raw)
	if err != nil {
		t.Fatalf("decoding raw message: %v", err)
	}

	msg := string(decoded)
	// Pure ASCII subjects should pass through as-is (no encoding needed)
	if !strings.Contains(msg, "Subject: Plain ASCII") {
		t.Errorf("ASCII subject should not be encoded, got: %s", msg)
	}
}

func TestBuildRawMessageUTF8Body(t *testing.T) {
	body := "H\u00e9llo, this has acc\u00e9nts and em\u2014dashes and \u201ccurly quotes\u201d"
	raw := buildRawMessage("bob@example.com", "Test", body, "", "", "", "", "")
	decoded, err := base64.URLEncoding.DecodeString(raw)
	if err != nil {
		t.Fatalf("decoding raw message: %v", err)
	}

	msg := string(decoded)
	if !strings.Contains(msg, body) {
		t.Error("UTF-8 body should be preserved in the raw message")
	}
	if !strings.Contains(msg, "Content-Transfer-Encoding: 8bit") {
		t.Error("UTF-8 body requires Content-Transfer-Encoding: 8bit")
	}
	if !strings.Contains(msg, `charset="UTF-8"`) {
		t.Error("must declare UTF-8 charset")
	}
}
