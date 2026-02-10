package sheets

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"google.golang.org/api/sheets/v4"

	"github.com/evert/google-workspace-mcp-go/internal/middleware"
	"github.com/evert/google-workspace-mcp-go/internal/pkg/response"
	"github.com/evert/google-workspace-mcp-go/internal/services"
)

// --- list_spreadsheets (extended) ---

type ListSpreadsheetsInput struct {
	UserEmail string `json:"user_google_email" jsonschema:"required" jsonschema_description:"The user's Google email address"`
	PageSize  int    `json:"page_size,omitempty" jsonschema_description:"Maximum results (default 10)"`
	Query     string `json:"query,omitempty" jsonschema_description:"Additional Drive query filter"`
}

type SpreadsheetSummary struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	ModifiedTime string `json:"modified_time,omitempty"`
	WebViewLink  string `json:"web_view_link,omitempty"`
}

type ListSpreadsheetsOutput struct {
	Spreadsheets []SpreadsheetSummary `json:"spreadsheets"`
}

func createListSpreadsheetsHandler(factory *services.Factory) mcp.ToolHandlerFor[ListSpreadsheetsInput, ListSpreadsheetsOutput] {
	return func(ctx context.Context, req *mcp.CallToolRequest, input ListSpreadsheetsInput) (*mcp.CallToolResult, ListSpreadsheetsOutput, error) {
		if input.PageSize == 0 {
			input.PageSize = 10
		}

		// Use Drive API to search for spreadsheets
		drvSrv, err := factory.Drive(ctx, input.UserEmail)
		if err != nil {
			return nil, ListSpreadsheetsOutput{}, middleware.HandleGoogleAPIError(err)
		}

		q := "mimeType='application/vnd.google-apps.spreadsheet' and trashed=false"
		if input.Query != "" {
			q += " and " + input.Query
		}

		result, err := drvSrv.Files.List().
			Q(q).
			PageSize(int64(input.PageSize)).
			Fields("files(id, name, modifiedTime, webViewLink)").
			SupportsAllDrives(true).
			IncludeItemsFromAllDrives(true).
			Context(ctx).Do()
		if err != nil {
			return nil, ListSpreadsheetsOutput{}, middleware.HandleGoogleAPIError(err)
		}

		spreadsheets := make([]SpreadsheetSummary, 0, len(result.Files))
		rb := response.New()
		rb.Header("Spreadsheets")
		rb.KeyValue("Count", len(result.Files))
		rb.Blank()

		for _, f := range result.Files {
			spreadsheets = append(spreadsheets, SpreadsheetSummary{
				ID:           f.Id,
				Name:         f.Name,
				ModifiedTime: f.ModifiedTime,
				WebViewLink:  f.WebViewLink,
			})
			rb.Item("%s", f.Name)
			rb.Line("    ID: %s | Modified: %s", f.Id, f.ModifiedTime)
		}

		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: rb.Build()}},
		}, ListSpreadsheetsOutput{Spreadsheets: spreadsheets}, nil
	}
}

// --- get_spreadsheet_info (extended) ---

type GetSpreadsheetInfoInput struct {
	UserEmail     string `json:"user_google_email" jsonschema:"required" jsonschema_description:"The user's Google email address"`
	SpreadsheetID string `json:"spreadsheet_id" jsonschema:"required" jsonschema_description:"The spreadsheet ID"`
}

type SheetInfo struct {
	SheetID  int64  `json:"sheet_id"`
	Title    string `json:"title"`
	RowCount int64  `json:"row_count"`
	ColCount int64  `json:"col_count"`
}

type GetSpreadsheetInfoOutput struct {
	Title  string      `json:"title"`
	URL    string      `json:"url"`
	Locale string      `json:"locale"`
	Sheets []SheetInfo `json:"sheets"`
}

