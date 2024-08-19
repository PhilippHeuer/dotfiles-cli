package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/tabwriter"

	"github.com/PhilippHeuer/dotfiles-cli/pkg/config"
	"github.com/PhilippHeuer/dotfiles-cli/pkg/util"
	"github.com/iancoleman/strcase"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

func queryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "query",
		Short: "query the application config or state",
		Run: func(cmd *cobra.Command, args []string) {
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

			// load config
			conf, err := config.Load(filepath.Join(state.Source, "dotfiles.yaml"))
			if err != nil {
				log.Fatal().Err(err).Str("file", filepath.Join(state.Source, "config.yaml")).Msg("failed to parse config file")
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
				_, _ = fmt.Fprintln(w, "CosmicTheme\t"+at.CosmicTheme)
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
			case "cosmictheme":
				requireActiveTheme(state)
				fmt.Println(state.ActiveTheme.CosmicTheme)
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

				log.Fatal().Str("key", key).Msg("property not found")
				os.Exit(1)
			}
		},
	}

	return cmd
}

func requireActiveTheme(state *config.DotfileState) {
	if state.ActiveTheme == nil {
		log.Fatal().Msg("active theme not set")
	}
}
