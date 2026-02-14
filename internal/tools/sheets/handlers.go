package sheets

import (
	"context"
	"fmt"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"google.golang.org/api/sheets/v4"

	"github.com/evert/google-workspace-mcp-go/internal/middleware"
	"github.com/evert/google-workspace-mcp-go/internal/pkg/response"
	"github.com/evert/google-workspace-mcp-go/internal/services"
)

// --- create_spreadsheet ---

type CreateSpreadsheetInput struct {
	UserEmail  string   `json:"user_google_email" jsonschema:"required" jsonschema_description:"The user's Google email address"`
	Title      string   `json:"title" jsonschema:"required" jsonschema_description:"Title for the new spreadsheet"`
	SheetNames []string `json:"sheet_names,omitempty" jsonschema_description:"Sheet tab names to create (default: one sheet with default name)"`
}

func createCreateSpreadsheetHandler(factory *services.Factory) mcp.ToolHandlerFor[CreateSpreadsheetInput, any] {
	return func(ctx context.Context, req *mcp.CallToolRequest, input CreateSpreadsheetInput) (*mcp.CallToolResult, any, error) {
		srv, err := factory.Sheets(ctx, input.UserEmail)
		if err != nil {
			return nil, nil, middleware.HandleGoogleAPIError(err)
		}

		spreadsheet := &sheets.Spreadsheet{
			Properties: &sheets.SpreadsheetProperties{
				Title: input.Title,
			},
		}

		if len(input.SheetNames) > 0 {
			sheetsList := make([]*sheets.Sheet, 0, len(input.SheetNames))
			for _, name := range input.SheetNames {
				sheetsList = append(sheetsList, &sheets.Sheet{
					Properties: &sheets.SheetProperties{
						Title: name,
					},
				})
			}
			spreadsheet.Sheets = sheetsList
		}

		created, err := srv.Spreadsheets.Create(spreadsheet).Context(ctx).Do()
		if err != nil {
			return nil, nil, middleware.HandleGoogleAPIError(err)
		}

		rb := response.New()
		rb.Header("Spreadsheet Created")
		rb.KeyValue("Title", created.Properties.Title)
		rb.KeyValue("ID", created.SpreadsheetId)
		rb.KeyValue("URL", created.SpreadsheetUrl)
		rb.KeyValue("Locale", created.Properties.Locale)
		if len(created.Sheets) > 0 {
			rb.Blank()
			rb.Section("Sheets")
			for _, s := range created.Sheets {
				rb.Item("%s", s.Properties.Title)
			}
		}

		return rb.TextResult(), nil, nil
	}
}

// --- read_sheet_values ---

type ReadSheetValuesInput struct {
	UserEmail     string `json:"user_google_email" jsonschema:"required" jsonschema_description:"The user's Google email address"`
	SpreadsheetID string `json:"spreadsheet_id" jsonschema:"required" jsonschema_description:"The ID of the spreadsheet"`
	RangeName     string `json:"range_name,omitempty" jsonschema_description:"Range to read (e.g. Sheet1!A1:D10). Default: A1:Z1000"`
}

type ReadSheetValuesOutput struct {
	Values [][]interface{} `json:"values"`
	Range  string          `json:"range"`
}

func createReadSheetValuesHandler(factory *services.Factory) mcp.ToolHandlerFor[ReadSheetValuesInput, ReadSheetValuesOutput] {
	return func(ctx context.Context, req *mcp.CallToolRequest, input ReadSheetValuesInput) (*mcp.CallToolResult, ReadSheetValuesOutput, error) {
		srv, err := factory.Sheets(ctx, input.UserEmail)
		if err != nil {
			return nil, ReadSheetValuesOutput{}, middleware.HandleGoogleAPIError(err)
		}

		rangeName := input.RangeName
		if rangeName == "" {
			rangeName = "A1:Z1000"
		}

		result, err := srv.Spreadsheets.Values.Get(input.SpreadsheetID, rangeName).Context(ctx).Do()
		if err != nil {
			return nil, ReadSheetValuesOutput{}, middleware.HandleGoogleAPIError(err)
		}

		rb := response.New()
		rb.Header("Sheet Values")
		rb.KeyValue("Spreadsheet", input.SpreadsheetID)
		rb.KeyValue("Range", result.Range)
		rb.KeyValue("Rows", len(result.Values))
		rb.Blank()

		for i, row := range result.Values {
			cells := make([]string, 0, len(row))
			for _, cell := range row {
				cells = append(cells, fmt.Sprintf("%v", cell))
			}
			rb.Line("Row %d: %s", i+1, strings.Join(cells, " | "))
		}

		return rb.TextResult(), ReadSheetValuesOutput{Values: result.Values, Range: result.Range}, nil
	}
}

