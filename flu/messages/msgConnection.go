package messages

import (
	"encoding/binary"

	"github.com/flu-network/client/common"
)

type OpenConnectionRequest struct {
	Sha1Hash  *common.Sha1Hash // which file we want
	Chunk     uint16           // which chunk of that file we want
	WindowCap uint16           // how many unacked requests we'll allow
}

func (r *OpenConnectionRequest) Serialize() []byte {
	result := make([]byte, 32)

	// message type
	result[0] = openLineRequest

	// sha1hash
	copy(result[1:21], r.Sha1Hash.Slice())

	// Chunk
	binary.BigEndian.PutUint16(result[21:23], r.Chunk)

	// window cap
	binary.BigEndian.PutUint16(result[23:32], r.WindowCap)

	return result
}

// Type returns a uint8 that identifies this message type
func (r *OpenConnectionRequest) Type() byte {
	return openLineRequest
}

// ResponseType returns a uint8 that identidies the type of response expected for this message
func (r *OpenConnectionRequest) ResponseType() byte {
	return dataPacket
}
