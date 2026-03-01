package qrcode

// ErrorCorrectionLevel specifies the error correction capability.
type ErrorCorrectionLevel int

const (
	LevelL ErrorCorrectionLevel = iota // ~7% recovery
	LevelM                             // ~15% recovery
	LevelQ                             // ~25% recovery
	LevelH                             // ~30% recovery
)

// mode indicates the encoding mode for data segments.
type mode int

const (
	modeNumeric      mode = iota // 0-9
	modeAlphanumeric             // 0-9, A-Z, space, $%*+-./:
	modeByte                     // ISO 8859-1 / UTF-8
)

// modeIndicator returns the 4-bit mode indicator.
func modeIndicator(m mode) int {
	switch m {
	case modeNumeric:
		return 0b0001
	case modeAlphanumeric:
		return 0b0010
	case modeByte:
		return 0b0100
	}
	return 0b0100
}

// charCountBits returns the number of character count indicator bits for a given version and mode.
func charCountBits(version int, m mode) int {
	var idx int
	switch {
	case version <= 9:
		idx = 0
	case version <= 26:
		idx = 1
	default:
		idx = 2
	}
	table := [3][3]int{
		{10, 9, 8},   // v1-9:  numeric, alphanum, byte
		{12, 11, 16}, // v10-26
		{14, 13, 16}, // v27-40
	}
	return table[idx][m]
}

// ecBlockInfo describes the error correction block structure.
type ecBlockInfo struct {
	totalCodewords int // total data + EC codewords in this version/level
	ecPerBlock     int // EC codewords per block
	group1Blocks   int // number of blocks in group 1
	group1Data     int // data codewords per block in group 1
	group2Blocks   int // number of blocks in group 2 (0 if none)
	group2Data     int // data codewords per block in group 2
}

// dataCapacity returns the total data codewords for a version/level.
func (e ecBlockInfo) dataCapacity() int {
	return e.group1Blocks*e.group1Data + e.group2Blocks*e.group2Data
}

