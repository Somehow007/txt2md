package renderer

import (
	"fmt"
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
		if i > 0 {
			// Add blank line between blocks
			sb.WriteString("\n\n")
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
	text := strings.TrimSpace(block.Lines[0].Raw)
	// Remove existing markers
	text = strings.TrimPrefix(text, "# ")
	text = strings.TrimPrefix(text, "## ")
	text = strings.TrimPrefix(text, "### ")

	// Determine level based on confidence/context
	level := 2 // default
	if block.Confidence > 0.8 {
		level = 1
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

	if opts.Pretty {
		text = beautifyTypography(text)
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

func beautifyTypography(text string) string {
	// Basic Chinese/English spacing
	// This is a simplified version; full implementation in formatter.go
	return text
}
