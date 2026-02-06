package response

import (
	"strings"
	"testing"
)

func TestBuilderHeader(t *testing.T) {
	got := New().Header("Test %s", "Header").Build()
	if !strings.Contains(got, "Test Header") {
		t.Errorf("Header missing expected text, got: %s", got)
	}
	if !strings.Contains(got, "═══") {
		t.Errorf("Header missing decoration, got: %s", got)
	}
}

func TestBuilderKeyValue(t *testing.T) {
	got := New().KeyValue("Name", "Alice").Build()
	want := "• Name: Alice\n"
	if got != want {
		t.Errorf("KeyValue = %q, want %q", got, want)
	}
}

func TestBuilderItem(t *testing.T) {
	got := New().Item("item %d", 1).Build()
	want := "  → item 1\n"
	if got != want {
		t.Errorf("Item = %q, want %q", got, want)
	}
}

func TestBuilderLine(t *testing.T) {
	got := New().Line("hello %s", "world").Build()
	want := "hello world\n"
	if got != want {
		t.Errorf("Line = %q, want %q", got, want)
	}
}

func TestBuilderBlank(t *testing.T) {
	got := New().Blank().Build()
	if got != "\n" {
		t.Errorf("Blank = %q, want %q", got, "\n")
	}
}

func TestBuilderComposite(t *testing.T) {
	got := New().
		Header("Results").
		KeyValue("Count", 3).
		Blank().
		Item("First").
		Item("Second").
		Separator().
		Section("Details").
		Line("Some detail").
		Build()

	if !strings.Contains(got, "Results") {
		t.Error("missing header")
	}
	if !strings.Contains(got, "Count: 3") {
		t.Error("missing key-value")
	}
	if !strings.Contains(got, "→ First") {
		t.Error("missing first item")
	}
	if !strings.Contains(got, "→ Second") {
		t.Error("missing second item")
	}
	if !strings.Contains(got, "Details") {
		t.Error("missing section")
	}
	if !strings.Contains(got, "Some detail") {
		t.Error("missing detail line")
	}
}
