package calendar

import (
	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/evert/google-workspace-mcp-go/internal/services"
)

var serviceIcons = []mcp.Icon{{
	Source:   "https://www.gstatic.com/images/branding/product/1x/calendar_2020q4_48dp.png",
	MIMEType: "image/png",
	Sizes:    []string{"48x48"},
}}

// Register registers all core Calendar tools with the MCP server.
func Register(server *mcp.Server, factory *services.Factory) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_calendars",
		Icons:       serviceIcons,
		Description: "List all calendars accessible to the authenticated user, including primary calendar, shared calendars, and subscribed calendars.",
		Annotations: &mcp.ToolAnnotations{
			Title:         "List Calendars",
			ReadOnlyHint:  true,
			OpenWorldHint: ptrBool(true),
		},
	}, createListCalendarsHandler(factory))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_events",
		Icons:       serviceIcons,
		Description: "Get calendar events. Can retrieve a single event by ID or list events within a time range with optional search.",
		Annotations: &mcp.ToolAnnotations{
			Title:         "Get Calendar Events",
			ReadOnlyHint:  true,
			OpenWorldHint: ptrBool(true),
		},
	}, createGetEventsHandler(factory))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "create_event",
		Icons:       serviceIcons,
		Description: "Create a new calendar event with optional attendees, location, reminders, and Google Meet link.",
		Annotations: &mcp.ToolAnnotations{
			Title:         "Create Calendar Event",
			OpenWorldHint: ptrBool(true),
		},
	}, createCreateEventHandler(factory))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "modify_event",
		Icons:       serviceIcons,
		Description: "Modify an existing calendar event. Only specified fields are updated; omitted fields remain unchanged.",
		Annotations: &mcp.ToolAnnotations{
			Title:          "Modify Calendar Event",
			IdempotentHint: true,
			OpenWorldHint:  ptrBool(true),
		},
	}, createModifyEventHandler(factory))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "delete_event",
		Icons:       serviceIcons,
		Description: "Permanently delete a calendar event. This action cannot be undone.",
		Annotations: &mcp.ToolAnnotations{
			Title:           "Delete Calendar Event",
			DestructiveHint: ptrBool(true),
			OpenWorldHint:   ptrBool(true),
		},
	}, createDeleteEventHandler(factory))

	// --- Extended tools ---

	mcp.AddTool(server, &mcp.Tool{
		Name:        "query_freebusy",
		Icons:       serviceIcons,
		Description: "Query free/busy times for one or more calendars within a time range.",
		Annotations: &mcp.ToolAnnotations{
			Title:         "Query Free/Busy",
			ReadOnlyHint:  true,
			OpenWorldHint: ptrBool(true),
		},
	}, createQueryFreeBusyHandler(factory))
}

func ptrBool(b bool) *bool { return &b }
