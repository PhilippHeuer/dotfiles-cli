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
	return os.MkdirAll(filepath.Dir(path), 0755)
}

func LinkFile(source string, target string, dryRun bool, mode string, properties map[string]string) error {
	if dryRun {
		return nil
	}

	if err := CreateParentDirectory(target); err != nil {
		return err
	}

	if _, err := os.Stat(target); err == nil {
		return nil
	}

	switch mode {
	case "template":
		return copyFileWithTemplate(source, target, properties)
	case "copy":
		return copyFile(source, target)
	case "symlink":
		return createOrUpdateSymlink(source, target)
	default:
		return fmt.Errorf("invalid mode: %s (valid values: copy, symlink)", mode)
	}
}

func copyFile(source string, target string) error {
	src, err := os.Open(source)
	if err != nil {
		return err
	}
	defer src.Close()

	tgt, err := os.Create(target)
	if err != nil {
		return err
	}
	defer tgt.Close()

	if _, err := io.Copy(tgt, src); err != nil {
		return err
	}

	return ensureExecutable(source, target)
}

func copyFileWithTemplate(source string, target string, data map[string]string) error {
	src, err := os.Open(source)
	if err != nil {
		return err
	}
	defer src.Close()

	tgt, err := os.Create(target)
	if err != nil {
		return err
	}
	defer tgt.Close()

	content, err := io.ReadAll(src)
	if err != nil {
		return err
	}

	tmpl, err := template.New("template").Parse(string(content))
	if err != nil {
		return err
	}
	if err := tmpl.Execute(tgt, data); err != nil {
		return err
	}

	return ensureExecutable(source, target)
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

func ensureExecutable(source, target string) error {
	srcInfo, err := os.Stat(source)
	if err != nil || srcInfo.Mode()&0100 == 0 {
		return nil
	}
	tgtInfo, err := os.Stat(target)
	if err != nil {
		return err
	}
	return os.Chmod(target, tgtInfo.Mode()|0100)
}
