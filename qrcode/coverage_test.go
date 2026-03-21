package qrcode

import (
	"testing"
)

// ---------------------------------------------------------------------------
// writeVersionInfo tests (version >= 7)
// ---------------------------------------------------------------------------

func TestWriteVersionInfo_Version7(t *testing.T) {
	version := 7
	size := moduleSize(version) // 17 + 7*4 = 45
	m := newMatrix(size)

	m.writeVersionInfo(version)

	bits := versionInfo[version-7]

	// Verify bits are placed in the correct positions.
	for i := 0; i < 18; i++ {
		row := i / 3
		col := i % 3
		dark := (bits>>i)&1 == 1

		// Near bottom-left.
		if m.modules[size-11+col][row] != dark {
			t.Errorf("bottom-left bit %d at (%d,%d): got %v, want %v",
				i, size-11+col, row, m.modules[size-11+col][row], dark)
		}
		// Near top-right.
		if m.modules[row][size-11+col] != dark {
			t.Errorf("top-right bit %d at (%d,%d): got %v, want %v",
				i, row, size-11+col, m.modules[row][size-11+col], dark)
		}
	}
}

func TestWriteVersionInfo_SkipsBelow7(t *testing.T) {
	version := 6
	size := moduleSize(version)
	m := newMatrix(size)

	// No panic, no changes.
	m.writeVersionInfo(version)

	// All modules should still be false (default).
	for r := 0; r < size; r++ {
		for c := 0; c < size; c++ {
			if m.modules[r][c] {
				t.Errorf("module (%d,%d) should be false for version < 7", r, c)
			}
		}
	}
}

func TestWriteVersionInfo_Version40(t *testing.T) {
	version := 40
	size := moduleSize(version)
	m := newMatrix(size)

	// Should not panic.
	m.writeVersionInfo(version)

	bits := versionInfo[version-7]
	// Spot-check bit 0.
	dark := bits&1 == 1
	if m.modules[size-11][0] != dark {
		t.Errorf("version 40 bit 0: got %v, want %v", m.modules[size-11][0], dark)
	}
}

// ---------------------------------------------------------------------------
// reserveVersionArea tests (version >= 7)
// ---------------------------------------------------------------------------

func TestReserveVersionArea_Version7(t *testing.T) {
	version := 7
	size := moduleSize(version)
	m := newMatrix(size)

	m.reserveVersionArea(version, size)

	// Verify 6x3 reserved blocks are marked as set.
	for i := 0; i < 6; i++ {
		for j := 0; j < 3; j++ {
			// Near bottom-left.
			if !m.set[size-11+j][i] {
				t.Errorf("bottom-left (%d,%d) not reserved", size-11+j, i)
			}
			// Near top-right.
			if !m.set[i][size-11+j] {
				t.Errorf("top-right (%d,%d) not reserved", i, size-11+j)
			}
		}
	}
}

func TestReserveVersionArea_SkipsBelow7(t *testing.T) {
	version := 6
	size := moduleSize(version)
	m := newMatrix(size)

	m.reserveVersionArea(version, size)

	// No modules should be set.
	for r := 0; r < size; r++ {
		for c := 0; c < size; c++ {
			if m.set[r][c] {
				t.Errorf("module (%d,%d) should not be set for version < 7", r, c)
			}
		}
	}
}

// ---------------------------------------------------------------------------
// gfDiv tests
// ---------------------------------------------------------------------------

func TestGfDiv_ZeroDividend(t *testing.T) {
	// 0 / a = 0 for any nonzero a.
	for i := 1; i < 256; i++ {
		if got := gfDiv(0, byte(i)); got != 0 {
			t.Errorf("gfDiv(0, %d) = %d, want 0", i, got)
		}
	}
}

func TestGfDiv_DivisionByZeroPanics(t *testing.T) {
	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("expected panic for division by zero")
		}
	}()
	gfDiv(1, 0)
}

func TestGfDiv_NonInverse(t *testing.T) {
	// a / b should satisfy (a/b) * b = a.
	testCases := [][2]byte{
		{7, 3},
		{255, 1},
		{100, 50},
		{1, 255},
		{128, 64},
	}
	for _, tc := range testCases {
		a, b := tc[0], tc[1]
		result := gfDiv(a, b)
		// Verify: result * b = a
		product := gfMul(result, b)
		if product != a {
			t.Errorf("gfDiv(%d, %d) = %d, but %d * %d = %d, want %d", a, b, result, result, b, product, a)
		}
	}
}

func TestGfDiv_AllNonZero(t *testing.T) {
	// For all nonzero a, a / 1 = a.
	for i := 1; i < 256; i++ {
		if got := gfDiv(byte(i), 1); got != byte(i) {
			t.Errorf("gfDiv(%d, 1) = %d, want %d", i, got, i)
		}
	}
}

// ---------------------------------------------------------------------------
// isSet tests
// ---------------------------------------------------------------------------

func TestIsSet_OutOfBounds(t *testing.T) {
	m := newMatrix(5)

	// Out-of-bounds should return true.
	tests := []struct {
		row, col int
		desc     string
	}{
		{-1, 0, "negative row"},
		{0, -1, "negative col"},
		{-1, -1, "both negative"},
		{5, 0, "row == size"},
		{0, 5, "col == size"},
		{5, 5, "both == size"},
		{100, 0, "large row"},
		{0, 100, "large col"},
	}
	for _, tc := range tests {
		if !m.isSet(tc.row, tc.col) {
			t.Errorf("isSet(%d, %d) [%s] = false, want true", tc.row, tc.col, tc.desc)
		}
	}
}

func TestIsSet_InBounds_NotSet(t *testing.T) {
	m := newMatrix(5)
	// A module that hasn't been placed should return false.
	if m.isSet(2, 2) {
		t.Error("isSet(2, 2) = true, want false for unplaced module")
	}
}

func TestIsSet_InBounds_Set(t *testing.T) {
	m := newMatrix(5)
	m.setModule(2, 2, true)
	if !m.isSet(2, 2) {
		t.Error("isSet(2, 2) = false, want true for placed module")
	}
}

// ---------------------------------------------------------------------------
// Integration: buildMatrix with version >= 7
// ---------------------------------------------------------------------------

func TestBuildMatrix_Version7(t *testing.T) {
	// Version 7 requires 45 modules. Generate enough data to fill.
	version := 7
	size := moduleSize(version)
	dataLen := size * size / 8 // approximate
	data := make([]byte, dataLen)
	for i := range data {
		data[i] = 0xAA
	}

	m := buildMatrix(data, version, LevelM, 0)
	if m.size != size {
		t.Errorf("matrix size = %d, want %d", m.size, size)
	}

	// Verify version info area is written (check a spot in version area).
	bits := versionInfo[0] // version 7 = index 0
	dark := bits&1 == 1
	if m.modules[size-11][0] != dark {
		t.Errorf("version info bit 0 mismatch")
	}
}
