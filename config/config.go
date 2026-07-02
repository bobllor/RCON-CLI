package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/goccy/go-yaml"
)

type Configuration struct {
	DefaultRcon string               `yaml:"default_rcon"`
	RconEntries map[string]RconEntry `yaml:"rcons"`
}

type RconEntry struct {
	Address  string `yaml:"address"`
	Password string `yaml:"password"`
}

// NewConfiguration unmarshals the bytes into a Configuration struct.
func NewConfiguration(b []byte) (*Configuration, error) {
	var config Configuration
	err := yaml.Unmarshal(b, &config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}

// ReadYaml searches the root path for the configuration yaml file and reads
// file and return the bytes.
//
// This file must be named config followed by a YAML extension.
// The extension can YAML or YML, case insensitive.
func ReadYaml(root string) ([]byte, error) {
	entries, err := os.ReadDir(root)
	if err != nil {
		return nil, err
	}

	var filePath string
	for _, entry := range entries {
		file := strings.ToLower(entry.Name())

		if file == "config.yml" || file == "config.yaml" {
			filePath = filepath.Join(root, entry.Name())
			break
		}
	}

	if filePath == "" {
		return nil, fmt.Errorf("config.yml does not exist in %s", root)
	}

	b, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	return b, nil
}
