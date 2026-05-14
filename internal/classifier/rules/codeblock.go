package rules

import (
	"strings"

	"github.com/Somehow007/txt2md/internal/scanner"
)

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

	if strings.HasPrefix(text, "```") || strings.HasPrefix(text, "~~~") {
		fence := "```"
		if strings.HasPrefix(text, "~~~") {
			fence = "~~~"
		}

		consumed := 1
		foundClose := false
		for idx+consumed < len(lines) {
			nextTrimmed := strings.TrimSpace(lines[idx+consumed].Raw)
			if strings.HasPrefix(nextTrimmed, fence) {
				consumed++
				foundClose = true
				break
			}

			if !foundClose && lines[idx+consumed].IsEmpty {
				break
			}

			consumed++
		}

		return &scanner.Block{
			Type:       scanner.BlockCode,
			Lines:      lines[idx : idx+consumed],
			Confidence: 1.0,
		}, consumed
	}

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

	if looksLikeCode(text) && idx+1 < len(lines) {
		codeLines := []scanner.Line{line}
		consumed := 1
		codeScore := 1
		blankStreak := 0
		for idx+consumed < len(lines) && consumed < 40 {
			next := lines[idx+consumed]
			if next.IsEmpty {
				blankStreak++
				if blankStreak > 1 {
					break
				}
				if idx+consumed+1 < len(lines) {
					nextNextText := strings.TrimSpace(lines[idx+consumed+1].Raw)
					if looksLikeCode(nextNextText) || isCodeContinuation(nextNextText, line.Indent) {
						codeLines = append(codeLines, next)
						consumed++
						continue
					}
				}
				break
			}
			blankStreak = 0
			nextText := strings.TrimSpace(next.Raw)
			if looksLikeCode(nextText) {
				codeLines = append(codeLines, next)
				codeScore++
				consumed++
			} else if codeScore >= 2 && isCodeContinuation(nextText, line.Indent) {
				codeLines = append(codeLines, next)
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

	if isDiagramStart(text) {
		codeLines := []scanner.Line{line}
		consumed := 1
		for idx+consumed < len(lines) {
			next := lines[idx+consumed]
			if next.IsEmpty {
				break
			}
			nextText := strings.TrimSpace(next.Raw)
			if isDiagramLine(nextText) {
				codeLines = append(codeLines, next)
				consumed++
			} else {
				break
			}
		}
		if len(codeLines) >= 2 {
			return &scanner.Block{
				Type:       scanner.BlockCode,
				Lines:      codeLines,
				Confidence: 0.8,
			}, consumed
		}
	}

	if looksLikeDiagramHeader(text) && idx+1 < len(lines) {
		nextText := strings.TrimSpace(lines[idx+1].Raw)
		if isDiagramLine(nextText) {
			codeLines := []scanner.Line{line}
			consumed := 1
			for idx+consumed < len(lines) {
				next := lines[idx+consumed]
				if next.IsEmpty {
					break
				}
				nextText := strings.TrimSpace(next.Raw)
				if isDiagramLine(nextText) || looksLikeDiagramHeader(nextText) {
					codeLines = append(codeLines, next)
					consumed++
				} else {
					break
				}
			}
			if len(codeLines) >= 2 {
				return &scanner.Block{
					Type:       scanner.BlockCode,
					Lines:      codeLines,
					Confidence: 0.8,
				}, consumed
			}
		}
	}

	return nil, 0
}

func isCodeContinuation(text string, baseIndent int) bool {
	if len(text) == 0 {
		return false
	}
	if strings.HasSuffix(text, ",") || strings.HasSuffix(text, ";") ||
		strings.HasSuffix(text, "(") || strings.HasSuffix(text, "{") ||
		strings.HasSuffix(text, "[") || strings.HasSuffix(text, "||") ||
		strings.HasSuffix(text, "&&") || strings.HasSuffix(text, "+") ||
		strings.HasSuffix(text, "->") || strings.HasSuffix(text, "=>") {
		return true
	}
	if strings.HasPrefix(text, ",") || strings.HasPrefix(text, ")") ||
		strings.HasPrefix(text, "}") || strings.HasPrefix(text, "]") ||
		strings.HasPrefix(text, ".") || strings.HasPrefix(text, "::") {
		return true
	}
	return false
}

func isDiagramStart(text string) bool {
	if strings.Contains(text, "──>") || strings.Contains(text, "<──") ||
		strings.Contains(text, "─>") || strings.Contains(text, "<─") {
		if !strings.ContainsAny(text, "┌┐└┘├┤┬┴┼") {
			return true
		}
	}
	if strings.Contains(text, "│") && (strings.Contains(text, "──>") || strings.Contains(text, "<──")) {
		if !strings.ContainsAny(text, "┌┐└┘├┤┬┴┼") {
			return true
		}
	}
	return false
}

func isDiagramLine(text string) bool {
	if strings.Contains(text, "│") && !strings.ContainsAny(text, "┌┐└┘├┤┬┴┼") {
		return true
	}
	if isDiagramStart(text) {
		return true
	}
	return false
}

func looksLikeDiagramHeader(text string) bool {
	if len(text) < 10 {
		return false
	}
	for _, r := range text {
		if r >= 0x2500 && r <= 0x257F {
			return false
		}
	}
	maxSpaceRun := 0
	spaceRun := 0
	for _, c := range text {
		if c == ' ' {
			spaceRun++
			if spaceRun > maxSpaceRun {
				maxSpaceRun = spaceRun
			}
		} else {
			spaceRun = 0
		}
	}
	if maxSpaceRun >= 5 {
		parts := strings.Fields(text)
		if len(parts) >= 2 && len(parts) <= 6 {
			spaceCount := strings.Count(text, " ")
			if float64(spaceCount)/float64(len(text)) > 0.3 {
				return true
			}
		}
	}
	return false
}

func looksLikeCode(text string) bool {
	if len(text) == 0 {
		return false
	}

	trimmed := strings.TrimSpace(text)

	if strings.HasPrefix(trimmed, "//") || strings.HasPrefix(trimmed, "///") ||
		strings.HasPrefix(trimmed, "# ") || strings.HasPrefix(trimmed, "#!") {
		return true
	}

	if trimmed == "{" || trimmed == "}" || trimmed == "{}" ||
		trimmed == "(" || trimmed == ")" || trimmed == "()" ||
		trimmed == "[" || trimmed == "]" || trimmed == "[]" {
		return true
	}

	if trimmed == "}" || trimmed == "};" || trimmed == "}," {
		return true
	}

	codeChars := 0
	for _, c := range text {
		if c == '{' || c == '}' || c == '(' || c == ')' || c == '[' || c == ']' ||
			c == '=' || c == ';' || c == '<' || c == '>' || c == '/' || c == '*' ||
			c == '&' || c == '|' || c == '!' || c == '#' || c == '@' || c == '$' ||
			c == '"' || c == '\'' || c == ':' || c == '%' || c == '+' || c == '-' {
			codeChars++
		}
	}

	ratio := float64(codeChars) / float64(len(text))
	if ratio > 0.2 && len(text) > 5 {
		return true
	}
	if ratio > 0.1 && len(text) > 10 {
		return true
	}

	keywords := []string{"func ", "var ", "const ", "import ", "package ", "def ", "class ",
		"return ", "if ", "for ", "while ", "else ", "try ", "catch ", "public ", "private ",
		"void ", "int ", "String ", "boolean ", "@Override", "@Autowired", "@PostMapping",
		"new ", "this.", "extends ", "implements ", "@GetMapping", "@RequestBody",
		"redisTemplate.", "userService.", "JSON.", "UserHolder.", "Result.",
		"TimeUnit.", "StringRedisTemplate", "HttpServletRequest", "HttpServletResponse",
		"HandlerInterceptor", "Map.of(", "UUID.randomUUID",
	}
	for _, kw := range keywords {
		if strings.Contains(text, kw) {
			return true
		}
	}

	return false
}
