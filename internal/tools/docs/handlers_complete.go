package docs

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	docspb "google.golang.org/api/docs/v1"

	"github.com/evert/google-workspace-mcp-go/internal/middleware"
	"github.com/evert/google-workspace-mcp-go/internal/pkg/response"
	"github.com/evert/google-workspace-mcp-go/internal/services"
)

// --- insert_doc_image (complete) ---

type InsertDocImageInput struct {
	UserEmail  string `json:"user_google_email" jsonschema:"required" jsonschema_description:"The user's Google email address"`
	DocumentID string `json:"document_id" jsonschema:"required" jsonschema_description:"The Google Doc document ID"`
	ImageURI   string `json:"image_uri" jsonschema:"required" jsonschema_description:"Public URL of the image to insert"`
	Index      int    `json:"index" jsonschema:"required" jsonschema_description:"Character index position where the image should be inserted (1-based)"`
	Width      int    `json:"width,omitempty" jsonschema_description:"Image width in points"`
	Height     int    `json:"height,omitempty" jsonschema_description:"Image height in points"`
}

func createInsertDocImageHandler(factory *services.Factory) mcp.ToolHandlerFor[InsertDocImageInput, any] {
	return func(ctx context.Context, req *mcp.CallToolRequest, input InsertDocImageInput) (*mcp.CallToolResult, any, error) {
		srv, err := factory.Docs(ctx, input.UserEmail)
		if err != nil {
			return nil, nil, middleware.HandleGoogleAPIError(err)
		}

		insertReq := &docspb.InsertInlineImageRequest{
			Uri: input.ImageURI,
			Location: &docspb.Location{
				Index: int64(input.Index),
			},
		}

		if input.Width > 0 && input.Height > 0 {
			insertReq.ObjectSize = &docspb.Size{
				Width: &docspb.Dimension{
					Magnitude: float64(input.Width),
					Unit:      "PT",
				},
				Height: &docspb.Dimension{
					Magnitude: float64(input.Height),
					Unit:      "PT",
				},
			}
		}

		batchReq := &docspb.BatchUpdateDocumentRequest{
			Requests: []*docspb.Request{
				{InsertInlineImage: insertReq},
			},
		}

		_, err = srv.Documents.BatchUpdate(input.DocumentID, batchReq).Context(ctx).Do()
		if err != nil {
			return nil, nil, middleware.HandleGoogleAPIError(err)
		}

		rb := response.New()
		rb.Header("Image Inserted")
		rb.KeyValue("Document ID", input.DocumentID)
		rb.KeyValue("Image URI", input.ImageURI)
		rb.KeyValue("Position", input.Index)

		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: rb.Build()}},
		}, nil, nil
	}
}

// --- update_doc_headers_footers (complete) ---

type UpdateHeadersFootersInput struct {
	UserEmail    string `json:"user_google_email" jsonschema:"required" jsonschema_description:"The user's Google email address"`
	DocumentID   string `json:"document_id" jsonschema:"required" jsonschema_description:"The Google Doc document ID"`
	HeaderText   string `json:"header_text,omitempty" jsonschema_description:"Text to insert into the default header"`
	FooterText   string `json:"footer_text,omitempty" jsonschema_description:"Text to insert into the default footer"`
	RemoveHeader bool   `json:"remove_header,omitempty" jsonschema_description:"Remove the default header"`
	RemoveFooter bool   `json:"remove_footer,omitempty" jsonschema_description:"Remove the default footer"`
}

