package gmail

import (
	"github.com/modelcontextprotocol/go-sdk/mcp"

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
			OpenWorldHint: ptrBool(true),
		},
	}, createSearchMessagesHandler(factory))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_gmail_message_content",
		Icons:       serviceIcons,
		Description: "Get the full content of a specific Gmail message including subject, sender, recipients, and body text.",
		Annotations: &mcp.ToolAnnotations{
			Title:         "Get Gmail Message Content",
			ReadOnlyHint:  true,
			OpenWorldHint: ptrBool(true),
		},
	}, createGetMessageContentHandler(factory))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_gmail_messages_content_batch",
		Icons:       serviceIcons,
		Description: "Get the content of multiple Gmail messages in a single request. Supports up to 25 messages per batch. Reports progress during retrieval.",
		Annotations: &mcp.ToolAnnotations{
			Title:         "Get Gmail Messages (Batch)",
			ReadOnlyHint:  true,
			OpenWorldHint: ptrBool(true),
		},
	}, createBatchGetMessagesHandler(factory))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "send_gmail_message",
		Icons:       serviceIcons,
		Description: "Send an email using the user's Gmail account. Supports new emails and replies with threading.",
		Annotations: &mcp.ToolAnnotations{
			Title:         "Send Gmail Message",
			OpenWorldHint: ptrBool(true),
		},
	}, createSendMessageHandler(factory))

	// --- Extended tools ---

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_gmail_attachment_content",
		Icons:       serviceIcons,
		Description: "Get the content of a Gmail message attachment by attachment ID.",
		Annotations: &mcp.ToolAnnotations{
			Title:         "Get Gmail Attachment",
			ReadOnlyHint:  true,
			OpenWorldHint: ptrBool(true),
		},
	}, createGetAttachmentHandler(factory))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_gmail_thread_content",
		Icons:       serviceIcons,
		Description: "Get all messages in a Gmail thread, including full body content for each message.",
		Annotations: &mcp.ToolAnnotations{
			Title:         "Get Gmail Thread",
			ReadOnlyHint:  true,
			OpenWorldHint: ptrBool(true),
		},
	}, createGetThreadHandler(factory))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "modify_gmail_message_labels",
		Icons:       serviceIcons,
		Description: "Add or remove labels from a Gmail message.",
		Annotations: &mcp.ToolAnnotations{
			Title:          "Modify Message Labels",
			IdempotentHint: true,
			OpenWorldHint:  ptrBool(true),
		},
	}, createModifyLabelsHandler(factory))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_gmail_labels",
		Icons:       serviceIcons,
		Description: "List all Gmail labels including system and user-created labels.",
		Annotations: &mcp.ToolAnnotations{
			Title:         "List Gmail Labels",
			ReadOnlyHint:  true,
			OpenWorldHint: ptrBool(true),
		},
	}, createListLabelsHandler(factory))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "manage_gmail_label",
		Icons:       serviceIcons,
		Description: "Create, update, or delete a Gmail label.",
		Annotations: &mcp.ToolAnnotations{
			Title:         "Manage Gmail Label",
			OpenWorldHint: ptrBool(true),
		},
	}, createManageLabelHandler(factory))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "draft_gmail_message",
		Icons:       serviceIcons,
		Description: "Create a draft email message that can be edited and sent later.",
		Annotations: &mcp.ToolAnnotations{
			Title:         "Draft Gmail Message",
			OpenWorldHint: ptrBool(true),
		},
	}, createDraftMessageHandler(factory))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_gmail_filters",
		Icons:       serviceIcons,
		Description: "List all email filters configured for the user.",
		Annotations: &mcp.ToolAnnotations{
			Title:         "List Gmail Filters",
			ReadOnlyHint:  true,
			OpenWorldHint: ptrBool(true),
		},
	}, createListFiltersHandler(factory))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "create_gmail_filter",
		Icons:       serviceIcons,
		Description: "Create an email filter to automatically process matching messages.",
		Annotations: &mcp.ToolAnnotations{
			Title:         "Create Gmail Filter",
			OpenWorldHint: ptrBool(true),
		},
	}, createCreateFilterHandler(factory))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "delete_gmail_filter",
		Icons:       serviceIcons,
		Description: "Permanently delete an email filter.",
		Annotations: &mcp.ToolAnnotations{
			Title:           "Delete Gmail Filter",
			DestructiveHint: ptrBool(true),
			OpenWorldHint:   ptrBool(true),
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
			OpenWorldHint: ptrBool(true),
		},
	}, createBatchGetThreadsHandler(factory))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "batch_modify_gmail_message_labels",
		Icons:       serviceIcons,
		Description: "Add or remove labels from multiple Gmail messages in a single batch operation.",
		Annotations: &mcp.ToolAnnotations{
			Title:          "Batch Modify Message Labels",
			IdempotentHint: true,
			OpenWorldHint:  ptrBool(true),
		},
	}, createBatchModifyLabelsHandler(factory))
}

func ptrBool(b bool) *bool { return &b }
