package customrules

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultCustomRules(t *testing.T) {
	rules := DefaultCustomRules()

	assert.Equal(t, 3, rules.Heading.MinLength)
	assert.Equal(t, 60, rules.Heading.MaxLength)
	assert.True(t, rules.Heading.DetectUppercase)
	assert.True(t, rules.Heading.DetectNumbered)
	assert.True(t, rules.Heading.DetectContextual)

	assert.Equal(t, []string{"-", "*", "+", "•"}, rules.List.Markers)
	assert.Equal(t, 8, rules.List.MaxIndent)

	assert.Equal(t, 4, rules.Code.MinIndent)
	assert.Equal(t, 2, rules.Code.MinLines)

	assert.Equal(t, 2, rules.Table.MinColumns)
	assert.Equal(t, 2, rules.Table.MinRows)

	assert.True(t, rules.Pretty.CJKSpacing)
}

func TestLoadFromYAML(t *testing.T) {
	content := `heading:
  min-length: 5
  max-length: 80
  detect-uppercase: false
  detect-numbered: true
  detect-contextual: true
list:
  markers:
    - "-"
    - "*"
    - ">"
  max-indent: 12
code:
  min-indent: 2
  min-lines: 3
table:
  min-columns: 3
  min-rows: 1
pretty:
  cjk-spacing: false
`
	dir := t.TempDir()
	path := filepath.Join(dir, "rules.yaml")
	err := os.WriteFile(path, []byte(content), 0644)
	require.NoError(t, err)

	rules, err := Load(path)
	require.NoError(t, err)

	assert.Equal(t, 5, rules.Heading.MinLength)
	assert.Equal(t, 80, rules.Heading.MaxLength)
	assert.False(t, rules.Heading.DetectUppercase)
	assert.True(t, rules.Heading.DetectNumbered)
	assert.True(t, rules.Heading.DetectContextual)

	assert.Equal(t, []string{"-", "*", ">"}, rules.List.Markers)
	assert.Equal(t, 12, rules.List.MaxIndent)

	assert.Equal(t, 2, rules.Code.MinIndent)
	assert.Equal(t, 3, rules.Code.MinLines)

	assert.Equal(t, 3, rules.Table.MinColumns)
	assert.Equal(t, 1, rules.Table.MinRows)

	assert.False(t, rules.Pretty.CJKSpacing)
}

func TestLoadFromYML(t *testing.T) {
	content := `heading:
  min-length: 10
list:
  max-indent: 6
`
	dir := t.TempDir()
	path := filepath.Join(dir, "rules.yml")
	err := os.WriteFile(path, []byte(content), 0644)
	require.NoError(t, err)

	rules, err := Load(path)
	require.NoError(t, err)

	assert.Equal(t, 10, rules.Heading.MinLength)
	assert.Equal(t, 6, rules.List.MaxIndent)
}

func TestLoadFromTOML(t *testing.T) {
	content := `[heading]
min-length = 5
max-length = 80
detect-uppercase = false
detect-numbered = true
detect-contextual = true

[list]
markers = ["-", "*", ">"]
max-indent = 12

[code]
min-indent = 2
min-lines = 3

[table]
min-columns = 3
min-rows = 1

[pretty]
cjk-spacing = false
`
	dir := t.TempDir()
	path := filepath.Join(dir, "rules.toml")
	err := os.WriteFile(path, []byte(content), 0644)
	require.NoError(t, err)

	rules, err := Load(path)
	require.NoError(t, err)

	assert.Equal(t, 5, rules.Heading.MinLength)
	assert.Equal(t, 80, rules.Heading.MaxLength)
	assert.False(t, rules.Heading.DetectUppercase)
	assert.True(t, rules.Heading.DetectNumbered)
	assert.True(t, rules.Heading.DetectContextual)

	assert.Equal(t, []string{"-", "*", ">"}, rules.List.Markers)
	assert.Equal(t, 12, rules.List.MaxIndent)

	assert.Equal(t, 2, rules.Code.MinIndent)
	assert.Equal(t, 3, rules.Code.MinLines)

	assert.Equal(t, 3, rules.Table.MinColumns)
	assert.Equal(t, 1, rules.Table.MinRows)

	assert.False(t, rules.Pretty.CJKSpacing)
}

func TestLoadEmptyYAML(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "rules.yaml")
	err := os.WriteFile(path, []byte(""), 0644)
	require.NoError(t, err)

	rules, err := Load(path)
	require.NoError(t, err)

	defaults := DefaultCustomRules()
	assert.Equal(t, defaults.Heading.MinLength, rules.Heading.MinLength)
	assert.Equal(t, defaults.Heading.MaxLength, rules.Heading.MaxLength)
	assert.Equal(t, defaults.Heading.DetectUppercase, rules.Heading.DetectUppercase)
	assert.Equal(t, defaults.Heading.DetectNumbered, rules.Heading.DetectNumbered)
	assert.Equal(t, defaults.Heading.DetectContextual, rules.Heading.DetectContextual)
	assert.Equal(t, defaults.List.Markers, rules.List.Markers)
	assert.Equal(t, defaults.List.MaxIndent, rules.List.MaxIndent)
	assert.Equal(t, defaults.Code.MinIndent, rules.Code.MinIndent)
	assert.Equal(t, defaults.Code.MinLines, rules.Code.MinLines)
	assert.Equal(t, defaults.Table.MinColumns, rules.Table.MinColumns)
	assert.Equal(t, defaults.Table.MinRows, rules.Table.MinRows)
	assert.Equal(t, defaults.Pretty.CJKSpacing, rules.Pretty.CJKSpacing)
}

