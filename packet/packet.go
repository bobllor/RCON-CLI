package packet

import (
	"bytes"
	"encoding/binary"
	"math/rand"
	"time"
)

type Packet struct {
	// Length is the size of the payload. It is the size of all values in the packet
	// except Length itself.
	Length int32
	// Request ID is a randomly generated client ID associated with the
	// packet.
	RequestId int32
	// Type is the type of packet being used for the RCON communication.
	// It can be only three values: 3 for login, 2 for a command,
	// and 0 for a multi-packet response.
	Type int32
	// Payload is the payload of the bytes used. This represents the
	// command being executed via RCON.
	Payload []byte
	// null is the termination end of the payload. This is only a single byte, and
	// by default will be false.
	null bool
	// pad is the padding of the end of a packet. This is only a single byte, and
	// by default will be false.
	pad bool
}

type PacketType string

const (
	PacketLogin   PacketType = "LOGIN"
	PacketCommand PacketType = "COMMAND"
	PacketMulti   PacketType = "MULTI"
)

// NewRequestId generates a new int32 request ID.
func NewRequestId() int32 {
	rand := rand.New(rand.NewSource(time.Now().Unix()))

	return rand.Int31()
}

func NewPacket(payload []byte, packetType PacketType) *Packet {
	var packetT int32
	switch packetType {
	case PacketCommand:
		packetT = 2
	case PacketLogin:
		packetT = 3
	default:
		packetT = 0
	}

	return &Packet{
		Length:    packetLength(len(payload)),
		RequestId: NewRequestId(),
		Type:      packetT,
		Payload:   payload,
		null:      false,
		pad:       false,
	}
}

// ToBytes converts the packet into a byte format. The bytes
// will be in a little-endian format.
func (p *Packet) ToBytes() ([]byte, error) {
	buf := bytes.Buffer{}
	data := []any{p.Length, p.RequestId, p.Type, p.Payload, p.null, p.pad}

	for _, d := range data {
		err := binary.Write(&buf, binary.LittleEndian, d)
		if err != nil {
			return nil, err
		}
	}

	return buf.Bytes(), nil
}

// packetLength calculates the length of the packet to send over
// the connection.
func packetLength(payloadSize int) int32 {
	// int32 + int32 + payload + null-terminated + null
	// the first length int32 is not included (you forgot this again)
	return int32(4 + 4 + payloadSize + 2)
}
