package paths

import (
	"fmt"
	"os"
)

type AppPath struct {
	// Home is the default home path of the logged in user on the device.
	Home string
	// Config is the path where the configuration file is stored in.
	Config string
	// Log is the path where logging is stored in.
	Log string
	// Runtime is the path for files created during the program's runtime.
	Runtime string
}

// MkdirAll creates all the paths in AppPath. If the path already
// exists, then it will do nothing. By default all paths are created
// with permissions 700.
//
// If an error occurs, a slice of errors will be returned. If no errors occur,
// the return will be nil.
func (a *AppPath) MkdirAll() []error {
	paths := map[string]string{
		"Home":    a.Home,
		"Config":  a.Config,
		"Logs":    a.Log,
		"Runtime": a.Runtime,
	}

	errors := []error{}

	for key, path := range paths {
		if path == "" {
			errors = append(errors, fmt.Errorf("error: %s is an empty path", key))
		}

		err := os.MkdirAll(path, 0o700)
		if err != nil {
			errors = append(errors, err)
		}
	}

	if len(errors) == 0 {
		return nil
	}

	return errors
}
