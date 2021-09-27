package flu

import (
	"encoding/binary"

	"github.com/flu-network/client/common"
)

// import "fmt"

// // Message is a clever wrapper around a []byte.
// type Message struct {
// 	data []byte
// }

// func (m *Message) decode() (interface{}, error) {
// 	messageType, data := m.data[0], m.data[1:]

// 	switch messageType {
// 	case discoverHostRequest:
// 		// todo
// 	case discoverHostResponse:
// 		// todo
// 	}
// 	return nil, fmt.Errorf("Unknown message type %d", m.messageType)
// }

// type FluMessage interface {
// 	schama() []uint8
// }

// const (
// 	dataTypeSha1Hash    = uint8(iota)
// 	dataTypeUint16      = uint8(iota)
// 	dataTypeSliceUint16 = uint8(iota)
// )

// byteReader is a wrapper around a []byte that makes it easier to parse things if you know what
// data to expect
type byteReader struct {
	Data  []byte
	index int
}

func (b *byteReader) readByte() uint8 {
	result := b.Data[b.index]
	b.index++
	return result
}

func (b *byteReader) readBytes(count int) []byte {
	result := b.Data[b.index : b.index+count]
	b.index += count
	return result
}

func (b *byteReader) readSha1Hash() *common.Sha1Hash {
	hashData := b.Data[b.index : b.index+20]
	b.index += 20
	return (&common.Sha1Hash{}).FromSlice(hashData)
}

func (b *byteReader) readUint16() uint16 {
	result := binary.BigEndian.Uint16(b.Data[b.index : b.index+2])
	b.index += 2
	return result
}

func (b *byteReader) readSliceUint16() []uint16 {
	length := int(b.readByte())
	result := make([]uint16, length)
	for i := 0; i < length; i++ {
		result[i] = b.readUint16()
	}
	return result
}
