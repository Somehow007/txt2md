package rules

import (
	"strings"

	"github.com/Somehow007/txt2md/internal/scanner"
)

// CodeBlockRule detects code blocks based on code-like patterns.
type CodeBlockRule struct{}

func (r *CodeBlockRule) Name() string {
	return "codeblock"
}

func (r *CodeBlockRule) Detect(lines []scanner.Line, idx int, opts Options) (*scanner.Block, int) {
	line := lines[idx]
	text := strings.TrimSpace(line.Raw)

	if text == "" {
		return nil, 0
	}

	// Check for fenced code blocks
	if strings.HasPrefix(text, "```") {
		// Find closing fence
		consumed := 1
		for idx+consumed < len(lines) {
			if strings.HasPrefix(strings.TrimSpace(lines[idx+consumed].Raw), "```") {
				consumed++
				break
			}
			consumed++
		}
		codeLines := lines[idx : idx+consumed]
		return &scanner.Block{
			Type:       scanner.BlockCode,
			Lines:      codeLines,
			Confidence: 1.0,
		}, consumed
	}

	// Detect indented code block (4+ spaces or tab)
	if line.Indent >= 4 || strings.HasPrefix(line.Raw, "\t") {
		codeLines := []scanner.Line{line}
		consumed := 1
		for idx+consumed < len(lines) {
			next := lines[idx+consumed]
			if next.IsEmpty && idx+consumed+1 < len(lines) && lines[idx+consumed+1].Indent >= 4 {
				codeLines = append(codeLines, next)
				consumed++
				continue
			}
			if next.Indent >= 4 || strings.HasPrefix(next.Raw, "\t") {
				codeLines = append(codeLines, next)
				consumed++
			} else {
				break
			}
		}
		return &scanner.Block{
			Type:       scanner.BlockCode,
			Lines:      codeLines,
			Confidence: 0.85,
		}, consumed
	}

	// Detect code-like content (high density of special chars, keywords)
	if looksLikeCode(text) && idx+1 < len(lines) && !lines[idx+1].IsEmpty {
		// Check if next few lines also look like code
		codeLines := []scanner.Line{line}
		consumed := 1
		codeScore := 1
		for idx+consumed < len(lines) && consumed < 20 {
			next := lines[idx+consumed]
			if next.IsEmpty {
				break
			}
			if looksLikeCode(strings.TrimSpace(next.Raw)) {
				codeLines = append(codeLines, next)
				codeScore++
				consumed++
			} else {
				break
			}
		}
		if codeScore >= 2 {
			return &scanner.Block{
				Type:       scanner.BlockCode,
				Lines:      codeLines,
				Confidence: 0.7,
			}, consumed
		}
	}

	return nil, 0
}

func looksLikeCode(text string) bool {
	if len(text) == 0 {
		return false
	}

	// Count code-like characters
	codeChars := 0
	for _, c := range text {
		if c == '{' || c == '}' || c == '(' || c == ')' || c == '[' || c == ']' ||
			c == '=' || c == ';' || c == '<' || c == '>' || c == '/' || c == '*' ||
			c == '&' || c == '|' || c == '!' || c == '#' || c == '@' || c == '$' {
			codeChars++
		}
	}

	// High ratio of code characters
	ratio := float64(codeChars) / float64(len(text))
	if ratio > 0.15 && len(text) > 10 {
		return true
	}

	// Common programming keywords
	keywords := []string{"func ", "var ", "const ", "import ", "package ", "def ", "class ",
		"return ", "if ", "for ", "while ", "else ", "try ", "catch ", "public ", "private "}
	for _, kw := range keywords {
		if strings.Contains(text, kw) {
			return true
		}
	}

	return false
}
