package main

import (
	"log/slog"

	"github.com/PhilippHeuer/dotfiles-cli/pkg/cmd"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
	status  = "clean"
)

// Init Hook
func init() {
	// pass version info the version cmd
	cmd.Version = version
	cmd.CommitHash = commit
	cmd.BuildAt = date
	cmd.RepositoryStatus = status
}

// CLI Main Entrypoint
func main() {
	if err := cmd.Execute(); err != nil {
		slog.Error("cli error", "err", err)
	}
}
