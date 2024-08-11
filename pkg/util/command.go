package util

import (
	"os"
	"os/exec"
)

// RunCommand executes a given shell command and returns an error if the command fails.
func RunCommand(command string) error {
	command = ResolvePath(command) // support placeholders such as ~ and $HOME

	cmd := exec.Command("sh", "-c", command)

	cmd.Stdout = os.Stderr
	cmd.Stderr = os.Stderr

	return cmd.Run()
}
