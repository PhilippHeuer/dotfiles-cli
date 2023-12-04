package util

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func GetAllFiles(root string) ([]string, error) {
	var files []string

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			files = append(files, path)
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return files, nil
}

func ResolvePath(path string) string {
	// replace ~ with $HOME
	path = strings.Replace(path, "~", "$HOME", 1)

	// expand environment variables
	path = os.ExpandEnv(path)

	return path
}

func CreateParentDirectory(path string) error {
	// get parent directory
	parentDir := filepath.Dir(path)

	// create parent directory
	if err := os.MkdirAll(parentDir, 0755); err != nil {
		return err
	}

	return nil
}

func LinkFile(source string, target string, dryRun bool, mode string) error {
	if dryRun {
		return nil
	}

	// create parent directory
	if err := CreateParentDirectory(target); err != nil {
		return err
	}

	// check if file exists
	if _, err := os.Stat(target); err == nil {
		return nil
	}

	// create symlink
	if mode == "copy" {
		if err := copyFile(source, target); err != nil {
			return fmt.Errorf("failed to copy file: %w", err)
		}
	} else if mode == "symlink" {
		if err := os.Symlink(source, target); err != nil {
			return fmt.Errorf("failed to create symlink: %w", err)
		}
	} else {
		return fmt.Errorf("invalid mode: %s (valid values: copy, symlink)", mode)
	}

	return nil
}

func copyFile(source string, target string) error {
	sourceFile, err := os.Open(source)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	targetFile, err := os.Create(target)
	if err != nil {
		return err
	}
	defer targetFile.Close()

	_, err = io.Copy(targetFile, sourceFile)
	if err != nil {
		return err
	}

	return nil
}
