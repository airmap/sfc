package sfc

// Point is a point in multi-dimensional space.
type Point []Bitmask

// Clone returns a deep copy of Point
func (pt Point) Clone() Point {
	ptCopy := make(Point, len(pt), len(pt))
	copy(ptCopy, pt)
	return ptCopy
}
