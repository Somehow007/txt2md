package rules

import (
	"strings"
	"unicode"

	"github.com/Somehow007/txt2md/internal/scanner"
)

type HeadingRule struct{}

func (r *HeadingRule) Name() string {
	return "heading"
}

func (r *HeadingRule) Detect(lines []scanner.Line, idx int, opts Options) (*scanner.Block, int) {
	line := lines[idx]
	text := strings.TrimSpace(line.Raw)

	if text == "" {
		return nil, 0
	}

	if strings.Contains(text, ".") && strings.Contains(text, " ") && len(text) > 50 {
		return nil, 0
	}

	if isLikelyCodeComment(text, lines, idx) {
		return nil, 0
	}

	if isListItem(text, line.Indent) {
		return nil, 0
	}

	if looksLikeShellCommand(text) {
		return nil, 0
	}

	if endsWithSentencePunctuation(text) {
		return nil, 0
	}

	isHeading := false
	confidence := 0.8
	level := 2

	if strings.HasPrefix(text, "#") {
		hashCount := 0
		for _, r := range text {
			if r == '#' {
				hashCount++
			} else {
				break
			}
		}
		if hashCount >= 1 && hashCount <= 6 && len(text) > hashCount && text[hashCount] == ' ' {
			isHeading = true
			confidence = 1.0
			level = hashCount
		}
	}

	if !isHeading && (strings.HasPrefix(text, "// ") || strings.HasPrefix(text, "/// ")) {
		if !isInCodeContext(lines, idx) && !isNearCodeContent(lines, idx) {
			isHeading = true
			confidence = 0.95
			if strings.HasPrefix(text, "/// ") {
				level = 3
			} else {
				level = 2
			}
		}
	}

	if !isHeading && text == strings.ToUpper(text) && len(text) > 3 && len(text) < 50 {
		if isMostlyUppercaseLatin(text) {
			isHeading = true
			level = 2
		}
	}

	if !isHeading && isNumberedHeading(text) {
		isHeading = true
		confidence = 0.9
		level = determineHeadingLevel(text, line.Indent)
	}

	if !isHeading && isStepPattern(text) {
		isHeading = true
		confidence = 0.85
		level = 3
	}

	if !isHeading && isNumberedSectionTitle(text) {
		isHeading = true
		confidence = 0.8
		level = 2
	}

	if !isHeading && isContextualHeading(text, lines, idx) {
		isHeading = true
		confidence = 0.75
		level = determineContextualHeadingLevel(text, line.Indent)
	}

	if !isHeading {
		return nil, 0
	}

	return &scanner.Block{
		Type:         scanner.BlockHeading,
		Lines:        []scanner.Line{line},
		Confidence:   confidence,
		HeadingLevel: level,
	}, 1
}

func endsWithSentencePunctuation(text string) bool {
	if len(text) == 0 {
		return false
	}
	runes := []rune(text)
	last := runes[len(runes)-1]
	return last == '。' || last == '！' || last == '？' ||
		last == '.' || last == '!' || last == '?'
}

func isContextualHeading(text string, lines []scanner.Line, idx int) bool {
	if len(text) < 2 || len(text) > 60 {
		return false
	}

	if isListItem(text, lines[idx].Indent) {
		return false
	}

	if !hasLetterOrChinese(text) {
		return false
	}

	if looksLikeCode(text) {
		return false
	}

	if isCommandLine(text) {
		return false
	}

	if looksLikeShellCommand(text) {
		return false
	}

	if strings.HasPrefix(text, "@") {
		return false
	}

	if endsWithSentencePunctuation(text) {
		return false
	}

	if isPartOfParagraph(lines, idx) {
		return false
	}

	if isOnlyDigits(text) {
		return false
	}

	hasBlankAfter := idx+1 < len(lines) && lines[idx+1].IsEmpty
	hasContentAfter := false
	for i := idx + 1; i < len(lines) && i <= idx+3; i++ {
		if !lines[i].IsEmpty {
			hasContentAfter = true
			break
		}
	}

	hasBlankBefore := idx == 0 || lines[idx-1].IsEmpty

	consecutiveBlanksAfter := 0
	for i := idx + 1; i < len(lines); i++ {
		if lines[i].IsEmpty {
			consecutiveBlanksAfter++
		} else {
			break
		}
	}

	if consecutiveBlanksAfter >= 2 {
		return false
	}

	if hasBlankBefore && hasBlankAfter && hasContentAfter {
		return true
	}

	return false
}

func isOnlyDigits(text string) bool {
	for _, r := range text {
		if !unicode.IsDigit(r) && r != ' ' && r != '.' {
			return false
		}
	}
	return len(text) <= 5
}

func isCommandLine(text string) bool {
	trimmed := strings.TrimSpace(text)
	if len(trimmed) == 0 {
		return false
	}

	commandPrefixes := []string{
		"txt2md ", "git ", "npm ", "go ", "python ", "pip ",
		"docker ", "curl ", "wget ", "make ", "cargo ",
		"brew ", "apt ", "yum ", "sudo ", "chmod ",
		"mv ", "cp ", "rm ", "ls ", "cat ", "echo ",
		"export ", "source ", "cd ",
	}
	for _, prefix := range commandPrefixes {
		if strings.HasPrefix(trimmed, prefix) {
			return true
		}
	}

	if strings.Contains(trimmed, " --") || strings.Contains(trimmed, " -") {
		if strings.Contains(trimmed, "=") || strings.Count(trimmed, " ") >= 2 {
			return true
		}
	}

	return false
}

