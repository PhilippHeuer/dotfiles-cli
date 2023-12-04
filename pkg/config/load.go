package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

func Load(file string) (*DotfilesConfig, error) {
	cfg := DotfilesConfig{}

	fileContent, fileReadErr := os.ReadFile(file)
	if fileReadErr != nil {
		return nil, fileReadErr
	}

	yamlErr := yaml.Unmarshal(fileContent, &cfg)
	if yamlErr != nil {
		return nil, yamlErr
	}

	return &cfg, nil
}
