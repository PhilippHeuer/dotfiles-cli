package util

import (
	"os"
	"os/exec"
)

// RunCommand executes a given shell command and returns an error if the command fails.
func RunCommand(command string) error {
	cmd := exec.Command("sh", "-c", command)

	cmd.Stdout = os.Stderr
	cmd.Stderr = os.Stderr

	return cmd.Run()
}
