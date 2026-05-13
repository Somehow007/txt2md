package renderer

import "regexp"

// Beautify applies typography improvements to text.
func Beautify(text string) string {
	// Chinese/English spacing
	text = addSpacesBetweenCJKAndLatin(text)
	return text
}

// CJK character ranges
var (
	cjkChar    = regexp.MustCompile(`[\p{Han}\p{Hangul}\p{Hiragana}\p{Katakana}]`)
	notCjkChar = regexp.MustCompile(`[^\p{Han}\p{Hangul}\p{Hiragana}\p{Katakana}\s]`)
)

func addSpacesBetweenCJKAndLatin(text string) string {
	result := []rune(text)
	var newResult []rune

	for i := 0; i < len(result); i++ {
		newResult = append(newResult, result[i])

		if i+1 < len(result) {
			curr := result[i]
			next := result[i+1]

			// CJK followed by Latin/number
			if isCJK(curr) && isLatinOrNumber(next) {
				newResult = append(newResult, ' ')
			}
			// Latin/number followed by CJK
			if isLatinOrNumber(curr) && isCJK(next) {
				newResult = append(newResult, ' ')
			}
		}
	}

	return string(newResult)
}

func isCJK(r rune) bool {
	return (r >= 0x4E00 && r <= 0x9FFF) || // CJK Unified Ideographs
		(r >= 0x3040 && r <= 0x30FF) || // Japanese
		(r >= 0xAC00 && r <= 0xD7AF) // Korean
}

func isLatinOrNumber(r rune) bool {
	return (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9')
}
