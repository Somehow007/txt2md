package customrules

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/pelletier/go-toml/v2"
	"gopkg.in/yaml.v3"
)

type CustomRules struct {
	Heading HeadingRule  `yaml:"heading" toml:"heading"`
	List    ListRule     `yaml:"list" toml:"list"`
	Code    CodeRule     `yaml:"code" toml:"code"`
	Table   TableRule    `yaml:"table" toml:"table"`
	Pretty  PrettyRules  `yaml:"pretty" toml:"pretty"`
}

type HeadingRule struct {
	MinLength        int  `yaml:"min-length" toml:"min-length"`
	MaxLength        int  `yaml:"max-length" toml:"max-length"`
	DetectUppercase  bool `yaml:"detect-uppercase" toml:"detect-uppercase"`
	DetectNumbered   bool `yaml:"detect-numbered" toml:"detect-numbered"`
	DetectContextual bool `yaml:"detect-contextual" toml:"detect-contextual"`
}

type ListRule struct {
	Markers   []string `yaml:"markers" toml:"markers"`
	MaxIndent int      `yaml:"max-indent" toml:"max-indent"`
}

type CodeRule struct {
	MinIndent int `yaml:"min-indent" toml:"min-indent"`
	MinLines  int `yaml:"min-lines" toml:"min-lines"`
}

type TableRule struct {
	MinColumns int `yaml:"min-columns" toml:"min-columns"`
	MinRows    int `yaml:"min-rows" toml:"min-rows"`
}

type PrettyRules struct {
	CJKSpacing bool `yaml:"cjk-spacing" toml:"cjk-spacing"`
}

func DefaultCustomRules() *CustomRules {
	return &CustomRules{
		Heading: HeadingRule{
			MinLength:        3,
			MaxLength:        60,
			DetectUppercase:  true,
			DetectNumbered:   true,
			DetectContextual: true,
		},
		List: ListRule{
			Markers:   []string{"-", "*", "+", "•"},
			MaxIndent: 8,
		},
		Code: CodeRule{
			MinIndent: 4,
			MinLines:  2,
		},
		Table: TableRule{
			MinColumns: 2,
			MinRows:    2,
		},
		Pretty: PrettyRules{
			CJKSpacing: true,
		},
	}
}

func Load(path string) (*CustomRules, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read rules file: %w", err)
	}

	rules := DefaultCustomRules()

	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".yaml", ".yml":
		var m map[string]any
		if err := yaml.Unmarshal(data, &m); err != nil {
			return nil, fmt.Errorf("failed to parse YAML rules: %w", err)
		}
		applyMap(rules, m)
	case ".toml":
		var m map[string]any
		if err := toml.Unmarshal(data, &m); err != nil {
			return nil, fmt.Errorf("failed to parse TOML rules: %w", err)
		}
		applyMap(rules, m)
	default:
		return nil, fmt.Errorf("unsupported rules file format: %s", ext)
	}

	return rules, nil
}

func applyMap(rules *CustomRules, m map[string]any) {
	if v, ok := m["heading"]; ok {
		if sub, ok := v.(map[string]any); ok {
			applyHeadingMap(&rules.Heading, sub)
		}
	}
	if v, ok := m["list"]; ok {
		if sub, ok := v.(map[string]any); ok {
			applyListMap(&rules.List, sub)
		}
	}
	if v, ok := m["code"]; ok {
		if sub, ok := v.(map[string]any); ok {
			applyCodeMap(&rules.Code, sub)
		}
	}
	if v, ok := m["table"]; ok {
		if sub, ok := v.(map[string]any); ok {
			applyTableMap(&rules.Table, sub)
		}
	}
	if v, ok := m["pretty"]; ok {
		if sub, ok := v.(map[string]any); ok {
			applyPrettyMap(&rules.Pretty, sub)
		}
	}
}

func applyHeadingMap(h *HeadingRule, m map[string]any) {
	if v, ok := intFromMap(m, "min-length"); ok {
		h.MinLength = v
	}
	if v, ok := intFromMap(m, "max-length"); ok {
		h.MaxLength = v
	}
	if v, ok := boolFromMap(m, "detect-uppercase"); ok {
		h.DetectUppercase = v
	}
	if v, ok := boolFromMap(m, "detect-numbered"); ok {
		h.DetectNumbered = v
	}
	if v, ok := boolFromMap(m, "detect-contextual"); ok {
		h.DetectContextual = v
	}
}

func applyListMap(l *ListRule, m map[string]any) {
	if v, ok := stringSliceFromMap(m, "markers"); ok {
		l.Markers = v
	}
	if v, ok := intFromMap(m, "max-indent"); ok {
		l.MaxIndent = v
	}
}

func applyCodeMap(c *CodeRule, m map[string]any) {
	if v, ok := intFromMap(m, "min-indent"); ok {
		c.MinIndent = v
	}
	if v, ok := intFromMap(m, "min-lines"); ok {
		c.MinLines = v
	}
}

func applyTableMap(t *TableRule, m map[string]any) {
	if v, ok := intFromMap(m, "min-columns"); ok {
		t.MinColumns = v
	}
	if v, ok := intFromMap(m, "min-rows"); ok {
		t.MinRows = v
	}
}

func applyPrettyMap(p *PrettyRules, m map[string]any) {
	if v, ok := boolFromMap(m, "cjk-spacing"); ok {
		p.CJKSpacing = v
	}
}

func intFromMap(m map[string]any, key string) (int, bool) {
	v, ok := m[key]
	if !ok {
		return 0, false
	}
	switch n := v.(type) {
	case int:
		return n, true
	case int64:
		return int(n), true
	case float64:
		return int(n), true
	}
	return 0, false
}

func boolFromMap(m map[string]any, key string) (bool, bool) {
	v, ok := m[key]
	if !ok {
		return false, false
	}
	b, ok := v.(bool)
	return b, ok
}

func stringSliceFromMap(m map[string]any, key string) ([]string, bool) {
	v, ok := m[key]
	if !ok {
		return nil, false
	}
	slice, ok := v.([]any)
	if !ok {
		return nil, false
	}
	result := make([]string, 0, len(slice))
	for _, item := range slice {
		s, ok := item.(string)
		if !ok {
			return nil, false
		}
		result = append(result, s)
	}
	return result, true
}
