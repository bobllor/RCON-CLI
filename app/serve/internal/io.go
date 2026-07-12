package internal

import (
	"os"
	"runtime"
	"strconv"
	"syscall"
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
//
// A slice of errors will be returned or nil if no errors occurred.
func RemoveFiles(paths ...string) []error {
	errs := []error{}
	for _, path := range paths {
		err := os.Remove(path)
		if err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) == 0 {
		return nil
	}

	return errs
}

// CheckProcessRunning checks the process if it is running or not. It will
// return true if the process is running.
//
// This handles both Unix and Windows.
func CheckProcessRunning(pid int) bool {
	p, err := os.FindProcess(pid)
	if err != nil {
		// i think windows will error out. error could be returned if yes.
		// if that is the case then this should be false. for now test it out.
		return false
	}

	if runtime.GOOS == "linux" || runtime.GOOS == "darwin" {
		err = p.Signal(syscall.Signal(0))
		if err != nil {
			return false
		}
	}

	return true
}
