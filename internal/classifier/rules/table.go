package rules

import (
	"regexp"
	"strings"

	"github.com/Somehow007/txt2md/internal/scanner"
)

// unicodeTableRow represents a row in a Unicode box-drawing table.
type unicodeTableRow struct {
	rawLine  string
	cells    []string
	isBorder bool
}

// TableRule detects text tables in various formats:
// 1. Unicode box-drawing tables (┌─┐│─├─┤└─┘)
// 2. Simple space/tab separated tables
type TableRule struct{}

func (r *TableRule) Name() string {
	return "table"
}

func (r *TableRule) Detect(lines []scanner.Line, idx int, opts Options) (*scanner.Block, int) {
	// Try Unicode box-drawing table first (highest confidence, most distinctive)
	if block, consumed := r.detectUnicodeTable(lines, idx); block != nil {
		return block, consumed
	}
	// Try ASCII art table with +---+ borders
	if block, consumed := r.detectASCIITable(lines, idx); block != nil {
		return block, consumed
	}
	// Fall back to simple space/tab separated table
	return r.detectSimpleTable(lines, idx)
}

// detectUnicodeTable detects tables using Unicode box-drawing characters.
func (r *TableRule) detectUnicodeTable(lines []scanner.Line, idx int) (*scanner.Block, int) {
	if idx >= len(lines) || lines[idx].IsEmpty {
		return nil, 0
	}

	raw := lines[idx].Raw
	if !containsBoxDrawing(raw) {
		return nil, 0
	}

	// Must contain box-drawing characters that indicate a table
	if !strings.ContainsAny(raw, "┌┐└┘├┤┬┴┼│") {
		return nil, 0
	}

	// Collect all rows
	var rows []unicodeTableRow
	consumed := 0

	for i := idx; i < len(lines); i++ {
		line := lines[i].Raw
		trimmed := strings.TrimSpace(line)

		// Stop at bottom border
		if isUnicodeBottomBorder(trimmed) {
			consumed = i - idx + 1
			break
		}

		// Stop if line doesn't look like part of the table
		if !strings.ContainsAny(line, "│├┤┬┴┼") && !isUnicodeBorderLine(trimmed) {
			if len(rows) > 0 {
				break
			}
			return nil, 0
		}

		// Check if this is a border/separator line
		if isUnicodeBorderLine(trimmed) {
			rows = append(rows, unicodeTableRow{
				rawLine:  line,
				cells:    nil,
				isBorder: true,
			})
			consumed = i - idx + 1
			continue
		}

		// Extract cells from content row
		cells := extractUnicodeCells(line)
		if len(cells) >= 2 {
			rows = append(rows, unicodeTableRow{
				rawLine:  line,
				cells:    cells,
				isBorder: false,
			})
			consumed = i - idx + 1
		} else {
			break
		}
	}

	// Filter to only content rows
	var segments [][]unicodeTableRow
	var currentSegment []unicodeTableRow

	for _, row := range rows {
		if row.isBorder {
			if len(currentSegment) > 0 {
				segments = append(segments, currentSegment)
				currentSegment = nil
			}
			continue
		}
		currentSegment = append(currentSegment, row)
	}
	if len(currentSegment) > 0 {
		segments = append(segments, currentSegment)
	}

	if len(segments) == 0 {
		return nil, 0
	}

	var allMergedRows [][]string
	for _, seg := range segments {
		merged := mergeSegmentRows(seg)
		allMergedRows = append(allMergedRows, merged...)
	}

	allLines := make([]scanner.Line, consumed)
	copy(allLines, lines[idx:idx+consumed])

	return &scanner.Block{
		Type:       scanner.BlockTable,
		Lines:      allLines,
		Confidence: 0.95,
		TableData: &scanner.TableData{
			Rows: allMergedRows,
		},
	}, consumed
}

// mergeUnicodeRows merges consecutive rows where cells continue across lines.
// Only merges if the next row is clearly a continuation (not a new row).
func mergeSegmentRows(rows []unicodeTableRow) [][]string {
	if len(rows) == 0 {
		return nil
	}

	result := [][]string{rows[0].cells}

	for i := 1; i < len(rows); i++ {
		prev := result[len(result)-1]
		next := rows[i].cells

		if shouldMergeCells(prev, next) {
			for j := range prev {
				if j < len(next) && next[j] != "" {
					prev[j] = strings.TrimSpace(prev[j] + next[j])
				}
			}
		} else {
			result = append(result, next)
		}
	}

	return result
}

