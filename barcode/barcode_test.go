package barcode

import (
	"encoding/binary"
	"testing"
)

func TestEncodeCode128Letters(t *testing.T) {
	// Simple ASCII string should use Code B.
	bc, err := Encode("Hello", Code128)
	if err != nil {
		t.Fatalf("Encode(%q) error: %v", "Hello", err)
	}
	if bc.Data != "Hello" {
		t.Errorf("Data = %q, want %q", bc.Data, "Hello")
	}
	if bc.Format != Code128 {
		t.Errorf("Format = %d, want %d", bc.Format, Code128)
	}
	if len(bc.Pattern) == 0 {
		t.Error("Pattern is empty")
	}
}

func TestEncodeCode128Digits(t *testing.T) {
	// Digit-only string with 4+ digits should use Code C optimization.
	bc, err := Encode("123456", Code128)
	if err != nil {
		t.Fatalf("Encode(%q) error: %v", "123456", err)
	}
	if len(bc.Pattern) == 0 {
		t.Error("Pattern is empty")
	}

	// Verify Code C produces a shorter pattern than Code B would for digits.
	// Code C: StartC + 3 pairs + checksum + stop = 6 symbols
	// Code B: StartB + 6 chars + checksum + stop = 9 symbols
	// Code C pattern should be shorter.
	bcLong, err := Encode("ABCDEF", Code128)
	if err != nil {
		t.Fatalf("Encode(%q) error: %v", "ABCDEF", err)
	}
	if len(bc.Pattern) >= len(bcLong.Pattern) {
		t.Errorf("digit-only pattern (%d modules) should be shorter than letter pattern (%d modules)",
			len(bc.Pattern), len(bcLong.Pattern))
	}
}

func TestEncodeCode128Mixed(t *testing.T) {
	// Mixed content: letters and digits.
	bc, err := Encode("ABC12345678def", Code128)
	if err != nil {
		t.Fatalf("Encode error: %v", err)
	}
	if len(bc.Pattern) == 0 {
		t.Error("Pattern is empty")
	}
}

func TestEncodeCode128Checksum(t *testing.T) {
	// Manually verify checksum for a known simple case.
	// "A" in Code B: StartB(104), A=33, checksum = (104 + 1*33) % 103 = 137 % 103 = 34
	symbols, err := encodeCode128("A")
	if err != nil {
		t.Fatalf("encodeCode128(%q) error: %v", "A", err)
	}

	// Expected: [104, 33, 34, 106]
	// 104 = StartB, 33 = 'A'-32, 34 = checksum, 106 = Stop
	expected := []int{104, 33, 34, 106}
	if len(symbols) != len(expected) {
		t.Fatalf("symbols length = %d, want %d; got %v", len(symbols), len(expected), symbols)
	}
	for i, s := range symbols {
		if s != expected[i] {
			t.Errorf("symbols[%d] = %d, want %d", i, s, expected[i])
		}
	}
}

func TestEncodeCode128ChecksumMultiChar(t *testing.T) {
	// "AB" in Code B:
	// StartB(104), A=33, B=34
	// checksum = (104 + 1*33 + 2*34) % 103 = (104+33+68) % 103 = 205 % 103 = 102
	symbols, err := encodeCode128("AB")
	if err != nil {
		t.Fatalf("encodeCode128(%q) error: %v", "AB", err)
	}

	expected := []int{104, 33, 34, 102, 106}
	if len(symbols) != len(expected) {
		t.Fatalf("symbols length = %d, want %d; got %v", len(symbols), len(expected), symbols)
	}
	for i, s := range symbols {
		if s != expected[i] {
			t.Errorf("symbols[%d] = %d, want %d", i, s, expected[i])
		}
	}
}

