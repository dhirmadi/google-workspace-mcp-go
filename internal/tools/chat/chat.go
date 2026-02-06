package chat

import (
	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/evert/google-workspace-mcp-go/internal/services"
)

var serviceIcons = []mcp.Icon{{
	Source:   "https://www.gstatic.com/images/branding/product/1x/chat_2020q4_48dp.png",
	MIMEType: "image/png",
	Sizes:    []string{"48x48"},
}}

// Register registers all Chat tools (core) with the MCP server.
// All Chat tools use the chat_ prefix to avoid naming collisions with Gmail tools.
// NOTE: The Chat API requires the app to be configured as a Chat app in the
// Google Workspace Admin Console. It does NOT work with consumer Gmail accounts.
func Register(server *mcp.Server, factory *services.Factory) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_chat_spaces",
		Icons:       serviceIcons,
		Description: "List Google Chat spaces (rooms, DMs) accessible to the user. Required to get space names for other chat operations.",
		Annotations: &mcp.ToolAnnotations{
			Title:         "List Chat Spaces",
			ReadOnlyHint:  true,
			OpenWorldHint: ptrBool(true),
		},
	}, createListChatSpacesHandler(factory))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_chat_messages",
		Icons:       serviceIcons,
		Description: "Get messages from a Google Chat space with pagination support.",
		Annotations: &mcp.ToolAnnotations{
			Title:         "Get Chat Messages",
			ReadOnlyHint:  true,
			OpenWorldHint: ptrBool(true),
		},
	}, createGetChatMessagesHandler(factory))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "search_chat_messages",
		Icons:       serviceIcons,
		Description: "Search for messages across Google Chat spaces by text content.",
		Annotations: &mcp.ToolAnnotations{
			Title:         "Search Chat Messages",
			ReadOnlyHint:  true,
			OpenWorldHint: ptrBool(true),
		},
	}, createSearchChatMessagesHandler(factory))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "send_chat_message",
		Icons:       serviceIcons,
		Description: "Send a message to a Google Chat space. Can send to a specific thread or create a new one.",
		Annotations: &mcp.ToolAnnotations{
			Title:         "Send Chat Message",
			OpenWorldHint: ptrBool(true),
		},
	}, createSendChatMessageHandler(factory))
}
