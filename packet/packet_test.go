package packet

import (
	"testing"

	"github.com/bobllor/assert"
)

func TestToBytes(t *testing.T) {
	packet := NewPacket([]byte("payload"), PacketLogin)

	b, err := packet.ToBytes()
	assert.Nil(t, err)
	assert.NotEqual(t, len(b), 0)
}
