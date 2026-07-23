package config

import (
	"os"
	"os/user"
	"slices"

	"github.com/cidverse/go-rules/pkg/expr"
	"log/slog"
)

type DotfilesConfig struct {
	Themes      []ThemeConfig  `yaml:"themes"`             // Themes defines theme-specific configurations
	Commands    []ThemeCommand `yaml:"activationCommands"` // Commands to run when a theme is activated
	Directories []Dir          `yaml:"directories"`        // Directories to copy
	Includes    []string       `yaml:"includes"`           // Include optional configuration files
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
	Path          string        `yaml:"path"`
	Paths         []string      `yaml:"paths"` // Can be used to specify multiple possible paths, first one that exists will be used.
	Target        string        `yaml:"target"`
	Mode          string        `yaml:"mode"`          // Override global mode for this directory (copy, symlink)
	Rules         []Rules       `yaml:"rules"`         // At least one condition must match for the rule to apply
	TemplateFiles []string      `yaml:"templateFiles"` // Files that need to be processed as templates, allowing the use of theme properties
	ThemeFiles    []ThemeFile   `yaml:"themeFiles"`    // Theme-specific files to copy
	LinkFiles     []LinkFile    `yaml:"linkFiles"`     // Individual file symlinks with fallback paths
}

type Rules struct {
	Rule    string   `yaml:"rule"`
	Exclude []string `yaml:"exclude"` // Exclude paths or files
}

type ThemeFile struct {
	Target  string            `yaml:"target"`
	Sources map[string]string `yaml:"sources"`
}

type LinkFile struct {
	Paths  []string `yaml:"paths"`  // Ordered list of source candidates (absolute or ~/ paths), first that exists wins
	Target string   `yaml:"target"` // Destination path (supports ~/ and env vars)
	Mode   string   `yaml:"mode"`   // Override global mode for this file (copy, symlink)
}

func EvaluateRules(conditions []Rules, sourceFile string) bool {
	return EvaluateRulesWithContext(BuildRuleContext(), conditions, sourceFile)
}

// RuleContext holds the shared context for rule evaluation (user, theme, wsl).
type RuleContext map[string]interface{}

// BuildRuleContext creates a shared rule context that can be reused across files.
// Only the "file" field changes per-file; everything else is constant.
// Available variables: user, home, hostname, theme, wsl
func BuildRuleContext() RuleContext {
	// user info
	var username, homeDir string
	currentUser, err := user.Current()
	if err != nil {
		slog.Debug("failed to get current user, falling back to env vars", "err", err)
		username = os.Getenv("USER")
		homeDir = os.Getenv("HOME")
	} else {
		username = currentUser.Username
		homeDir = currentUser.HomeDir
	}

	// hostname
	hostname, _ := os.Hostname()

	// context
	ctx := map[string]interface{}{
		"user":     username,
		"home":     homeDir,
		"hostname": hostname,
		"theme":    os.Getenv("DOTFILE_THEME"),
		"wsl":      os.Getenv("WSL_DISTRO_NAME") != "",
	}

	return RuleContext(ctx)
}

func EvaluateRulesWithContext(ctx RuleContext, conditions []Rules, sourceFile string) bool {
	if len(conditions) == 0 {
		return true
	}

	ctx["file"] = sourceFile

	// evaluate
	for _, c := range conditions {
		// excludes
		if slices.Contains(c.Exclude, sourceFile) {
			return false
		}

		// match expression
		match, cErr := expr.EvaluateRule(c.Rule, ctx)
		if cErr != nil {
			slog.Error("failed to evaluate condition, check your configuration file syntax", "rule", c.Rule, "err", cErr)
			os.Exit(1)
		}
		if match {
			return true
		}
	}

	return false
}
