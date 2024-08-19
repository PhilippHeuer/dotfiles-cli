package config

import (
	"os"
	"os/user"
	"slices"

	"github.com/cidverse/go-rules/pkg/expr"
	"github.com/rs/zerolog/log"
)

type DotfilesConfig struct {
	Themes      []ThemeConfig  `yaml:"themes"`             // Themes defines theme-specific configurations
	Commands    []ThemeCommand `yaml:"activationCommands"` // Commands to run when a theme is activated
	Directories []Dir          `yaml:"directories"`        // Directories to copy
}

func (c *DotfilesConfig) GetTheme(name string) *ThemeConfig {
	for _, t := range c.Themes {
		if t.Name == name {
			return &t
		}
	}
	return nil
}

type ThemeConfig struct {
	Name         string            `yaml:"name"`
	ColorScheme  string            `yaml:"colorScheme"`
	WallpaperDir string            `yaml:"wallpaperDir"`
	FontFamily   string            `yaml:"fontFamily"`
	FontSize     string            `yaml:"fontSize"`
	CosmicTheme  string            `yaml:"cosmicTheme"`
	GtkTheme     string            `yaml:"gtkTheme"`
	IconTheme    string            `yaml:"iconTheme"`
	CursorTheme  string            `yaml:"cursorTheme"`
	Properties   map[string]string `yaml:"properties"`
	Commands     []ThemeCommand    `yaml:"commands"`
}

type ThemeCommand struct {
	Command   string `yaml:"command"`
	OnChange  bool   `yaml:"onChange"`
	Condition string `yaml:"condition"`
}

type Dir struct {
	Path          string      `yaml:"path"`
	Paths         []string    `yaml:"paths"` // Can be used to specify multiple possible paths, first one that exists will be used.
	Target        string      `yaml:"target"`
	Rules         []Rules     `yaml:"rules"`         // At least one condition must match for the rule to apply
	TemplateFiles []string    `yaml:"templateFiles"` // Files that need to be processed as templates, allowing the use of theme properties
	ThemeFiles    []ThemeFile `yaml:"themeFiles"`    // Theme-specific files to copy
}

type Rules struct {
	Rule    string   `yaml:"rule"`
	Exclude []string `yaml:"exclude"` // Exclude paths or files
}

type ThemeFile struct {
	Target  string            `yaml:"target"`
	Sources map[string]string `yaml:"sources"`
}

func EvaluateRules(conditions []Rules, sourceFile string) bool {
	if len(conditions) == 0 {
		return true
	}

	// user
	currentUser, err := user.Current()
	if err != nil {
		log.Fatal().Err(err).Msg("failed to get current user")
	}

	// context
	ctx := map[string]interface{}{
		"user":  currentUser.Username,
		"theme": os.Getenv("DOTFILE_THEME"),
		"file":  sourceFile,
	}

	// wsl distro
	wslDistro := os.Getenv("WSL_DISTRO_NAME")
	if wslDistro != "" {
		ctx["wsl"] = true
	} else {
		ctx["wsl"] = false
	}

	// evaluate
	for _, c := range conditions {
		// excludes
		if slices.Contains(c.Exclude, sourceFile) {
			return false
		}

		// match expression
		match, cErr := expr.EvaluateRule(c.Rule, ctx)
		if cErr != nil {
			log.Fatal().Err(cErr).Str("rule", c.Rule).Msg("failed to evaluate condition, check your configuration file syntax")
		}
		if match {
			return true
		}
	}

	return false
}
