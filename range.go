package sfc

import (
	"sort"
)

// Range represents a range in 1 dimensional space. E.g. Hilbert space.
type Range struct {
	MinValue Bitmask
	MaxValue Bitmask
}

// Ranges is a slice of multiple ranges
type Ranges []Range

// implement sort interface

func (r Ranges) Len() int      { return len(r) }
func (r Ranges) Swap(i, j int) { r[i], r[j] = r[j], r[i] }
func (r Ranges) Less(i, j int) bool {
	return r[i].MinValue < r[j].MinValue
}

// joinRanges takes a slice of ranges and combines any overlapping or adjacent
// ranges into single entries.
//
// The slice is modified in place and a new slice with the subset of ranges is
// returned.
func joinRanges(in Ranges) Ranges {
	sort.Sort(in)

	out := in[:1]

	for i := 0; i < len(in); i++ {
		// last element in out
		lo := len(out) - 1
		if in[i].MinValue == 0 || in[i].MinValue-1 <= out[lo].MaxValue {
			if in[i].MaxValue > out[lo].MaxValue {
				out[lo].MaxValue = in[i].MaxValue
			}
		} else {
			out = append(out, in[i])
		}
	}

	return out
}
