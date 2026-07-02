package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/bobllor/assert"
	"github.com/goccy/go-yaml"
)

var testYamlFixture = map[string]any{
	"default_rcon": "",
	"rcons": map[string]map[string]any{
		"rcon 1": {
			"address":  ":25525",
			"password": "password1",
		},
		"rcon 2": {
			"address":  ":21123",
			"password": "password2",
		},
	},
}

func TestYamlRead(t *testing.T) {
	dir := t.TempDir()

	err := writeYaml(dir)
	assert.Nil(t, err)

	b, err := readYaml(dir)
	assert.Nil(t, err)
	assert.NotEqual(t, len(b), 0)
}

func TestNewConfiguration(t *testing.T) {
	dir := t.TempDir()

	err := writeYaml(dir)
	assert.Nil(t, err)

	config, err := NewConfiguration(dir)
	assert.Nil(t, err)

	rcon_one := testYamlFixture["rcons"].(map[string]map[string]any)["rcon 1"]
	rcon_two := testYamlFixture["rcons"].(map[string]map[string]any)["rcon 2"]

	assert.Equal(t, config.DefaultRcon, testYamlFixture["default_rcon"])
	assert.Equal(t, config.RconEntries["rcon 1"].Address, rcon_one["address"])
	assert.Equal(t, config.RconEntries["rcon 2"].Address, rcon_two["address"])

	assert.Equal(t, config.RconEntries["rcon 1"].Password, rcon_one["password"])
	assert.Equal(t, config.RconEntries["rcon 2"].Password, rcon_two["password"])
}

// writeYaml writes the test yaml fixture to the given directory.
// It will automatically write it as config.yml.
//
// The file path will be returned.
func writeYaml(dir string) error {
	b, err := yaml.Marshal(testYamlFixture)
	if err != nil {
		return err
	}

	writeErr := os.WriteFile(filepath.Join(dir, "config.yml"), b, 0o600)
	if writeErr != nil {
		return writeErr
	}

	return nil
}
