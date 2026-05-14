package scanner

import (
	"regexp"
	"strings"
)

var (
	mdHeadingRe     = regexp.MustCompile(`^ {0,3}#{1,6}\s+\S`)
	mdFenceRe       = regexp.MustCompile("^ {0,3}```")
	mdTildeFenceRe  = regexp.MustCompile("^ {0,3}~~~")
	mdBlockquoteRe  = regexp.MustCompile(`^ {0,3}>`)
	mdListRe        = regexp.MustCompile(`^ {0,3}[-*+]\s`)
	mdOrderedListRe = regexp.MustCompile(`^ {0,3}\d+[.)]\s`)
	mdTablePipeRe   = regexp.MustCompile(`\|`)
	mdImageRe       = regexp.MustCompile(`!\[.*?\]\(.*?\)`)
	mdLinkRe        = regexp.MustCompile(`\[[^\]]*\]\(.*?\)`)
	mdBoldRe        = regexp.MustCompile(`(\*\*|__).*?(\*\*|__)`)
	mdItalicRe      = regexp.MustCompile(`(\*|_)[^\s].*?(\*|_)`)
	mdInlineCodeRe  = regexp.MustCompile("`[^`]+`")
)

func detectMarkdownSyntax(raw string) bool {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return false
	}

	if mdHeadingRe.MatchString(trimmed) {
		return true
	}
	if mdFenceRe.MatchString(trimmed) || mdTildeFenceRe.MatchString(trimmed) {
		return true
	}
	if mdBlockquoteRe.MatchString(trimmed) {
		return true
	}
	if isHorizontalRule(trimmed) {
		return true
	}
	if mdListRe.MatchString(trimmed) {
		return true
	}
	if mdOrderedListRe.MatchString(trimmed) {
		return true
	}

	if mdImageRe.MatchString(trimmed) {
		return true
	}
	if mdLinkRe.MatchString(trimmed) {
		return true
	}
	if mdBoldRe.MatchString(trimmed) {
		return true
	}
	if mdItalicRe.MatchString(trimmed) {
		return true
	}
	if mdInlineCodeRe.MatchString(trimmed) {
		return true
	}

	if mdTablePipeRe.MatchString(trimmed) && strings.Count(trimmed, "|") >= 2 {
		return true
	}

	return false
}

func isHorizontalRule(text string) bool {
	if len(text) < 3 {
		return false
	}
	first := rune(text[0])
	if first != '-' && first != '*' && first != '_' {
		return false
	}
	for _, ch := range text {
		if ch != first && ch != ' ' && ch != '\t' {
			return false
		}
	}
	count := 0
	for _, ch := range text {
		if ch == first {
			count++
		}
	}
	return count >= 3
}

func Scan(input string) []Line {
	rawLines := strings.Split(input, "\n")
	lines := make([]Line, 0, len(rawLines))

	for i, raw := range rawLines {
		raw = strings.TrimRight(raw, "\r")
		line := NewLine(raw, i+1)
		line.IsMarkdown = detectMarkdownSyntax(raw)
		lines = append(lines, line)
	}

	return lines
}
