package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"text/tabwriter"

	"github.com/PhilippHeuer/dotfiles-cli/pkg/config"
	"github.com/PhilippHeuer/dotfiles-cli/pkg/util"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

func listThemeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list-themes",
		Short: "list all available themes",
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

			// output
			w := tabwriter.NewWriter(os.Stdout, 1, 1, 1, ' ', 0)
			_, _ = fmt.Fprintln(w, "NAME")
			for _, theme := range conf.Themes {
				_, _ = fmt.Fprintln(w, theme.Name)
			}
			_ = w.Flush()
		},
	}

	return cmd
}
