package forms

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	formspb "google.golang.org/api/forms/v1"

	"github.com/evert/google-workspace-mcp-go/internal/middleware"
	"github.com/evert/google-workspace-mcp-go/internal/pkg/response"
	"github.com/evert/google-workspace-mcp-go/internal/services"
)

// --- create_form (core) ---

type CreateFormInput struct {
	UserEmail   string `json:"user_google_email" jsonschema:"required" jsonschema_description:"The user's Google email address"`
	Title       string `json:"title" jsonschema:"required" jsonschema_description:"Title for the new form"`
	Description string `json:"description,omitempty" jsonschema_description:"Form description shown to respondents"`
}

func createCreateFormHandler(factory *services.Factory) mcp.ToolHandlerFor[CreateFormInput, any] {
	return func(ctx context.Context, req *mcp.CallToolRequest, input CreateFormInput) (*mcp.CallToolResult, any, error) {
		srv, err := factory.Forms(ctx, input.UserEmail)
		if err != nil {
			return nil, nil, middleware.HandleGoogleAPIError(err)
		}

		form := &formspb.Form{
			Info: &formspb.Info{
				Title:       input.Title,
				Description: input.Description,
			},
		}

		created, err := srv.Forms.Create(form).Context(ctx).Do()
		if err != nil {
			return nil, nil, middleware.HandleGoogleAPIError(err)
		}

		rb := response.New()
		rb.Header("Form Created")
		rb.KeyValue("Title", created.Info.Title)
		rb.KeyValue("Form ID", created.FormId)
		rb.KeyValue("Responder URI", created.ResponderUri)
		rb.KeyValue("Edit URL", fmt.Sprintf("https://docs.google.com/forms/d/%s/edit", created.FormId))

		return rb.TextResult(), nil, nil
	}
}

// --- get_form (core) ---

type GetFormInput struct {
	UserEmail string `json:"user_google_email" jsonschema:"required" jsonschema_description:"The user's Google email address"`
	FormID    string `json:"form_id" jsonschema:"required" jsonschema_description:"The Google Form ID"`
}

type FormOutput struct {
	FormID       string     `json:"form_id"`
	Title        string     `json:"title"`
	Description  string     `json:"description"`
	ResponderURI string     `json:"responder_uri"`
	Items        []FormItem `json:"items"`
}

type FormItem struct {
	ItemID      string `json:"item_id"`
	Title       string `json:"title"`
	Description string `json:"description,omitempty"`
	Type        string `json:"type"`
}

func createGetFormHandler(factory *services.Factory) mcp.ToolHandlerFor[GetFormInput, FormOutput] {
	return func(ctx context.Context, req *mcp.CallToolRequest, input GetFormInput) (*mcp.CallToolResult, FormOutput, error) {
		srv, err := factory.Forms(ctx, input.UserEmail)
		if err != nil {
			return nil, FormOutput{}, middleware.HandleGoogleAPIError(err)
		}

		form, err := srv.Forms.Get(input.FormID).Context(ctx).Do()
		if err != nil {
			return nil, FormOutput{}, middleware.HandleGoogleAPIError(err)
		}

		items := make([]FormItem, 0, len(form.Items))
		rb := response.New()
		rb.Header("Form Details")
		rb.KeyValue("Title", form.Info.Title)
		if form.Info.Description != "" {
			rb.KeyValue("Description", form.Info.Description)
		}
		rb.KeyValue("Form ID", form.FormId)
		rb.KeyValue("Responder URI", form.ResponderUri)
		rb.Blank()
		rb.Line("Items (%d):", len(form.Items))

		for _, item := range form.Items {
			fi := FormItem{
				ItemID: item.ItemId,
				Title:  item.Title,
			}
			if item.Description != "" {
				fi.Description = item.Description
			}
			fi.Type = classifyFormItem(item)
			items = append(items, fi)

			rb.Item("[%s] %s", fi.Type, fi.Title)
			rb.Line("    ID: %s", fi.ItemID)
			if fi.Description != "" {
				rb.Line("    Description: %s", fi.Description)
			}
		}

		output := FormOutput{
			FormID:       form.FormId,
			Title:        form.Info.Title,
			Description:  form.Info.Description,
			ResponderURI: form.ResponderUri,
			Items:        items,
		}

		return rb.TextResult(), output, nil
	}
}

// --- list_form_responses (extended) ---

type ListFormResponsesInput struct {
	UserEmail string `json:"user_google_email" jsonschema:"required" jsonschema_description:"The user's Google email address"`
	FormID    string `json:"form_id" jsonschema:"required" jsonschema_description:"The Google Form ID"`
	PageSize  int    `json:"page_size,omitempty" jsonschema_description:"Max responses to return (default 25)"`
	PageToken string `json:"page_token,omitempty" jsonschema_description:"Token for next page of results"`
}

type ListFormResponsesOutput struct {
	Responses     []FormResponseSummary `json:"responses"`
	NextPageToken string                `json:"next_page_token,omitempty"`
}

