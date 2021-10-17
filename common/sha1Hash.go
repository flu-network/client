package common

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"math"
)

// Sha1Hash is a convenient wrapper around the 20 bytes that make up a sha1Hash. Should be passed
// by pointer to avoid making copies.
type Sha1Hash struct {
	Data [20]byte // The underlying data in the hash. Should not be manipulated directly
}

// Array returns a pointer to the underlying data
func (h *Sha1Hash) Array() *[20]byte {
	return &h.Data
}

// String returns a hex-encoded string representation of the hash.
func (h *Sha1Hash) String() string {
	return hex.EncodeToString(h.Data[:])
}

// Slice returns a slice of the underlying hash data. The data is not copied.
func (h *Sha1Hash) Slice() []byte {
	return h.Data[:]
}

// SliceCopy copies and returns a slice of the underlying hash data.
func (h *Sha1Hash) SliceCopy() []byte {
	result := make([]byte, 20)
	copy(result, h.Data[:])
	return result
}

// FromString reads the bytes from a string and attempts to copy them into the underlying hash.
// It assumes the input is valid, and returns itself for syntactic convenience. If the input is
// invalid it will panic.
func (h *Sha1Hash) FromString(str string) *Sha1Hash {
	bytes, err := hex.DecodeString(str)
	if err != nil {
		panic(err)
	}

	copy(h.Data[:], bytes)
	return h
}

// FromStringSafe reads the bytes from a string and attempts to copy them into the underlying hash.
// If the provided string isn't hex-encoded and of the right length, an error is returned.
func (h *Sha1Hash) FromStringSafe(str string) error {
	bytes, err := hex.DecodeString(str)

	if err != nil {
		return err
	}

	if len(bytes) != 20 {
		return fmt.Errorf("Expected 20-byte string but got %d", len(bytes))
	}

	copy(h.Data[:], bytes)
	return nil
}

// FromSlice assumes the input is a slice of 20 bytes, and copies those bytes into the underlying
// hash. If there's any ambiguity about whether the input is valid, use FromSliceSafe(). It returns
// itself for syntactic convenience.
func (h *Sha1Hash) FromSlice(s []byte) *Sha1Hash {
	copy(h.Data[:], s)
	return h
}

// FromSliceSafe validates the length of the input and then copies it into the underlying hash. If
// the input is the wrong length, it returns an error.
func (h *Sha1Hash) FromSliceSafe(s []byte) error {
	if len(s) != 20 {
		return fmt.Errorf("Expected 20-byte string but got %d", len(s))
	}

	copy(h.Data[:], s)
	return nil
}

// Blank overwrites the underlying hash to FF...FF. Used through out flu as the 'null' hash. Returns
// itself.
func (h *Sha1Hash) Blank() *Sha1Hash {
	for i := 0; i < 20; i++ {
		h.Data[i] = 255
	}
	return h
}

// IsBlank returns true if the sha1hash is blank (i.e., FF...FF); i.e., a 'null' hash. Benchmarked
// to be 4x faster than the naive loop=based approach
func (h *Sha1Hash) IsBlank() bool {
	c1 := binary.BigEndian.Uint64(h.Data[:8])
	c2 := binary.BigEndian.Uint64(h.Data[8:16])
	c3 := binary.BigEndian.Uint32(h.Data[16:20])
	return c1 == math.MaxUint64 && c1 == c2 && c3 == math.MaxUint32
}