func TestEncodeCode128DigitPairs(t *testing.T) {
	// "1234" should use Code C: StartC(105), 12, 34
	// checksum = (105 + 1*12 + 2*34) % 103 = (105+12+68) % 103 = 185 % 103 = 82
	symbols, err := encodeCode128("1234")
	if err != nil {
		t.Fatalf("encodeCode128(%q) error: %v", "1234", err)
	}

	expected := []int{105, 12, 34, 82, 106}
	if len(symbols) != len(expected) {
		t.Fatalf("symbols length = %d, want %d; got %v", len(symbols), len(expected), symbols)
	}
	for i, s := range symbols {
		if s != expected[i] {
			t.Errorf("symbols[%d] = %d, want %d", i, s, expected[i])
		}
	}
}

func TestPNGOutput(t *testing.T) {
	bc, err := Encode("Test123", Code128)
	if err != nil {
		t.Fatalf("Encode error: %v", err)
	}

	png, err := bc.PNG(2, 50)
	if err != nil {
		t.Fatalf("PNG error: %v", err)
	}

	// Check PNG signature.
	sig := []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}
	if len(png) < 8 {
		t.Fatal("PNG too short for signature")
	}
	for i, b := range sig {
		if png[i] != b {
			t.Errorf("signature byte %d = 0x%02x, want 0x%02x", i, png[i], b)
		}
	}

	// Check IHDR chunk.
	// After signature (8 bytes): 4-byte length + 4-byte type "IHDR" + 13-byte data + 4-byte CRC
	if len(png) < 33 {
		t.Fatal("PNG too short for IHDR")
	}
	ihdrType := string(png[12:16])
	if ihdrType != "IHDR" {
		t.Errorf("first chunk type = %q, want %q", ihdrType, "IHDR")
	}

	// Read width and height from IHDR data (starts at offset 16).
	ihdrWidth := binary.BigEndian.Uint32(png[16:20])
	ihdrHeight := binary.BigEndian.Uint32(png[20:24])

	expectedWidth := uint32((quietZone + len(bc.Pattern) + quietZone) * 2)
	if ihdrWidth != expectedWidth {
		t.Errorf("IHDR width = %d, want %d", ihdrWidth, expectedWidth)
	}
	if ihdrHeight != 50 {
		t.Errorf("IHDR height = %d, want 50", ihdrHeight)
	}

	// Check bit depth and color type.
	if png[24] != 8 {
		t.Errorf("bit depth = %d, want 8", png[24])
	}
	if png[25] != 2 {
		t.Errorf("color type = %d, want 2 (RGB)", png[25])
	}
}

func TestPNGDimensions(t *testing.T) {
	bc, err := Encode("X", Code128)
	if err != nil {
		t.Fatalf("Encode error: %v", err)
	}

	barWidth := 3
	height := 100
	png, err := bc.PNG(barWidth, height)
	if err != nil {
		t.Fatalf("PNG error: %v", err)
	}

	ihdrWidth := binary.BigEndian.Uint32(png[16:20])
	ihdrHeight := binary.BigEndian.Uint32(png[20:24])

	expectedWidth := uint32((quietZone + len(bc.Pattern) + quietZone) * barWidth)
	if ihdrWidth != expectedWidth {
		t.Errorf("IHDR width = %d, want %d", ihdrWidth, expectedWidth)
	}
	if ihdrHeight != uint32(height) {
		t.Errorf("IHDR height = %d, want %d", ihdrHeight, height)
	}
}

func TestEncodeEmpty(t *testing.T) {
	_, err := Encode("", Code128)
	if err == nil {
		t.Error("expected error for empty string, got nil")
	}
}

func TestEncodeNonASCII(t *testing.T) {
	_, err := Encode("\x80", Code128)
	if err == nil {
		t.Error("expected error for non-ASCII character, got nil")
	}
}

func TestPNGInvalidParams(t *testing.T) {
	bc, err := Encode("A", Code128)
	if err != nil {
		t.Fatalf("Encode error: %v", err)
	}

	if _, err := bc.PNG(0, 50); err == nil {
		t.Error("expected error for barWidth=0")
	}
	if _, err := bc.PNG(1, 0); err == nil {
		t.Error("expected error for height=0")
	}
}

