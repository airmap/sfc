package sfc_test

import (
	"reflect"
	"sort"
	"testing"

	"github.com/airmap/sfc"
)

// bruteAllValues calculates the min and max value within a given hilbert
// bounding box by brute force.
func bruteAllValues(bounds sfc.Box, coord sfc.Point, hc *sfc.Hilbert,
	dim uint32, values *[]sfc.Bitmask) {

	for coord[dim] = bounds[dim].Min; coord[dim] <= bounds[dim].Max; coord[dim]++ {

		if int(dim) == len(bounds)-1 {
			value := sfc.Encode(sfc.Bitmask(hc.Order()), coord)
			*values = append(*values, value)
		} else {
			bruteAllValues(bounds, coord, hc, dim+1, values)
		}
	}
}

func TestHilbertDecomposeSpans(t *testing.T) {

	type tcase struct {
		dim      uint32
		order    uint32
		minTier  uint32
		maxTier  uint32
		bounds   sfc.Box
		expected sfc.Spans
	}

	fn := func(t *testing.T, tc tcase) {
		uut, err := sfc.NewHilbert(tc.dim, tc.order)
		if err != nil {
			t.Fatalf("error creating hilbert curve, %v", err)
		}

		result, err := uut.DecomposeSpans(tc.minTier, tc.maxTier, &tc.bounds)
		if err != nil {
			t.Fatalf("error decomposing region, %v", err)
		}
		// evaluate consistent results.
		sort.Sort(result)

		if reflect.DeepEqual(result, tc.expected) == false {
			t.Errorf("invalid result, expected %v got %v", tc.expected, result)
		}
	}

	tcases := map[string]tcase{
		"test1": {
			dim:     2,
			order:   3,
			minTier: 0,
			maxTier: 0,
			bounds: sfc.NewBox(
				[]sfc.Bitmask{2, 1},
				[]sfc.Bitmask{4, 5},
			),
			expected: sfc.Spans{{Min: 0, Max: 63}},
		},
		"test2": {
			dim:     2,
			order:   3,
			minTier: 0,
			maxTier: 1,
			bounds: sfc.NewBox(
				[]sfc.Bitmask{2, 1},
				[]sfc.Bitmask{4, 5},
			),
			expected: sfc.Spans{{Min: 4, Max: 11}, {Min: 28, Max: 35}, {Min: 52, Max: 59}},
		},
		"test3": {
			dim:     2,
			order:   3,
			minTier: 0,
			maxTier: 2,
			bounds: sfc.NewBox(
				[]sfc.Bitmask{2, 1},
				[]sfc.Bitmask{4, 5},
			),
			expected: sfc.Spans{{Min: 6, Max: 11}, {Min: 28, Max: 32}, {Min: 35, Max: 35}, {Min: 53, Max: 54}, {Min: 57, Max: 57}},
		},
	}

	for k, v := range tcases {
		tc := v
		t.Run(k, func(t *testing.T) { fn(t, tc) })

	}
}

func BenchmarkHilbertDecomposeSpans(b *testing.B) {

	uut, err := sfc.NewHilbert(2, 32)
	if err != nil {
		b.Fatalf("error creating hilbert curve, %v", err)
	}

	for i := 0; i < b.N; i++ {
		box := sfc.NewBox(
			sfc.Point{32000, 35000},
			sfc.Point{45000, 38000},
		)
		_, err := uut.DecomposeSpans(0, 32, &box)
		if err != nil {
			b.Fatalf("error decomposing region, %v", err)
		}
	}
}

