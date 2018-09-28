package sfc

import (
	"fmt"
)

// Bounds represents a bounding box. Points are inclusive.
type Bounds struct {
	Min Point
	Max Point
}

// Dimensions returns the number of dimensions in the bounding box.
func (b *Bounds) Dimensions() uint32 {
	return uint32(len(b.Min))
}

// Contains returns true if b completely overlaps all points in other.
func (b *Bounds) Contains(other *Bounds) (bool, error) {

	if b.Dimensions() != other.Dimensions() {
		return false, fmt.Errorf("dimensions do not match")
	}

	for d := uint32(0); d < b.Dimensions(); d++ {
		if other.Max[d] < b.Min[d] ||
			other.Min[d] > b.Max[d] ||
			other.Min[d] < b.Min[d] ||
			other.Max[d] > b.Max[d] {
			return false, nil
		}
	}

	return true, nil
}

// Intersects returns true if b touchs other at any point.
func (b *Bounds) Intersects(other *Bounds) (bool, error) {

	if b.Dimensions() != other.Dimensions() {
		return false, fmt.Errorf("dimensions do not match")
	}

	for d := uint32(0); d < b.Dimensions(); d++ {
		if b.Max[d] < other.Min[d] || other.Max[d] < b.Min[d] {
			return false, nil
		}
	}

	return true, nil
}
