package rules

import (
	"strings"
	"unicode"

	"github.com/Somehow007/txt2md/internal/scanner"
)

type ListRule struct{}

func (r *ListRule) Name() string {
	return "list"
}

func (r *ListRule) Detect(lines []scanner.Line, idx int, opts Options) (*scanner.Block, int) {
	line := lines[idx]
	text := strings.TrimSpace(line.Raw)

	if text == "" {
		return nil, 0
	}

	if !isListItem(text, line.Indent) {
		return nil, 0
	}

	listLines := []scanner.Line{line}
	consumed := 1

	for idx+consumed < len(lines) {
		next := lines[idx+consumed]
		if next.IsEmpty {
			if idx+consumed+1 < len(lines) {
				nextNext := lines[idx+consumed+1]
				nextNextText := strings.TrimSpace(nextNext.Raw)
				if isListItem(nextNextText, nextNext.Indent) && !isNumberedHeading(nextNextText) {
					consumed++
					continue
				}
			}
			break
		}
		nextText := strings.TrimSpace(next.Raw)
		if isNumberedHeading(nextText) {
			break
		}
		if isListItem(nextText, next.Indent) {
			listLines = append(listLines, next)
			consumed++
		} else if isListContinuation(nextText, next.Indent, lines, idx+consumed) {
			listLines = append(listLines, next)
			consumed++
		} else {
			break
		}
	}

	return &scanner.Block{
		Type:       scanner.BlockList,
		Lines:      listLines,
		Confidence: 0.9,
	}, consumed
}

func isListContinuation(text string, indent int, lines []scanner.Line, idx int) bool {
	if len(text) == 0 {
		return false
	}

	if looksLikeShellCommand(text) {
		return false
	}

	if looksLikeCode(text) && isConsecutiveCodeLine(lines, idx) {
		return false
	}

	if idx > 0 {
		prev := lines[idx-1]
		prevText := strings.TrimSpace(prev.Raw)
		if isListItem(prevText, prev.Indent) {
			if !endsWithPunctuation(prevText) {
				return true
			}
		}
		if isListContinuation(prevText, prev.Indent, lines, idx-1) {
			if !endsWithPunctuation(prevText) {
				return true
			}
		}
	}

	return false
}

func isConsecutiveCodeLine(lines []scanner.Line, idx int) bool {
	codeCount := 0
	for i := idx + 1; i < len(lines) && i <= idx+2; i++ {
		if lines[i].IsEmpty {
			continue
		}
		nextText := strings.TrimSpace(lines[i].Raw)
		if looksLikeCode(nextText) || looksLikeShellCommand(nextText) {
			codeCount++
		}
	}
	return codeCount >= 1
}

func endsWithPunctuation(s string) bool {
	if len(s) == 0 {
		return false
	}
	runes := []rune(s)
	last := runes[len(runes)-1]
	return last == '。' || last == '！' || last == '？' ||
		last == '.' || last == '!' || last == '?' ||
		last == ';' || last == '；' || last == '：' || last == ':'
}

func isListItem(text string, indent int) bool {
	if len(text) == 0 {
		return false
	}

	if isNumberedHeading(text) {
		return false
	}

	if len(text) >= 2 {
		first := rune(text[0])
		if (first == '-' || first == '*' || first == '+' || first == '•') && (text[1] == ' ' || text[1] == '\t') {
			return true
		}
	}

	if isChineseNumberedList(text) {
		return true
	}

	if isSubNumberedList(text) {
		return true
	}

	if isNumberedSectionTitle(text) {
		return false
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
		if (text[0] >= 'a' && text[0] <= 'z') || (text[0] >= 'A' && text[0] <= 'Z') {
			if len(text) > 1 && (text[1] == '.' || text[1] == ')') && len(text) > 2 && (text[2] == ' ' || text[2] == '\t') {
				return true
			}
		}
	}

	return false
}

func isChineseNumberedList(text string) bool {
	if len(text) < 3 {
		return false
	}

	i := 0
	for i < len(text) && i < 5 {
		r, size := utf8DecodeRune(text[i:])
		if r == 0 {
			break
		}
		if unicode.IsDigit(r) {
			i += size
			continue
		}
		if r == '、' && i > 0 {
			if i+1 < len(text) {
				nextR, _ := utf8DecodeRune(text[i+size:])
				if nextR != ' ' && nextR != '\t' && nextR != 0 {
					return true
				}
			}
		}
		break
	}
	return false
}

func isSubNumberedList(text string) bool {
	dotCount := 0
	digitBefore := false
	digitAfter := false
	i := 0
	for i < len(text) && i < 10 {
		r, size := utf8DecodeRune(text[i:])
		if r == 0 {
			break
		}
		if unicode.IsDigit(r) {
			if dotCount == 0 {
				digitBefore = true
			} else if dotCount == 1 {
				digitAfter = true
			}
		} else if r == '.' && digitBefore {
			dotCount++
		} else if (r == ' ' || r == '\t') && digitBefore && digitAfter && dotCount == 1 {
			return true
		} else if !unicode.IsDigit(r) && r != '.' {
			break
		}
		i += size
	}
	return false
}

func isNumberedSectionTitle(text string) bool {
	if len(text) < 4 {
		return false
	}

	i := 0
	for i < len(text) && i < 3 {
		if text[i] >= '0' && text[i] <= '9' {
			i++
			continue
		}
		break
	}

	if i == 0 {
		return false
	}

	if i >= len(text) || text[i] != ' ' {
		return false
	}

	i++
	for i < len(text) {
		r, _ := utf8DecodeRune(text[i:])
		if r >= 0x4E00 && r <= 0x9FFF {
			return true
		}
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') {
			return true
		}
		break
	}

	return false
}

func utf8DecodeRune(s string) (rune, int) {
	if len(s) == 0 {
		return 0, 0
	}
	r := rune(s[0])
	size := 1
	if s[0] >= 0xC0 {
		if s[0] >= 0xF0 && len(s) >= 4 {
			r = rune(s[0]&0x07)<<18 | rune(s[1]&0x3F)<<12 | rune(s[2]&0x3F)<<6 | rune(s[3]&0x3F)
			size = 4
		} else if s[0] >= 0xE0 && len(s) >= 3 {
			r = rune(s[0]&0x0F)<<12 | rune(s[1]&0x3F)<<6 | rune(s[2]&0x3F)
			size = 3
		} else if s[0] >= 0xC0 && len(s) >= 2 {
			r = rune(s[0]&0x1F)<<6 | rune(s[1]&0x3F)
			size = 2
		}
	}
	return r, size
}
