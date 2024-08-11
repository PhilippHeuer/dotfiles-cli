package dotfiles

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/PhilippHeuer/dotfiles-cli/pkg/config"
	"github.com/PhilippHeuer/dotfiles-cli/pkg/util"
	"github.com/adrg/xdg"
	"github.com/rs/zerolog/log"
)

type File struct {
	Source string
	Target string
}

func Install(dir string, mode string, dryRun bool) error {
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
	if dir != "" {
		source = dir
	} else if state.Source != "" {
		source = state.Source
	} else {
		log.Fatal().Msg("provide the source directory as first argument")
	}
	state.Source = source

	// theme
	theme := os.Getenv("DOTFILE_THEME")
	if theme != "" {
		state.Theme = theme
	} else {
		theme = state.Theme
	}

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
		var filesToProcess []File
		for _, file := range files {
			relativeFile, fileErr := filepath.Rel(fullPath, file)
			if fileErr != nil {
				return errors.New("failed to determinate relative file path for: " + file)
			}
			targetFile := filepath.Join(targetPath, relativeFile)

			filesToProcess = append(filesToProcess, File{
				Source: file,
				Target: targetFile,
			})
		}

		// theme-specific files
		if len(dir.ThemeFiles) > 0 {
			for _, tf := range dir.ThemeFiles {
				// use theme-specific source, fallback to first source if not available
				source := tf.Sources[theme]
				if source == "" {
					for _, src := range tf.Sources {
						source = src
						break
					}
				}

				// skip if no source
				if source == "" {
					continue
				}

				// resolve full path if not absolute
				if !filepath.IsAbs(source) {
					source = filepath.Join(fullPath, source)
				}

				// append to files
				filesToProcess = append(filesToProcess, File{
					Source: source,
					Target: util.ResolvePath(tf.Target),
				})
			}
		}

		// process files
		for _, f := range filesToProcess {
			// skip if conditions do not match
			match := config.EvaluateRules(dir.Rules, f.Source)
			log.Debug().Str("dir", f.Source).Str("target", f.Target).Bool("condition-result", match).Msg("processing file")
			if !match {
				continue
			}

			// copy or link file
			linkErr := util.LinkFile(f.Source, f.Target, dryRun, mode)
			if linkErr != nil {
				log.Fatal().Err(linkErr).Str("source", f.Source).Str("target", f.Target).Msg("failed to link file")
			}
			log.Trace().Str("source", f.Source).Str("target", f.Target).Str("mode", mode).Msg("process file")

			// state
			state.ManagedFiles = append(state.ManagedFiles, f.Target)
		}
	}

	// theme activation
	if state.Theme != "" {
		var themeCommands []config.ThemeCommand
		for _, theme := range conf.Themes {
			if theme.Name == state.Theme { // Assuming 'Name' is the field that identifies the theme
				themeCommands = theme.Commands
				break
			}
		}

		err = writeStateFiles(state.Source, state.Theme)
		if err != nil {
			log.Fatal().Err(err).Msg("failed to write state files")
		}

		for _, cmd := range themeCommands {
			log.Debug().Str("command", cmd.Command).Msg("executing theme command")
			if !dryRun {
				err := util.RunCommand(cmd.Command)
				if err != nil {
					log.Warn().Err(err).Str("command", cmd.Command).Msg("failed to execute theme activation command")
				}
			}
		}
	}

	return nil
}

func writeStateFiles(sourceDir string, theme string) error {
	// create state directory
	stateDir := filepath.Join(xdg.StateHome, "dotfiles")
	err := util.CreateParentDirectory(stateDir)
	if err != nil {
		return err
	}

	// write source file
	sourceFile := filepath.Join(stateDir, "source-dir")
	err = os.WriteFile(sourceFile, []byte(sourceDir), 0644)
	if err != nil {
		return err
	}

	// write theme file
	themeFile := filepath.Join(stateDir, "current-theme")
	err = os.WriteFile(themeFile, []byte(theme), 0644)
	if err != nil {
		return err
	}

	return nil
}

func calculateFullPath(source string, path string) string {
	fullPath := path
	if !filepath.IsAbs(path) && path != "" && path[0] != filepath.Separator {
		fullPath = filepath.Join(source, path)
	}
	return fullPath
}