func createUpdateHeadersFootersHandler(factory *services.Factory) mcp.ToolHandlerFor[UpdateHeadersFootersInput, any] {
	return func(ctx context.Context, req *mcp.CallToolRequest, input UpdateHeadersFootersInput) (*mcp.CallToolResult, any, error) {
		srv, err := factory.Docs(ctx, input.UserEmail)
		if err != nil {
			return nil, nil, middleware.HandleGoogleAPIError(err)
		}

		// Get the document to find existing header/footer IDs
		doc, err := srv.Documents.Get(input.DocumentID).Context(ctx).Do()
		if err != nil {
			return nil, nil, middleware.HandleGoogleAPIError(err)
		}

		var requests []*docspb.Request

		// Handle header
		if input.RemoveHeader {
			if doc.DocumentStyle != nil && doc.DocumentStyle.DefaultHeaderId != "" {
				requests = append(requests, &docspb.Request{
					DeleteHeader: &docspb.DeleteHeaderRequest{
						HeaderId: doc.DocumentStyle.DefaultHeaderId,
					},
				})
			}
		} else if input.HeaderText != "" {
			headerID := ""
			if doc.DocumentStyle != nil {
				headerID = doc.DocumentStyle.DefaultHeaderId
			}
			if headerID == "" {
				// Create a header first
				requests = append(requests, &docspb.Request{
					CreateHeader: &docspb.CreateHeaderRequest{
						Type: "DEFAULT",
					},
				})
			} else {
				// Insert text into existing header
				requests = append(requests, &docspb.Request{
					InsertText: &docspb.InsertTextRequest{
						Text: input.HeaderText,
						Location: &docspb.Location{
							SegmentId: headerID,
							Index:     0,
						},
					},
				})
			}
		}

		// Handle footer
		if input.RemoveFooter {
			if doc.DocumentStyle != nil && doc.DocumentStyle.DefaultFooterId != "" {
				requests = append(requests, &docspb.Request{
					DeleteFooter: &docspb.DeleteFooterRequest{
						FooterId: doc.DocumentStyle.DefaultFooterId,
					},
				})
			}
		} else if input.FooterText != "" {
			footerID := ""
			if doc.DocumentStyle != nil {
				footerID = doc.DocumentStyle.DefaultFooterId
			}
			if footerID == "" {
				requests = append(requests, &docspb.Request{
					CreateFooter: &docspb.CreateFooterRequest{
						Type: "DEFAULT",
					},
				})
			} else {
				requests = append(requests, &docspb.Request{
					InsertText: &docspb.InsertTextRequest{
						Text: input.FooterText,
						Location: &docspb.Location{
							SegmentId: footerID,
							Index:     0,
						},
					},
				})
			}
		}

		if len(requests) == 0 {
			return nil, nil, fmt.Errorf("no header/footer changes specified - set header_text, footer_text, remove_header, or remove_footer")
		}

		batchReq := &docspb.BatchUpdateDocumentRequest{
			Requests: requests,
		}

		_, err = srv.Documents.BatchUpdate(input.DocumentID, batchReq).Context(ctx).Do()
		if err != nil {
			return nil, nil, middleware.HandleGoogleAPIError(err)
		}

		rb := response.New()
		rb.Header("Headers/Footers Updated")
		rb.KeyValue("Document ID", input.DocumentID)
		rb.KeyValue("Changes Applied", len(requests))

		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: rb.Build()}},
		}, nil, nil
	}
}

// --- batch_update_doc (complete) ---

type BatchUpdateDocInput struct {
	UserEmail  string `json:"user_google_email" jsonschema:"required" jsonschema_description:"The user's Google email address"`
	DocumentID string `json:"document_id" jsonschema:"required" jsonschema_description:"The Google Doc document ID"`
	Requests   string `json:"requests" jsonschema:"required" jsonschema_description:"JSON array of document update requests. Each request can contain insertText deleteContentRange insertInlineImage updateTextStyle updateParagraphStyle etc."`
}

func createBatchUpdateDocHandler(factory *services.Factory) mcp.ToolHandlerFor[BatchUpdateDocInput, any] {
	return func(ctx context.Context, req *mcp.CallToolRequest, input BatchUpdateDocInput) (*mcp.CallToolResult, any, error) {
		srv, err := factory.Docs(ctx, input.UserEmail)
		if err != nil {
			return nil, nil, middleware.HandleGoogleAPIError(err)
		}

		var requests []*docspb.Request
		if err := json.Unmarshal([]byte(input.Requests), &requests); err != nil {
			return nil, nil, fmt.Errorf("invalid requests JSON - provide a JSON array of document update request objects: %w", err)
		}

		batchReq := &docspb.BatchUpdateDocumentRequest{
			Requests: requests,
		}

		result, err := srv.Documents.BatchUpdate(input.DocumentID, batchReq).Context(ctx).Do()
		if err != nil {
			return nil, nil, middleware.HandleGoogleAPIError(err)
		}

		rb := response.New()
		rb.Header("Document Updated")
		rb.KeyValue("Document ID", result.DocumentId)
		rb.KeyValue("Replies", len(result.Replies))

		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: rb.Build()}},
		}, nil, nil
	}
}

// --- inspect_doc_structure (complete) ---

type InspectDocStructureInput struct {
	UserEmail  string `json:"user_google_email" jsonschema:"required" jsonschema_description:"The user's Google email address"`
	DocumentID string `json:"document_id" jsonschema:"required" jsonschema_description:"The Google Doc document ID"`
}

