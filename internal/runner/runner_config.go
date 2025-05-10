package runner

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/Tariomka/desktop-led-controller/internal/common"
)

const configFilename = "config.json"

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

	return *config
}

func readConfigFromFile() (*RunnerConfig, error) {
	fullPath, err := common.GetFullPathFromRelativePath(configFilename)
	if err != nil {
		fmt.Printf("[CONFIG] Error reading full path.\n")
		return nil, err
	}

	data, err := os.ReadFile(fullPath)
	if err != nil {
		fmt.Printf("[CONFIG] Error reading config.\n")
		return nil, err
	}

	var rc RunnerConfig
	if err := json.Unmarshal(data, &rc); err != nil {
		fmt.Printf("[CONFIG] Error unmarshalling config.\n")
		return nil, err
	}

	return &rc, nil
}

func defaultConfig() RunnerConfig {
	return RunnerConfig{
		IP:   "192.168.0.169",
		Port: 42069,
	}
}
