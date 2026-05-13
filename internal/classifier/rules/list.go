package rules

import (
	"strings"

	"github.com/user/txt2md/internal/scanner"
)

// ListRule detects ordered and unordered lists.
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

	// Check if this line is a list item
	if !isListItem(text, line.Indent) {
		return nil, 0
	}

	// Collect consecutive list items
	listLines := []scanner.Line{line}
	consumed := 1

	for idx+consumed < len(lines) {
		next := lines[idx+consumed]
		if next.IsEmpty {
			// Check if next non-empty line is also a list item
			if idx+consumed+1 < len(lines) {
				nextNext := lines[idx+consumed+1]
				if isListItem(strings.TrimSpace(nextNext.Raw), nextNext.Indent) {
					consumed++ // skip empty line
					continue
				}
			}
			break
		}
		if isListItem(strings.TrimSpace(next.Raw), next.Indent) {
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

func isListItem(text string, indent int) bool {
	if len(text) == 0 {
		return false
	}

	// Unordered list markers: -, *, +, •
	if len(text) >= 2 {
		first := rune(text[0])
		if (first == '-' || first == '*' || first == '+' || first == '•') && (text[1] == ' ' || text[1] == '\t') {
			return true
		}
	}

	// Ordered list: "1. ", "1) ", "a. ", etc.
	if len(text) >= 3 {
		// Check for patterns like "1. ", "1) "
		for i := 0; i < len(text) && i < 5; i++ {
			if text[i] >= '0' && text[i] <= '9' {
				continue
			}
			if i > 0 && (text[i] == '.' || text[i] == ')') && i+1 < len(text) && (text[i+1] == ' ' || text[i+1] == '\t') {
				return true
			}
			break
		}
		// Check for "a. " or "A. "
		if (text[0] >= 'a' && text[0] <= 'z') || (text[0] >= 'A' && text[0] <= 'Z') {
			if len(text) > 1 && (text[1] == '.' || text[1] == ')') && len(text) > 2 && (text[2] == ' ' || text[2] == '\t') {
				return true
			}
		}
	}

	// Chinese numbered list: "一、", "二、", etc.
	prefixes := []string{"一、", "二、", "三、", "四、", "五、", "六、", "七、", "八、", "九、", "十、"}
	for _, p := range prefixes {
		if strings.HasPrefix(text, p) {
			return true
		}
	}

	return false
}
