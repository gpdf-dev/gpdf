package qrcode

// maskFunc returns true if the module at (row, col) should be flipped.
type maskFunc func(row, col int) bool

var maskFuncs = [8]maskFunc{
	func(r, c int) bool { return (r+c)%2 == 0 },
	func(r, c int) bool { return r%2 == 0 },
	func(r, c int) bool { return c%3 == 0 },
	func(r, c int) bool { return (r+c)%3 == 0 },
	func(r, c int) bool { return (r/2+c/3)%2 == 0 },
	func(r, c int) bool { return (r*c)%2+(r*c)%3 == 0 },
	func(r, c int) bool { return ((r*c)%2+(r*c)%3)%2 == 0 },
	func(r, c int) bool { return ((r+c)%2+(r*c)%3)%2 == 0 },
}

// applyMask flips data modules according to the mask pattern.
// Only data modules (not function patterns) are affected.
func applyMask(m *matrix, maskID int) {
	fn := maskFuncs[maskID]
	size := m.size

	// We need to know which modules are "function patterns" (finder, timing,
	// alignment, format/version info). We use a separate matrix to track
	// which modules were set BEFORE data placement. However, since placeData
	// has already been called, we need to identify function pattern modules.
	// The trick: function pattern modules were set before placeData. After
	// placeData, ALL modules are set. We reconstruct function patterns by
	// building a fresh marker matrix.
	funcPat := newFunctionPatternMask(m.size, guessVersion(m.size))

	for r := 0; r < size; r++ {
		for c := 0; c < size; c++ {
			if funcPat[r][c] {
				continue
			}
			if fn(r, c) {
				m.modules[r][c] = !m.modules[r][c]
			}
		}
	}
}

func guessVersion(size int) int {
	return (size - 17) / 4
}

// newFunctionPatternMask returns a mask where true means the module is
// part of a function pattern (not data).
func newFunctionPatternMask(size, version int) [][]bool {
	mask := make([][]bool, size)
	for i := range mask {
		mask[i] = make([]bool, size)
	}

	// Finder patterns + separators.
	markRect(mask, 0, 0, 9, 9)      // top-left
	markRect(mask, 0, size-8, 9, 8) // top-right
	markRect(mask, size-8, 0, 8, 9) // bottom-left

	// Timing patterns.
	for i := 8; i < size-8; i++ {
		mask[6][i] = true
		mask[i][6] = true
	}

	// Alignment patterns.
	positions := alignmentPositions[version]
	for _, row := range positions {
		for _, col := range positions {
			// Skip if overlap with finder pattern area.
			if overlapsFinder(row, col, size) {
				continue
			}
			markRect(mask, row-2, col-2, 5, 5)
		}
	}

	// Format info.
	for i := 0; i <= 8; i++ {
		mask[8][i] = true
		mask[i][8] = true
	}
	for i := 0; i <= 7; i++ {
		mask[8][size-1-i] = true
		mask[size-1-i][8] = true
	}

	// Version info (for version >= 7).
	if version >= 7 {
		for i := 0; i < 6; i++ {
			for j := 0; j < 3; j++ {
				mask[size-11+j][i] = true
				mask[i][size-11+j] = true
			}
		}
	}

	return mask
}

func overlapsFinder(row, col, size int) bool {
	// Top-left finder occupies [0..8, 0..8]
	if row <= 8 && col <= 8 {
		return true
	}
	// Top-right finder occupies [0..8, size-9..size-1]
	if row <= 8 && col >= size-9 {
		return true
	}
	// Bottom-left finder occupies [size-9..size-1, 0..8]
	if row >= size-9 && col <= 8 {
		return true
	}
	return false
}

func markRect(mask [][]bool, row, col, h, w int) {
	for r := row; r < row+h && r < len(mask); r++ {
		for c := col; c < col+w && c < len(mask[0]); c++ {
			if r >= 0 && c >= 0 {
				mask[r][c] = true
			}
		}
	}
}

// chooseBestMask evaluates all 8 mask patterns and returns the one with
// the lowest penalty score.
func chooseBestMask(data []byte, version int, ecLevel ErrorCorrectionLevel) int {
	bestMask := 0
	bestScore := -1

	for maskID := 0; maskID < 8; maskID++ {
		m := buildMatrixForScoring(data, version, ecLevel, maskID)
		score := penaltyScore(m)
		if bestScore < 0 || score < bestScore {
			bestScore = score
			bestMask = maskID
		}
	}
	return bestMask
}

