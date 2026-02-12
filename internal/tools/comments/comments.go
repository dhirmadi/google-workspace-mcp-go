// Package comments implements shared comment tools for Google Docs, Sheets, and Slides.
// All three services use the Drive API for comment management, so a single implementation
// is parameterized by resource type (document, spreadsheet, presentation).
package comments

import (
	"context"
	"fmt"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"google.golang.org/api/drive/v3"

	"github.com/evert/google-workspace-mcp-go/internal/middleware"
	"github.com/evert/google-workspace-mcp-go/internal/pkg/ptr"
	"github.com/evert/google-workspace-mcp-go/internal/pkg/response"
	"github.com/evert/google-workspace-mcp-go/internal/services"
)

// CommentSummary is a compact representation of a Drive comment.
type CommentSummary struct {
	ID        string         `json:"id"`
	Author    string         `json:"author"`
	Content   string         `json:"content"`
	CreatedAt string         `json:"created_at"`
	Resolved  bool           `json:"resolved"`
	Replies   []ReplySummary `json:"replies,omitempty"`
}

// ReplySummary is a compact representation of a comment reply.
type ReplySummary struct {
	ID        string `json:"id"`
	Author    string `json:"author"`
	Content   string `json:"content"`
	CreatedAt string `json:"created_at"`
}

// --- Input/Output types ---

type ReadCommentsInput struct {
	UserEmail string `json:"user_google_email" jsonschema:"required" jsonschema_description:"The user's Google email address"`
	FileID    string `json:"file_id" jsonschema:"required" jsonschema_description:"The Google Drive file ID"`
}

type ReadCommentsOutput struct {
	Comments []CommentSummary `json:"comments"`
}

type CreateCommentInput struct {
	UserEmail string `json:"user_google_email" jsonschema:"required" jsonschema_description:"The user's Google email address"`
	FileID    string `json:"file_id" jsonschema:"required" jsonschema_description:"The Google Drive file ID"`
	Content   string `json:"content" jsonschema:"required" jsonschema_description:"Comment text content"`
}

type ReplyToCommentInput struct {
	UserEmail string `json:"user_google_email" jsonschema:"required" jsonschema_description:"The user's Google email address"`
	FileID    string `json:"file_id" jsonschema:"required" jsonschema_description:"The Google Drive file ID"`
	CommentID string `json:"comment_id" jsonschema:"required" jsonschema_description:"The comment ID to reply to"`
	Content   string `json:"content" jsonschema:"required" jsonschema_description:"Reply text content"`
}

type ResolveCommentInput struct {
	UserEmail string `json:"user_google_email" jsonschema:"required" jsonschema_description:"The user's Google email address"`
	FileID    string `json:"file_id" jsonschema:"required" jsonschema_description:"The Google Drive file ID"`
	CommentID string `json:"comment_id" jsonschema:"required" jsonschema_description:"The comment ID to resolve"`
}

// Register registers comment tools for a specific resource type.
// resourceType: "document", "spreadsheet", or "presentation"
// icons are inherited from the parent service (Docs, Sheets, or Slides).
func Register(server *mcp.Server, factory *services.Factory, resourceType string, icons ...[]mcp.Icon) {
	prefix := resourceType

	var toolIcons []mcp.Icon
	if len(icons) > 0 {
		toolIcons = icons[0]
	}

	mcp.AddTool(server, &mcp.Tool{
		Name:        fmt.Sprintf("read_%s_comments", prefix),
		Icons:       toolIcons,
		Description: fmt.Sprintf("Read all comments from a Google %s including replies and resolution status.", capitalize(resourceType)),
		Annotations: &mcp.ToolAnnotations{
			Title:         fmt.Sprintf("Read %s Comments", capitalize(resourceType)),
			ReadOnlyHint:  true,
			OpenWorldHint: ptr.Bool(true),
		},
	}, createReadCommentsHandler(factory, resourceType))

	mcp.AddTool(server, &mcp.Tool{
		Name:        fmt.Sprintf("create_%s_comment", prefix),
		Icons:       toolIcons,
		Description: fmt.Sprintf("Add a new comment to a Google %s.", capitalize(resourceType)),
		Annotations: &mcp.ToolAnnotations{
			Title:         fmt.Sprintf("Create %s Comment", capitalize(resourceType)),
			OpenWorldHint: ptr.Bool(true),
		},
	}, createCreateCommentHandler(factory, resourceType))

	mcp.AddTool(server, &mcp.Tool{
		Name:        fmt.Sprintf("reply_to_%s_comment", prefix),
		Icons:       toolIcons,
		Description: fmt.Sprintf("Reply to an existing comment on a Google %s.", capitalize(resourceType)),
		Annotations: &mcp.ToolAnnotations{
			Title:         fmt.Sprintf("Reply to %s Comment", capitalize(resourceType)),
			OpenWorldHint: ptr.Bool(true),
		},
	}, createReplyToCommentHandler(factory, resourceType))

	mcp.AddTool(server, &mcp.Tool{
		Name:        fmt.Sprintf("resolve_%s_comment", prefix),
		Icons:       toolIcons,
		Description: fmt.Sprintf("Mark a comment on a Google %s as resolved.", capitalize(resourceType)),
		Annotations: &mcp.ToolAnnotations{
			Title:          fmt.Sprintf("Resolve %s Comment", capitalize(resourceType)),
			IdempotentHint: true,
			OpenWorldHint:  ptr.Bool(true),
		},
	}, createResolveCommentHandler(factory, resourceType))
}

