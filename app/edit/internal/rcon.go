package internal

import (
	"errors"
	"fmt"
	"strings"

	"github.com/bobllor/rcon-cli/app/utils"
	"github.com/bobllor/rcon-cli/config"
)

// HandleEditRconName handles changing the RCON name of the entry. It will delete the
// base name entry if the new name is validated successfully.
//
// If the base name is also the default entry, then the default entry will be set to the new
// entry name.
func HandleEditCfgRconName(cfg *config.Configuration, baseName string, newName string) error {
	if cfg.HasEntry(newName) {
		return fmt.Errorf("%s already exists as an entry, no changes were made", newName)
	}
	if strings.TrimSpace(newName) == "" {
		return errors.New("cannot have an empty name")
	}

	baseDefault := cfg.DefaultRcon

	cfg.DeleteEntry(baseName)

	if baseName == baseDefault {
		cfg.DefaultRcon = newName
	}

	return nil
}

// HandleEditAddress handles validating and setting the RCON address.
//
// This will update the given entry if the address is valid.
func HandleEditAddress(entry *config.RconEntry, addr string) error {
	err := utils.ValidateAddress(addr)
	if err != nil {
		return err
	}

	entry.Address = addr

	return nil
}

// HandleEditPassword handles validating and setting the RCON password.
//
// This will update the given entry if the password is valid.
func HandleEditPassword(entry *config.RconEntry, pw string) error {
	err := utils.ValidatePassword(pw)
	if err != nil {
		return err
	}

	entry.Password = pw

	return nil
}

// HandleInteractiveMode returns a string read from Stdin. It prints out
// a non-newline message before reading from Stdin.
//
// It can handle both hidden and non-hidden reads.
func HandleInteractiveMode(message string, hidden bool) (string, error) {
	fmt.Print(message)
	if !hidden {
		return utils.ReadInput()
	} else {
		return utils.ReadInputHidden()
	}
}
