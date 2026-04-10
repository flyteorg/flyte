package data

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
)

type FileIOConfig struct {
	Path         string `json:"path"`
	VariableName string `json:"variableName"`
}

func LoadFileIOConfigs(fileIOConfigFilePath string, fileIOConfigDir string) (map[string]FileIOConfig, error) {
	raw, err := os.ReadFile(fileIOConfigFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file config file %q: %w", fileIOConfigFilePath, err)
	}
	var configsList []FileIOConfig
	if err := json.Unmarshal(raw, &configsList); err != nil {
		return nil, fmt.Errorf("failed to unmarshal file config: %w", err)
	}
	configs := make(map[string]FileIOConfig)
	for _, config := range configsList {
		config.Path = path.Join(fileIOConfigDir, config.Path)
		configs[config.VariableName] = config
	}
	return configs, nil
}
