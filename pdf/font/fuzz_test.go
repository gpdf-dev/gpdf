package font

import (
	"testing"
)

// FuzzParseTrueType tests that arbitrary byte sequences do not cause
// panics when parsed as TrueType font data.
func FuzzParseTrueType(f *testing.F) {
	// Minimal valid-ish TrueType header (will fail validation but
	// exercises the parser).
	header := make([]byte, 12)
	header[0] = 0x00
	header[1] = 0x01
	header[2] = 0x00
	header[3] = 0x00
	f.Add(header)

	// Empty input.
	f.Add([]byte{})

	// Random short inputs.
	f.Add([]byte{0xFF, 0xFF, 0xFF, 0xFF})
	f.Add([]byte("true"))

	f.Fuzz(func(t *testing.T, data []byte) {
		// ParseTrueType should return an error for invalid data,
		// never panic.
		_, _ = ParseTrueType(data)
	})
}
