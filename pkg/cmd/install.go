package cmd

import (
	"os"
	"path/filepath"

	"github.com/PhilippHeuer/dotfilessetup/pkg/config"
	"github.com/PhilippHeuer/dotfilessetup/pkg/util"
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
			if len(args) != 1 {
				log.Fatal().Msg("provide one source directory")
			}
			source := args[0]
			if source == "" {
				log.Fatal().Msg("source directory can not be empty")
			}

			// load config
			conf, err := config.Load(filepath.Join(source, "dotfiles.yaml"))
			if err != nil {
				log.Fatal().Err(err).Str("file", filepath.Join(source, "config.yaml")).Msg("failed to parse config file")
			}

			// information
			log.Info().Bool("dry-run", dryRun).Str("mode", mode).Str("source", source).Msg("installing dotfiles")

			// load state
			stateFile := config.StateFile()
			err = util.CreateParentDirectory(stateFile)
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

			// process directories
			for _, dir := range conf.Directories {
				fullPath := filepath.Join(source, dir.Path)
				targetPath := util.ResolvePath(dir.Target)
				log.Debug().Str("dir", fullPath).Str("target", targetPath).Msg("processing directory")

				// get all files in source
				files, filesErr := util.GetAllFiles(fullPath)
				if filesErr != nil {
					log.Fatal().Err(filesErr).Str("source", source).Msg("failed to get files")
				}

				// process files
				for _, file := range files {
					relativeFile, fileErr := filepath.Rel(fullPath, file)
					if fileErr != nil {
						log.Fatal().Err(fileErr).Str("source", file).Msg("failed to get relative file")
					}
					targetFile := filepath.Join(targetPath, relativeFile)

					// copy or link file
					linkErr := util.LinkFile(file, targetFile, dryRun, mode)
					if linkErr != nil {
						log.Fatal().Err(linkErr).Str("source", file).Str("target", targetFile).Msg("failed to link file")
					}

					// state
					managedFiles = append(managedFiles, targetFile)
				}
			}

			// save state
			state.ManagedFiles = managedFiles
			saveErr := config.SaveState(stateFile, state)
			if saveErr != nil {
				log.Fatal().Err(saveErr).Msg("failed to save state")
			}
		},
	}

	cmd.PersistentFlags().String("mode", "copy", "copy or symlink")
	cmd.PersistentFlags().BoolP("dry-run", "d", false, "dry run")

	return cmd
}
