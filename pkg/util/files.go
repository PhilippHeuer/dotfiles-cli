package util

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

func GetAllFiles(root string) ([]string, error) {
	var files []string

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() {
			absPath, absPathErr := filepath.Abs(path)
			if absPathErr != nil {
				return fmt.Errorf("failed to get absolute path for %s: %w", path, absPathErr)
			}
			files = append(files, absPath)
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

func LinkFile(source string, target string, dryRun bool, mode string, properties map[string]string) error {
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
	switch mode {
	case "template":
		if err := copyFileWithTemplate(source, target, properties); err != nil {
			return fmt.Errorf("failed to copy file with template: %w", err)
		}
	case "copy":
		if err := copyFile(source, target); err != nil {
			return fmt.Errorf("failed to copy file: %w", err)
		}
	case "symlink":
		if err := createOrUpdateSymlink(source, target); err != nil {
			return fmt.Errorf("failed to create symlink: %w", err)
		}
	default:
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

	if isExecutable(source) {
		err = makeExecutableByOwner(target)
		if err != nil {
			return err
		}
	}

	return nil
}

func copyFileWithTemplate(source string, target string, data map[string]string) error {
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

	sourceContent, err := io.ReadAll(sourceFile)
	if err != nil {
		return err
	}
	tmpl, err := template.New("template").Parse(string(sourceContent))
	if err != nil {
		return err
	}
	err = tmpl.Execute(targetFile, data)
	if err != nil {
		return err
	}

	if isExecutable(source) {
		err = makeExecutableByOwner(target)
		if err != nil {
			return err
		}
	}

	return nil
}

func createOrUpdateSymlink(source string, target string) error {
	// check if symlink exists
	linkInfo, err := os.Lstat(target)
	if err == nil {
		if linkInfo.Mode()&os.ModeSymlink != 0 { // is symlink
			currentTarget, err := os.Readlink(target)
			if err != nil {
				return fmt.Errorf("failed to read existing symlink target: %w", err)
			}

			// skip if symlink points to the same target
			if currentTarget == source {
				return nil
			}

			// remove old symlink
			if err := os.Remove(target); err != nil {
				return fmt.Errorf("failed to remove existing symlink: %w", err)
			}
		} else { // is file/directory
			if err := os.Remove(target); err != nil {
				return fmt.Errorf("failed to remove existing file/directory: %w", err)
			}
		}
	}

	// create symlink
	if err := os.Symlink(source, target); err != nil {
		return fmt.Errorf("failed to create symlink: %w", err)
	}

	return nil
}

func isExecutable(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}

	return info.Mode()&0100 != 0
}

func makeExecutableByOwner(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		return err
	}

	newMode := info.Mode() | 0100
	if err := os.Chmod(path, newMode); err != nil {
		return err
	}

	return nil
}
