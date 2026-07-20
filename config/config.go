package config

import (
	"errors"
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
	path := filepath.Join(root, DEFAULT_YAML_NAME)

	_, err := os.Stat(path)
	if errors.Is(err, os.ErrNotExist) {
		writeErr := config.WriteFile(root)
		if writeErr != nil {
			return nil, writeErr
		}
	} else if err != nil {
		return nil, err
	}

	return LoadConfiguration(root)
}

// NewConfiguration creates a new Configuraton struct with
// zeroed values.
func NewConfiguration() *Configuration {
	return &Configuration{
		RconEntries: make(map[string]RconEntry),
	}
}

// WriteFile writes the Configuration data into a given root path. It will
// create a swap file and rename to the value of DEFAULT_YAML_NAME.
//
// If the root path does not exist, then it will be created.
//
// All data in the original file will be overwritten if this is called.
func (c *Configuration) WriteFile(root string) error {
	b, err := yaml.Marshal(c)
	if err != nil {
		return err
	}

	err = os.MkdirAll(root, 0o700)
	if err != nil {
		return err
	}

	tempFi, err := os.CreateTemp(root, ".*config.swp")
	if err != nil {
		return err
	}

	_, err = tempFi.Write(b)
	if err != nil {
		return err
	}
	tempFi.Close()

	path := tempFi.Name()

	return os.Rename(path, filepath.Join(root, DEFAULT_YAML_NAME))
}

// HasEntry checks for the existence of the entry in the RCON entries map.
// It will return true if the entryName exists, otherwise it false.
func (c *Configuration) HasEntry(entryName string) bool {
	_, ok := c.RconEntries[entryName]

	return ok
}

// AddEntry adds a new entry to the RconEntries map. This will overwrite
// an existing entry.
func (c *Configuration) AddEntry(entryName string, entry RconEntry) {
	c.entryNilCheck()

	c.RconEntries[entryName] = entry
}

// DeleteEntry deletes an entry from the RconEntries map. If the entry
// does not exist, then it will do nothing.
//
// If the given entry is also the default entry, then the default entry is
// reset to an empty string.
//
// It will return true or false depending on if the key exists.
func (c *Configuration) DeleteEntry(entryName string) bool {
	c.entryNilCheck()

	_, ok := c.RconEntries[entryName]
	delete(c.RconEntries, entryName)

	if c.DefaultRcon == entryName {
		c.DefaultRcon = ""
	}

	return ok
}

// entryNilCheck checks if the entry map is nil. If it is, then it will initialize
// a new map. If it aready exists, then this does nothing.
func (c *Configuration) entryNilCheck() {
	// sanity check
	if c.RconEntries == nil {
		c.RconEntries = make(map[string]RconEntry)
	}
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