func createInspectDocStructureHandler(factory *services.Factory) mcp.ToolHandlerFor[InspectDocStructureInput, DocStructureOutput] {
	return func(ctx context.Context, req *mcp.CallToolRequest, input InspectDocStructureInput) (*mcp.CallToolResult, DocStructureOutput, error) {
		srv, err := factory.Docs(ctx, input.UserEmail)
		if err != nil {
			return nil, DocStructureOutput{}, middleware.HandleGoogleAPIError(err)
		}

		doc, err := srv.Documents.Get(input.DocumentID).Context(ctx).Do()
		if err != nil {
			return nil, DocStructureOutput{}, middleware.HandleGoogleAPIError(err)
		}

		elements := extractStructureElements(doc)

		rb := response.New()
		rb.Header("Document Structure")
		rb.KeyValue("Title", doc.Title)
		rb.KeyValue("Document ID", doc.DocumentId)
		rb.KeyValue("Elements", len(elements))
		rb.Blank()

		for _, e := range elements {
			content := e.Content
			if len(content) > 80 {
				content = content[:80] + "..."
			}
			content = strings.ReplaceAll(content, "\n", "\\n")
			rb.Item("[%s] %d–%d: %s", e.Type, e.StartIndex, e.EndIndex, content)
		}

		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: rb.Build()}},
		}, DocStructureOutput{DocumentID: doc.DocumentId, Title: doc.Title, Elements: elements}, nil
	}
}

// --- create_table_with_data (complete) ---

type CreateTableWithDataInput struct {
	UserEmail  string     `json:"user_google_email" jsonschema:"required" jsonschema_description:"The user's Google email address"`
	DocumentID string     `json:"document_id" jsonschema:"required" jsonschema_description:"The Google Doc document ID"`
	Index      int        `json:"index" jsonschema:"required" jsonschema_description:"Character index where the table should be inserted (1-based)"`
	Rows       int        `json:"rows" jsonschema:"required" jsonschema_description:"Number of rows"`
	Columns    int        `json:"columns" jsonschema:"required" jsonschema_description:"Number of columns"`
	Data       [][]string `json:"data,omitempty" jsonschema_description:"2D array of cell values (rows x columns). If provided fills the table."`
}

func createCreateTableHandler(factory *services.Factory) mcp.ToolHandlerFor[CreateTableWithDataInput, any] {
	return func(ctx context.Context, req *mcp.CallToolRequest, input CreateTableWithDataInput) (*mcp.CallToolResult, any, error) {
		srv, err := factory.Docs(ctx, input.UserEmail)
		if err != nil {
			return nil, nil, middleware.HandleGoogleAPIError(err)
		}

		// First, insert the table
		insertReq := &docspb.BatchUpdateDocumentRequest{
			Requests: []*docspb.Request{
				{
					InsertTable: &docspb.InsertTableRequest{
						Rows:    int64(input.Rows),
						Columns: int64(input.Columns),
						Location: &docspb.Location{
							Index: int64(input.Index),
						},
					},
				},
			},
		}

		_, err = srv.Documents.BatchUpdate(input.DocumentID, insertReq).Context(ctx).Do()
		if err != nil {
			return nil, nil, middleware.HandleGoogleAPIError(err)
		}

		// If data is provided, fill the table cells
		if len(input.Data) > 0 {
			// Re-fetch the doc to get updated indices
			doc, err := srv.Documents.Get(input.DocumentID).Context(ctx).Do()
			if err != nil {
				return nil, nil, middleware.HandleGoogleAPIError(err)
			}

			// Find the table we just inserted
			var table *docspb.Table
			for _, elem := range doc.Body.Content {
				if elem.Table != nil && elem.StartIndex >= int64(input.Index) {
					table = elem.Table
					break
				}
			}

			if table != nil {
				var dataRequests []*docspb.Request
				// Iterate in reverse to avoid index shifting
				for r := len(input.Data) - 1; r >= 0 && r < int(table.Rows); r-- {
					row := input.Data[r]
					for c := len(row) - 1; c >= 0 && c < int(table.Columns); c-- {
						if row[c] == "" {
							continue
						}
						if r < len(table.TableRows) && c < len(table.TableRows[r].TableCells) {
							cell := table.TableRows[r].TableCells[c]
							if len(cell.Content) > 0 {
								idx := cell.Content[0].StartIndex
								dataRequests = append(dataRequests, &docspb.Request{
									InsertText: &docspb.InsertTextRequest{
										Text: row[c],
										Location: &docspb.Location{
											Index: idx,
										},
									},
								})
							}
						}
					}
				}

				if len(dataRequests) > 0 {
					dataBatch := &docspb.BatchUpdateDocumentRequest{
						Requests: dataRequests,
					}
					_, err = srv.Documents.BatchUpdate(input.DocumentID, dataBatch).Context(ctx).Do()
					if err != nil {
						return nil, nil, middleware.HandleGoogleAPIError(err)
					}
				}
			}
		}

		rb := response.New()
		rb.Header("Table Created")
		rb.KeyValue("Document ID", input.DocumentID)
		rb.KeyValue("Size", fmt.Sprintf("%dx%d", input.Rows, input.Columns))
		if len(input.Data) > 0 {
			rb.KeyValue("Data Rows", len(input.Data))
		}

		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: rb.Build()}},
		}, nil, nil
	}
}

