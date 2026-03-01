package qrcode

// matrix represents a QR code module grid.
type matrix struct {
	size    int
	modules [][]bool // true = dark module
	set     [][]bool // true = this module has been placed (not available for data)
}

func newMatrix(size int) *matrix {
	m := &matrix{
		size:    size,
		modules: make([][]bool, size),
		set:     make([][]bool, size),
	}
	for i := range m.modules {
		m.modules[i] = make([]bool, size)
		m.set[i] = make([]bool, size)
	}
	return m
}

// setModule sets a module at (row, col) and marks it as placed.
func (m *matrix) setModule(row, col int, dark bool) {
	if row >= 0 && row < m.size && col >= 0 && col < m.size {
		m.modules[row][col] = dark
		m.set[row][col] = true
	}
}

// isSet returns true if the module at (row, col) has been placed.
func (m *matrix) isSet(row, col int) bool {
	if row >= 0 && row < m.size && col >= 0 && col < m.size {
		return m.set[row][col]
	}
	return true // out of bounds is treated as set
}

// placeFinderPattern places a 7x7 finder pattern with its center at (centerRow, centerCol).
func (m *matrix) placeFinderPattern(centerRow, centerCol int) {
	for dr := -3; dr <= 3; dr++ {
		for dc := -3; dc <= 3; dc++ {
			r, c := centerRow+dr, centerCol+dc
			// Dark if on the border, or in the center 3x3, but not the ring between.
			abs_dr, abs_dc := dr, dc
			if abs_dr < 0 {
				abs_dr = -abs_dr
			}
			if abs_dc < 0 {
				abs_dc = -abs_dc
			}
			maxDist := abs_dr
			if abs_dc > maxDist {
				maxDist = abs_dc
			}
			dark := maxDist == 3 || maxDist == 0 || (abs_dr <= 1 && abs_dc <= 1)
			m.setModule(r, c, dark)
		}
	}
}

// placeSeparators places the white separator lines around finder patterns.
func (m *matrix) placeSeparators(size int) {
	// Top-left finder: separator along row 7 (cols 0-7) and col 7 (rows 0-7).
	for i := 0; i <= 7; i++ {
		m.setModule(7, i, false)
		m.setModule(i, 7, false)
	}
	// Top-right finder: separator along row 7 (cols size-8 to size-1) and col size-8 (rows 0-7).
	for i := 0; i <= 7; i++ {
		m.setModule(7, size-8+i, false)
		m.setModule(i, size-8, false)
	}
	// Bottom-left finder: separator along row size-8 (cols 0-7) and col 7 (rows size-8 to size-1).
	for i := 0; i <= 7; i++ {
		m.setModule(size-8, i, false)
		m.setModule(size-8+i, 7, false)
	}
}

// placeTimingPatterns places the horizontal and vertical timing patterns.
func (m *matrix) placeTimingPatterns(size int) {
	for i := 8; i < size-8; i++ {
		dark := i%2 == 0
		if !m.isSet(6, i) {
			m.setModule(6, i, dark) // horizontal
		}
		if !m.isSet(i, 6) {
			m.setModule(i, 6, dark) // vertical
		}
	}
}

// placeAlignmentPatterns places alignment patterns for the given version.
func (m *matrix) placeAlignmentPatterns(version int) {
	positions := alignmentPositions[version]
	for _, row := range positions {
		for _, col := range positions {
			// Skip if it overlaps with a finder pattern.
			if m.isSet(row, col) {
				continue
			}
			m.placeAlignmentPattern(row, col)
		}
	}
}

// placeAlignmentPattern places a single 5x5 alignment pattern centered at (row, col).
func (m *matrix) placeAlignmentPattern(row, col int) {
	for dr := -2; dr <= 2; dr++ {
		for dc := -2; dc <= 2; dc++ {
			abs_dr, abs_dc := dr, dc
			if abs_dr < 0 {
				abs_dr = -abs_dr
			}
			if abs_dc < 0 {
				abs_dc = -abs_dc
			}
			maxDist := abs_dr
			if abs_dc > maxDist {
				maxDist = abs_dc
			}
			dark := maxDist == 2 || maxDist == 0
			m.setModule(row+dr, col+dc, dark)
		}
	}
}

// reserveFormatArea reserves the format information bits
// (they are written later after masking).
func (m *matrix) reserveFormatArea(size int) {
	// Around top-left finder pattern.
	for i := 0; i <= 8; i++ {
		if !m.isSet(8, i) {
			m.setModule(8, i, false)
		}
		if !m.isSet(i, 8) {
			m.setModule(i, 8, false)
		}
	}
	// Below top-right finder.
	for i := 0; i <= 7; i++ {
		if !m.isSet(8, size-1-i) {
			m.setModule(8, size-1-i, false)
		}
	}
	// Right of bottom-left finder.
	for i := 0; i <= 7; i++ {
		if !m.isSet(size-1-i, 8) {
			m.setModule(size-1-i, 8, false)
		}
	}
	// Dark module (always required).
	m.setModule(size-8, 8, true)
}

