package sheets

import (
	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/evert/google-workspace-mcp-go/internal/pkg/ptr"
	"github.com/evert/google-workspace-mcp-go/internal/services"
	"github.com/evert/google-workspace-mcp-go/internal/tools/comments"
)

var serviceIcons = []mcp.Icon{{
	Source:   "https://www.gstatic.com/images/branding/product/1x/sheets_2020q4_48dp.png",
	MIMEType: "image/png",
	Sizes:    []string{"48x48"},
}}

// Register registers all Sheets tools (core + extended + comments) with the MCP server.
func Register(server *mcp.Server, factory *services.Factory) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "create_spreadsheet",
		Icons:       serviceIcons,
		Description: "Create a new Google Spreadsheet with optional sheet tab names.",
		Annotations: &mcp.ToolAnnotations{
			Title:         "Create Spreadsheet",
			OpenWorldHint: ptr.Bool(true),
		},
	}, createCreateSpreadsheetHandler(factory))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "read_sheet_values",
		Icons:       serviceIcons,
		Description: "Read cell values from a specific range in a Google Sheet. Returns raw values in a 2D array.",
		Annotations: &mcp.ToolAnnotations{
			Title:         "Read Sheet Values",
			ReadOnlyHint:  true,
			OpenWorldHint: ptr.Bool(true),
		},
	}, createReadSheetValuesHandler(factory))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "modify_sheet_values",
		Icons:       serviceIcons,
		Description: "Write, update, or clear values in a Google Sheet range. Supports raw or user-entered value interpretation.",
		Annotations: &mcp.ToolAnnotations{
			Title:          "Modify Sheet Values",
			IdempotentHint: true,
			OpenWorldHint:  ptr.Bool(true),
		},
	}, createModifySheetValuesHandler(factory))

	// --- Extended tools ---

	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_spreadsheets",
		Icons:       serviceIcons,
		Description: "List Google Spreadsheets in the user's Drive.",
		Annotations: &mcp.ToolAnnotations{
			Title:         "List Spreadsheets",
			ReadOnlyHint:  true,
			OpenWorldHint: ptr.Bool(true),
		},
	}, createListSpreadsheetsHandler(factory))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_spreadsheet_info",
		Icons:       serviceIcons,
		Description: "Get spreadsheet metadata including sheet names, dimensions, and properties.",
		Annotations: &mcp.ToolAnnotations{
			Title:         "Get Spreadsheet Info",
			ReadOnlyHint:  true,
			OpenWorldHint: ptr.Bool(true),
		},
	}, createGetSpreadsheetInfoHandler(factory))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "format_sheet_range",
		Icons:       serviceIcons,
		Description: "Format cells in a range (bold, italic, colors, alignment, number format, borders).",
		Annotations: &mcp.ToolAnnotations{
			Title:         "Format Sheet Range",
			OpenWorldHint: ptr.Bool(true),
		},
	}, createFormatSheetRangeHandler(factory))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "add_conditional_formatting",
		Icons:       serviceIcons,
		Description: "Add a conditional formatting rule to a sheet range.",
		Annotations: &mcp.ToolAnnotations{
			Title:         "Add Conditional Formatting",
			OpenWorldHint: ptr.Bool(true),
		},
	}, createAddConditionalFormattingHandler(factory))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "update_conditional_formatting",
		Icons:       serviceIcons,
		Description: "Update an existing conditional formatting rule by index.",
		Annotations: &mcp.ToolAnnotations{
			Title:          "Update Conditional Formatting",
			IdempotentHint: true,
			OpenWorldHint:  ptr.Bool(true),
		},
	}, createUpdateConditionalFormattingHandler(factory))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "delete_conditional_formatting",
		Icons:       serviceIcons,
		Description: "Delete a conditional formatting rule by index.",
		Annotations: &mcp.ToolAnnotations{
			Title:           "Delete Conditional Formatting",
			DestructiveHint: ptr.Bool(true),
			OpenWorldHint:   ptr.Bool(true),
		},
	}, createDeleteConditionalFormattingHandler(factory))

	// --- Complete tools ---

	mcp.AddTool(server, &mcp.Tool{
		Name:        "create_sheet",
		Icons:       serviceIcons,
		Description: "Create a new sheet tab in an existing Google Spreadsheet.",
		Annotations: &mcp.ToolAnnotations{
			Title:         "Create Sheet Tab",
			OpenWorldHint: ptr.Bool(true),
		},
	}, createCreateSheetHandler(factory))

	// --- Comment tools (via shared Drive API) ---
	comments.Register(server, factory, "spreadsheet", serviceIcons)
}
