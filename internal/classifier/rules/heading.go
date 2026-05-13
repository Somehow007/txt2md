package rules

import (
	"strings"

	"github.com/Somehow007/txt2md/internal/scanner"
)

// HeadingRule detects headings based on line length, position, and patterns.
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

	// Skip if line contains common non-heading patterns
	if strings.Contains(text, ".") && strings.Contains(text, " ") && len(text) > 50 {
		return nil, 0
	}

	// Heading indicators
	isHeading := false

	// All caps short line (likely a title)
	if text == strings.ToUpper(text) && len(text) > 3 && len(text) < 50 {
		isHeading = true
	}

	// Line ends with colon (Chinese/English) - disabled to reduce false positives
	// if (strings.HasSuffix(text, ":") || strings.HasSuffix(text, "：")) && len(text) < 15 {
	// 	isHeading = true
	// }

	// Next line is empty and line is very short (likely a heading)
	// Disabled to reduce false positives
	// if idx+1 < len(lines) && lines[idx+1].IsEmpty && len(text) < 20 {
	// 	isHeading = true
	// }

	// Numbered heading like "1. Title" or "一、Title"
	if isNumberedHeading(text) {
		isHeading = true
	}

	if !isHeading {
		return nil, 0
	}

	// Determine heading level (currently unused, reserved for future enhancement)
	_ = determineHeadingLevel(text, line.Indent)

	return &scanner.Block{
		Type:       scanner.BlockHeading,
		Lines:      []scanner.Line{line},
		Confidence: 0.8,
	}, 1
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
	// Simple heuristic: less indent = higher level
	if indent == 0 {
		if len(text) < 20 {
			return 1
		}
		return 2
	}
	return 3 + indent/4
}
