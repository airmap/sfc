package sfc

import (
	"fmt"
)

// Box represents a multi-dimensional box. Points are inclusive.
//
// While it would be nicer to give all Box's receivers non-pointers, the
// pointers must be there to implement Intersecter
type Box []Span

// NewBox constructs a new box.
//
// min the point that presents the min values in the box.
// max the point that presents the max values in the box.
func NewBox(min, max Point) Box {
	if len(min) != len(max) {
		panic("min and max have different dimensions")
	}

	var result Box = make([]Span, len(min))
	result.SetMin(min)
	result.SetMax(max)

	return result
}

// Dimensions returns the number of dimensions in the bounding box.
func (b *Box) Dimensions() uint32 {
	return uint32(len(*b))
}

// Clone returns a deep copy of b.
func (b *Box) Clone() *Box {
	result := make(Box, len(*b))
	for i := range *b {
		result[i].Min = (*b)[i].Min
		result[i].Max = (*b)[i].Max
	}
	return &result
}

// Contains returns true if b completely overlaps all points in other.
func (b *Box) Contains(other *Box) (bool, error) {

	if b.Dimensions() != other.Dimensions() {
		return false, fmt.Errorf("dimensions do not match")
	}

	for d := uint32(0); d < b.Dimensions(); d++ {
		if (*other)[d].Max < (*b)[d].Min ||
			(*other)[d].Min > (*b)[d].Max ||
			(*other)[d].Min < (*b)[d].Min ||
			(*other)[d].Max > (*b)[d].Max {
			return false, nil
		}
	}

	return true, nil
}

// Intersects returns true if b touchs other at any point.
func (b *Box) Intersects(other *Box) (bool, error) {

	if b.Dimensions() != other.Dimensions() {
		return false, fmt.Errorf("dimensions do not match")
	}

	for d := uint32(0); d < b.Dimensions(); d++ {
		if (*b)[d].Max < (*other)[d].Min || (*other)[d].Max < (*b)[d].Min {
			return false, nil
		}
	}

	return true, nil
}

// SetMax sets the max values on all dimension to the associated values in p.
//
// If p and b have a different number of dimensions the function panics.
func (b *Box) SetMax(p Point) {
	if len(p) != len(*b) {
		panic("point is not the same dimension as the box")
	}
	for d := 0; d < len(*b); d++ {
		(*b)[d].Max = p[d]
	}
}

// SetMin sets the min values on all dimension to the associated values in p.
//
// If p and b have a different number of dimensions the function panics.
func (b *Box) SetMin(p Point) {
	if len(p) != len(*b) {
		panic("point is not the same dimension as the box")
	}
	for d := 0; d < len(*b); d++ {
		(*b)[d].Min = p[d]
	}
}
