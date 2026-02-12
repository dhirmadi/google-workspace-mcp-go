package drive

import (
	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/evert/google-workspace-mcp-go/internal/pkg/ptr"
	"github.com/evert/google-workspace-mcp-go/internal/services"
)

var serviceIcons = []mcp.Icon{{
	Source:   "https://www.gstatic.com/images/branding/product/1x/drive_2020q4_48dp.png",
	MIMEType: "image/png",
	Sizes:    []string{"48x48"},
}}

// Register registers all core Drive tools with the MCP server.
func Register(server *mcp.Server, factory *services.Factory) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "search_drive_files",
		Icons:       serviceIcons,
		Description: "Search for files and folders in Google Drive using Drive query syntax. Returns file metadata including IDs for further operations.",
		Annotations: &mcp.ToolAnnotations{
			Title:         "Search Drive Files",
			ReadOnlyHint:  true,
			OpenWorldHint: ptr.Bool(true),
		},
	}, createSearchFilesHandler(factory))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_drive_file_content",
		Icons:       serviceIcons,
		Description: "Get the text content of a Google Drive file. Exports Google Docs/Sheets/Slides as text. Extracts text from Office files (.docx, .xlsx, .pptx).",
		Annotations: &mcp.ToolAnnotations{
			Title:         "Get Drive File Content",
			ReadOnlyHint:  true,
			OpenWorldHint: ptr.Bool(true),
		},
	}, createGetFileContentHandler(factory))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_drive_file_download_url",
		Icons:       serviceIcons,
		Description: "Get a download URL for a Google Drive file. For Google native files, exports to a useful format (PDF, DOCX, etc.).",
		Annotations: &mcp.ToolAnnotations{
			Title:         "Get Drive File Download URL",
			ReadOnlyHint:  true,
			OpenWorldHint: ptr.Bool(true),
		},
	}, createGetDownloadURLHandler(factory))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "create_drive_file",
		Icons:       serviceIcons,
		Description: "Create a new file in Google Drive with optional content. Supports text files and Google Workspace native types.",
		Annotations: &mcp.ToolAnnotations{
			Title:         "Create Drive File",
			OpenWorldHint: ptr.Bool(true),
		},
	}, createCreateFileHandler(factory))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "import_to_google_doc",
		Icons:       serviceIcons,
		Description: "Import a file from Google Drive into Google Doc format. Creates a copy as a Google Doc.",
		Annotations: &mcp.ToolAnnotations{
			Title:         "Import to Google Doc",
			OpenWorldHint: ptr.Bool(true),
		},
	}, createImportToDocHandler(factory))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "share_drive_file",
		Icons:       serviceIcons,
		Description: "Share a Google Drive file or folder with a user, group, domain, or anyone with the link.",
		Annotations: &mcp.ToolAnnotations{
			Title:          "Share Drive File",
			IdempotentHint: true,
			OpenWorldHint:  ptr.Bool(true),
		},
	}, createShareFileHandler(factory))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_drive_shareable_link",
		Icons:       serviceIcons,
		Description: "Get the shareable link for a Google Drive file or folder, along with its current sharing permissions.",
		Annotations: &mcp.ToolAnnotations{
			Title:         "Get Drive Shareable Link",
			ReadOnlyHint:  true,
			OpenWorldHint: ptr.Bool(true),
		},
	}, createGetShareableLinkHandler(factory))

	// --- Extended tools ---

	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_drive_items",
		Icons:       serviceIcons,
		Description: "List files and folders in a specific Drive folder with pagination.",
		Annotations: &mcp.ToolAnnotations{
			Title:         "List Drive Items",
			ReadOnlyHint:  true,
			OpenWorldHint: ptr.Bool(true),
		},
	}, createListDriveItemsHandler(factory))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "copy_drive_file",
		Icons:       serviceIcons,
		Description: "Create a copy of a Google Drive file, optionally in a different folder.",
		Annotations: &mcp.ToolAnnotations{
			Title:         "Copy Drive File",
			OpenWorldHint: ptr.Bool(true),
		},
	}, createCopyFileHandler(factory))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "update_drive_file",
		Icons:       serviceIcons,
		Description: "Update a file's name, content, or location in Google Drive.",
		Annotations: &mcp.ToolAnnotations{
			Title:          "Update Drive File",
			IdempotentHint: true,
			OpenWorldHint:  ptr.Bool(true),
		},
	}, createUpdateFileHandler(factory))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "update_drive_permission",
		Icons:       serviceIcons,
		Description: "Modify an existing sharing permission on a Drive file (e.g. change role from reader to writer).",
		Annotations: &mcp.ToolAnnotations{
			Title:          "Update Drive Permission",
			IdempotentHint: true,
			OpenWorldHint:  ptr.Bool(true),
		},
	}, createUpdatePermissionHandler(factory))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "remove_drive_permission",
		Icons:       serviceIcons,
		Description: "Remove a sharing permission from a Drive file or folder.",
		Annotations: &mcp.ToolAnnotations{
			Title:           "Remove Drive Permission",
			DestructiveHint: ptr.Bool(true),
			OpenWorldHint:   ptr.Bool(true),
		},
	}, createRemovePermissionHandler(factory))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "transfer_drive_ownership",
		Icons:       serviceIcons,
		Description: "Transfer ownership of a Drive file to another user.",
		Annotations: &mcp.ToolAnnotations{
			Title:         "Transfer Drive Ownership",
			OpenWorldHint: ptr.Bool(true),
		},
	}, createTransferOwnershipHandler(factory))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "batch_share_drive_file",
		Icons:       serviceIcons,
		Description: "Share multiple Drive files with a user in a single operation. Reports progress during batch processing.",
		Annotations: &mcp.ToolAnnotations{
			Title:         "Batch Share Drive Files",
			OpenWorldHint: ptr.Bool(true),
		},
	}, createBatchShareHandler(factory))

	// --- Complete tools ---

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_drive_file_permissions",
		Icons:       serviceIcons,
		Description: "List all permissions on a Drive file including role, type, email, and domain details.",
		Annotations: &mcp.ToolAnnotations{
			Title:         "Get File Permissions",
			ReadOnlyHint:  true,
			OpenWorldHint: ptr.Bool(true),
		},
	}, createGetFilePermissionsHandler(factory))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "check_drive_file_public_access",
		Icons:       serviceIcons,
		Description: "Check if a Drive file is publicly accessible or shared with a domain.",
		Annotations: &mcp.ToolAnnotations{
			Title:         "Check Public Access",
			ReadOnlyHint:  true,
			OpenWorldHint: ptr.Bool(true),
		},
	}, createCheckPublicAccessHandler(factory))
}
