package qrcode

// GF(2^8) finite field arithmetic with primitive polynomial 0x11d (x^8 + x^4 + x^3 + x^2 + 1).

var gfExp [512]byte // antilog table (doubled for convenience)
var gfLog [256]byte // log table

func init() {
	// Build exponential and logarithm tables.
	x := 1
	for i := 0; i < 255; i++ {
		gfExp[i] = byte(x)
		gfLog[x] = byte(i)
		x <<= 1
		if x >= 256 {
			x ^= 0x11d
		}
	}
	// Duplicate for easy modular access.
	for i := 255; i < 512; i++ {
		gfExp[i] = gfExp[i-255]
	}
}

// gfMul returns a * b in GF(2^8).
func gfMul(a, b byte) byte {
	if a == 0 || b == 0 {
		return 0
	}
	return gfExp[int(gfLog[a])+int(gfLog[b])]
}

// gfDiv returns a / b in GF(2^8). Panics if b == 0.
func gfDiv(a, b byte) byte {
	if b == 0 {
		panic("gf256: division by zero")
	}
	if a == 0 {
		return 0
	}
	return gfExp[(int(gfLog[a])-int(gfLog[b]))+255]
}
