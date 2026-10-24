// Package config reads and writes the settings file next to the executable.
package config

import (
	"encoding/json"
	"os"

	"github.com/Tariomka/desktop-led-controller/internal/common"
)

const (
	fileName   = "config.json"
	maxRecent  = 8
	MinWidth   = 1280
	MinHeight  = 720
	fileMode   = os.FileMode(0644)
	DefaultIP  = "192.168.0.169"
	DefaultTCP = 42069
)

type Settings struct {
	IP     string `json:"IP,omitempty"`
	Port   uint16 `json:"Port"`
	Window Window `json:"Window"`
	Panels Panels `json:"Panels"`
	// Recent is the list of recently imported file paths, newest first.
	Recent []string `json:"Recent,omitempty"`
}

type Window struct {
	Width, Height int
}

// Panels is the layout state of the side and bottom panels.
type Panels struct {
	LeftOpen, RightOpen, BottomOpen     bool
	LeftWidth, RightWidth, BottomHeight float32
	RightTab, BottomTab                 int
}

func Default() Settings {
	return Settings{
		IP:     DefaultIP,
		Port:   DefaultTCP,
		Window: Window{Width: 1600, Height: 900},
		Panels: Panels{LeftOpen: true, RightOpen: true, BottomOpen: true, LeftWidth: 260, RightWidth: 320, BottomHeight: 220},
	}
}

// Load reads the settings; missing or broken files give the defaults (and the error).
func Load() (Settings, error) {
	settings := Default()
	path, err := common.GetFullPathFromRelativePath(fileName)
	if err != nil {
		return settings, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return settings, err
	}
	loaded := Default()
	if err := json.Unmarshal(data, &loaded); err != nil {
		return settings, err
	}
	loaded.normalize()
	return loaded, nil
}

func (this *Settings) Save() error {
	this.normalize()
	path, err := common.GetFullPathFromRelativePath(fileName)
	if err != nil {
		return err
	}
	data, err := json.MarshalIndent(this, "", "    ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, fileMode)
}

// AddRecent puts path first in the recent list.
func (this *Settings) AddRecent(path string) {
	recent := []string{path}
	for _, item := range this.Recent {
		if item != path && len(recent) < maxRecent {
			recent = append(recent, item)
		}
	}
	this.Recent = recent
}

func (this *Settings) normalize() {
	defaults := Default()
	if this.Port == 0 {
		this.Port = defaults.Port
	}
	if this.Window.Width < MinWidth || this.Window.Height < MinHeight {
		this.Window = defaults.Window
	}
	if this.Panels.LeftWidth < 160 {
		this.Panels.LeftWidth = defaults.Panels.LeftWidth
	}
	if this.Panels.RightWidth < 240 {
		this.Panels.RightWidth = defaults.Panels.RightWidth
	}
	if this.Panels.BottomHeight < 120 {
		this.Panels.BottomHeight = defaults.Panels.BottomHeight
	}
}
