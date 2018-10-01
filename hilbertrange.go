package sfc

// Intersecter provides method for determining the relationship between a
// region (this) and bounds.
//
// This must be thread safe.
type Intersecter interface {
	// Contains returns true if bounds is fully contained by the region.
	Contains(bounds *Bounds) (bool, error)

	// Intersects returns true if a region overlaps with bounds. This will
	// return false if bounds is adjacent to or outside of the region.
	Intersects(bounds *Bounds) (bool, error)
}

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
	bounds  Bounds
	region  Intersecter
}

// DecomposeRanges breaks a region up into a series of hilbert value ranges.
//
// minTier - The minimum tier in the hilbert curve to start the decomposition.
// Setting this too high may result in a large number of ranges.
//
// maxTier - The maximum tier to recurse down to during the decomposition.
// Setting maxTier to a high value may results in a very large number of
// ranges.
func (hc *Hilbert) DecomposeRanges(minTier, maxTier uint32,
	region Intersecter) (Ranges, error) {

	cell := make(Point, hc.dim, hc.dim)
	it := hc.cellIterator(0, cell)

	dc := decomposeCall{
		bounds:  Bounds{Min: cell.Clone(), Max: cell.Clone()},
		minTier: minTier,
		maxTier: maxTier,
		region:  region,
	}

	result := Ranges{}

	for it() {
		err := hc.decomposeRanges(0, cell.Clone(), &dc, &result)
		if err != nil {
			return Ranges{}, err
		}
	}

	result = joinRanges(result)

	return result, nil
}

func (hc *Hilbert) decomposeRanges(tier uint32, cell Point, dc *decomposeCall, result *Ranges) error {

	tierBit := Bitmask(1) << (Bitmask(hc.order) - Bitmask(tier) - 1)
	upperBits := tierBit - 1

	// calculate the upper bound
	copy(dc.bounds.Min, cell)
	copy(dc.bounds.Max, cell)
	for d := uint32(0); d < hc.dim; d++ {
		dc.bounds.Max[d] |= upperBits
	}

	intersects, err := dc.region.Intersects(&dc.bounds)
	if err != nil {
		return err
	}
	// if the region intersects the bounds of this tier/cell
	if intersects {

		// if we're in the reporting range
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

				r := Range{
					MinValue: value & ^tierValueBits,
					MaxValue: value | tierValueBits,
				}
				*result = append(*result, r)
			} else {
				// if we only partially overlap and we aren't at the max
				// tier

				it := hc.cellIterator(tier+1, cell)
				// go through all the child cells at this tier
				for it() {
					hc.decomposeRanges(tier+1, cell, dc, result)
				}
			}
			// if we aren't in the reporting range, just recurse
		} else {
			it := hc.cellIterator(tier+1, cell)
			// go through all the child cells at this tier
			for it() {
				hc.decomposeRanges(tier+1, cell, dc, result)
			}
		}
	}

	return nil
}

// DecomposeRegion breaks a region up into a series of hilbert value cells.
//
// minTier - The minimum tier in the hilbert curve to start the decomposition.
// Setting this too high may result in a large number of ranges.
//
// maxTier - The maximum tier to recurse down to during the decomposition.
// Setting maxTier to a high value may results in a very large number of
// ranges.
func (hc *Hilbert) DecomposeRegion(minTier, maxTier uint32,
	region Intersecter) ([]Cell, error) {

	cell := make(Point, hc.dim, hc.dim)
	it := hc.cellIterator(0, cell)

	dc := decomposeCall{
		bounds:  Bounds{Min: cell.Clone(), Max: cell.Clone()},
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

	return result, nil
}

func (hc *Hilbert) decomposeRegion(tier uint32, cell Point, dc *decomposeCall, result *[]Cell) error {

	tierBit := Bitmask(1) << (Bitmask(hc.order) - Bitmask(tier) - 1)
	upperBits := tierBit - 1

	// calculate the upper bound
	copy(dc.bounds.Min, cell)
	copy(dc.bounds.Max, cell)
	for d := uint32(0); d < hc.dim; d++ {
		dc.bounds.Max[d] |= upperBits
	}

	intersects, err := dc.region.Intersects(&dc.bounds)
	if err != nil {
		return err
	}
	// if the region intersects the bounds of this tier/cell
	if intersects {

		// if we're in the reporting range
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
			// if we aren't in the reporting range, just recurse
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
