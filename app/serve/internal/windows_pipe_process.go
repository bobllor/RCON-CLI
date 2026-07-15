package internal

import (
	"fmt"
	"os"
)

// WindowsPipeProcess is a struct used to handle IPC between
// processes on Windows.
type WindowsPipeProcess struct {
	PipeProcess
}

func NewWindowsPipeProcess(proc PipeProcess) *WindowsPipeProcess {
	return &WindowsPipeProcess{
		PipeProcess: proc,
	}
}

// Report reports the PipeInfo information back to the parent.
//
// It will print to stdout due to Windows pipes using exec.Cmd.StdoutPipe.
func (w *WindowsPipeProcess) Report() error {
	json, err := w.MarshalString()
	if err != nil {
		return err
	}
	fmt.Println(json)

	return nil
}

func (w *WindowsPipeProcess) ValidateHandshake() error {
	// is set if the internal flag is used
	envValue := os.Getenv(WindowsEnvKey)
	if envValue == "" {
		return ErrStartNotAllowed
	}

	return nil
}
