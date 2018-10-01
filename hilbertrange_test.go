package sfc_test

import (
	"reflect"
	"sort"
	"testing"

	"github.com/airmap/sfc"
)

// bruteMinMaxValue calculates the min and max value within a given hilbert
// bounding box by brute force.
func bruteAllValues(bounds sfc.Bounds, coord sfc.Point, hc *sfc.Hilbert,
	dim uint32, values *[]sfc.Bitmask) {

	for coord[dim] = bounds.Min[dim]; coord[dim] <= bounds.Max[dim]; coord[dim]++ {

		if int(dim) == len(bounds.Min)-1 {
			value := sfc.Encode(sfc.Bitmask(hc.Order()), coord)
			*values = append(*values, value)
		} else {
			bruteAllValues(bounds, coord, hc, dim+1, values)
		}
	}
}

func TestHilbertDecomposeRanges(t *testing.T) {

	type tcase struct {
		dim      uint32
		order    uint32
		minTier  uint32
		maxTier  uint32
		bounds   sfc.Intersecter
		expected sfc.Ranges
	}

	fn := func(t *testing.T, tc tcase) {
		uut, err := sfc.NewHilbert(tc.dim, tc.order)
		if err != nil {
			t.Fatalf("error creating hilbert curve, %v", err)
		}

		result, err := uut.DecomposeRanges(tc.minTier, tc.maxTier, tc.bounds)
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
			bounds: &sfc.Bounds{
				Min: []sfc.Bitmask{2, 1},
				Max: []sfc.Bitmask{4, 5},
			},
			expected: sfc.Ranges{{MinValue: 0, MaxValue: 63}},
		},
		"test2": {
			dim:     2,
			order:   3,
			minTier: 0,
			maxTier: 1,
			bounds: &sfc.Bounds{
				Min: []sfc.Bitmask{2, 1},
				Max: []sfc.Bitmask{4, 5},
			},
			expected: sfc.Ranges{{MinValue: 4, MaxValue: 11}, {MinValue: 28, MaxValue: 35}, {MinValue: 52, MaxValue: 59}},
		},
		"test3": {
			dim:     2,
			order:   3,
			minTier: 0,
			maxTier: 2,
			bounds: &sfc.Bounds{
				Min: []sfc.Bitmask{2, 1},
				Max: []sfc.Bitmask{4, 5},
			},
			expected: sfc.Ranges{{MinValue: 6, MaxValue: 11}, {MinValue: 28, MaxValue: 32}, {MinValue: 35, MaxValue: 35}, {MinValue: 53, MaxValue: 54}, {MinValue: 57, MaxValue: 57}},
		},
	}

	for k, v := range tcases {
		tc := v
		t.Run(k, func(t *testing.T) { fn(t, tc) })

	}
}

func BenchmarkHilbertDecomposeRanges(b *testing.B) {

	uut, err := sfc.NewHilbert(2, 32)
	if err != nil {
		b.Fatalf("error creating hilbert curve, %v", err)
	}

	for i := 0; i < b.N; i++ {
		_, err := uut.DecomposeRanges(0, 32, &sfc.Bounds{
			Min: sfc.Point{32000, 35000},
			Max: sfc.Point{45000, 38000},
		})
		if err != nil {
			b.Fatalf("error decomposing region, %v", err)
		}
	}
}

// TestHilbertDecomposeRanges2 uses brute force to extract all values in a
// range and then compares them against the ranges returned. This will only
// flag an error if a range doesn't contain a value, not if it contains too
// many values.
func TestHilbertDecomposeRanges2(t *testing.T) {

	type tcase struct {
		dim     uint32
		order   uint32
		minTier uint32
		maxTier uint32
		bounds  *sfc.Bounds
	}

	fn := func(t *testing.T, tc tcase) {
		uut, err := sfc.NewHilbert(tc.dim, tc.order)
		if err != nil {
			t.Fatalf("error creating hilbert curve, %v", err)
		}

		result, err := uut.DecomposeRanges(tc.minTier, tc.maxTier, tc.bounds)
		if err != nil {
			t.Fatalf("error decomposing region, %v", err)
		}

		coord := make(sfc.Point, tc.dim)
		allValues := make([]sfc.Bitmask, 0)
		bruteAllValues(*tc.bounds, coord, uut, 0, &allValues)

		for i := range allValues {
			foundIt := false
			for j := range result {
				if result[j].MinValue <= allValues[i] &&
					result[j].MaxValue >= allValues[i] {
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
			bounds: &sfc.Bounds{
				Min: []sfc.Bitmask{2, 1},
				Max: []sfc.Bitmask{4, 5},
			},
		},
		"test2": {
			dim:     3,
			order:   3,
			minTier: 0,
			maxTier: 2,
			bounds: &sfc.Bounds{
				Min: []sfc.Bitmask{2, 1, 2},
				Max: []sfc.Bitmask{4, 5, 7},
			},
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
		bounds   sfc.Intersecter
		expected []sfc.Cell
	}

	fn := func(t *testing.T, tc tcase) {
		uut, err := sfc.NewHilbert(tc.dim, tc.order)
		if err != nil {
			t.Fatalf("error creating hilbert curve, %v", err)
		}

		result, err := uut.DecomposeRegion(tc.minTier, tc.maxTier, tc.bounds)
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
			bounds: &sfc.Bounds{
				Min: []sfc.Bitmask{2, 1},
				Max: []sfc.Bitmask{4, 5},
			},
			expected: []sfc.Cell{{Value: 0, Tier: 0}, {Value: 3, Tier: 0}, {Value: 1, Tier: 0}, {Value: 2, Tier: 0}},
		},
		"test2": {
			dim:     2,
			order:   3,
			minTier: 0,
			maxTier: 1,
			bounds: &sfc.Bounds{
				Min: []sfc.Bitmask{3, 4},
				Max: []sfc.Bitmask{7, 7},
			},
			expected: []sfc.Cell{{Value: 7, Tier: 1}, {Value: 6, Tier: 1}, {Value: 2, Tier: 0}},
		},
		"test3": {
			dim:     3,
			order:   3,
			minTier: 0,
			maxTier: 2,
			bounds: &sfc.Bounds{
				Min: []sfc.Bitmask{1, 2, 3},
				Max: []sfc.Bitmask{1, 2, 4},
			},
			expected: []sfc.Cell{{Value: 48, Tier: 2}, {Value: 123, Tier: 2}},
		},
	}

	for k, v := range tcases {
		tc := v
		t.Run(k, func(t *testing.T) { fn(t, tc) })

	}
}
