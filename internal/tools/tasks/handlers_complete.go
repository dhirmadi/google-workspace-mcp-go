package tasks

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	taskspb "google.golang.org/api/tasks/v1"

	"github.com/evert/google-workspace-mcp-go/internal/middleware"
	"github.com/evert/google-workspace-mcp-go/internal/pkg/response"
	"github.com/evert/google-workspace-mcp-go/internal/services"
)

// --- get_task_list (complete) ---

type GetTaskListInput struct {
	UserEmail  string `json:"user_google_email" jsonschema:"required" jsonschema_description:"The user's Google email address"`
	TaskListID string `json:"task_list_id" jsonschema:"required" jsonschema_description:"The task list ID"`
}

type GetTaskListOutput struct {
	TaskList TaskListSummary `json:"task_list"`
}

func createGetTaskListHandler(factory *services.Factory) mcp.ToolHandlerFor[GetTaskListInput, GetTaskListOutput] {
	return func(ctx context.Context, req *mcp.CallToolRequest, input GetTaskListInput) (*mcp.CallToolResult, GetTaskListOutput, error) {
		srv, err := factory.Tasks(ctx, input.UserEmail)
		if err != nil {
			return nil, GetTaskListOutput{}, middleware.HandleGoogleAPIError(err)
		}

		tl, err := srv.Tasklists.Get(input.TaskListID).Context(ctx).Do()
		if err != nil {
			return nil, GetTaskListOutput{}, middleware.HandleGoogleAPIError(err)
		}

		summary := taskListToSummary(tl)

		rb := response.New()
		rb.Header("Task List Details")
		rb.KeyValue("Title", tl.Title)
		rb.KeyValue("ID", tl.Id)
		rb.KeyValue("Updated", tl.Updated)

		return rb.TextResult(), GetTaskListOutput{TaskList: summary}, nil
	}
}

// --- create_task_list (complete) ---

type CreateTaskListInput struct {
	UserEmail string `json:"user_google_email" jsonschema:"required" jsonschema_description:"The user's Google email address"`
	Title     string `json:"title" jsonschema:"required" jsonschema_description:"Title for the new task list"`
}

func createCreateTaskListHandler(factory *services.Factory) mcp.ToolHandlerFor[CreateTaskListInput, any] {
	return func(ctx context.Context, req *mcp.CallToolRequest, input CreateTaskListInput) (*mcp.CallToolResult, any, error) {
		srv, err := factory.Tasks(ctx, input.UserEmail)
		if err != nil {
			return nil, nil, middleware.HandleGoogleAPIError(err)
		}

		tl := &taskspb.TaskList{
			Title: input.Title,
		}

		created, err := srv.Tasklists.Insert(tl).Context(ctx).Do()
		if err != nil {
			return nil, nil, middleware.HandleGoogleAPIError(err)
		}

		rb := response.New()
		rb.Header("Task List Created")
		rb.KeyValue("Title", created.Title)
		rb.KeyValue("ID", created.Id)

		return rb.TextResult(), nil, nil
	}
}

// --- update_task_list (complete) ---

type UpdateTaskListInput struct {
	UserEmail  string `json:"user_google_email" jsonschema:"required" jsonschema_description:"The user's Google email address"`
	TaskListID string `json:"task_list_id" jsonschema:"required" jsonschema_description:"The task list ID to update"`
	Title      string `json:"title" jsonschema:"required" jsonschema_description:"New title for the task list"`
}

func createUpdateTaskListHandler(factory *services.Factory) mcp.ToolHandlerFor[UpdateTaskListInput, any] {
	return func(ctx context.Context, req *mcp.CallToolRequest, input UpdateTaskListInput) (*mcp.CallToolResult, any, error) {
		srv, err := factory.Tasks(ctx, input.UserEmail)
		if err != nil {
			return nil, nil, middleware.HandleGoogleAPIError(err)
		}

		tl := &taskspb.TaskList{
			Title: input.Title,
		}

		updated, err := srv.Tasklists.Update(input.TaskListID, tl).Context(ctx).Do()
		if err != nil {
			return nil, nil, middleware.HandleGoogleAPIError(err)
		}

		rb := response.New()
		rb.Header("Task List Updated")
		rb.KeyValue("Title", updated.Title)
		rb.KeyValue("ID", updated.Id)

		return rb.TextResult(), nil, nil
	}
}

