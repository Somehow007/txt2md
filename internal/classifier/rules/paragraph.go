package rules

import (
	"strings"

	"github.com/Somehow007/txt2md/internal/scanner"
)

// ParagraphRule detects paragraphs (default fallback).
type ParagraphRule struct{}

func (r *ParagraphRule) Name() string {
	return "paragraph"
}

func (r *ParagraphRule) Detect(lines []scanner.Line, idx int, opts Options) (*scanner.Block, int) {
	line := lines[idx]

	if line.IsEmpty {
		return nil, 0
	}

	// Collect consecutive non-empty lines as a paragraph
	paraLines := []scanner.Line{line}
	consumed := 1

	for idx+consumed < len(lines) {
		next := lines[idx+consumed]
		if next.IsEmpty {
			// Check if paragraph is truly ended
			break
		}
		// Stop if next line is a list item
		nextText := strings.TrimSpace(next.Raw)
		if isListItem(nextText, next.Indent) {
			break
		}
		paraLines = append(paraLines, next)
		consumed++
	}

	return &scanner.Block{
		Type:       scanner.BlockParagraph,
		Lines:      paraLines,
		Confidence: 0.6,
	}, consumed
}
