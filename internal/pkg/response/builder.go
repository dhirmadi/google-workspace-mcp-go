package response

import (
	"fmt"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// Builder constructs formatted text responses for MCP tool results.
// Provides a consistent output format across all tools.
type Builder struct {
	sb strings.Builder
}

// New creates a new response Builder.
func New() *Builder {
	return &Builder{}
}

// Header writes a header line with optional formatting arguments.
func (b *Builder) Header(format string, args ...any) *Builder {
	text := fmt.Sprintf(format, args...)
	b.sb.WriteString("═══ ")
	b.sb.WriteString(text)
	b.sb.WriteString(" ═══\n")
	return b
}

// KeyValue writes a key-value pair.
func (b *Builder) KeyValue(key string, value any) *Builder {
	b.sb.WriteString(fmt.Sprintf("• %s: %v\n", key, value))
	return b
}

// Item writes a bulleted item with optional formatting arguments.
func (b *Builder) Item(format string, args ...any) *Builder {
	text := fmt.Sprintf(format, args...)
	b.sb.WriteString("  → ")
	b.sb.WriteString(text)
	b.sb.WriteByte('\n')
	return b
}

// Line writes a plain line with optional formatting arguments.
func (b *Builder) Line(format string, args ...any) *Builder {
	b.sb.WriteString(fmt.Sprintf(format, args...))
	b.sb.WriteByte('\n')
	return b
}

// Blank writes an empty line.
func (b *Builder) Blank() *Builder {
	b.sb.WriteByte('\n')
	return b
}

// Separator writes a visual separator line.
func (b *Builder) Separator() *Builder {
	b.sb.WriteString("───────────────────────────────\n")
	return b
}

// Section writes a section header (smaller than Header).
func (b *Builder) Section(format string, args ...any) *Builder {
	text := fmt.Sprintf(format, args...)
	b.sb.WriteString("── ")
	b.sb.WriteString(text)
	b.sb.WriteString(" ──\n")
	return b
}

// Raw writes raw text without any formatting.
func (b *Builder) Raw(text string) *Builder {
	b.sb.WriteString(text)
	return b
}

// Build returns the assembled string.
func (b *Builder) Build() string {
	return b.sb.String()
}

// TextResult constructs an MCP CallToolResult from the builder's text content.
// This is the standard return pattern for all tool handlers.
func (b *Builder) TextResult() *mcp.CallToolResult {
	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: b.sb.String()}},
	}
}
