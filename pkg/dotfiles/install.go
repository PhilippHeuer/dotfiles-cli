package dotfiles

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"

	"github.com/PhilippHeuer/dotfiles-cli/pkg/config"
	"github.com/PhilippHeuer/dotfiles-cli/pkg/util"
	"github.com/adrg/xdg"
	"github.com/rs/zerolog/log"
)

type File struct {
	Source         string
	Target         string
	IsTemplateFile bool
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

	// load config
	conf, err := config.Load(filepath.Join(source, "dotfiles.yaml"))
	if err != nil {
		log.Fatal().Err(err).Str("file", filepath.Join(source, "config.yaml")).Msg("failed to parse config file")
	}

	// theme
	themeName := os.Getenv("DOTFILE_THEME")
	if themeName != "" {
		state.Theme = themeName
	} else {
		themeName = state.Theme
	}
	theme := conf.GetTheme(themeName)

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
	state.ManagedFiles = DeleteManagedFiles(state.ManagedFiles, dryRun)

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

			// force template mode for designated files
			isTemplateFile := false
			if slices.Contains(dir.TemplateFiles, filepath.Join(dir.Path, relativeFile)) {
				isTemplateFile = true
			}

			filesToProcess = append(filesToProcess, File{
				Source:         file,
				Target:         targetFile,
				IsTemplateFile: isTemplateFile,
			})
		}

		// theme-specific files
		if len(dir.ThemeFiles) > 0 {
			for _, tf := range dir.ThemeFiles {
				// use theme-specific source, fallback to first source if not available
				source := tf.Sources[themeName]
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

				// force template mode for designated files
				isTemplateFile := false
				if slices.Contains(dir.TemplateFiles, source) {
					isTemplateFile = true
				}

				// resolve full path if not absolute
				if !filepath.IsAbs(source) {
					source = filepath.Join(fullPath, source)
				}

				// append to files
				filesToProcess = append(filesToProcess, File{
					Source:         source,
					Target:         util.ResolvePath(tf.Target),
					IsTemplateFile: isTemplateFile,
				})
			}
		}

		// properties
		properties := map[string]string{}
		if theme != nil {
			properties = theme.Properties
		}

		// process files
		for _, f := range filesToProcess {
			// skip if conditions do not match
			match := config.EvaluateRules(dir.Rules, f.Source)
			log.Debug().Str("dir", f.Source).Str("target", f.Target).Bool("condition-result", match).Msg("processing file")
			if !match {
				continue
			}

			// determine mode
			fileMode := mode
			if f.IsTemplateFile {
				fileMode = "template"
			}

			// copy or link file
			linkErr := util.LinkFile(f.Source, f.Target, dryRun, fileMode, properties)
			if linkErr != nil {
				log.Fatal().Err(linkErr).Str("source", f.Source).Str("target", f.Target).Msg("failed to link file")
			}
			log.Trace().Str("source", f.Source).Str("target", f.Target).Str("mode", fileMode).Msg("process file")

			// state
			state.ManagedFiles = append(state.ManagedFiles, f.Target)
		}
	}

	// write state files
	err = writeStateFiles(state.Source, state.Theme)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to write state files")
	}

	// theme activation
	if theme != nil && !dryRun {
		err = activateTheme(theme)
		if err != nil {
			log.Fatal().Err(err).Str("theme", themeName).Msg("failed to activate theme")
		}
	}

	return nil
}

// activateTheme executes the theme activation commands, if available
func activateTheme(theme *config.ThemeConfig) error {
	for _, cmd := range theme.Commands {
		log.Debug().Str("command", cmd.Command).Msg("executing theme command")

		err := util.RunCommand(cmd.Command)
		if err != nil {
			return errors.Join(fmt.Errorf("failed to execute theme activation command: %s", cmd), err)
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
