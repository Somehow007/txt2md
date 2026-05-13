package pipeline

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConvertBasic(t *testing.T) {
	input := `Hello World

This is a paragraph.

- Item 1
- Item 2
- Item 3`

	result, err := Convert(input, Options{})
	require.NoError(t, err)
	assert.Contains(t, result, "Hello World")
	assert.Contains(t, result, "This is a paragraph.")
	assert.Contains(t, result, "- Item 1")
}

func TestGoldenFiles(t *testing.T) {
	// Find testdata directory relative to this test file
	dir := "../../testdata"
	// Try to find the correct path
	paths := []string{
		"../../testdata",
		"testdata",
		filepath.Join("..", "..", "testdata"),
	}
	var found bool
	for _, p := range paths {
		if _, err := os.Stat(p); err == nil {
			dir = p
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("Could not find testdata directory")
	}
	files, err := os.ReadDir(dir)
	require.NoError(t, err)

	for _, f := range files {
		if !strings.HasPrefix(f.Name(), "input_") || !strings.HasSuffix(f.Name(), ".txt") {
			continue
		}

		t.Run(f.Name(), func(t *testing.T) {
			inputPath := filepath.Join(dir, f.Name())

			// Calculate golden file path: input_xxx.txt -> golden_xxx.md
			baseName := strings.TrimPrefix(f.Name(), "input_")
			baseName = strings.TrimSuffix(baseName, ".txt")
			goldenPath := filepath.Join(dir, "golden_"+baseName+".md")

			input, err := os.ReadFile(inputPath)
			require.NoError(t, err)

			expected, err := os.ReadFile(goldenPath)
			require.NoError(t, err)

			result, err := Convert(string(input), Options{})
			require.NoError(t, err)

			assert.Equal(t, string(expected), result)
		})
	}
}
