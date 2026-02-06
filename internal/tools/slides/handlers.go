package slides

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	slidespb "google.golang.org/api/slides/v1"

	"github.com/evert/google-workspace-mcp-go/internal/middleware"
	"github.com/evert/google-workspace-mcp-go/internal/pkg/response"
	"github.com/evert/google-workspace-mcp-go/internal/services"
)

// --- create_presentation (core) ---

type CreatePresentationInput struct {
	UserEmail string `json:"user_google_email" jsonschema:"required" jsonschema_description:"The user's Google email address"`
	Title     string `json:"title" jsonschema:"required" jsonschema_description:"Title for the new presentation"`
}

func createCreatePresentationHandler(factory *services.Factory) mcp.ToolHandlerFor[CreatePresentationInput, any] {
	return func(ctx context.Context, req *mcp.CallToolRequest, input CreatePresentationInput) (*mcp.CallToolResult, any, error) {
		srv, err := factory.Slides(ctx, input.UserEmail)
		if err != nil {
			return nil, nil, middleware.HandleGoogleAPIError(err)
		}

		pres := &slidespb.Presentation{
			Title: input.Title,
		}

		created, err := srv.Presentations.Create(pres).Context(ctx).Do()
		if err != nil {
			return nil, nil, middleware.HandleGoogleAPIError(err)
		}

		rb := response.New()
		rb.Header("Presentation Created")
		rb.KeyValue("Title", created.Title)
		rb.KeyValue("Presentation ID", created.PresentationId)
		rb.KeyValue("Slides", len(created.Slides))
		rb.KeyValue("URL", fmt.Sprintf("https://docs.google.com/presentation/d/%s/edit", created.PresentationId))

		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: rb.Build()}},
		}, nil, nil
	}
}

// --- get_presentation (core) ---

type GetPresentationInput struct {
	UserEmail      string `json:"user_google_email" jsonschema:"required" jsonschema_description:"The user's Google email address"`
	PresentationID string `json:"presentation_id" jsonschema:"required" jsonschema_description:"The Google Slides presentation ID"`
}

type PresentationOutput struct {
	PresentationID string         `json:"presentation_id"`
	Title          string         `json:"title"`
	SlideCount     int            `json:"slide_count"`
	Slides         []SlideSummary `json:"slides"`
	PageWidth      float64        `json:"page_width_pt"`
	PageHeight     float64        `json:"page_height_pt"`
}

type SlideSummary struct {
	ObjectID     string `json:"object_id"`
	ElementCount int    `json:"element_count"`
}

func createGetPresentationHandler(factory *services.Factory) mcp.ToolHandlerFor[GetPresentationInput, PresentationOutput] {
	return func(ctx context.Context, req *mcp.CallToolRequest, input GetPresentationInput) (*mcp.CallToolResult, PresentationOutput, error) {
		srv, err := factory.Slides(ctx, input.UserEmail)
		if err != nil {
			return nil, PresentationOutput{}, middleware.HandleGoogleAPIError(err)
		}

		pres, err := srv.Presentations.Get(input.PresentationID).Context(ctx).Do()
		if err != nil {
			return nil, PresentationOutput{}, middleware.HandleGoogleAPIError(err)
		}

		slides := make([]SlideSummary, 0, len(pres.Slides))
		rb := response.New()
		rb.Header("Presentation Details")
		rb.KeyValue("Title", pres.Title)
		rb.KeyValue("Presentation ID", pres.PresentationId)
		rb.KeyValue("Slides", len(pres.Slides))

		if pres.PageSize != nil && pres.PageSize.Width != nil && pres.PageSize.Height != nil {
			rb.KeyValue("Page Size", fmt.Sprintf("%.0f x %.0f pt",
				pres.PageSize.Width.Magnitude, pres.PageSize.Height.Magnitude))
		}
		rb.Blank()

		for i, slide := range pres.Slides {
			ss := SlideSummary{
				ObjectID:     slide.ObjectId,
				ElementCount: len(slide.PageElements),
			}
			slides = append(slides, ss)
			rb.Item("Slide %d: %s (%d elements)", i+1, slide.ObjectId, len(slide.PageElements))
		}

		var pw, ph float64
		if pres.PageSize != nil {
			if pres.PageSize.Width != nil {
				pw = pres.PageSize.Width.Magnitude
			}
			if pres.PageSize.Height != nil {
				ph = pres.PageSize.Height.Magnitude
			}
		}

		output := PresentationOutput{
			PresentationID: pres.PresentationId,
			Title:          pres.Title,
			SlideCount:     len(pres.Slides),
			Slides:         slides,
			PageWidth:      pw,
			PageHeight:     ph,
		}

		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: rb.Build()}},
		}, output, nil
	}
}

