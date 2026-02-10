package htmlutil

import (
	"testing"
)

func TestToPlainTextEmpty(t *testing.T) {
	if got := ToPlainText(""); got != "" {
		t.Errorf("ToPlainText(\"\") = %q, want \"\"", got)
	}
}

func TestToPlainTextBasicHTML(t *testing.T) {
	input := "<p>Hello <b>World</b></p>"
	got := ToPlainText(input)
	if got != "Hello World" {
		t.Errorf("got %q, want %q", got, "Hello World")
	}
}

func TestToPlainTextLineBreaks(t *testing.T) {
	input := "Line one<br>Line two<br/>Line three"
	got := ToPlainText(input)
	want := "Line one\nLine two\nLine three"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestToPlainTextEntities(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"&amp; &lt; &gt;", "& < >"},
		{"&quot;hello&quot;", `"hello"`},
		{"&nbsp;space&nbsp;", "space"},
		{"&mdash; dash &ndash;", "— dash –"},
		{"&copy; 2025", "© 2025"},
	}
	for _, tt := range tests {
		got := ToPlainText(tt.input)
		if got != tt.want {
			t.Errorf("ToPlainText(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestToPlainTextNumericEntities(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"decimal entity", "&#8212; em dash", "— em dash"},
		{"hex entity lowercase", "&#x2019; curly quote", "\u2019 curly quote"},
		{"hex entity uppercase", "&#x2018; curly quote", "\u2018 curly quote"},
		{"star entity", "&#9733; star", "★ star"},
		{"emoji hex entity", "&#x1F600; grin", "\U0001F600 grin"},
		{"mixed entities", "&amp; &#38; &#x26;", "& & &"},
		{"entity in HTML", "<p>Price: &#8364;10</p>", "Price: €10"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ToPlainText(tt.input)
			if got != tt.want {
				t.Errorf("ToPlainText(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestToPlainTextStripStyleScript(t *testing.T) {
	input := `<style>body { color: red; }</style><p>Hello</p><script>alert("x")</script>`
	got := ToPlainText(input)
	if got != "Hello" {
		t.Errorf("got %q, want %q", got, "Hello")
	}
}

func TestToPlainTextBlockElements(t *testing.T) {
	input := "<h1>Title</h1><p>Paragraph one</p><p>Paragraph two</p>"
	got := ToPlainText(input)
	// Should have newlines between blocks, content preserved
	if !containsAll(got, "Title", "Paragraph one", "Paragraph two") {
		t.Errorf("missing content in %q", got)
	}
}

func TestToPlainTextCollapseWhitespace(t *testing.T) {
	input := "  lots   of    spaces  "
	got := ToPlainText(input)
	if got != "lots of spaces" {
		t.Errorf("got %q, want %q", got, "lots of spaces")
	}
}

func TestToPlainTextCollapseBlankLines(t *testing.T) {
	input := "Line one\n\n\n\n\nLine two"
	got := ToPlainText(input)
	if got != "Line one\n\nLine two" {
		t.Errorf("got %q, want %q", got, "Line one\n\nLine two")
	}
}

func containsAll(s string, subs ...string) bool {
	for _, sub := range subs {
		if !contains(s, sub) {
			return false
		}
	}
	return true
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && searchString(s, sub)
}

func searchString(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
