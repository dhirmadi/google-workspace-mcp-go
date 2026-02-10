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

func TestExtractAttachments(t *testing.T) {
	tests := []struct {
		name    string
		payload *gmail.MessagePart
		want    int // expected number of attachments
	}{
		{
			name: "no attachments",
			payload: &gmail.MessagePart{
				MimeType: "text/plain",
				Body:     &gmail.MessagePartBody{Data: "dGVzdA=="},
			},
			want: 0,
		},
		{
			name: "single attachment",
			payload: &gmail.MessagePart{
				MimeType: "multipart/mixed",
				Parts: []*gmail.MessagePart{
					{
						MimeType: "text/plain",
						Body:     &gmail.MessagePartBody{Data: "dGVzdA=="},
					},
					{
						MimeType: "application/pdf",
						Filename: "report.pdf",
						Body: &gmail.MessagePartBody{
							AttachmentId: "att-1",
							Size:         12345,
						},
					},
				},
			},
			want: 1,
		},
		{
			name: "multiple nested attachments",
			payload: &gmail.MessagePart{
				MimeType: "multipart/mixed",
				Parts: []*gmail.MessagePart{
					{
						MimeType: "multipart/alternative",
						Parts: []*gmail.MessagePart{
							{MimeType: "text/plain", Body: &gmail.MessagePartBody{Data: "dGVzdA=="}},
							{MimeType: "text/html", Body: &gmail.MessagePartBody{Data: "PGI+dGVzdDwvYj4="}},
						},
					},
					{
						MimeType: "image/png",
						Filename: "screenshot.png",
						Body: &gmail.MessagePartBody{
							AttachmentId: "att-2",
							Size:         54321,
						},
					},
					{
						MimeType: "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
						Filename: "doc.docx",
						Body: &gmail.MessagePartBody{
							AttachmentId: "att-3",
							Size:         99999,
						},
					},
				},
			},
			want: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractAttachments(tt.payload)
			if len(got) != tt.want {
				t.Errorf("extractAttachments() returned %d attachments, want %d", len(got), tt.want)
			}
		})
	}
}

func TestExtractAttachmentsFields(t *testing.T) {
	payload := &gmail.MessagePart{
		MimeType: "multipart/mixed",
		Parts: []*gmail.MessagePart{
			{
				MimeType: "application/pdf",
				Filename: "invoice.pdf",
				Body: &gmail.MessagePartBody{
					AttachmentId: "att-abc-123",
					Size:         42000,
				},
			},
		},
	}

	attachments := extractAttachments(payload)
	if len(attachments) != 1 {
		t.Fatalf("expected 1 attachment, got %d", len(attachments))
	}

	a := attachments[0]
	if a.AttachmentID != "att-abc-123" {
		t.Errorf("AttachmentID = %q, want %q", a.AttachmentID, "att-abc-123")
	}
	if a.Filename != "invoice.pdf" {
		t.Errorf("Filename = %q, want %q", a.Filename, "invoice.pdf")
	}
	if a.MimeType != "application/pdf" {
		t.Errorf("MimeType = %q, want %q", a.MimeType, "application/pdf")
	}
	if a.Size != 42000 {
		t.Errorf("Size = %d, want %d", a.Size, 42000)
	}
}

func TestFindAttachmentPart(t *testing.T) {
	payload := &gmail.MessagePart{
		MimeType: "multipart/mixed",
		Parts: []*gmail.MessagePart{
			{MimeType: "text/plain", Body: &gmail.MessagePartBody{Data: "dGVzdA=="}},
			{
				MimeType: "image/jpeg",
				Filename: "photo.jpg",
				Body: &gmail.MessagePartBody{
					AttachmentId: "target-id",
					Size:         8000,
				},
			},
			{
				MimeType: "application/pdf",
				Filename: "doc.pdf",
				Body: &gmail.MessagePartBody{
					AttachmentId: "other-id",
					Size:         5000,
				},
			},
		},
	}

	t.Run("found", func(t *testing.T) {
		info := findAttachmentPart(payload, "target-id")
		if info == nil {
			t.Fatal("expected to find attachment, got nil")
		}
		if info.Filename != "photo.jpg" {
			t.Errorf("Filename = %q, want %q", info.Filename, "photo.jpg")
		}
		if info.MimeType != "image/jpeg" {
			t.Errorf("MimeType = %q, want %q", info.MimeType, "image/jpeg")
		}
	})

	t.Run("not found", func(t *testing.T) {
		info := findAttachmentPart(payload, "nonexistent-id")
		if info != nil {
			t.Errorf("expected nil for nonexistent ID, got %+v", info)
		}
	})
}

