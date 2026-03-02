package qrcode

import (
	"testing"
)

// FuzzEncode tests that arbitrary data strings do not cause panics
// during QR code encoding.
func FuzzEncode(f *testing.F) {
	f.Add("Hello, World!")
	f.Add("https://example.com")
	f.Add("12345678")
	f.Add("ABCDEFGHIJKLMNOP")
	f.Add("日本語テスト")
	f.Add("a")

	f.Fuzz(func(t *testing.T, data string) {
		if len(data) == 0 {
			return // empty is rejected by contract
		}
		qr, err := Encode(data, LevelM)
		if err != nil {
			return // data too long or unsupported is expected
		}
		// If encoding succeeded, PNG generation should not panic.
		_, _ = qr.PNG(1)
	})
}

// FuzzEncodeAllLevels tests encoding at all error correction levels.
func FuzzEncodeAllLevels(f *testing.F) {
	f.Add("test", byte(0))
	f.Add("test", byte(1))
	f.Add("test", byte(2))
	f.Add("test", byte(3))

	f.Fuzz(func(t *testing.T, data string, levelByte byte) {
		if len(data) == 0 {
			return
		}
		level := ErrorCorrectionLevel(levelByte % 4)
		qr, err := Encode(data, level)
		if err != nil {
			return
		}
		_, _ = qr.PNG(1)
	})
}
