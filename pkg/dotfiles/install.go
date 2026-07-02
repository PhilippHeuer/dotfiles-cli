package dotfiles

import (
	"errors"
	"log/slog"
	"os"
	"path/filepath"
	"slices"

	"github.com/PhilippHeuer/dotfiles-cli/pkg/config"
	"github.com/PhilippHeuer/dotfiles-cli/pkg/util"
	"github.com/cidverse/go-rules/pkg/expr"
	"github.com/iancoleman/strcase"
)

type File struct {
	Source         string
	Target         string
	IsTemplateFile bool
}

func Install(dir string, mode string, dryRun bool) error {
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

	// source dir (first arg or from state)
	var source string
	if dir != "" {
		source = dir
	} else if state.Source != "" {
		source = state.Source
	} else {
		slog.Error("provide the source directory as first argument")
		os.Exit(1)
	}
	state.Source = source

	// load config
	conf, err := config.Load(filepath.Join(source, "dotfiles.yaml"), true)
	if err != nil {
		slog.Error("failed to parse config file", "file", filepath.Join(source, "dotfiles.yaml"), "err", err)
		os.Exit(1)
	}

	// theme
	themeName := os.Getenv("DOTFILE_THEME")
	originalThemeName := state.Theme
	if themeName != "" {
		state.Theme = themeName
	} else {
		themeName = state.Theme
	}
	theme := conf.GetTheme(themeName)
	state.ActiveTheme = theme

	// information
	slog.Info("installing dotfiles", "dry-run", dryRun, "mode", mode, "source", source)

	// remove files
	state.ManagedFiles = DeleteManagedFiles(state.ManagedFiles, dryRun)

	// properties (built once, reused for all directories)
	properties := map[string]string{
		"Home": os.Getenv("HOME"),
		"User": os.Getenv("USER"),
	}
	if theme != nil {
		properties["Name"] = themeName
		properties["ColorScheme"] = theme.ColorScheme
		properties["WallpaperDir"] = theme.WallpaperDir
		properties["FontFamily"] = theme.FontFamily
		properties["FontSize"] = theme.FontSize
		properties["GtkTheme"] = theme.GtkTheme
		properties["IconTheme"] = theme.IconTheme
		properties["CursorTheme"] = theme.CursorTheme
		for k, v := range theme.Properties {
			properties[strcase.ToCamel(k)] = v
		}
	}

	// rule context (built once, reused for all files)
	ruleCtx := config.BuildRuleContext()

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
			slog.Info("source does not exist, skipping", "source", source, "err", filesErr)
			continue
		}

		// process files
		var filesToProcess []File
		for _, file := range files {
			relativeFile, fileErr := filepath.Rel(fullPath, file)
			if fileErr != nil {
				return errors.New("failed to determine relative file path for: " + file)
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
		if theme != nil && len(dir.ThemeFiles) > 0 {
			for _, tf := range dir.ThemeFiles {
				// use theme-specific source
				src := tf.Sources[theme.Name]
				if src == "" { // fallback to color scheme
					src = tf.Sources[theme.ColorScheme]
				}
				if src == "" { // fallback to first source
					for _, s := range tf.Sources {
						src = s
						break
					}
				}

				// skip if no source
				if src == "" {
					continue
				}

				// force template mode for designated files
				isTemplateFile := false
				if slices.Contains(dir.TemplateFiles, src) {
					isTemplateFile = true
				}

				// resolve full path if not absolute
				if !filepath.IsAbs(src) {
					src = filepath.Join(fullPath, src)
				}

				// append to files
				filesToProcess = append(filesToProcess, File{
					Source:         src,
					Target:         util.ResolvePath(tf.Target),
					IsTemplateFile: isTemplateFile,
				})
			}
		}

		// determine directory mode (dir config > global flag, template always wins)
		dirMode := mode
		if dir.Mode != "" {
			dirMode = dir.Mode
		}

		// process files
		for _, f := range filesToProcess {
			// skip if conditions do not match
			match := config.EvaluateRulesWithContext(ruleCtx, dir.Rules, f.Source)
			slog.Debug("processing file", "dir", f.Source, "target", f.Target, "condition-result", match)
			if !match {
				continue
			}

			// determine mode (template > dir config > global flag)
			fileMode := dirMode
			if f.IsTemplateFile {
				fileMode = "template"
			}

			// copy or link file
			if linkErr := util.LinkFile(f.Source, f.Target, dryRun, fileMode, properties); linkErr != nil {
				slog.Error("failed to link file", "source", f.Source, "target", f.Target, "err", linkErr)
				os.Exit(1)
			}
			slog.Debug("process file", "source", f.Source, "target", f.Target, "mode", fileMode)

			// state
			state.ManagedFiles = append(state.ManagedFiles, f.Target)
		}
	}

	// persist state (in case any of the commands query the state)
	if saveErr := config.SaveState(stateFile, state); saveErr != nil {
		slog.Error("failed to save state", "err", saveErr)
		os.Exit(1)
	}

	// theme activation
	if theme != nil && !dryRun {
		if err := activateTheme(theme, conf.Commands, originalThemeName); err != nil {
			slog.Error("failed to activate theme", "theme", themeName, "err", err)
			os.Exit(1)
		}
	}

	return nil
}

// activateTheme executes the theme activation commands, if available
func activateTheme(theme *config.ThemeConfig, activationCommands []config.ThemeCommand, originalThemeName string) error {
	for _, cmd := range append(activationCommands, theme.Commands...) {
		slog.Debug("executing theme command", "command", cmd.Command)

		if cmd.Condition != "" {
			match, err := expr.EvalBooleanExpression(cmd.Condition, map[string]interface{}{
				"env": os.Environ(),
			})
			if err != nil {
				slog.Warn("failed to evaluate theme activation command condition", "condition", cmd.Condition, "err", err)
				continue
			}

			if !match {
				continue
			}
		}
		if cmd.OnChange && originalThemeName == theme.Name {
			slog.Debug("command not executed, theme did not change", "command", cmd.Command)
			continue
		}

		if err := util.RunCommand(cmd.Command); err != nil {
			slog.Warn("failed to execute theme activation command", "command", cmd.Command, "err", err)
		}
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
