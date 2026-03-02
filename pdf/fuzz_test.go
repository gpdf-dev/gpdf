package pdf

import (
	"bytes"
	"testing"
)

// FuzzNameWriteTo tests that arbitrary Name strings do not panic
// during PDF serialization.
func FuzzNameWriteTo(f *testing.F) {
	f.Add("Type")
	f.Add("Font")
	f.Add("")
	f.Add("hello world")
	f.Add("日本語")
	f.Add("a#b/c(d)e<f>g[h]i{j}k%l")

	f.Fuzz(func(t *testing.T, s string) {
		var buf bytes.Buffer
		n := Name(s)
		_, _ = n.WriteTo(&buf)
	})
}

// FuzzLiteralStringWriteTo tests that arbitrary literal strings do not panic.
func FuzzLiteralStringWriteTo(f *testing.F) {
	f.Add("Hello World")
	f.Add("")
	f.Add("(nested)")
	f.Add(`back\slash`)
	f.Add("line\nbreak")
	f.Add("\x00\x01\x02\xff")

	f.Fuzz(func(t *testing.T, s string) {
		var buf bytes.Buffer
		ls := LiteralString(s)
		_, _ = ls.WriteTo(&buf)
	})
}

// FuzzHexStringWriteTo tests that arbitrary hex strings do not panic.
func FuzzHexStringWriteTo(f *testing.F) {
	f.Add("Hello")
	f.Add("")
	f.Add("\x00\xff")

	f.Fuzz(func(t *testing.T, s string) {
		var buf bytes.Buffer
		hs := HexString(s)
		_, _ = hs.WriteTo(&buf)
	})
}

// FuzzStreamWriteTo tests that arbitrary stream content does not panic.
func FuzzStreamWriteTo(f *testing.F) {
	f.Add([]byte("stream content"))
	f.Add([]byte{})
	f.Add([]byte{0x00, 0xFF, 0x01, 0xFE})

	f.Fuzz(func(t *testing.T, data []byte) {
		var buf bytes.Buffer
		s := Stream{
			Dict:    Dict{Name("Type"): Name("XObject")},
			Content: data,
		}
		_, _ = s.WriteTo(&buf)
	})
}
