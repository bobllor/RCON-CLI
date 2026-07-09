package rcon

import (
	"bytes"
	"encoding/binary"
	"errors"
	"net"
	"time"

	"github.com/bobllor/rcon/packet"
)

type Rcon struct {
	Conn net.Conn
}

func NewRcon(address string) (*Rcon, error) {
	conn, err := net.DialTimeout("tcp", address, time.Second*5)
	if err != nil {
		return nil, err
	}

	rcon := &Rcon{
		Conn: conn,
	}

	return rcon, nil
}

func (r *Rcon) Authenticate(password string) error {
	loginPacket := packet.NewPacket([]byte(password), packet.PacketLogin)
	payload, err := loginPacket.ToBytes()
	if err != nil {
		return err
	}

	writeErr := r.write(payload)
	if writeErr != nil {
		return err
	}

	// only the request ID matters here
	res, err := r.read()
	if err != nil {
		return err
	}
	if res.RequestId == -1 {
		return errors.New("Failed to authenticate")
	}

	return nil
}

// Command sends a command to the server. It returns the
// output of the command.
//
// The output of the command can be empty, it is dependent on what
// command was sent over RCON.
func (r *Rcon) Command(command string) (string, error) {
	commandPacket := packet.NewPacket([]byte(command), packet.PacketCommand)
	payload, err := commandPacket.ToBytes()
	if err != nil {
		return "", err
	}

	writeErr := r.write(payload)
	if writeErr != nil {
		return "", err
	}

	resPacket, err := r.read()
	if err != nil {
		return "", err
	}

	return string(resPacket.Payload), nil
}

// Closes the connection.
func (r *Rcon) Close() error {
	return r.Conn.Close()
}

// write writes the payload to the server.
func (r *Rcon) write(payload []byte) error {
	_, err := r.Conn.Write(payload)
	if err != nil {
		return err
	}

	return nil
}

// read reads the connection into a new Packet response.
//
// This should only be called after the client -> server communication.
func (r *Rcon) read() (*packet.Packet, error) {
	MAX_PAYLOAD_LENGTH := 4096
	// make([]byte, 0, MAX_PAYLOAD_LENGTH) does not work, res ends up being empty
	res := make([]byte, MAX_PAYLOAD_LENGTH)

	// NOTE: the response can be fragmented, check https://minecraft.wiki/w/RCON#Fragmentation
	// TODO: deal with potential multi-packet response

	// NOTE: according to docs the response is not reliable
	// if bugs occur good luck. TODO: add logging buddy
	_, err := r.Conn.Read(res)
	if err != nil {
		return nil, err
	}

	// length (4) + request ID (4) + type (4) + payload (remainder)
	if len(res) < 12 {
		return nil, errors.New("ERROR: packet response size is invalid")
	}

	packet := packet.NewPacket(res, packet.PacketMulti)
	packet.Length = int32(binary.LittleEndian.Uint32(res[0:4]))
	packet.RequestId = int32(binary.LittleEndian.Uint32(res[4:8]))
	packet.Type = int32(binary.LittleEndian.Uint32(res[8:12]))
	packet.Payload = bytes.TrimRight(res[12:], "\x00")

	return packet, nil
}