func mergeUnicodeRows(rows []unicodeTableRow, hasSeparator bool) [][]string {
	if len(rows) == 0 {
		return nil
	}

	result := [][]string{}
	result = append(result, rows[0].cells)

	startIdx := 1
	if hasSeparator && len(rows) > 1 {
		result = append(result, rows[1].cells)
		startIdx = 2
	}

	for i := startIdx; i < len(rows); i++ {
		next := rows[i].cells

		if i+1 < len(rows) && shouldMergeCells(next, rows[i+1].cells) {
			merged := make([]string, len(next))
			copy(merged, next)
			for j := range merged {
				if j < len(rows[i+1].cells) && rows[i+1].cells[j] != "" {
					merged[j] = strings.TrimSpace(merged[j] + rows[i+1].cells[j])
				}
			}
			result = append(result, merged)
			i++
		} else {
			result = append(result, next)
		}
	}

	return result
}

// shouldMergeCells checks if next row continues a cell from prev row.
// Returns true only if BOTH rows have ONLY ONE non-empty cell in the SAME column.
func shouldMergeCells(prev, next []string) bool {
	if len(prev) != len(next) {
		return false
	}

	nextNonEmpty := 0
	for _, cell := range next {
		if cell != "" {
			nextNonEmpty++
		}
	}

	if nextNonEmpty == 0 {
		return false
	}

	prevNonEmpty := 0
	for _, cell := range prev {
		if cell != "" {
			prevNonEmpty++
		}
	}

	if prevNonEmpty == 0 {
		return false
	}

	if nextNonEmpty > prevNonEmpty {
		return false
	}

	shortContinuationCount := 0
	for j, nextCell := range next {
		if nextCell == "" || j >= len(prev) {
			continue
		}
		prevCell := prev[j]
		if prevCell == "" {
			continue
		}
		if !endsWithPunctuation(prevCell) && runeLen(nextCell) <= 5 {
			shortContinuationCount++
		}
	}

	if shortContinuationCount > 0 && float64(shortContinuationCount)/float64(nextNonEmpty) >= 0.5 {
		return true
	}

	allIncomplete := true
	incompleteCount := 0
	for j, nextCell := range next {
		if nextCell == "" || j >= len(prev) {
			continue
		}
		prevCell := prev[j]
		if prevCell == "" {
			continue
		}
		if !endsWithPunctuation(prevCell) {
			incompleteCount++
		} else {
			allIncomplete = false
		}
	}

	if incompleteCount == 0 {
		return false
	}

	if nextNonEmpty < prevNonEmpty {
		return true
	}

	return allIncomplete
}

func extractUnicodeCells(line string) []string {
	// Use rune-based splitting for correct Unicode handling
	runes := []rune(line)
	var cells []string
	start := 0
	firstCell := true

	for i, r := range runes {
		if r == '│' {
			// Extract text between previous │ and this one
			cell := strings.Trim(string(runes[start:i]), " ├┤┬┴┼┌┐└┘─━")
			// Skip the very first cell (before first │) as it's the left border
			if !firstCell {
				cells = append(cells, cell)
			}
			firstCell = false
			start = i + 1
		}
	}
	// Handle text after last │ (the right border - skip it)
	// if start < len(runes) {
	// 	cell := strings.Trim(string(runes[start:]), " ├┤┬┴┼┌┐└┘─━")
	// 	cells = append(cells, cell)
	// }

	return cells
}

// containsBoxDrawing checks if the string contains Unicode box-drawing characters (U+2500-U+257F).
func containsBoxDrawing(s string) bool {
	for _, r := range s {
		if r >= 0x2500 && r <= 0x257F {
			return true
		}
	}
	return false
}

// isBoxDrawingRune checks if a rune is a Unicode box-drawing character (U+2500 to U+257F).
func isBoxDrawingRune(r rune) bool {
	return r >= 0x2500 && r <= 0x257F
}