func createGetSpreadsheetInfoHandler(factory *services.Factory) mcp.ToolHandlerFor[GetSpreadsheetInfoInput, GetSpreadsheetInfoOutput] {
	return func(ctx context.Context, req *mcp.CallToolRequest, input GetSpreadsheetInfoInput) (*mcp.CallToolResult, GetSpreadsheetInfoOutput, error) {
		srv, err := factory.Sheets(ctx, input.UserEmail)
		if err != nil {
			return nil, GetSpreadsheetInfoOutput{}, middleware.HandleGoogleAPIError(err)
		}

		ss, err := srv.Spreadsheets.Get(input.SpreadsheetID).Context(ctx).Do()
		if err != nil {
			return nil, GetSpreadsheetInfoOutput{}, middleware.HandleGoogleAPIError(err)
		}

		sheetInfos := make([]SheetInfo, 0, len(ss.Sheets))
		rb := response.New()
		rb.Header("Spreadsheet Info")
		rb.KeyValue("Title", ss.Properties.Title)
		rb.KeyValue("ID", ss.SpreadsheetId)
		rb.KeyValue("URL", ss.SpreadsheetUrl)
		rb.KeyValue("Locale", ss.Properties.Locale)
		rb.Blank()
		rb.Section("Sheets")

		for _, s := range ss.Sheets {
			si := SheetInfo{
				SheetID: s.Properties.SheetId,
				Title:   s.Properties.Title,
			}
			if s.Properties.GridProperties != nil {
				si.RowCount = s.Properties.GridProperties.RowCount
				si.ColCount = s.Properties.GridProperties.ColumnCount
			}
			sheetInfos = append(sheetInfos, si)
			rb.Item("%s (ID: %d, %dx%d)", si.Title, si.SheetID, si.RowCount, si.ColCount)
		}

		return &mcp.CallToolResult{
				Content: []mcp.Content{&mcp.TextContent{Text: rb.Build()}},
			}, GetSpreadsheetInfoOutput{
				Title:  ss.Properties.Title,
				URL:    ss.SpreadsheetUrl,
				Locale: ss.Properties.Locale,
				Sheets: sheetInfos,
			}, nil
	}
}

// --- format_sheet_range (extended) ---

type FormatSheetRangeInput struct {
	UserEmail     string `json:"user_google_email" jsonschema:"required" jsonschema_description:"The user's Google email address"`
	SpreadsheetID string `json:"spreadsheet_id" jsonschema:"required" jsonschema_description:"The spreadsheet ID"`
	SheetID       int64  `json:"sheet_id" jsonschema:"required" jsonschema_description:"The sheet ID (tab ID, not name)"`
	StartRow      int64  `json:"start_row" jsonschema:"required" jsonschema_description:"Start row index (0-based)"`
	EndRow        int64  `json:"end_row" jsonschema:"required" jsonschema_description:"End row index (exclusive)"`
	StartCol      int64  `json:"start_col" jsonschema:"required" jsonschema_description:"Start column index (0-based)"`
	EndCol        int64  `json:"end_col" jsonschema:"required" jsonschema_description:"End column index (exclusive)"`
	Bold          *bool  `json:"bold,omitempty" jsonschema_description:"Make text bold"`
	Italic        *bool  `json:"italic,omitempty" jsonschema_description:"Make text italic"`
	FontSize      *int64 `json:"font_size,omitempty" jsonschema_description:"Font size in points"`
	TextColor     string `json:"text_color,omitempty" jsonschema_description:"Text color as hex (#RRGGBB)"`
	BgColor       string `json:"background_color,omitempty" jsonschema_description:"Background color as hex (#RRGGBB)"`
	HAlign        string `json:"horizontal_alignment,omitempty" jsonschema_description:"Horizontal alignment: LEFT CENTER RIGHT"`
	NumberFormat  string `json:"number_format,omitempty" jsonschema_description:"Number format pattern (e.g. #,##0.00)"`
	NumberType    string `json:"number_format_type,omitempty" jsonschema_description:"Number format type: TEXT NUMBER PERCENT CURRENCY DATE TIME DATE_TIME SCIENTIFIC"`
}

