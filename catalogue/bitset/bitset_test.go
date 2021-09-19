package bitset

import (
	"fmt"
	"math/rand"
	"reflect"
	"strconv"
	"testing"
)

const (
	start = 1000 * 1000
	limit = 100 * 1000
	items = 10 * 1000
)

func TestBitsetInit(t *testing.T) {
	t.Run("Test initialization", func(t *testing.T) {
		b := NewBitset(128)
		if len(b.data) != 2 {
			t.Fatalf("Get should've returned %d for %d\n", 2, len(b.data))
		}

		b = NewBitset(129)
		if len(b.data) != 3 {
			t.Fatalf("Get should've returned %d for %d\n", 3, len(b.data))
		}
	})

	t.Run("size tracking", func(t *testing.T) {
		maxSize := 0
		b1 := NewBitset(maxSize)
		m1 := make(map[uint64]struct{})

		// Call Set a bunch
		for i := 0; i < items; i++ {
			x := start + uint64(rand.Intn(limit))
			maxSize = max(maxSize, int(x)+1)
			b1.Set(x)
			if maxSize != b1.size {
				t.Fatalf("Bitset size should be %d but got %d\n", maxSize, b1.size)
			}
			m1[x] = struct{}{}
		}
	})
}

func TestBitsetOperations(t *testing.T) {
	t.Run("Basic Operation", func(t *testing.T) {
		b := NewBitset(0)

		for i := 0; i < items; i++ {
			x := uint64(rand.Intn(limit))
			b.Set(x)
			if b.Get(x) != true {
				t.Fatalf("Get should've returned %t for %d after Set\n", true, x)
			}
			b.Unset(x)
			if b.Get(x) != false {
				t.Fatalf("Get should've returned %t for %d after Unset\n", false, x)
			}
		}
	})

	t.Run("Test Union() & Intersect()", func(t *testing.T) {
		b1 := NewBitset(0)
		m1 := make(map[uint64]struct{})

		// Call Set a bunch
		for i := 0; i < items; i++ {
			x := start + uint64(rand.Intn(limit))
			b1.Set(x)
			m1[x] = struct{}{}
		}

		// Make sure subsequent Get works as expected
		for x := uint64(0); x < start+limit+wordSize; x++ {
			_, ok := m1[x]
			if ok != b1.Get(x) {
				t.Fatalf("Get should've returned %t for %d\n", ok, x)
			}
		}

		b2 := NewBitset(0)
		m2 := make(map[uint64]struct{})

		// Call Set a bunch
		for i := 0; i < items; i++ {
			x := uint64(rand.Intn(limit))
			b2.Set(x)
			m2[x] = struct{}{}
		}

		union := b1.Union(b2)
		intersect := b1.Intersect(b2)
		for x := uint64(0); x < start+limit+wordSize; x++ {
			_, ok1 := m1[x]
			_, ok2 := m2[x]
			if (ok1 || ok2) != union.Get(x) {
				t.Fatalf("Union: Get should've returned %t for %d\n", ok1 || ok2, x)
			}
			if (ok1 && ok2) != intersect.Get(x) {
				t.Fatalf("Intersect: Get should've returned %t for %d\n", ok1 && ok2, x)
			}
		}
	})
}

func TestBitsetFillAndFull(t *testing.T) {
	t.Run("Empty case", func(t *testing.T) {
		b1 := NewBitset(0)
		b1.Fill()
		if !b1.Full() {
			t.Fatalf("Even an empty filled bitset should be full")
		}
	})

	t.Run("single int case", func(t *testing.T) {
		b1 := NewBitset(0)
		if !b1.Full() {
			t.Fatalf("an empty bitset of size 0 should be full")
		}

		b1.Set(15)
		if b1.Full() {
			t.Fatalf("a non-full bitset should not be full")
		}

		b1.Fill()
		if !b1.Full() {
			t.Fatalf("a filled bitset should be full")
		}
	})

	t.Run("exact int case", func(t *testing.T) {
		limit := 64
		b1 := NewBitset(limit)
		if b1.size != limit {
			t.Fatalf("Bitset initialized with wrong size\n")
		}

		b1.Fill()
		prev := -1
		for i := 0; i < limit; i++ {
			if i%64 == 0 {
				prev++
				fmt.Printf("%d: %s\n", prev, strconv.FormatInt(int64(b1.data[prev]), 2))
			}
			if b1.Get(uint64(i)) != true {
				t.Fatalf("Expected filled bitset to actually be filled but %d was unset", i)
			}
		}
	})

	t.Run("many int case", func(t *testing.T) {
		limit := 123456
		b1 := NewBitset(limit)
		if b1.size != limit {
			t.Fatalf("Bitset initialized with wrong size\n")
		}

		b1.Fill()
		prev := -1
		for i := 0; i < limit; i++ {
			if i%64 == 0 {
				prev++
				fmt.Printf("%d: %s\n", prev, strconv.FormatInt(int64(b1.data[prev]), 2))
			}
			if b1.Get(uint64(i)) != true {
				fmt.Printf("%d: %s\n", prev+1, strconv.FormatInt(int64(b1.data[prev+1]), 2))
				t.Fatalf("Expected filled bitset to actually be filled but %d was unset", i)
			}
		}
	})
}

func TestSerialization(t *testing.T) {
	for size := 9; size < 10000; size++ {

		b1 := NewBitset(size)
		// set roughly 10% of available slots
		for i := 0; i < size/10; i++ {
			b1.Set(uint64(rand.Intn(size)))
		}

		data := b1.Serialize()
		b2, err := Deserialize(data)

		if err != nil {
			t.Fatal(err)
		}

		if b1.size != b2.size {
			t.Fatalf("Deserialized bitset size did not match original: %d != %d\n", b1.size, b2.size)
		}

		if reflect.DeepEqual(b1.data, b2.data) != true {
			for i, b1Data := range b1.data {
				b2Data := b2.data[i]
				if b1Data != b2Data {
					pos := fmt.Sprintf("Deserialization error at index %d\n", i)
					l1 := fmt.Sprintf("Original: % 64b\n", b1Data)
					l2 := fmt.Sprintf("Deserial: % 64b\n", b2Data)
					t.Fatalf("%s%s%s", pos, l1, l2)
				}
			}
		}
	}
}
