package renderer

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/Somehow007/txt2md/internal/scanner"
)

// Options configures rendering behavior.
type Options struct {
	Pretty bool
	Style  string // "compact" or "spacious"
}

// Render converts classified blocks to Markdown text.
func Render(blocks []scanner.Block, opts Options) string {
	var sb strings.Builder

	for i, block := range blocks {
		// Add blank line between blocks based on style
		if i > 0 {
			if opts.Style == "compact" {
				sb.WriteString("\n")
			} else {
				sb.WriteString("\n\n")
			}
		}

		switch block.Type {
		case scanner.BlockHeading:
			renderHeading(&sb, block)
		case scanner.BlockParagraph:
			renderParagraph(&sb, block, opts)
		case scanner.BlockList:
			renderList(&sb, block, opts)
		case scanner.BlockCode:
			renderCodeBlock(&sb, block)
		case scanner.BlockQuote:
			renderBlockquote(&sb, block)
		case scanner.BlockTable:
			renderTable(&sb, block)
		case scanner.BlockHorizontal:
			renderHorizontal(&sb)
		default:
			// Fallback: render as paragraph
			renderParagraph(&sb, block, opts)
		}
	}

	return sb.String()
}

func renderHeading(sb *strings.Builder, block scanner.Block) {
	if len(block.Lines) == 0 {
		return
	}
	rawText := strings.TrimSpace(block.Lines[0].Raw)

	// Detect heading level from original text (count leading # characters)
	level := 0
	for _, r := range rawText {
		if r == '#' {
			level++
		} else {
			break
		}
	}

	// If markdown-style heading (## Title), use detected level
	// Otherwise, clean the text and use default level
	var text string
	if level > 0 && level <= 6 && len(rawText) > level && rawText[level] == ' ' {
		// Markdown-style heading: use detected level
		text = strings.TrimSpace(rawText[level+1:])
	} else {
		// Not markdown-style, clean prefixes
		text = rawText
		text = strings.TrimPrefix(text, "# ")
		text = strings.TrimPrefix(text, "## ")
		text = strings.TrimPrefix(text, "### ")
		text = strings.TrimPrefix(text, "// ")
		text = strings.TrimPrefix(text, "/// ")
		text = strings.TrimSpace(text)

		// Default level based on confidence
		level = 2
		if block.Confidence > 0.8 {
			level = 1
		}
	}

	sb.WriteString(fmt.Sprintf("%s %s", strings.Repeat("#", level), text))
}

func renderParagraph(sb *strings.Builder, block scanner.Block, opts Options) {
	parts := make([]string, 0, len(block.Lines))
	for _, line := range block.Lines {
		text := strings.TrimSpace(line.Raw)
		if text != "" {
			parts = append(parts, text)
		}
	}
	text := strings.Join(parts, " ")

	// URL auto-detection
	text = linkifyURLs(text)

	// Pretty typography
	if opts.Pretty {
		text = Beautify(text)
	}

	sb.WriteString(text)
}

func renderList(sb *strings.Builder, block scanner.Block, opts Options) {
	for i, line := range block.Lines {
		text := strings.TrimSpace(line.Raw) // Normalize indentation

		if text == "" {
			continue
		}

		sb.WriteString(text)
		// Add newline after each line except the last
		if i < len(block.Lines)-1 {
			sb.WriteString("\n")
		}
	}
}

func renderCodeBlock(sb *strings.Builder, block scanner.Block) {
	// Check if original had a language identifier
	lang := ""
	if len(block.Lines) > 0 {
		first := strings.TrimSpace(block.Lines[0].Raw)
		if strings.HasPrefix(first, "```") {
			lang = strings.TrimPrefix(first, "```")
			lang = strings.TrimSpace(lang)
		}
	}

	sb.WriteString("```")
	if lang != "" {
		sb.WriteString(lang)
	}
	sb.WriteString("\n")

	for _, line := range block.Lines {
		text := line.Raw
		// Skip original fence lines
		if strings.HasPrefix(strings.TrimSpace(text), "```") {
			continue
		}
		sb.WriteString(text)
		sb.WriteString("\n")
	}
	sb.WriteString("```")
}

func renderBlockquote(sb *strings.Builder, block scanner.Block) {
	for _, line := range block.Lines {
		text := strings.TrimSpace(line.Raw)
		if text == "" {
			sb.WriteString(">\n")
			continue
		}
		if strings.HasPrefix(text, ">") {
			text = strings.TrimPrefix(text, ">")
			text = strings.TrimPrefix(text, " ")
		}
		sb.WriteString("> ")
		sb.WriteString(text)
		sb.WriteString("\n")
	}
}

func renderTable(sb *strings.Builder, block scanner.Block) {
	// If TableData is populated (new path), use it directly
	if block.TableData != nil && len(block.TableData.Rows) > 0 {
		renderTableFromData(sb, block.TableData.Rows)
		return
	}

	// Fallback: original line-based parsing (for backward compatibility)
	if len(block.Lines) < 2 {
		return
	}

	// Parse rows
	var rows [][]string
	for _, line := range block.Lines {
		text := strings.TrimSpace(line.Raw)
		// Skip separator rows in output
		if isSeparatorLine(text) {
			continue
		}
		row := parseTableRow(text)
		if len(row) > 0 {
			rows = append(rows, row)
		}
	}

	if len(rows) == 0 {
		return
	}

	// Normalize columns
	// Determine max columns
	maxCols := 0
	for _, row := range rows {
		if len(row) > maxCols {
			maxCols = len(row)
		}
	}

	// Normalize all rows to maxCols
	for i := range rows {
		for len(rows[i]) < maxCols {
			rows[i] = append(rows[i], "")
		}
	}

	renderTableFromData(sb, rows)
}

