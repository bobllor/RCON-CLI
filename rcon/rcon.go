package rcon

import (
	"encoding/binary"
	"errors"
	"fmt"
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

	res := make([]byte, 4096)
	resN, err := r.Conn.Read(res)
	if err != nil {
		return err
	}
	id := int32(binary.LittleEndian.Uint32(res[4:8]))
	if id == -1 {
		return errors.New("Failed to authenticate")
	}
	fmt.Printf("Read %d bytes from server\n", resN)

	return nil
}

// Command sends a command to the server.
func (r *Rcon) Command(command string) error {
	commandPacket := packet.NewPacket([]byte(command), packet.PacketCommand)
	payload, err := commandPacket.ToBytes()
	if err != nil {
		return err
	}

	writeErr := r.write(payload)
	if writeErr != nil {
		return err
	}

	return nil
}

// Closes the connection.
func (r *Rcon) Close() error {
	return r.Conn.Close()
}

// write writes the payload to the server.
func (r *Rcon) write(payload []byte) error {
	n, err := r.Conn.Write(payload)
	if err != nil {
		return err
	}
	fmt.Printf("Wrote %d bytes to payload\n", n)

	return nil
}
