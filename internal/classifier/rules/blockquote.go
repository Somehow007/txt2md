package rules

import (
	"strings"

	"github.com/Somehow007/txt2md/internal/scanner"
)

// BlockquoteRule detects blockquotes (lines starting with >).
type BlockquoteRule struct{}

func (r *BlockquoteRule) Name() string {
	return "blockquote"
}

func (r *BlockquoteRule) Detect(lines []scanner.Line, idx int, opts Options) (*scanner.Block, int) {
	line := lines[idx]
	text := strings.TrimSpace(line.Raw)

	if text == "" {
		return nil, 0
	}

	// Check if line starts with > (blockquote marker)
	if !strings.HasPrefix(text, ">") {
		return nil, 0
	}

	// Collect consecutive blockquote lines
	quoteLines := []scanner.Line{line}
	consumed := 1

	for idx+consumed < len(lines) {
		next := lines[idx+consumed]
		nextText := strings.TrimSpace(next.Raw)
		if strings.HasPrefix(nextText, ">") {
			quoteLines = append(quoteLines, next)
			consumed++
		} else if next.IsEmpty {
			// Check if next non-empty line is also a blockquote
			isFollowedByQuote := false
			for j := idx + consumed + 1; j < len(lines); j++ {
				if lines[j].IsEmpty {
					continue
				}
				if strings.HasPrefix(strings.TrimSpace(lines[j].Raw), ">") {
					isFollowedByQuote = true
				}
				break
			}
			if isFollowedByQuote {
				quoteLines = append(quoteLines, next)
				consumed++
			} else {
				break
			}
		} else {
			break
		}
	}

	return &scanner.Block{
		Type:       scanner.BlockQuote,
		Lines:      quoteLines,
		Confidence: 0.95,
	}, consumed
}