// --- modify_sheet_values ---

type ModifySheetValuesInput struct {
	UserEmail        string     `json:"user_google_email" jsonschema:"required" jsonschema_description:"The user's Google email address"`
	SpreadsheetID    string     `json:"spreadsheet_id" jsonschema:"required" jsonschema_description:"The ID of the spreadsheet"`
	RangeName        string     `json:"range_name" jsonschema:"required" jsonschema_description:"Range to modify (e.g. Sheet1!A1:D10)"`
	Values           [][]string `json:"values,omitempty" jsonschema_description:"2D array of values to write. Required unless clear_values is true."`
	ValueInputOption string     `json:"value_input_option,omitempty" jsonschema_description:"How to interpret input: RAW or USER_ENTERED (default USER_ENTERED)"`
	ClearValues      bool       `json:"clear_values,omitempty" jsonschema_description:"If true clears the range instead of writing values"`
}

func createModifySheetValuesHandler(factory *services.Factory) mcp.ToolHandlerFor[ModifySheetValuesInput, any] {
	return func(ctx context.Context, req *mcp.CallToolRequest, input ModifySheetValuesInput) (*mcp.CallToolResult, any, error) {
		srv, err := factory.Sheets(ctx, input.UserEmail)
		if err != nil {
			return nil, nil, middleware.HandleGoogleAPIError(err)
		}

		rb := response.New()

		if input.ClearValues {
			_, err := srv.Spreadsheets.Values.Clear(input.SpreadsheetID, input.RangeName, &sheets.ClearValuesRequest{}).
				Context(ctx).Do()
			if err != nil {
				return nil, nil, middleware.HandleGoogleAPIError(err)
			}

			rb.Header("Values Cleared")
			rb.KeyValue("Spreadsheet", input.SpreadsheetID)
			rb.KeyValue("Range", input.RangeName)
		} else {
			if len(input.Values) == 0 {
				return nil, nil, fmt.Errorf("values are required when clear_values is false — provide a 2D array of values or set clear_values to true")
			}

			valueInputOption := input.ValueInputOption
			if valueInputOption == "" {
				valueInputOption = "USER_ENTERED"
			}

			// Convert [][]string to [][]interface{}
			iface := make([][]interface{}, 0, len(input.Values))
			for _, row := range input.Values {
				ifaceRow := make([]interface{}, 0, len(row))
				for _, cell := range row {
					ifaceRow = append(ifaceRow, cell)
				}
				iface = append(iface, ifaceRow)
			}

			vr := &sheets.ValueRange{
				Values: iface,
			}

			result, err := srv.Spreadsheets.Values.Update(input.SpreadsheetID, input.RangeName, vr).
				ValueInputOption(valueInputOption).
				Context(ctx).Do()
			if err != nil {
				return nil, nil, middleware.HandleGoogleAPIError(err)
			}

			rb.Header("Values Updated")
			rb.KeyValue("Spreadsheet", input.SpreadsheetID)
			rb.KeyValue("Range", result.UpdatedRange)
			rb.KeyValue("Updated rows", result.UpdatedRows)
			rb.KeyValue("Updated columns", result.UpdatedColumns)
			rb.KeyValue("Updated cells", result.UpdatedCells)
		}

		return rb.TextResult(), nil, nil
	}
}