// isUnicodeBorderLine checks if a line is a Unicode table border/separator.
func isUnicodeBorderLine(line string) bool {
	trimmed := strings.TrimSpace(line)
	if len(trimmed) == 0 {
		return false
	}
	// Must contain box-drawing characters and dashes
	return containsBoxDrawing(trimmed) && strings.ContainsAny(trimmed, "─━")
}

// isUnicodeBottomBorder checks if a line is a Unicode table bottom border.
func isUnicodeBottomBorder(line string) bool {
	trimmed := strings.TrimSpace(line)
	// Check for └ and ┘ which indicate bottom border
	return strings.Contains(trimmed, "└") || strings.Contains(trimmed, "┘")
}

// detectASCIITable detects ASCII art tables with +---+ borders.
func (r *TableRule) detectASCIITable(lines []scanner.Line, idx int) (*scanner.Block, int) {
	if idx >= len(lines) || lines[idx].IsEmpty {
		return nil, 0
	}

	raw := lines[idx].Raw
	trimmed := strings.TrimSpace(raw)

	// Must look like an ASCII table border (starts/ends with +, contains -)
	if !strings.HasPrefix(trimmed, "+") || !strings.HasSuffix(trimmed, "+") {
		return nil, 0
	}
	if !strings.Contains(trimmed, "-") {
		return nil, 0
	}

	// Now collect the table rows
	var tableLines []scanner.Line
	tableLines = append(tableLines, lines[idx]) // border line
	consumed := 1

	// Look for content rows until the next border
	for i := idx + 1; i < len(lines); i++ {
		line := lines[i].Raw
		trimmedLine := strings.TrimSpace(line)

		// Skip empty lines
		if trimmedLine == "" {
			break
		}

		// Check if this is a border line
		if strings.HasPrefix(trimmedLine, "+") && strings.HasSuffix(trimmedLine, "+") && strings.Contains(trimmedLine, "-") {
			tableLines = append(tableLines, lines[i])
			consumed++
			// Check if this is the bottom border (same structure as top border)
			if i+1 >= len(lines) || lines[i+1].IsEmpty {
				consumed = i - idx + 1
				break
			}
			continue
		}

		// Check if this is a content row (contains |)
		if strings.Contains(line, "|") {
			tableLines = append(tableLines, lines[i])
			consumed++
			continue
		}

		// If we get here, line doesn't belong to table
		break
	}

	if len(tableLines) < 3 { // Need at least top border, content, bottom border
		return nil, 0
	}

	// Parse rows from the table lines
	var rows [][]string
	for _, line := range tableLines {
		text := strings.TrimSpace(line.Raw)
		// Skip border lines
		if strings.HasPrefix(text, "+") {
			continue
		}
		// Parse content row
		row := parseASCIICells(text)
		if len(row) >= 2 {
			rows = append(rows, row)
		}
	}

	if len(rows) < 1 {
		return nil, 0
	}

	allLines := make([]scanner.Line, consumed)
	copy(allLines, lines[idx:idx+consumed])

	return &scanner.Block{
		Type:       scanner.BlockTable,
		Lines:      allLines,
		Confidence: 0.9,
		TableData: &scanner.TableData{
			Rows: rows,
		},
	}, consumed
}

// parseASCIICells extracts cell content from an ASCII table row.
// Finds | characters in the line and extracts text between them.
func parseASCIICells(raw string) []string {
	var cells []string
	runes := []rune(raw)

	start := -1
	for i, r := range runes {
		if r == '|' {
			if start >= 0 {
				// Extract text between previous | and this one
				cell := strings.TrimSpace(string(runes[start+1 : i]))
				cells = append(cells, cell)
			}
			start = i
		}
	}
	// Handle text after last |
	if start >= 0 && start < len(runes)-1 {
		cell := strings.TrimSpace(string(runes[start+1:]))
		if cell != "" {
			cells = append(cells, cell)
		}
	}

	return cells
}

