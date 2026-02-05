// Package readsolomon implements Reed-Solomon error correction for MaxiCode.
package readsolomon

// Encoder holds precomputed tables for Reed-Solomon encoding.
type Encoder struct {
	logSize      int
	eccSymbols   int
	logTable     []int
	alogTable    []int
	rsPolynomial []int
}

// NewEncoder creates a new Reed-Solomon encoder with the given polynomial, ECC symbol count, and starting index.
func NewEncoder(polynomial, eccSymbols, index int) *Encoder {
	size := 0

	// Find the top bit, and hence the symbol size
	b := 1
	for ; b <= polynomial; b <<= 1 {
		size++
	}

	size--
	b >>= 1

	// Build the log and antilog tables.
	logSize := (1 << uint(size)) - 1 //nolint:gosec // size is bounded by polynomial bit width
	logTable := make([]int, logSize+1)
	alogTable := make([]int, logSize)
	rsPolynomial := make([]int, eccSymbols+1)

	for p, v := 1, 0; v < logSize; v++ {
		alogTable[v] = p
		logTable[p] = v
		p <<= 1
		if (p & b) != 0 {
			p ^= polynomial
		}
	}

	rsPolynomial[0] = 1
	for i := 1; i <= eccSymbols; i++ {
		rsPolynomial[i] = 1
		for k := i - 1; k > 0; k-- {
			if rsPolynomial[k] != 0 {
				rsPolynomial[k] = alogTable[(logTable[rsPolynomial[k]]+index)%logSize]
			}

			rsPolynomial[k] ^= rsPolynomial[k-1]
		}

		rsPolynomial[0] = alogTable[(logTable[rsPolynomial[0]]+index)%logSize]
		index++
	}

	return &Encoder{
		logSize:      logSize,
		eccSymbols:   eccSymbols,
		logTable:     logTable,
		alogTable:    alogTable,
		rsPolynomial: rsPolynomial,
	}
}

// Encode performs Reed-Solomon encoding on the given data, writing error correction bytes to ecc.
func (e *Encoder) Encode(length int, data, ecc []byte) {
	for i := 0; i < length; i++ {
		m := ecc[e.eccSymbols-1] ^ data[i]

		for j := e.eccSymbols - 1; j > 0; j-- {
			if m != 0 && e.rsPolynomial[j] != 0 {
				ecc[j] = ecc[j-1] ^ byte(e.alogTable[(e.logTable[m]+e.logTable[e.rsPolynomial[j]])%e.logSize])
			} else {
				ecc[j] = ecc[j-1]
			}
		}

		if m != 0 && e.rsPolynomial[0] != 0 {
			ecc[0] = byte(e.alogTable[(e.logTable[m]+e.logTable[e.rsPolynomial[0]])%e.logSize])
		} else {
			ecc[0] = 0
		}
	}
}
