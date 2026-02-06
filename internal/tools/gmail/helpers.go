package gmail

import (
	"encoding/base64"
	"fmt"
	"strings"

	"google.golang.org/api/gmail/v1"

	"github.com/evert/google-workspace-mcp-go/internal/pkg/htmlutil"
)

// MessageSummary is a compact representation of a Gmail message.
type MessageSummary struct {
	ID       string   `json:"id"`
	ThreadID string   `json:"thread_id"`
	Subject  string   `json:"subject,omitempty"`
	From     string   `json:"from,omitempty"`
	To       string   `json:"to,omitempty"`
	Date     string   `json:"date,omitempty"`
	Snippet  string   `json:"snippet,omitempty"`
	LabelIDs []string `json:"label_ids,omitempty"`
}

// MessageDetail is the full content of a Gmail message.
type MessageDetail struct {
	ID        string   `json:"id"`
	ThreadID  string   `json:"thread_id"`
	Subject   string   `json:"subject"`
	From      string   `json:"from"`
	To        string   `json:"to"`
	CC        string   `json:"cc,omitempty"`
	Date      string   `json:"date"`
	MessageID string   `json:"message_id,omitempty"`
	Body      string   `json:"body"`
	LabelIDs  []string `json:"label_ids,omitempty"`
}

// extractHeader returns the value of a named header from a Gmail message.
func extractHeader(msg *gmail.Message, name string) string {
	if msg.Payload == nil {
		return ""
	}
	for _, h := range msg.Payload.Headers {
		if strings.EqualFold(h.Name, name) {
			return h.Value
		}
	}
	return ""
}

// extractBody extracts the plain text body from a Gmail message.
// It prefers text/plain, falling back to text/html with HTML-to-text conversion.
func extractBody(msg *gmail.Message) string {
	if msg.Payload == nil {
		return ""
	}

	// Try to find plain text first
	if body := findBodyPart(msg.Payload, "text/plain"); body != "" {
		return body
	}

	// Fall back to HTML and convert
	if body := findBodyPart(msg.Payload, "text/html"); body != "" {
		return htmlutil.ToPlainText(body)
	}

	return ""
}

// findBodyPart recursively searches the message payload for a part with the given MIME type.
func findBodyPart(part *gmail.MessagePart, mimeType string) string {
	if part.MimeType == mimeType && part.Body != nil && part.Body.Data != "" {
		data, err := base64.URLEncoding.DecodeString(part.Body.Data)
		if err != nil {
			return ""
		}
		return string(data)
	}

	for _, child := range part.Parts {
		if body := findBodyPart(child, mimeType); body != "" {
			return body
		}
	}

	return ""
}

// messageToSummary converts a Gmail message to a compact summary.
func messageToSummary(msg *gmail.Message) MessageSummary {
	return MessageSummary{
		ID:       msg.Id,
		ThreadID: msg.ThreadId,
		Subject:  extractHeader(msg, "Subject"),
		From:     extractHeader(msg, "From"),
		To:       extractHeader(msg, "To"),
		Date:     extractHeader(msg, "Date"),
		Snippet:  msg.Snippet,
		LabelIDs: msg.LabelIds,
	}
}

// messageToDetail converts a Gmail message to full detail including body.
func messageToDetail(msg *gmail.Message) MessageDetail {
	return MessageDetail{
		ID:        msg.Id,
		ThreadID:  msg.ThreadId,
		Subject:   extractHeader(msg, "Subject"),
		From:      extractHeader(msg, "From"),
		To:        extractHeader(msg, "To"),
		CC:        extractHeader(msg, "Cc"),
		Date:      extractHeader(msg, "Date"),
		MessageID: extractHeader(msg, "Message-ID"),
		Body:      extractBody(msg),
		LabelIDs:  msg.LabelIds,
	}
}

// buildRawMessage builds an RFC 2822 message for the Gmail API.
// Returns base64url-encoded raw message.
func buildRawMessage(to, subject, body, cc, bcc, threadID, inReplyTo, references string) string {
	var msg strings.Builder

	msg.WriteString(fmt.Sprintf("To: %s\r\n", to))
	if cc != "" {
		msg.WriteString(fmt.Sprintf("Cc: %s\r\n", cc))
	}
	if bcc != "" {
		msg.WriteString(fmt.Sprintf("Bcc: %s\r\n", bcc))
	}
	msg.WriteString(fmt.Sprintf("Subject: %s\r\n", subject))

	if inReplyTo != "" {
		msg.WriteString(fmt.Sprintf("In-Reply-To: %s\r\n", inReplyTo))
	}
	if references != "" {
		msg.WriteString(fmt.Sprintf("References: %s\r\n", references))
	}

	msg.WriteString("MIME-Version: 1.0\r\n")
	msg.WriteString("Content-Type: text/plain; charset=\"UTF-8\"\r\n")
	msg.WriteString("\r\n")
	msg.WriteString(body)

	return base64.URLEncoding.EncodeToString([]byte(msg.String()))
}
