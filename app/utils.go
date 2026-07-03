package app

import (
	"bufio"
	"fmt"
	"os"
	"strings"

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

// readInput reads the STDIN and returns the given input.
func ReadInput() (string, error) {
	reader := bufio.NewReader(os.Stdin)

	address, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}

	address = strings.TrimSpace(address)

	return address, nil
}

// readInputHidden reads the STDIN with a hidden input and
// returns the given input.
func ReadInputHidden() (string, error) {
	b, err := term.ReadPassword(0)
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
