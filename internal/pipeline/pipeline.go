package pipeline

import (
	"strings"

	"github.com/Somehow007/txt2md/internal/classifier"
	"github.com/Somehow007/txt2md/internal/classifier/rules"
	"github.com/Somehow007/txt2md/internal/diff"
	"github.com/Somehow007/txt2md/internal/renderer"
	"github.com/Somehow007/txt2md/internal/scanner"
)

type Options struct {
	TabWidth int
	Pretty   bool
	Style    string
	Diff     bool
}

type Result struct {
	Output  string
	Diff    *diff.Result
	HasDiff bool
}

func Convert(input string, opts Options) (string, error) {
	result, err := ConvertWithDiff(input, opts)
	if err != nil {
		return "", err
	}
	return result.Output, nil
}

func ConvertWithDiff(input string, opts Options) (*Result, error) {
	lines := scanner.Scan(input)

	lines = preprocessLines(lines)

	mdRatio := calcMarkdownRatio(lines)

	if mdRatio > 0.4 {
		output := preserveMarkdown(lines)
		result := &Result{Output: output, HasDiff: false}
		if opts.Diff {
			d := diff.Compare(input, output)
			result.Diff = d
			result.HasDiff = d.Summary.AddedLines > 0 || d.Summary.RemovedLines > 0 || d.Summary.ChangedLines > 0
		}
		return result, nil
	}

	engine := classifier.NewEngine(
		rules.Options{TabWidth: opts.TabWidth},
		&rules.HeadingRule{},
		&rules.ListRule{},
		&rules.CodeBlockRule{},
		&rules.BlockquoteRule{},
		&rules.TableRule{},
		&rules.HorizontalRule{},
		&rules.ParagraphRule{},
	)
	blocks := engine.Classify(lines)

	output := renderer.Render(blocks, renderer.Options{
		Pretty: opts.Pretty,
		Style:  opts.Style,
	})

	result := &Result{
		Output:  output,
		HasDiff: false,
	}

	if opts.Diff {
		d := diff.Compare(input, output)
		result.Diff = d
		result.HasDiff = d.Summary.AddedLines > 0 || d.Summary.RemovedLines > 0 || d.Summary.ChangedLines > 0
	}

	return result, nil
}

func calcMarkdownRatio(lines []scanner.Line) float64 {
	if len(lines) == 0 {
		return 0
	}
	blockMdCount := 0
	contentCount := 0
	for _, line := range lines {
		if line.IsEmpty {
			continue
		}
		contentCount++
		trimmed := strings.TrimSpace(line.Raw)
		if isBlockLevelMarkdown(trimmed) {
			blockMdCount++
		}
	}
	if contentCount == 0 {
		return 0
	}
	return float64(blockMdCount) / float64(contentCount)
}

func isBlockLevelMarkdown(text string) bool {
	if len(text) == 0 {
		return false
	}
	if strings.HasPrefix(text, "#") {
		return true
	}
	if strings.HasPrefix(text, "```") || strings.HasPrefix(text, "~~~") {
		return true
	}
	if strings.HasPrefix(text, "> ") {
		return true
	}
	if len(text) >= 2 {
		first := rune(text[0])
		if (first == '-' || first == '*' || first == '+') && (text[1] == ' ' || text[1] == '\t') {
			return true
		}
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

func preserveMarkdown(lines []scanner.Line) string {
	var sb strings.Builder
	inCodeBlock := false
	var codeBlockLines []string

	for i, line := range lines {
		text := line.Raw
		trimmed := strings.TrimSpace(text)

		if trimmed == "" {
			if inCodeBlock {
				codeBlockLines = append(codeBlockLines, text)
			} else {
				sb.WriteString("\n")
			}
			continue
		}

		if strings.HasPrefix(trimmed, "```") || strings.HasPrefix(trimmed, "~~~") {
			if inCodeBlock {
				flushCodeBlockLines(&sb, codeBlockLines)
				codeBlockLines = nil
				inCodeBlock = false
				sb.WriteString(trimmed)
				sb.WriteString("\n")
			} else {
				inCodeBlock = true
				if i > 0 {
					sb.WriteString("\n")
				}
				sb.WriteString(trimmed)
				sb.WriteString("\n")
				codeBlockLines = nil
			}
			continue
		}

		if inCodeBlock {
			codeBlockLines = append(codeBlockLines, text)
			continue
		}

		cleaned := normalizeMarkdownLine(trimmed)

		if i > 0 {
			sb.WriteString("\n")
		}
		sb.WriteString(cleaned)
	}

	if inCodeBlock && len(codeBlockLines) > 0 {
		flushCodeBlockLines(&sb, codeBlockLines)
	}

	output := sb.String()

	output = fixUnclosedCodeBlocks(output)

	output = strings.TrimRight(output, "\n\r")

	return output
}

func flushCodeBlockLines(sb *strings.Builder, lines []string) {
	if len(lines) == 0 {
		return
	}

	minIndent := -1
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		indent := len(line) - len(strings.TrimLeft(line, " \t"))
		if minIndent < 0 || indent < minIndent {
			minIndent = indent
		}
	}

	if minIndent <= 0 {
		for _, line := range lines {
			sb.WriteString(strings.TrimRight(line, " \t\r"))
			sb.WriteString("\n")
		}
		return
	}

	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			sb.WriteString("\n")
			continue
		}
		if len(line) > minIndent {
			sb.WriteString(line[minIndent:])
		} else {
			sb.WriteString(strings.TrimRight(line, " \t\r"))
		}
		sb.WriteString("\n")
	}
}

