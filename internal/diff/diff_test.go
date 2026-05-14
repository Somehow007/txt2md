package diff

import (
	"strings"
	"testing"
)

func TestCompareIdentical(t *testing.T) {
	text := "line1\nline2\nline3"
	result := Compare(text, text)

	if result.Summary.AddedLines != 0 || result.Summary.RemovedLines != 0 || result.Summary.ChangedLines != 0 {
		t.Errorf("Expected no changes for identical text, got added=%d, removed=%d, changed=%d",
			result.Summary.AddedLines, result.Summary.RemovedLines, result.Summary.ChangedLines)
	}
	if result.Summary.UnchangedLines != 3 {
		t.Errorf("Expected 3 unchanged lines, got %d", result.Summary.UnchangedLines)
	}
}

func TestCompareAdded(t *testing.T) {
	old := "line1\nline3"
	new := "line1\nline2\nline3"
	result := Compare(old, new)

	if result.Summary.AddedLines != 1 {
		t.Errorf("Expected 1 added line, got %d", result.Summary.AddedLines)
	}
}

func TestCompareRemoved(t *testing.T) {
	old := "line1\nline2\nline3"
	new := "line1\nline3"
	result := Compare(old, new)

	if result.Summary.RemovedLines != 1 {
		t.Errorf("Expected 1 removed line, got %d", result.Summary.RemovedLines)
	}
}

func TestCompareChanged(t *testing.T) {
	old := "line1\nold line\nline3"
	new := "line1\nnew line\nline3"
	result := Compare(old, new)

	if result.Summary.RemovedLines != 1 || result.Summary.AddedLines != 1 {
		t.Errorf("Expected 1 removed and 1 added line, got removed=%d, added=%d",
			result.Summary.RemovedLines, result.Summary.AddedLines)
	}
}

func TestUnified(t *testing.T) {
	old := "line1\nline2"
	new := "line1\nline3"
	result := Compare(old, new)

	unified := result.Unified()
	if !strings.Contains(unified, "- line2") {
		t.Error("Unified output should contain removed line")
	}
	if !strings.Contains(unified, "+ line3") {
		t.Error("Unified output should contain added line")
	}
}

func TestStats(t *testing.T) {
	old := "line1\nline2"
	new := "line1\nline3"
	result := Compare(old, new)

	stats := result.Stats()
	if !strings.Contains(stats, "Added:") || !strings.Contains(stats, "Removed:") {
		t.Errorf("Stats should contain change info, got: %s", stats)
	}
}

func TestCompareEmpty(t *testing.T) {
	result := Compare("", "")
	stats := result.Stats()
	if !strings.Contains(stats, "0") {
		t.Errorf("Expected no changes for empty input, got: %s", stats)
	}
}
