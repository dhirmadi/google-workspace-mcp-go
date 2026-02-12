package slides

import (
	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/evert/google-workspace-mcp-go/internal/pkg/ptr"
	"github.com/evert/google-workspace-mcp-go/internal/services"
	"github.com/evert/google-workspace-mcp-go/internal/tools/comments"
)

var serviceIcons = []mcp.Icon{{
	Source:   "https://www.gstatic.com/images/branding/product/1x/slides_2020q4_48dp.png",
	MIMEType: "image/png",
	Sizes:    []string{"48x48"},
}}

// Register registers all Slides tools (core + extended + complete) with the MCP server.
func Register(server *mcp.Server, factory *services.Factory) {
	// --- Core tools ---

	mcp.AddTool(server, &mcp.Tool{
		Name:        "create_presentation",
		Icons:       serviceIcons,
		Description: "Create a new Google Slides presentation.",
		Annotations: &mcp.ToolAnnotations{
			Title:         "Create Presentation",
			OpenWorldHint: ptr.Bool(true),
		},
	}, createCreatePresentationHandler(factory))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_presentation",
		Icons:       serviceIcons,
		Description: "Get details of a Google Slides presentation including all slides, layouts, and masters.",
		Annotations: &mcp.ToolAnnotations{
			Title:         "Get Presentation",
			ReadOnlyHint:  true,
			OpenWorldHint: ptr.Bool(true),
		},
	}, createGetPresentationHandler(factory))

	// --- Extended tools ---

	mcp.AddTool(server, &mcp.Tool{
		Name:        "batch_update_presentation",
		Icons:       serviceIcons,
		Description: "Perform batch updates on a Google Slides presentation: create slides, insert text, images, shapes, tables, and more.",
		Annotations: &mcp.ToolAnnotations{
			Title:         "Batch Update Presentation",
			OpenWorldHint: ptr.Bool(true),
		},
	}, createBatchUpdatePresentationHandler(factory))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_page",
		Icons:       serviceIcons,
		Description: "Get details of a specific slide/page from a Google Slides presentation.",
		Annotations: &mcp.ToolAnnotations{
			Title:         "Get Slide Page",
			ReadOnlyHint:  true,
			OpenWorldHint: ptr.Bool(true),
		},
	}, createGetPageHandler(factory))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_page_thumbnail",
		Icons:       serviceIcons,
		Description: "Get a thumbnail image URL for a specific slide in a Google Slides presentation.",
		Annotations: &mcp.ToolAnnotations{
			Title:         "Get Slide Thumbnail",
			ReadOnlyHint:  true,
			OpenWorldHint: ptr.Bool(true),
		},
	}, createGetPageThumbnailHandler(factory))

	// --- Comment tools (via shared Drive API) ---
	comments.Register(server, factory, "presentation", serviceIcons)
}