func TestPatternStartsAndEndsWithBar(t *testing.T) {
	// Code 128 patterns always start with a bar and end with a bar.
	bc, err := Encode("Test", Code128)
	if err != nil {
		t.Fatalf("Encode error: %v", err)
	}
	if !bc.Pattern[0] {
		t.Error("pattern should start with a bar (true)")
	}
	if !bc.Pattern[len(bc.Pattern)-1] {
		t.Error("pattern should end with a bar (true)")
	}
}

func TestEncodeCode128ControlCharacters(t *testing.T) {
	// String starting with a control character should use Code A.
	// \x01 is a control char (< 32), so chooseStartCode returns StartA.
	bc, err := Encode("\x01", Code128)
	if err != nil {
		t.Fatalf("Encode(ctrl) error: %v", err)
	}
	if len(bc.Pattern) == 0 {
		t.Error("Pattern is empty")
	}
}

func TestEncodeSetA_PrintableASCII(t *testing.T) {
	// Control char followed by printable ASCII in Code A range (32-95).
	// \x01 starts in Code A, then 'A' (65) stays in A.
	bc, err := Encode("\x01A", Code128)
	if err != nil {
		t.Fatalf("Encode error: %v", err)
	}
	if len(bc.Pattern) == 0 {
		t.Error("Pattern is empty")
	}
}

func TestEncodeSetA_SwitchToB(t *testing.T) {
	// Control char followed by lowercase (>95) forces switch A→B.
	bc, err := Encode("\x01a", Code128)
	if err != nil {
		t.Fatalf("Encode error: %v", err)
	}
	if len(bc.Pattern) == 0 {
		t.Error("Pattern is empty")
	}
}

func TestEncodeSetA_SwitchToC(t *testing.T) {
	// Control char followed by 4+ digits forces switch A→C.
	bc, err := Encode("\x011234", Code128)
	if err != nil {
		t.Fatalf("Encode error: %v", err)
	}
	if len(bc.Pattern) == 0 {
		t.Error("Pattern is empty")
	}
}

func TestEncodeSetB_ControlCharPath(t *testing.T) {
	// Code B with a control character triggers temporary switch to A.
	// "A\x01B" starts in B, 'A' in B, \x01 switches to A and back, 'B' in B.
	symbols, err := encodeCode128("A\x01B")
	if err != nil {
		t.Fatalf("encodeCode128 error: %v", err)
	}
	// Should contain SwitchA and SwitchB symbols.
	foundSwitchA := false
	foundSwitchB := false
	for _, s := range symbols {
		if s == code128SwitchA {
			foundSwitchA = true
		}
		if s == code128SwitchB {
			foundSwitchB = true
		}
	}
	if !foundSwitchA {
		t.Error("expected SwitchA symbol for control char in Code B")
	}
	if !foundSwitchB {
		t.Error("expected SwitchB symbol after control char")
	}
}

func TestChooseStartCode_ControlChar(t *testing.T) {
	start, set := chooseStartCode("\x01ABC")
	if start != code128StartA {
		t.Errorf("start = %d, want %d (StartA)", start, code128StartA)
	}
	if set != 'A' {
		t.Errorf("set = %c, want A", set)
	}
}

func TestEncodeSetA_MultipleControlChars(t *testing.T) {
	// Multiple control characters stay in Code A.
	bc, err := Encode("\x01\x02\x03", Code128)
	if err != nil {
		t.Fatalf("Encode error: %v", err)
	}
	if len(bc.Pattern) == 0 {
		t.Error("Pattern is empty")
	}
}

func TestCountDigits(t *testing.T) {
	tests := []struct {
		data string
		pos  int
		want int
	}{
		{"123456", 0, 6},
		{"abc123", 3, 3},
		{"abc", 0, 0},
		{"12ab34", 0, 2},
		{"12ab34", 4, 2},
	}
	for _, tt := range tests {
		got := countDigits(tt.data, tt.pos)
		if got != tt.want {
			t.Errorf("countDigits(%q, %d) = %d, want %d", tt.data, tt.pos, got, tt.want)
		}
	}
}