// reserveVersionArea reserves the version information bits for version >= 7.
func (m *matrix) reserveVersionArea(version, size int) {
	if version < 7 {
		return
	}
	// Two 6x3 blocks: one near bottom-left, one near top-right.
	for i := 0; i < 6; i++ {
		for j := 0; j < 3; j++ {
			m.setModule(size-11+j, i, false)
			m.setModule(i, size-11+j, false)
		}
	}
}

// placeData places the data bits in the zigzag pattern.
func (m *matrix) placeData(data []byte) {
	size := m.size
	bitIdx := 0
	totalBits := len(data) * 8

	// Right-to-left column pairs, skipping column 6 (timing pattern).
	for col := size - 1; col >= 0; col -= 2 {
		if col == 6 {
			col-- // skip timing pattern column
		}
		// Upward or downward depending on column pair position.
		upward := ((size-1-col)/2)%2 == 0

		for i := 0; i < size; i++ {
			row := i
			if upward {
				row = size - 1 - i
			}

			for dc := 0; dc <= 1; dc++ {
				c := col - dc
				if c < 0 {
					continue
				}
				if m.isSet(row, c) {
					continue
				}
				if bitIdx < totalBits {
					byteIdx := bitIdx / 8
					bitPos := 7 - bitIdx%8
					dark := (data[byteIdx]>>uint(bitPos))&1 == 1
					m.modules[row][c] = dark
					m.set[row][c] = true
					bitIdx++
				} else {
					// Fill remaining with false (white).
					m.modules[row][c] = false
					m.set[row][c] = true
				}
			}
		}
	}
}

// writeFormatInfo writes the 15-bit format information.
func (m *matrix) writeFormatInfo(ecLevel ErrorCorrectionLevel, maskID int) {
	bits := formatInfo[ecLevel][maskID]
	size := m.size

	// Horizontal strip near top-left finder, and vertical strip.
	for i := 0; i <= 5; i++ {
		m.modules[8][i] = (bits>>(14-i))&1 == 1
		m.modules[i][8] = (bits>>(i))&1 == 1
	}
	// Bit 6 at (8, 7).
	m.modules[8][7] = (bits>>8)&1 == 1
	// Bit 7 at (8, 8).
	m.modules[8][8] = (bits>>7)&1 == 1
	// Bit 8 at (7, 8).
	m.modules[7][8] = (bits>>6)&1 == 1

	// Bits 9-14 along left side.
	for i := 9; i <= 14; i++ {
		m.modules[14-i][8] = (bits>>(14-i))&1 == 1
	}

	// Second copy: horizontal near top-right.
	for i := 0; i <= 7; i++ {
		m.modules[8][size-1-i] = (bits>>i)&1 == 1
	}

	// Second copy: vertical near bottom-left.
	for i := 0; i <= 6; i++ {
		m.modules[size-1-i][8] = (bits>>(14-i))&1 == 1
	}
	// Dark module (always dark).
	m.modules[size-8][8] = true
}

// writeVersionInfo writes the 18-bit version information for version >= 7.
func (m *matrix) writeVersionInfo(version int) {
	if version < 7 {
		return
	}
	bits := versionInfo[version-7]
	size := m.size

	for i := 0; i < 18; i++ {
		row := i / 3
		col := i % 3
		dark := (bits>>i)&1 == 1
		// Near bottom-left.
		m.modules[size-11+col][row] = dark
		// Near top-right.
		m.modules[row][size-11+col] = dark
	}
}

// buildMatrix constructs the QR code matrix for the given data, version,
// EC level, and mask pattern. It returns the completed matrix.
func buildMatrix(data []byte, version int, ecLevel ErrorCorrectionLevel, maskID int) *matrix {
	size := moduleSize(version)
	m := newMatrix(size)

	// 1. Finder patterns.
	m.placeFinderPattern(3, 3)      // top-left
	m.placeFinderPattern(3, size-4) // top-right
	m.placeFinderPattern(size-4, 3) // bottom-left
	m.placeSeparators(size)

	// 2. Alignment patterns.
	m.placeAlignmentPatterns(version)

	// 3. Timing patterns.
	m.placeTimingPatterns(size)

	// 4. Reserve format and version info areas.
	m.reserveFormatArea(size)
	m.reserveVersionArea(version, size)

	// 5. Place data bits.
	m.placeData(data)

	// 6. Apply mask.
	applyMask(m, maskID)

	// 7. Write format info.
	m.writeFormatInfo(ecLevel, maskID)

	// 8. Write version info.
	m.writeVersionInfo(version)

	return m
}