// --- batch_update_presentation (extended) ---

type BatchUpdatePresentationInput struct {
	UserEmail      string `json:"user_google_email" jsonschema:"required" jsonschema_description:"The user's Google email address"`
	PresentationID string `json:"presentation_id" jsonschema:"required" jsonschema_description:"The Google Slides presentation ID"`
	Requests       string `json:"requests" jsonschema:"required" jsonschema_description:"JSON array of presentation update requests. Each request can contain createSlide insertText insertImage createShape deleteObject replaceAllText etc."`
}

func createBatchUpdatePresentationHandler(factory *services.Factory) mcp.ToolHandlerFor[BatchUpdatePresentationInput, any] {
	return func(ctx context.Context, req *mcp.CallToolRequest, input BatchUpdatePresentationInput) (*mcp.CallToolResult, any, error) {
		srv, err := factory.Slides(ctx, input.UserEmail)
		if err != nil {
			return nil, nil, middleware.HandleGoogleAPIError(err)
		}

		var requests []*slidespb.Request
		if err := json.Unmarshal([]byte(input.Requests), &requests); err != nil {
			return nil, nil, fmt.Errorf("invalid requests JSON - provide a JSON array of presentation update request objects: %w", err)
		}

		batchReq := &slidespb.BatchUpdatePresentationRequest{
			Requests: requests,
		}

		result, err := srv.Presentations.BatchUpdate(input.PresentationID, batchReq).Context(ctx).Do()
		if err != nil {
			return nil, nil, middleware.HandleGoogleAPIError(err)
		}

		rb := response.New()
		rb.Header("Presentation Updated")
		rb.KeyValue("Presentation ID", result.PresentationId)
		rb.KeyValue("Replies", len(result.Replies))

		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: rb.Build()}},
		}, nil, nil
	}
}

// --- get_page (extended) ---

type GetPageInput struct {
	UserEmail      string `json:"user_google_email" jsonschema:"required" jsonschema_description:"The user's Google email address"`
	PresentationID string `json:"presentation_id" jsonschema:"required" jsonschema_description:"The Google Slides presentation ID"`
	PageObjectID   string `json:"page_object_id" jsonschema:"required" jsonschema_description:"The object ID of the page/slide to retrieve"`
}

type PageOutput struct {
	ObjectID     string        `json:"object_id"`
	ElementCount int           `json:"element_count"`
	Elements     []PageElement `json:"elements"`
}

type PageElement struct {
	ObjectID    string  `json:"object_id"`
	Type        string  `json:"type"`
	Title       string  `json:"title,omitempty"`
	Description string  `json:"description,omitempty"`
	Text        string  `json:"text,omitempty"`
	Width       float64 `json:"width_pt,omitempty"`
	Height      float64 `json:"height_pt,omitempty"`
}

func createGetPageHandler(factory *services.Factory) mcp.ToolHandlerFor[GetPageInput, PageOutput] {
	return func(ctx context.Context, req *mcp.CallToolRequest, input GetPageInput) (*mcp.CallToolResult, PageOutput, error) {
		srv, err := factory.Slides(ctx, input.UserEmail)
		if err != nil {
			return nil, PageOutput{}, middleware.HandleGoogleAPIError(err)
		}

		page, err := srv.Presentations.Pages.Get(input.PresentationID, input.PageObjectID).Context(ctx).Do()
		if err != nil {
			return nil, PageOutput{}, middleware.HandleGoogleAPIError(err)
		}

		elements := make([]PageElement, 0, len(page.PageElements))
		rb := response.New()
		rb.Header("Page Details")
		rb.KeyValue("Object ID", page.ObjectId)
		rb.KeyValue("Elements", len(page.PageElements))
		rb.Blank()

		for _, el := range page.PageElements {
			pe := classifyPageElement(el)
			elements = append(elements, pe)

			rb.Item("[%s] %s", pe.Type, pe.ObjectID)
			if pe.Title != "" {
				rb.Line("    Title: %s", pe.Title)
			}
			if pe.Text != "" {
				text := pe.Text
				if len(text) > 100 {
					text = text[:100] + "..."
				}
				rb.Line("    Text: %s", text)
			}
		}

		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: rb.Build()}},
		}, PageOutput{ObjectID: page.ObjectId, ElementCount: len(page.PageElements), Elements: elements}, nil
	}
}

