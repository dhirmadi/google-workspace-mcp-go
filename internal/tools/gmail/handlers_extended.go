package gmail

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"google.golang.org/api/gmail/v1"

	"github.com/evert/google-workspace-mcp-go/internal/middleware"
	"github.com/evert/google-workspace-mcp-go/internal/pkg/response"
	"github.com/evert/google-workspace-mcp-go/internal/services"
)

// --- get_gmail_attachment_content (extended) ---

type GetAttachmentInput struct {
	UserEmail    string `json:"user_google_email" jsonschema:"required" jsonschema_description:"The user's Google email address"`
	MessageID    string `json:"message_id" jsonschema:"required" jsonschema_description:"The message ID containing the attachment"`
	AttachmentID string `json:"attachment_id" jsonschema:"required" jsonschema_description:"The attachment ID to retrieve"`
}

type GetAttachmentOutput struct {
	Data     string `json:"data"`
	Size     int64  `json:"size"`
	MimeType string `json:"mime_type,omitempty"`
}

func createGetAttachmentHandler(factory *services.Factory) mcp.ToolHandlerFor[GetAttachmentInput, GetAttachmentOutput] {
	return func(ctx context.Context, req *mcp.CallToolRequest, input GetAttachmentInput) (*mcp.CallToolResult, GetAttachmentOutput, error) {
		srv, err := factory.Gmail(ctx, input.UserEmail)
		if err != nil {
			return nil, GetAttachmentOutput{}, middleware.HandleGoogleAPIError(err)
		}

		attachment, err := srv.Users.Messages.Attachments.Get(input.UserEmail, input.MessageID, input.AttachmentID).
			Context(ctx).Do()
		if err != nil {
			return nil, GetAttachmentOutput{}, middleware.HandleGoogleAPIError(err)
		}

		rb := response.New()
		rb.Header("Gmail Attachment")
		rb.KeyValue("Message ID", input.MessageID)
		rb.KeyValue("Attachment ID", input.AttachmentID)
		rb.KeyValue("Size", fmt.Sprintf("%d bytes", attachment.Size))
		rb.Line("Data is base64url-encoded.")

		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: rb.Build()}},
		}, GetAttachmentOutput{Data: attachment.Data, Size: attachment.Size}, nil
	}
}

// --- get_gmail_thread_content (extended) ---

type GetThreadInput struct {
	UserEmail string `json:"user_google_email" jsonschema:"required" jsonschema_description:"The user's Google email address"`
	ThreadID  string `json:"thread_id" jsonschema:"required" jsonschema_description:"The Gmail thread ID"`
}

type GetThreadOutput struct {
	ThreadID string          `json:"thread_id"`
	Messages []MessageDetail `json:"messages"`
}

func createGetThreadHandler(factory *services.Factory) mcp.ToolHandlerFor[GetThreadInput, GetThreadOutput] {
	return func(ctx context.Context, req *mcp.CallToolRequest, input GetThreadInput) (*mcp.CallToolResult, GetThreadOutput, error) {
		srv, err := factory.Gmail(ctx, input.UserEmail)
		if err != nil {
			return nil, GetThreadOutput{}, middleware.HandleGoogleAPIError(err)
		}

		thread, err := srv.Users.Threads.Get(input.UserEmail, input.ThreadID).
			Format("full").
			Context(ctx).Do()
		if err != nil {
			return nil, GetThreadOutput{}, middleware.HandleGoogleAPIError(err)
		}

		messages := make([]MessageDetail, 0, len(thread.Messages))
		rb := response.New()
		rb.Header("Gmail Thread")
		rb.KeyValue("Thread ID", thread.Id)
		rb.KeyValue("Messages", len(thread.Messages))
		rb.Blank()

		for _, msg := range thread.Messages {
			detail := messageToDetail(msg)
			messages = append(messages, detail)

			rb.Separator()
			rb.KeyValue("Subject", detail.Subject)
			rb.KeyValue("From", detail.From)
			rb.KeyValue("Date", detail.Date)
			rb.Blank()
			rb.Raw(detail.Body)
			if len(detail.Attachments) > 0 {
				rb.Blank()
				rb.Section("Attachments (%d)", len(detail.Attachments))
				for _, a := range detail.Attachments {
					rb.Item("%s (%s, %s)", a.Filename, a.MimeType, formatAttachmentSize(a.Size))
					rb.Line("    Attachment ID: %s", a.AttachmentID)
				}
			}
			rb.Blank()
		}

		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: rb.Build()}},
		}, GetThreadOutput{ThreadID: thread.Id, Messages: messages}, nil
	}
}

