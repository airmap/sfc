package sfc

import (
	"reflect"
	"testing"
)

func TestHilbertCellIterator(t *testing.T) {

	type tcase struct {
		dim      uint32
		order    uint32
		tier     uint32
		mask     []Bitmask
		expected [][]Bitmask
	}

	fn := func(t *testing.T, tc tcase) {
		uut := Hilbert{dim: tc.dim, order: tc.order}
		it := uut.cellIterator(tc.tier, tc.mask)

		result := make([][]Bitmask, 0)
		for it() {
			cellCp := make([]Bitmask, len(tc.mask))
			copy(cellCp, tc.mask)
			result = append(result, cellCp)
		}

		if reflect.DeepEqual(result, tc.expected) == false {
			t.Errorf("invalid result, expected %v got %v", tc.expected, result)
		}
	}

	// results were verified against the original C code.
	tcases := map[string]tcase{
		"test1": {
			dim:      2,
			order:    3,
			tier:     0,
			mask:     []Bitmask{0, 0},
			expected: [][]Bitmask{{0, 0}, {4, 0}, {0, 4}, {4, 4}},
		},
		"test2": {
			dim:      3,
			order:    3,
			tier:     1,
			mask:     []Bitmask{4, 0, 4},
			expected: [][]Bitmask{{4, 0, 4}, {6, 0, 4}, {4, 2, 4}, {6, 2, 4}, {4, 0, 6}, {6, 0, 6}, {4, 2, 6}, {6, 2, 6}},
		},
	}

	for k, v := range tcases {
		tc := v
		t.Run(k, func(t *testing.T) { fn(t, tc) })

	}
}
