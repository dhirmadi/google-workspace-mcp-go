package tasks

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	taskspb "google.golang.org/api/tasks/v1"

	"github.com/evert/google-workspace-mcp-go/internal/middleware"
	"github.com/evert/google-workspace-mcp-go/internal/pkg/response"
	"github.com/evert/google-workspace-mcp-go/internal/services"
)

// --- list_task_lists (core) ---

type ListTaskListsInput struct {
	UserEmail string `json:"user_google_email" jsonschema:"required" jsonschema_description:"The user's Google email address"`
}

type ListTaskListsOutput struct {
	TaskLists []TaskListSummary `json:"task_lists"`
}

func createListTaskListsHandler(factory *services.Factory) mcp.ToolHandlerFor[ListTaskListsInput, ListTaskListsOutput] {
	return func(ctx context.Context, req *mcp.CallToolRequest, input ListTaskListsInput) (*mcp.CallToolResult, ListTaskListsOutput, error) {
		srv, err := factory.Tasks(ctx, input.UserEmail)
		if err != nil {
			return nil, ListTaskListsOutput{}, middleware.HandleGoogleAPIError(err)
		}

		result, err := srv.Tasklists.List().Context(ctx).Do()
		if err != nil {
			return nil, ListTaskListsOutput{}, middleware.HandleGoogleAPIError(err)
		}

		lists := make([]TaskListSummary, 0, len(result.Items))
		rb := response.New()
		rb.Header("Task Lists")
		rb.KeyValue("Count", len(result.Items))
		rb.Blank()

		for _, tl := range result.Items {
			lists = append(lists, taskListToSummary(tl))
			rb.Item("%s", tl.Title)
			rb.Line("    ID: %s", tl.Id)
		}

		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: rb.Build()}},
		}, ListTaskListsOutput{TaskLists: lists}, nil
	}
}

// --- list_tasks (core) ---

type ListTasksInput struct {
	UserEmail     string `json:"user_google_email" jsonschema:"required" jsonschema_description:"The user's Google email address"`
	TaskListID    string `json:"task_list_id" jsonschema:"required" jsonschema_description:"The ID of the task list"`
	MaxResults    int    `json:"max_results,omitempty" jsonschema_description:"Maximum tasks to return (default 20)"`
	ShowCompleted bool   `json:"show_completed,omitempty" jsonschema_description:"Include completed tasks (default true)"`
	ShowHidden    bool   `json:"show_hidden,omitempty" jsonschema_description:"Include hidden tasks (default false)"`
	DueMin        string `json:"due_min,omitempty" jsonschema_description:"Lower bound for due date (RFC 3339)"`
	DueMax        string `json:"due_max,omitempty" jsonschema_description:"Upper bound for due date (RFC 3339)"`
	PageToken     string `json:"page_token,omitempty" jsonschema_description:"Token for pagination"`
}

type ListTasksOutput struct {
	Tasks         []TaskSummary `json:"tasks"`
	NextPageToken string        `json:"next_page_token,omitempty"`
}

func createListTasksHandler(factory *services.Factory) mcp.ToolHandlerFor[ListTasksInput, ListTasksOutput] {
	return func(ctx context.Context, req *mcp.CallToolRequest, input ListTasksInput) (*mcp.CallToolResult, ListTasksOutput, error) {
		srv, err := factory.Tasks(ctx, input.UserEmail)
		if err != nil {
			return nil, ListTasksOutput{}, middleware.HandleGoogleAPIError(err)
		}

		if input.MaxResults == 0 {
			input.MaxResults = 20
		}

		call := srv.Tasks.List(input.TaskListID).
			MaxResults(int64(input.MaxResults)).
			ShowCompleted(true).
			Context(ctx)

		if !input.ShowCompleted {
			call = call.ShowCompleted(false)
		}
		if input.ShowHidden {
			call = call.ShowHidden(true)
		}
		if input.DueMin != "" {
			call = call.DueMin(input.DueMin)
		}
		if input.DueMax != "" {
			call = call.DueMax(input.DueMax)
		}
		if input.PageToken != "" {
			call = call.PageToken(input.PageToken)
		}

		result, err := call.Do()
		if err != nil {
			return nil, ListTasksOutput{}, middleware.HandleGoogleAPIError(err)
		}

		taskList := make([]TaskSummary, 0, len(result.Items))
		rb := response.New()
		rb.Header("Tasks")
		rb.KeyValue("Task List", input.TaskListID)
		rb.KeyValue("Count", len(result.Items))
		if result.NextPageToken != "" {
			rb.KeyValue("Next page token", result.NextPageToken)
		}
		rb.Blank()

		for _, t := range result.Items {
			ts := taskToSummary(t)
			taskList = append(taskList, ts)

			status := "○"
			if ts.Status == "completed" {
				status = "✓"
			}
			rb.Item("[%s] %s", status, ts.Title)
			if ts.Due != "" {
				rb.Line("    Due: %s", ts.Due)
			}
			if ts.Notes != "" {
				rb.Line("    Notes: %s", ts.Notes)
			}
			rb.Line("    ID: %s", ts.ID)
		}

		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: rb.Build()}},
		}, ListTasksOutput{Tasks: taskList, NextPageToken: result.NextPageToken}, nil
	}
}

