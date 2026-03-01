package qrcode

import (
	"encoding/binary"
	"testing"
)

func TestGF256Multiply(t *testing.T) {
	// a * 1 = a
	for i := 0; i < 256; i++ {
		if got := gfMul(byte(i), 1); got != byte(i) {
			t.Errorf("gfMul(%d, 1) = %d, want %d", i, got, i)
		}
	}
	// a * 0 = 0
	for i := 0; i < 256; i++ {
		if got := gfMul(byte(i), 0); got != 0 {
			t.Errorf("gfMul(%d, 0) = %d, want 0", i, got)
		}
	}
	// Known value: 0x02 * 0x02 = 0x04
	if got := gfMul(2, 2); got != 4 {
		t.Errorf("gfMul(2,2) = %d, want 4", got)
	}
}

func TestGF256DivInverse(t *testing.T) {
	// a / a = 1 for all nonzero a.
	for i := 1; i < 256; i++ {
		if got := gfDiv(byte(i), byte(i)); got != 1 {
			t.Errorf("gfDiv(%d, %d) = %d, want 1", i, i, got)
		}
	}
}

func TestReedSolomon(t *testing.T) {
	// Version 1-M data codewords with known correct EC output.
	data := []byte{32, 91, 11, 120, 209, 114, 220, 77, 67, 64, 236, 17, 236, 17, 236, 17}
	ec := rsEncode(data, 10)
	want := []byte{196, 35, 39, 119, 235, 215, 231, 226, 93, 23}
	if len(ec) != len(want) {
		t.Fatalf("rsEncode returned %d EC codewords, want %d", len(ec), len(want))
	}
	for i := range ec {
		if ec[i] != want[i] {
			t.Errorf("rsEncode[%d] = %d, want %d", i, ec[i], want[i])
		}
	}
}

func TestBitBuffer(t *testing.T) {
	buf := &bitBuffer{}
	buf.put(0b1010, 4)
	buf.put(0b1100, 4)
	got := buf.bytes()
	if len(got) != 1 || got[0] != 0xAC {
		t.Errorf("bitBuffer got %x, want AC", got)
	}
	if buf.len != 8 {
		t.Errorf("bitBuffer len = %d, want 8", buf.len)
	}
}

func TestDetectMode(t *testing.T) {
	tests := []struct {
		input string
		want  mode
	}{
		{"12345", modeNumeric},
		{"0", modeNumeric},
		{"HELLO WORLD", modeAlphanumeric},
		{"A1", modeAlphanumeric},
		{"hello", modeByte},
		{"https://gpdf.dev", modeByte},
		{"こんにちは", modeByte},
	}
	for _, tt := range tests {
		if got := detectMode([]byte(tt.input)); got != tt.want {
			t.Errorf("detectMode(%q) = %d, want %d", tt.input, got, tt.want)
		}
	}
}

func TestSelectVersion(t *testing.T) {
	// Short numeric data should fit in version 1.
	v, err := selectVersion([]byte("12345"), modeNumeric, LevelM)
	if err != nil {
		t.Fatal(err)
	}
	if v != 1 {
		t.Errorf("selectVersion for '12345' = %d, want 1", v)
	}

	// URL should pick a reasonable version.
	v, err = selectVersion([]byte("https://gpdf.dev"), modeByte, LevelM)
	if err != nil {
		t.Fatal(err)
	}
	if v < 1 || v > 40 {
		t.Errorf("selectVersion returned invalid version %d", v)
	}
}

func TestEncodeNumeric(t *testing.T) {
	qr, err := Encode("01234567", LevelM)
	if err != nil {
		t.Fatal(err)
	}
	if qr.Version != 1 {
		t.Errorf("version = %d, want 1", qr.Version)
	}
	expectedSize := moduleSize(1) // 21
	if qr.Size() != expectedSize {
		t.Errorf("size = %d, want %d", qr.Size(), expectedSize)
	}
}

