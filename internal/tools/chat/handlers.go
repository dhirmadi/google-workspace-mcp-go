package chat

import (
	"context"
	"fmt"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	chatpb "google.golang.org/api/chat/v1"

	"github.com/evert/google-workspace-mcp-go/internal/middleware"
	"github.com/evert/google-workspace-mcp-go/internal/pkg/response"
	"github.com/evert/google-workspace-mcp-go/internal/services"
)

// --- SpaceSummary / MessageSummary ---

type SpaceSummary struct {
	Name        string `json:"name"`
	DisplayName string `json:"display_name"`
	Type        string `json:"type"`
	Threaded    bool   `json:"threaded,omitempty"`
}

type ChatMessageSummary struct {
	Name       string `json:"name"`
	Sender     string `json:"sender"`
	Text       string `json:"text"`
	CreateTime string `json:"create_time"`
	ThreadName string `json:"thread_name,omitempty"`
}

// --- list_chat_spaces (core) ---

type ListChatSpacesInput struct {
	UserEmail string `json:"user_google_email" jsonschema:"required" jsonschema_description:"The user's Google email address"`
	PageSize  int    `json:"page_size,omitempty" jsonschema_description:"Maximum spaces to return (default 25)"`
	PageToken string `json:"page_token,omitempty" jsonschema_description:"Token for pagination"`
}

type ListChatSpacesOutput struct {
	Spaces        []SpaceSummary `json:"spaces"`
	NextPageToken string         `json:"next_page_token,omitempty"`
}

func createListChatSpacesHandler(factory *services.Factory) mcp.ToolHandlerFor[ListChatSpacesInput, ListChatSpacesOutput] {
	return func(ctx context.Context, req *mcp.CallToolRequest, input ListChatSpacesInput) (*mcp.CallToolResult, ListChatSpacesOutput, error) {
		if input.PageSize == 0 {
			input.PageSize = 25
		}

		srv, err := factory.Chat(ctx, input.UserEmail)
		if err != nil {
			return nil, ListChatSpacesOutput{}, middleware.HandleGoogleAPIError(err)
		}

		call := srv.Spaces.List().
			PageSize(int64(input.PageSize)).
			Context(ctx)

		if input.PageToken != "" {
			call = call.PageToken(input.PageToken)
		}

		result, err := call.Do()
		if err != nil {
			return nil, ListChatSpacesOutput{}, middleware.HandleGoogleAPIError(err)
		}

		spaces := make([]SpaceSummary, 0, len(result.Spaces))
		rb := response.New()
		rb.Header("Chat Spaces")
		rb.KeyValue("Count", len(result.Spaces))
		if result.NextPageToken != "" {
			rb.KeyValue("Next page token", result.NextPageToken)
		}
		rb.Blank()

		for _, s := range result.Spaces {
			ss := SpaceSummary{
				Name:        s.Name,
				DisplayName: s.DisplayName,
				Type:        s.Type,
				Threaded:    s.Threaded,
			}
			spaces = append(spaces, ss)
			rb.Item("%s (%s)", ss.DisplayName, ss.Type)
			rb.Line("    Name: %s", ss.Name)
		}

		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: rb.Build()}},
		}, ListChatSpacesOutput{Spaces: spaces, NextPageToken: result.NextPageToken}, nil
	}
}

// --- get_chat_messages (core) ---

type GetChatMessagesInput struct {
	UserEmail string `json:"user_google_email" jsonschema:"required" jsonschema_description:"The user's Google email address"`
	SpaceName string `json:"space_name" jsonschema:"required" jsonschema_description:"The space resource name (e.g. spaces/AAAAAA)"`
	PageSize  int    `json:"page_size,omitempty" jsonschema_description:"Maximum messages to return (default 25)"`
	PageToken string `json:"page_token,omitempty" jsonschema_description:"Token for pagination"`
}

type GetChatMessagesOutput struct {
	Messages      []ChatMessageSummary `json:"messages"`
	NextPageToken string               `json:"next_page_token,omitempty"`
}

func createGetChatMessagesHandler(factory *services.Factory) mcp.ToolHandlerFor[GetChatMessagesInput, GetChatMessagesOutput] {
	return func(ctx context.Context, req *mcp.CallToolRequest, input GetChatMessagesInput) (*mcp.CallToolResult, GetChatMessagesOutput, error) {
		if input.PageSize == 0 {
			input.PageSize = 25
		}

		srv, err := factory.Chat(ctx, input.UserEmail)
		if err != nil {
			return nil, GetChatMessagesOutput{}, middleware.HandleGoogleAPIError(err)
		}

		call := srv.Spaces.Messages.List(input.SpaceName).
			PageSize(int64(input.PageSize)).
			Context(ctx)

		if input.PageToken != "" {
			call = call.PageToken(input.PageToken)
		}

		result, err := call.Do()
		if err != nil {
			return nil, GetChatMessagesOutput{}, middleware.HandleGoogleAPIError(err)
		}

		messages := make([]ChatMessageSummary, 0, len(result.Messages))
		rb := response.New()
		rb.Header("Chat Messages")
		rb.KeyValue("Space", input.SpaceName)
		rb.KeyValue("Count", len(result.Messages))
		if result.NextPageToken != "" {
			rb.KeyValue("Next page token", result.NextPageToken)
		}
		rb.Blank()

		for _, m := range result.Messages {
			sender := ""
			if m.Sender != nil {
				sender = m.Sender.DisplayName
				if sender == "" {
					sender = m.Sender.Name
				}
			}

			threadName := ""
			if m.Thread != nil {
				threadName = m.Thread.Name
			}

			ms := ChatMessageSummary{
				Name:       m.Name,
				Sender:     sender,
				Text:       m.Text,
				CreateTime: m.CreateTime,
				ThreadName: threadName,
			}
			messages = append(messages, ms)

			rb.Item("%s: %s", ms.Sender, truncateText(ms.Text, 100))
			rb.Line("    Time: %s | Name: %s", ms.CreateTime, ms.Name)
		}

		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: rb.Build()}},
		}, GetChatMessagesOutput{Messages: messages, NextPageToken: result.NextPageToken}, nil
	}
}