// --- modify_gmail_message_labels (extended) ---

type ModifyLabelsInput struct {
	UserEmail    string   `json:"user_google_email" jsonschema:"required" jsonschema_description:"The user's Google email address"`
	MessageID    string   `json:"message_id" jsonschema:"required" jsonschema_description:"The message ID to modify"`
	AddLabels    []string `json:"add_label_ids,omitempty" jsonschema_description:"Label IDs to add"`
	RemoveLabels []string `json:"remove_label_ids,omitempty" jsonschema_description:"Label IDs to remove"`
}

func createModifyLabelsHandler(factory *services.Factory) mcp.ToolHandlerFor[ModifyLabelsInput, any] {
	return func(ctx context.Context, req *mcp.CallToolRequest, input ModifyLabelsInput) (*mcp.CallToolResult, any, error) {
		if len(input.AddLabels) == 0 && len(input.RemoveLabels) == 0 {
			return nil, nil, fmt.Errorf("specify at least one label to add or remove")
		}

		srv, err := factory.Gmail(ctx, input.UserEmail)
		if err != nil {
			return nil, nil, middleware.HandleGoogleAPIError(err)
		}

		_, err = srv.Users.Messages.Modify(input.UserEmail, input.MessageID, &gmail.ModifyMessageRequest{
			AddLabelIds:    input.AddLabels,
			RemoveLabelIds: input.RemoveLabels,
		}).Context(ctx).Do()
		if err != nil {
			return nil, nil, middleware.HandleGoogleAPIError(err)
		}

		rb := response.New()
		rb.Header("Message Labels Modified")
		rb.KeyValue("Message ID", input.MessageID)
		if len(input.AddLabels) > 0 {
			rb.KeyValue("Added", input.AddLabels)
		}
		if len(input.RemoveLabels) > 0 {
			rb.KeyValue("Removed", input.RemoveLabels)
		}

		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: rb.Build()}},
		}, nil, nil
	}
}

// --- list_gmail_labels (extended) ---

type ListLabelsInput struct {
	UserEmail string `json:"user_google_email" jsonschema:"required" jsonschema_description:"The user's Google email address"`
}

type LabelInfo struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Type string `json:"type"`
}

type ListLabelsOutput struct {
	Labels []LabelInfo `json:"labels"`
}

func createListLabelsHandler(factory *services.Factory) mcp.ToolHandlerFor[ListLabelsInput, ListLabelsOutput] {
	return func(ctx context.Context, req *mcp.CallToolRequest, input ListLabelsInput) (*mcp.CallToolResult, ListLabelsOutput, error) {
		srv, err := factory.Gmail(ctx, input.UserEmail)
		if err != nil {
			return nil, ListLabelsOutput{}, middleware.HandleGoogleAPIError(err)
		}

		result, err := srv.Users.Labels.List(input.UserEmail).Context(ctx).Do()
		if err != nil {
			return nil, ListLabelsOutput{}, middleware.HandleGoogleAPIError(err)
		}

		labels := make([]LabelInfo, 0, len(result.Labels))
		rb := response.New()
		rb.Header("Gmail Labels")
		rb.KeyValue("Count", len(result.Labels))
		rb.Blank()

		for _, l := range result.Labels {
			labels = append(labels, LabelInfo{
				ID:   l.Id,
				Name: l.Name,
				Type: l.Type,
			})
			rb.Item("%s (%s)", l.Name, l.Type)
			rb.Line("    ID: %s", l.Id)
		}

		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: rb.Build()}},
		}, ListLabelsOutput{Labels: labels}, nil
	}
}

