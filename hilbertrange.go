package sfc

import (
	"sync"
)

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
	ranges  chan Range
	region  Intersecter
	err     chan error
	wg      *sync.WaitGroup
}

// DecomposeRegion breaks a region up into a series of hilbert value ranges.
//
// minTier - The minimum tier in the hilbert curve to start the decomposition.
// Setting this too high may result in a large number of ranges.
//
// maxTier - The maximum tier to recurse down to during the decomposition.
// Setting maxTier to a high value may results in a very large number of
// ranges.
func (hc *Hilbert) DecomposeRegion(minTier, maxTier uint32,
	region Intersecter) (Ranges, error) {

	cell := make(Point, hc.dim, hc.dim)
	it := hc.cellIterator(0, cell)

	dc := decomposeCall{
		err:     make(chan error),
		minTier: minTier,
		maxTier: maxTier,
		ranges:  make(chan Range),
		region:  region,
		wg:      &sync.WaitGroup{},
	}

	result := Ranges{}

	for it() {
		err := hc.decomposeRegion(0, cell.Clone(), &dc, &result)
		if err != nil {
			return Ranges{}, err
		}
	}

	result = joinRanges(result)

	return result, nil
}

func (hc *Hilbert) decomposeRegion(tier uint32, cell Point, dc *decomposeCall, result *Ranges) error {

	tierBit := Bitmask(1) << (Bitmask(hc.order) - Bitmask(tier) - 1)
	upperBits := tierBit - 1
	upper := make(Point, hc.dim, hc.dim)

	// calculate the upper bound
	copy(upper, cell)
	for d := uint32(0); d < hc.dim; d++ {
		upper[d] |= upperBits
	}

	bounds := &Bounds{Min: cell, Max: upper}

	intersects, err := dc.region.Intersects(bounds)
	if err != nil {
		return err
	}
	// if the region intersects the bounds of this tier/cell
	if intersects {

		// if we're in the reporting range
		if tier >= dc.minTier {

			contains, err := dc.region.Contains(bounds)
			if err != nil {
				return err
			}

			// if we've reached the max tier, or are fully contained
			if tier == dc.maxTier || contains {
				tmp1 := cell.Clone()
				tmp2 := upper.Clone()

				lower, err := BBoxLowerValue(Bitmask(hc.order), tmp1, tmp2)
				if err != nil {
					return err
				}

				copy(tmp1, cell)
				copy(tmp2, upper)
				upper, err := BBoxUpperValue(Bitmask(hc.order), tmp1, tmp2)
				if err != nil {
					return err
				}

				r := Range{
					MinValue: lower,
					MaxValue: upper,
				}
				*result = append(*result, r)
				// if we only partially overlap and we aren't at the max
				// tier
			} else {
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

// DecomposeRegionThreads breaks a region up into a series of hilbert value
// ranges.
//
// This is similar to DecomposeRegion, but multiple threads are used. In simple
// cases this is a little slower and uses far more CPUs, but if contains and
// intersects are expesive operations this may provide some performance
// improvement.
//
// minTier - The minimum tier in the hilbert curve to start the decomposition.
// Setting this too high may result in a large number of ranges.
//
// maxTier - The maximum tier to recurse down to during the decomposition.
// Setting maxTier to a high value may results in a very large number of
// ranges.
func (hc *Hilbert) DecomposeRegionThreads(minTier, maxTier uint32,
	region Intersecter) (Ranges, error) {

	cell := make(Point, hc.dim, hc.dim)
	it := hc.cellIterator(0, cell)

	dc := decomposeCall{
		err:     make(chan error),
		minTier: minTier,
		maxTier: maxTier,
		ranges:  make(chan Range),
		region:  region,
		wg:      &sync.WaitGroup{},
	}

	go func() {
		for it() {
			dc.wg.Add(1)
			go hc.decomposeRegionThreads(0, cell.Clone(), &dc)
		}

		dc.wg.Wait()

		close(dc.ranges)
		close(dc.err)
	}()

	result := Ranges{}

	for {
		select {
		case r, ok := <-dc.ranges:
			if !ok {
				dc.ranges = nil
			} else {
				result = append(result, r)
			}
		case err, ok := <-dc.err:
			if !ok {
				dc.err = nil
			} else {
				return Ranges{}, err
			}
		}

		if dc.ranges == nil && dc.err == nil {
			break
		}
	}

	result = joinRanges(result)

	return result, nil
}

func (hc *Hilbert) decomposeRegionThreads(tier uint32, cell Point, dc *decomposeCall) {

	defer dc.wg.Done()

	tierBit := Bitmask(1) << (Bitmask(hc.order) - Bitmask(tier) - 1)
	upperBits := tierBit - 1
	upper := make(Point, hc.dim, hc.dim)

	// calculate the upper bound
	copy(upper, cell)
	for d := uint32(0); d < hc.dim; d++ {
		upper[d] |= upperBits
	}

	bounds := &Bounds{Min: cell, Max: upper}

	intersects, err := dc.region.Intersects(bounds)
	if err != nil {
		dc.err <- err
		return
	}
	// if the region intersects the bounds of this tier/cell
	if intersects {

		// if we're in the reporting range
		if tier >= dc.minTier {

			contains, err := dc.region.Contains(bounds)
			if err != nil {
				dc.err <- err
				return
			}

			// if we've reached the max tier, or are fully contained
			if tier == dc.maxTier || contains {

				lower, err := BBoxLowerValue(Bitmask(hc.order), cell.Clone(), upper.Clone())
				if err != nil {
					dc.err <- err
					return
				}

				upper, err := BBoxUpperValue(Bitmask(hc.order), cell.Clone(), upper.Clone())
				if err != nil {
					dc.err <- err
					return
				}

				r := Range{
					MinValue: lower,
					MaxValue: upper,
				}
				dc.ranges <- r
				// if we only partially overlap and we aren't at the max
				// tier
			} else {
				c := cell.Clone()
				it := hc.cellIterator(tier+1, c)
				// go through all the child cells at this tier
				for it() {
					dc.wg.Add(1)
					go hc.decomposeRegionThreads(tier+1, c.Clone(), dc)
				}
			}
			// if we aren't in the reporting range, just recurse
		} else {
			c := cell.Clone()
			it := hc.cellIterator(tier+1, c)
			// go through all the child cells at this tier
			for it() {
				dc.wg.Add(1)
				go hc.decomposeRegionThreads(tier+1, c.Clone(), dc)
			}
		}
	}

	return
}
