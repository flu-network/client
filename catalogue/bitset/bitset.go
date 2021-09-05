package bitset

const wordSize = 64

// Bitset is simple bitset backed by a []uint64. The size of the bitset is unbounded and it will
// grow as required to support any Set() operations invoked. Bitset provides no compression.
type Bitset struct {
	data []uint64
}

// NewBitset returns a new bitset of the specified size. Panics if size < 0. Providing a size is
// convenient for serialization, but not necessary. Get() and Set() operations work as described.
func NewBitset(size int) *Bitset {
	if size < 0 {
		panic("Bitmaps cannot have a negative size")
	}
	result := Bitset{}

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
}

// Unset sets the specified index to False. If the specified index is out of bounds, nothing
// happens.
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

// utilities
func shorterFirst(a, b *Bitset) (*Bitset, *Bitset) {
	if len(a.data) < len(b.data) {
		return a, b
	}
	return b, a
}