// buildMatrixForScoring builds a matrix without writing format/version info
// (format info is written, but we need it for penalty calculation).
func buildMatrixForScoring(data []byte, version int, ecLevel ErrorCorrectionLevel, maskID int) *matrix {
	return buildMatrix(data, version, ecLevel, maskID)
}

// penaltyScore computes the total penalty score for a QR matrix (4 rules).
func penaltyScore(m *matrix) int {
	return penalty1(m) + penalty2(m) + penalty3(m) + penalty4(m)
}

// penalty1: Adjacent modules in row/column that are same color.
// N1 (3) + (count - 5) for runs of 5+ same-color modules.
func penalty1(m *matrix) int {
	score := 0
	size := m.size

	// Horizontal.
	for r := 0; r < size; r++ {
		count := 1
		for c := 1; c < size; c++ {
			if m.modules[r][c] == m.modules[r][c-1] {
				count++
			} else {
				if count >= 5 {
					score += count - 2
				}
				count = 1
			}
		}
		if count >= 5 {
			score += count - 2
		}
	}

	// Vertical.
	for c := 0; c < size; c++ {
		count := 1
		for r := 1; r < size; r++ {
			if m.modules[r][c] == m.modules[r-1][c] {
				count++
			} else {
				if count >= 5 {
					score += count - 2
				}
				count = 1
			}
		}
		if count >= 5 {
			score += count - 2
		}
	}

	return score
}

// penalty2: 2x2 blocks of same-color modules.
// N2 (3) per 2x2 block.
func penalty2(m *matrix) int {
	score := 0
	size := m.size

	for r := 0; r < size-1; r++ {
		for c := 0; c < size-1; c++ {
			v := m.modules[r][c]
			if v == m.modules[r][c+1] && v == m.modules[r+1][c] && v == m.modules[r+1][c+1] {
				score += 3
			}
		}
	}
	return score
}

// penalty3: Patterns that look like finder patterns.
// N3 (40) for each occurrence of 1:1:3:1:1 pattern preceded/followed by 4 white modules.
func penalty3(m *matrix) int {
	score := 0
	size := m.size
	// Pattern: dark, light, dark*3, light, dark, then 4 light (or reverse).
	pattern1 := [11]bool{true, false, true, true, true, false, true, false, false, false, false}
	pattern2 := [11]bool{false, false, false, false, true, false, true, true, true, false, true}

	for r := 0; r < size; r++ {
		for c := 0; c <= size-11; c++ {
			match1, match2 := true, true
			for k := 0; k < 11; k++ {
				if m.modules[r][c+k] != pattern1[k] {
					match1 = false
				}
				if m.modules[r][c+k] != pattern2[k] {
					match2 = false
				}
			}
			if match1 || match2 {
				score += 40
			}
		}
	}

	for c := 0; c < size; c++ {
		for r := 0; r <= size-11; r++ {
			match1, match2 := true, true
			for k := 0; k < 11; k++ {
				if m.modules[r+k][c] != pattern1[k] {
					match1 = false
				}
				if m.modules[r+k][c] != pattern2[k] {
					match2 = false
				}
			}
			if match1 || match2 {
				score += 40
			}
		}
	}

	return score
}

// penalty4: Proportion of dark modules.
// N4 (10) * floor(|(dark% - 50)| / 5).
func penalty4(m *matrix) int {
	total := m.size * m.size
	dark := 0
	for r := 0; r < m.size; r++ {
		for c := 0; c < m.size; c++ {
			if m.modules[r][c] {
				dark++
			}
		}
	}

	pct := dark * 100 / total
	// Steps of 5% from 50%.
	prev5 := (pct / 5) * 5
	next5 := prev5 + 5

	diff1 := prev5 - 50
	if diff1 < 0 {
		diff1 = -diff1
	}
	diff2 := next5 - 50
	if diff2 < 0 {
		diff2 = -diff2
	}
	minDiff := diff1
	if diff2 < minDiff {
		minDiff = diff2
	}

	return minDiff / 5 * 10
}
