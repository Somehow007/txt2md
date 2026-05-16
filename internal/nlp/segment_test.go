package nlp

import "testing"

func TestBasicSegmentation(t *testing.T) {
	input := "This is the first paragraph.\nThis is the second paragraph."
	expected := "This is the first paragraph.\n\nThis is the second paragraph."
	result := Segment(input)
	if result != expected {
		t.Errorf("Basic segmentation failed.\nExpected:\n%q\nGot:\n%q", expected, result)
	}
}

func TestChineseTextSegmentation(t *testing.T) {
	input := "这是第一段内容。这是第二段内容。"
	expected := "这是第一段内容。\n\n这是第二段内容。"
	result := Segment(input)
	if result != expected {
		t.Errorf("Chinese text segmentation failed.\nExpected:\n%q\nGot:\n%q", expected, result)
	}
}

func TestMixedChineseEnglish(t *testing.T) {
	input := "This is English text.这是中文内容。"
	expected := "This is English text.\n\n这是中文内容。"
	result := Segment(input)
	if result != expected {
		t.Errorf("Mixed Chinese/English segmentation failed.\nExpected:\n%q\nGot:\n%q", expected, result)
	}
}

func TestAlreadyWellFormatted(t *testing.T) {
	input := "First paragraph.\n\nSecond paragraph."
	expected := "First paragraph.\n\nSecond paragraph."
	result := Segment(input)
	if result != expected {
		t.Errorf("Already well-formatted text should not change.\nExpected:\n%q\nGot:\n%q", expected, result)
	}
}

func TestEmptyInput(t *testing.T) {
	result := Segment("")
	if result != "" {
		t.Errorf("Empty input should return empty string, got: %q", result)
	}
}

func TestSingleParagraph(t *testing.T) {
	input := "This is a single paragraph that spans multiple lines\nwithout any clear paragraph boundary indicators\nso the text should remain unchanged"
	expected := input
	result := Segment(input)
	if result != expected {
		t.Errorf("Single paragraph should remain unchanged.\nExpected:\n%q\nGot:\n%q", expected, result)
	}
}

func TestIsSentenceEnd(t *testing.T) {
	tests := []struct {
		r        rune
		expected bool
	}{
		{'。', true},
		{'！', true},
		{'？', true},
		{'.', true},
		{'!', true},
		{'?', true},
		{',', false},
		{'a', false},
		{'，', false},
		{' ', false},
	}
	for _, tt := range tests {
		if got := isSentenceEnd(tt.r); got != tt.expected {
			t.Errorf("isSentenceEnd(%q) = %v, want %v", tt.r, got, tt.expected)
		}
	}
}

func TestIsParagraphStart(t *testing.T) {
	tests := []struct {
		text     string
		expected bool
	}{
		{"Hello", true},
		{"This is a test", true},
		{"hello", false},
		{"这是中文", true},
		{"第一章 开始", true},
		{"Chapter 1", true},
		{"chapter one", true},
		{"Section 2", true},
		{"Part III", true},
		{"", false},
		{"lowercase", false},
	}
	for _, tt := range tests {
		if got := isParagraphStart(tt.text); got != tt.expected {
			t.Errorf("isParagraphStart(%q) = %v, want %v", tt.text, got, tt.expected)
		}
	}
}

func TestSectionMarker(t *testing.T) {
	input := "Some introductory text here.\n第一章 故事开始\n这是章节内容。"
	result := Segment(input)
	if result == input {
		t.Errorf("Section marker should trigger paragraph break.\nInput:\n%q\nGot:\n%q", input, result)
	}
}

func TestIndentationChange(t *testing.T) {
	input := "First line with no indent\n    Indented line here"
	result := Segment(input)
	if result == input {
		t.Errorf("Indentation change should trigger paragraph break.\nInput:\n%q\nGot:\n%q", input, result)
	}
}

func TestShortLineAsHeading(t *testing.T) {
	input := "This is a long paragraph with substantial content that spans a good length.\nIntro\nThis is another long paragraph with substantial content that spans a good length."
	result := Segment(input)
	if result == input {
		t.Errorf("Short line between long lines should trigger paragraph breaks.\nInput:\n%q\nGot:\n%q", input, result)
	}
}
