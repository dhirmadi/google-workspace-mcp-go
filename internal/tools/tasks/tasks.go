package tasks

import (
	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/evert/google-workspace-mcp-go/internal/services"
)

var serviceIcons = []mcp.Icon{{
	Source:   "https://www.gstatic.com/images/branding/product/1x/tasks_2021_48dp.png",
	MIMEType: "image/png",
	Sizes:    []string{"48x48"},
}}

// Register registers all Tasks tools (core + extended) with the MCP server.
func Register(server *mcp.Server, factory *services.Factory) {
	// --- Core tools ---

	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_task_lists",
		Icons:       serviceIcons,
		Description: "List all task lists for the user. Returns task list IDs needed for other task operations.",
		Annotations: &mcp.ToolAnnotations{
			Title:         "List Task Lists",
			ReadOnlyHint:  true,
			OpenWorldHint: ptrBool(true),
		},
	}, createListTaskListsHandler(factory))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_tasks",
		Icons:       serviceIcons,
		Description: "List tasks in a specific task list with optional filtering by completion status, due date, and pagination.",
		Annotations: &mcp.ToolAnnotations{
			Title:         "List Tasks",
			ReadOnlyHint:  true,
			OpenWorldHint: ptrBool(true),
		},
	}, createListTasksHandler(factory))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_task",
		Icons:       serviceIcons,
		Description: "Get detailed information about a specific task including title, status, due date, and notes.",
		Annotations: &mcp.ToolAnnotations{
			Title:         "Get Task",
			ReadOnlyHint:  true,
			OpenWorldHint: ptrBool(true),
		},
	}, createGetTaskHandler(factory))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "create_task",
		Icons:       serviceIcons,
		Description: "Create a new task in a task list. Supports subtasks via parent parameter and positioning via previous parameter.",
		Annotations: &mcp.ToolAnnotations{
			Title:         "Create Task",
			OpenWorldHint: ptrBool(true),
		},
	}, createCreateTaskHandler(factory))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "update_task",
		Icons:       serviceIcons,
		Description: "Update an existing task's title, notes, status, or due date. Only specified fields are changed.",
		Annotations: &mcp.ToolAnnotations{
			Title:          "Update Task",
			IdempotentHint: true,
			OpenWorldHint:  ptrBool(true),
		},
	}, createUpdateTaskHandler(factory))

	// --- Extended tools ---

	mcp.AddTool(server, &mcp.Tool{
		Name:        "delete_task",
		Icons:       serviceIcons,
		Description: "Permanently delete a task. This action cannot be undone.",
		Annotations: &mcp.ToolAnnotations{
			Title:           "Delete Task",
			DestructiveHint: ptrBool(true),
			OpenWorldHint:   ptrBool(true),
		},
	}, createDeleteTaskHandler(factory))

	// --- Complete tools ---

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_task_list",
		Icons:       serviceIcons,
		Description: "Get details of a specific task list by ID.",
		Annotations: &mcp.ToolAnnotations{
			Title:         "Get Task List",
			ReadOnlyHint:  true,
			OpenWorldHint: ptrBool(true),
		},
	}, createGetTaskListHandler(factory))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "create_task_list",
		Icons:       serviceIcons,
		Description: "Create a new task list.",
		Annotations: &mcp.ToolAnnotations{
			Title:         "Create Task List",
			OpenWorldHint: ptrBool(true),
		},
	}, createCreateTaskListHandler(factory))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "update_task_list",
		Icons:       serviceIcons,
		Description: "Update the title of a task list.",
		Annotations: &mcp.ToolAnnotations{
			Title:          "Update Task List",
			IdempotentHint: true,
			OpenWorldHint:  ptrBool(true),
		},
	}, createUpdateTaskListHandler(factory))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "delete_task_list",
		Icons:       serviceIcons,
		Description: "Permanently delete a task list and all its tasks.",
		Annotations: &mcp.ToolAnnotations{
			Title:           "Delete Task List",
			DestructiveHint: ptrBool(true),
			OpenWorldHint:   ptrBool(true),
		},
	}, createDeleteTaskListHandler(factory))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "move_task",
		Icons:       serviceIcons,
		Description: "Move a task to a new position within its task list, or make it a subtask of another task.",
		Annotations: &mcp.ToolAnnotations{
			Title:         "Move Task",
			OpenWorldHint: ptrBool(true),
		},
	}, createMoveTaskHandler(factory))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "clear_completed_tasks",
		Icons:       serviceIcons,
		Description: "Remove all completed tasks from a task list. This cannot be undone.",
		Annotations: &mcp.ToolAnnotations{
			Title:           "Clear Completed Tasks",
			DestructiveHint: ptrBool(true),
			OpenWorldHint:   ptrBool(true),
		},
	}, createClearCompletedTasksHandler(factory))
}
