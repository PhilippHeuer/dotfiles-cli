package cmd

import (
	"log/slog"
	"os"

	"github.com/PhilippHeuer/dotfiles-cli/pkg/config"
	"github.com/PhilippHeuer/dotfiles-cli/pkg/dotfiles"
	"github.com/PhilippHeuer/dotfiles-cli/pkg/util"
	"github.com/spf13/cobra"
)

func cleanCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "clean",
		Short: "clean dotfiles",
		Run: func(cmd *cobra.Command, args []string) {
			// properties
			dryRun, _ := cmd.Flags().GetBool("dry-run")

			// load state
			stateFile := config.StateFile()
			if err := util.CreateParentDirectory(stateFile); err != nil {
				slog.Error("failed to create state directory", "file", stateFile, "err", err)
				os.Exit(1)
			}
			state, err := config.LoadState(stateFile)
			if err != nil {
				slog.Error("failed to parse state file", "file", stateFile, "err", err)
				os.Exit(1)
			}

			// remove files
			state.ManagedFiles = dotfiles.DeleteManagedFiles(state.ManagedFiles, dryRun)

			// save state
			if !dryRun {
				if saveErr := config.SaveState(stateFile, state); saveErr != nil {
					slog.Error("failed to save state", "err", saveErr)
					os.Exit(1)
				}
			}
		},
	}

	cmd.PersistentFlags().BoolP("dry-run", "d", false, "dry run")

	return cmd
}
