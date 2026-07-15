package listenertest

import (
	"bytes"
	"encoding/binary"
	"errors"
	"net"

	"github.com/bobllor/rcon/packet"
)

type TcpListener struct {
	net.Listener
	isAuthenicated bool
}

const AuthPassword = "testpassword123"

// NewTcpListener creates a new TCP test listener that listens on
// the loopback address with an ephemeral port.
//
// The caller should call Close when it is finished.
func NewTcpListener() (*TcpListener, error) {
	li, err := net.Listen("tcp", ":0")
	if err != nil {
		return nil, err
	}

	return &TcpListener{
		Listener: li,
	}, nil
}

func (t *TcpListener) HandleConnection() error {
	for {
		conn, err := t.Accept()
		if err != nil {
			return err
		}

		go t.handleConnection(conn)
	}
}

func (t *TcpListener) handleConnection(conn net.Conn) {
	defer conn.Close()

	b := make([]byte, 4096)
	_, err := conn.Read(b)
	if err != nil {
		return
	}

	bytes.TrimRight(b, "\x00")

	err = t.authenticate(b)
	if err != nil {
		packet := packet.NewPacket([]byte(err.Error()), packet.PacketLogin)

		packet.RequestId = -1

		packetB, err := packet.ToBytes()
		if err != nil {
			return
		}

		_, err = conn.Write(packetB)
		if err != nil {
			return
		}
	}

	packet := packet.NewPacket(b, packet.PacketCommand)
	packetB, err := packet.ToBytes()
	if err != nil {
		return
	}

	_, err = conn.Write(packetB)
	if err != nil {
		return
	}
}

// authenticate handles the authentication.
func (t *TcpListener) authenticate(payload []byte) error {
	if t.isAuthenicated {
		return nil
	}

	// minimum has to be 12 or more
	if len(payload) < 12 {
		return errors.New("invalid payload size")
	}
	id := binary.LittleEndian.Uint32(payload[8:12])

	// 3 is the login packet
	if id != uint32(3) && !t.isAuthenicated {
		return errors.New("tried to execute command while not authenticated")
	}

	data := bytes.TrimRight(payload[12:], "\x00")
	// just for testing, the valid password is going to be hard coded
	if string(data) != AuthPassword {
		return errors.New("invalid password")
	}

	t.isAuthenicated = true

	return nil
}
