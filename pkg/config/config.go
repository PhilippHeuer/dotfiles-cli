package config

import (
	"os"
	"os/user"
	"slices"

	"github.com/cidverse/go-rules/pkg/expr"
	"github.com/rs/zerolog/log"
)

type DotfilesConfig struct {
	Directories []Dir `yaml:"directories"`
}

type Dir struct {
	Path   string   `yaml:"path"`
	Paths  []string `yaml:"paths"` // Can be used to specify multiple possible paths, first one that exists will be used.
	Target string   `yaml:"target"`
	Rules  []Rules  `yaml:"rules"` // At least one condition must match for the rule to apply
}

type Rules struct {
	Rule    string   `yaml:"rule"`
	Exclude []string `yaml:"exclude"` // Exclude paths or files
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
