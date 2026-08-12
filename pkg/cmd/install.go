package cmd

import (
	"log/slog"
	"strings"

	"github.com/PhilippHeuer/dotfiles-cli/pkg/dotfiles"
	"github.com/PhilippHeuer/dotfiles-cli/pkg/util"
	"github.com/spf13/cobra"
)

func installCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "install",
		Short: "install dotfiles",
		Run: func(cmd *cobra.Command, args []string) {
			// properties
			mode, _ := cmd.Flags().GetString("mode")
			dryRun, _ := cmd.Flags().GetBool("dry-run")
			theme, _ := cmd.Flags().GetString("theme")

			dir := ""
			if len(args) == 1 && args[0] != "" {
				dir = args[0]
			}

			// extra context from CLI flags
			extraContext := make(map[string]interface{})
			contextFile, _ := cmd.Flags().GetString("context-file")
			if contextFile != "" {
				fileCtx, err := util.LoadContextFile(util.ResolvePath(contextFile))
				if err != nil {
					slog.Error("failed to load context file", "file", contextFile, "err", err)
				} else {
					for k, v := range fileCtx {
						extraContext[k] = v
					}
				}
			}
			contextPairs, _ := cmd.Flags().GetStringSlice("context")
			for _, pair := range contextPairs {
				k, v, found := strings.Cut(pair, "=")
				if !found {
					slog.Warn("skipping malformed context pair", "value", pair)
					continue
				}
				extraContext[strings.TrimSpace(k)] = util.ParseContextValue(strings.TrimSpace(v))
			}

			// install
			if err := dotfiles.Install(dir, mode, dryRun, extraContext, theme); err != nil {
				slog.Error("failed to install dotfiles", "err", err)
			}
		},
	}

	cmd.PersistentFlags().String("mode", "copy", "copy or symlink")
	cmd.PersistentFlags().BoolP("dry-run", "d", false, "dry run")
	cmd.PersistentFlags().String("theme", "", "theme to install (overrides DOTFILE_THEME env var)")
	cmd.PersistentFlags().String("context-file", "", "path to a key=value context file")
	cmd.PersistentFlags().StringSlice("context", []string{}, "additional context key=value pairs")

	return cmd
}
