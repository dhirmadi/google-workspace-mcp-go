package docs

import (
	"context"
	"fmt"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	docspb "google.golang.org/api/docs/v1"

	"github.com/evert/google-workspace-mcp-go/internal/middleware"
	"github.com/evert/google-workspace-mcp-go/internal/pkg/response"
	"github.com/evert/google-workspace-mcp-go/internal/services"
)

// --- get_doc_content (core) ---

type GetDocContentInput struct {
	UserEmail  string `json:"user_google_email" jsonschema:"required" jsonschema_description:"The user's Google email address"`
	DocumentID string `json:"document_id" jsonschema:"required" jsonschema_description:"The Google Docs document ID"`
}

func createGetDocContentHandler(factory *services.Factory) mcp.ToolHandlerFor[GetDocContentInput, DocContentOutput] {
	return func(ctx context.Context, req *mcp.CallToolRequest, input GetDocContentInput) (*mcp.CallToolResult, DocContentOutput, error) {
		srv, err := factory.Docs(ctx, input.UserEmail)
		if err != nil {
			return nil, DocContentOutput{}, middleware.HandleGoogleAPIError(err)
		}

		doc, err := srv.Documents.Get(input.DocumentID).Context(ctx).Do()
		if err != nil {
			return nil, DocContentOutput{}, middleware.HandleGoogleAPIError(err)
		}

		content := extractDocText(doc)

		rb := response.New()
		rb.Header("Document Content")
		rb.KeyValue("Title", doc.Title)
		rb.KeyValue("Document ID", doc.DocumentId)
		rb.Blank()
		rb.Raw(content)

		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: rb.Build()}},
		}, DocContentOutput{DocumentID: doc.DocumentId, Title: doc.Title, Content: content}, nil
	}
}

// --- create_doc (core) ---

type CreateDocInput struct {
	UserEmail string `json:"user_google_email" jsonschema:"required" jsonschema_description:"The user's Google email address"`
	Title     string `json:"title" jsonschema:"required" jsonschema_description:"Title for the new document"`
	Content   string `json:"content,omitempty" jsonschema_description:"Initial text content to insert"`
}

func createCreateDocHandler(factory *services.Factory) mcp.ToolHandlerFor[CreateDocInput, any] {
	return func(ctx context.Context, req *mcp.CallToolRequest, input CreateDocInput) (*mcp.CallToolResult, any, error) {
		srv, err := factory.Docs(ctx, input.UserEmail)
		if err != nil {
			return nil, nil, middleware.HandleGoogleAPIError(err)
		}

		doc := &docspb.Document{
			Title: input.Title,
		}

		created, err := srv.Documents.Create(doc).Context(ctx).Do()
		if err != nil {
			return nil, nil, middleware.HandleGoogleAPIError(err)
		}

		// If initial content was provided, insert it
		if input.Content != "" {
			insertReq := &docspb.BatchUpdateDocumentRequest{
				Requests: []*docspb.Request{
					{
						InsertText: &docspb.InsertTextRequest{
							Text: input.Content,
							Location: &docspb.Location{
								Index: 1,
							},
						},
					},
				},
			}
			_, err = srv.Documents.BatchUpdate(created.DocumentId, insertReq).Context(ctx).Do()
			if err != nil {
				return nil, nil, middleware.HandleGoogleAPIError(err)
			}
		}

		rb := response.New()
		rb.Header("Document Created")
		rb.KeyValue("Title", created.Title)
		rb.KeyValue("Document ID", created.DocumentId)
		rb.KeyValue("Link", fmt.Sprintf("https://docs.google.com/document/d/%s/edit", created.DocumentId))

		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: rb.Build()}},
		}, nil, nil
	}
}

// --- modify_doc_text (core) ---

type ModifyDocTextInput struct {
	UserEmail       string `json:"user_google_email" jsonschema:"required" jsonschema_description:"The user's Google email address"`
	DocumentID      string `json:"document_id" jsonschema:"required" jsonschema_description:"The document ID to update"`
	StartIndex      int64  `json:"start_index" jsonschema:"required" jsonschema_description:"Start position for operation (1-based)"`
	EndIndex        *int64 `json:"end_index,omitempty" jsonschema_description:"End position for text replacement/formatting. If not provided with text the text is inserted."`
	Text            string `json:"text,omitempty" jsonschema_description:"New text to insert or replace with"`
	Bold            *bool  `json:"bold,omitempty" jsonschema_description:"Make text bold (true/false)"`
	Italic          *bool  `json:"italic,omitempty" jsonschema_description:"Make text italic (true/false)"`
	Underline       *bool  `json:"underline,omitempty" jsonschema_description:"Underline text (true/false)"`
	FontSize        *int   `json:"font_size,omitempty" jsonschema_description:"Font size in points"`
	FontFamily      string `json:"font_family,omitempty" jsonschema_description:"Font family name (e.g. Arial)"`
	TextColor       string `json:"text_color,omitempty" jsonschema_description:"Text color as hex (#RRGGBB)"`
	BackgroundColor string `json:"background_color,omitempty" jsonschema_description:"Background/highlight color as hex (#RRGGBB)"`
}

