package config

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/PhilippHeuer/dotfiles-cli/pkg/util"
	"github.com/adrg/xdg"
)

type DotfileState struct {
	Theme        string       `json:"theme"`
	ActiveTheme  *ThemeConfig `json:"active_theme"`
	Source       string       `json:"source"`
	ManagedFiles []string     `json:"managed_files"`
}

func StateFile() string {
	if os.Getenv("DOTFILE_STATE_FILE") != "" {
		return os.ExpandEnv("$DOTFILE_STATE_FILE")
	}
	return filepath.Join(xdg.StateHome, "dotfiles", "state.json")
}

func LoadState(file string) (*DotfileState, error) {
	s := &DotfileState{
		ManagedFiles: []string{},
	}

	// if file does not exist, return empty state
	if _, err := os.Stat(file); os.IsNotExist(err) {
		return s, nil
	}

	// read file
	data, err := os.ReadFile(file)
	if err != nil {
		return s, err
	}

	// unmarshal
	if err := json.Unmarshal(data, &s); err != nil {
		return s, err
	}

	return s, nil
}

func SaveState(file string, state *DotfileState) error {
	// create state directory
	err := util.CreateParentDirectory(file)
	if err != nil {
		return err
	}

	// write to file
	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(file, data, 0644); err != nil {
		return err
	}

	return nil
}