func TestEncodeAlphanumeric(t *testing.T) {
	qr, err := Encode("HELLO WORLD", LevelQ)
	if err != nil {
		t.Fatal(err)
	}
	if qr.Version != 1 {
		t.Errorf("version = %d, want 1", qr.Version)
	}
	if qr.Size() != 21 {
		t.Errorf("size = %d, want 21", qr.Size())
	}
}

func TestEncodeByte(t *testing.T) {
	qr, err := Encode("https://gpdf.dev", LevelM)
	if err != nil {
		t.Fatal(err)
	}
	if qr.Version < 1 {
		t.Errorf("unexpected version %d", qr.Version)
	}
	if qr.Size() != moduleSize(qr.Version) {
		t.Errorf("size mismatch")
	}
}

func TestEncodeJapanese(t *testing.T) {
	qr, err := Encode("こんにちは", LevelH)
	if err != nil {
		t.Fatal(err)
	}
	if qr.Version < 1 {
		t.Errorf("unexpected version %d", qr.Version)
	}
}

func TestEncodeAllLevels(t *testing.T) {
	levels := []ErrorCorrectionLevel{LevelL, LevelM, LevelQ, LevelH}
	for _, level := range levels {
		qr, err := Encode("TEST", level)
		if err != nil {
			t.Errorf("Encode with level %d failed: %v", level, err)
			continue
		}
		if qr.ECLevel != level {
			t.Errorf("EC level = %d, want %d", qr.ECLevel, level)
		}
	}
}

func TestPNGOutput(t *testing.T) {
	qr, err := Encode("https://gpdf.dev", LevelM)
	if err != nil {
		t.Fatal(err)
	}

	pngData, err := qr.PNG(4)
	if err != nil {
		t.Fatal(err)
	}

	// Check PNG signature.
	sig := []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}
	for i, b := range sig {
		if pngData[i] != b {
			t.Fatalf("invalid PNG signature at byte %d: got %02x, want %02x", i, pngData[i], b)
		}
	}

	// Check IHDR dimensions.
	// IHDR starts at offset 8: length(4) + "IHDR"(4) + width(4) + height(4)
	width := binary.BigEndian.Uint32(pngData[16:20])
	height := binary.BigEndian.Uint32(pngData[20:24])

	quiet := 4
	expectedDim := uint32((qr.Size() + quiet*2) * 4) // scale=4
	if width != expectedDim || height != expectedDim {
		t.Errorf("PNG dimensions = %dx%d, want %dx%d", width, height, expectedDim, expectedDim)
	}
}

func TestPNGScale(t *testing.T) {
	qr, err := Encode("A", LevelL)
	if err != nil {
		t.Fatal(err)
	}

	for _, scale := range []int{1, 5, 10} {
		png, err := qr.PNG(scale)
		if err != nil {
			t.Errorf("PNG(scale=%d) failed: %v", scale, err)
			continue
		}
		width := binary.BigEndian.Uint32(png[16:20])
		expected := uint32((qr.Size() + 8) * scale)
		if width != expected {
			t.Errorf("PNG(scale=%d) width = %d, want %d", scale, width, expected)
		}
	}
}

func TestEncodeEmpty(t *testing.T) {
	_, err := Encode("", LevelM)
	if err == nil {
		t.Error("expected error for empty data")
	}
}

func TestMatrixSymmetry(t *testing.T) {
	// Finder patterns should be in correct positions.
	qr, err := Encode("A", LevelL)
	if err != nil {
		t.Fatal(err)
	}

	size := qr.Size()

	// Top-left finder: (0,0) to (6,6) - corners should be dark.
	if !qr.Matrix[0][0] {
		t.Error("top-left finder corner should be dark")
	}
	if !qr.Matrix[0][6] {
		t.Error("top-left finder corner (0,6) should be dark")
	}

	// Top-right finder: (0, size-7) to (0, size-1).
	if !qr.Matrix[0][size-1] {
		t.Error("top-right finder corner should be dark")
	}

	// Bottom-left finder: (size-7, 0) to (size-1, 0).
	if !qr.Matrix[size-1][0] {
		t.Error("bottom-left finder corner should be dark")
	}
}
