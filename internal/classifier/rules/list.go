package rules

import (
	"strings"

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
				if isListItem(nextNextText, nextNext.Indent) {
					consumed++
					continue
				}
			}
			break
		}
		nextText := strings.TrimSpace(next.Raw)
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

	if len(text) >= 2 {
		first := rune(text[0])
		if (first == '-' || first == '*' || first == '+' || first == '•') && (text[1] == ' ' || text[1] == '\t') {
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
		if (text[0] >= 'a' && text[0] <= 'z') || (text[0] >= 'A' && text[0] <= 'Z') {
			if len(text) > 1 && (text[1] == '.' || text[1] == ')') && len(text) > 2 && (text[2] == ' ' || text[2] == '\t') {
				return true
			}
		}
	}

	prefixes := []string{"一、", "二、", "三、", "四、", "五、", "六、", "七、", "八、", "九、", "十、"}
	for _, p := range prefixes {
		if strings.HasPrefix(text, p) {
			return true
		}
	}

	return false
}
