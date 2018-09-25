package sfc_test

import (
	"reflect"
	"testing"

	"github.com/airmap/sfc"
)

// distance2 returns the distance^2 between two points
func distance2(p1, p2 []sfc.Bitmask) float64 {
	if len(p1) != len(p2) {
		panic("points must have the same number of dimensions")
	}

	sum := 0.0
	for i := range p1 {
		diff := p1[i] - p2[i]
		sum += float64(diff * diff)
	}

	return sum
}

// TestHilbertConsistency tests that the hilbert code gives self-consistent
// results. Coming up with test cases in 5 dimensions with order 10 is a
// difficult thing to do manually. This test simply ensures that each
// increment of 1 in the curve also results in a change of 1 in space. It
// also ensures that going from value -> point -> value gives a consistent
// result.
func TestHilbertConsistency(t *testing.T) {

	type tcase struct {
		dim        sfc.Bitmask
		order      sfc.Bitmask
		startValue sfc.Bitmask
		endValue   sfc.Bitmask
	}

	fn := func(t *testing.T, tc tcase) {
		pt := make([]sfc.Bitmask, tc.dim, tc.dim)
		lastPt := make([]sfc.Bitmask, tc.dim, tc.dim)

		for i := tc.startValue; i <= tc.endValue; i++ {
			sfc.Decode(tc.dim, tc.order, i, pt)

			if i != tc.startValue {
				dist := distance2(pt, lastPt)
				if dist > 1.01 || dist < 0.99 {
					t.Errorf("encoded point (%v) is not 1 distance from the last point (%v), expected 1 got %v", dist, pt, lastPt)
				}
			}

			enc := sfc.Encode(tc.dim, tc.order, pt)
			if enc != i {
				t.Errorf("original hilbert value isn't consistent with the encoded/decoded value, expected %v got %v", i, enc)
			}

			copy(lastPt, pt)
		}
	}

	// results were verified against the original C code.
	tcases := map[string]tcase{
		"test1": {
			dim:        2,
			order:      3,
			startValue: 0,
			endValue:   63,
		},
		"test2": {
			dim:        4,
			order:      3,
			startValue: 0,
			endValue:   511,
		},
		"test3": {
			dim:        4,
			order:      5,
			startValue: 200,
			endValue:   2000,
		},
		"test4": {
			dim:        5,
			order:      10,
			startValue: 10000000,
			endValue:   10001000,
		},
	}

	for k, v := range tcases {
		tc := v
		t.Run(k, func(t *testing.T) { fn(t, tc) })

	}
}

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