// ecTable contains error correction parameters for versions 1-40, all 4 levels.
// Index: [version-1][level]
var ecTable = [40][4]ecBlockInfo{
	// Version 1
	{{26, 7, 1, 19, 0, 0}, {26, 10, 1, 16, 0, 0}, {26, 13, 1, 13, 0, 0}, {26, 17, 1, 9, 0, 0}},
	// Version 2
	{{44, 10, 1, 34, 0, 0}, {44, 16, 1, 28, 0, 0}, {44, 22, 1, 22, 0, 0}, {44, 28, 1, 16, 0, 0}},
	// Version 3
	{{70, 15, 1, 55, 0, 0}, {70, 26, 1, 44, 0, 0}, {70, 18, 2, 17, 0, 0}, {70, 22, 2, 13, 0, 0}},
	// Version 4
	{{100, 20, 1, 80, 0, 0}, {100, 18, 2, 32, 0, 0}, {100, 26, 2, 24, 0, 0}, {100, 16, 4, 9, 0, 0}},
	// Version 5
	{{134, 26, 1, 108, 0, 0}, {134, 24, 2, 43, 0, 0}, {134, 18, 2, 15, 2, 16}, {134, 22, 2, 11, 2, 12}},
	// Version 6
	{{172, 18, 2, 68, 0, 0}, {172, 16, 4, 27, 0, 0}, {172, 24, 4, 19, 0, 0}, {172, 28, 4, 15, 0, 0}},
	// Version 7
	{{196, 20, 2, 78, 0, 0}, {196, 18, 4, 31, 0, 0}, {196, 18, 2, 14, 4, 15}, {196, 26, 4, 13, 1, 14}},
	// Version 8
	{{242, 24, 2, 97, 0, 0}, {242, 22, 2, 38, 2, 39}, {242, 22, 4, 18, 2, 19}, {242, 26, 4, 14, 2, 15}},
	// Version 9
	{{292, 30, 2, 116, 0, 0}, {292, 22, 3, 36, 2, 37}, {292, 20, 4, 16, 4, 17}, {292, 24, 4, 12, 4, 13}},
	// Version 10
	{{346, 18, 2, 68, 2, 69}, {346, 26, 4, 43, 1, 44}, {346, 24, 6, 19, 2, 20}, {346, 28, 6, 15, 2, 16}},
	// Version 11
	{{404, 20, 4, 81, 0, 0}, {404, 30, 1, 50, 4, 51}, {404, 28, 4, 22, 4, 23}, {404, 24, 3, 12, 8, 13}},
	// Version 12
	{{466, 24, 2, 92, 2, 93}, {466, 22, 6, 36, 2, 37}, {466, 26, 4, 20, 6, 21}, {466, 28, 7, 14, 4, 15}},
	// Version 13
	{{532, 26, 4, 107, 0, 0}, {532, 22, 8, 37, 1, 38}, {532, 24, 8, 20, 4, 21}, {532, 22, 12, 11, 4, 12}},
	// Version 14
	{{581, 30, 3, 115, 1, 116}, {581, 24, 4, 40, 5, 41}, {581, 20, 11, 16, 5, 17}, {581, 24, 11, 12, 5, 13}},
	// Version 15
	{{655, 22, 5, 87, 1, 88}, {655, 24, 5, 41, 5, 42}, {655, 30, 5, 24, 7, 25}, {655, 24, 11, 12, 7, 13}},
	// Version 16
	{{733, 24, 5, 98, 1, 99}, {733, 28, 7, 45, 3, 46}, {733, 24, 15, 19, 2, 20}, {733, 30, 3, 15, 13, 16}},
	// Version 17
	{{815, 28, 1, 107, 5, 108}, {815, 28, 10, 46, 1, 47}, {815, 28, 1, 22, 15, 23}, {815, 28, 2, 14, 17, 15}},
	// Version 18
	{{901, 30, 5, 120, 1, 121}, {901, 26, 9, 43, 4, 44}, {901, 28, 17, 22, 1, 23}, {901, 28, 2, 14, 19, 15}},
	// Version 19
	{{991, 28, 3, 113, 4, 114}, {991, 26, 3, 44, 11, 45}, {991, 26, 17, 21, 4, 22}, {991, 26, 9, 13, 16, 14}},
	// Version 20
	{{1085, 28, 3, 107, 5, 108}, {1085, 26, 3, 41, 13, 42}, {1085, 30, 15, 24, 5, 25}, {1085, 28, 15, 15, 10, 16}},
	// Version 21
	{{1156, 28, 4, 116, 4, 117}, {1156, 26, 17, 42, 0, 0}, {1156, 28, 17, 22, 6, 23}, {1156, 30, 19, 16, 6, 17}},
	// Version 22
	{{1258, 28, 2, 111, 7, 112}, {1258, 28, 17, 46, 0, 0}, {1258, 30, 7, 24, 16, 25}, {1258, 24, 34, 13, 0, 0}},
	// Version 23
	{{1364, 30, 4, 121, 5, 122}, {1364, 28, 4, 47, 14, 48}, {1364, 30, 11, 24, 14, 25}, {1364, 30, 16, 15, 14, 16}},
	// Version 24
	{{1474, 30, 6, 117, 4, 118}, {1474, 28, 6, 45, 14, 46}, {1474, 30, 11, 24, 16, 25}, {1474, 30, 30, 16, 2, 17}},
	// Version 25
	{{1588, 26, 8, 106, 4, 107}, {1588, 28, 8, 47, 13, 48}, {1588, 30, 7, 24, 22, 25}, {1588, 30, 22, 15, 13, 16}},
	// Version 26
	{{1706, 28, 10, 114, 2, 115}, {1706, 28, 19, 46, 4, 47}, {1706, 28, 28, 22, 6, 23}, {1706, 30, 33, 16, 4, 17}},
	// Version 27
	{{1828, 30, 8, 122, 4, 123}, {1828, 28, 22, 45, 3, 46}, {1828, 30, 8, 23, 26, 24}, {1828, 30, 12, 15, 28, 16}},
	// Version 28
	{{1921, 30, 3, 117, 10, 118}, {1921, 28, 3, 45, 23, 46}, {1921, 30, 4, 24, 31, 25}, {1921, 30, 11, 15, 31, 16}},
	// Version 29
	{{2051, 30, 7, 116, 7, 117}, {2051, 28, 21, 45, 7, 46}, {2051, 30, 1, 23, 37, 24}, {2051, 30, 19, 15, 26, 16}},
	// Version 30
	{{2185, 30, 5, 115, 10, 116}, {2185, 28, 19, 47, 10, 48}, {2185, 30, 15, 24, 25, 25}, {2185, 30, 23, 15, 25, 16}},
	// Version 31
	{{2323, 30, 13, 115, 3, 116}, {2323, 28, 2, 46, 29, 47}, {2323, 30, 42, 24, 1, 25}, {2323, 30, 23, 15, 28, 16}},
	// Version 32
	{{2465, 30, 17, 115, 0, 0}, {2465, 28, 10, 46, 23, 47}, {2465, 30, 10, 24, 35, 25}, {2465, 30, 19, 15, 35, 16}},
	// Version 33
	{{2611, 30, 17, 115, 1, 116}, {2611, 28, 14, 46, 21, 47}, {2611, 30, 29, 24, 19, 25}, {2611, 30, 11, 15, 46, 16}},
	// Version 34
	{{2761, 30, 13, 115, 6, 116}, {2761, 28, 14, 46, 23, 47}, {2761, 30, 44, 24, 7, 25}, {2761, 30, 59, 16, 1, 17}},
	// Version 35
	{{2876, 30, 12, 121, 7, 122}, {2876, 28, 12, 47, 26, 48}, {2876, 30, 39, 24, 14, 25}, {2876, 30, 22, 15, 41, 16}},
	// Version 36
	{{3034, 30, 6, 121, 14, 122}, {3034, 28, 6, 47, 34, 48}, {3034, 30, 46, 24, 10, 25}, {3034, 30, 2, 15, 64, 16}},
	// Version 37
	{{3196, 30, 17, 122, 4, 123}, {3196, 28, 29, 46, 14, 47}, {3196, 30, 49, 24, 10, 25}, {3196, 30, 24, 15, 46, 16}},
	// Version 38
	{{3362, 30, 4, 122, 18, 123}, {3362, 28, 13, 46, 32, 47}, {3362, 30, 48, 24, 14, 25}, {3362, 30, 42, 15, 32, 16}},
	// Version 39
	{{3532, 30, 20, 117, 4, 118}, {3532, 28, 40, 47, 7, 48}, {3532, 30, 43, 24, 22, 25}, {3532, 30, 10, 15, 67, 16}},
	// Version 40
	{{3706, 30, 19, 118, 6, 119}, {3706, 28, 18, 47, 31, 48}, {3706, 30, 34, 24, 34, 25}, {3706, 30, 20, 15, 61, 16}},
}

