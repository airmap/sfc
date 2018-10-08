package sfc

import (
	"sort"
)

// Span represents a span in 1 dimensional space. E.g. Hilbert space.
type Span struct {
	Min Bitmask
	Max Bitmask
}

// Spans is a slice of multiple spans
type Spans []Span

// implement sort interface

func (r Spans) Len() int      { return len(r) }
func (r Spans) Swap(i, j int) { r[i], r[j] = r[j], r[i] }
func (r Spans) Less(i, j int) bool {
	return r[i].Min < r[j].Min
}

// joinSpans takes a slice of spans and combines any overlapping or adjacent
// spans into single entries.
//
// The slice is modified in place and a new slice with the subset of spans is
// returned.
func joinSpans(in Spans) Spans {
	sort.Sort(in)

	out := in[:1]

	for i := range in {
		// last element in out
		lo := len(out) - 1
		if in[i].Min == 0 || in[i].Min-1 <= out[lo].Max {
			if in[i].Max > out[lo].Max {
				out[lo].Max = in[i].Max
			}
		} else {
			out = append(out, in[i])
		}
	}

	return out
}
