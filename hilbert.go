package sfc

import (
	"fmt"
)

// Bitmask is the datatype that contains integer values in both hilbert
// space and coordinate space.
type Bitmask uint64

// Hilbert defines the hilbert space.
type Hilbert struct {
	// dim is the number of dimensions, must be >= 1
	dim uint32
	// order is the number of bits per dimension, must be >= 1 and
	// <= 63
	order uint32
}

// NewHilbert returns a new Hilbert curve.
//
// dim - number of dimensions represented
//
// order - number of bits per dimension
//
// NOTE: dim * order must be <= 64
func NewHilbert(dim, order uint32) (*Hilbert, error) {
	if dim*order > 64 {
		return nil, fmt.Errorf("dim * order must be <= 64")
	}

	return &Hilbert{dim: dim, order: order}, nil
}

//
// nd1Ones - can be expensive, so pass it in each time.
// nd1Ones := ones(nDims) >> 1
func adjustRotation(rotation, nd1Ones, nDims, bits Bitmask) Bitmask {

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

// BBoxLowerValue returns the lower bound hilbert value for a given bounding
// box. The lower bound is placed into the original minBound/maxBound arrays.
//
// If minBound or maxBound are outside the bit range specified in order the
// results are undefined.
func BBoxLowerValue(order Bitmask, minBound, maxBound Point) (Bitmask, error) {

	if len(minBound) != len(maxBound) {
		return 0, fmt.Errorf("min and max bounds must be the same size")
	}
	nDim := Bitmask(len(minBound))

	if order*nDim > 64 {
		return 0, fmt.Errorf("dimension * order must be <= 64")
	}

	// reverse the coordinate so that coord[0] = X, coord[1] = Y, ...
	reverse(minBound)
	reverse(maxBound)

	hilbertBoxPt(order, true, minBound, maxBound)

	// reverse back before returning
	reverse(minBound)

	return Encode(order, minBound), nil
}

// BBoxUpperValue returns the upper bound hilbert value for a given bounding
// box. The upper bound is placed into the original minBound/maxBound arrays.
//
// If minBound or maxBound are outside the bit range specified in order the
// results are undefined.
func BBoxUpperValue(order Bitmask, minBound, maxBound Point) (Bitmask, error) {

	if len(minBound) != len(maxBound) {
		return 0, fmt.Errorf("min and max bounds must be the same size")
	}
	nDim := Bitmask(len(minBound))

	if order*nDim > 64 {
		return 0, fmt.Errorf("dimension * order must be <= 64")
	}

	// reverse the coordinate so that coord[0] = X, coord[1] = Y, ...
	reverse(minBound)
	reverse(maxBound)

	hilbertBoxPt(order, false, minBound, maxBound)

	// reverse back before returning
	reverse(maxBound)

	return Encode(order, maxBound), nil
}

func rdbit(w, k Bitmask) Bitmask {
	return (w >> k) & 1
}

// reverse copies and reversed the specified array.
func reverse(r []Bitmask) {

	for i := 0; i < len(r)/2; i++ {
		e := len(r) - 1 - i
		r[i], r[e] = r[e], r[i]
	}
}

// getBits reads nDims bits out of a Bitmask and consilidates them into
// a consecutive string of bits.
//
// Each dimension is represented with nBytes, this method will retrieve the
// yth bit out of each dimension and return that as a new Bitmask.
func getBits(bm []Bitmask, y Bitmask) Bitmask {
	bits := Bitmask(0)

	ld := Bitmask(len(bm))
	for d := Bitmask(0); d < ld; d++ {
		bits |= (bm[d] >> y & 1) << d
	}

	return bits
}

// func (bm Bitmask) writeBits(d, nBytes, y, fold Bitmask) {
// 	bit := y

// }

// propogateIntBits flips the bit in the yth position of c[d].
// If fold is 0 and the yth bit is flipped to 0, set all the bits below y to 1
// If fold is 0 and the yth bit is flipped to 1, set all the bits below y to 0
func propogateIntBits(d Bitmask, c []Bitmask, y Bitmask, fold Bitmask) {
	bthbit := Bitmask(1) << y
	c[d] ^= bthbit

	if fold == 0 {
		// notbit is 0 if bit y in c[d] is 1, otherwise all 1s
		notbit := ((c[d] >> y) & 1) - 1
		if notbit != 0 {
			// set all bits below y to 1
			c[d] |= bthbit - 1
		} else {
			// set all the bits below y 0
			c[d] &= -bthbit
		}
	}
}

// ones returns k bits with the value 1.
func ones(k Bitmask) Bitmask {
	return (Bitmask(1) << k) - 1
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
func Decode(nBits, index Bitmask, coord []Bitmask) {
	nDims := Bitmask(len(coord))

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
				rotation = adjustRotation(rotation, ndOnes>>1, nDims, bits)

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

// Dim returns the number of dimensions in the curve
func (hc *Hilbert) Dim() uint32 {
	return hc.dim
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
//
// Encoding values with a dimension > 10 is not supported. This is solely as an
// optimization to reduce memory churn.
func Encode(nBits Bitmask, coord []Bitmask) Bitmask {
	nDims := Bitmask(len(coord))

	// reverse the coordinate so that coord[0] = X, coord[1] = Y, ...
	reverse(coord)
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
				rotation = adjustRotation(rotation, ndOnes>>1, nDims, bits)

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

		reverse(coord)

		return index
	}

	return coord[0]
}

/*****************************************************************
 * hilbertBoxPt
 *
 * Determine the first or last point of a box to lie on a Hilbert curve
 * Inputs:
 *  nDims:      Number of coordinates.
 *  nBytes:     Number of bytes/coordinate.
 *  nBits:      Number of bits/coordinate.
 *  findMin:    Is it the least vertex sought?
 *  coord1:     Array of nDims nBytes-byte coordinates - one corner of box
 *  coord2:     Array of nDims nBytes-byte coordinates - opposite corner
 * Output:
 *      c1 and c2 modified to refer to least point
 * Assumptions:
 *      nBits <= (sizeof bitmask_t) * (bits_per_byte)
 */
func hilbertBoxPtWork(nBits, findMin,
	max, y Bitmask, c1, c2 []Bitmask,
	rotation, bits, index Bitmask) Bitmask {
	nDims := Bitmask(len(c1))

	one := Bitmask(1)
	var fold1, fold2, smearSum Bitmask

	for y > max {
		y--
		reflection := getBits(c1, y)
		diff := reflection ^ getBits(c2, y)

		if diff != 0 {
			smear := rotateRight(diff, rotation, nDims) >> 1
			digit := rotateRight(bits^reflection, rotation, nDims)

			for d := Bitmask(1); d < nDims; d *= 2 {
				index ^= index >> d
				digit ^= (digit >> d) & ^smear
				smear |= smear >> d
			}

			smearSum += smear
			index &= 1
			if (index ^ (y^findMin)&1) != 0 {
				digit ^= smear + 1
			}
			digit = rotateLeft(digit, rotation, nDims) & diff
			reflection ^= digit

			for d := Bitmask(0); d < nDims; d++ {
				if rdbit(diff, d) != 0 {
					// way := rdbit(digit, d);
					// char* c = way? c1: c2;
					// bitmask_t fold = way? fold1: fold2;
					// propogateBits(d, nBytes, c, y, rdbit(fold, d));

					if rdbit(digit, d) != 0 {
						propogateIntBits(d, c1, y, rdbit(fold1, d))
					} else {
						propogateIntBits(d, c2, y, rdbit(fold2, d))
					}
				}
			}

			diff ^= digit
			fold1 |= digit
			fold2 |= diff
		}

		bits ^= reflection
		bits = rotateRight(bits, rotation, nDims)
		index ^= bits
		reflection ^= one << rotation
		rotation = adjustRotation(rotation, ones(nDims)>>1, nDims, bits)
		bits = reflection
	}

	return smearSum
}

// hilbertBoxPt
// NOTE: If you visualize the 2D hilbert curve with the lower left as 0,0 then
// c1 and c2 take the form {Y, X}.
//
func hilbertBoxPt(nBits Bitmask, findMin bool,
	c1, c2 []Bitmask) Bitmask {

	nDims := Bitmask(len(c1))
	one := Bitmask(1)
	bits := one << (nDims - 1)
	var fm Bitmask
	// yeah, these appear to be reversed when nBits < 8. Dunno why.
	if findMin && nBits < 8 || findMin == false && nBits >= 8 {
		fm = 0
	} else {
		fm = 1
	}
	return hilbertBoxPtWork(nBits, fm, 0, nBits, c1, c2, 0, bits, bits)
}

// Order returns the number of bits per dimension in the curve.
func (hc *Hilbert) Order() uint32 {
	return hc.order
}
