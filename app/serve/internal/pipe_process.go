package internal

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
)

// PipeProcessError is a struct used to handle IPC.
type PipeProcess struct {
	OK  bool   `json:"ok"`
	Msg string `json:"msg"`
}

// SetError sets an error message and sets OK to false.
func (p *PipeProcess) SetError(msg string) {
	p.OK = false
	p.Msg = msg
}

// SetErrorf sets an error message and sets OK to false using a
// formatted string.
func (p *PipeProcess) SetErrorf(format string, a ...any) {
	p.OK = false
	p.Msg = fmt.Sprintf(format, a...)
}

// Marshal marshals the struct into a JSON string.
func (p *PipeProcess) MarshalString() (string, error) {
	b, err := json.Marshal(p)
	if err != nil {
		return "", err
	}

	return string(b), nil
}

// Encode encodes the data to any io.Writer.
func (p *PipeProcess) Encode(w io.Writer) error {
	// fi.Write does not work due to it causing a hang in the parent
	// if io.ReadAll/io.ReadFull is used. unsure why json.NewEncoder/Decoder works.
	return json.NewEncoder(w).Encode(p)
}

// ToError creates a new error from the message.
func (p *PipeProcess) ToError() error {
	return errors.New(p.Msg)
}