// alignmentPositions lists the row/column coordinates for alignment patterns
// per version. Version 1 has no alignment patterns.
var alignmentPositions = [41][]int{
	{},                             // version 0 (unused)
	{},                             // version 1
	{6, 18},                        // version 2
	{6, 22},                        // version 3
	{6, 26},                        // version 4
	{6, 30},                        // version 5
	{6, 34},                        // version 6
	{6, 22, 38},                    // version 7
	{6, 24, 42},                    // version 8
	{6, 26, 46},                    // version 9
	{6, 28, 50},                    // version 10
	{6, 30, 54},                    // version 11
	{6, 32, 58},                    // version 12
	{6, 34, 62},                    // version 13
	{6, 26, 46, 66},                // version 14
	{6, 26, 48, 70},                // version 15
	{6, 26, 50, 74},                // version 16
	{6, 30, 54, 78},                // version 17
	{6, 30, 56, 82},                // version 18
	{6, 30, 58, 86},                // version 19
	{6, 34, 62, 90},                // version 20
	{6, 28, 50, 72, 94},            // version 21
	{6, 26, 50, 74, 98},            // version 22
	{6, 30, 54, 78, 102},           // version 23
	{6, 28, 54, 80, 106},           // version 24
	{6, 32, 58, 84, 110},           // version 25
	{6, 30, 58, 86, 114},           // version 26
	{6, 34, 62, 90, 118},           // version 27
	{6, 26, 50, 74, 98, 122},       // version 28
	{6, 30, 54, 78, 102, 126},      // version 29
	{6, 26, 52, 78, 104, 130},      // version 30
	{6, 30, 56, 82, 108, 134},      // version 31
	{6, 34, 60, 86, 112, 138},      // version 32
	{6, 30, 58, 86, 114, 142},      // version 33
	{6, 34, 62, 90, 118, 146},      // version 34
	{6, 30, 54, 78, 102, 126, 150}, // version 35
	{6, 24, 50, 76, 102, 128, 154}, // version 36
	{6, 28, 54, 80, 106, 132, 158}, // version 37
	{6, 32, 58, 84, 110, 136, 162}, // version 38
	{6, 26, 54, 82, 110, 138, 166}, // version 39
	{6, 30, 58, 86, 114, 142, 170}, // version 40
}

// formatInfo contains the 15-bit format information strings for each
// EC level and mask combination, pre-masked with 0x5412.
// Index: [ecLevel][mask]
var formatInfo = [4][8]int{
	// Level L
	{0x77c4, 0x72f3, 0x7daa, 0x789d, 0x662f, 0x6318, 0x6c41, 0x6976},
	// Level M
	{0x5412, 0x5125, 0x5e7c, 0x5b4b, 0x45f9, 0x40ce, 0x4f97, 0x4aa0},
	// Level Q
	{0x355f, 0x3068, 0x3f31, 0x3a06, 0x24b4, 0x2183, 0x2eda, 0x2bed},
	// Level H
	{0x1689, 0x13be, 0x1ce7, 0x19d0, 0x0762, 0x0255, 0x0d0c, 0x083b},
}

// versionInfo contains the 18-bit version information strings for
// versions 7-40. Index: version - 7.
var versionInfo = [34]int{
	0x07c94, 0x085bc, 0x09a99, 0x0a4d3, 0x0bbf6, 0x0c762, 0x0d847, 0x0e60d,
	0x0f928, 0x10b78, 0x1145d, 0x12a17, 0x13532, 0x149a6, 0x15683, 0x168c9,
	0x177ec, 0x18ec4, 0x191e1, 0x1afab, 0x1b08e, 0x1cc1a, 0x1d33f, 0x1ed75,
	0x1f250, 0x209d5, 0x216f0, 0x228ba, 0x2379f, 0x24b0b, 0x2542e, 0x26a64,
	0x27541, 0x28c69,
}

// moduleSize returns the number of modules per side for a given version.
func moduleSize(version int) int {
	return 17 + version*4
}