type FormResponseSummary struct {
	ResponseID      string            `json:"response_id"`
	CreateTime      string            `json:"create_time"`
	LastSubmitTime  string            `json:"last_submit_time"`
	RespondentEmail string            `json:"respondent_email,omitempty"`
	Answers         map[string]string `json:"answers"`
}

func createListFormResponsesHandler(factory *services.Factory) mcp.ToolHandlerFor[ListFormResponsesInput, ListFormResponsesOutput] {
	return func(ctx context.Context, req *mcp.CallToolRequest, input ListFormResponsesInput) (*mcp.CallToolResult, ListFormResponsesOutput, error) {
		srv, err := factory.Forms(ctx, input.UserEmail)
		if err != nil {
			return nil, ListFormResponsesOutput{}, middleware.HandleGoogleAPIError(err)
		}

		if input.PageSize == 0 {
			input.PageSize = 25
		}

		call := srv.Forms.Responses.List(input.FormID).
			PageSize(int64(input.PageSize)).
			Context(ctx)
		if input.PageToken != "" {
			call = call.PageToken(input.PageToken)
		}

		result, err := call.Do()
		if err != nil {
			return nil, ListFormResponsesOutput{}, middleware.HandleGoogleAPIError(err)
		}

		responses := make([]FormResponseSummary, 0, len(result.Responses))
		rb := response.New()
		rb.Header("Form Responses")
		rb.KeyValue("Form ID", input.FormID)
		rb.KeyValue("Count", len(result.Responses))
		if result.NextPageToken != "" {
			rb.KeyValue("Next page token", result.NextPageToken)
		}
		rb.Blank()

		for _, r := range result.Responses {
			frs := FormResponseSummary{
				ResponseID:     r.ResponseId,
				CreateTime:     r.CreateTime,
				LastSubmitTime: r.LastSubmittedTime,
				Answers:        make(map[string]string),
			}
			if r.RespondentEmail != "" {
				frs.RespondentEmail = r.RespondentEmail
			}
			for qID, ans := range r.Answers {
				frs.Answers[qID] = formatAnswer(ans)
			}
			responses = append(responses, frs)

			rb.Item("Response: %s", frs.ResponseID)
			rb.Line("    Submitted: %s", frs.LastSubmitTime)
			if frs.RespondentEmail != "" {
				rb.Line("    Respondent: %s", frs.RespondentEmail)
			}
			rb.Line("    Answers: %d", len(frs.Answers))
		}

		return rb.TextResult(), ListFormResponsesOutput{Responses: responses, NextPageToken: result.NextPageToken}, nil
	}
}

// --- set_publish_settings (complete) ---

type SetPublishSettingsInput struct {
	UserEmail          string `json:"user_google_email" jsonschema:"required" jsonschema_description:"The user's Google email address"`
	FormID             string `json:"form_id" jsonschema:"required" jsonschema_description:"The Google Form ID"`
	AcceptingResponses bool   `json:"accepting_responses" jsonschema_description:"Whether the form is accepting responses (default true)"`
	IsQuiz             bool   `json:"is_quiz,omitempty" jsonschema_description:"Whether the form is a quiz"`
}

func createSetPublishSettingsHandler(factory *services.Factory) mcp.ToolHandlerFor[SetPublishSettingsInput, any] {
	return func(ctx context.Context, req *mcp.CallToolRequest, input SetPublishSettingsInput) (*mcp.CallToolResult, any, error) {
		srv, err := factory.Forms(ctx, input.UserEmail)
		if err != nil {
			return nil, nil, middleware.HandleGoogleAPIError(err)
		}

		batchReq := &formspb.BatchUpdateFormRequest{
			Requests: []*formspb.Request{
				{
					UpdateSettings: &formspb.UpdateSettingsRequest{
						Settings: &formspb.FormSettings{
							QuizSettings: &formspb.QuizSettings{
								IsQuiz: input.IsQuiz,
							},
						},
						UpdateMask: "quizSettings.isQuiz",
					},
				},
			},
		}

		_, err = srv.Forms.BatchUpdate(input.FormID, batchReq).Context(ctx).Do()
		if err != nil {
			return nil, nil, middleware.HandleGoogleAPIError(err)
		}

		rb := response.New()
		rb.Header("Form Settings Updated")
		rb.KeyValue("Form ID", input.FormID)
		rb.KeyValue("Is Quiz", input.IsQuiz)

		return rb.TextResult(), nil, nil
	}
}

// --- get_form_response (complete) ---

type GetFormResponseInput struct {
	UserEmail  string `json:"user_google_email" jsonschema:"required" jsonschema_description:"The user's Google email address"`
	FormID     string `json:"form_id" jsonschema:"required" jsonschema_description:"The Google Form ID"`
	ResponseID string `json:"response_id" jsonschema:"required" jsonschema_description:"The response ID to retrieve"`
}

type GetFormResponseOutput struct {
	ResponseID      string            `json:"response_id"`
	CreateTime      string            `json:"create_time"`
	LastSubmitTime  string            `json:"last_submit_time"`
	RespondentEmail string            `json:"respondent_email,omitempty"`
	Answers         map[string]string `json:"answers"`
}

