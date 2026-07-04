package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/goccy/go-yaml"
)

const DEFAULT_YAML_NAME = "config.yml"

type Configuration struct {
	DefaultRcon string               `yaml:"default_rcon"`
	RconEntries map[string]RconEntry `yaml:"rcons"`
}

type RconEntry struct {
	Address  string `yaml:"address"`
	Password string `yaml:"password"`
}

// LoadConfiguration loads the config file from the given root string.
//
// The file must exist otherwise an os.ErrNotExist error will be returned.
// Other errors can occur while attempting to read the root or file and if
// the YAML parsing fails.
func LoadConfiguration(root string) (*Configuration, error) {
	var config Configuration

	b, err := readYaml(root)
	if err != nil {
		return nil, err
	}
	unmarshalErr := yaml.Unmarshal(b, &config)
	if unmarshalErr != nil {
		return nil, err
	}

	// unmarshal overwrites it to nil if the map does not exist in the file
	if config.RconEntries == nil {
		config.RconEntries = make(map[string]RconEntry)
	}

	return &config, nil
}

// LoadConfigurationIfMissing loads the config file from the given root string.
//
// This is a wrapper around LoadConfiguration, but instead the file is created if
// it does not exist.
func LoadConfigurationIfMissing(root string) (*Configuration, error) {
	var config Configuration

	_, err := os.Stat(filepath.Join(root, DEFAULT_YAML_NAME))
	if errors.Is(err, os.ErrNotExist) {
		writeErr := config.WriteFile(root)
		if writeErr != nil {
			return nil, writeErr
		}
		fmt.Println("created config file")
	} else if err != nil {
		return nil, err
	}

	return LoadConfiguration(root)
}

// WriteFile writes the current data structure into the YAML path.
//
// All data in the original file will be overwritten if this is called.
func (c *Configuration) WriteFile(root string) error {
	b, err := yaml.Marshal(c)
	if err != nil {
		return err
	}

	return os.WriteFile(filepath.Join(root, DEFAULT_YAML_NAME), b, 0o600)
}

// EntryExist checks for the existence of the entry in the RCON entries map.
// It will return true if the entryName exists, otherwise it false.
func (c *Configuration) EntryExist(entryName string) bool {
	_, ok := c.RconEntries[entryName]

	return ok
}

// AddEntry adds a new entry to the RconEntries map. This will overwrite
// an existing entry.
func (c *Configuration) AddEntry(entryName string, entry RconEntry) {
	// sanity check
	if c.RconEntries == nil {
		c.RconEntries = make(map[string]RconEntry)
	}

	c.RconEntries[entryName] = entry
}

// readYaml searches the root path for the configuration YAML file, reads the file,
// and return the bytes.
//
// If the file cannot be found, a os.ErrNotExist error will be returned.
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
		return nil, os.ErrNotExist
	}

	b, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	return b, nil
}
