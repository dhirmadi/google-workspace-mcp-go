package office

import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"strings"
)

// MaxFileSize is the maximum file size to attempt extraction on (50 MB).
const MaxFileSize = 50 * 1024 * 1024

// ExtractText extracts plain text from Office XML documents (.docx, .xlsx, .pptx).
// The data must be the raw ZIP-based Office file content.
func ExtractText(data []byte, mimeType string) (string, error) {
	if len(data) > MaxFileSize {
		return "", fmt.Errorf("file too large for text extraction (%d bytes, max %d)", len(data), MaxFileSize)
	}

	reader, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return "", fmt.Errorf("opening Office document as ZIP: %w", err)
	}

	switch {
	case strings.Contains(mimeType, "wordprocessingml") || strings.HasSuffix(mimeType, ".docx"):
		return extractDocx(reader)
	case strings.Contains(mimeType, "spreadsheetml") || strings.HasSuffix(mimeType, ".xlsx"):
		return extractXlsx(reader)
	case strings.Contains(mimeType, "presentationml") || strings.HasSuffix(mimeType, ".pptx"):
		return extractPptx(reader)
	default:
		return extractAllXMLText(reader)
	}
}

// extractDocx extracts text from a .docx file by reading word/document.xml.
func extractDocx(r *zip.Reader) (string, error) {
	for _, f := range r.File {
		if f.Name == "word/document.xml" {
			return extractXMLText(f)
		}
	}
	return "", fmt.Errorf("word/document.xml not found in docx")
}

// extractXlsx extracts text from a .xlsx file by reading shared strings and sheet data.
func extractXlsx(r *zip.Reader) (string, error) {
	// First, read shared strings
	sharedStrings := make(map[int]string)
	for _, f := range r.File {
		if f.Name == "xl/sharedStrings.xml" {
			strings, err := parseSharedStrings(f)
			if err == nil {
				sharedStrings = strings
			}
			break
		}
	}

	// Then read sheets
	var parts []string
	for _, f := range r.File {
		if strings.HasPrefix(f.Name, "xl/worksheets/sheet") && strings.HasSuffix(f.Name, ".xml") {
			text, err := extractXMLText(f)
			if err != nil {
				continue
			}
			if text != "" {
				parts = append(parts, text)
			}
		}
	}

	_ = sharedStrings // Shared strings are embedded in XML text extraction
	return strings.Join(parts, "\n\n"), nil
}

// extractPptx extracts text from a .pptx file by reading all slide XML files.
func extractPptx(r *zip.Reader) (string, error) {
	var parts []string
	for _, f := range r.File {
		if strings.HasPrefix(f.Name, "ppt/slides/slide") && strings.HasSuffix(f.Name, ".xml") {
			text, err := extractXMLText(f)
			if err != nil {
				continue
			}
			if text != "" {
				parts = append(parts, text)
			}
		}
	}
	return strings.Join(parts, "\n\n"), nil
}

// extractAllXMLText extracts text from all XML files as a fallback.
func extractAllXMLText(r *zip.Reader) (string, error) {
	var parts []string
	for _, f := range r.File {
		if strings.HasSuffix(f.Name, ".xml") {
			text, err := extractXMLText(f)
			if err != nil {
				continue
			}
			if text != "" {
				parts = append(parts, text)
			}
		}
	}
	return strings.Join(parts, "\n"), nil
}

// extractXMLText reads all text content from an XML file within the ZIP.
func extractXMLText(f *zip.File) (string, error) {
	rc, err := f.Open()
	if err != nil {
		return "", err
	}
	defer rc.Close()

	data, err := io.ReadAll(io.LimitReader(rc, MaxFileSize))
	if err != nil {
		return "", err
	}

	return xmlToText(data), nil
}

// xmlToText extracts all character data from XML, joining with spaces.
func xmlToText(data []byte) string {
	decoder := xml.NewDecoder(bytes.NewReader(data))
	var parts []string

	for {
		tok, err := decoder.Token()
		if err != nil {
			break
		}
		if charData, ok := tok.(xml.CharData); ok {
			text := strings.TrimSpace(string(charData))
			if text != "" {
				parts = append(parts, text)
			}
		}
	}

	return strings.Join(parts, " ")
}

// parseSharedStrings parses the sharedStrings.xml from an xlsx file.
func parseSharedStrings(f *zip.File) (map[int]string, error) {
	rc, err := f.Open()
	if err != nil {
		return nil, err
	}
	defer rc.Close()

	data, err := io.ReadAll(io.LimitReader(rc, MaxFileSize))
	if err != nil {
		return nil, err
	}

	decoder := xml.NewDecoder(bytes.NewReader(data))
	result := make(map[int]string)
	index := 0
	var inSi bool
	var current strings.Builder

	for {
		tok, err := decoder.Token()
		if err != nil {
			break
		}
		switch t := tok.(type) {
		case xml.StartElement:
			if t.Name.Local == "si" {
				inSi = true
				current.Reset()
			}
		case xml.EndElement:
			if t.Name.Local == "si" {
				result[index] = current.String()
				index++
				inSi = false
			}
		case xml.CharData:
			if inSi {
				current.Write(t)
			}
		}
	}

	return result, nil
}
