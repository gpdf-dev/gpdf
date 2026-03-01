package qrcode

// rsGenPoly builds a Reed-Solomon generator polynomial of the given degree.
// The polynomial is represented as coefficients [g0, g1, ..., gn] where
// g(x) = (x - alpha^0)(x - alpha^1)...(x - alpha^(n-1)).
func rsGenPoly(degree int) []byte {
	gen := make([]byte, degree+1)
	gen[0] = 1

	for i := 0; i < degree; i++ {
		// Multiply gen by (x - alpha^i).
		for j := degree; j > 0; j-- {
			gen[j] = gen[j-1] ^ gfMul(gen[j], gfExp[i])
		}
		gen[0] = gfMul(gen[0], gfExp[i])
	}
	return gen
}

// rsEncode computes Reed-Solomon error correction codewords for the given data.
// ecCount is the number of EC codewords to generate.
func rsEncode(data []byte, ecCount int) []byte {
	gen := rsGenPoly(ecCount)

	// Polynomial long division.
	result := make([]byte, ecCount)
	for _, b := range data {
		coeff := b ^ result[0]
		// Shift result left by 1.
		copy(result, result[1:])
		result[ecCount-1] = 0
		// Subtract gen * coeff.
		if coeff != 0 {
			for j := 0; j < ecCount; j++ {
				result[j] ^= gfMul(gen[ecCount-1-j], coeff)
			}
		}
	}
	return result
}
