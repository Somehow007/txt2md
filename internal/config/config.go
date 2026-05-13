package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

// Config holds the application configuration.
type Config struct {
	TabWidth int    `mapstructure:"tab-width"`
	Style    string `mapstructure:"style"`
	Pretty   bool   `mapstructure:"pretty"`
}

// DefaultConfig returns the default configuration.
func DefaultConfig() *Config {
	return &Config{
		TabWidth: 4,
		Style:    "spacious",
		Pretty:   false,
	}
}

// Load loads configuration from .txt2md.yaml files.
// Search order: ./.txt2md.yaml, $HOME/.txt2md.yaml
func Load() (*Config, error) {
	cfg := DefaultConfig()

	v := viper.New()
	v.SetDefault("tab-width", 4)
	v.SetDefault("style", "spacious")
	v.SetDefault("pretty", false)

	// Search in current directory
	v.AddConfigPath(".")
	v.SetConfigName(".txt2md")
	v.SetConfigType("yaml")

	// Also search in home directory
	if home, err := os.UserHomeDir(); err == nil {
		v.AddConfigPath(filepath.Join(home))
	}

	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("failed to read config: %w", err)
		}
		// Config file not found is OK, use defaults
	}

	if err := v.Unmarshal(cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	return cfg, nil
}

// MergeFlags merges CLI flags into config (CLI flags take precedence).
func (c *Config) MergeFlags(tabWidth int, pretty bool, style string) {
	if tabWidth != 4 { // 4 is the default, only override if user specified
		c.TabWidth = tabWidth
	}
	if pretty {
		c.Pretty = pretty
	}
	if style != "spacious" { // spacious is the default
		c.Style = style
	}
}