func createFormatSheetRangeHandler(factory *services.Factory) mcp.ToolHandlerFor[FormatSheetRangeInput, any] {
	return func(ctx context.Context, req *mcp.CallToolRequest, input FormatSheetRangeInput) (*mcp.CallToolResult, any, error) {
		srv, err := factory.Sheets(ctx, input.UserEmail)
		if err != nil {
			return nil, nil, middleware.HandleGoogleAPIError(err)
		}

		cellFormat := &sheets.CellFormat{}
		fields := make([]string, 0)

		// Text format
		textFormat := &sheets.TextFormat{}
		hasTextFormat := false
		if input.Bold != nil {
			textFormat.Bold = *input.Bold
			hasTextFormat = true
		}
		if input.Italic != nil {
			textFormat.Italic = *input.Italic
			hasTextFormat = true
		}
		if input.FontSize != nil {
			textFormat.FontSize = int64(*input.FontSize)
			hasTextFormat = true
		}
		if input.TextColor != "" {
			textFormat.ForegroundColor = parseSheetColor(input.TextColor)
			hasTextFormat = true
		}
		if hasTextFormat {
			cellFormat.TextFormat = textFormat
			fields = append(fields, "userEnteredFormat.textFormat")
		}

		// Background color
		if input.BgColor != "" {
			cellFormat.BackgroundColor = parseSheetColor(input.BgColor)
			fields = append(fields, "userEnteredFormat.backgroundColor")
		}

		// Alignment
		if input.HAlign != "" {
			cellFormat.HorizontalAlignment = input.HAlign
			fields = append(fields, "userEnteredFormat.horizontalAlignment")
		}

		// Number format
		if input.NumberFormat != "" || input.NumberType != "" {
			nf := &sheets.NumberFormat{}
			if input.NumberFormat != "" {
				nf.Pattern = input.NumberFormat
			}
			if input.NumberType != "" {
				nf.Type = input.NumberType
			}
			cellFormat.NumberFormat = nf
			fields = append(fields, "userEnteredFormat.numberFormat")
		}

		if len(fields) == 0 {
			return nil, nil, fmt.Errorf("no formatting specified — provide at least one formatting parameter")
		}

		batchReq := &sheets.BatchUpdateSpreadsheetRequest{
			Requests: []*sheets.Request{
				{
					RepeatCell: &sheets.RepeatCellRequest{
						Range: &sheets.GridRange{
							SheetId:          input.SheetID,
							StartRowIndex:    input.StartRow,
							EndRowIndex:      input.EndRow,
							StartColumnIndex: input.StartCol,
							EndColumnIndex:   input.EndCol,
						},
						Cell: &sheets.CellData{
							UserEnteredFormat: cellFormat,
						},
						Fields: joinFields(fields),
					},
				},
			},
		}

		_, err = srv.Spreadsheets.BatchUpdate(input.SpreadsheetID, batchReq).Context(ctx).Do()
		if err != nil {
			return nil, nil, middleware.HandleGoogleAPIError(err)
		}

		rb := response.New()
		rb.Header("Range Formatted")
		rb.KeyValue("Spreadsheet", input.SpreadsheetID)
		rb.KeyValue("Range", fmt.Sprintf("Sheet %d: R%d:R%d C%d:C%d", input.SheetID, input.StartRow, input.EndRow, input.StartCol, input.EndCol))

		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: rb.Build()}},
		}, nil, nil
	}
}

// --- add_conditional_formatting (extended) ---

type AddConditionalFormattingInput struct {
	UserEmail     string   `json:"user_google_email" jsonschema:"required" jsonschema_description:"The user's Google email address"`
	SpreadsheetID string   `json:"spreadsheet_id" jsonschema:"required" jsonschema_description:"The spreadsheet ID"`
	SheetID       int64    `json:"sheet_id" jsonschema:"required" jsonschema_description:"The sheet ID"`
	StartRow      int64    `json:"start_row" jsonschema:"required" jsonschema_description:"Start row (0-based)"`
	EndRow        int64    `json:"end_row" jsonschema:"required" jsonschema_description:"End row (exclusive)"`
	StartCol      int64    `json:"start_col" jsonschema:"required" jsonschema_description:"Start column (0-based)"`
	EndCol        int64    `json:"end_col" jsonschema:"required" jsonschema_description:"End column (exclusive)"`
	RuleType      string   `json:"rule_type" jsonschema:"required" jsonschema_description:"Rule type: CUSTOM_FORMULA or NUMBER_GREATER,enum=CUSTOM_FORMULA,enum=NUMBER_GREATER,enum=NUMBER_LESS,enum=TEXT_CONTAINS,enum=TEXT_NOT_CONTAINS"`
	Values        []string `json:"values" jsonschema:"required" jsonschema_description:"Condition values (formula for CUSTOM_FORMULA or threshold values)"`
	BgColor       string   `json:"background_color,omitempty" jsonschema_description:"Background color to apply (#RRGGBB)"`
	TextColor     string   `json:"text_color,omitempty" jsonschema_description:"Text color to apply (#RRGGBB)"`
	Bold          *bool    `json:"bold,omitempty" jsonschema_description:"Make matching text bold"`
}

