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
	confidence := 0.8

	// Markdown-style heading: "# Title", "## Title", etc.
	if strings.HasPrefix(text, "#") {
		level := 0
		for _, r := range text {
			if r == '#' {
				level++
			} else {
				break
			}
		}
		if level >= 1 && level <= 6 && len(text) > level && text[level] == ' ' {
			isHeading = true
			confidence = 1.0
		}
	}

	// Comment-style heading: "// Title" or "/// Title"
	if !isHeading && (strings.HasPrefix(text, "// ") || strings.HasPrefix(text, "/// ")) {
		isHeading = true
		confidence = 0.95
	}

	// All caps short line (likely a title)
	// Must have at least some letters/Chinese characters, not just symbols
	if !isHeading && text == strings.ToUpper(text) && len(text) > 3 && len(text) < 50 {
		if hasLetterOrChinese(text) {
			isHeading = true
		}
	}

	// Numbered heading like "1. Title" or "一、Title"
	if !isHeading && isNumberedHeading(text) {
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
		Confidence: confidence,
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

// hasLetterOrChinese checks if the text contains at least one letter or Chinese character.
func hasLetterOrChinese(s string) bool {
	for _, r := range s {
		// Check for ASCII letters
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') {
			return true
		}
		// Check for Chinese characters (CJK Unified Ideographs range)
		if r >= 0x4E00 && r <= 0x9FFF {
			return true
		}
	}
	return false
}
