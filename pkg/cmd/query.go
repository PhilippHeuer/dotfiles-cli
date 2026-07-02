package cmd

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"text/tabwriter"

	"github.com/PhilippHeuer/dotfiles-cli/pkg/config"
	"github.com/PhilippHeuer/dotfiles-cli/pkg/util"
	"github.com/iancoleman/strcase"
	"github.com/spf13/cobra"
)

func queryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "query",
		Short: "query the application config or state",
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 {
				slog.Error("provide a key to query")
				os.Exit(1)
			}

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

			// load config
			conf, err := config.Load(filepath.Join(state.Source, "dotfiles.yaml"), true)
			if err != nil {
				slog.Error("failed to parse config file", "file", filepath.Join(state.Source, "config.yaml"), "err", err)
				os.Exit(1)
			}

			// check if the key is a property
			key := strings.ToLower(strcase.ToCamel(args[0]))
			switch key {
			case "themes":
				w := tabwriter.NewWriter(os.Stdout, 1, 1, 1, ' ', 0)
				for _, theme := range conf.Themes {
					_, _ = fmt.Fprintln(w, theme.Name)
				}
				_ = w.Flush()
			case "themeoverview":
				requireActiveTheme(state)
				at := state.ActiveTheme
				if len(args) == 2 && args[1] != "" {
					at = conf.GetTheme(args[1])
				}

				w := tabwriter.NewWriter(os.Stdout, 1, 1, 1, ' ', 0)
				_, _ = fmt.Fprintln(w, "Name\t"+at.Name)
				_, _ = fmt.Fprintln(w, "ColorScheme\t"+at.ColorScheme)
				_, _ = fmt.Fprintln(w, "WallpaperDir\t"+at.WallpaperDir)
				_, _ = fmt.Fprintln(w, "FontFamily\t"+at.FontFamily)
				_, _ = fmt.Fprintln(w, "FontSize\t"+at.FontSize)
				_, _ = fmt.Fprintln(w, "GtkTheme\t"+at.GtkTheme)
				_, _ = fmt.Fprintln(w, "IconTheme\t"+at.IconTheme)
				_, _ = fmt.Fprintln(w, "CursorTheme\t"+at.CursorTheme)
				_ = w.Flush()
			case "source":
				fmt.Println(state.Source)
			case "theme":
				requireActiveTheme(state)
				fmt.Println(state.ActiveTheme.Name)
			case "colorscheme":
				requireActiveTheme(state)
				fmt.Println(state.ActiveTheme.ColorScheme)
			case "wallpaperdir":
				requireActiveTheme(state)
				fmt.Println(util.ResolvePath(state.ActiveTheme.WallpaperDir))
			case "fontfamily":
				requireActiveTheme(state)
				fmt.Println(state.ActiveTheme.FontFamily)
			case "fontsize":
				requireActiveTheme(state)
				fmt.Println(state.ActiveTheme.FontSize)
			case "gtktheme":
				requireActiveTheme(state)
				fmt.Println(state.ActiveTheme.GtkTheme)
			case "icontheme":
				requireActiveTheme(state)
				fmt.Println(state.ActiveTheme.IconTheme)
			case "cursortheme":
				requireActiveTheme(state)
				fmt.Println(state.ActiveTheme.CursorTheme)
			case "properties":
				requireActiveTheme(state)
				for k, v := range state.ActiveTheme.Properties {
					fmt.Println(k + "\t" + v)
				}
			default:
				if state.ActiveTheme != nil {
					for k, v := range state.ActiveTheme.Properties {
						if strings.ToLower(strcase.ToCamel(k)) == key {
							fmt.Println(v)
							break
						}
					}
				}

				slog.Error("property not found", "key", key)
				os.Exit(1)
			}
		},
	}

	return cmd
}

func requireActiveTheme(state *config.DotfileState) {
	if state.ActiveTheme == nil {
		slog.Error("active theme not set")
		os.Exit(1)
	}
}
