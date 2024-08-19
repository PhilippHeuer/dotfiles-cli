package cmd

import (
	"fmt"
	"os"

	"github.com/PhilippHeuer/dotfiles-cli/pkg/config"
	"github.com/PhilippHeuer/dotfiles-cli/pkg/util"
	"github.com/iancoleman/strcase"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

func queryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "query",
		Short: "query the application state",
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
			if state == nil {
				log.Fatal().Msg("failed to load state, state is nil")
			}

			// check if the key is a property
			key := strcase.ToCamel(args[0])
			switch key {
			case "Source":
				fmt.Println(state.Source)
			case "Theme":
				requireActiveTheme(state)
				fmt.Println(state.ActiveTheme.Name)
			case "ColorScheme":
				requireActiveTheme(state)
				fmt.Println(state.ActiveTheme.ColorScheme)
			case "WallpaperDir":
				requireActiveTheme(state)
				fmt.Println(util.ResolvePath(state.ActiveTheme.WallpaperDir))
			case "FontFamily":
				requireActiveTheme(state)
				fmt.Println(state.ActiveTheme.FontFamily)
			case "FontSize":
				requireActiveTheme(state)
				fmt.Println(state.ActiveTheme.FontSize)
			case "CosmicTheme":
				requireActiveTheme(state)
				fmt.Println(state.ActiveTheme.CosmicTheme)
			case "GtkTheme":
				requireActiveTheme(state)
				fmt.Println(state.ActiveTheme.GtkTheme)
			case "IconTheme":
				requireActiveTheme(state)
				fmt.Println(state.ActiveTheme.IconTheme)
			case "CursorTheme":
				requireActiveTheme(state)
				fmt.Println(state.ActiveTheme.CursorTheme)
			default:
				if state.ActiveTheme != nil {
					for k, v := range state.ActiveTheme.Properties {
						if k == key {
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