// --- debug_table_structure (complete) ---

type DebugTableStructureInput struct {
	UserEmail  string `json:"user_google_email" jsonschema:"required" jsonschema_description:"The user's Google email address"`
	DocumentID string `json:"document_id" jsonschema:"required" jsonschema_description:"The Google Doc document ID"`
	TableIndex int    `json:"table_index,omitempty" jsonschema_description:"Which table to inspect (0-based default 0 for the first table)"`
}

type DebugTableOutput struct {
	DocumentID string         `json:"document_id"`
	TableIndex int            `json:"table_index"`
	Rows       int            `json:"rows"`
	Columns    int            `json:"columns"`
	StartIndex int64          `json:"start_index"`
	EndIndex   int64          `json:"end_index"`
	Cells      [][]CellDebug  `json:"cells"`
}

type CellDebug struct {
	Row        int    `json:"row"`
	Col        int    `json:"col"`
	StartIndex int64  `json:"start_index"`
	EndIndex   int64  `json:"end_index"`
	Content    string `json:"content"`
}

func createDebugTableStructureHandler(factory *services.Factory) mcp.ToolHandlerFor[DebugTableStructureInput, DebugTableOutput] {
	return func(ctx context.Context, req *mcp.CallToolRequest, input DebugTableStructureInput) (*mcp.CallToolResult, DebugTableOutput, error) {
		srv, err := factory.Docs(ctx, input.UserEmail)
		if err != nil {
			return nil, DebugTableOutput{}, middleware.HandleGoogleAPIError(err)
		}

		doc, err := srv.Documents.Get(input.DocumentID).Context(ctx).Do()
		if err != nil {
			return nil, DebugTableOutput{}, middleware.HandleGoogleAPIError(err)
		}

		// Find the nth table
		tableIdx := 0
		var tableElem *docspb.StructuralElement
		for _, elem := range doc.Body.Content {
			if elem.Table != nil {
				if tableIdx == input.TableIndex {
					tableElem = elem
					break
				}
				tableIdx++
			}
		}

		if tableElem == nil || tableElem.Table == nil {
			return nil, DebugTableOutput{}, fmt.Errorf("table at index %d not found in document - verify the table_index parameter", input.TableIndex)
		}

		table := tableElem.Table
		output := DebugTableOutput{
			DocumentID: doc.DocumentId,
			TableIndex: input.TableIndex,
			Rows:       int(table.Rows),
			Columns:    int(table.Columns),
			StartIndex: tableElem.StartIndex,
			EndIndex:   tableElem.EndIndex,
		}

		rb := response.New()
		rb.Header("Table Structure Debug")
		rb.KeyValue("Document", doc.Title)
		rb.KeyValue("Table Index", input.TableIndex)
		rb.KeyValue("Size", fmt.Sprintf("%dx%d", table.Rows, table.Columns))
		rb.KeyValue("Range", fmt.Sprintf("%d–%d", tableElem.StartIndex, tableElem.EndIndex))
		rb.Blank()

		for r, row := range table.TableRows {
			var rowCells []CellDebug
			for c, cell := range row.TableCells {
				cellContent := extractCellText(cell)
				startIdx := int64(0)
				endIdx := int64(0)
				if len(cell.Content) > 0 {
					startIdx = cell.Content[0].StartIndex
					endIdx = cell.Content[len(cell.Content)-1].EndIndex
				}

				cd := CellDebug{
					Row:        r,
					Col:        c,
					StartIndex: startIdx,
					EndIndex:   endIdx,
					Content:    cellContent,
				}
				rowCells = append(rowCells, cd)

				content := cellContent
				if len(content) > 40 {
					content = content[:40] + "..."
				}
				rb.Item("[%d,%d] %d–%d: %q", r, c, startIdx, endIdx, content)
			}
			output.Cells = append(output.Cells, rowCells)
		}

		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: rb.Build()}},
		}, output, nil
	}
}