func TestFormatAttachmentSize(t *testing.T) {
	tests := []struct {
		bytes int64
		want  string
	}{
		{0, "unknown size"},
		{500, "500 B"},
		{1024, "1.0 KB"},
		{1536, "1.5 KB"},
		{1048576, "1.0 MB"},
		{5242880, "5.0 MB"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := formatAttachmentSize(tt.bytes)
			if got != tt.want {
				t.Errorf("formatAttachmentSize(%d) = %q, want %q", tt.bytes, got, tt.want)
			}
		})
	}
}

func TestIsOfficeType(t *testing.T) {
	tests := []struct {
		mimeType string
		want     bool
	}{
		{"application/vnd.openxmlformats-officedocument.wordprocessingml.document", true},
		{"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", true},
		{"application/vnd.openxmlformats-officedocument.presentationml.presentation", true},
		{"application/pdf", false},
		{"text/plain", false},
		{"image/png", false},
	}

	for _, tt := range tests {
		t.Run(tt.mimeType, func(t *testing.T) {
			got := isOfficeType(tt.mimeType)
			if got != tt.want {
				t.Errorf("isOfficeType(%q) = %v, want %v", tt.mimeType, got, tt.want)
			}
		})
	}
}

func TestMessageToDetailWithAttachments(t *testing.T) {
	plainText := base64.URLEncoding.EncodeToString([]byte("Body text"))
	msg := &gmail.Message{
		Id:       "msg-with-att",
		ThreadId: "thread-1",
		Payload: &gmail.MessagePart{
			MimeType: "multipart/mixed",
			Headers: []*gmail.MessagePartHeader{
				{Name: "Subject", Value: "Has Attachment"},
				{Name: "From", Value: "sender@example.com"},
				{Name: "To", Value: "receiver@example.com"},
			},
			Parts: []*gmail.MessagePart{
				{
					MimeType: "text/plain",
					Body:     &gmail.MessagePartBody{Data: plainText},
				},
				{
					MimeType: "application/pdf",
					Filename: "report.pdf",
					Body: &gmail.MessagePartBody{
						AttachmentId: "att-pdf-1",
						Size:         10240,
					},
				},
			},
		},
	}

	detail := messageToDetail(msg)

	if detail.Subject != "Has Attachment" {
		t.Errorf("Subject = %q, want %q", detail.Subject, "Has Attachment")
	}
	if detail.Body != "Body text" {
		t.Errorf("Body = %q, want %q", detail.Body, "Body text")
	}
	if len(detail.Attachments) != 1 {
		t.Fatalf("Attachments count = %d, want 1", len(detail.Attachments))
	}
	att := detail.Attachments[0]
	if att.AttachmentID != "att-pdf-1" {
		t.Errorf("AttachmentID = %q, want %q", att.AttachmentID, "att-pdf-1")
	}
	if att.Filename != "report.pdf" {
		t.Errorf("Filename = %q, want %q", att.Filename, "report.pdf")
	}
}

func TestMessageToDetailWithoutAttachments(t *testing.T) {
	plainText := base64.URLEncoding.EncodeToString([]byte("Just text"))
	msg := &gmail.Message{
		Id:       "msg-no-att",
		ThreadId: "thread-2",
		Payload: &gmail.MessagePart{
			MimeType: "text/plain",
			Headers: []*gmail.MessagePartHeader{
				{Name: "Subject", Value: "No Attachments"},
			},
			Body: &gmail.MessagePartBody{Data: plainText},
		},
	}

	detail := messageToDetail(msg)

	if len(detail.Attachments) != 0 {
		t.Errorf("expected no attachments, got %d", len(detail.Attachments))
	}
}
