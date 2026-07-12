package internal

import (
	"os"
	"strconv"
)

// ReadPID reads the given path for the PID file.
// The file must exist and the contents must be a number.
func ReadPID(path string) (int, error) {
	buf, err := os.ReadFile(path)
	if err != nil {
		return 0, err
	}

	pid, err := strconv.Atoi(string(buf))
	if err != nil {
		return 0, err
	}

	return pid, nil
}

// RemoveFiles removes the given file paths. Directories are not supported.
func RemoveFiles(paths ...string) error {
	for _, path := range paths {
		err := os.Remove(path)
		if err != nil {
			return err
		}
	}

	return nil
}

type ConnReader interface {
	Read(p []byte) (int, error)
	Close() error
}
