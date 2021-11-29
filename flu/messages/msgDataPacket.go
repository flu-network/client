package messages

import (
	"encoding/binary"

	"github.com/flu-network/client/common"
)

// DataPacket is a slice of data of an arbitrary length that is generally part of a larger message.
// The Offset indicates the position in the larger message where this packet belongs. By convention,
// if the offset is zero, the first 20 bytes are the sha1hash of the data the connection is scoped
// to, and the next four bytes (indices 20, 21, 22, 23) are the number of bytes of salient data
// we hope to transmit.
//
// DataPackets are different from other flu messages in that they are parsed depending on their
// context. The first byte is used as a control structure that tells the receiver how to parse the
// rest of the DataPacket.
type DataPacket struct {
	Offset uint32
	Data   []byte
}

// Type returns a uint8 that identifies this message type
func (r *DataPacket) Type() byte {
	return dataPacket
}

// Serialize converts its subject into a []byte for transmission over the wire
func (r *DataPacket) Serialize() []byte {
	result := make([]byte, 5+len(r.Data))

	// message type
	result[0] = dataPacket

	// offset
	binary.BigEndian.PutUint32(result[1:5], r.Offset)

	// data
	copy(result[5:], r.Data)

	return result
}

// Split interprets the DataPacket as the first message with a zero offset and returns the
// the individual components of that message, namely: a hash of its contents, the size of the
// overall data being transmitted in this connection, and the raw data itself
func (r *DataPacket) Split() (*common.Sha1Hash, uint32, []byte) {
	hash := common.Sha1Hash{Data: [20]byte{}}
	copy(hash.Data[:], r.Data[0:20]) // copy hash
	chunkSize := binary.BigEndian.Uint32(r.Data[20:24])
	data := r.Data[24:]
	return &hash, chunkSize, data
}

// ParseAsDataPacket takes raw bytes from the wire and parses them as a DataPacket. It skips over
// the first byte since that's used to indicate what type the message is, and this function assumes
// the caller already has good reason to know this is a DataPacket and not something else.
func ParseAsDataPacket(data []byte) *DataPacket {
	return &DataPacket{
		Offset: binary.BigEndian.Uint32(data[1:5]),
		Data:   data[5:],
	}
}

// DataPacketAck is sent by the receiver to the sender with flow control and retransmission
// information
type DataPacketAck struct {
	Offset uint32
}

func (ack *DataPacketAck) Serialize() []byte {
	result := make([]byte, 5)

	// message type
	result[0] = dataPacketAck

	// offset
	binary.BigEndian.PutUint32(result[1:], ack.Offset)

	return result
}

func (ack *DataPacketAck) Type() byte {
	return dataPacketAck
}
