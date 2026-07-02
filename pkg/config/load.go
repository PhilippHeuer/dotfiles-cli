package config

import (
	"log/slog"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

func Load(file string, require bool) (*DotfilesConfig, error) {
	cfg := DotfilesConfig{}

	absFile, err := filepath.Abs(file)
	if err != nil {
		return nil, err
	}
	baseDir := filepath.Dir(absFile)

	// if file does not exist, return empty state
	if _, err := os.Stat(file); os.IsNotExist(err) {
		if !require {
			return &cfg, nil
		}
		return nil, err
	}

	// read file
	fileContent, fileReadErr := os.ReadFile(file)
	if fileReadErr != nil {
		return nil, fileReadErr
	}

	// unmarshal
	yamlErr := yaml.Unmarshal(fileContent, &cfg)
	if yamlErr != nil {
		return nil, yamlErr
	}

	if len(cfg.Includes) > 0 {
		for _, include := range cfg.Includes {
			includePath := include
			if !filepath.IsAbs(include) {
				includePath = filepath.Join(baseDir, include)
			}
			slog.Debug("including config file", "file", includePath)

			includeCfg, err := Load(includePath, false)
			if err != nil {
				return nil, err
			}

			cfg = *mergeConfigs(cfg, *includeCfg)
		}
	}

	return &cfg, nil
}

func mergeConfigs(a DotfilesConfig, b DotfilesConfig) *DotfilesConfig {
	merged := a

	merged.Themes = append(merged.Themes, b.Themes...)
	merged.Commands = append(merged.Commands, b.Commands...)
	merged.Directories = append(merged.Directories, b.Directories...)

	return &merged
}
