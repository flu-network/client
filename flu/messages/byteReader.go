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

func (b *byteReader) readSliceUint16() []uint16 {
	length := int(b.readByte())
	result := make([]uint16, length)
	for i := 0; i < length; i++ {
		result[i] = b.readUint16()
	}
	return result
}
