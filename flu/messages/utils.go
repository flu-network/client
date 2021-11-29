package messages

import (
	"encoding/binary"

	"github.com/flu-network/client/common"
)

// byteReader is a wrapper around a []byte that makes it easier to parse things if you know what
// data to expect. For example, if you know the message contains two uint8s and a sha1Hash you could
// just call br.readByte(); br.readByte(); br.readSha1Hash() in that order.
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

func (b *byteReader) readUint32() uint32 {
	result := binary.BigEndian.Uint32(b.Data[b.index : b.index+4])
	b.index += 4
	return result
}

func (b *byteReader) readUint64() uint64 {
	result := binary.BigEndian.Uint64(b.Data[b.index : b.index+8])
	b.index += 8
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

func (b *byteReader) readString256() string {
	length := int(b.readByte())
	result := string(b.Data[b.index : b.index+length])
	b.index += length
	return result
}

// Serialization utilities

// SerializeString255 serializaes the first 255 bytes of a string. Later bytes are ignored.
func SerializeString255(s string) []byte {
	strLen := len(s)
	if strLen > 255 {
		strLen = 255
	}
	result := make([]byte, 1+strLen)

	resultLen := copy(result[1:], s[:])
	result[0] = uint8(resultLen)
	return result
}
