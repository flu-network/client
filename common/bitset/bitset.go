package bitset

import (
	"encoding/binary"
	"fmt"
	"math"
	"strings"

	"github.com/flu-network/client/common"
)

const wordSize = 64

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

// Size returns the number of elements in the bitset, both set and unset. Size is always
// non-negative
func (b *Bitset) Size() int {
	return b.size
}

// Count returns the number of elements in the bitset that are set (true)
func (b *Bitset) Count() int {
	result := uint64(0)
	for _, unit := range b.data {
		for unit > 0 {
			result += (unit & 1)
			unit = unit >> 1
		}
	}
	return int(result)
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
func (b *Bitset) Set(x uint64) *Bitset {
	offset := int(x) / wordSize
	if offset > len(b.data)-1 {
		diff := offset - (len(b.data) - 1)
		b.data = append(b.data, make([]uint64, diff)...)
	}
	b.data[x/wordSize] |= (1 << (x % wordSize))
	b.size = max(b.size, int(x)+1)
	return b
}

// Fill sets all items between 0:size to true. Returns itself for syntactic convenience
func (b *Bitset) Fill() *Bitset {
	if b.size == 0 {
		return b
	}

	maxOffset := (b.size - 1) / wordSize
	for i := 0; i < maxOffset; i++ {
		b.data[i] = math.MaxUint64
	}

	lastIndex := (b.size - 1) % wordSize
	b.data[maxOffset] = (((1 << lastIndex) - 1) << 1) | 1
	return b
}

// Full returns true if all items between 0:size are set to true, and false if not.
func (b *Bitset) Full() bool {
	if b.size == 0 {
		return true
	}

	maxOffset := (b.size - 1) / wordSize
	for i := 0; i < maxOffset; i++ {
		if b.data[i] != math.MaxUint64 {
			return false
		}
	}

	lastIndex := (b.size - 1) % wordSize
	return b.data[maxOffset] == (((1<<lastIndex)-1)<<1)|1
}

// Unset sets the specified index to False. If the specified index is out of bounds, nothing
// happens. Does not cause size to shrink. Returns itself
func (b *Bitset) Unset(x uint64) *Bitset {
	offset := int(x) / wordSize
	if offset >= len(b.data) {
		return b
	}
	b.data[offset] = b.data[offset] - (1 << (x % wordSize))
	return b
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

// Overlap takes a sorted  list of ranges ([]uint16 of even length) where each consecutive pair of
// numbers represents an inclusive range [start, end] of bits. The return value is a sorted list of
// non-overallping ranges that are set to 'on' within the underlying bitset
func (b *Bitset) Overlap(ranges []uint16) []uint16 {
	result := make([]uint16, 0, len(ranges))
	for i := 0; i < len(ranges); i += 2 {
		start, end := ranges[i], ranges[i+1]
		result = append(result, b.filledRanges(int(start), int(end))...)
	}
	return result
}

// Ranges returns a sorted, non-overallping list of ranges of the underlying bitset that are set to
// true.
func (b *Bitset) Ranges() []uint16 {
	return b.filledRanges(0, b.size)
}

func (b *Bitset) filledRanges(start, end int) []uint16 {
	result := make([]uint16, 0, 2)
	rStart, rEnd := start, start-1
	for i := start; i <= end; i++ {
		if b.Get(uint64(i)) {
			rEnd = i
		} else {
			if rEnd >= rStart {
				result = append(result, uint16(rStart), uint16(rEnd))
			}
			rStart = i + 1
		}
	}
	if rEnd >= rStart {
		result = append(result, uint16(rStart), uint16(rEnd))
	}
	return result
}

// UnfilledRanges returns a sorted, non-overlapping list of ranges of the underlying bitset that are
// set to false.
func (b *Bitset) UnfilledRanges() []common.Range {
	start, end := 0, b.size
	result := make([]common.Range, 0, 1)
	rStart, rEnd := start, start-1
	for i := start; i < end; i++ {
		if !b.Get(uint64(i)) {
			rEnd = i
		} else {
			if rEnd >= rStart {
				result = append(result, common.NewRange(uint16(rStart), uint16(rEnd)))
			}
			rStart = i + 1
		}
	}
	if rEnd >= rStart {
		result = append(result, common.NewRange(uint16(rStart), uint16(rEnd)))
	}
	return result
}

// UnfilledItems returns the first `count` items in the bitset that are set to false. If there are
// fewer than `count` items it returns that many.
// TODO: this is super gross and inefficient... do something better
func (b *Bitset) UnfilledItems(count int) []uint16 {
	result := make([]uint16, 0, count)

	for i := (uint64(0)); i < uint64(b.size) && len(result) < cap(result); i++ {
		if !b.Get(i) {
			result = append(result, uint16(i))
		}
	}

	return result
}

// Print returns a 0/1-formatted string representation of the underlying bitset
func (b *Bitset) Print() string {
	result := strings.Builder{}
	result.WriteString(fmt.Sprintf("size: %d\n", b.size))
	for segmentIndex := range b.data {
		result.WriteString(fmt.Sprintf("%064b\n", b.data[segmentIndex]))
	}
	return result.String()
}

func (b *Bitset) Deb() []uint64 {
	return b.data
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
		return nil, fmt.Errorf("data segment of bitset data must be a multiple of 8. Got %d", l)
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

// Copy returns a copy of the bitset. Mutating the copy will not affect the original
func (b *Bitset) Copy() *Bitset {
	data := make([]uint64, len(b.data), cap(b.data))
	copy(data, b.data)
	return &Bitset{data: data, size: b.size}
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