// --- manage_gmail_label (extended) ---

type ManageLabelInput struct {
	UserEmail string `json:"user_google_email" jsonschema:"required" jsonschema_description:"The user's Google email address"`
	Action    string `json:"action" jsonschema:"required" jsonschema_description:"Action: create update or delete,enum=create,enum=update,enum=delete"`
	LabelID   string `json:"label_id,omitempty" jsonschema_description:"Label ID (required for update/delete)"`
	Name      string `json:"name,omitempty" jsonschema_description:"Label name (required for create/update)"`
}

func createManageLabelHandler(factory *services.Factory) mcp.ToolHandlerFor[ManageLabelInput, any] {
	return func(ctx context.Context, req *mcp.CallToolRequest, input ManageLabelInput) (*mcp.CallToolResult, any, error) {
		srv, err := factory.Gmail(ctx, input.UserEmail)
		if err != nil {
			return nil, nil, middleware.HandleGoogleAPIError(err)
		}

		rb := response.New()

		switch input.Action {
		case "create":
			if input.Name == "" {
				return nil, nil, fmt.Errorf("name is required for label creation")
			}
			label, err := srv.Users.Labels.Create(input.UserEmail, &gmail.Label{
				Name:                  input.Name,
				LabelListVisibility:   "labelShow",
				MessageListVisibility: "show",
			}).Context(ctx).Do()
			if err != nil {
				return nil, nil, middleware.HandleGoogleAPIError(err)
			}
			rb.Header("Label Created")
			rb.KeyValue("Name", label.Name)
			rb.KeyValue("ID", label.Id)

		case "update":
			if input.LabelID == "" {
				return nil, nil, fmt.Errorf("label_id is required for update")
			}
			if input.Name == "" {
				return nil, nil, fmt.Errorf("name is required for label update")
			}
			label, err := srv.Users.Labels.Update(input.UserEmail, input.LabelID, &gmail.Label{
				Name: input.Name,
			}).Context(ctx).Do()
			if err != nil {
				return nil, nil, middleware.HandleGoogleAPIError(err)
			}
			rb.Header("Label Updated")
			rb.KeyValue("Name", label.Name)
			rb.KeyValue("ID", label.Id)

		case "delete":
			if input.LabelID == "" {
				return nil, nil, fmt.Errorf("label_id is required for deletion")
			}
			err := srv.Users.Labels.Delete(input.UserEmail, input.LabelID).Context(ctx).Do()
			if err != nil {
				return nil, nil, middleware.HandleGoogleAPIError(err)
			}
			rb.Header("Label Deleted")
			rb.KeyValue("Label ID", input.LabelID)

		default:
			return nil, nil, fmt.Errorf("invalid action %q — use create, update, or delete", input.Action)
		}

		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: rb.Build()}},
		}, nil, nil
	}
}

// --- draft_gmail_message (extended) ---

type DraftMessageInput struct {
	UserEmail string `json:"user_google_email" jsonschema:"required" jsonschema_description:"The user's Google email address"`
	To        string `json:"to" jsonschema:"required" jsonschema_description:"Recipient email address"`
	Subject   string `json:"subject" jsonschema:"required" jsonschema_description:"Email subject"`
	Body      string `json:"body" jsonschema:"required" jsonschema_description:"Email body content"`
	CC        string `json:"cc,omitempty" jsonschema_description:"CC email address"`
	BCC       string `json:"bcc,omitempty" jsonschema_description:"BCC email address"`
	ThreadID  string `json:"thread_id,omitempty" jsonschema_description:"Thread ID to reply in"`
}