func createModifyDocTextHandler(factory *services.Factory) mcp.ToolHandlerFor[ModifyDocTextInput, any] {
	return func(ctx context.Context, req *mcp.CallToolRequest, input ModifyDocTextInput) (*mcp.CallToolResult, any, error) {
		srv, err := factory.Docs(ctx, input.UserEmail)
		if err != nil {
			return nil, nil, middleware.HandleGoogleAPIError(err)
		}

		requests := make([]*docspb.Request, 0, 2)
		operations := make([]string, 0, 2)

		// Text modification (insert or replace)
		if input.Text != "" {
			if input.EndIndex != nil {
				// Replace: delete then insert
				requests = append(requests,
					&docspb.Request{
						DeleteContentRange: &docspb.DeleteContentRangeRequest{
							Range: &docspb.Range{
								StartIndex: input.StartIndex,
								EndIndex:   *input.EndIndex,
							},
						},
					},
					&docspb.Request{
						InsertText: &docspb.InsertTextRequest{
							Text: input.Text,
							Location: &docspb.Location{
								Index: input.StartIndex,
							},
						},
					},
				)
				operations = append(operations, fmt.Sprintf("Replaced text at %d-%d", input.StartIndex, *input.EndIndex))
			} else {
				// Insert at position
				requests = append(requests, &docspb.Request{
					InsertText: &docspb.InsertTextRequest{
						Text: input.Text,
						Location: &docspb.Location{
							Index: input.StartIndex,
						},
					},
				})
				operations = append(operations, fmt.Sprintf("Inserted text at %d", input.StartIndex))
			}
		}

		// Formatting
		style := buildTextStyle(input.Bold, input.Italic, input.Underline, input.FontSize, input.FontFamily, input.TextColor, input.BackgroundColor)
		if style != nil {
			endIndex := input.StartIndex + int64(len(input.Text))
			if input.EndIndex != nil && input.Text == "" {
				endIndex = *input.EndIndex
			}
			fields := buildTextStyleFields(input.Bold, input.Italic, input.Underline, input.FontSize, input.FontFamily, input.TextColor, input.BackgroundColor)

			requests = append(requests, &docspb.Request{
				UpdateTextStyle: &docspb.UpdateTextStyleRequest{
					TextStyle: style,
					Range: &docspb.Range{
						StartIndex: input.StartIndex,
						EndIndex:   endIndex,
					},
					Fields: fields,
				},
			})
			operations = append(operations, fmt.Sprintf("Applied formatting (%s)", fields))
		}

		if len(requests) == 0 {
			return nil, nil, fmt.Errorf("no operation specified — provide text to insert/replace or formatting parameters")
		}

		_, err = srv.Documents.BatchUpdate(input.DocumentID, &docspb.BatchUpdateDocumentRequest{
			Requests: requests,
		}).Context(ctx).Do()
		if err != nil {
			return nil, nil, middleware.HandleGoogleAPIError(err)
		}

		rb := response.New()
		rb.Header("Document Modified")
		rb.KeyValue("Document ID", input.DocumentID)
		for _, op := range operations {
			rb.Item("%s", op)
		}

		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: rb.Build()}},
		}, nil, nil
	}
}

// --- export_doc_to_pdf (extended) ---

type ExportDocToPDFInput struct {
	UserEmail  string `json:"user_google_email" jsonschema:"required" jsonschema_description:"The user's Google email address"`
	DocumentID string `json:"document_id" jsonschema:"required" jsonschema_description:"The document ID to export"`
}

type ExportDocToPDFOutput struct {
	DownloadURL string `json:"download_url"`
	DocumentID  string `json:"document_id"`
}

func createExportDocToPDFHandler(factory *services.Factory) mcp.ToolHandlerFor[ExportDocToPDFInput, ExportDocToPDFOutput] {
	return func(ctx context.Context, req *mcp.CallToolRequest, input ExportDocToPDFInput) (*mcp.CallToolResult, ExportDocToPDFOutput, error) {
		url := fmt.Sprintf("https://www.googleapis.com/drive/v3/files/%s/export?mimeType=application/pdf", input.DocumentID)

		rb := response.New()
		rb.Header("Document Export URL")
		rb.KeyValue("Document ID", input.DocumentID)
		rb.KeyValue("Format", "PDF")
		rb.KeyValue("Download URL", url)

		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: rb.Build()}},
		}, ExportDocToPDFOutput{DownloadURL: url, DocumentID: input.DocumentID}, nil
	}
}

