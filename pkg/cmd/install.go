package cmd

import (
	"github.com/PhilippHeuer/dotfiles-cli/pkg/dotfiles"
	"github.com/rs/zerolog/log"
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
			err := dotfiles.Install(dir, mode, dryRun)
			if err != nil {
				log.Error().Err(err).Msg("failed to install dotfiles")
			}
		},
	}

	cmd.PersistentFlags().String("mode", "copy", "copy or symlink")
	cmd.PersistentFlags().BoolP("dry-run", "d", false, "dry run")

	return cmd
}
