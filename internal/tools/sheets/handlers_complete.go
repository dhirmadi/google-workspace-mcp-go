package sheets

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	sheetspb "google.golang.org/api/sheets/v4"

	"github.com/evert/google-workspace-mcp-go/internal/middleware"
	"github.com/evert/google-workspace-mcp-go/internal/pkg/response"
	"github.com/evert/google-workspace-mcp-go/internal/services"
)

// --- create_sheet (complete) ---

type CreateSheetInput struct {
	UserEmail     string `json:"user_google_email" jsonschema:"required" jsonschema_description:"The user's Google email address"`
	SpreadsheetID string `json:"spreadsheet_id" jsonschema:"required" jsonschema_description:"The Google Sheets spreadsheet ID"`
	Title         string `json:"title" jsonschema:"required" jsonschema_description:"Title for the new sheet tab"`
	Index         int    `json:"index,omitempty" jsonschema_description:"Position of the new sheet (0-based). If omitted the sheet is added at the end."`
}

func createCreateSheetHandler(factory *services.Factory) mcp.ToolHandlerFor[CreateSheetInput, any] {
	return func(ctx context.Context, req *mcp.CallToolRequest, input CreateSheetInput) (*mcp.CallToolResult, any, error) {
		srv, err := factory.Sheets(ctx, input.UserEmail)
		if err != nil {
			return nil, nil, middleware.HandleGoogleAPIError(err)
		}

		addReq := &sheetspb.AddSheetRequest{
			Properties: &sheetspb.SheetProperties{
				Title: input.Title,
			},
		}
		if input.Index > 0 {
			addReq.Properties.Index = int64(input.Index)
		}

		batchReq := &sheetspb.BatchUpdateSpreadsheetRequest{
			Requests: []*sheetspb.Request{
				{AddSheet: addReq},
			},
		}

		result, err := srv.Spreadsheets.BatchUpdate(input.SpreadsheetID, batchReq).Context(ctx).Do()
		if err != nil {
			return nil, nil, middleware.HandleGoogleAPIError(err)
		}

		rb := response.New()
		rb.Header("Sheet Created")
		rb.KeyValue("Spreadsheet ID", input.SpreadsheetID)
		rb.KeyValue("Title", input.Title)
		if len(result.Replies) > 0 && result.Replies[0].AddSheet != nil {
			props := result.Replies[0].AddSheet.Properties
			rb.KeyValue("Sheet ID", fmt.Sprintf("%d", props.SheetId))
			rb.KeyValue("Index", fmt.Sprintf("%d", props.Index))
		}

		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: rb.Build()}},
		}, nil, nil
	}
}