func createDraftMessageHandler(factory *services.Factory) mcp.ToolHandlerFor[DraftMessageInput, any] {
	return func(ctx context.Context, req *mcp.CallToolRequest, input DraftMessageInput) (*mcp.CallToolResult, any, error) {
		srv, err := factory.Gmail(ctx, input.UserEmail)
		if err != nil {
			return nil, nil, middleware.HandleGoogleAPIError(err)
		}

		rawMsg := buildRawMessage(input.To, input.Subject, input.Body, input.CC, input.BCC, input.ThreadID, "", "")

		msg := &gmail.Message{Raw: rawMsg}
		if input.ThreadID != "" {
			msg.ThreadId = input.ThreadID
		}

		draft, err := srv.Users.Drafts.Create(input.UserEmail, &gmail.Draft{
			Message: msg,
		}).Context(ctx).Do()
		if err != nil {
			return nil, nil, middleware.HandleGoogleAPIError(err)
		}

		rb := response.New()
		rb.Header("Draft Created")
		rb.KeyValue("Draft ID", draft.Id)
		rb.KeyValue("To", input.To)
		rb.KeyValue("Subject", input.Subject)
		if draft.Message != nil {
			rb.KeyValue("Message ID", draft.Message.Id)
		}

		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: rb.Build()}},
		}, nil, nil
	}
}

// --- list_gmail_filters (extended) ---

type ListFiltersInput struct {
	UserEmail string `json:"user_google_email" jsonschema:"required" jsonschema_description:"The user's Google email address"`
}

type FilterInfo struct {
	ID       string            `json:"id"`
	Criteria map[string]string `json:"criteria,omitempty"`
	Actions  map[string]string `json:"actions,omitempty"`
}

type ListFiltersOutput struct {
	Filters []FilterInfo `json:"filters"`
}

func createListFiltersHandler(factory *services.Factory) mcp.ToolHandlerFor[ListFiltersInput, ListFiltersOutput] {
	return func(ctx context.Context, req *mcp.CallToolRequest, input ListFiltersInput) (*mcp.CallToolResult, ListFiltersOutput, error) {
		srv, err := factory.Gmail(ctx, input.UserEmail)
		if err != nil {
			return nil, ListFiltersOutput{}, middleware.HandleGoogleAPIError(err)
		}

		result, err := srv.Users.Settings.Filters.List(input.UserEmail).Context(ctx).Do()
		if err != nil {
			return nil, ListFiltersOutput{}, middleware.HandleGoogleAPIError(err)
		}

		filters := make([]FilterInfo, 0, len(result.Filter))
		rb := response.New()
		rb.Header("Gmail Filters")
		rb.KeyValue("Count", len(result.Filter))
		rb.Blank()

		for _, f := range result.Filter {
			fi := FilterInfo{ID: f.Id}

			if f.Criteria != nil {
				fi.Criteria = make(map[string]string)
				if f.Criteria.From != "" {
					fi.Criteria["from"] = f.Criteria.From
				}
				if f.Criteria.To != "" {
					fi.Criteria["to"] = f.Criteria.To
				}
				if f.Criteria.Subject != "" {
					fi.Criteria["subject"] = f.Criteria.Subject
				}
				if f.Criteria.Query != "" {
					fi.Criteria["query"] = f.Criteria.Query
				}
			}

			if f.Action != nil {
				fi.Actions = make(map[string]string)
				if len(f.Action.AddLabelIds) > 0 {
					fi.Actions["addLabels"] = fmt.Sprintf("%v", f.Action.AddLabelIds)
				}
				if len(f.Action.RemoveLabelIds) > 0 {
					fi.Actions["removeLabels"] = fmt.Sprintf("%v", f.Action.RemoveLabelIds)
				}
				if f.Action.Forward != "" {
					fi.Actions["forward"] = f.Action.Forward
				}
			}

			filters = append(filters, fi)
			criteriaJSON, _ := json.Marshal(fi.Criteria)
			rb.Item("Filter %s", f.Id)
			rb.Line("    Criteria: %s", string(criteriaJSON))
		}

		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: rb.Build()}},
		}, ListFiltersOutput{Filters: filters}, nil
	}
}

