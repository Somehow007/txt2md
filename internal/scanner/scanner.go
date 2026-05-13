package scanner

import "strings"

// Scan splits input text into lines.
func Scan(input string) []Line {
	rawLines := strings.Split(input, "\n")
	lines := make([]Line, 0, len(rawLines))

	for i, raw := range rawLines {
		// Remove trailing \r (Windows line endings)
		raw = strings.TrimRight(raw, "\r")
		lines = append(lines, NewLine(raw, i+1))
	}

	return lines
}
