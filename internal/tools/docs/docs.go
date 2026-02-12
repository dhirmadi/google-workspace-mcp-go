package docs

import (
	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/evert/google-workspace-mcp-go/internal/pkg/ptr"
	"github.com/evert/google-workspace-mcp-go/internal/services"
	"github.com/evert/google-workspace-mcp-go/internal/tools/comments"
)

var serviceIcons = []mcp.Icon{{
	Source:   "https://www.gstatic.com/images/branding/product/1x/docs_2020q4_48dp.png",
	MIMEType: "image/png",
	Sizes:    []string{"48x48"},
}}

// Register registers all Docs tools (core + extended) with the MCP server.
func Register(server *mcp.Server, factory *services.Factory) {
	// --- Core tools ---

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_doc_content",
		Icons:       serviceIcons,
		Description: "Get the full text content of a Google Doc including tables.",
		Annotations: &mcp.ToolAnnotations{
			Title:         "Get Document Content",
			ReadOnlyHint:  true,
			OpenWorldHint: ptr.Bool(true),
		},
	}, createGetDocContentHandler(factory))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "create_doc",
		Icons:       serviceIcons,
		Description: "Create a new Google Doc with an optional initial text content.",
		Annotations: &mcp.ToolAnnotations{
			Title:         "Create Document",
			OpenWorldHint: ptr.Bool(true),
		},
	}, createCreateDocHandler(factory))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "modify_doc_text",
		Icons:       serviceIcons,
		Description: "Insert or replace text in a Google Doc with optional formatting (bold, italic, color, font). Can also format existing text without changing content.",
		Annotations: &mcp.ToolAnnotations{
			Title:         "Modify Document Text",
			OpenWorldHint: ptr.Bool(true),
		},
	}, createModifyDocTextHandler(factory))

	// --- Extended tools ---

	mcp.AddTool(server, &mcp.Tool{
		Name:        "export_doc_to_pdf",
		Icons:       serviceIcons,
		Description: "Get a download URL to export a Google Doc as PDF.",
		Annotations: &mcp.ToolAnnotations{
			Title:         "Export Document to PDF",
			ReadOnlyHint:  true,
			OpenWorldHint: ptr.Bool(true),
		},
	}, createExportDocToPDFHandler(factory))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "search_docs",
		Icons:       serviceIcons,
		Description: "Search for Google Docs in Drive. Uses Drive search query syntax scoped to Google Docs MIME type.",
		Annotations: &mcp.ToolAnnotations{
			Title:         "Search Documents",
			ReadOnlyHint:  true,
			OpenWorldHint: ptr.Bool(true),
		},
	}, createSearchDocsHandler(factory))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "find_and_replace_doc",
		Icons:       serviceIcons,
		Description: "Find and replace all occurrences of text in a Google Doc.",
		Annotations: &mcp.ToolAnnotations{
			Title:         "Find and Replace in Document",
			OpenWorldHint: ptr.Bool(true),
		},
	}, createFindAndReplaceDocHandler(factory))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_docs_in_folder",
		Icons:       serviceIcons,
		Description: "List all Google Docs in a specific Drive folder.",
		Annotations: &mcp.ToolAnnotations{
			Title:         "List Documents in Folder",
			ReadOnlyHint:  true,
			OpenWorldHint: ptr.Bool(true),
		},
	}, createListDocsInFolderHandler(factory))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "insert_doc_elements",
		Icons:       serviceIcons,
		Description: "Insert paragraphs or list items into a Google Doc at specified positions.",
		Annotations: &mcp.ToolAnnotations{
			Title:         "Insert Document Elements",
			OpenWorldHint: ptr.Bool(true),
		},
	}, createInsertDocElementsHandler(factory))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "update_paragraph_style",
		Icons:       serviceIcons,
		Description: "Update paragraph styling (headings, alignment, spacing, indent) in a Google Doc.",
		Annotations: &mcp.ToolAnnotations{
			Title:          "Update Paragraph Style",
			IdempotentHint: true,
			OpenWorldHint:  ptr.Bool(true),
		},
	}, createUpdateParagraphStyleHandler(factory))

	// --- Complete tools ---

	mcp.AddTool(server, &mcp.Tool{
		Name:        "insert_doc_image",
		Icons:       serviceIcons,
		Description: "Insert an image into a Google Doc from a public URL at a specified position.",
		Annotations: &mcp.ToolAnnotations{
			Title:         "Insert Document Image",
			OpenWorldHint: ptr.Bool(true),
		},
	}, createInsertDocImageHandler(factory))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "update_doc_headers_footers",
		Icons:       serviceIcons,
		Description: "Add, update, or remove headers and footers in a Google Doc.",
		Annotations: &mcp.ToolAnnotations{
			Title:          "Update Headers/Footers",
			IdempotentHint: true,
			OpenWorldHint:  ptr.Bool(true),
		},
	}, createUpdateHeadersFootersHandler(factory))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "batch_update_doc",
		Icons:       serviceIcons,
		Description: "Perform batch updates on a Google Doc using a JSON array of update requests.",
		Annotations: &mcp.ToolAnnotations{
			Title:         "Batch Update Document",
			OpenWorldHint: ptr.Bool(true),
		},
	}, createBatchUpdateDocHandler(factory))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "inspect_doc_structure",
		Icons:       serviceIcons,
		Description: "Inspect the structural elements of a Google Doc (paragraphs, tables, sections) with their index ranges.",
		Annotations: &mcp.ToolAnnotations{
			Title:         "Inspect Document Structure",
			ReadOnlyHint:  true,
			OpenWorldHint: ptr.Bool(true),
		},
	}, createInspectDocStructureHandler(factory))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "create_table_with_data",
		Icons:       serviceIcons,
		Description: "Create a table in a Google Doc and optionally populate it with data.",
		Annotations: &mcp.ToolAnnotations{
			Title:         "Create Table with Data",
			OpenWorldHint: ptr.Bool(true),
		},
	}, createCreateTableHandler(factory))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "debug_table_structure",
		Icons:       serviceIcons,
		Description: "Debug the internal structure of a table in a Google Doc showing cell indices and content.",
		Annotations: &mcp.ToolAnnotations{
			Title:         "Debug Table Structure",
			ReadOnlyHint:  true,
			OpenWorldHint: ptr.Bool(true),
		},
	}, createDebugTableStructureHandler(factory))

	// --- Comment tools (via shared Drive API) ---
	comments.Register(server, factory, "document", serviceIcons)
}
