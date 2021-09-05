package bitset

import (
	"math/rand"
	"testing"
)

const (
	start = 1000 * 1000
	limit = 100 * 1000
	items = 10 * 1000
)

func TestBitset(t *testing.T) {
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