func createAddConditionalFormattingHandler(factory *services.Factory) mcp.ToolHandlerFor[AddConditionalFormattingInput, any] {
	return func(ctx context.Context, req *mcp.CallToolRequest, input AddConditionalFormattingInput) (*mcp.CallToolResult, any, error) {
		srv, err := factory.Sheets(ctx, input.UserEmail)
		if err != nil {
			return nil, nil, middleware.HandleGoogleAPIError(err)
		}

		condValues := make([]*sheets.ConditionValue, 0, len(input.Values))
		for _, v := range input.Values {
			condValues = append(condValues, &sheets.ConditionValue{UserEnteredValue: v})
		}

		format := &sheets.CellFormat{}
		if input.BgColor != "" {
			format.BackgroundColor = parseSheetColor(input.BgColor)
		}
		textFmt := &sheets.TextFormat{}
		hasTextFmt := false
		if input.TextColor != "" {
			textFmt.ForegroundColor = parseSheetColor(input.TextColor)
			hasTextFmt = true
		}
		if input.Bold != nil {
			textFmt.Bold = *input.Bold
			hasTextFmt = true
		}
		if hasTextFmt {
			format.TextFormat = textFmt
		}

		batchReq := &sheets.BatchUpdateSpreadsheetRequest{
			Requests: []*sheets.Request{
				{
					AddConditionalFormatRule: &sheets.AddConditionalFormatRuleRequest{
						Rule: &sheets.ConditionalFormatRule{
							Ranges: []*sheets.GridRange{
								{
									SheetId:          input.SheetID,
									StartRowIndex:    input.StartRow,
									EndRowIndex:      input.EndRow,
									StartColumnIndex: input.StartCol,
									EndColumnIndex:   input.EndCol,
								},
							},
							BooleanRule: &sheets.BooleanRule{
								Condition: &sheets.BooleanCondition{
									Type:   input.RuleType,
									Values: condValues,
								},
								Format: format,
							},
						},
						Index: 0,
					},
				},
			},
		}

		_, err = srv.Spreadsheets.BatchUpdate(input.SpreadsheetID, batchReq).Context(ctx).Do()
		if err != nil {
			return nil, nil, middleware.HandleGoogleAPIError(err)
		}

		rb := response.New()
		rb.Header("Conditional Formatting Added")
		rb.KeyValue("Spreadsheet", input.SpreadsheetID)
		rb.KeyValue("Rule Type", input.RuleType)

		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: rb.Build()}},
		}, nil, nil
	}
}

// --- update_conditional_formatting (extended) ---

type UpdateConditionalFormattingInput struct {
	UserEmail     string   `json:"user_google_email" jsonschema:"required" jsonschema_description:"The user's Google email address"`
	SpreadsheetID string   `json:"spreadsheet_id" jsonschema:"required" jsonschema_description:"The spreadsheet ID"`
	RuleIndex     int64    `json:"rule_index" jsonschema:"required" jsonschema_description:"Index of the rule to update (0-based)"`
	SheetID       int64    `json:"sheet_id" jsonschema:"required" jsonschema_description:"The sheet ID"`
	StartRow      int64    `json:"start_row" jsonschema:"required" jsonschema_description:"Start row (0-based)"`
	EndRow        int64    `json:"end_row" jsonschema:"required" jsonschema_description:"End row (exclusive)"`
	StartCol      int64    `json:"start_col" jsonschema:"required" jsonschema_description:"Start column (0-based)"`
	EndCol        int64    `json:"end_col" jsonschema:"required" jsonschema_description:"End column (exclusive)"`
	RuleType      string   `json:"rule_type" jsonschema:"required" jsonschema_description:"Rule type"`
	Values        []string `json:"values" jsonschema:"required" jsonschema_description:"Condition values"`
	BgColor       string   `json:"background_color,omitempty" jsonschema_description:"Background color (#RRGGBB)"`
}

