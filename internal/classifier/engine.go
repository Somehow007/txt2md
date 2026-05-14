package classifier

import (
	"github.com/Somehow007/txt2md/internal/classifier/rules"
	"github.com/Somehow007/txt2md/internal/scanner"
)

type Engine struct {
	rules []rules.Rule
	opts  rules.Options
}

func NewEngine(opts rules.Options, ruleList ...rules.Rule) *Engine {
	return &Engine{
		rules: ruleList,
		opts:  opts,
	}
}

func (e *Engine) Classify(lines []scanner.Line) []scanner.Block {
	var blocks []scanner.Block
	idx := 0

	for idx < len(lines) {
		if lines[idx].IsEmpty {
			idx++
			continue
		}

		var matched bool
		for _, rule := range e.rules {
			block, consumed := rule.Detect(lines, idx, e.opts)
			if block != nil {
				blocks = append(blocks, *block)
				idx += consumed
				matched = true
				break
			}
		}

		if !matched {
			paraLines := []scanner.Line{lines[idx]}
			j := idx + 1
			for j < len(lines) && !lines[j].IsEmpty {
				paraLines = append(paraLines, lines[j])
				j++
			}
			blocks = append(blocks, scanner.Block{
				Type:       scanner.BlockParagraph,
				Lines:      paraLines,
				Confidence: 0.5,
			})
			idx = j
		}
	}

	return blocks
}
