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
//
// Root is the root path of where the configuration file is being held.
//
// An error can occur while attempting to read the root or file and if
// the YAML parsing fails.
func NewConfiguration(root string) (*Configuration, error) {
	b, err := readYaml(root)
	if err != nil {
		return nil, err
	}

	var config Configuration
	unmarshalErr := yaml.Unmarshal(b, &config)
	if unmarshalErr != nil {
		return nil, err
	}

	return &config, nil
}

// WriteFile writes the current data structure into the YAML path.
//
// All data in the original file will be overwritten if this is called.
func (c *Configuration) WriteFile(filePath string) error {
	b, err := yaml.Marshal(c)
	if err != nil {
		return err
	}

	return os.WriteFile(filePath, b, 0o600)
}

// readYaml searches the root path for the configuration YAML file, reads the file,
// and return the bytes.
//
// This file must be named config followed by a YAML extension.
// The extension can YAML or YML, case insensitive.
func readYaml(root string) ([]byte, error) {
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
