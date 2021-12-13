package common

// Range is a tuple of inclusive start and end indices, each of which is limited to 16 bits
type Range struct {
	Start uint16
	End   uint16
}

// NewRange returns a range by value. Note: Since a Range is only 32 bits (smaller than a pointer),
// it is returned by value and not by pointer
func NewRange(start, end uint16) Range {
	return Range{
		Start: start,
		End:   end,
	}
}
