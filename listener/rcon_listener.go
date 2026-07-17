package listener

import (
	"bytes"
	"net"
	"strings"
	"time"

	"github.com/bobllor/rcon-cli/rcon"
)

type RconListener struct {
	net.Listener
}

// NewRconListener creates a new listener for RCON commands. It uses
// sockets to listen on.
func NewRconListener(addr string) (*RconListener, error) {
	ln, err := net.Listen("unix", addr)
	if err != nil {
		return nil, err
	}

	return &RconListener{Listener: ln}, err
}

// HandleConnection handles the incoming connection to the listener. It
// will return an error if any errors occur during Accept().
//
// It is expected the payload of the connection is a command string. This will
// be used to execute the RCON command.
//
// This will be a blocking action and start the loop for the listener.
// The caller is responsible for closing the connection.
func (r *RconListener) HandleConnection(rconn *rcon.Rcon) error {
	for {
		con, err := r.Accept()
		if err != nil {
			return err
		}

		go r.handleConnection(con, rconn)
	}
}

// Stop stops the server after a duration of time passes.
func (r *RconListener) Stop(d time.Duration) error {
	time.Sleep(d)

	return r.Close()
}

// handleConnection handles writing and reading to the connection, as well
// as run the command over RCON.
//
// conn will be written with the output of the RCON command and is closed
// in the method.
func (r *RconListener) handleConnection(conn net.Conn, rconn *rcon.Rcon) {
	defer conn.Close()

	buf := make([]byte, 4096)
	_, err := conn.Read(buf)
	if err != nil {
		// TODO: do something about this?
		return
	}

	buf = bytes.TrimRight(buf, "\x00")

	res, err := rconn.Command(string(buf))
	if err != nil {
		return
	}

	errIndex := strings.Index(res, "error")
	if errIndex != -1 {
		errLength := errIndex + len("error")
		res = res[:errLength] + "\n" + res[errLength:]
	}

	_, err = conn.Write([]byte(res))
	if err != nil {
		return
	}
}
