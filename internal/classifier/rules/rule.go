package rules

import (
	"github.com/user/txt2md/internal/scanner"
)

// Rule defines the interface for classification rules.
type Rule interface {
	// Name returns the rule identifier.
	Name() string
	// Detect examines a group of lines and returns a Block if matched.
	// idx is the current line index, lines is the full line slice.
	Detect(lines []scanner.Line, idx int, opts Options) (*scanner.Block, int)
}

// Options provides context for rule detection.
type Options struct {
	TabWidth int
}
