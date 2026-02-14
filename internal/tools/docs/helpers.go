package docs

import (
	"fmt"
	"strings"

	docspb "google.golang.org/api/docs/v1"

	"github.com/evert/google-workspace-mcp-go/internal/pkg/color"
)

// DocSummary is a compact representation of a Google Doc.
type DocSummary struct {
	DocumentID string `json:"document_id"`
	Title      string `json:"title"`
	RevisionID string `json:"revision_id,omitempty"`
}

// DocContentOutput is the structured output for get_doc_content.
type DocContentOutput struct {
	DocumentID string `json:"document_id"`
	Title      string `json:"title"`
	Content    string `json:"content"`
}

// DocStructureOutput is the structured output for inspect_doc_structure.
type DocStructureOutput struct {
	DocumentID string             `json:"document_id"`
	Title      string             `json:"title"`
	Elements   []StructureElement `json:"elements"`
}

// StructureElement represents a structural element in a document.
type StructureElement struct {
	Type       string `json:"type"`
	StartIndex int64  `json:"start_index"`
	EndIndex   int64  `json:"end_index"`
	Content    string `json:"content,omitempty"`
}

// extractDocText extracts all plain text from a Google Doc body.
func extractDocText(doc *docspb.Document) string {
	if doc.Body == nil {
		return ""
	}

	var sb strings.Builder
	for _, elem := range doc.Body.Content {
		if elem.Paragraph != nil {
			for _, pe := range elem.Paragraph.Elements {
				if pe.TextRun != nil {
					sb.WriteString(pe.TextRun.Content)
				}
			}
		}
		if elem.Table != nil {
			for _, row := range elem.Table.TableRows {
				cells := make([]string, 0, len(row.TableCells))
				for _, cell := range row.TableCells {
					cellText := extractCellText(cell)
					cells = append(cells, cellText)
				}
				sb.WriteString(strings.Join(cells, " | "))
				sb.WriteByte('\n')
			}
		}
	}
	return sb.String()
}

// extractCellText extracts text from a table cell.
func extractCellText(cell *docspb.TableCell) string {
	var sb strings.Builder
	for _, elem := range cell.Content {
		if elem.Paragraph != nil {
			for _, pe := range elem.Paragraph.Elements {
				if pe.TextRun != nil {
					sb.WriteString(strings.TrimSpace(pe.TextRun.Content))
				}
			}
		}
	}
	return sb.String()
}

// extractStructureElements extracts structural information from a document.
func extractStructureElements(doc *docspb.Document) []StructureElement {
	if doc.Body == nil {
		return nil
	}

	elements := make([]StructureElement, 0, len(doc.Body.Content))
	for _, elem := range doc.Body.Content {
		se := StructureElement{
			StartIndex: elem.StartIndex,
			EndIndex:   elem.EndIndex,
		}

		switch {
		case elem.Paragraph != nil:
			se.Type = "paragraph"
			var content strings.Builder
			for _, pe := range elem.Paragraph.Elements {
				if pe.TextRun != nil {
					content.WriteString(pe.TextRun.Content)
				}
			}
			se.Content = content.String()
			if elem.Paragraph.ParagraphStyle != nil && elem.Paragraph.ParagraphStyle.NamedStyleType != "" {
				se.Type = fmt.Sprintf("paragraph(%s)", elem.Paragraph.ParagraphStyle.NamedStyleType)
			}
		case elem.Table != nil:
			se.Type = fmt.Sprintf("table(%dx%d)", elem.Table.Rows, elem.Table.Columns)
		case elem.SectionBreak != nil:
			se.Type = "section_break"
		case elem.TableOfContents != nil:
			se.Type = "table_of_contents"
		}

		elements = append(elements, se)
	}
	return elements
}

// buildTextStyle constructs a TextStyle from formatting parameters.
func buildTextStyle(bold, italic, underline *bool, fontSize *int, fontFamily, textColor, bgColor string) *docspb.TextStyle {
	style := &docspb.TextStyle{}
	hasStyle := false

	if bold != nil {
		style.Bold = *bold
		hasStyle = true
	}
	if italic != nil {
		style.Italic = *italic
		hasStyle = true
	}
	if underline != nil {
		style.Underline = *underline
		hasStyle = true
	}
	if fontSize != nil && *fontSize > 0 {
		style.FontSize = &docspb.Dimension{
			Magnitude: float64(*fontSize),
			Unit:      "PT",
		}
		hasStyle = true
	}
	if fontFamily != "" {
		style.WeightedFontFamily = &docspb.WeightedFontFamily{
			FontFamily: fontFamily,
		}
		hasStyle = true
	}
	if textColor != "" {
		style.ForegroundColor = parseColor(textColor)
		hasStyle = true
	}
	if bgColor != "" {
		style.BackgroundColor = parseColor(bgColor)
		hasStyle = true
	}

	if !hasStyle {
		return nil
	}
	return style
}

// buildTextStyleFields builds the fields mask for a TextStyle update.
func buildTextStyleFields(bold, italic, underline *bool, fontSize *int, fontFamily, textColor, bgColor string) string {
	fields := make([]string, 0, 7)
	if bold != nil {
		fields = append(fields, "bold")
	}
	if italic != nil {
		fields = append(fields, "italic")
	}
	if underline != nil {
		fields = append(fields, "underline")
	}
	if fontSize != nil {
		fields = append(fields, "fontSize")
	}
	if fontFamily != "" {
		fields = append(fields, "weightedFontFamily")
	}
	if textColor != "" {
		fields = append(fields, "foregroundColor")
	}
	if bgColor != "" {
		fields = append(fields, "backgroundColor")
	}
	return strings.Join(fields, ",")
}

// parseColor converts a hex color (#RRGGBB) to a Docs OptionalColor.
func parseColor(hex string) *docspb.OptionalColor {
	r, g, b, ok := color.HexToRGB(hex)
	if !ok {
		return nil
	}
	return &docspb.OptionalColor{
		Color: &docspb.Color{
			RgbColor: &docspb.RgbColor{
				Red:   r,
				Green: g,
				Blue:  b,
			},
		},
	}
}
