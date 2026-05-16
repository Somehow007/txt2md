package nlp

import (
	"strings"
	"unicode"
)

func isSentenceEnd(r rune) bool {
	return r == '。' || r == '！' || r == '？' || r == '.' || r == '!' || r == '?'
}

func isCJK(r rune) bool {
	return unicode.Is(unicode.Han, r) ||
		unicode.Is(unicode.Hiragana, r) ||
		unicode.Is(unicode.Katakana, r) ||
		unicode.Is(unicode.Hangul, r)
}

func isSectionMarker(text string) bool {
	trimmed := strings.TrimSpace(text)
	if len(trimmed) == 0 {
		return false
	}
	runes := []rune(trimmed)
	if len(runes) >= 2 && runes[0] == '第' {
		return true
	}
	upper := strings.ToUpper(trimmed)
	if strings.HasPrefix(upper, "CHAPTER ") ||
		strings.HasPrefix(upper, "SECTION ") ||
		strings.HasPrefix(upper, "PART ") {
		return true
	}
	return false
}

func isParagraphStart(text string) bool {
	if len(text) == 0 {
		return false
	}
	r := []rune(text)[0]
	if isCJK(r) {
		return true
	}
	if r >= 'A' && r <= 'Z' {
		return true
	}
	if isSectionMarker(text) {
		return true
	}
	return false
}

func indentLevel(line string) int {
	count := 0
	for _, r := range line {
		if r == ' ' || r == '\t' {
			count++
		} else {
			break
		}
	}
	return count
}

func splitWithinLine(text string) string {
	runes := []rune(text)
	var result []rune
	for i := 0; i < len(runes); i++ {
		result = append(result, runes[i])
		if isSentenceEnd(runes[i]) {
			j := i + 1
			for j < len(runes) && (runes[j] == ' ' || runes[j] == '\t') {
				j++
			}
			if j < len(runes) {
				if isCJK(runes[j]) || isSectionMarker(string(runes[j:])) {
					for k := i + 1; k < j; k++ {
						result = append(result, runes[k])
					}
					result = append(result, '\n')
				}
			}
		}
	}
	return string(result)
}

func Segment(text string) string {
	if text == "" {
		return ""
	}

	text = splitWithinLine(text)

	lines := strings.Split(text, "\n")
	if len(lines) <= 1 {
		return text
	}

	totalLen := 0
	nonEmptyCount := 0
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if len(trimmed) > 0 {
			totalLen += len([]rune(trimmed))
			nonEmptyCount++
		}
	}
	avgLen := 0
	if nonEmptyCount > 0 {
		avgLen = totalLen / nonEmptyCount
	}

	var result []string
	for i := 0; i < len(lines); i++ {
		result = append(result, lines[i])

		if i >= len(lines)-1 {
			continue
		}

		current := lines[i]
		next := lines[i+1]

		currentTrimmed := strings.TrimSpace(current)
		nextTrimmed := strings.TrimSpace(next)

		if currentTrimmed == "" || nextTrimmed == "" {
			continue
		}

		needsBreak := false

		currentRunes := []rune(currentTrimmed)
		nextRunes := []rune(nextTrimmed)

		if len(currentRunes) > 0 && isSentenceEnd(currentRunes[len(currentRunes)-1]) && isParagraphStart(nextTrimmed) {
			needsBreak = true
		}

		if !needsBreak && len(nextRunes) > 0 && avgLen > 0 {
			nextLen := len(nextRunes)
			if nextLen < avgLen*2/5 && nextLen < 30 {
				needsBreak = true
			}
		}

		if !needsBreak && len(currentRunes) > 0 && avgLen > 0 {
			currentLen := len(currentRunes)
			if currentLen < avgLen*2/5 && currentLen < 30 {
				needsBreak = true
			}
		}

		if !needsBreak {
			currentIndent := indentLevel(current)
			nextIndent := indentLevel(next)
			if currentIndent != nextIndent && (currentIndent > 0 || nextIndent > 0) {
				needsBreak = true
			}
		}

		if !needsBreak && isSectionMarker(nextTrimmed) {
			needsBreak = true
		}

		if needsBreak {
			result = append(result, "")
		}
	}

	return strings.Join(result, "\n")
}