// --- get_page_thumbnail (extended) ---

type GetPageThumbnailInput struct {
	UserEmail      string `json:"user_google_email" jsonschema:"required" jsonschema_description:"The user's Google email address"`
	PresentationID string `json:"presentation_id" jsonschema:"required" jsonschema_description:"The Google Slides presentation ID"`
	PageObjectID   string `json:"page_object_id" jsonschema:"required" jsonschema_description:"The object ID of the page/slide"`
}

type PageThumbnailOutput struct {
	ContentURL string `json:"content_url"`
	Width      int64  `json:"width"`
	Height     int64  `json:"height"`
}

func createGetPageThumbnailHandler(factory *services.Factory) mcp.ToolHandlerFor[GetPageThumbnailInput, PageThumbnailOutput] {
	return func(ctx context.Context, req *mcp.CallToolRequest, input GetPageThumbnailInput) (*mcp.CallToolResult, PageThumbnailOutput, error) {
		srv, err := factory.Slides(ctx, input.UserEmail)
		if err != nil {
			return nil, PageThumbnailOutput{}, middleware.HandleGoogleAPIError(err)
		}

		thumbnail, err := srv.Presentations.Pages.GetThumbnail(input.PresentationID, input.PageObjectID).
			ThumbnailPropertiesMimeType("PNG").
			Context(ctx).
			Do()
		if err != nil {
			return nil, PageThumbnailOutput{}, middleware.HandleGoogleAPIError(err)
		}

		output := PageThumbnailOutput{
			ContentURL: thumbnail.ContentUrl,
			Width:      thumbnail.Width,
			Height:     thumbnail.Height,
		}

		rb := response.New()
		rb.Header("Page Thumbnail")
		rb.KeyValue("Page", input.PageObjectID)
		rb.KeyValue("Size", fmt.Sprintf("%dx%d", thumbnail.Width, thumbnail.Height))
		rb.KeyValue("URL", thumbnail.ContentUrl)

		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: rb.Build()}},
		}, output, nil
	}
}

// --- Helper functions ---

func classifyPageElement(el *slidespb.PageElement) PageElement {
	pe := PageElement{
		ObjectID:    el.ObjectId,
		Title:       el.Title,
		Description: el.Description,
	}

	if el.Size != nil {
		if el.Size.Width != nil {
			pe.Width = el.Size.Width.Magnitude
		}
		if el.Size.Height != nil {
			pe.Height = el.Size.Height.Magnitude
		}
	}

	switch {
	case el.Shape != nil:
		pe.Type = "shape"
		if el.Shape.Text != nil {
			pe.Text = extractTextFromTextElements(el.Shape.Text.TextElements)
		}
	case el.Image != nil:
		pe.Type = "image"
	case el.Table != nil:
		pe.Type = "table"
	case el.Video != nil:
		pe.Type = "video"
	case el.Line != nil:
		pe.Type = "line"
	case el.SheetsChart != nil:
		pe.Type = "sheets_chart"
	case el.WordArt != nil:
		pe.Type = "word_art"
	case el.ElementGroup != nil:
		pe.Type = "group"
	default:
		pe.Type = "unknown"
	}

	return pe
}

func extractTextFromTextElements(elements []*slidespb.TextElement) string {
	var text string
	for _, te := range elements {
		if te.TextRun != nil {
			text += te.TextRun.Content
		}
	}
	return text
}
