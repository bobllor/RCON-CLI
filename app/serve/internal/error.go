package internal

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
)

// PipeProcessError is a struct used to handle errors between
// IPC processes.
type PipeProcessError struct {
	OK  bool
	Msg string
}

// SetError sets an error message and sets OK to false.
func (p *PipeProcessError) SetError(msg string) {
	p.OK = false
	p.Msg = msg
}

// SetErrorf sets an error message and sets OK to false using a
// formatted string.
func (p *PipeProcessError) SetErrorf(format string, a ...any) {
	p.OK = false
	p.Msg = fmt.Sprintf(format, a...)
}

// Encode encodes the data to any io.Writer.
func (p *PipeProcessError) Encode(w io.Writer) error {
	// fi.Write does not work due to it causing a hang in the parent
	// if io.ReadAll/io.ReadFull is used. unsure why json.NewEncoder/Decoder works.
	return json.NewEncoder(w).Encode(p)
}

// ToError creates a new error from the message.
func (p *PipeProcessError) ToError() error {
	return errors.New(p.Msg)
}
