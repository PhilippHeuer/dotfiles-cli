package cmd

import (
	"os"
	"strings"

	"github.com/cidverse/cidverseutils/zerologconfig"
	"github.com/spf13/cobra"
)

var (
	cfg = struct {
		LogLevel  string
		LogFormat string
		LogCaller bool
	}{}
	validLogLevels  = []string{"trace", "debug", "info", "warn", "error"}
	validLogFormats = []string{"plain", "color", "json"}
)

func rootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   `tms`,
		Short: `scans source directories for projects to create tmux sessions`,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			zerologconfig.Configure(cfg)
		},
		Args: cobra.MinimumNArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			_ = cmd.Help()
			os.Exit(0)
		},
	}

	cmd.PersistentFlags().StringVar(&cfg.LogLevel, "log-level", "info", "log level - allowed: "+strings.Join(validLogLevels, ","))
	cmd.PersistentFlags().StringVar(&cfg.LogFormat, "log-format", "color", "log format - allowed: "+strings.Join(validLogFormats, ","))
	cmd.PersistentFlags().BoolVar(&cfg.LogCaller, "log-caller", false, "include caller in log functions")

	cmd.AddCommand(installCmd())
	cmd.AddCommand(cleanCmd())
	cmd.AddCommand(queryCmd())
	cmd.AddCommand(versionCmd())

	return cmd
}

// Execute executes the root command.
func Execute() error {
	return rootCmd().Execute()
}