// --- Handler factories ---

func createReadCommentsHandler(factory *services.Factory, _ string) mcp.ToolHandlerFor[ReadCommentsInput, ReadCommentsOutput] {
	return func(ctx context.Context, req *mcp.CallToolRequest, input ReadCommentsInput) (*mcp.CallToolResult, ReadCommentsOutput, error) {
		srv, err := factory.Drive(ctx, input.UserEmail)
		if err != nil {
			return nil, ReadCommentsOutput{}, middleware.HandleGoogleAPIError(err)
		}

		result, err := srv.Comments.List(input.FileID).
			Fields("comments(id, content, author(displayName), createdTime, resolved, replies(id, content, author(displayName), createdTime))").
			Context(ctx).
			Do()
		if err != nil {
			return nil, ReadCommentsOutput{}, middleware.HandleGoogleAPIError(err)
		}

		comments := make([]CommentSummary, 0, len(result.Comments))
		rb := response.New()
		rb.Header("Comments")
		rb.KeyValue("Count", len(result.Comments))
		rb.Blank()

		for _, c := range result.Comments {
			cs := commentToSummary(c)
			comments = append(comments, cs)

			status := "open"
			if cs.Resolved {
				status = "resolved"
			}
			rb.Item("[%s] %s — %s", status, cs.Author, cs.Content)
			rb.Line("    ID: %s | Created: %s", cs.ID, cs.CreatedAt)
			for _, r := range cs.Replies {
				rb.Line("      ↳ %s — %s", r.Author, r.Content)
			}
		}

		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: rb.Build()}},
		}, ReadCommentsOutput{Comments: comments}, nil
	}
}

func createCreateCommentHandler(factory *services.Factory, _ string) mcp.ToolHandlerFor[CreateCommentInput, any] {
	return func(ctx context.Context, req *mcp.CallToolRequest, input CreateCommentInput) (*mcp.CallToolResult, any, error) {
		srv, err := factory.Drive(ctx, input.UserEmail)
		if err != nil {
			return nil, nil, middleware.HandleGoogleAPIError(err)
		}

		comment := &drive.Comment{
			Content: input.Content,
		}

		created, err := srv.Comments.Create(input.FileID, comment).
			Fields("id, content, author(displayName), createdTime").
			Context(ctx).
			Do()
		if err != nil {
			return nil, nil, middleware.HandleGoogleAPIError(err)
		}

		rb := response.New()
		rb.Header("Comment Created")
		rb.KeyValue("Content", created.Content)
		rb.KeyValue("ID", created.Id)
		rb.KeyValue("Author", created.Author.DisplayName)

		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: rb.Build()}},
		}, nil, nil
	}
}

func createReplyToCommentHandler(factory *services.Factory, _ string) mcp.ToolHandlerFor[ReplyToCommentInput, any] {
	return func(ctx context.Context, req *mcp.CallToolRequest, input ReplyToCommentInput) (*mcp.CallToolResult, any, error) {
		srv, err := factory.Drive(ctx, input.UserEmail)
		if err != nil {
			return nil, nil, middleware.HandleGoogleAPIError(err)
		}

		reply := &drive.Reply{
			Content: input.Content,
		}

		created, err := srv.Replies.Create(input.FileID, input.CommentID, reply).
			Fields("id, content, author(displayName), createdTime").
			Context(ctx).
			Do()
		if err != nil {
			return nil, nil, middleware.HandleGoogleAPIError(err)
		}

		rb := response.New()
		rb.Header("Reply Created")
		rb.KeyValue("Content", created.Content)
		rb.KeyValue("Reply ID", created.Id)
		rb.KeyValue("Comment ID", input.CommentID)

		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: rb.Build()}},
		}, nil, nil
	}
}

func createResolveCommentHandler(factory *services.Factory, _ string) mcp.ToolHandlerFor[ResolveCommentInput, any] {
	return func(ctx context.Context, req *mcp.CallToolRequest, input ResolveCommentInput) (*mcp.CallToolResult, any, error) {
		srv, err := factory.Drive(ctx, input.UserEmail)
		if err != nil {
			return nil, nil, middleware.HandleGoogleAPIError(err)
		}

		comment := &drive.Comment{
			Resolved: true,
		}

		_, err = srv.Comments.Update(input.FileID, input.CommentID, comment).
			Fields("id, resolved").
			Context(ctx).
			Do()
		if err != nil {
			return nil, nil, middleware.HandleGoogleAPIError(err)
		}

		rb := response.New()
		rb.Header("Comment Resolved")
		rb.KeyValue("Comment ID", input.CommentID)
		rb.KeyValue("File ID", input.FileID)

		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: rb.Build()}},
		}, nil, nil
	}
}

// --- Helper functions ---

func commentToSummary(c *drive.Comment) CommentSummary {
	replies := make([]ReplySummary, 0, len(c.Replies))
	for _, r := range c.Replies {
		replies = append(replies, ReplySummary{
			ID:        r.Id,
			Author:    r.Author.DisplayName,
			Content:   r.Content,
			CreatedAt: r.CreatedTime,
		})
	}

	author := ""
	if c.Author != nil {
		author = c.Author.DisplayName
	}

	return CommentSummary{
		ID:        c.Id,
		Author:    author,
		Content:   c.Content,
		CreatedAt: c.CreatedTime,
		Resolved:  c.Resolved,
		Replies:   replies,
	}
}

func capitalize(s string) string {
	if s == "" {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

