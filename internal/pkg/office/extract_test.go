package office

import (
	"archive/zip"
	"bytes"
	"testing"
)

// createTestDocx creates a minimal .docx file in memory for testing.
func createTestDocx(content string) []byte {
	buf := new(bytes.Buffer)
	w := zip.NewWriter(buf)

	f, _ := w.Create("word/document.xml")
	_, _ = f.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?>
<w:document xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main">
  <w:body>
    <w:p><w:r><w:t>` + content + `</w:t></w:r></w:p>
  </w:body>
</w:document>`))

	_ = w.Close()
	return buf.Bytes()
}

// createTestPptx creates a minimal .pptx file in memory for testing.
func createTestPptx(content string) []byte {
	buf := new(bytes.Buffer)
	w := zip.NewWriter(buf)

	f, _ := w.Create("ppt/slides/slide1.xml")
	_, _ = f.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?>
<p:sld xmlns:p="http://schemas.openxmlformats.org/presentationml/2006/main"
       xmlns:a="http://schemas.openxmlformats.org/drawingml/2006/main">
  <p:cSld><p:spTree><p:sp><p:txBody><a:p><a:r><a:t>` + content + `</a:t></a:r></a:p></p:txBody></p:sp></p:spTree></p:cSld>
</p:sld>`))

	_ = w.Close()
	return buf.Bytes()
}

func TestExtractDocx(t *testing.T) {
	data := createTestDocx("Hello World")
	text, err := ExtractText(data, "application/vnd.openxmlformats-officedocument.wordprocessingml.document")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if text != "Hello World" {
		t.Errorf("got %q, want %q", text, "Hello World")
	}
}

func TestExtractPptx(t *testing.T) {
	data := createTestPptx("Slide Title")
	text, err := ExtractText(data, "application/vnd.openxmlformats-officedocument.presentationml.presentation")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if text != "Slide Title" {
		t.Errorf("got %q, want %q", text, "Slide Title")
	}
}

func TestExtractTextInvalidZip(t *testing.T) {
	_, err := ExtractText([]byte("not a zip file"), "application/vnd.openxmlformats-officedocument.wordprocessingml.document")
	if err == nil {
		t.Error("expected error for invalid ZIP data")
	}
}

func TestExtractTextTooLarge(t *testing.T) {
	data := make([]byte, MaxFileSize+1)
	_, err := ExtractText(data, "application/vnd.openxmlformats-officedocument.wordprocessingml.document")
	if err == nil {
		t.Error("expected error for oversized file")
	}
}
