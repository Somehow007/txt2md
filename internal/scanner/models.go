package scanner

import "strings"

// BlockType represents the type of a text block.
type BlockType int

const (
	BlockParagraph BlockType = iota
	BlockHeading
	BlockList
	BlockCode
	BlockQuote
	BlockTable
	BlockHorizontal
)

// Line represents a single line of text with metadata.
type Line struct {
	Raw        string // original line content
	Number     int    // line number (1-based)
	Indent     int    // leading whitespace count
	IsEmpty    bool   // whether line is blank
	IsMarkdown bool   // whether line contains existing Markdown syntax
}

// TableData holds pre-parsed table structure for rendering.
type TableData struct {
	Rows [][]string // each row is a slice of cell strings
}

// Block represents a contiguous group of lines with the same type.
type Block struct {
	Type         BlockType
	Lines        []Line
	Confidence   float64    // classification confidence (0.0-1.0)
	TableData    *TableData // populated when Type == BlockTable
	HeadingLevel int        // populated when Type == BlockHeading (1-6)
}

// NewLine creates a Line from a raw string.
func NewLine(raw string, number int) Line {
	indent := len(raw) - len(strings.TrimLeft(raw, " \t"))
	return Line{
		Raw:     raw,
		Number:  number,
		Indent:  indent,
		IsEmpty: strings.TrimSpace(raw) == "",
	}
}
