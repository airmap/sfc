package sfc

import (
	"fmt"
)

// CellIterator is a function that iterates to the next cell in cellIterator
type CellIterator func() bool

// Cell represents a specific hilbert curve value at a specific tier/order
type Cell struct {
	Value Bitmask
	Tier  uint32
}

// cellIterator returns a function that enables iterating over 2 ^ dim cells
// at a given tier/location.
//
// tier - the tier to iterate over
//
// mask - the area to iterate around
func (hc *Hilbert) cellIterator(tier uint32, mask []Bitmask) CellIterator {
	cell := mask
	tierBit := Bitmask(1) << (Bitmask(hc.order) - Bitmask(tier) - 1)
	first := true

	return func() bool {
		if first {
			first = false
			return true
		}

		dim := 0
		// if this dim is rolling over.
		for cell[dim]&tierBit != 0 {
			// clear this dim
			cell[dim] ^= tierBit
			dim++
			if dim == int(hc.dim) {
				// clear all the lower bits to reset the state
				for i := uint32(0); i < hc.dim; i++ {
					cell[i] &= ^(tierBit - 1)
				}

				// all done
				return false
			}
		}

		cell[dim] ^= tierBit

		return true
	}
}

type decomposeCall struct {
	minTier uint32
	maxTier uint32
	bounds  Box
	region  Intersecter
}

// DecomposeSpans breaks a region up into a series of hilbert value spans.
//
// minTier - The minimum tier in the hilbert curve to start the decomposition.
// Setting this too high may result in a large number of spans.
//
// maxTier - The maximum tier to recurse down to during the decomposition.
// Setting maxTier to a high value may results in a very large number of
// spans.
func (hc *Hilbert) DecomposeSpans(minTier, maxTier uint32,
	region Intersecter) (Spans, error) {

	cell := make(Point, hc.dim, hc.dim)
	it := hc.cellIterator(0, cell)

	dc := decomposeCall{
		bounds:  Box{{cell[0], cell[1]}},
		minTier: minTier,
		maxTier: maxTier,
		region:  region,
	}

	result := Spans{}

	for it() {
		err := hc.decomposeSpans(0, cell.Clone(), &dc, &result)
		if err != nil {
			return Spans{}, err
		}
	}

	result = joinSpans(result)

	return result, nil
}

func (hc *Hilbert) decomposeSpans(tier uint32, cell Point, dc *decomposeCall,
	result *Spans) error {

	tierBit := Bitmask(1) << (Bitmask(hc.order) - Bitmask(tier) - 1)
	upperBits := tierBit - 1

	// calculate the upper bound
	dc.bounds = NewBox(cell, cell)
	for d := uint32(0); d < hc.dim; d++ {
		dc.bounds[d].Max |= upperBits
	}

	intersects, err := dc.region.Intersects(&dc.bounds)
	if err != nil {
		return err
	}
	// if the region intersects the bounds of this tier/cell
	if intersects {

		// if we're in the reporting span
		if tier >= dc.minTier {

			contains, err := dc.region.Contains(&dc.bounds)
			if err != nil {
				return err
			}

			// if we've reached the max tier, or are fully contained
			if tier == dc.maxTier || contains {

				value := Encode(Bitmask(hc.order), cell)
				tierValueBits := Bitmask(1) << ((hc.order - tier - 1) * hc.dim)
				tierValueBits--

				r := Span{
					Min: value & ^tierValueBits,
					Max: value | tierValueBits,
				}
				*result = append(*result, r)
			} else {
				// if we only partially overlap and we aren't at the max
				// tier

				it := hc.cellIterator(tier+1, cell)
				// go through all the child cells at this tier
				for it() {
					hc.decomposeSpans(tier+1, cell, dc, result)
				}
			}
			// if we aren't in the reporting span, just recurse
		} else {
			it := hc.cellIterator(tier+1, cell)
			// go through all the child cells at this tier
			for it() {
				hc.decomposeSpans(tier+1, cell, dc, result)
			}
		}
	}

	return nil
}

// DecomposeRegion breaks a region up into a series of hilbert value cells.
//
// minTier - The minimum tier in the hilbert curve to start the decomposition.
// Setting this too high may result in a large number of spans.
//
// maxTier - The maximum tier to recurse down to during the decomposition.
// Setting maxTier to a high value may results in a very large number of
// spans.
func (hc *Hilbert) DecomposeRegion(minTier, maxTier uint32,
	region Intersecter) ([]Cell, error) {

	if maxTier >= hc.order {
		return []Cell{}, fmt.Errorf("error decomposing region, maxTier (%v)"+
			" must be less than %v", maxTier, hc.order)
	}
	if minTier > maxTier {
		return []Cell{}, fmt.Errorf("error decomposing region, minTier (%v)"+
			" must be less than or equal to maxTier (%v)", minTier, maxTier)
	}

	cell := make(Point, hc.dim, hc.dim)
	it := hc.cellIterator(0, cell)

	dc := decomposeCall{
		bounds:  NewBox(cell, cell),
		minTier: minTier,
		maxTier: maxTier,
		region:  region,
	}

	result := []Cell{}

	for it() {
		err := hc.decomposeRegion(0, cell.Clone(), &dc, &result)
		if err != nil {
			return []Cell{}, err
		}
	}

	if len(result) == 0 {
		return []Cell{}, ErrNoOverlappingCells
	}

	return result, nil
}

func (hc *Hilbert) decomposeRegion(tier uint32, cell Point, dc *decomposeCall, result *[]Cell) error {

	tierBit := Bitmask(1) << (Bitmask(hc.order) - Bitmask(tier) - 1)
	upperBits := tierBit - 1

	// calculate the upper bound
	dc.bounds = NewBox(cell, cell)
	for d := uint32(0); d < hc.dim; d++ {
		dc.bounds[d].Max |= upperBits
	}

	intersects, err := dc.region.Intersects(&dc.bounds)
	if err != nil {
		return err
	}
	// if the region intersects the bounds of this tier/cell
	if intersects {
		// if we're in the reporting span
		if tier >= dc.minTier {

			contains, err := dc.region.Contains(&dc.bounds)
			if err != nil {
				return err
			}

			// if we've reached the max tier, or are fully contained
			if tier == dc.maxTier || contains {
				tmp := make([]Bitmask, hc.dim, hc.dim)
				for i := range cell {
					tmp[i] = cell[i] >> (hc.order - tier - 1)
				}

				value := Encode(Bitmask(tier+1), tmp)
				*result = append(*result, Cell{Value: value, Tier: tier})
			} else {
				// if we only partially overlap and we aren't at the max
				// tier

				it := hc.cellIterator(tier+1, cell)
				// go through all the child cells at this tier
				for it() {
					hc.decomposeRegion(tier+1, cell, dc, result)
				}
			}
			// if we aren't in the reporting span, just recurse
		} else {
			it := hc.cellIterator(tier+1, cell)
			// go through all the child cells at this tier
			for it() {
				hc.decomposeRegion(tier+1, cell, dc, result)
			}
		}
	}

	return nil
}