// --- search_docs (extended) ---

type SearchDocsInput struct {
	UserEmail string `json:"user_google_email" jsonschema:"required" jsonschema_description:"The user's Google email address"`
	Query     string `json:"query" jsonschema:"required" jsonschema_description:"Search query for Google Docs"`
	PageSize  int    `json:"page_size,omitempty" jsonschema_description:"Maximum results to return (default 10)"`
}

type SearchDocsOutput struct {
	Files []DocSearchResult `json:"files"`
}

type DocSearchResult struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	ModifiedTime string `json:"modified_time,omitempty"`
	WebViewLink  string `json:"web_view_link,omitempty"`
}

func createSearchDocsHandler(factory *services.Factory) mcp.ToolHandlerFor[SearchDocsInput, SearchDocsOutput] {
	return func(ctx context.Context, req *mcp.CallToolRequest, input SearchDocsInput) (*mcp.CallToolResult, SearchDocsOutput, error) {
		if input.PageSize == 0 {
			input.PageSize = 10
		}

		srv, err := factory.Drive(ctx, input.UserEmail)
		if err != nil {
			return nil, SearchDocsOutput{}, middleware.HandleGoogleAPIError(err)
		}

		// Search for Google Docs only
		q := fmt.Sprintf("mimeType='application/vnd.google-apps.document' and %s", input.Query)

		result, err := srv.Files.List().
			Q(q).
			PageSize(int64(input.PageSize)).
			Fields("files(id, name, modifiedTime, webViewLink)").
			SupportsAllDrives(true).
			IncludeItemsFromAllDrives(true).
			Context(ctx).
			Do()
		if err != nil {
			return nil, SearchDocsOutput{}, middleware.HandleGoogleAPIError(err)
		}

		files := make([]DocSearchResult, 0, len(result.Files))
		rb := response.New()
		rb.Header("Google Docs Search Results")
		rb.KeyValue("Query", input.Query)
		rb.KeyValue("Results", len(result.Files))
		rb.Blank()

		for _, f := range result.Files {
			files = append(files, DocSearchResult{
				ID:           f.Id,
				Name:         f.Name,
				ModifiedTime: f.ModifiedTime,
				WebViewLink:  f.WebViewLink,
			})
			rb.Item("%s", f.Name)
			rb.Line("    ID: %s | Modified: %s", f.Id, f.ModifiedTime)
			if f.WebViewLink != "" {
				rb.Line("    Link: %s", f.WebViewLink)
			}
		}

		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: rb.Build()}},
		}, SearchDocsOutput{Files: files}, nil
	}
}

// --- find_and_replace_doc (extended) ---

type FindAndReplaceDocInput struct {
	UserEmail  string `json:"user_google_email" jsonschema:"required" jsonschema_description:"The user's Google email address"`
	DocumentID string `json:"document_id" jsonschema:"required" jsonschema_description:"The document ID"`
	FindText   string `json:"find_text" jsonschema:"required" jsonschema_description:"Text to find"`
	ReplaceText string `json:"replace_text" jsonschema:"required" jsonschema_description:"Text to replace with"`
	MatchCase   bool   `json:"match_case,omitempty" jsonschema_description:"Case-sensitive matching (default false)"`
}

func createFindAndReplaceDocHandler(factory *services.Factory) mcp.ToolHandlerFor[FindAndReplaceDocInput, any] {
	return func(ctx context.Context, req *mcp.CallToolRequest, input FindAndReplaceDocInput) (*mcp.CallToolResult, any, error) {
		srv, err := factory.Docs(ctx, input.UserEmail)
		if err != nil {
			return nil, nil, middleware.HandleGoogleAPIError(err)
		}

		result, err := srv.Documents.BatchUpdate(input.DocumentID, &docspb.BatchUpdateDocumentRequest{
			Requests: []*docspb.Request{
				{
					ReplaceAllText: &docspb.ReplaceAllTextRequest{
						ContainsText: &docspb.SubstringMatchCriteria{
							Text:      input.FindText,
							MatchCase: input.MatchCase,
						},
						ReplaceText: input.ReplaceText,
					},
				},
			},
		}).Context(ctx).Do()
		if err != nil {
			return nil, nil, middleware.HandleGoogleAPIError(err)
		}

		replacements := 0
		if len(result.Replies) > 0 && result.Replies[0].ReplaceAllText != nil {
			replacements = int(result.Replies[0].ReplaceAllText.OccurrencesChanged)
		}

		rb := response.New()
		rb.Header("Find and Replace Complete")
		rb.KeyValue("Document ID", input.DocumentID)
		rb.KeyValue("Find", input.FindText)
		rb.KeyValue("Replace", input.ReplaceText)
		rb.KeyValue("Replacements", replacements)

		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: rb.Build()}},
		}, nil, nil
	}
}