func isPartOfParagraph(lines []scanner.Line, idx int) bool {
	if idx > 0 && !lines[idx-1].IsEmpty {
		prevText := strings.TrimSpace(lines[idx-1].Raw)
		if prevText != "" && !endsWithPunctuation(prevText) {
			return true
		}
	}
	return false
}

func isStepPattern(text string) bool {
	lower := strings.ToLower(text)
	if strings.HasPrefix(lower, "step ") {
		rest := lower[5:]
		dotIdx := strings.Index(rest, ":")
		if dotIdx > 0 && dotIdx <= 3 {
			beforeColon := rest[:dotIdx]
			allDigit := true
			for _, c := range beforeColon {
				if !unicode.IsDigit(c) {
					allDigit = false
					break
				}
			}
			if allDigit {
				return true
			}
		}
	}
	return false
}

func determineContextualHeadingLevel(text string, indent int) int {
	if indent > 0 {
		level := 3 + indent/4
		if level > 6 {
			level = 6
		}
		return level
	}
	if len(text) <= 20 {
		return 2
	}
	return 3
}

func isLikelyCodeComment(text string, lines []scanner.Line, idx int) bool {
	if !strings.HasPrefix(text, "//") {
		return false
	}

	codeContextCount := 0
	nonEmptyCount := 0

	for i := 1; i <= 2; i++ {
		if idx-i >= 0 {
			prev := lines[idx-i]
			prevText := strings.TrimSpace(prev.Raw)
			if prevText != "" && !prev.IsEmpty {
				nonEmptyCount++
				if hasStrongCodeIndicators(prevText) {
					codeContextCount++
				}
			}
		}
		if idx+i < len(lines) {
			next := lines[idx+i]
			nextText := strings.TrimSpace(next.Raw)
			if nextText != "" && !next.IsEmpty {
				nonEmptyCount++
				if hasStrongCodeIndicators(nextText) {
					codeContextCount++
				}
			}
		}
	}

	return codeContextCount >= 1 && nonEmptyCount > 0
}

func isInCodeContext(lines []scanner.Line, idx int) bool {
	codeLines := 0
	checkRange := 3

	for i := 1; i <= checkRange; i++ {
		if idx-i >= 0 {
			prev := lines[idx-i]
			prevText := strings.TrimSpace(prev.Raw)
			if prevText != "" && !prev.IsEmpty && hasStrongCodeIndicators(prevText) {
				codeLines++
			}
		}
		if idx+i < len(lines) {
			next := lines[idx+i]
			nextText := strings.TrimSpace(next.Raw)
			if nextText != "" && !next.IsEmpty && hasStrongCodeIndicators(nextText) {
				codeLines++
			}
		}
	}

	return codeLines >= 2
}

func isNearCodeContent(lines []scanner.Line, idx int) bool {
	for i := 1; i <= 3; i++ {
		if idx+i < len(lines) {
			next := lines[idx+i]
			nextText := strings.TrimSpace(next.Raw)
			if nextText != "" && !next.IsEmpty {
				if looksLikeCode(nextText) || looksLikeShellCommand(nextText) {
					return true
				}
				break
			}
		}
	}
	for i := 1; i <= 2; i++ {
		if idx-i >= 0 {
			prev := lines[idx-i]
			prevText := strings.TrimSpace(prev.Raw)
			if prevText != "" && !prev.IsEmpty {
				if looksLikeCode(prevText) || looksLikeShellCommand(prevText) {
					return true
				}
				break
			}
		}
	}
	return false
}

func hasStrongCodeIndicators(text string) bool {
	strongIndicators := []string{
		"func ", "var ", "const ", "return ", "class ",
		"public ", "private ", "void ", "new ",
		"import ", "package ", "@Override", "@Autowired",
		"@PostMapping", "@GetMapping", "@RequestBody",
		"redisTemplate.", "userService.", "JSON.",
		"UserHolder.", "Result.", "TimeUnit.",
		"StringRedisTemplate", "HttpServletRequest",
		"HttpServletResponse", "HandlerInterceptor",
	}
	for _, indicator := range strongIndicators {
		if strings.Contains(text, indicator) {
			return true
		}
	}

	braceCount := 0
	for _, c := range text {
		if c == '{' || c == '}' {
			braceCount++
		}
	}
	if braceCount >= 2 {
		return true
	}

	return false
}

func isNumberedHeading(text string) bool {
	prefixes := []string{"一、", "二、", "三、", "四、", "五、", "六、", "七、", "八、", "九、", "十、"}
	for _, p := range prefixes {
		if strings.HasPrefix(text, p) {
			return true
		}
	}
	return false
}

func determineHeadingLevel(text string, indent int) int {
	if indent == 0 {
		if len(text) < 20 {
			return 2
		}
		return 3
	}
	level := 3 + indent/4
	if level > 6 {
		level = 6
	}
	return level
}

func hasLatinLetter(s string) bool {
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') {
			return true
		}
	}
	return false
}

func isMostlyUppercaseLatin(s string) bool {
	latinCount := 0
	upperCount := 0
	totalLetters := 0
	for _, r := range s {
		if r >= 'a' && r <= 'z' {
			latinCount++
			totalLetters++
		} else if r >= 'A' && r <= 'Z' {
			latinCount++
			upperCount++
			totalLetters++
		} else if r >= 0x4E00 && r <= 0x9FFF {
			totalLetters++
		}
	}
	if latinCount == 0 {
		return false
	}
	if float64(upperCount)/float64(latinCount) <= 0.5 {
		return false
	}
	if float64(latinCount)/float64(totalLetters) < 0.5 {
		return false
	}
	return true
}

func hasLetterOrChinese(s string) bool {
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') {
			return true
		}
		if r >= 0x4E00 && r <= 0x9FFF {
			return true
		}
	}
	return false
}
