package cmd

import (
	"os"

	"github.com/PhilippHeuer/dotfiles-cli/pkg/config"
	"github.com/PhilippHeuer/dotfiles-cli/pkg/util"
	"github.com/rs/zerolog/log"
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
			err := util.CreateParentDirectory(stateFile)
			if err != nil {
				log.Fatal().Err(err).Str("file", stateFile).Msg("failed to create state directory")
			}
			state, err := config.LoadState(stateFile)
			if err != nil {
				log.Fatal().Err(err).Str("file", stateFile).Msg("failed to parse state file")
			}

			// remove files
			var managedFiles []string
			for _, file := range state.ManagedFiles {
				log.Debug().Str("file", file).Msg("removing file")
				if dryRun {
					continue
				}

				// check if file exists
				if _, err := os.Stat(file); os.IsNotExist(err) {
					log.Trace().Str("file", file).Msg("file does not exist, already deleted")
					continue
				}

				// delete file
				removeErr := os.Remove(file)
				if removeErr != nil {
					managedFiles = append(managedFiles, file)
					log.Warn().Str("file", file).Msg("failed to remove file")
				}
			}
			state.ManagedFiles = managedFiles

			// save state
			if !dryRun {
				saveErr := config.SaveState(stateFile, state)
				if saveErr != nil {
					log.Fatal().Err(saveErr).Msg("failed to save state")
				}
			}
		},
	}

	cmd.PersistentFlags().BoolP("dry-run", "d", false, "dry run")

	return cmd
}
