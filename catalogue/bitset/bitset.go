package bitset

import (
	"encoding/binary"
	"fmt"
)

const wordSize = 64
const allOn = ^uint64(0)

// Bitset is simple bitset backed by a []uint64. The size of the bitset is unbounded and it will
// grow as required to support any Set() operations invoked. Bitset provides no compression.
type Bitset struct {
	data []uint64
	size int
}

// NewBitset returns a new bitset of the specified size. Panics if size < 0. Providing a size is
// convenient for serialization, but not necessary. Get() and Set() operations work as described.
func NewBitset(size int) *Bitset {
	if size < 0 {
		panic("Bitmaps cannot have a negative size")
	}
	result := Bitset{size: size}

	if size > 0 {
		maxIndex := uint64(size - 1)
		result.Set(maxIndex)
		result.Unset(maxIndex)
	}

	return &result
}

// Get returns true if the specified index has been previously set and never subsequently unset.
// Safely false if the specified index is out of bounds.
func (b *Bitset) Get(x uint64) bool {
	if int(x)/wordSize >= len(b.data) {
		return false
	}
	return b.data[x/wordSize]&(1<<(x%wordSize)) != 0
}

// Set sets the specified index to true. If the specified index is out of bounds, the bitset
// expands to incorporate the given index. The bitset will never shrink once expanded.
func (b *Bitset) Set(x uint64) {
	offset := int(x) / wordSize
	if offset > len(b.data)-1 {
		diff := offset - (len(b.data) - 1)
		b.data = append(b.data, make([]uint64, diff)...)
	}
	b.data[x/wordSize] |= (1 << (x % wordSize))
	b.size = max(b.size, int(x)+1)
}

// Fill sets all items between 0:size to true.
func (b *Bitset) Fill() {
	if b.size == 0 {
		return
	}

	maxOffset := (b.size - 1) / wordSize
	for i := 0; i < maxOffset; i++ {
		b.data[i] = allOn
	}

	lastIndex := (b.size - 1) % wordSize
	b.data[maxOffset] = (((1 << lastIndex) - 1) << 1) | 1
}

// Full returns true if all items between 0:size are set to true, and false if not.
func (b *Bitset) Full() bool {
	if b.size == 0 {
		return true
	}

	maxOffset := (b.size - 1) / wordSize
	for i := 0; i < maxOffset; i++ {
		if b.data[i] != allOn {
			return false
		}
	}

	lastIndex := (b.size - 1) % wordSize
	return b.data[maxOffset] == (((1<<lastIndex)-1)<<1)|1
}

// Unset sets the specified index to False. If the specified index is out of bounds, nothing
// happens. Does not cause size to shrink.
func (b *Bitset) Unset(x uint64) {
	offset := int(x) / wordSize
	if offset >= len(b.data) {
		return
	}
	b.data[offset] = b.data[offset] - (1 << (x % wordSize))
}

// Union returns the union between any two bitsets
func (b *Bitset) Union(a *Bitset) *Bitset {
	shorter, longer := shorterFirst(a, b)
	result := make([]uint64, len(shorter.data))

	for i := 0; i < len(shorter.data); i++ {
		result[i] = shorter.data[i] | longer.data[i]
	}

	remainder := make([]uint64, len(longer.data)-len(shorter.data))
	copy(remainder, longer.data[len(shorter.data):])

	result = append(result, remainder...)
	return &Bitset{data: result}
}

// Intersect returns the intersect between any two bitsets
func (b *Bitset) Intersect(a *Bitset) *Bitset {
	shorter, longer := shorterFirst(a, b)
	result := make([]uint64, len(shorter.data))

	for i, x := range shorter.data {
		result[i] = x & longer.data[i]
	}

	return &Bitset{data: result}
}

// Serialize converts the bitset into a []byte so it can be transmitted somewhere. It makes a copy
// of its underlying data, but is not threadsafe. If race conditions are possible it is the
// caller's responsibility to maintain exclusivity.
func (b *Bitset) Serialize() []byte {
	sizeBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(sizeBytes, uint64(b.size))

	rest := make([]byte, len(b.data)*8)
	for i := 0; i < len(b.data); i++ {
		byteIndex := i * 8
		dest := rest[byteIndex : byteIndex+8]
		binary.BigEndian.PutUint64(dest, b.data[i])
	}

	return append(sizeBytes, rest...)
}

// Deserialize converts a previously-serialized bitset into a realized bitset.
func Deserialize(data []byte) (*Bitset, error) {
	size := int(binary.BigEndian.Uint64(data[:8]))

	if l := len(data[8:]); l%8 != 0 {
		return nil, fmt.Errorf("Data segment of bitset data must be a multiple of 8. Got %d", l)
	}

	setData := make([]uint64, len(data[8:])/8)
	for i := 0; i < len(setData); i++ {
		byteIndex := 8 + (i * 8)
		setData[i] = binary.BigEndian.Uint64(data[byteIndex : byteIndex+8])
	}

	return &Bitset{
		data: setData,
		size: size,
	}, nil
}

/*
Private utilities
*/

func shorterFirst(a, b *Bitset) (*Bitset, *Bitset) {
	if len(a.data) < len(b.data) {
		return a, b
	}
	return b, a
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
