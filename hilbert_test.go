package sfc_test

import (
	"reflect"
	"testing"

	"github.com/airmap/sfc"
)

func TestHilbertDecode(t *testing.T) {

	type tcase struct {
		dim      sfc.Bitmask
		order    sfc.Bitmask
		value    sfc.Bitmask
		expected []sfc.Bitmask
	}

	fn := func(t *testing.T, tc tcase) {
		result := make([]sfc.Bitmask, tc.dim, tc.dim)
		sfc.Decode(tc.dim, tc.order, tc.value, result)

		if reflect.DeepEqual(result, tc.expected) == false {
			t.Errorf("invalid result, expected %v got %v", tc.expected, result)
		}
	}

	// results were verified against the original C code.
	tcases := map[string]tcase{
		"test1": {
			dim:      2,
			order:    3,
			value:    13,
			expected: []sfc.Bitmask{1, 2},
		},
		"test2": {
			dim:      2,
			order:    3,
			value:    46,
			expected: []sfc.Bitmask{6, 4},
		},
		"test3": {
			dim:      2,
			order:    2,
			value:    6,
			expected: []sfc.Bitmask{1, 3},
		},
		"test4": {
			dim:      1,
			order:    4,
			value:    13,
			expected: []sfc.Bitmask{13},
		},
	}

	for k, v := range tcases {
		tc := v
		t.Run(k, func(t *testing.T) { fn(t, tc) })

	}
}

func TestHilbertEncode(t *testing.T) {

	type tcase struct {
		dim      sfc.Bitmask
		order    sfc.Bitmask
		pt       []sfc.Bitmask
		expected sfc.Bitmask
	}

	fn := func(t *testing.T, tc tcase) {
		result := sfc.Encode(tc.dim, tc.order, tc.pt)

		if result != tc.expected {
			t.Errorf("invalid result, expected 0x%X got 0x%X", tc.expected, result)
		}
	}

	// results were verified against the original C code.
	tcases := map[string]tcase{
		"test1": {
			dim:      2,
			order:    3,
			pt:       []sfc.Bitmask{1, 2},
			expected: 13,
		},
		"test2": {
			dim:      2,
			order:    3,
			pt:       []sfc.Bitmask{2, 1},
			expected: 7,
		},
		"test3": {
			dim:      2,
			order:    3,
			pt:       []sfc.Bitmask{6, 1},
			expected: 61,
		},
		"test4": {
			dim:      2,
			order:    3,
			pt:       []sfc.Bitmask{4, 6},
			expected: 36,
		},
	}

	for k, v := range tcases {
		tc := v
		t.Run(k, func(t *testing.T) { fn(t, tc) })

	}
}
