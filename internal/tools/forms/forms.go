package forms

import (
	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/evert/google-workspace-mcp-go/internal/services"
)

var serviceIcons = []mcp.Icon{{
	Source:   "https://www.gstatic.com/images/branding/product/1x/forms_2020q4_48dp.png",
	MIMEType: "image/png",
	Sizes:    []string{"48x48"},
}}

// Register registers all Forms tools (core + extended + complete) with the MCP server.
func Register(server *mcp.Server, factory *services.Factory) {
	// --- Core tools ---

	mcp.AddTool(server, &mcp.Tool{
		Name:        "create_form",
		Icons:       serviceIcons,
		Description: "Create a new Google Form with a title and optional description.",
		Annotations: &mcp.ToolAnnotations{
			Title:         "Create Form",
			OpenWorldHint: ptrBool(true),
		},
	}, createCreateFormHandler(factory))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_form",
		Icons:       serviceIcons,
		Description: "Get details of a Google Form including title, description, and all questions/items.",
		Annotations: &mcp.ToolAnnotations{
			Title:         "Get Form",
			ReadOnlyHint:  true,
			OpenWorldHint: ptrBool(true),
		},
	}, createGetFormHandler(factory))

	// --- Extended tools ---

	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_form_responses",
		Icons:       serviceIcons,
		Description: "List all responses submitted to a Google Form.",
		Annotations: &mcp.ToolAnnotations{
			Title:         "List Form Responses",
			ReadOnlyHint:  true,
			OpenWorldHint: ptrBool(true),
		},
	}, createListFormResponsesHandler(factory))

	// --- Complete tools ---

	mcp.AddTool(server, &mcp.Tool{
		Name:        "set_publish_settings",
		Icons:       serviceIcons,
		Description: "Set publish settings for a Google Form (accepting responses, quiz mode).",
		Annotations: &mcp.ToolAnnotations{
			Title:          "Set Form Publish Settings",
			IdempotentHint: true,
			OpenWorldHint:  ptrBool(true),
		},
	}, createSetPublishSettingsHandler(factory))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_form_response",
		Icons:       serviceIcons,
		Description: "Get a single form response by its response ID.",
		Annotations: &mcp.ToolAnnotations{
			Title:         "Get Form Response",
			ReadOnlyHint:  true,
			OpenWorldHint: ptrBool(true),
		},
	}, createGetFormResponseHandler(factory))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "batch_update_form",
		Icons:       serviceIcons,
		Description: "Perform batch updates on a Google Form: add/update/delete items, update info, or modify settings.",
		Annotations: &mcp.ToolAnnotations{
			Title:         "Batch Update Form",
			OpenWorldHint: ptrBool(true),
		},
	}, createBatchUpdateFormHandler(factory))
}

func ptrBool(b bool) *bool { return &b }