func createGetFormResponseHandler(factory *services.Factory) mcp.ToolHandlerFor[GetFormResponseInput, GetFormResponseOutput] {
	return func(ctx context.Context, req *mcp.CallToolRequest, input GetFormResponseInput) (*mcp.CallToolResult, GetFormResponseOutput, error) {
		srv, err := factory.Forms(ctx, input.UserEmail)
		if err != nil {
			return nil, GetFormResponseOutput{}, middleware.HandleGoogleAPIError(err)
		}

		r, err := srv.Forms.Responses.Get(input.FormID, input.ResponseID).Context(ctx).Do()
		if err != nil {
			return nil, GetFormResponseOutput{}, middleware.HandleGoogleAPIError(err)
		}

		output := GetFormResponseOutput{
			ResponseID:     r.ResponseId,
			CreateTime:     r.CreateTime,
			LastSubmitTime: r.LastSubmittedTime,
			Answers:        make(map[string]string),
		}
		if r.RespondentEmail != "" {
			output.RespondentEmail = r.RespondentEmail
		}

		rb := response.New()
		rb.Header("Form Response")
		rb.KeyValue("Response ID", r.ResponseId)
		rb.KeyValue("Created", r.CreateTime)
		rb.KeyValue("Last Submitted", r.LastSubmittedTime)
		if r.RespondentEmail != "" {
			rb.KeyValue("Respondent", r.RespondentEmail)
		}
		rb.Blank()

		for qID, ans := range r.Answers {
			val := formatAnswer(ans)
			output.Answers[qID] = val
			rb.Item("Q[%s]: %s", qID, val)
		}

		return rb.TextResult(), output, nil
	}
}

// --- batch_update_form (complete) ---

type BatchUpdateFormInput struct {
	UserEmail string `json:"user_google_email" jsonschema:"required" jsonschema_description:"The user's Google email address"`
	FormID    string `json:"form_id" jsonschema:"required" jsonschema_description:"The Google Form ID"`
	Requests  string `json:"requests" jsonschema:"required" jsonschema_description:"JSON array of form update requests. Each request can contain createItem updateItem deleteItem updateFormInfo or updateSettings."`
}

func createBatchUpdateFormHandler(factory *services.Factory) mcp.ToolHandlerFor[BatchUpdateFormInput, any] {
	return func(ctx context.Context, req *mcp.CallToolRequest, input BatchUpdateFormInput) (*mcp.CallToolResult, any, error) {
		srv, err := factory.Forms(ctx, input.UserEmail)
		if err != nil {
			return nil, nil, middleware.HandleGoogleAPIError(err)
		}

		var requests []*formspb.Request
		if err := json.Unmarshal([]byte(input.Requests), &requests); err != nil {
			return nil, nil, fmt.Errorf("invalid requests JSON - provide a JSON array of form update request objects: %w", err)
		}

		batchReq := &formspb.BatchUpdateFormRequest{
			Requests: requests,
		}

		result, err := srv.Forms.BatchUpdate(input.FormID, batchReq).Context(ctx).Do()
		if err != nil {
			return nil, nil, middleware.HandleGoogleAPIError(err)
		}

		rb := response.New()
		rb.Header("Form Updated")
		rb.KeyValue("Form ID", input.FormID)
		rb.KeyValue("Updates Applied", len(result.Replies))
		if result.Form != nil && result.Form.Info != nil {
			rb.KeyValue("Title", result.Form.Info.Title)
		}

		return rb.TextResult(), nil, nil
	}
}

// --- Helper functions ---

func classifyFormItem(item *formspb.Item) string {
	if item.QuestionItem != nil {
		q := item.QuestionItem.Question
		if q == nil {
			return "question"
		}
		switch {
		case q.ChoiceQuestion != nil:
			return "choice"
		case q.TextQuestion != nil:
			return "text"
		case q.ScaleQuestion != nil:
			return "scale"
		case q.DateQuestion != nil:
			return "date"
		case q.TimeQuestion != nil:
			return "time"
		case q.FileUploadQuestion != nil:
			return "file_upload"
		default:
			return "question"
		}
	}
	if item.QuestionGroupItem != nil {
		return "question_group"
	}
	if item.PageBreakItem != nil {
		return "page_break"
	}
	if item.TextItem != nil {
		return "text_section"
	}
	if item.ImageItem != nil {
		return "image"
	}
	if item.VideoItem != nil {
		return "video"
	}
	return "unknown"
}

func formatAnswer(ans formspb.Answer) string {
	if ans.TextAnswers == nil {
		return ""
	}
	var parts []string
	for _, a := range ans.TextAnswers.Answers {
		parts = append(parts, a.Value)
	}
	if len(parts) == 1 {
		return parts[0]
	}
	result := ""
	for i, p := range parts {
		if i > 0 {
			result += ", "
		}
		result += p
	}
	return result
}
