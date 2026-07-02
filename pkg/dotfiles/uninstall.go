package dotfiles

import (
	"log/slog"
	"os"
)

// DeleteManagedFiles deletes all files listed in managedFiles.
// If dryRun is true, no files are deleted but those that would be deleted are returned.
// It returns a slice of files that could not be deleted.
func DeleteManagedFiles(managedFiles []string, dryRun bool) []string {
	var failedToDelete []string

	for _, file := range managedFiles {
		slog.Debug("removing file", "file", file)

		if dryRun {
			failedToDelete = append(failedToDelete, file)
			continue
		}

		if _, err := os.Stat(file); os.IsNotExist(err) {
			slog.Debug("file does not exist, already deleted", "file", file)
			continue
		}

		if err := os.Remove(file); err != nil {
			failedToDelete = append(failedToDelete, file)
			slog.Debug("failed to remove file", "file", file, "err", err)
		}
	}

	return failedToDelete
}