func normalizeMarkdownLine(text string) string {
	if strings.HasPrefix(text, "---") && len(strings.TrimLeft(text, "-")) == 0 && len(text) >= 3 {
		return "---"
	}
	if strings.HasPrefix(text, "***") && len(strings.TrimLeft(text, "*")) == 0 && len(text) >= 3 {
		return "---"
	}
	if strings.HasPrefix(text, "___") && len(strings.TrimLeft(text, "_")) == 0 && len(text) >= 3 {
		return "---"
	}
	if strings.HasPrefix(text, "===") && len(strings.TrimLeft(text, "=")) == 0 && len(text) >= 3 {
		return "---"
	}

	if isUnorderedListItem(text) {
		markerEnd := findMarkerEnd(text)
		marker := strings.TrimSpace(text[:markerEnd])
		content := strings.TrimSpace(text[markerEnd:])
		if marker == "*" || marker == "+" || marker == "•" || marker == "·" {
			return "- " + content
		}
		return "- " + content
	}

	return text
}

func isUnorderedListItem(text string) bool {
	if len(text) < 2 {
		return false
	}
	first := rune(text[0])
	if (first == '-' || first == '*' || first == '+' || first == '•' || first == '·') && len(text) > 1 && (text[1] == ' ' || text[1] == '\t') {
		return true
	}
	return false
}

func findMarkerEnd(text string) int {
	idx := 1
	for idx < len(text) && (text[idx] == ' ' || text[idx] == '\t') {
		idx++
	}
	return idx
}

func fixUnclosedCodeBlocks(text string) string {
	lines := strings.Split(text, "\n")
	inBlock := false
	fence := ""
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "```") {
			if !inBlock {
				inBlock = true
				fence = "```"
			} else {
				inBlock = false
				fence = ""
			}
			continue
		}
		if strings.HasPrefix(trimmed, "~~~") {
			if !inBlock {
				inBlock = true
				fence = "~~~"
			} else {
				inBlock = false
				fence = ""
			}
			continue
		}
	}

	if inBlock {
		text += "\n" + fence
	}

	return text
}

func preprocessLines(lines []scanner.Line) []scanner.Line {
	result := make([]scanner.Line, 0, len(lines))
	for i, line := range lines {
		if isOrphanNumber(line, lines, i) {
			continue
		}
		result = append(result, line)
	}
	return result
}

func isOrphanNumber(line scanner.Line, lines []scanner.Line, idx int) bool {
	if line.IsEmpty {
		return false
	}

	trimmed := strings.TrimSpace(line.Raw)
	if len(trimmed) == 0 {
		return false
	}

	if !isOnlyDigits(trimmed) {
		return false
	}

	hasBlankBefore := idx == 0 || lines[idx-1].IsEmpty
	hasBlankAfter := idx+1 >= len(lines) || lines[idx+1].IsEmpty

	if hasBlankBefore && hasBlankAfter {
		return true
	}

	return false
}

func isOnlyDigits(text string) bool {
	for _, r := range text {
		if r < '0' || r > '9' {
			return false
		}
	}
	return len(text) > 0 && len(text) <= 3
}