// --- get_task (core) ---

type GetTaskInput struct {
	UserEmail  string `json:"user_google_email" jsonschema:"required" jsonschema_description:"The user's Google email address"`
	TaskListID string `json:"task_list_id" jsonschema:"required" jsonschema_description:"The task list ID"`
	TaskID     string `json:"task_id" jsonschema:"required" jsonschema_description:"The task ID"`
}

type GetTaskOutput struct {
	Task TaskSummary `json:"task"`
}

func createGetTaskHandler(factory *services.Factory) mcp.ToolHandlerFor[GetTaskInput, GetTaskOutput] {
	return func(ctx context.Context, req *mcp.CallToolRequest, input GetTaskInput) (*mcp.CallToolResult, GetTaskOutput, error) {
		srv, err := factory.Tasks(ctx, input.UserEmail)
		if err != nil {
			return nil, GetTaskOutput{}, middleware.HandleGoogleAPIError(err)
		}

		task, err := srv.Tasks.Get(input.TaskListID, input.TaskID).Context(ctx).Do()
		if err != nil {
			return nil, GetTaskOutput{}, middleware.HandleGoogleAPIError(err)
		}

		ts := taskToSummary(task)

		rb := response.New()
		rb.Header("Task Details")
		rb.KeyValue("Title", ts.Title)
		rb.KeyValue("Status", ts.Status)
		rb.KeyValue("ID", ts.ID)
		if ts.Due != "" {
			rb.KeyValue("Due", ts.Due)
		}
		if ts.Notes != "" {
			rb.KeyValue("Notes", ts.Notes)
		}
		if ts.Parent != "" {
			rb.KeyValue("Parent", ts.Parent)
		}
		if ts.Completed != "" {
			rb.KeyValue("Completed", ts.Completed)
		}
		rb.KeyValue("Updated", ts.Updated)

		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: rb.Build()}},
		}, GetTaskOutput{Task: ts}, nil
	}
}

// --- create_task (core) ---

type CreateTaskInput struct {
	UserEmail  string `json:"user_google_email" jsonschema:"required" jsonschema_description:"The user's Google email address"`
	TaskListID string `json:"task_list_id" jsonschema:"required" jsonschema_description:"The task list ID"`
	Title      string `json:"title" jsonschema:"required" jsonschema_description:"Task title"`
	Notes      string `json:"notes,omitempty" jsonschema_description:"Task notes/description"`
	Due        string `json:"due,omitempty" jsonschema_description:"Due date in RFC 3339 format (e.g. 2025-12-31T23:59:59Z)"`
	Parent     string `json:"parent,omitempty" jsonschema_description:"Parent task ID (for subtasks)"`
	Previous   string `json:"previous,omitempty" jsonschema_description:"Previous sibling task ID (for positioning)"`
}

