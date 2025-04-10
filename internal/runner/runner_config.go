package runner

import (
	"encoding/json"
	"fmt"
	"os"
)

type RunnerConfig struct {
	IP   string
	Port uint16
}

func NewConfig() RunnerConfig {
	return readConfigFromFile()
}

func readConfigFromFile() RunnerConfig {
	data, err := os.ReadFile("config.json")
	if err != nil {
		fmt.Printf("Error reading config: %s. Creating a default config.\n", err.Error())
		return defaultConfig()
	}

	var rc RunnerConfig
	if err := json.Unmarshal(data, &rc); err != nil {
		fmt.Printf("Error unmarshalling config: %s. Creating a default config.\n", err.Error())
		return defaultConfig()
	}
	return rc
}

func defaultConfig() RunnerConfig {
	return RunnerConfig{
		IP:   "192.168.0.169",
		Port: 42069,
	}
}