// TestHilbertDecomposeSpans2 uses brute force to extract all values in a
// range and then compares them against the ranges returned. This will only
// flag an error if a range doesn't contain a value, not if it contains too
// many values.
func TestHilbertDecomposeSpans2(t *testing.T) {

	type tcase struct {
		dim     uint32
		order   uint32
		minTier uint32
		maxTier uint32
		bounds  sfc.Box
	}

	fn := func(t *testing.T, tc tcase) {
		uut, err := sfc.NewHilbert(tc.dim, tc.order)
		if err != nil {
			t.Fatalf("error creating hilbert curve, %v", err)
		}

		result, err := uut.DecomposeSpans(tc.minTier, tc.maxTier, &tc.bounds)
		if err != nil {
			t.Fatalf("error decomposing region, %v", err)
		}

		coord := make(sfc.Point, tc.dim)
		allValues := make([]sfc.Bitmask, 0)
		bruteAllValues(tc.bounds, coord, uut, 0, &allValues)

		for i := range allValues {
			foundIt := false
			for j := range result {
				if result[j].Min <= allValues[i] &&
					result[j].Max >= allValues[i] {
					foundIt = true
				}
			}

			if foundIt == false {
				t.Errorf("expected value (%v) isn't in any range (%v)", allValues[i], result)
			}
		}
	}

	tcases := map[string]tcase{
		"test1": {
			dim:     2,
			order:   3,
			minTier: 0,
			maxTier: 0,
			bounds: sfc.NewBox(
				[]sfc.Bitmask{2, 1},
				[]sfc.Bitmask{4, 5},
			),
		},
		"test2": {
			dim:     3,
			order:   3,
			minTier: 0,
			maxTier: 2,
			bounds: sfc.NewBox(
				[]sfc.Bitmask{2, 1, 2},
				[]sfc.Bitmask{4, 5, 7},
			),
		},
	}

	for k, v := range tcases {
		tc := v
		t.Run(k, func(t *testing.T) { fn(t, tc) })

	}
}

func TestHilbertDecomposeRegion(t *testing.T) {

	type tcase struct {
		dim      uint32
		order    uint32
		minTier  uint32
		maxTier  uint32
		bounds   sfc.Box
		expected []sfc.Cell
	}

	fn := func(t *testing.T, tc tcase) {
		uut, err := sfc.NewHilbert(tc.dim, tc.order)
		if err != nil {
			t.Fatalf("error creating hilbert curve, %v", err)
		}

		result, err := uut.DecomposeRegion(tc.minTier, tc.maxTier, &tc.bounds)
		if err != nil {
			t.Fatalf("error decomposing region, %v", err)
		}

		// order doesn't matter so try sorting if this test ever fails
		if reflect.DeepEqual(result, tc.expected) == false {
			t.Errorf("invalid result, expected %v got %v", tc.expected, result)
		}
	}

	tcases := map[string]tcase{
		"test1": {
			dim:     2,
			order:   3,
			minTier: 0,
			maxTier: 0,
			bounds: sfc.NewBox(
				[]sfc.Bitmask{2, 1},
				[]sfc.Bitmask{4, 5},
			),
			expected: []sfc.Cell{{Value: 0, Tier: 0}, {Value: 3, Tier: 0}, {Value: 1, Tier: 0}, {Value: 2, Tier: 0}},
		},
		"test2": {
			dim:     2,
			order:   3,
			minTier: 0,
			maxTier: 1,
			bounds: sfc.NewBox(
				[]sfc.Bitmask{3, 4},
				[]sfc.Bitmask{7, 7},
			),
			expected: []sfc.Cell{{Value: 7, Tier: 1}, {Value: 6, Tier: 1}, {Value: 2, Tier: 0}},
		},
		"test3": {
			dim:     3,
			order:   3,
			minTier: 0,
			maxTier: 2,
			bounds: sfc.NewBox(
				[]sfc.Bitmask{1, 2, 3},
				[]sfc.Bitmask{1, 2, 4},
			),
			expected: []sfc.Cell{{Value: 48, Tier: 2}, {Value: 123, Tier: 2}},
		},
		"test4": {
			dim:     2,
			order:   32,
			minTier: 0,
			maxTier: 31,
			bounds: sfc.NewBox(
				[]sfc.Bitmask{10000, 200000},
				[]sfc.Bitmask{10000, 200000},
			),
			expected: []sfc.Cell{{Value: 21714213632, Tier: 31}},
		},
	}

	for k, v := range tcases {
		tc := v
		t.Run(k, func(t *testing.T) { fn(t, tc) })

	}
}
