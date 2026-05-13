package rules

import (
	"regexp"
	"strings"

	"github.com/Somehow007/txt2md/internal/scanner"
)

// TableRule detects text tables separated by 2+ spaces or tabs.
type TableRule struct{}

func (r *TableRule) Name() string {
	return "table"
}

func (r *TableRule) Detect(lines []scanner.Line, idx int, opts Options) (*scanner.Block, int) {
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

// isSeparatorRow checks if a line is a Markdown table separator (e.g., ---|---|---)
func (r *TableRule) isSeparatorRow(raw string, expectedCols int) bool {
	// Must contain at least one '-' or '|'
	if !strings.ContainsAny(raw, "-|") {
		return false
	}

	// Common patterns: "---|---|---" or "---  ---  ---" or "--- | --- | ---"
	colCount := 0
	hasDash := false

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

		colCount = 0
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
