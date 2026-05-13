package rules

import (
	"strings"

	"github.com/Somehow007/txt2md/internal/scanner"
)

// HorizontalRule detects horizontal rules (---, ***, ___ with 3+ chars).
type HorizontalRule struct{}

func (r *HorizontalRule) Name() string {
	return "horizontal"
}

func (r *HorizontalRule) Detect(lines []scanner.Line, idx int, opts Options) (*scanner.Block, int) {
	if idx >= len(lines) {
		return nil, 0
	}

	line := lines[idx]
	if line.IsEmpty {
		return nil, 0
	}

	text := strings.TrimSpace(line.Raw)
	if text == "" {
		return nil, 0
	}

	// Must be a single line (no other content)
	if len(text) < 3 {
		return nil, 0
	}

	// Check for horizontal rule patterns:
	// --- (3+ hyphens)
	// *** (3+ asterisks)
	// ___ (3+ underscores)
	// Or repeated characters like ===, +++, ###
	if r.isHorizontalRule(text) {
		return &scanner.Block{
			Type:       scanner.BlockHorizontal,
			Lines:      []scanner.Line{line},
			Confidence: 0.95,
		}, 1
	}

	return nil, 0
}

func (r *HorizontalRule) isHorizontalRule(text string) bool {
	if len(text) < 3 {
		return false
	}

	// Get the first character
	first := rune(text[0])
	if !strings.ContainsRune("-*_=+#", first) {
		return false
	}

	// Check if all characters are the same
	for _, ch := range text {
		if ch != first {
			return false
		}
	}

	return true
}