// renderTableFromData renders table rows that have already been parsed.
func renderTableFromData(sb *strings.Builder, rows [][]string) {
	if len(rows) == 0 {
		return
	}

	// Normalize columns
	maxCols := 0
	for _, row := range rows {
		if len(row) > maxCols {
			maxCols = len(row)
		}
	}
	for i := range rows {
		for len(rows[i]) < maxCols {
			rows[i] = append(rows[i], "")
		}
	}

	// Render header row
	header := rows[0]
	sb.WriteString("| ")
	sb.WriteString(strings.Join(header, " | "))
	sb.WriteString(" |")
	sb.WriteString("\n")

	// Render separator row
	sb.WriteString("|")
	for i := 0; i < maxCols; i++ {
		sb.WriteString(" --- |")
	}
	sb.WriteString("\n")

	// Render data rows
	for i := 1; i < len(rows); i++ {
		sb.WriteString("| ")
		sb.WriteString(strings.Join(rows[i], " | "))
		sb.WriteString(" |")
		if i < len(rows)-1 {
			sb.WriteString("\n")
		}
	}
}

func renderHorizontal(sb *strings.Builder) {
	sb.WriteString("---")
}

func parseTableRow(raw string) []string {
	// Handle Unicode box-drawing vertical line character
	if strings.Contains(raw, "│") {
		return parseUnicodeTableRow(raw)
	}
	// Split by '|' if present
	if strings.Contains(raw, "|") {
		parts := strings.Split(raw, "|")
		cols := make([]string, 0, len(parts))
		for _, p := range parts {
			trimmed := strings.TrimSpace(p)
			if trimmed != "" {
				cols = append(cols, trimmed)
			}
		}
		return cols
	}

	// Split by tab first
	raw = strings.ReplaceAll(raw, "\t", "  ")

	// Split by 2+ consecutive spaces
	spaceRe := regexp.MustCompile(` {2,}`)
	parts := spaceRe.Split(raw, -1)

	cols := make([]string, 0, len(parts))
	for _, p := range parts {
		trimmed := strings.TrimSpace(p)
		if trimmed != "" {
			cols = append(cols, trimmed)
		}
	}

	return cols
}

// parseUnicodeTableRow splits a line by Unicode box-drawing vertical line │.
func parseUnicodeTableRow(raw string) []string {
	parts := strings.Split(raw, "│")
	cols := make([]string, 0, len(parts))
	for _, p := range parts {
		trimmed := strings.TrimSpace(p)
		if trimmed != "" {
			cols = append(cols, trimmed)
		}
	}
	return cols
}

// isSeparatorLine checks if the line is a table separator like ---|---|--- or ---   ---   ---
// Also handles Unicode box-drawing separators like ├─┼─┤.
func isSeparatorLine(text string) bool {
	trimmed := strings.TrimSpace(text)

	// Check for Unicode box-drawing separator (├─┼─┤, ┝━┿━┥, etc.)
	if hasBoxDrawing(trimmed) {
		// Must contain dash characters (─, ━, etc.)
		if strings.ContainsAny(trimmed, "─━") {
			return true
		}
	}

	// Check for | style separator: ---|---|---
	if strings.Contains(trimmed, "|") {
		parts := strings.Split(trimmed, "|")
		for _, p := range parts {
			if !regexp.MustCompile(`^[\s\-|:*_]+$`).MatchString(p) {
				return false
			}
		}
		return true
	}

	// Check for space-separated separator: ---   ---   ---
	// Split by 2+ spaces
	spaceRe := regexp.MustCompile(` {2,}`)
	parts := spaceRe.Split(trimmed, -1)
	if len(parts) >= 2 {
		for _, p := range parts {
			p = strings.TrimSpace(p)
			if !regexp.MustCompile(`^[-*_]+$`).MatchString(p) {
				return false
			}
		}
		return true
	}

	return false
}

// hasBoxDrawing checks if the string contains Unicode box-drawing characters (U+2500-U+257F).
func hasBoxDrawing(s string) bool {
	for _, r := range s {
		if r >= 0x2500 && r <= 0x257F {
			return true
		}
	}
	return false
}

// URL regex patterns
var (
	urlPattern = regexp.MustCompile(`(https?://[^\s<]+)`)
	wwwPattern = regexp.MustCompile(`\b(www\.[^\s<]+)`)
)

// linkifyURLs converts plain URLs to Markdown links.
func linkifyURLs(text string) string {
	// Handle https?:// URLs
	text = urlPattern.ReplaceAllStringFunc(text, func(match string) string {
		// Don't linkify if already in markdown link format
		if strings.Contains(text, "]("+match+")") {
			return match
		}
		return fmt.Sprintf("[%s](%s)", match, match)
	})

	// Handle www. URLs
	text = wwwPattern.ReplaceAllStringFunc(text, func(match string) string {
		// Don't linkify if already in markdown link format
		return fmt.Sprintf("[%s](https://%s)", match, match)
	})

	return text
}

func normalizeListMarker(text string) string {
	// Convert various list markers to standard "-"
	prefixes := []string{"* ", "+ ", "• ", "· "}
	for _, p := range prefixes {
		if strings.HasPrefix(text, p) {
			return "- " + strings.TrimPrefix(text, p)
		}
	}
	return text
}

// beautifyTypography is deprecated, use Beautify from formatter.go instead.
func beautifyTypography(text string) string {
	return text
}
