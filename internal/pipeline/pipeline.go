package pipeline

import (
	"github.com/Somehow007/txt2md/internal/classifier"
	"github.com/Somehow007/txt2md/internal/classifier/rules"
	"github.com/Somehow007/txt2md/internal/renderer"
	"github.com/Somehow007/txt2md/internal/scanner"
)

// Options configures the conversion pipeline.
type Options struct {
	TabWidth int
	Pretty   bool
	Style    string
}

// Convert transforms plain text into Markdown.
func Convert(input string, opts Options) (string, error) {
	// Step 1: Scan - split into lines
	lines := scanner.Scan(input)

	// Step 2: Classify - detect block types
	engine := classifier.NewEngine(
		rules.Options{TabWidth: opts.TabWidth},
		&rules.HeadingRule{},
		&rules.ListRule{},
		&rules.CodeBlockRule{},
		&rules.BlockquoteRule{},
		&rules.ParagraphRule{},
	)
	blocks := engine.Classify(lines)

	// Step 3: Render - generate Markdown
	output := renderer.Render(blocks, renderer.Options{
		Pretty: opts.Pretty,
		Style:  opts.Style,
	})

	return output, nil
}