// --- search_chat_messages (core) ---

type SearchChatMessagesInput struct {
	UserEmail string `json:"user_google_email" jsonschema:"required" jsonschema_description:"The user's Google email address"`
	Query     string `json:"query" jsonschema:"required" jsonschema_description:"Search query for chat messages"`
	PageSize  int    `json:"page_size,omitempty" jsonschema_description:"Maximum results (default 25)"`
	PageToken string `json:"page_token,omitempty" jsonschema_description:"Token for pagination"`
}

type SearchChatMessagesOutput struct {
	Messages      []ChatMessageSummary `json:"messages"`
	NextPageToken string               `json:"next_page_token,omitempty"`
}

func createSearchChatMessagesHandler(factory *services.Factory) mcp.ToolHandlerFor[SearchChatMessagesInput, SearchChatMessagesOutput] {
	return func(ctx context.Context, req *mcp.CallToolRequest, input SearchChatMessagesInput) (*mcp.CallToolResult, SearchChatMessagesOutput, error) {
		if input.PageSize == 0 {
			input.PageSize = 25
		}

		srv, err := factory.Chat(ctx, input.UserEmail)
		if err != nil {
			return nil, SearchChatMessagesOutput{}, middleware.HandleGoogleAPIError(err)
		}

		call := srv.Spaces.Messages.List("spaces/-").
			Filter(fmt.Sprintf("text:%q", input.Query)).
			PageSize(int64(input.PageSize)).
			Context(ctx)

		if input.PageToken != "" {
			call = call.PageToken(input.PageToken)
		}

		result, err := call.Do()
		if err != nil {
			return nil, SearchChatMessagesOutput{}, middleware.HandleGoogleAPIError(err)
		}

		messages := make([]ChatMessageSummary, 0, len(result.Messages))
		rb := response.New()
		rb.Header("Chat Search Results")
		rb.KeyValue("Query", input.Query)
		rb.KeyValue("Results", len(result.Messages))
		if result.NextPageToken != "" {
			rb.KeyValue("Next page token", result.NextPageToken)
		}
		rb.Blank()

		for _, m := range result.Messages {
			sender := ""
			if m.Sender != nil {
				sender = m.Sender.DisplayName
				if sender == "" {
					sender = m.Sender.Name
				}
			}
			ms := ChatMessageSummary{
				Name:       m.Name,
				Sender:     sender,
				Text:       m.Text,
				CreateTime: m.CreateTime,
			}
			messages = append(messages, ms)
			rb.Item("%s: %s", ms.Sender, truncateText(ms.Text, 100))
			rb.Line("    Time: %s", ms.CreateTime)
		}

		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: rb.Build()}},
		}, SearchChatMessagesOutput{Messages: messages, NextPageToken: result.NextPageToken}, nil
	}
}

// --- send_chat_message (core) ---

type SendChatMessageInput struct {
	UserEmail  string `json:"user_google_email" jsonschema:"required" jsonschema_description:"The user's Google email address"`
	SpaceName  string `json:"space_name" jsonschema:"required" jsonschema_description:"The space resource name (e.g. spaces/AAAAAA)"`
	Text       string `json:"text" jsonschema:"required" jsonschema_description:"Message text to send"`
	ThreadName string `json:"thread_name,omitempty" jsonschema_description:"Thread name to reply in (creates new thread if omitted)"`
}

func createSendChatMessageHandler(factory *services.Factory) mcp.ToolHandlerFor[SendChatMessageInput, any] {
	return func(ctx context.Context, req *mcp.CallToolRequest, input SendChatMessageInput) (*mcp.CallToolResult, any, error) {
		srv, err := factory.Chat(ctx, input.UserEmail)
		if err != nil {
			return nil, nil, middleware.HandleGoogleAPIError(err)
		}

		msg := &chatpb.Message{
			Text: input.Text,
		}

		if input.ThreadName != "" {
			msg.Thread = &chatpb.Thread{
				Name: input.ThreadName,
			}
		}

		sent, err := srv.Spaces.Messages.Create(input.SpaceName, msg).Context(ctx).Do()
		if err != nil {
			return nil, nil, middleware.HandleGoogleAPIError(err)
		}

		rb := response.New()
		rb.Header("Chat Message Sent")
		rb.KeyValue("Space", input.SpaceName)
		rb.KeyValue("Message Name", sent.Name)
		if sent.Thread != nil {
			rb.KeyValue("Thread", sent.Thread.Name)
		}

		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: rb.Build()}},
		}, nil, nil
	}
}

// --- helper functions ---

func truncateText(text string, maxLen int) string {
	text = strings.ReplaceAll(text, "\n", " ")
	if len(text) <= maxLen {
		return text
	}
	return text[:maxLen] + "..."
}

func ptrBool(b bool) *bool { return &b }
