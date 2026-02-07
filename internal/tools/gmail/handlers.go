package gmail

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"google.golang.org/api/gmail/v1"

	"github.com/evert/google-workspace-mcp-go/internal/middleware"
	"github.com/evert/google-workspace-mcp-go/internal/pkg/response"
	"github.com/evert/google-workspace-mcp-go/internal/services"
)

// --- search_gmail_messages ---

// SearchMessagesInput is the input for search_gmail_messages.
type SearchMessagesInput struct {
	UserEmail string `json:"user_google_email" jsonschema:"required" jsonschema_description:"The user's Google email address"`
	Query     string `json:"query" jsonschema:"required" jsonschema_description:"Gmail search query using standard Gmail search operators"`
	PageSize  int    `json:"page_size,omitempty" jsonschema_description:"Maximum number of results to return (default 10)"`
	PageToken string `json:"page_token,omitempty" jsonschema_description:"Token for retrieving the next page of results"`
}

// SearchMessagesOutput is the structured output for search_gmail_messages.
type SearchMessagesOutput struct {
	Messages      []MessageSummary `json:"messages"`
	Query         string           `json:"query"`
	NextPageToken string           `json:"next_page_token,omitempty"`
	ResultCount   int              `json:"result_count"`
}

func createSearchMessagesHandler(factory *services.Factory) mcp.ToolHandlerFor[SearchMessagesInput, SearchMessagesOutput] {
	return func(ctx context.Context, req *mcp.CallToolRequest, input SearchMessagesInput) (*mcp.CallToolResult, SearchMessagesOutput, error) {
		if input.PageSize == 0 {
			input.PageSize = 10
		}

		srv, err := factory.Gmail(ctx, input.UserEmail)
		if err != nil {
			return nil, SearchMessagesOutput{}, middleware.HandleGoogleAPIError(err)
		}

		result, err := srv.Users.Messages.List(input.UserEmail).
			Q(input.Query).
			MaxResults(int64(input.PageSize)).
			PageToken(input.PageToken).
			Context(ctx).
			Do()
		if err != nil {
			return nil, SearchMessagesOutput{}, middleware.HandleGoogleAPIError(err)
		}

		// Fetch minimal metadata for each message
		summaries := make([]MessageSummary, 0, len(result.Messages))
		for _, m := range result.Messages {
			msg, err := srv.Users.Messages.Get(input.UserEmail, m.Id).
				Format("metadata").
				MetadataHeaders("Subject", "From", "To", "Date").
				Context(ctx).
				Do()
			if err != nil {
				continue
			}
			summaries = append(summaries, messageToSummary(msg))
		}

		// Build text output
		rb := response.New()
		rb.Header("Gmail Search Results")
		rb.KeyValue("Query", input.Query)
		rb.KeyValue("Results", len(summaries))
		if result.NextPageToken != "" {
			rb.KeyValue("Next page token", result.NextPageToken)
		}
		rb.Blank()
		for _, s := range summaries {
			rb.Item("Subject: %s", s.Subject)
			rb.Line("    From: %s | Date: %s", s.From, s.Date)
			rb.Line("    ID: %s (Thread: %s)", s.ID, s.ThreadID)
		}

		output := SearchMessagesOutput{
			Messages:      summaries,
			Query:         input.Query,
			NextPageToken: result.NextPageToken,
			ResultCount:   len(summaries),
		}

		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: rb.Build()}},
		}, output, nil
	}
}

// --- get_gmail_message_content ---

// GetMessageContentInput is the input for get_gmail_message_content.
type GetMessageContentInput struct {
	UserEmail string `json:"user_google_email" jsonschema:"required" jsonschema_description:"The user's Google email address"`
	MessageID string `json:"message_id" jsonschema:"required" jsonschema_description:"The unique ID of the Gmail message to retrieve"`
}

// GetMessageContentOutput is the structured output for get_gmail_message_content.
type GetMessageContentOutput struct {
	Message MessageDetail `json:"message"`
}

func createGetMessageContentHandler(factory *services.Factory) mcp.ToolHandlerFor[GetMessageContentInput, GetMessageContentOutput] {
	return func(ctx context.Context, req *mcp.CallToolRequest, input GetMessageContentInput) (*mcp.CallToolResult, GetMessageContentOutput, error) {
		srv, err := factory.Gmail(ctx, input.UserEmail)
		if err != nil {
			return nil, GetMessageContentOutput{}, middleware.HandleGoogleAPIError(err)
		}

		msg, err := srv.Users.Messages.Get(input.UserEmail, input.MessageID).
			Format("full").
			Context(ctx).
			Do()
		if err != nil {
			return nil, GetMessageContentOutput{}, middleware.HandleGoogleAPIError(err)
		}

		detail := messageToDetail(msg)

		rb := response.New()
		rb.Header("Gmail Message")
		rb.KeyValue("Subject", detail.Subject)
		rb.KeyValue("From", detail.From)
		rb.KeyValue("To", detail.To)
		if detail.CC != "" {
			rb.KeyValue("CC", detail.CC)
		}
		rb.KeyValue("Date", detail.Date)
		rb.KeyValue("Message ID", detail.ID)
		if detail.MessageID != "" {
			rb.KeyValue("Message-ID Header", detail.MessageID)
		}
		if len(detail.Attachments) > 0 {
			rb.Blank()
			rb.Section("Attachments")
			for _, a := range detail.Attachments {
				rb.Item("%s (%s, %d bytes)", a.Filename, a.MimeType, a.Size)
				rb.Line("    Attachment ID: %s", a.AttachmentID)
			}
		}
		rb.Blank()
		rb.Section("Body")
		rb.Raw(detail.Body)

		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: rb.Build()}},
		}, GetMessageContentOutput{Message: detail}, nil
	}
}