// --- create_gmail_filter (extended) ---

type CreateFilterInput struct {
	UserEmail     string   `json:"user_google_email" jsonschema:"required" jsonschema_description:"The user's Google email address"`
	From          string   `json:"from,omitempty" jsonschema_description:"Match messages from this sender"`
	To            string   `json:"to,omitempty" jsonschema_description:"Match messages to this recipient"`
	Subject       string   `json:"subject,omitempty" jsonschema_description:"Match messages with this subject"`
	Query         string   `json:"query,omitempty" jsonschema_description:"Match messages with this query"`
	AddLabelIDs   []string `json:"add_label_ids,omitempty" jsonschema_description:"Label IDs to add to matching messages"`
	RemoveLabelIDs []string `json:"remove_label_ids,omitempty" jsonschema_description:"Label IDs to remove from matching messages"`
	Forward       string   `json:"forward,omitempty" jsonschema_description:"Email address to forward matching messages to"`
}

func createCreateFilterHandler(factory *services.Factory) mcp.ToolHandlerFor[CreateFilterInput, any] {
	return func(ctx context.Context, req *mcp.CallToolRequest, input CreateFilterInput) (*mcp.CallToolResult, any, error) {
		srv, err := factory.Gmail(ctx, input.UserEmail)
		if err != nil {
			return nil, nil, middleware.HandleGoogleAPIError(err)
		}

		filter := &gmail.Filter{
			Criteria: &gmail.FilterCriteria{
				From:    input.From,
				To:      input.To,
				Subject: input.Subject,
				Query:   input.Query,
			},
			Action: &gmail.FilterAction{
				AddLabelIds:    input.AddLabelIDs,
				RemoveLabelIds: input.RemoveLabelIDs,
				Forward:        input.Forward,
			},
		}

		created, err := srv.Users.Settings.Filters.Create(input.UserEmail, filter).Context(ctx).Do()
		if err != nil {
			return nil, nil, middleware.HandleGoogleAPIError(err)
		}

		rb := response.New()
		rb.Header("Filter Created")
		rb.KeyValue("Filter ID", created.Id)
		if input.From != "" {
			rb.KeyValue("From", input.From)
		}
		if input.Query != "" {
			rb.KeyValue("Query", input.Query)
		}

		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: rb.Build()}},
		}, nil, nil
	}
}

// --- delete_gmail_filter (extended) ---

type DeleteFilterInput struct {
	UserEmail string `json:"user_google_email" jsonschema:"required" jsonschema_description:"The user's Google email address"`
	FilterID  string `json:"filter_id" jsonschema:"required" jsonschema_description:"The filter ID to delete"`
}

func createDeleteFilterHandler(factory *services.Factory) mcp.ToolHandlerFor[DeleteFilterInput, any] {
	return func(ctx context.Context, req *mcp.CallToolRequest, input DeleteFilterInput) (*mcp.CallToolResult, any, error) {
		srv, err := factory.Gmail(ctx, input.UserEmail)
		if err != nil {
			return nil, nil, middleware.HandleGoogleAPIError(err)
		}

		err = srv.Users.Settings.Filters.Delete(input.UserEmail, input.FilterID).Context(ctx).Do()
		if err != nil {
			return nil, nil, middleware.HandleGoogleAPIError(err)
		}

		rb := response.New()
		rb.Header("Filter Deleted")
		rb.KeyValue("Filter ID", input.FilterID)

		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: rb.Build()}},
		}, nil, nil
	}
}

