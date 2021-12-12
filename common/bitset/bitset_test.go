package bitset

import (
	"fmt"
	"math/rand"
	"reflect"
	"testing"

	"github.com/flu-network/client/common"
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

		b1 = NewBitset(1)
		b1.Set(0)
		if !b1.Full() {
			t.Fatalf("an full bitset of size 1 should be full")
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
			}
			if b1.Get(uint64(i)) != true {
				t.Fatalf("Expected filled bitset to actually be filled but %d was unset", i)
			}
		}
	})
}

func TestCount(t *testing.T) {
	t.Run("Count() does not basics/mutation", func(t *testing.T) {
		b1 := NewBitset(100)

		// empty case
		if c := b1.Count(); c != 0 {
			t.Fatalf("New bitset should have count() == 0. Got %d\n", c)
		}

		// set twice and check that it works
		b1.Set(0)
		b1.Set(99)
		if c := b1.Count(); c != 2 {
			t.Fatalf("Count inaccurate: Expected %d but got %d\n", 2, c)
		}

		// do it again, just to make sure that count()ing didn't mutate anything.
		if c := b1.Count(); c != 2 {
			t.Fatalf("Count inaccurate: Expected %d but got %d\n", 2, c)
		}
	})

	t.Run("Count() accuracy", func(t *testing.T) {
		sizes := []int{31}

		for _, size := range sizes {
			b1 := NewBitset(size)
			setCount := 0
			// set roughly 10% of available slots
			for i := 0; i < size/10; i++ {
				if !b1.Get(uint64(i)) {
					b1.Set(uint64(rand.Intn(size)))
					setCount++
				}
			}

			if c := b1.Count(); setCount != c {
				t.Fatalf("Count inaccurate at size %d: Expected %d but got %d\n", size, setCount, c)
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
			t.Fatalf("Deserialized bitset size did not match original: %d != %d\n",
				b1.size, b2.size)
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

func TestOverlap(t *testing.T) {
	type expectation struct {
		desc   string
		bitset Bitset
		input  []uint16
		output []uint16
	}

	testCases := []expectation{
		{
			desc:   "no overlap",
			bitset: *NewBitset(10).Fill(),
			input:  []uint16{11, 15},
			output: []uint16{},
		},
		{
			desc:   "full overlap",
			bitset: *NewBitset(10).Fill(),
			input:  []uint16{0, 9},
			output: []uint16{0, 9},
		},
		{
			desc:   "overlap with start of query",
			bitset: *NewBitset(10).Fill(),
			input:  []uint16{5, 15},
			output: []uint16{5, 9},
		},
		{
			desc:   "overlap with end of query",
			bitset: *NewBitset(10).Fill().Unset(0).Unset(1),
			input:  []uint16{0, 5},
			output: []uint16{2, 5},
		},
		{
			desc:   "disjoint overlap",
			bitset: *NewBitset(10).Fill().Unset(1).Unset(5).Unset(8),
			input:  []uint16{0, 9},
			output: []uint16{0, 0, 2, 4, 6, 7, 9, 9},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.desc, func(t *testing.T) {
			result := testCase.bitset.Overlap(testCase.input)
			if !reflect.DeepEqual(result, testCase.output) {
				t.Fatalf("Expected %v to equal %v\n", result, testCase.output)
			}
		})
	}
}

func TestRanges(t *testing.T) {
	type expectation struct {
		desc   string
		bitset Bitset
		output []common.Range
	}

	testCases := []expectation{
		{
			desc:   "single empty",
			bitset: *NewBitset(1),
			output: []common.Range{common.NewRange(0, 0)},
		},
		{
			desc:   "single filled",
			bitset: *NewBitset(1).Fill(),
			output: make([]common.Range, 0, 2),
		},
		{
			desc:   "left single empty",
			bitset: *NewBitset(10).Fill().Unset(0),
			output: []common.Range{common.NewRange(0, 0)},
		},
		{
			desc:   "left many empty",
			bitset: *NewBitset(10).Fill().Unset(0).Unset(1),
			output: []common.Range{common.NewRange(0, 1)},
		},
		{
			desc:   "right single empty",
			bitset: *NewBitset(10).Fill().Unset(9),
			output: []common.Range{common.NewRange(9, 9)},
		},
		{
			desc:   "right many empty",
			bitset: *NewBitset(10).Fill().Unset(9).Unset(8),
			output: []common.Range{common.NewRange(8, 9)},
		},
		{
			desc:   "mid single empty",
			bitset: *NewBitset(10).Fill().Unset(4),
			output: []common.Range{common.NewRange(4, 4)},
		},
		{
			desc:   "mid multiple empty",
			bitset: *NewBitset(10).Fill().Unset(4).Unset(5),
			output: []common.Range{common.NewRange(4, 5)},
		},
		{
			desc:   "mid disjoint empty",
			bitset: *NewBitset(10).Fill().Unset(2).Unset(3).Unset(5),
			output: []common.Range{common.NewRange(2, 3), common.NewRange(5, 5)},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.desc, func(t *testing.T) {
			result := testCase.bitset.UnfilledRanges()
			if !reflect.DeepEqual(result, testCase.output) {
				t.Fatalf("Expected %v to equal %v\n", result, testCase.output)
			}
		})
	}
}