// --- get_gmail_messages_content_batch ---

// BatchGetMessagesInput is the input for get_gmail_messages_content_batch.
type BatchGetMessagesInput struct {
	UserEmail  string   `json:"user_google_email" jsonschema:"required" jsonschema_description:"The user's Google email address"`
	MessageIDs []string `json:"message_ids" jsonschema:"required" jsonschema_description:"List of Gmail message IDs to retrieve (max 25)"`
	Format     string   `json:"format,omitempty" jsonschema_description:"Message format: full or metadata (default full),enum=full,enum=metadata"`
}

// BatchGetMessagesOutput is the structured output for get_gmail_messages_content_batch.
type BatchGetMessagesOutput struct {
	Messages []MessageDetail `json:"messages"`
}

func createBatchGetMessagesHandler(factory *services.Factory) mcp.ToolHandlerFor[BatchGetMessagesInput, BatchGetMessagesOutput] {
	return func(ctx context.Context, req *mcp.CallToolRequest, input BatchGetMessagesInput) (*mcp.CallToolResult, BatchGetMessagesOutput, error) {
		if len(input.MessageIDs) > 25 {
			return nil, BatchGetMessagesOutput{}, fmt.Errorf("maximum 25 messages per batch request, got %d — split into multiple calls", len(input.MessageIDs))
		}

		if input.Format == "" {
			input.Format = "full"
		}

		srv, err := factory.Gmail(ctx, input.UserEmail)
		if err != nil {
			return nil, BatchGetMessagesOutput{}, middleware.HandleGoogleAPIError(err)
		}

		total := len(input.MessageIDs)
		messages := make([]MessageDetail, 0, total)

		for i, id := range input.MessageIDs {
			// Report progress if client supports it
			if pt := req.Params.GetProgressToken(); pt != nil {
				_ = req.Session.NotifyProgress(ctx, &mcp.ProgressNotificationParams{
					ProgressToken: pt,
					Progress:      float64(i),
					Total:         float64(total),
					Message:       fmt.Sprintf("Fetching message %d/%d", i+1, total),
				})
			}

			msg, err := srv.Users.Messages.Get(input.UserEmail, id).
				Format(input.Format).
				Context(ctx).
				Do()
			if err != nil {
				// Log but continue — don't fail the whole batch for one bad ID
				continue
			}
			messages = append(messages, messageToDetail(msg))
		}

		rb := response.New()
		rb.Header("Gmail Batch Messages")
		rb.KeyValue("Requested", total)
		rb.KeyValue("Retrieved", len(messages))
		rb.Blank()
		for _, m := range messages {
			rb.Separator()
			rb.KeyValue("Subject", m.Subject)
			rb.KeyValue("From", m.From)
			rb.KeyValue("Date", m.Date)
			rb.KeyValue("ID", m.ID)
			if len(m.Attachments) > 0 {
				rb.KeyValue("Attachments", len(m.Attachments))
				for _, a := range m.Attachments {
					rb.Line("    • %s (%s, %d bytes) — ID: %s", a.Filename, a.MimeType, a.Size, a.AttachmentID)
				}
			}
			rb.Blank()
			rb.Raw(m.Body)
			rb.Blank()
		}

		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: rb.Build()}},
		}, BatchGetMessagesOutput{Messages: messages}, nil
	}
}

// --- send_gmail_message ---

// SendMessageInput is the input for send_gmail_message.
type SendMessageInput struct {
	UserEmail  string `json:"user_google_email" jsonschema:"required" jsonschema_description:"The user's Google email address"`
	To         string `json:"to" jsonschema:"required" jsonschema_description:"Recipient email address"`
	Subject    string `json:"subject" jsonschema:"required" jsonschema_description:"Email subject"`
	Body       string `json:"body" jsonschema:"required" jsonschema_description:"Email body content (plain text)"`
	CC         string `json:"cc,omitempty" jsonschema_description:"CC email address"`
	BCC        string `json:"bcc,omitempty" jsonschema_description:"BCC email address"`
	ThreadID   string `json:"thread_id,omitempty" jsonschema_description:"Gmail thread ID to reply within"`
	InReplyTo  string `json:"in_reply_to,omitempty" jsonschema_description:"Message-ID of the message being replied to"`
	References string `json:"references,omitempty" jsonschema_description:"Chain of Message-IDs for proper threading"`
}

func createSendMessageHandler(factory *services.Factory) mcp.ToolHandlerFor[SendMessageInput, any] {
	return func(ctx context.Context, req *mcp.CallToolRequest, input SendMessageInput) (*mcp.CallToolResult, any, error) {
		srv, err := factory.Gmail(ctx, input.UserEmail)
		if err != nil {
			return nil, nil, middleware.HandleGoogleAPIError(err)
		}

		rawMsg := buildRawMessage(input.To, input.Subject, input.Body, input.CC, input.BCC, input.ThreadID, input.InReplyTo, input.References)

		gmailMsg := &gmail.Message{
			Raw: rawMsg,
		}
		if input.ThreadID != "" {
			gmailMsg.ThreadId = input.ThreadID
		}

		sent, err := srv.Users.Messages.Send(input.UserEmail, gmailMsg).
			Context(ctx).
			Do()
		if err != nil {
			return nil, nil, middleware.HandleGoogleAPIError(err)
		}

		rb := response.New()
		rb.Header("Message Sent")
		rb.KeyValue("To", input.To)
		rb.KeyValue("Subject", input.Subject)
		rb.KeyValue("Message ID", sent.Id)
		rb.KeyValue("Thread ID", sent.ThreadId)
		if input.CC != "" {
			rb.KeyValue("CC", input.CC)
		}

		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: rb.Build()}},
		}, nil, nil
	}
}
