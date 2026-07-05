package app

import (
	"errors"
	"fmt"
	"net"
	"os"
	"path/filepath"

	"github.com/bobllor/rcon/config"
)

type App struct {
	root *RootCommand
	add  *AddCommand
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

	paths := AppPath{
		Home:   home,
		Config: configPath,
	}

	mkErr := MkdirAll(configPath)
	// errors will not exit
	if mkErr != nil {
		fmt.Fprintf(os.Stderr, "encountered failures while making folders: %v\n", mkErr)
	}

	rootCmd := NewRootCommand(paths)
	addCmd := NewAddCommand(paths)
	listCmd := NewListCommand(paths)
	rmCmd := NewRemoveCommand(paths)

	rootCmd.Cmd.AddCommand(addCmd.Cmd)
	rootCmd.Cmd.AddCommand(listCmd.Cmd)
	rootCmd.Cmd.AddCommand(rmCmd.Cmd)

	err = rootCmd.Cmd.Execute()
	if err != nil {
		PrintFatal(err)
	}
}

// initEntry initializes the entries if the values are empty. If the values
// are not empty, then this will do nothing.
//
// This does not validate the input aside from empty string checks.
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
		addErr := validateAddress(address)
		if addErr != nil {
			return addErr
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
	err := validateAddress(entry.Address)
	if err != nil {
		return err
	}

	if entry.Password == "" {
		return errors.New("cannot have an empty RCON password")
	}

	return nil
}

// initRconName asks for the input of the name and returns the input.
func initRconName() (string, error) {
	fmt.Print(`Enter a unique RCON name identifier: `)
	name, err := ReadInput()
	if err != nil {
		return "", err
	}
	if name == "" {
		return "", errors.New("cannot have an empty RCON name identifier")
	}

	return name, nil
}

// validateAddress validates if the address string is a valid
// address string.
func validateAddress(address string) error {
	_, _, err := net.SplitHostPort(address)
	if err != nil {
		return err
	}

	return nil
}

// loadConfiguration loads the configuration and returns a new Configuration.
//
// If the file does not exist, it will return a zeroed Configuration.
func loadConfiguration(root string) (*config.Configuration, error) {
	cfg := config.NewConfiguration()

	loadedCfg, cfgErr := config.LoadConfiguration(root)
	// no errors, will fall back to terminal if an error occurs
	if errors.Is(cfgErr, os.ErrNotExist) {
		fmt.Println("mcrcon config not found")
	} else if cfgErr != nil {
		return nil, cfgErr
	} else {
		return loadedCfg, nil
	}

	return cfg, nil
}