// --- delete_task_list (complete) ---

type DeleteTaskListInput struct {
	UserEmail  string `json:"user_google_email" jsonschema:"required" jsonschema_description:"The user's Google email address"`
	TaskListID string `json:"task_list_id" jsonschema:"required" jsonschema_description:"The task list ID to delete"`
}

func createDeleteTaskListHandler(factory *services.Factory) mcp.ToolHandlerFor[DeleteTaskListInput, any] {
	return func(ctx context.Context, req *mcp.CallToolRequest, input DeleteTaskListInput) (*mcp.CallToolResult, any, error) {
		srv, err := factory.Tasks(ctx, input.UserEmail)
		if err != nil {
			return nil, nil, middleware.HandleGoogleAPIError(err)
		}

		err = srv.Tasklists.Delete(input.TaskListID).Context(ctx).Do()
		if err != nil {
			return nil, nil, middleware.HandleGoogleAPIError(err)
		}

		rb := response.New()
		rb.Header("Task List Deleted")
		rb.KeyValue("Task List ID", input.TaskListID)

		return rb.TextResult(), nil, nil
	}
}

// --- move_task (complete) ---

type MoveTaskInput struct {
	UserEmail  string `json:"user_google_email" jsonschema:"required" jsonschema_description:"The user's Google email address"`
	TaskListID string `json:"task_list_id" jsonschema:"required" jsonschema_description:"The task list ID"`
	TaskID     string `json:"task_id" jsonschema:"required" jsonschema_description:"The task ID to move"`
	Parent     string `json:"parent,omitempty" jsonschema_description:"New parent task ID (makes this a subtask). Empty to make top-level."`
	Previous   string `json:"previous,omitempty" jsonschema_description:"Task ID of the previous sibling (for ordering)"`
}

func createMoveTaskHandler(factory *services.Factory) mcp.ToolHandlerFor[MoveTaskInput, any] {
	return func(ctx context.Context, req *mcp.CallToolRequest, input MoveTaskInput) (*mcp.CallToolResult, any, error) {
		srv, err := factory.Tasks(ctx, input.UserEmail)
		if err != nil {
			return nil, nil, middleware.HandleGoogleAPIError(err)
		}

		call := srv.Tasks.Move(input.TaskListID, input.TaskID).Context(ctx)
		if input.Parent != "" {
			call = call.Parent(input.Parent)
		}
		if input.Previous != "" {
			call = call.Previous(input.Previous)
		}

		moved, err := call.Do()
		if err != nil {
			return nil, nil, middleware.HandleGoogleAPIError(err)
		}

		rb := response.New()
		rb.Header("Task Moved")
		rb.KeyValue("Title", moved.Title)
		rb.KeyValue("ID", moved.Id)
		rb.KeyValue("Position", moved.Position)
		if moved.Parent != "" {
			rb.KeyValue("Parent", moved.Parent)
		}

		return rb.TextResult(), nil, nil
	}
}

// --- clear_completed_tasks (complete) ---

type ClearCompletedTasksInput struct {
	UserEmail  string `json:"user_google_email" jsonschema:"required" jsonschema_description:"The user's Google email address"`
	TaskListID string `json:"task_list_id" jsonschema:"required" jsonschema_description:"The task list ID to clear completed tasks from"`
}

func createClearCompletedTasksHandler(factory *services.Factory) mcp.ToolHandlerFor[ClearCompletedTasksInput, any] {
	return func(ctx context.Context, req *mcp.CallToolRequest, input ClearCompletedTasksInput) (*mcp.CallToolResult, any, error) {
		srv, err := factory.Tasks(ctx, input.UserEmail)
		if err != nil {
			return nil, nil, middleware.HandleGoogleAPIError(err)
		}

		err = srv.Tasks.Clear(input.TaskListID).Context(ctx).Do()
		if err != nil {
			return nil, nil, middleware.HandleGoogleAPIError(err)
		}

		rb := response.New()
		rb.Header("Completed Tasks Cleared")
		rb.KeyValue("Task List ID", input.TaskListID)
		rb.Line("All completed tasks have been removed from the task list.")

		return rb.TextResult(), nil, nil
	}
}
