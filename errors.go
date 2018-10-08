package sfc

import (
	"errors"
)

// ErrNoOverlappingCells DecomposeRegion didn't find any appropriately
// overlapping cells with the specified region.
var ErrNoOverlappingCells = errors.New("no cells overlap region")
