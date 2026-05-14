package diff

import (
	"fmt"
	"strings"
)

type LineDiff struct {
	OldLine string
	NewLine string
	Type    string
}

type Result struct {
	OldText string
	NewText string
	Diffs   []LineDiff
	Summary Summary
}

type Summary struct {
	TotalOldLines  int
	TotalNewLines  int
	AddedLines     int
	RemovedLines   int
	ChangedLines   int
	UnchangedLines int
}

func Compare(oldText, newText string) *Result {
	oldLines := strings.Split(oldText, "\n")
	newLines := strings.Split(newText, "\n")

	lcs := computeLCS(oldLines, newLines)
	diffs := buildDiffs(oldLines, newLines, lcs)

	summary := Summary{
		TotalOldLines: len(oldLines),
		TotalNewLines: len(newLines),
	}

	for _, d := range diffs {
		switch d.Type {
		case "added":
			summary.AddedLines++
		case "removed":
			summary.RemovedLines++
		case "changed":
			summary.ChangedLines++
		case "unchanged":
			summary.UnchangedLines++
		}
	}

	return &Result{
		OldText: oldText,
		NewText: newText,
		Diffs:   diffs,
		Summary: summary,
	}
}

func (r *Result) Unified() string {
	var sb strings.Builder

	for _, d := range r.Diffs {
		switch d.Type {
		case "added":
			sb.WriteString(fmt.Sprintf("+ %s\n", d.NewLine))
		case "removed":
			sb.WriteString(fmt.Sprintf("- %s\n", d.OldLine))
		case "changed":
			sb.WriteString(fmt.Sprintf("- %s\n", d.OldLine))
			sb.WriteString(fmt.Sprintf("+ %s\n", d.NewLine))
		case "unchanged":
			sb.WriteString(fmt.Sprintf("  %s\n", d.NewLine))
		}
	}

	return sb.String()
}

func (r *Result) SideBySide(width int) string {
	if width <= 0 {
		width = 80
	}
	halfWidth := width / 2

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("%-*s | %s\n", halfWidth, "--- Original ---", "--- Converted ---"))
	sb.WriteString(strings.Repeat("-", halfWidth) + "-+-" + strings.Repeat("-", halfWidth) + "\n")

	for _, d := range r.Diffs {
		var oldPart, newPart string
		switch d.Type {
		case "added":
			oldPart = ""
			newPart = "+ " + d.NewLine
		case "removed":
			oldPart = "- " + d.OldLine
			newPart = ""
		case "changed":
			oldPart = "- " + d.OldLine
			newPart = "+ " + d.NewLine
		case "unchanged":
			oldPart = "  " + d.OldLine
			newPart = "  " + d.NewLine
		}

		if len(oldPart) > halfWidth {
			oldPart = oldPart[:halfWidth]
		}
		if len(newPart) > halfWidth {
			newPart = newPart[:halfWidth]
		}

		sb.WriteString(fmt.Sprintf("%-*s | %s\n", halfWidth, oldPart, newPart))
	}

	return sb.String()
}

func (r *Result) Stats() string {
	s := r.Summary
	total := s.AddedLines + s.RemovedLines + s.ChangedLines + s.UnchangedLines
	if total == 0 {
		return "No changes detected."
	}
	changePercent := float64(s.AddedLines+s.RemovedLines+s.ChangedLines) / float64(total) * 100
	return fmt.Sprintf("Lines: %d -> %d | Added: %d | Removed: %d | Changed: %d | Unchanged: %d | Change rate: %.1f%%",
		s.TotalOldLines, s.TotalNewLines, s.AddedLines, s.RemovedLines, s.ChangedLines, s.UnchangedLines, changePercent)
}

func computeLCS(a, b []string) [][]int {
	m := len(a)
	n := len(b)

	dp := make([][]int, m+1)
	for i := range dp {
		dp[i] = make([]int, n+1)
	}

	for i := 1; i <= m; i++ {
		for j := 1; j <= n; j++ {
			if a[i-1] == b[j-1] {
				dp[i][j] = dp[i-1][j-1] + 1
			} else if dp[i-1][j] > dp[i][j-1] {
				dp[i][j] = dp[i-1][j]
			} else {
				dp[i][j] = dp[i][j-1]
			}
		}
	}

	return dp
}

func buildDiffs(oldLines, newLines []string, lcs [][]int) []LineDiff {
	var diffs []LineDiff

	i := len(oldLines)
	j := len(newLines)

	type action struct {
		diff LineDiff
	}

	var actions []action

	for i > 0 && j > 0 {
		if oldLines[i-1] == newLines[j-1] {
			actions = append(actions, action{LineDiff{OldLine: oldLines[i-1], NewLine: newLines[j-1], Type: "unchanged"}})
			i--
			j--
		} else if lcs[i-1][j] > lcs[i][j-1] {
			actions = append(actions, action{LineDiff{OldLine: oldLines[i-1], NewLine: "", Type: "removed"}})
			i--
		} else {
			actions = append(actions, action{LineDiff{OldLine: "", NewLine: newLines[j-1], Type: "added"}})
			j--
		}
	}

	for i > 0 {
		actions = append(actions, action{LineDiff{OldLine: oldLines[i-1], NewLine: "", Type: "removed"}})
		i--
	}

	for j > 0 {
		actions = append(actions, action{LineDiff{OldLine: "", NewLine: newLines[j-1], Type: "added"}})
		j--
	}

	for k := len(actions) - 1; k >= 0; k-- {
		diffs = append(diffs, actions[k].diff)
	}

	return diffs
}
