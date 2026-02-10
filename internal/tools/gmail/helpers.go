package gmail

import (
	"context"
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

// AttachmentInfo describes an attachment on a Gmail message.
type AttachmentInfo struct {
	AttachmentID string `json:"attachment_id"`
	Filename     string `json:"filename"`
	MimeType     string `json:"mime_type"`
	Size         int64  `json:"size"`
}

// MessageDetail is the full content of a Gmail message.
type MessageDetail struct {
	ID          string           `json:"id"`
	ThreadID    string           `json:"thread_id"`
	Subject     string           `json:"subject"`
	From        string           `json:"from"`
	To          string           `json:"to"`
	CC          string           `json:"cc,omitempty"`
	Date        string           `json:"date"`
	MessageID   string           `json:"message_id,omitempty"`
	Body        string           `json:"body"`
	LabelIDs    []string         `json:"label_ids,omitempty"`
	Attachments []AttachmentInfo `json:"attachments,omitempty"`
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

// extractAttachments recursively collects attachment metadata from message parts.
func extractAttachments(part *gmail.MessagePart) []AttachmentInfo {
	var result []AttachmentInfo
	if part.Body != nil && part.Body.AttachmentId != "" {
		result = append(result, AttachmentInfo{
			AttachmentID: part.Body.AttachmentId,
			Filename:     part.Filename,
			MimeType:     part.MimeType,
			Size:         part.Body.Size,
		})
	}
	for _, child := range part.Parts {
		result = append(result, extractAttachments(child)...)
	}
	return result
}

// findAttachmentPart recursively locates the MessagePart matching the given attachment ID.
func findAttachmentPart(part *gmail.MessagePart, attachmentID string) *AttachmentInfo {
	if part.Body != nil && part.Body.AttachmentId == attachmentID {
		return &AttachmentInfo{
			AttachmentID: attachmentID,
			Filename:     part.Filename,
			MimeType:     part.MimeType,
			Size:         part.Body.Size,
		}
	}
	for _, child := range part.Parts {
		if info := findAttachmentPart(child, attachmentID); info != nil {
			return info
		}
	}
	return nil
}

// resolveAttachmentMeta fetches the parent message to find the MIME type and
// filename for a given attachment ID. Returns sensible defaults on failure.
func resolveAttachmentMeta(ctx context.Context, srv *gmail.Service, userEmail, messageID, attachmentID string) (mimeType, filename string) {
	msg, err := srv.Users.Messages.Get(userEmail, messageID).
		Format("full").
		Fields("payload").
		Context(ctx).
		Do()
	if err != nil || msg.Payload == nil {
		return "application/octet-stream", "attachment"
	}
	if info := findAttachmentPart(msg.Payload, attachmentID); info != nil {
		return info.MimeType, info.Filename
	}
	return "application/octet-stream", "attachment"
}

// formatAttachmentSize returns a human-readable size string.
func formatAttachmentSize(bytes int64) string {
	if bytes == 0 {
		return "unknown size"
	}
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// isOfficeType returns true if the MIME type is a Microsoft Office XML format.
func isOfficeType(mimeType string) bool {
	return strings.Contains(mimeType, "officedocument") ||
		strings.HasSuffix(mimeType, ".docx") ||
		strings.HasSuffix(mimeType, ".xlsx") ||
		strings.HasSuffix(mimeType, ".pptx")
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
	var attachments []AttachmentInfo
	if msg.Payload != nil {
		attachments = extractAttachments(msg.Payload)
	}
	return MessageDetail{
		ID:          msg.Id,
		ThreadID:    msg.ThreadId,
		Subject:     extractHeader(msg, "Subject"),
		From:        extractHeader(msg, "From"),
		To:          extractHeader(msg, "To"),
		CC:          extractHeader(msg, "Cc"),
		Date:        extractHeader(msg, "Date"),
		MessageID:   extractHeader(msg, "Message-ID"),
		Body:        extractBody(msg),
		LabelIDs:    msg.LabelIds,
		Attachments: attachments,
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