func createUpdateConditionalFormattingHandler(factory *services.Factory) mcp.ToolHandlerFor[UpdateConditionalFormattingInput, any] {
	return func(ctx context.Context, req *mcp.CallToolRequest, input UpdateConditionalFormattingInput) (*mcp.CallToolResult, any, error) {
		srv, err := factory.Sheets(ctx, input.UserEmail)
		if err != nil {
			return nil, nil, middleware.HandleGoogleAPIError(err)
		}

		condValues := make([]*sheets.ConditionValue, 0, len(input.Values))
		for _, v := range input.Values {
			condValues = append(condValues, &sheets.ConditionValue{UserEnteredValue: v})
		}

		format := &sheets.CellFormat{}
		if input.BgColor != "" {
			format.BackgroundColor = parseSheetColor(input.BgColor)
		}

		batchReq := &sheets.BatchUpdateSpreadsheetRequest{
			Requests: []*sheets.Request{
				{
					UpdateConditionalFormatRule: &sheets.UpdateConditionalFormatRuleRequest{
						Rule: &sheets.ConditionalFormatRule{
							Ranges: []*sheets.GridRange{
								{
									SheetId:          input.SheetID,
									StartRowIndex:    input.StartRow,
									EndRowIndex:      input.EndRow,
									StartColumnIndex: input.StartCol,
									EndColumnIndex:   input.EndCol,
								},
							},
							BooleanRule: &sheets.BooleanRule{
								Condition: &sheets.BooleanCondition{
									Type:   input.RuleType,
									Values: condValues,
								},
								Format: format,
							},
						},
						Index: input.RuleIndex,
					},
				},
			},
		}

		_, err = srv.Spreadsheets.BatchUpdate(input.SpreadsheetID, batchReq).Context(ctx).Do()
		if err != nil {
			return nil, nil, middleware.HandleGoogleAPIError(err)
		}

		rb := response.New()
		rb.Header("Conditional Formatting Updated")
		rb.KeyValue("Spreadsheet", input.SpreadsheetID)
		rb.KeyValue("Rule Index", input.RuleIndex)

		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: rb.Build()}},
		}, nil, nil
	}
}

// --- delete_conditional_formatting (extended) ---

type DeleteConditionalFormattingInput struct {
	UserEmail     string `json:"user_google_email" jsonschema:"required" jsonschema_description:"The user's Google email address"`
	SpreadsheetID string `json:"spreadsheet_id" jsonschema:"required" jsonschema_description:"The spreadsheet ID"`
	SheetID       int64  `json:"sheet_id" jsonschema:"required" jsonschema_description:"The sheet ID"`
	RuleIndex     int64  `json:"rule_index" jsonschema:"required" jsonschema_description:"Index of the rule to delete (0-based)"`
}

func createDeleteConditionalFormattingHandler(factory *services.Factory) mcp.ToolHandlerFor[DeleteConditionalFormattingInput, any] {
	return func(ctx context.Context, req *mcp.CallToolRequest, input DeleteConditionalFormattingInput) (*mcp.CallToolResult, any, error) {
		srv, err := factory.Sheets(ctx, input.UserEmail)
		if err != nil {
			return nil, nil, middleware.HandleGoogleAPIError(err)
		}

		batchReq := &sheets.BatchUpdateSpreadsheetRequest{
			Requests: []*sheets.Request{
				{
					DeleteConditionalFormatRule: &sheets.DeleteConditionalFormatRuleRequest{
						SheetId: input.SheetID,
						Index:   input.RuleIndex,
					},
				},
			},
		}

		_, err = srv.Spreadsheets.BatchUpdate(input.SpreadsheetID, batchReq).Context(ctx).Do()
		if err != nil {
			return nil, nil, middleware.HandleGoogleAPIError(err)
		}

		rb := response.New()
		rb.Header("Conditional Formatting Deleted")
		rb.KeyValue("Spreadsheet", input.SpreadsheetID)
		rb.KeyValue("Rule Index", input.RuleIndex)

		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: rb.Build()}},
		}, nil, nil
	}
}

// --- helper functions ---

// parseSheetColor converts a hex color (#RRGGBB) to a Sheets Color.
func parseSheetColor(hex string) *sheets.Color {
	hex = trimHash(hex)
	if len(hex) != 6 {
		return nil
	}
	return &sheets.Color{
		Red:   float64(hexVal(hex[0:2])) / 255.0,
		Green: float64(hexVal(hex[2:4])) / 255.0,
		Blue:  float64(hexVal(hex[4:6])) / 255.0,
	}
}

func trimHash(s string) string {
	if s != "" && s[0] == '#' {
		return s[1:]
	}
	return s
}

func hexVal(hex string) byte {
	var val byte
	for _, c := range hex {
		val *= 16
		switch {
		case c >= '0' && c <= '9':
			val += byte(c - '0')
		case c >= 'a' && c <= 'f':
			val += byte(c-'a') + 10
		case c >= 'A' && c <= 'F':
			val += byte(c-'A') + 10
		}
	}
	return val
}

func joinFields(fields []string) string {
	result := ""
	for i, f := range fields {
		if i > 0 {
			result += ","
		}
		result += f
	}
	return result
}
