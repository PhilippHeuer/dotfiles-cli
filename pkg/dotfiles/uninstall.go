package dotfiles

import (
	"os"

	"github.com/rs/zerolog/log"
)

// DeleteManagedFiles deletes all files listed in managedFiles.
// If dryRun is true, no files are deleted but those that would be deleted are returned.
// It returns a slice of files that could not be deleted.
func DeleteManagedFiles(managedFiles []string, dryRun bool) []string {
	var failedToDelete []string

	for _, file := range managedFiles {
		log.Debug().Str("file", file).Msg("removing file")

		if dryRun {
			failedToDelete = append(failedToDelete, file)
			continue
		}

		if _, err := os.Stat(file); os.IsNotExist(err) {
			log.Trace().Str("file", file).Msg("file does not exist, already deleted")
			continue
		}

		if err := os.Remove(file); err != nil {
			failedToDelete = append(failedToDelete, file)
			log.Debug().Err(err).Str("file", file).Msg("failed to remove file")
		}
	}

	return failedToDelete
}
