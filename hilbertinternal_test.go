package sfc

import (
	"testing"
)

func TestHilbertBitTranspose(t *testing.T) {

	type tcase struct {
		dim      Bitmask
		order    Bitmask
		value    Bitmask
		expected Bitmask
	}

	fn := func(t *testing.T, tc tcase) {
		result := bitTranspose(tc.dim, tc.order, tc.value)

		if result != tc.expected {
			t.Errorf("invalid result, expected 0x%X got 0x%X", tc.expected, result)
		}
	}

	// results were verified against the original C code.
	tcases := map[string]tcase{
		"test1": {
			dim:      2,
			order:    3,
			value:    0x7,
			expected: 0x15,
		},
		"test2": {
			dim:      2,
			order:    3,
			value:    0x3,
			expected: 0x5,
		},
		"test3": {
			dim:      3,
			order:    12,
			value:    0x321,
			expected: 0x9008001,
		},
	}

	for k, v := range tcases {
		tc := v
		t.Run(k, func(t *testing.T) { fn(t, tc) })

	}
}
