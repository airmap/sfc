package sfc

// Intersecter provides method for determining the relationship between a
// region (this) and bounds.
//
// This must be thread safe.
type Intersecter interface {
	// Contains returns true if bounds is fully contained by the region.
	Contains(bounds *Box) (bool, error)

	// Intersects returns true if a region overlaps with bounds. This will
	// return false if bounds is adjacent to or outside of the region.
	Intersects(bounds *Box) (bool, error)
}