// --- list_docs_in_folder (extended) ---

type ListDocsInFolderInput struct {
	UserEmail string `json:"user_google_email" jsonschema:"required" jsonschema_description:"The user's Google email address"`
	FolderID  string `json:"folder_id" jsonschema:"required" jsonschema_description:"The Drive folder ID to list documents from"`
	PageSize  int    `json:"page_size,omitempty" jsonschema_description:"Maximum results (default 25)"`
}

type ListDocsInFolderOutput struct {
	Documents []DocSearchResult `json:"documents"`
}

func createListDocsInFolderHandler(factory *services.Factory) mcp.ToolHandlerFor[ListDocsInFolderInput, ListDocsInFolderOutput] {
	return func(ctx context.Context, req *mcp.CallToolRequest, input ListDocsInFolderInput) (*mcp.CallToolResult, ListDocsInFolderOutput, error) {
		if input.PageSize == 0 {
			input.PageSize = 25
		}

		srv, err := factory.Drive(ctx, input.UserEmail)
		if err != nil {
			return nil, ListDocsInFolderOutput{}, middleware.HandleGoogleAPIError(err)
		}

		q := fmt.Sprintf("'%s' in parents and mimeType='application/vnd.google-apps.document' and trashed=false", input.FolderID)

		result, err := srv.Files.List().
			Q(q).
			PageSize(int64(input.PageSize)).
			Fields("files(id, name, modifiedTime, webViewLink)").
			SupportsAllDrives(true).
			IncludeItemsFromAllDrives(true).
			Context(ctx).
			Do()
		if err != nil {
			return nil, ListDocsInFolderOutput{}, middleware.HandleGoogleAPIError(err)
		}

		docs := make([]DocSearchResult, 0, len(result.Files))
		rb := response.New()
		rb.Header("Documents in Folder")
		rb.KeyValue("Folder ID", input.FolderID)
		rb.KeyValue("Count", len(result.Files))
		rb.Blank()

		for _, f := range result.Files {
			docs = append(docs, DocSearchResult{
				ID:           f.Id,
				Name:         f.Name,
				ModifiedTime: f.ModifiedTime,
				WebViewLink:  f.WebViewLink,
			})
			rb.Item("%s", f.Name)
			rb.Line("    ID: %s | Modified: %s", f.Id, f.ModifiedTime)
		}

		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: rb.Build()}},
		}, ListDocsInFolderOutput{Documents: docs}, nil
	}
}

// --- insert_doc_elements (extended) ---

type InsertDocElementsInput struct {
	UserEmail  string       `json:"user_google_email" jsonschema:"required" jsonschema_description:"The user's Google email address"`
	DocumentID string       `json:"document_id" jsonschema:"required" jsonschema_description:"The document ID"`
	Elements   []DocElement `json:"elements" jsonschema:"required" jsonschema_description:"Array of elements to insert"`
}

// DocElement represents a document element to insert.
type DocElement struct {
	Type    string `json:"type" jsonschema:"required" jsonschema_description:"Element type: paragraph or list_item,enum=paragraph,enum=list_item"`
	Text    string `json:"text" jsonschema:"required" jsonschema_description:"Text content"`
	Index   int64  `json:"index" jsonschema:"required" jsonschema_description:"Insertion index (1-based)"`
}

func createInsertDocElementsHandler(factory *services.Factory) mcp.ToolHandlerFor[InsertDocElementsInput, any] {
	return func(ctx context.Context, req *mcp.CallToolRequest, input InsertDocElementsInput) (*mcp.CallToolResult, any, error) {
		srv, err := factory.Docs(ctx, input.UserEmail)
		if err != nil {
			return nil, nil, middleware.HandleGoogleAPIError(err)
		}

		// Build requests in reverse order to maintain correct indices
		requests := make([]*docspb.Request, 0, len(input.Elements))
		for i := len(input.Elements) - 1; i >= 0; i-- {
			elem := input.Elements[i]
			text := elem.Text
			if !strings.HasSuffix(text, "\n") {
				text += "\n"
			}
			requests = append(requests, &docspb.Request{
				InsertText: &docspb.InsertTextRequest{
					Text: text,
					Location: &docspb.Location{
						Index: elem.Index,
					},
				},
			})
		}

		_, err = srv.Documents.BatchUpdate(input.DocumentID, &docspb.BatchUpdateDocumentRequest{
			Requests: requests,
		}).Context(ctx).Do()
		if err != nil {
			return nil, nil, middleware.HandleGoogleAPIError(err)
		}

		rb := response.New()
		rb.Header("Elements Inserted")
		rb.KeyValue("Document ID", input.DocumentID)
		rb.KeyValue("Elements inserted", len(input.Elements))

		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: rb.Build()}},
		}, nil, nil
	}
}

