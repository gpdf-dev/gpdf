package barcode

import (
	"testing"
)

// FuzzEncode tests that arbitrary data strings do not cause panics
// during barcode encoding.
func FuzzEncode(f *testing.F) {
	f.Add("Hello123")
	f.Add("ABCDEFGHIJ")
	f.Add("0123456789")
	f.Add("!@#$%^&*()")
	f.Add("a")
	f.Add(" ")

	f.Fuzz(func(t *testing.T, data string) {
		if len(data) == 0 {
			return
		}
		bc, err := Encode(data, Code128)
		if err != nil {
			return // characters outside Code 128 range are expected errors
		}
		// If encoding succeeded, PNG generation should not panic.
		_, _ = bc.PNG(1, 50)
	})
}