func createCreateTaskHandler(factory *services.Factory) mcp.ToolHandlerFor[CreateTaskInput, any] {
	return func(ctx context.Context, req *mcp.CallToolRequest, input CreateTaskInput) (*mcp.CallToolResult, any, error) {
		srv, err := factory.Tasks(ctx, input.UserEmail)
		if err != nil {
			return nil, nil, middleware.HandleGoogleAPIError(err)
		}

		task := &taskspb.Task{
			Title: input.Title,
			Notes: input.Notes,
		}
		if input.Due != "" {
			task.Due = input.Due
		}

		call := srv.Tasks.Insert(input.TaskListID, task).Context(ctx)
		if input.Parent != "" {
			call = call.Parent(input.Parent)
		}
		if input.Previous != "" {
			call = call.Previous(input.Previous)
		}

		created, err := call.Do()
		if err != nil {
			return nil, nil, middleware.HandleGoogleAPIError(err)
		}

		rb := response.New()
		rb.Header("Task Created")
		rb.KeyValue("Title", created.Title)
		rb.KeyValue("ID", created.Id)
		if created.Due != "" {
			rb.KeyValue("Due", created.Due)
		}

		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: rb.Build()}},
		}, nil, nil
	}
}

// --- update_task (core) ---

type UpdateTaskInput struct {
	UserEmail  string `json:"user_google_email" jsonschema:"required" jsonschema_description:"The user's Google email address"`
	TaskListID string `json:"task_list_id" jsonschema:"required" jsonschema_description:"The task list ID"`
	TaskID     string `json:"task_id" jsonschema:"required" jsonschema_description:"The task ID to update"`
	Title      string `json:"title,omitempty" jsonschema_description:"New task title"`
	Notes      string `json:"notes,omitempty" jsonschema_description:"New task notes"`
	Status     string `json:"status,omitempty" jsonschema_description:"New status: needsAction or completed,enum=needsAction,enum=completed"`
	Due        string `json:"due,omitempty" jsonschema_description:"New due date (RFC 3339)"`
}

func createUpdateTaskHandler(factory *services.Factory) mcp.ToolHandlerFor[UpdateTaskInput, any] {
	return func(ctx context.Context, req *mcp.CallToolRequest, input UpdateTaskInput) (*mcp.CallToolResult, any, error) {
		srv, err := factory.Tasks(ctx, input.UserEmail)
		if err != nil {
			return nil, nil, middleware.HandleGoogleAPIError(err)
		}

		// Get existing task
		existing, err := srv.Tasks.Get(input.TaskListID, input.TaskID).Context(ctx).Do()
		if err != nil {
			return nil, nil, middleware.HandleGoogleAPIError(err)
		}

		if input.Title != "" {
			existing.Title = input.Title
		}
		if input.Notes != "" {
			existing.Notes = input.Notes
		}
		if input.Status != "" {
			existing.Status = input.Status
		}
		if input.Due != "" {
			existing.Due = input.Due
		}

		updated, err := srv.Tasks.Update(input.TaskListID, input.TaskID, existing).Context(ctx).Do()
		if err != nil {
			return nil, nil, middleware.HandleGoogleAPIError(err)
		}

		rb := response.New()
		rb.Header("Task Updated")
		rb.KeyValue("Title", updated.Title)
		rb.KeyValue("Status", updated.Status)
		rb.KeyValue("ID", updated.Id)
		if updated.Due != "" {
			rb.KeyValue("Due", updated.Due)
		}

		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: rb.Build()}},
		}, nil, nil
	}
}

// --- delete_task (extended) ---

type DeleteTaskInput struct {
	UserEmail  string `json:"user_google_email" jsonschema:"required" jsonschema_description:"The user's Google email address"`
	TaskListID string `json:"task_list_id" jsonschema:"required" jsonschema_description:"The task list ID"`
	TaskID     string `json:"task_id" jsonschema:"required" jsonschema_description:"The task ID to delete"`
}

func createDeleteTaskHandler(factory *services.Factory) mcp.ToolHandlerFor[DeleteTaskInput, any] {
	return func(ctx context.Context, req *mcp.CallToolRequest, input DeleteTaskInput) (*mcp.CallToolResult, any, error) {
		srv, err := factory.Tasks(ctx, input.UserEmail)
		if err != nil {
			return nil, nil, middleware.HandleGoogleAPIError(err)
		}

		err = srv.Tasks.Delete(input.TaskListID, input.TaskID).Context(ctx).Do()
		if err != nil {
			return nil, nil, middleware.HandleGoogleAPIError(err)
		}

		rb := response.New()
		rb.Header("Task Deleted")
		rb.KeyValue("Task ID", input.TaskID)
		rb.KeyValue("Task List", input.TaskListID)

		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: rb.Build()}},
		}, nil, nil
	}
}