// --- update_paragraph_style (extended) ---

type UpdateParagraphStyleInput struct {
	UserEmail   string `json:"user_google_email" jsonschema:"required" jsonschema_description:"The user's Google email address"`
	DocumentID  string `json:"document_id" jsonschema:"required" jsonschema_description:"The document ID"`
	StartIndex  int64  `json:"start_index" jsonschema:"required" jsonschema_description:"Start of the range to style"`
	EndIndex    int64  `json:"end_index" jsonschema:"required" jsonschema_description:"End of the range to style"`
	NamedStyle  string `json:"named_style,omitempty" jsonschema_description:"Named style type: NORMAL_TEXT HEADING_1 HEADING_2 HEADING_3 HEADING_4 HEADING_5 HEADING_6 TITLE SUBTITLE"`
	Alignment   string `json:"alignment,omitempty" jsonschema_description:"Paragraph alignment: START CENTER END JUSTIFIED"`
	SpaceAbove  *float64 `json:"space_above,omitempty" jsonschema_description:"Space above paragraph in points"`
	SpaceBelow  *float64 `json:"space_below,omitempty" jsonschema_description:"Space below paragraph in points"`
	IndentStart *float64 `json:"indent_start,omitempty" jsonschema_description:"Left indent in points"`
}

func createUpdateParagraphStyleHandler(factory *services.Factory) mcp.ToolHandlerFor[UpdateParagraphStyleInput, any] {
	return func(ctx context.Context, req *mcp.CallToolRequest, input UpdateParagraphStyleInput) (*mcp.CallToolResult, any, error) {
		srv, err := factory.Docs(ctx, input.UserEmail)
		if err != nil {
			return nil, nil, middleware.HandleGoogleAPIError(err)
		}

		style := &docspb.ParagraphStyle{}
		fields := make([]string, 0)

		if input.NamedStyle != "" {
			style.NamedStyleType = input.NamedStyle
			fields = append(fields, "namedStyleType")
		}
		if input.Alignment != "" {
			style.Alignment = input.Alignment
			fields = append(fields, "alignment")
		}
		if input.SpaceAbove != nil {
			style.SpaceAbove = &docspb.Dimension{Magnitude: *input.SpaceAbove, Unit: "PT"}
			fields = append(fields, "spaceAbove")
		}
		if input.SpaceBelow != nil {
			style.SpaceBelow = &docspb.Dimension{Magnitude: *input.SpaceBelow, Unit: "PT"}
			fields = append(fields, "spaceBelow")
		}
		if input.IndentStart != nil {
			style.IndentStart = &docspb.Dimension{Magnitude: *input.IndentStart, Unit: "PT"}
			fields = append(fields, "indentStart")
		}

		if len(fields) == 0 {
			return nil, nil, fmt.Errorf("no style changes specified — provide at least one style parameter (named_style, alignment, space_above, space_below, indent_start)")
		}

		_, err = srv.Documents.BatchUpdate(input.DocumentID, &docspb.BatchUpdateDocumentRequest{
			Requests: []*docspb.Request{
				{
					UpdateParagraphStyle: &docspb.UpdateParagraphStyleRequest{
						ParagraphStyle: style,
						Range: &docspb.Range{
							StartIndex: input.StartIndex,
							EndIndex:   input.EndIndex,
						},
						Fields: strings.Join(fields, ","),
					},
				},
			},
		}).Context(ctx).Do()
		if err != nil {
			return nil, nil, middleware.HandleGoogleAPIError(err)
		}

		rb := response.New()
		rb.Header("Paragraph Style Updated")
		rb.KeyValue("Document ID", input.DocumentID)
		rb.KeyValue("Range", fmt.Sprintf("%d-%d", input.StartIndex, input.EndIndex))
		rb.KeyValue("Fields", strings.Join(fields, ", "))

		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: rb.Build()}},
		}, nil, nil
	}
}

