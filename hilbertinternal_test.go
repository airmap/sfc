package sfc

import (
	"reflect"
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

func TestHilbertBoxPt(t *testing.T) {

	type tcase struct {
		nBits Bitmask
		findMin bool
		c1 []Bitmask
		c2 []Bitmask
		expected []Bitmask
	}

	fn := func(t *testing.T, tc tcase) {
		hilbert_box_pt(tc.nBits, tc.findMin, tc.c1, tc.c2)

		if reflect.DeepEqual(tc.expected, tc.c1) == false {
			t.Errorf("invalid result, expected %v got %v", tc.expected, tc.c1)
		}
	}

	// results were verified against the original C code.
	tcases := map[string]tcase{
		"test1": {
			nBits: 3,
			findMin: false,
			c1: []Bitmask{4, 3},
			c2: []Bitmask{7, 6},
			expected: []Bitmask{4, 6},
		},
		"test2": {
			nBits: 3,
			findMin: true,
			c1: []Bitmask{4, 3},
			c2: []Bitmask{7, 6},
			expected: []Bitmask{7, 3},
		},
		"test3": {
			nBits: 3,
			findMin: true,
			c1: []Bitmask{2, 0},
			c2: []Bitmask{4, 3},
			expected: []Bitmask{2, 2},
		},
		"test4": {
			nBits: 3,
			findMin: false,
			c1: []Bitmask{2, 3},
			c2: []Bitmask{5, 7},
			expected: []Bitmask{2, 5},
		},
	}

	for k, v := range tcases {
		tc := v
		t.Run(k, func(t *testing.T) { fn(t, tc) })

	}
}

func TestHilbertPropogateIntBits(t *testing.T) {

	type tcase struct {
		d Bitmask
		c []Bitmask
		y Bitmask
		fold Bitmask
		expected []Bitmask
	}

	fn := func(t *testing.T, tc tcase) {
		propogateIntBits(tc.d, tc.c, tc.y, tc.fold)

		if reflect.DeepEqual(tc.c, tc.expected) == false {
			t.Errorf("invalid result, expected %v got %v", tc.expected, tc.c)
		}
	}

	// results were verified against the original C code.
	tcases := map[string]tcase{
		"test1": {
			d:      1,
			c:    []Bitmask{0x100, 0x200},
			y: 9,
			fold: 0,
			expected: []Bitmask{0x100, 0x1FF},
		},
		"test2": {
			d:      1,
			c:    []Bitmask{0x100, 0x200},
			y: 9,
			fold: 1,
			expected: []Bitmask{0x100, 0x0},
		},
		"test3": {
			d:      1,
			c:    []Bitmask{0x100, 0x200},
			y: 8,
			fold: 0,
			expected: []Bitmask{0x100, 0x300},
		},
		"test4": {
			d:      1,
			c:    []Bitmask{0x100, 0x200},
			y: 8,
			fold: 1,
			expected: []Bitmask{0x100, 0x300},
		},
		"test5": {
			d:      0,
			c:    []Bitmask{0x123, 0xABC},
			y: 8,
			fold: 1,
			expected: []Bitmask{0x23, 0xABC},
		},
		"test6": {
			d:      0,
			c:    []Bitmask{0x123, 0xABC},
			y: 8,
			fold: 0,
			expected: []Bitmask{0xFF, 0xABC},
		},
		"test7": {
			d:      0,
			c:    []Bitmask{0x123, 0xABC},
			y: 5,
			fold: 1,
			expected: []Bitmask{0x103, 0xABC},
		},
		"test8": {
			d:      0,
			c:    []Bitmask{0x123, 0xABC},
			y: 5,
			fold: 0,
			expected: []Bitmask{0x11F, 0xABC},
		},
	}

	for k, v := range tcases {
		tc := v
		t.Run(k, func(t *testing.T) { fn(t, tc) })

	}
}
