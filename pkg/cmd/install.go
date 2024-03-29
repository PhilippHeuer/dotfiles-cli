package cmd

import (
	"os"
	"path/filepath"

	"github.com/PhilippHeuer/dotfiles-cli/pkg/config"
	"github.com/PhilippHeuer/dotfiles-cli/pkg/util"
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

			// source dir (first arg or from state)
			var source string
			if len(args) == 1 && args[0] != "" {
				source = args[0]
			} else if len(args) == 0 && state.Source != "" {
				source = state.Source
			} else {
				log.Fatal().Msg("provide the source directory as first argument")
			}
			state.Source = source

			// load config
			conf, err := config.Load(filepath.Join(source, "dotfiles.yaml"))
			if err != nil {
				log.Fatal().Err(err).Str("file", filepath.Join(source, "config.yaml")).Msg("failed to parse config file")
			}

			// information
			log.Info().Bool("dry-run", dryRun).Str("mode", mode).Str("source", source).Msg("installing dotfiles")

			// save state on exit or error
			defer func() {
				saveErr := config.SaveState(stateFile, state)
				if saveErr != nil {
					log.Fatal().Err(saveErr).Msg("failed to save state")
				}
			}()

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
				fullPath := calculateFullPath(source, dir.Path)
				targetPath := util.ResolvePath(dir.Target)

				// check alternative paths
				if _, err := os.Stat(fullPath); os.IsNotExist(err) {
					for _, p := range dir.Paths {
						fp := calculateFullPath(source, p)
						if _, err := os.Stat(fp); !os.IsNotExist(err) {
							fullPath = fp
							break
						}
					}
				}

				// get all files in source
				files, filesErr := util.GetAllFiles(fullPath)
				if filesErr != nil {
					log.Info().Err(filesErr).Str("source", source).Msg("source does not exist, skipping")
					continue
				}

				// process files
				for _, file := range files {
					relativeFile, fileErr := filepath.Rel(fullPath, file)
					if fileErr != nil {
						log.Fatal().Err(fileErr).Str("source", file).Msg("failed to get relative file")
					}
					targetFile := filepath.Join(targetPath, relativeFile)

					// skip if conditions do not match
					match := config.EvaluateRules(dir.Rules, file)
					log.Debug().Str("dir", fullPath).Str("target", targetPath).Bool("condition-result", match).Msg("processing directory")
					if !match {
						continue
					}

					// copy or link file
					linkErr := util.LinkFile(file, targetFile, dryRun, mode)
					if linkErr != nil {
						log.Fatal().Err(linkErr).Str("source", file).Str("target", targetFile).Msg("failed to link file")
					}
					log.Trace().Str("source", file).Str("target", targetFile).Str("mode", mode).Msg("process file")

					// state
					state.ManagedFiles = append(state.ManagedFiles, targetFile)
				}
			}
		},
	}

	cmd.PersistentFlags().String("mode", "copy", "copy or symlink")
	cmd.PersistentFlags().BoolP("dry-run", "d", false, "dry run")

	return cmd
}

func calculateFullPath(source string, path string) string {
	fullPath := path
	if !filepath.IsAbs(path) && path != "" && path[0] != filepath.Separator {
		fullPath = filepath.Join(source, path)
	}
	return fullPath
}