func TestLoadEmptyTOML(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "rules.toml")
	err := os.WriteFile(path, []byte(""), 0644)
	require.NoError(t, err)

	rules, err := Load(path)
	require.NoError(t, err)

	defaults := DefaultCustomRules()
	assert.Equal(t, defaults.Heading.MinLength, rules.Heading.MinLength)
	assert.Equal(t, defaults.List.Markers, rules.List.Markers)
	assert.Equal(t, defaults.Code.MinIndent, rules.Code.MinIndent)
	assert.Equal(t, defaults.Table.MinColumns, rules.Table.MinColumns)
	assert.Equal(t, defaults.Pretty.CJKSpacing, rules.Pretty.CJKSpacing)
}

func TestLoadMissingFieldsYAML(t *testing.T) {
	content := `heading:
  min-length: 10
`
	dir := t.TempDir()
	path := filepath.Join(dir, "rules.yaml")
	err := os.WriteFile(path, []byte(content), 0644)
	require.NoError(t, err)

	rules, err := Load(path)
	require.NoError(t, err)

	assert.Equal(t, 10, rules.Heading.MinLength)
	assert.Equal(t, 60, rules.Heading.MaxLength)
	assert.True(t, rules.Heading.DetectUppercase)
	assert.True(t, rules.Heading.DetectNumbered)
	assert.True(t, rules.Heading.DetectContextual)
	assert.Equal(t, []string{"-", "*", "+", "•"}, rules.List.Markers)
	assert.Equal(t, 8, rules.List.MaxIndent)
	assert.Equal(t, 4, rules.Code.MinIndent)
	assert.Equal(t, 2, rules.Code.MinLines)
	assert.Equal(t, 2, rules.Table.MinColumns)
	assert.Equal(t, 2, rules.Table.MinRows)
	assert.True(t, rules.Pretty.CJKSpacing)
}

func TestLoadMissingFieldsTOML(t *testing.T) {
	content := `[heading]
min-length = 10
`
	dir := t.TempDir()
	path := filepath.Join(dir, "rules.toml")
	err := os.WriteFile(path, []byte(content), 0644)
	require.NoError(t, err)

	rules, err := Load(path)
	require.NoError(t, err)

	assert.Equal(t, 10, rules.Heading.MinLength)
	assert.Equal(t, 60, rules.Heading.MaxLength)
	assert.Equal(t, []string{"-", "*", "+", "•"}, rules.List.Markers)
	assert.Equal(t, 8, rules.List.MaxIndent)
	assert.Equal(t, 4, rules.Code.MinIndent)
	assert.Equal(t, 2, rules.Code.MinLines)
	assert.Equal(t, 2, rules.Table.MinColumns)
	assert.Equal(t, 2, rules.Table.MinRows)
}

func TestLoadInvalidFile(t *testing.T) {
	_, err := Load("/nonexistent/path/rules.yaml")
	assert.Error(t, err)
}

func TestLoadUnsupportedExtension(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "rules.json")
	err := os.WriteFile(path, []byte("{}"), 0644)
	require.NoError(t, err)

	_, err = Load(path)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported rules file format")
}

func TestLoadInvalidYAML(t *testing.T) {
	content := `heading:
  min-length: [invalid
`
	dir := t.TempDir()
	path := filepath.Join(dir, "rules.yaml")
	err := os.WriteFile(path, []byte(content), 0644)
	require.NoError(t, err)

	_, err = Load(path)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse YAML rules")
}

func TestLoadInvalidTOML(t *testing.T) {
	content := `[heading
min-length = invalid
`
	dir := t.TempDir()
	path := filepath.Join(dir, "rules.toml")
	err := os.WriteFile(path, []byte(content), 0644)
	require.NoError(t, err)

	_, err = Load(path)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse TOML rules")
}

func TestLoadPartialFieldsYAML(t *testing.T) {
	content := `code:
  min-lines: 5
pretty:
  cjk-spacing: true
`
	dir := t.TempDir()
	path := filepath.Join(dir, "rules.yaml")
	err := os.WriteFile(path, []byte(content), 0644)
	require.NoError(t, err)

	rules, err := Load(path)
	require.NoError(t, err)

	assert.Equal(t, 3, rules.Heading.MinLength)
	assert.Equal(t, 60, rules.Heading.MaxLength)
	assert.Equal(t, 4, rules.Code.MinIndent)
	assert.Equal(t, 5, rules.Code.MinLines)
	assert.True(t, rules.Pretty.CJKSpacing)
}

func TestLoadPartialFieldsTOML(t *testing.T) {
	content := `[code]
min-lines = 5

[pretty]
cjk-spacing = true
`
	dir := t.TempDir()
	path := filepath.Join(dir, "rules.toml")
	err := os.WriteFile(path, []byte(content), 0644)
	require.NoError(t, err)

	rules, err := Load(path)
	require.NoError(t, err)

	assert.Equal(t, 3, rules.Heading.MinLength)
	assert.Equal(t, 60, rules.Heading.MaxLength)
	assert.Equal(t, 4, rules.Code.MinIndent)
	assert.Equal(t, 5, rules.Code.MinLines)
	assert.True(t, rules.Pretty.CJKSpacing)
}
