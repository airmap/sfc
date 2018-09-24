package sfc

// Bitmask is the datatype that contains integer values in both hilbert
// space and coordinate space.
type Bitmask uint64
type halfmaskT uint32

// Hilbert defines the hilbert space.
type Hilbert struct {
	// dim is the number of dimensions, must be >= 1
	dim uint64
	// order is the number of bits per dimension, must be >= 1 and
	// <= 63
	order uint64
}

func adjustRotation(rotation, nDims, bits Bitmask) Bitmask {
	nd1Ones := ones(nDims) >> 1

	/* rotation = (rotation + 1 + ffs(bits)) % nDims; */
	bits &= -bits & nd1Ones
	for bits != 0 {
		bits >>= 1
		rotation++
	}
	rotation++
	if rotation >= nDims {
		rotation -= nDims
	}

	return rotation
}

// reverse copies and reversed the specified array.
func reverse(b []Bitmask) []Bitmask {
	r := make([]Bitmask, len(b), len(b))

	for i := range b {
		r[len(b)-1-i] = b[i]
	}

	return r
}

// ones returns k bits with the value 1.
func ones(k Bitmask) Bitmask {
	return (Bitmask(2) << (k - 1)) - 1
}

func rotateLeft(arg, nRots, nDims Bitmask) Bitmask {
	return (((arg) << (nRots)) | ((arg) >> ((nDims) - (nRots)))) & ones(nDims)
}

func rotateRight(arg, nRots, nDims Bitmask) Bitmask {
	return (((arg) >> (nRots)) | ((arg) << ((nDims) - (nRots)))) &
		ones(nDims)
}

func bitTranspose(nDims, nBits, inCoords Bitmask) Bitmask {
	nDims1 := nDims - 1
	inB := nBits
	inFieldEnds := Bitmask(1)
	inMask := ones(inB)
	coords := Bitmask(0)

	for utB := inB / 2; utB != 0; utB = inB / 2 {
		shiftAmt := nDims1 * utB
		utFieldEnds := inFieldEnds | (inFieldEnds << (shiftAmt + utB))
		utMask := (utFieldEnds << utB) - utFieldEnds
		utCoords := Bitmask(0)

		if inB&1 != 0 {
			inFieldStarts := inFieldEnds << (inB - 1)
			oddShift := 2 * shiftAmt

			for d := Bitmask(0); d < nDims; d++ {
				in := inCoords & inMask
				inCoords >>= inB
				coords |= (in & inFieldStarts) << oddShift
				oddShift++
				in &= ^inFieldStarts
				in = (in | (in << shiftAmt)) & utMask
				utCoords |= in << (d * utB)
			}
		} else {
			for d := Bitmask(0); d < nDims; d++ {
				in := inCoords & inMask
				inCoords >>= inB
				in = (in | (in << shiftAmt)) & utMask
				utCoords |= in << (d * utB)
			}
		}

		inCoords = utCoords
		inB = utB
		inFieldEnds = utFieldEnds
		inMask = utMask
	}

	coords |= inCoords
	return coords
}

// Decode converts an index into a Hilbert curve to a set of coordinates.
// Inputs:
//  nDims:      Number of coordinate axes.
//  nBits:      Number of bits per axis.
//  index:      The index, contains nDims*nBits bits
//              (so nDims*nBits must be <= 8*sizeof(Bitmask)).
// Outputs:
//  coord:      The list of nDims coordinates, each with nBits bits. coord
//              must be the same length as nDims or the method panics.
// Assumptions:
//      nDims*nBits <= (sizeof index)// (bits_per_byte)
//
//
func Decode(nDims, nBits, index Bitmask, coord []Bitmask) {
	if len(coord) != int(nDims) {
		panic("coord must have a length equal to nDims")
	}

	if nDims > 1 {
		coords := Bitmask(0)
		nbOnes := ones(nBits)

		if nBits > 1 {
			nDimsBits := nDims * nBits
			ndOnes := ones(nDims)
			b := nDimsBits
			rotation := Bitmask(0)
			flipBit := Bitmask(0)
			nthbits := ones(nDimsBits) / ndOnes
			index ^= (index ^ nthbits) >> 1

			for {
				b -= nDims
				bits := (index >> b) & ndOnes
				coords <<= nDims
				coords |= rotateLeft(bits, rotation, nDims) ^ flipBit
				flipBit = 1 << rotation
				rotation = adjustRotation(rotation, nDims, bits)

				if b == 0 {
					break
				}
			}

			for b = nDims; b < nDimsBits; b *= 2 {
				coords ^= coords >> b
			}
			coords = bitTranspose(nBits, nDims, coords)
		} else {
			coords = index ^ (index >> 1)
		}

		for d := Bitmask(0); d < nDims; d++ {
			coord[nDims-d-1] = coords & nbOnes
			coords >>= nBits
		}
	} else {
		coord[0] = index
	}
}

// Encode converts coordinates of a point on a Hilbert curve to its index.
// Inputs:
//  nDims:      Number of coordinates.
//  nBits:      Number of bits/coordinate.
//  coord:      Array of n nBits-bit coordinates.
// Outputs:
//  index:      Output index value.  nDims*nBits bits.
// Assumptions:
//      nDims*nBits <= (sizeof Bitmask) * (bits_per_byte)
func Encode(nDims, nBits Bitmask, coord []Bitmask) Bitmask {
	// reverse the coordinate so that coord[0] = X, coord[1] = Y, ...
	coord = reverse(coord)
	if nDims > 1 {
		nDimsBits := nDims * nBits
		coords := Bitmask(0)
		index := Bitmask(0)

		for d := int(nDims - 1); d >= 0; d-- {
			coords <<= nBits
			coords |= coord[d]
		}

		if nBits > 1 {
			ndOnes := ones(nDims)
			b := nDimsBits
			rotation := Bitmask(0)
			flipBit := Bitmask(0)
			nthbits := ones(nDimsBits) / ndOnes
			coords := bitTranspose(nDims, nBits, coords)
			coords ^= coords >> nDims
			index = Bitmask(0)

			for {
				b -= nDims
				bits := (coords >> b) & ndOnes
				bits = rotateRight(flipBit^bits, rotation, nDims)
				index <<= nDims
				index |= bits
				flipBit = Bitmask(1) << rotation
				rotation = adjustRotation(rotation, nDims, bits)

				if b == 0 {
					break
				}
			}
			index = index ^ (nthbits >> 1)
		} else {
			index = coords
		}

		for d := Bitmask(1); d < nDimsBits; d *= 2 {
			index = index ^ (index >> d)
		}

		return index
	}

	return coord[0]
}