// detectSimpleTable is the original table detection logic for space/tab separated tables.
func (r *TableRule) detectSimpleTable(lines []scanner.Line, idx int) (*scanner.Block, int) {
	if idx >= len(lines) {
		return nil, 0
	}

	// Need at least 2 lines for a table
	if idx+1 >= len(lines) {
		return nil, 0
	}

	line := lines[idx]
	if line.IsEmpty {
		return nil, 0
	}

	// Try to parse first line as table row
	firstRow := r.parseRow(line.Raw)
	if len(firstRow) < 2 {
		return nil, 0
	}

	// Check if next line is a separator row (optional)
	hasSeparator := false
	consumedLines := 1

	if idx+1 < len(lines) && !lines[idx+1].IsEmpty {
		nextRaw := strings.TrimSpace(lines[idx+1].Raw)
		if r.isSeparatorRow(nextRaw, len(firstRow)) {
			hasSeparator = true
			consumedLines = 2
		}
	}

	// Collect data rows
	var tableLines []scanner.Line
	tableLines = append(tableLines, lines[idx])

	if hasSeparator {
		tableLines = append(tableLines, lines[idx+1])
	}

	// Collect subsequent data rows
	for i := idx + consumedLines; i < len(lines); i++ {
		if lines[i].IsEmpty {
			break
		}
		row := r.parseRow(lines[i].Raw)
		if len(row) < 2 || len(row) != len(firstRow) {
			break
		}
		tableLines = append(tableLines, lines[i])
		consumedLines++
	}

	if len(tableLines) < 2 || (hasSeparator && len(tableLines) < 3) {
		return nil, 0
	}

	confidence := 0.7
	if hasSeparator {
		confidence = 0.9
	}

	return &scanner.Block{
		Type:       scanner.BlockTable,
		Lines:      tableLines,
		Confidence: confidence,
	}, consumedLines
}

// parseRow splits a line into table columns by 2+ spaces, tabs, or '|'.
func (r *TableRule) parseRow(raw string) []string {
	// Handle Unicode box-drawing vertical line character
	if strings.Contains(raw, "│") {
		return parseUnicodeTableRow(raw)
	}
	// Split by '|' if present
	if strings.Contains(raw, "|") {
		parts := strings.Split(raw, "|")
		cols := make([]string, 0, len(parts))
		for _, p := range parts {
			trimmed := strings.TrimSpace(p)
			if trimmed != "" {
				cols = append(cols, trimmed)
			}
		}
		return cols
	}

	// Split by tab first
	raw = strings.ReplaceAll(raw, "\t", "  ")

	// Split by 2+ consecutive spaces
	spaceRe := regexp.MustCompile(` {2,}`)
	parts := spaceRe.Split(raw, -1)

	cols := make([]string, 0, len(parts))
	for _, p := range parts {
		trimmed := strings.TrimSpace(p)
		if trimmed != "" {
			cols = append(cols, trimmed)
		}
	}

	return cols
}

// parseUnicodeTableRow splits a line by Unicode box-drawing vertical line │.
func parseUnicodeTableRow(raw string) []string {
	parts := strings.Split(raw, "│")
	cols := make([]string, 0, len(parts))
	for _, p := range parts {
		trimmed := strings.TrimSpace(p)
		if trimmed != "" {
			cols = append(cols, trimmed)
		}
	}
	return cols
}

func runeLen(s string) int {
	return len([]rune(s))
}

// isSeparatorRow checks if a line is a table separator.
func (r *TableRule) isSeparatorRow(raw string, expectedCols int) bool {
	// Must contain at least one '-' or '|'
	if !strings.ContainsAny(raw, "-|") {
		return false
	}

	hasDash := false
	colCount := 0

	// Count columns by '|' or double spaces
	if strings.Contains(raw, "|") {
		parts := strings.Split(raw, "|")
		colCount = len(parts)
		for _, p := range parts {
			if strings.TrimSpace(p) == "" {
				return false
			}
			if strings.Contains(strings.TrimSpace(p), "-") {
				hasDash = true
			}
		}
	} else {
		// Split by 2+ spaces
		spaceRe := regexp.MustCompile(` {2,}`)
		parts := spaceRe.Split(raw, -1)

		for _, p := range parts {
			trimmed := strings.TrimSpace(p)
			if trimmed != "" {
				colCount++
				if strings.Contains(trimmed, "-") {
					hasDash = true
				}
			}
		}
	}

	if colCount < 2 || colCount != expectedCols {
		return false
	}

	return hasDash || strings.Contains(raw, "---")
}
