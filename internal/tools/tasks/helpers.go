package tasks

import (
	"google.golang.org/api/tasks/v1"
)

// TaskListSummary is a compact representation of a task list.
type TaskListSummary struct {
	ID      string `json:"id"`
	Title   string `json:"title"`
	Updated string `json:"updated,omitempty"`
}

// TaskSummary is a compact representation of a task.
type TaskSummary struct {
	ID        string `json:"id"`
	Title     string `json:"title"`
	Status    string `json:"status"`
	Due       string `json:"due,omitempty"`
	Notes     string `json:"notes,omitempty"`
	Parent    string `json:"parent,omitempty"`
	Position  string `json:"position,omitempty"`
	Completed string `json:"completed,omitempty"`
	Updated   string `json:"updated,omitempty"`
}

// taskListToSummary converts a TaskList to a summary.
func taskListToSummary(tl *tasks.TaskList) TaskListSummary {
	return TaskListSummary{
		ID:      tl.Id,
		Title:   tl.Title,
		Updated: tl.Updated,
	}
}

// taskToSummary converts a Task to a summary.
func taskToSummary(t *tasks.Task) TaskSummary {
	completed := ""
	if t.Completed != nil {
		completed = *t.Completed
	}
	return TaskSummary{
		ID:        t.Id,
		Title:     t.Title,
		Status:    t.Status,
		Due:       t.Due,
		Notes:     t.Notes,
		Parent:    t.Parent,
		Position:  t.Position,
		Completed: completed,
		Updated:   t.Updated,
	}
}

func ptrBool(b bool) *bool { return &b }
