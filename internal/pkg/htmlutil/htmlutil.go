package htmlutil

import (
	"regexp"
	"strconv"
	"strings"
)

// Various regex patterns for HTML-to-text conversion.
var (
	tagRE         = regexp.MustCompile(`<[^>]+>`)
	brRE          = regexp.MustCompile(`(?i)<br\s*/?>`)
	blockRE       = regexp.MustCompile(`(?i)</?(p|div|h[1-6]|li|tr|blockquote)\b[^>]*>`)
	styleScriptRE = regexp.MustCompile(`(?is)<style[^>]*>.*?</style>`)
	scriptRE      = regexp.MustCompile(`(?is)<script[^>]*>.*?</script>`)
	entityRE      = regexp.MustCompile(`&[a-zA-Z]+;|&#\d+;|&#x[0-9a-fA-F]+;`)
	whitespaceRE  = regexp.MustCompile(`[ \t]+`)
	blankLineRE   = regexp.MustCompile(`\n{3,}`)
)

// HTML entity map for common named entities.
var entityMap = map[string]string{
	"&amp;":    "&",
	"&lt;":     "<",
	"&gt;":     ">",
	"&quot;":   `"`,
	"&apos;":   "'",
	"&nbsp;":   " ",
	"&mdash;":  "—",
	"&ndash;":  "–",
	"&hellip;": "…",
	"&laquo;":  "«",
	"&raquo;":  "»",
	"&copy;":   "©",
	"&reg;":    "®",
	"&trade;":  "™",
}

// ToPlainText converts HTML to readable plain text.
// It handles common HTML elements, entities, and formatting.
func ToPlainText(html string) string {
	if html == "" {
		return ""
	}

	// Remove style and script blocks entirely
	text := styleScriptRE.ReplaceAllString(html, "")
	text = scriptRE.ReplaceAllString(text, "")

	// Convert <br> to newlines
	text = brRE.ReplaceAllString(text, "\n")

	// Add newlines around block elements
	text = blockRE.ReplaceAllString(text, "\n")

	// Strip remaining HTML tags
	text = tagRE.ReplaceAllString(text, "")

	// Decode HTML entities
	text = decodeEntities(text)

	// Normalize whitespace: collapse runs of spaces/tabs within lines
	lines := strings.Split(text, "\n")
	for i, line := range lines {
		lines[i] = whitespaceRE.ReplaceAllString(strings.TrimSpace(line), " ")
	}
	text = strings.Join(lines, "\n")

	// Collapse 3+ consecutive newlines to 2
	text = blankLineRE.ReplaceAllString(text, "\n\n")

	return strings.TrimSpace(text)
}

// decodeEntities replaces HTML entities with their text equivalents.
// Supports named entities (e.g. &amp;) and numeric entities (&#123; and &#x1F600;).
func decodeEntities(text string) string {
	return entityRE.ReplaceAllStringFunc(text, func(entity string) string {
		if replacement, ok := entityMap[strings.ToLower(entity)]; ok {
			return replacement
		}
		// Decode numeric entities: &#123; (decimal) or &#x1F600; (hex)
		if strings.HasPrefix(entity, "&#") {
			numStr := entity[2 : len(entity)-1] // strip "&#" and ";"
			var codePoint int64
			var err error
			if strings.HasPrefix(numStr, "x") || strings.HasPrefix(numStr, "X") {
				codePoint, err = strconv.ParseInt(numStr[1:], 16, 32)
			} else {
				codePoint, err = strconv.ParseInt(numStr, 10, 32)
			}
			if err == nil && codePoint > 0 {
				return string(rune(codePoint))
			}
		}
		return entity
	})
}
