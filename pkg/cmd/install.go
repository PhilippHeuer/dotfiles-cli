package cmd

import (
	"log/slog"

	"github.com/PhilippHeuer/dotfiles-cli/pkg/dotfiles"
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

			dir := ""
			if len(args) == 1 && args[0] != "" {
				dir = args[0]
			}

			// install
			if err := dotfiles.Install(dir, mode, dryRun); err != nil {
				slog.Error("failed to install dotfiles", "err", err)
			}
		},
	}

	cmd.PersistentFlags().String("mode", "copy", "copy or symlink")
	cmd.PersistentFlags().BoolP("dry-run", "d", false, "dry run")

	return cmd
}
