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
	if !strings.Contains(msg, "Hello Bob!") {
		t.Error("missing body content")
	}
}

func TestExtractAttachments(t *testing.T) {
	msg := &gmail.MessagePart{
		MimeType: "multipart/mixed",
		Parts: []*gmail.MessagePart{
			{
				MimeType: "text/plain",
				Body:     &gmail.MessagePartBody{Data: base64.URLEncoding.EncodeToString([]byte("Hello"))},
			},
			{
				MimeType: "application/pdf",
				Filename: "report.pdf",
				Body: &gmail.MessagePartBody{
					AttachmentId: "att-123",
					Size:         1024,
				},
			},
			{
				MimeType: "image/png",
				Filename: "screenshot.png",
				Body: &gmail.MessagePartBody{
					AttachmentId: "att-456",
					Size:         2048,
				},
			},
		},
	}

	attachments := extractAttachments(msg)
	if len(attachments) != 2 {
		t.Fatalf("expected 2 attachments, got %d", len(attachments))
	}

	if attachments[0].Filename != "report.pdf" {
		t.Errorf("attachment[0].Filename = %q, want %q", attachments[0].Filename, "report.pdf")
	}
	if attachments[0].AttachmentID != "att-123" {
		t.Errorf("attachment[0].AttachmentID = %q, want %q", attachments[0].AttachmentID, "att-123")
	}
	if attachments[0].MimeType != "application/pdf" {
		t.Errorf("attachment[0].MimeType = %q, want %q", attachments[0].MimeType, "application/pdf")
	}
	if attachments[0].Size != 1024 {
		t.Errorf("attachment[0].Size = %d, want %d", attachments[0].Size, 1024)
	}

	if attachments[1].Filename != "screenshot.png" {
		t.Errorf("attachment[1].Filename = %q, want %q", attachments[1].Filename, "screenshot.png")
	}
	if attachments[1].AttachmentID != "att-456" {
		t.Errorf("attachment[1].AttachmentID = %q, want %q", attachments[1].AttachmentID, "att-456")
	}
}

func TestExtractAttachmentsNested(t *testing.T) {
	msg := &gmail.MessagePart{
		MimeType: "multipart/mixed",
		Parts: []*gmail.MessagePart{
			{
				MimeType: "multipart/alternative",
				Parts: []*gmail.MessagePart{
					{
						MimeType: "text/plain",
						Body:     &gmail.MessagePartBody{Data: base64.URLEncoding.EncodeToString([]byte("Body"))},
					},
				},
			},
			{
				MimeType: "application/zip",
				Filename: "archive.zip",
				Body: &gmail.MessagePartBody{
					AttachmentId: "att-nested",
					Size:         4096,
				},
			},
		},
	}

	attachments := extractAttachments(msg)
	if len(attachments) != 1 {
		t.Fatalf("expected 1 attachment, got %d", len(attachments))
	}
	if attachments[0].Filename != "archive.zip" {
		t.Errorf("Filename = %q, want %q", attachments[0].Filename, "archive.zip")
	}
}

func TestExtractAttachmentsNone(t *testing.T) {
	msg := &gmail.MessagePart{
		MimeType: "text/plain",
		Body:     &gmail.MessagePartBody{Data: base64.URLEncoding.EncodeToString([]byte("No attachments"))},
	}

	attachments := extractAttachments(msg)
	if len(attachments) != 0 {
		t.Errorf("expected 0 attachments, got %d", len(attachments))
	}
}

func TestMessageToDetailWithAttachments(t *testing.T) {
	msg := &gmail.Message{
		Id:       "msg-with-att",
		ThreadId: "thread-att",
		LabelIds: []string{"INBOX"},
		Payload: &gmail.MessagePart{
			MimeType: "multipart/mixed",
			Headers: []*gmail.MessagePartHeader{
				{Name: "Subject", Value: "File attached"},
				{Name: "From", Value: "sender@example.com"},
				{Name: "To", Value: "recipient@example.com"},
			},
			Parts: []*gmail.MessagePart{
				{
					MimeType: "text/plain",
					Body:     &gmail.MessagePartBody{Data: base64.URLEncoding.EncodeToString([]byte("See attached"))},
				},
				{
					MimeType: "application/pdf",
					Filename: "doc.pdf",
					Body: &gmail.MessagePartBody{
						AttachmentId: "att-doc",
						Size:         512,
					},
				},
			},
		},
	}

	detail := messageToDetail(msg)
	if len(detail.Attachments) != 1 {
		t.Fatalf("expected 1 attachment, got %d", len(detail.Attachments))
	}
	if detail.Attachments[0].Filename != "doc.pdf" {
		t.Errorf("Filename = %q, want %q", detail.Attachments[0].Filename, "doc.pdf")
	}
	if detail.Body != "See attached" {
		t.Errorf("Body = %q, want %q", detail.Body, "See attached")
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
