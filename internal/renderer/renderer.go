package renderer

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/Somehow007/txt2md/internal/scanner"
)

type Options struct {
	Pretty bool
	Style  string
}

func Render(blocks []scanner.Block, opts Options) string {
	var sb strings.Builder

	for i, block := range blocks {
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

	level := block.HeadingLevel
	if level <= 0 || level > 6 {
		level = 2
	}

	mdLevel := 0
	for _, r := range rawText {
		if r == '#' {
			mdLevel++
		} else {
			break
		}
	}

	var text string
	if mdLevel > 0 && mdLevel <= 6 && len(rawText) > mdLevel && rawText[mdLevel] == ' ' {
		text = strings.TrimSpace(rawText[mdLevel+1:])
		level = mdLevel
	} else if strings.HasPrefix(rawText, "// ") {
		text = strings.TrimSpace(strings.TrimPrefix(rawText, "// "))
	} else if strings.HasPrefix(rawText, "/// ") {
		text = strings.TrimSpace(strings.TrimPrefix(rawText, "/// "))
	} else {
		text = rawText
	}

	sb.WriteString(fmt.Sprintf("%s %s", strings.Repeat("#", level), text))
}

func renderParagraph(sb *strings.Builder, block scanner.Block, opts Options) {
	hasMarkdown := false
	for _, line := range block.Lines {
		if line.IsMarkdown {
			hasMarkdown = true
			break
		}
	}

	if hasMarkdown {
		for i, line := range block.Lines {
			text := strings.TrimRight(line.Raw, " \t\r")
			if i > 0 {
				sb.WriteString("\n")
			}
			sb.WriteString(text)
		}
		return
	}

	parts := make([]string, 0, len(block.Lines))
	for _, line := range block.Lines {
		text := strings.TrimSpace(line.Raw)
		if text != "" {
			parts = append(parts, text)
		}
	}
	text := strings.Join(parts, " ")

	text = linkifyURLs(text)

	if opts.Pretty {
		text = Beautify(text)
	}

	sb.WriteString(text)
}

func renderList(sb *strings.Builder, block scanner.Block, opts Options) {
	minIndent := -1
	for _, line := range block.Lines {
		if !line.IsEmpty && line.Indent >= 0 {
			if minIndent < 0 || line.Indent < minIndent {
				minIndent = line.Indent
			}
		}
	}
	if minIndent < 0 {
		minIndent = 0
	}

	type listItem struct {
		marker  string
		content string
	}

	var items []listItem
	var current *listItem

	for _, line := range block.Lines {
		if line.IsEmpty {
			continue
		}

		trimmed := strings.TrimSpace(line.Raw)
		if trimmed == "" {
			continue
		}

		relativeIndent := line.Indent - minIndent

		if isListItemText(trimmed) && relativeIndent == 0 {
			normalized := normalizeListMarker(trimmed)
			markerEnd := indexOfListMarkerEnd(normalized)
			marker := strings.TrimSpace(normalized[:markerEnd])
			content := strings.TrimSpace(normalized[markerEnd:])
			item := listItem{marker: marker, content: content}
			items = append(items, item)
			current = &items[len(items)-1]
		} else if current != nil {
			if current.content != "" {
				current.content += " " + trimmed
			} else {
				current.content = trimmed
			}
		}
	}

	for i, item := range items {
		sb.WriteString(item.marker)
		if item.content != "" {
			sb.WriteString(" ")
			sb.WriteString(item.content)
		}
		if i < len(items)-1 {
			sb.WriteString("\n")
		}
	}
}

func isListItemText(text string) bool {
	if len(text) < 2 {
		return false
	}
	first := rune(text[0])
	if (first == '-' || first == '*' || first == '+' || first == '•') && len(text) > 1 && (text[1] == ' ' || text[1] == '\t') {
		return true
	}
	if len(text) >= 3 {
		for i := 0; i < len(text) && i < 5; i++ {
			if text[i] >= '0' && text[i] <= '9' {
				continue
			}
			if i > 0 && (text[i] == '.' || text[i] == ')') && i+1 < len(text) && (text[i+1] == ' ' || text[i+1] == '\t') {
				return true
			}
			break
		}
	}
	return false
}

func indexOfListMarkerEnd(text string) int {
	if len(text) < 2 {
		return 0
	}
	first := rune(text[0])
	if (first == '-' || first == '*' || first == '+' || first == '•') && len(text) > 1 && (text[1] == ' ' || text[1] == '\t') {
		idx := 1
		for idx < len(text) && (text[idx] == ' ' || text[idx] == '\t') {
			idx++
		}
		return idx
	}
	for i := 0; i < len(text) && i < 5; i++ {
		if text[i] >= '0' && text[i] <= '9' {
			continue
		}
		if i > 0 && (text[i] == '.' || text[i] == ')') && i+1 < len(text) && (text[i+1] == ' ' || text[i+1] == '\t') {
			idx := i + 1
			for idx < len(text) && (text[idx] == ' ' || text[idx] == '\t') {
				idx++
			}
			return idx
		}
		break
	}
	return 0
}

func renderCodeBlock(sb *strings.Builder, block scanner.Block) {
	lang := ""
	hasOpenFence := false

	if len(block.Lines) > 0 {
		first := strings.TrimSpace(block.Lines[0].Raw)
		if strings.HasPrefix(first, "```") {
			lang = strings.TrimPrefix(first, "```")
			lang = strings.TrimSpace(lang)
			hasOpenFence = true
		} else if strings.HasPrefix(first, "~~~") {
			lang = strings.TrimPrefix(first, "~~~")
			lang = strings.TrimSpace(lang)
			hasOpenFence = true
		}
	}

	sb.WriteString("```")
	if lang != "" {
		sb.WriteString(lang)
	}
	sb.WriteString("\n")

	var contentLines []string
	for _, line := range block.Lines {
		trimmed := strings.TrimSpace(line.Raw)
		if strings.HasPrefix(trimmed, "```") || strings.HasPrefix(trimmed, "~~~") {
			continue
		}
		contentLines = append(contentLines, strings.TrimRight(line.Raw, " \t\r"))
	}

	if !hasOpenFence && len(contentLines) > 0 {
		minIndent := -1
		for _, line := range contentLines {
			if strings.TrimSpace(line) == "" {
				continue
			}
			indent := len(line) - len(strings.TrimLeft(line, " \t"))
			if minIndent < 0 || indent < minIndent {
				minIndent = indent
			}
		}
		if minIndent > 0 {
			for i, line := range contentLines {
				if strings.TrimSpace(line) == "" {
					contentLines[i] = ""
				} else if len(line) > minIndent {
					contentLines[i] = line[minIndent:]
				}
			}
		}
	}

	for _, line := range contentLines {
		sb.WriteString(line)
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

		depth := 0
		remaining := text
		for len(remaining) > 0 && remaining[0] == '>' {
			depth++
			remaining = remaining[1:]
			if len(remaining) > 0 && remaining[0] == ' ' {
				remaining = remaining[1:]
			}
		}

		if depth > 0 {
			prefix := strings.Repeat("> ", depth)
			sb.WriteString(prefix)
			sb.WriteString(remaining)
		} else {
			sb.WriteString("> ")
			sb.WriteString(remaining)
		}
		sb.WriteString("\n")
	}
}

func renderTable(sb *strings.Builder, block scanner.Block) {
	if block.TableData != nil && len(block.TableData.Rows) > 0 {
		renderTableFromData(sb, block.TableData.Rows)
		return
	}

	if len(block.Lines) < 2 {
		return
	}

	var rows [][]string
	for _, line := range block.Lines {
		text := strings.TrimSpace(line.Raw)
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

	renderTableFromData(sb, rows)
}

func renderTableFromData(sb *strings.Builder, rows [][]string) {
	if len(rows) == 0 {
		return
	}

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

	colWidths := make([]int, maxCols)
	for _, row := range rows {
		for j, cell := range row {
			cellLen := runeWidth(cell)
			if cellLen > colWidths[j] {
				colWidths[j] = cellLen
			}
		}
	}

	header := rows[0]
	cells := make([]string, maxCols)
	for j, cell := range header {
		cells[j] = padCell(cell, colWidths[j])
	}
	sb.WriteString("| " + strings.Join(cells, " | ") + " |")
	sb.WriteString("\n")

	sb.WriteString("|")
	for j := 0; j < maxCols; j++ {
		sb.WriteString(strings.Repeat("-", colWidths[j]+2))
		sb.WriteString("|")
	}
	sb.WriteString("\n")

	for i := 1; i < len(rows); i++ {
		cells := make([]string, maxCols)
		for j, cell := range rows[i] {
			cells[j] = padCell(cell, colWidths[j])
		}
		sb.WriteString("| " + strings.Join(cells, " | ") + " |")
		if i < len(rows)-1 {
			sb.WriteString("\n")
		}
	}
}

func padCell(cell string, width int) string {
	cellLen := runeWidth(cell)
	if cellLen >= width {
		return cell
	}
	return cell + strings.Repeat(" ", width-cellLen)
}

func runeWidth(s string) int {
	width := 0
	for _, r := range s {
		if r >= 0x4E00 && r <= 0x9FFF {
			width += 2
		} else if r >= 0x3040 && r <= 0x30FF {
			width += 2
		} else if r >= 0xAC00 && r <= 0xD7AF {
			width += 2
		} else if r >= 0xFF01 && r <= 0xFF60 {
			width += 2
		} else {
			width++
		}
	}
	return width
}

func renderHorizontal(sb *strings.Builder) {
	sb.WriteString("---")
}

func parseTableRow(raw string) []string {
	if strings.Contains(raw, "│") {
		return parseUnicodeTableRow(raw)
	}
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

	raw = strings.ReplaceAll(raw, "\t", "  ")

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

func isSeparatorLine(text string) bool {
	trimmed := strings.TrimSpace(text)

	if hasBoxDrawing(trimmed) {
		if strings.ContainsAny(trimmed, "─━") {
			return true
		}
	}

	if strings.Contains(trimmed, "|") {
		parts := strings.Split(trimmed, "|")
		for _, p := range parts {
			if !regexp.MustCompile(`^[\s\-|:*_]+$`).MatchString(p) {
				return false
			}
		}
		return true
	}

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

func hasBoxDrawing(s string) bool {
	for _, r := range s {
		if r >= 0x2500 && r <= 0x257F {
			return true
		}
	}
	return false
}

var (
	urlPattern       = regexp.MustCompile(`(https?://[^\s<)\]]+)`)
	wwwPattern       = regexp.MustCompile(`\b(www\.[^\s<)\]]+)`)
	mdExistingLinkRe = regexp.MustCompile(`\[[^\]]*\]\([^\)]*\)`)
)

func linkifyURLs(text string) string {
	if mdExistingLinkRe.MatchString(text) {
		return text
	}

	text = urlPattern.ReplaceAllStringFunc(text, func(match string) string {
		return fmt.Sprintf("[%s](%s)", match, match)
	})

	text = wwwPattern.ReplaceAllStringFunc(text, func(match string) string {
		return fmt.Sprintf("[%s](https://%s)", match, match)
	})

	return text
}

func normalizeListMarker(text string) string {
	prefixes := []string{"* ", "+ ", "• ", "· "}
	for _, p := range prefixes {
		if strings.HasPrefix(text, p) {
			return "- " + strings.TrimPrefix(text, p)
		}
	}
	return text
}
