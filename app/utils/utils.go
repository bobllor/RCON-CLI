package utils

import (
	"bufio"
	"errors"
	"fmt"
	"net"
	"os"
	"strings"
	"syscall"

	"github.com/bobllor/rcon-cli/config"
	"golang.org/x/term"
)

// PrintFatal prints the error and calls os.Exit(1).
func PrintFatal(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}

// PrintFatalString prints the string and calls os.Exit(1).
func PrintFatalString(msg string) {
	fmt.Fprintln(os.Stderr, msg)
	os.Exit(1)
}

// PrintFatalf prints the format string and its args
// and calls os.Exit(1).
func PrintFatalf(format string, a ...any) {
	fmt.Fprintf(os.Stderr, format+"\n", a...)
	os.Exit(1)
}

// readInput reads the STDIN and returns the given input.
//
// Spaces are automatically trimmed.
func ReadInput() (string, error) {
	reader := bufio.NewReader(os.Stdin)

	inp, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}

	inp = strings.TrimSpace(inp)

	return inp, nil
}

// readInputHidden reads the STDIN with a hidden input and
// returns the given input.
func ReadInputHidden() (string, error) {
	b, err := term.ReadPassword(int(syscall.Stdin))
	if err != nil {
		return "", err
	}

	// new line to break the print above
	fmt.Println()

	return string(b), nil
}

// MkdirAll creates a slice of folder paths and its children.
//
// By default it will create them with 0o600 permission.
func MkdirAll(paths ...string) error {
	errs := []string{}
	for _, p := range paths {
		err := os.MkdirAll(p, 0o700)
		if err != nil {
			errs = append(errs, err.Error())
		}
	}

	if len(errs) != 0 {
		return fmt.Errorf("had errors while creating folder: %s", strings.Join(errs, ";"))
	}

	return nil
}

// InitRconName asks for the input of the name and returns the input.
func InitRconName() (string, error) {
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

// InitEntry initializes the entries if the values are empty. If the values
// are not empty, then this will do nothing.
//
// This does not validate the input aside from empty string checks.
//
// If the password entry contains a "-", then it will prompt for an input.
func InitEntry(entry *config.RconEntry) error {
	if entry.Address == "" {
		fmt.Print("Enter the RCON address: ")
		address, err := ReadInput()
		if err != nil {
			return err
		}
		if address == "" {
			return errors.New("cannot have an empty RCON address")
		}
		addErr := ValidateAddress(address)
		if addErr != nil {
			return addErr
		}

		entry.Address = address
	}

	if entry.Password == "" || entry.Password == "-" {
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

// ValidateEntry validates the values of the entry. It will return
// an error if validation fails.
func ValidateEntry(entry config.RconEntry) error {
	if entry.Address == "" {
		return errors.New("cannot have an empty RCON address")
	}
	err := ValidateAddress(entry.Address)
	if err != nil {
		return err
	}

	if entry.Password == "" {
		return errors.New("cannot have an empty RCON password")
	}

	return nil
}

// ValidateAddress validates if the address string is a valid
// address string.
func ValidateAddress(address string) error {
	_, _, err := net.SplitHostPort(address)
	if err != nil {
		return err
	}

	return nil
}

// LoadConfiguration loads the configuration and returns a new Configuration.
//
// If the file does not exist, it will return a zeroed Configuration.
func LoadConfiguration(root string) (*config.Configuration, error) {
	cfg := config.NewConfiguration()

	loadedCfg, cfgErr := config.LoadConfiguration(root)
	// no errors, will fall back to terminal if an error occurs
	if errors.Is(cfgErr, os.ErrNotExist) {
		fmt.Println("rcon config not found")
	} else if cfgErr != nil {
		return nil, cfgErr
	} else {
		return loadedCfg, nil
	}

	return cfg, nil
}

// ValidatePassword validates the password string.
func ValidatePassword(pw string) error {
	if pw == "" {
		return errors.New("cannot have an empty password")
	}

	return nil
}
