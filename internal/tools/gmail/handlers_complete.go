package gmail

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	gmailpb "google.golang.org/api/gmail/v1"

	"github.com/evert/google-workspace-mcp-go/internal/middleware"
	"github.com/evert/google-workspace-mcp-go/internal/pkg/response"
	"github.com/evert/google-workspace-mcp-go/internal/services"
)

// --- get_gmail_threads_content_batch (complete) ---

type BatchGetThreadsInput struct {
	UserEmail string   `json:"user_google_email" jsonschema:"required" jsonschema_description:"The user's Google email address"`
	ThreadIDs []string `json:"thread_ids" jsonschema:"required" jsonschema_description:"Thread IDs to retrieve (max 25)"`
	Format    string   `json:"format,omitempty" jsonschema_description:"Response format: full or metadata (default full),enum=full,enum=metadata"`
}

type BatchGetThreadsOutput struct {
	Threads []ThreadSummary `json:"threads"`
}

type ThreadSummary struct {
	ThreadID     string           `json:"thread_id"`
	MessageCount int              `json:"message_count"`
	Messages     []MessageSummary `json:"messages"`
}

func createBatchGetThreadsHandler(factory *services.Factory) mcp.ToolHandlerFor[BatchGetThreadsInput, BatchGetThreadsOutput] {
	return func(ctx context.Context, req *mcp.CallToolRequest, input BatchGetThreadsInput) (*mcp.CallToolResult, BatchGetThreadsOutput, error) {
		if len(input.ThreadIDs) > 25 {
			return nil, BatchGetThreadsOutput{}, fmt.Errorf("maximum 25 threads per batch request, got %d - split into multiple calls", len(input.ThreadIDs))
		}

		srv, err := factory.Gmail(ctx, input.UserEmail)
		if err != nil {
			return nil, BatchGetThreadsOutput{}, middleware.HandleGoogleAPIError(err)
		}

		format := "full"
		if input.Format == "metadata" {
			format = "metadata"
		}

		threads := make([]ThreadSummary, 0, len(input.ThreadIDs))
		rb := response.New()
		rb.Header("Thread Contents (Batch)")
		rb.KeyValue("Threads", len(input.ThreadIDs))
		rb.Blank()

		total := len(input.ThreadIDs)
		for i, threadID := range input.ThreadIDs {
			// Report progress
			if pt := req.Params.GetProgressToken(); pt != nil {
				_ = req.Session.NotifyProgress(ctx, &mcp.ProgressNotificationParams{
					ProgressToken: pt,
					Progress:      float64(i),
					Total:         float64(total),
					Message:       fmt.Sprintf("Fetching thread %d/%d", i+1, total),
				})
			}

			thread, err := srv.Users.Threads.Get(input.UserEmail, threadID).
				Format(format).
				Context(ctx).Do()
			if err != nil {
				rb.Item("Thread %s: ERROR — %v", threadID, err)
				continue
			}

			ts := ThreadSummary{
				ThreadID:     thread.Id,
				MessageCount: len(thread.Messages),
			}

			rb.Item("Thread: %s (%d messages)", thread.Id, len(thread.Messages))

			for _, msg := range thread.Messages {
				ms := messageToSummary(msg)
				ts.Messages = append(ts.Messages, ms)
				rb.Line("    [%s] %s — %s", ms.Date, ms.From, ms.Subject)
			}

			threads = append(threads, ts)
		}

		return rb.TextResult(), BatchGetThreadsOutput{Threads: threads}, nil
	}
}

// --- batch_modify_gmail_message_labels (complete) ---

type BatchModifyLabelsInput struct {
	UserEmail      string   `json:"user_google_email" jsonschema:"required" jsonschema_description:"The user's Google email address"`
	MessageIDs     []string `json:"message_ids" jsonschema:"required" jsonschema_description:"Message IDs to modify"`
	AddLabelIDs    []string `json:"add_label_ids,omitempty" jsonschema_description:"Label IDs to add to messages"`
	RemoveLabelIDs []string `json:"remove_label_ids,omitempty" jsonschema_description:"Label IDs to remove from messages"`
}

func createBatchModifyLabelsHandler(factory *services.Factory) mcp.ToolHandlerFor[BatchModifyLabelsInput, any] {
	return func(ctx context.Context, req *mcp.CallToolRequest, input BatchModifyLabelsInput) (*mcp.CallToolResult, any, error) {
		if len(input.AddLabelIDs) == 0 && len(input.RemoveLabelIDs) == 0 {
			return nil, nil, fmt.Errorf("at least one of add_label_ids or remove_label_ids must be specified")
		}

		srv, err := factory.Gmail(ctx, input.UserEmail)
		if err != nil {
			return nil, nil, middleware.HandleGoogleAPIError(err)
		}

		modReq := &gmailpb.BatchModifyMessagesRequest{
			Ids:            input.MessageIDs,
			AddLabelIds:    input.AddLabelIDs,
			RemoveLabelIds: input.RemoveLabelIDs,
		}

		err = srv.Users.Messages.BatchModify(input.UserEmail, modReq).Context(ctx).Do()
		if err != nil {
			return nil, nil, middleware.HandleGoogleAPIError(err)
		}

		rb := response.New()
		rb.Header("Batch Label Modification Complete")
		rb.KeyValue("Messages Modified", len(input.MessageIDs))
		if len(input.AddLabelIDs) > 0 {
			rb.KeyValue("Labels Added", len(input.AddLabelIDs))
		}
		if len(input.RemoveLabelIDs) > 0 {
			rb.KeyValue("Labels Removed", len(input.RemoveLabelIDs))
		}

		return rb.TextResult(), nil, nil
	}
}
