package contacts

import (
	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/evert/google-workspace-mcp-go/internal/pkg/ptr"
	"github.com/evert/google-workspace-mcp-go/internal/services"
)

var serviceIcons = []mcp.Icon{{
	Source:   "https://www.gstatic.com/images/branding/product/1x/contacts_2022_48dp.png",
	MIMEType: "image/png",
	Sizes:    []string{"48x48"},
}}

// Register registers all Contacts tools (core + extended) with the MCP server.
// Uses the Google People API (the legacy Contacts API is deprecated).
func Register(server *mcp.Server, factory *services.Factory) {
	// --- Core tools ---

	mcp.AddTool(server, &mcp.Tool{
		Name:        "search_contacts",
		Icons:       serviceIcons,
		Description: "Search contacts by name or email address using the Google People API.",
		Annotations: &mcp.ToolAnnotations{
			Title:         "Search Contacts",
			ReadOnlyHint:  true,
			OpenWorldHint: ptr.Bool(true),
		},
	}, createSearchContactsHandler(factory))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_contact",
		Icons:       serviceIcons,
		Description: "Get detailed information about a specific contact including name, email, phone, and organization.",
		Annotations: &mcp.ToolAnnotations{
			Title:         "Get Contact",
			ReadOnlyHint:  true,
			OpenWorldHint: ptr.Bool(true),
		},
	}, createGetContactHandler(factory))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_contacts",
		Icons:       serviceIcons,
		Description: "List all contacts for the user with pagination support.",
		Annotations: &mcp.ToolAnnotations{
			Title:         "List Contacts",
			ReadOnlyHint:  true,
			OpenWorldHint: ptr.Bool(true),
		},
	}, createListContactsHandler(factory))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "create_contact",
		Icons:       serviceIcons,
		Description: "Create a new contact with name, email, phone, and organization details.",
		Annotations: &mcp.ToolAnnotations{
			Title:         "Create Contact",
			OpenWorldHint: ptr.Bool(true),
		},
	}, createCreateContactHandler(factory))

	// --- Extended tools ---

	mcp.AddTool(server, &mcp.Tool{
		Name:        "update_contact",
		Icons:       serviceIcons,
		Description: "Update an existing contact's details. Fetches current data first to preserve the etag.",
		Annotations: &mcp.ToolAnnotations{
			Title:          "Update Contact",
			IdempotentHint: true,
			OpenWorldHint:  ptr.Bool(true),
		},
	}, createUpdateContactHandler(factory))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "delete_contact",
		Icons:       serviceIcons,
		Description: "Permanently delete a contact. This action cannot be undone.",
		Annotations: &mcp.ToolAnnotations{
			Title:           "Delete Contact",
			DestructiveHint: ptr.Bool(true),
			OpenWorldHint:   ptr.Bool(true),
		},
	}, createDeleteContactHandler(factory))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_contact_groups",
		Icons:       serviceIcons,
		Description: "List all contact groups (labels) for the user.",
		Annotations: &mcp.ToolAnnotations{
			Title:         "List Contact Groups",
			ReadOnlyHint:  true,
			OpenWorldHint: ptr.Bool(true),
		},
	}, createListContactGroupsHandler(factory))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_contact_group",
		Icons:       serviceIcons,
		Description: "Get details of a specific contact group including member count.",
		Annotations: &mcp.ToolAnnotations{
			Title:         "Get Contact Group",
			ReadOnlyHint:  true,
			OpenWorldHint: ptr.Bool(true),
		},
	}, createGetContactGroupHandler(factory))

	// --- Complete tools ---

	mcp.AddTool(server, &mcp.Tool{
		Name:        "batch_create_contacts",
		Icons:       serviceIcons,
		Description: "Create multiple contacts in a single batch operation (max 200).",
		Annotations: &mcp.ToolAnnotations{
			Title:         "Batch Create Contacts",
			OpenWorldHint: ptr.Bool(true),
		},
	}, createBatchCreateContactsHandler(factory))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "batch_update_contacts",
		Icons:       serviceIcons,
		Description: "Update multiple contacts in a single batch operation.",
		Annotations: &mcp.ToolAnnotations{
			Title:          "Batch Update Contacts",
			IdempotentHint: true,
			OpenWorldHint:  ptr.Bool(true),
		},
	}, createBatchUpdateContactsHandler(factory))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "batch_delete_contacts",
		Icons:       serviceIcons,
		Description: "Delete multiple contacts in a single batch operation.",
		Annotations: &mcp.ToolAnnotations{
			Title:           "Batch Delete Contacts",
			DestructiveHint: ptr.Bool(true),
			OpenWorldHint:   ptr.Bool(true),
		},
	}, createBatchDeleteContactsHandler(factory))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "create_contact_group",
		Icons:       serviceIcons,
		Description: "Create a new contact group (label).",
		Annotations: &mcp.ToolAnnotations{
			Title:         "Create Contact Group",
			OpenWorldHint: ptr.Bool(true),
		},
	}, createCreateContactGroupHandler(factory))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "update_contact_group",
		Icons:       serviceIcons,
		Description: "Rename an existing contact group.",
		Annotations: &mcp.ToolAnnotations{
			Title:          "Update Contact Group",
			IdempotentHint: true,
			OpenWorldHint:  ptr.Bool(true),
		},
	}, createUpdateContactGroupHandler(factory))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "delete_contact_group",
		Icons:       serviceIcons,
		Description: "Delete a contact group. Optionally also delete contacts in the group.",
		Annotations: &mcp.ToolAnnotations{
			Title:           "Delete Contact Group",
			DestructiveHint: ptr.Bool(true),
			OpenWorldHint:   ptr.Bool(true),
		},
	}, createDeleteContactGroupHandler(factory))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "modify_contact_group_members",
		Icons:       serviceIcons,
		Description: "Add or remove contacts from a contact group.",
		Annotations: &mcp.ToolAnnotations{
			Title:          "Modify Group Members",
			IdempotentHint: true,
			OpenWorldHint:  ptr.Bool(true),
		},
	}, createModifyGroupMembersHandler(factory))
}
