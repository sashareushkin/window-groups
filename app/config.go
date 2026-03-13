package app

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

const appName = "Window Groups"
const version = "0.1.0"

// Config holds application configuration
type Config struct {
	GroupsPath string `json:"groups_path"`
	Hotkeys    map[string]string `json:"hotkeys"`
}

// Storage handles data persistence
type Storage struct {
	configPath string
	groupsPath string
}

// NewStorage creates a new storage handler
func NewStorage() (*Storage, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	
	configDir := filepath.Join(homeDir, ".config", "window-groups")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return nil, err
	}

	return &Storage{
		configPath: filepath.Join(configDir, "config.json"),
		groupsPath: filepath.Join(configDir, "groups.json"),
	}, nil
}

// LoadConfig loads application configuration
func (s *Storage) LoadConfig() (*Config, error) {
	data, err := os.ReadFile(s.configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return &Config{
				GroupsPath: s.groupsPath,
				Hotkeys:    make(map[string]string),
			}, nil
		}
		return nil, err
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, err
	}
	return &config, nil
}

// SaveConfig saves application configuration
func (s *Storage) SaveConfig(config *Config) error {
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.configPath, data, 0644)
}

// GetGroupsPath returns the path to groups storage
func (s *Storage) GetGroupsPath() string {
	return s.groupsPath
}

// Init initializes the application
func Init() error {
	fmt.Printf("%s v%s\n", appName, version)
	storage, err := NewStorage()
	if err != nil {
		return err
	}
	_, err = storage.LoadConfig()
	return err
}
