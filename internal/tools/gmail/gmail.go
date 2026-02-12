package gmail

import (
	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/evert/google-workspace-mcp-go/internal/pkg/ptr"
	"github.com/evert/google-workspace-mcp-go/internal/services"
)

var serviceIcons = []mcp.Icon{{
	Source:   "https://www.gstatic.com/images/branding/product/1x/gmail_2020q4_48dp.png",
	MIMEType: "image/png",
	Sizes:    []string{"48x48"},
}}

// Register registers all core Gmail tools with the MCP server.
func Register(server *mcp.Server, factory *services.Factory) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "search_gmail_messages",
		Icons:       serviceIcons,
		Description: "Search Gmail messages using standard Gmail search query syntax. Returns message summaries with IDs for further retrieval.",
		Annotations: &mcp.ToolAnnotations{
			Title:         "Search Gmail Messages",
			ReadOnlyHint:  true,
			OpenWorldHint: ptr.Bool(true),
		},
	}, createSearchMessagesHandler(factory))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_gmail_message_content",
		Icons:       serviceIcons,
		Description: "Get the full content of a specific Gmail message including subject, sender, recipients, body text, and attachment metadata (filenames, MIME types, attachment IDs for retrieval).",
		Annotations: &mcp.ToolAnnotations{
			Title:         "Get Gmail Message Content",
			ReadOnlyHint:  true,
			OpenWorldHint: ptr.Bool(true),
		},
	}, createGetMessageContentHandler(factory))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_gmail_messages_content_batch",
		Icons:       serviceIcons,
		Description: "Get the content of multiple Gmail messages in a single request. Supports up to 25 messages per batch. Reports progress during retrieval. Includes attachment metadata when present.",
		Annotations: &mcp.ToolAnnotations{
			Title:         "Get Gmail Messages (Batch)",
			ReadOnlyHint:  true,
			OpenWorldHint: ptr.Bool(true),
		},
	}, createBatchGetMessagesHandler(factory))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "send_gmail_message",
		Icons:       serviceIcons,
		Description: "Send an email using the user's Gmail account. Supports new emails and replies with threading.",
		Annotations: &mcp.ToolAnnotations{
			Title:         "Send Gmail Message",
			OpenWorldHint: ptr.Bool(true),
		},
	}, createSendMessageHandler(factory))

	// --- Extended tools ---

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_gmail_attachment_content",
		Icons:       serviceIcons,
		Description: "Get the content of a Gmail message attachment by attachment ID. Automatically extracts text from Office documents (.docx/.xlsx/.pptx) and text files. Returns images as inline image content for vision-capable models. Use get_gmail_message_content first to discover attachment IDs.",
		Annotations: &mcp.ToolAnnotations{
			Title:         "Get Gmail Attachment",
			ReadOnlyHint:  true,
			OpenWorldHint: ptr.Bool(true),
		},
	}, createGetAttachmentHandler(factory))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_gmail_thread_content",
		Icons:       serviceIcons,
		Description: "Get all messages in a Gmail thread, including full body content and attachment metadata for each message.",
		Annotations: &mcp.ToolAnnotations{
			Title:         "Get Gmail Thread",
			ReadOnlyHint:  true,
			OpenWorldHint: ptr.Bool(true),
		},
	}, createGetThreadHandler(factory))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "modify_gmail_message_labels",
		Icons:       serviceIcons,
		Description: "Add or remove labels from a Gmail message.",
		Annotations: &mcp.ToolAnnotations{
			Title:          "Modify Message Labels",
			IdempotentHint: true,
			OpenWorldHint:  ptr.Bool(true),
		},
	}, createModifyLabelsHandler(factory))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_gmail_labels",
		Icons:       serviceIcons,
		Description: "List all Gmail labels including system and user-created labels.",
		Annotations: &mcp.ToolAnnotations{
			Title:         "List Gmail Labels",
			ReadOnlyHint:  true,
			OpenWorldHint: ptr.Bool(true),
		},
	}, createListLabelsHandler(factory))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "manage_gmail_label",
		Icons:       serviceIcons,
		Description: "Create, update, or delete a Gmail label.",
		Annotations: &mcp.ToolAnnotations{
			Title:         "Manage Gmail Label",
			OpenWorldHint: ptr.Bool(true),
		},
	}, createManageLabelHandler(factory))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "draft_gmail_message",
		Icons:       serviceIcons,
		Description: "Create a draft email message that can be edited and sent later.",
		Annotations: &mcp.ToolAnnotations{
			Title:         "Draft Gmail Message",
			OpenWorldHint: ptr.Bool(true),
		},
	}, createDraftMessageHandler(factory))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_gmail_filters",
		Icons:       serviceIcons,
		Description: "List all email filters configured for the user.",
		Annotations: &mcp.ToolAnnotations{
			Title:         "List Gmail Filters",
			ReadOnlyHint:  true,
			OpenWorldHint: ptr.Bool(true),
		},
	}, createListFiltersHandler(factory))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "create_gmail_filter",
		Icons:       serviceIcons,
		Description: "Create an email filter to automatically process matching messages.",
		Annotations: &mcp.ToolAnnotations{
			Title:         "Create Gmail Filter",
			OpenWorldHint: ptr.Bool(true),
		},
	}, createCreateFilterHandler(factory))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "delete_gmail_filter",
		Icons:       serviceIcons,
		Description: "Permanently delete an email filter.",
		Annotations: &mcp.ToolAnnotations{
			Title:           "Delete Gmail Filter",
			DestructiveHint: ptr.Bool(true),
			OpenWorldHint:   ptr.Bool(true),
		},
	}, createDeleteFilterHandler(factory))

	// --- Complete tools ---

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_gmail_threads_content_batch",
		Icons:       serviceIcons,
		Description: "Get the content of multiple Gmail threads in a single request. Supports up to 25 threads per batch. Reports progress during retrieval.",
		Annotations: &mcp.ToolAnnotations{
			Title:         "Get Gmail Threads (Batch)",
			ReadOnlyHint:  true,
			OpenWorldHint: ptr.Bool(true),
		},
	}, createBatchGetThreadsHandler(factory))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "batch_modify_gmail_message_labels",
		Icons:       serviceIcons,
		Description: "Add or remove labels from multiple Gmail messages in a single batch operation.",
		Annotations: &mcp.ToolAnnotations{
			Title:          "Batch Modify Message Labels",
			IdempotentHint: true,
			OpenWorldHint:  ptr.Bool(true),
		},
	}, createBatchModifyLabelsHandler(factory))
}
