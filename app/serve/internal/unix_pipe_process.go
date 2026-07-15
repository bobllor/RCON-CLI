package internal

import "os"

// UnixPipeProcess is a struct used to handle IPC between
// processes on Linux and Darwin devices.
type UnixPipeProcess struct {
	PipeProcess
	// pipe is the pipe file for IPC.
	pipe *os.File
}

func NewUnixPipeProcess(proc PipeProcess, pipe *os.File) *UnixPipeProcess {
	return &UnixPipeProcess{
		PipeProcess: proc,
		pipe:        pipe,
	}
}

// Report encodes the PipeProcess into JSON format to
// the pipe.
func (u *UnixPipeProcess) Report() error {
	return u.Encode(u.pipe)
}

func (u *UnixPipeProcess) ValidateHandshake() error {
	_, err := u.pipe.Stat()
	if err != nil {
		return ErrStartNotAllowed
	}
	return nil
}
