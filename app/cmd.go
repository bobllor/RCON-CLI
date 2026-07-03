package app

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/bobllor/rcon/config"
)

type App struct {
	root *RootCommand
}

type AppPath struct {
	// Home is the default home path of the logged in user on the device.
	Home string
	// Config is the path where the configuration file is stored in.
	Config string
	// Log is the path where logging is stored in.
	Log string
}

// Execute is the main entry point of the root and its children. It will create a new
// App struct and use it for the rest of the info.
func Execute() {
	home, err := os.UserHomeDir()
	if err != nil {
		home = "."
	}
	configPath := filepath.Join(home, ".config", "mcrcon")

	paths := &AppPath{
		Home:   home,
		Config: configPath,
	}

	mkErr := MkdirAll(configPath)
	// errors will not exit
	if mkErr != nil {
		fmt.Fprintf(os.Stderr, "encountered failures while making folders: %v\n", mkErr)
	}

	rootCmd := NewRootCommand(paths)
	err = rootCmd.Cmd.Execute()
	if err != nil {
		PrintFatal(err)
	}
}

// NewConfig creates a new Configuration for use by reading from the
// given root for the config file.
// If errors occur during the reading/parsing process, then it will return an
// error.
//
// The config file must be named config and end with a valid YAML extension.
// It is case insenstive.
func NewConfig(rootPath string) (*config.Configuration, error) {
	cfg, err := config.NewConfiguration(rootPath)
	if err != nil {
		return nil, err
	}

	return cfg, nil
}

// initEntry initializes the entries if the values are empty.
//
// This does not validate the input.
func initEntry(entry *config.RconEntry) error {
	if entry.Address == "" {
		fmt.Print("Enter the RCON address: ")
		address, err := ReadInput()
		if err != nil {
			return err
		}
		if address == "" {
			return errors.New("cannot have an empty RCON address")
		}
		entry.Address = address
	}

	if entry.Password == "" {
		fmt.Print("Enter the RCON password: ")
		pw, err := ReadInputHidden()
		if err != nil {
			return err
		}
		if pw == "" {
			return errors.New("cannot have an empty RCON password")
		}

		entry.Password = pw
	}

	return nil
}

// validateEntry validates the values of the entry. It will return
// an error if validation fails.
func validateEntry(entry config.RconEntry) error {
	if entry.Address == "" {
		return errors.New("cannot have an empty RCON address")
	}

	if entry.Password == "" {
		return errors.New("cannot have an empty RCON password")
	}

	return nil
}
