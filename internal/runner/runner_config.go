package runner

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type RunnerConfig struct {
	IP   string `json:"IP,omitempty"`
	Port uint16 `json:"Port"`
}

func NewConfig() RunnerConfig {
	config, err := readConfigFromFile()
	if err != nil {
		fmt.Printf("Error occured while fetching config: %s. Creating a default config.\n", err.Error())
		return defaultConfig()
	}

	return config
}

func readConfigFromFile() (RunnerConfig, error) {
	path, err := os.Executable()
	if err != nil {
		fmt.Printf("[CONFIG] Error reading executable directory.\n")
		return RunnerConfig{}, err
	}

	data, err := os.ReadFile(fmt.Sprintf("%s/%s", filepath.Dir(path), "config.json"))
	if err != nil {
		fmt.Printf("[CONFIG] Error reading config.\n")
		return RunnerConfig{}, err
	}

	var rc RunnerConfig
	if err := json.Unmarshal(data, &rc); err != nil {
		fmt.Printf("[CONFIG] Error unmarshalling config.\n")
		return RunnerConfig{}, err
	}

	return rc, nil
}

func defaultConfig() RunnerConfig {
	return RunnerConfig{
		IP:   "192.168.0.169",
		Port: 42069,
	}
}
