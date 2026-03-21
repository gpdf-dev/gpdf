package font

import (
	"encoding/binary"
	"testing"
)

// ---------------------------------------------------------------------------
// parseCmapFormat12 tests
// ---------------------------------------------------------------------------

func TestParseCmapFormat12_Basic(t *testing.T) {
	// Build a format 12 subtable with 2 groups.
	// Header: format(2) + reserved(2) + length(4) + language(4) + numGroups(4) = 16 bytes
	// Each group: startCharCode(4) + endCharCode(4) + startGlyphID(4) = 12 bytes
	numGroups := 2
	dataLen := 16 + numGroups*12
	tbl := make([]byte, dataLen)

	binary.BigEndian.PutUint16(tbl[0:2], 12)                  // format
	binary.BigEndian.PutUint16(tbl[2:4], 0)                   // reserved
	binary.BigEndian.PutUint32(tbl[4:8], uint32(dataLen))     // length
	binary.BigEndian.PutUint32(tbl[8:12], 0)                  // language
	binary.BigEndian.PutUint32(tbl[12:16], uint32(numGroups)) // numGroups

	// Group 0: U+0041 ('A') to U+0043 ('C'), startGlyphID = 1
	off := 16
	binary.BigEndian.PutUint32(tbl[off:off+4], 0x0041)
	binary.BigEndian.PutUint32(tbl[off+4:off+8], 0x0043)
	binary.BigEndian.PutUint32(tbl[off+8:off+12], 1)

	// Group 1: U+1F600 to U+1F601, startGlyphID = 100
	off = 16 + 12
	binary.BigEndian.PutUint32(tbl[off:off+4], 0x1F600)
	binary.BigEndian.PutUint32(tbl[off+4:off+8], 0x1F601)
	binary.BigEndian.PutUint32(tbl[off+8:off+12], 100)

	f12, err := parseCmapFormat12(tbl, 0)
	if err != nil {
		t.Fatalf("parseCmapFormat12 failed: %v", err)
	}
	if len(f12.groups) != 2 {
		t.Fatalf("got %d groups, want 2", len(f12.groups))
	}
	if f12.groups[0].startCharCode != 0x0041 || f12.groups[0].endCharCode != 0x0043 || f12.groups[0].startGlyphID != 1 {
		t.Errorf("group 0 = %+v, want A-C gid 1", f12.groups[0])
	}
	if f12.groups[1].startCharCode != 0x1F600 || f12.groups[1].startGlyphID != 100 {
		t.Errorf("group 1 = %+v, want emoji range gid 100", f12.groups[1])
	}
}

func TestParseCmapFormat12_TooShort(t *testing.T) {
	tbl := make([]byte, 10) // shorter than 16 byte header
	_, err := parseCmapFormat12(tbl, 0)
	if err == nil {
		t.Fatal("expected error for too-short format 12 data")
	}
}

func TestParseCmapFormat12_TruncatedGroups(t *testing.T) {
	// Header says 2 groups but data only has room for 1.
	tbl := make([]byte, 16+12) // room for 1 group
	binary.BigEndian.PutUint16(tbl[0:2], 12)
	binary.BigEndian.PutUint32(tbl[4:8], uint32(16+12*2)) // length for 2 groups
	binary.BigEndian.PutUint32(tbl[12:16], 2)             // numGroups = 2

	// Fill group 0
	binary.BigEndian.PutUint32(tbl[16:20], 0x41)
	binary.BigEndian.PutUint32(tbl[20:24], 0x41)
	binary.BigEndian.PutUint32(tbl[24:28], 1)

	_, err := parseCmapFormat12(tbl, 0)
	if err == nil {
		t.Fatal("expected error for truncated format 12 groups")
	}
}

func TestParseCmapFormat12_ZeroGroups(t *testing.T) {
	tbl := make([]byte, 16)
	binary.BigEndian.PutUint16(tbl[0:2], 12)
	binary.BigEndian.PutUint32(tbl[4:8], 16)
	binary.BigEndian.PutUint32(tbl[12:16], 0) // 0 groups

	f12, err := parseCmapFormat12(tbl, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(f12.groups) != 0 {
		t.Errorf("got %d groups, want 0", len(f12.groups))
	}
}

// ---------------------------------------------------------------------------
// parseOS2 tests
// ---------------------------------------------------------------------------

func TestParseOS2_ShortTable(t *testing.T) {
	// Table < 72 bytes should be skipped entirely.
	ttf := &TrueTypeFont{}
	data := make([]byte, 200)
	binary.BigEndian.PutUint32(data[0:4], 0x00010000)
	binary.BigEndian.PutUint16(data[4:6], 1) // 1 table

	// Table directory at offset 12
	copy(data[12:16], []byte("OS/2"))
	binary.BigEndian.PutUint32(data[16:20], 0)  // checksum
	binary.BigEndian.PutUint32(data[20:24], 28) // offset = 28
	binary.BigEndian.PutUint32(data[24:28], 60) // length = 60 (< 72)

	tables := map[string]tableRecord{
		tagOS2: {Tag: [4]byte{'O', 'S', '/', '2'}, Offset: 28, Length: 60},
	}
	ttf.parseOS2(data, tables)
	if ttf.metrics.CapHeight != 0 || ttf.metrics.XHeight != 0 {
		t.Error("expected no metrics from short OS/2 table")
	}
}

func TestParseOS2_XHeightOnly(t *testing.T) {
	// Table of 88 bytes: has xHeight (offset 86-88) but no capHeight (offset 88-90).
	os2 := make([]byte, 88)
	putI16(os2[86:88], 500) // xHeight

	data := make([]byte, 200)
	copy(data[100:], os2)

	tables := map[string]tableRecord{
		tagOS2: {Tag: [4]byte{'O', 'S', '/', '2'}, Offset: 100, Length: 88},
	}
	ttf := &TrueTypeFont{}
	ttf.parseOS2(data, tables)

	if ttf.metrics.XHeight != 500 {
		t.Errorf("XHeight = %d, want 500", ttf.metrics.XHeight)
	}
	if ttf.metrics.CapHeight != 0 {
		t.Errorf("CapHeight = %d, want 0 (table too short)", ttf.metrics.CapHeight)
	}
}

func TestParseOS2_FullMetrics(t *testing.T) {
	// Table of 90+ bytes: has both xHeight and capHeight.
	os2 := make([]byte, 92)
	putI16(os2[86:88], 450) // xHeight
	putI16(os2[88:90], 700) // capHeight

	data := make([]byte, 200)
	copy(data[100:], os2)

	tables := map[string]tableRecord{
		tagOS2: {Tag: [4]byte{'O', 'S', '/', '2'}, Offset: 100, Length: 92},
	}
	ttf := &TrueTypeFont{}
	ttf.parseOS2(data, tables)

	if ttf.metrics.XHeight != 450 {
		t.Errorf("XHeight = %d, want 450", ttf.metrics.XHeight)
	}
	if ttf.metrics.CapHeight != 700 {
		t.Errorf("CapHeight = %d, want 700", ttf.metrics.CapHeight)
	}
}

func TestParseOS2_MissingTable(t *testing.T) {
	ttf := &TrueTypeFont{}
	tables := map[string]tableRecord{} // no OS/2
	ttf.parseOS2(nil, tables)
	// Should not panic, metrics should remain zero.
	if ttf.metrics.CapHeight != 0 || ttf.metrics.XHeight != 0 {
		t.Error("expected zero metrics when OS/2 table is missing")
	}
}

// ---------------------------------------------------------------------------
// cmapFormat4 lookup with idRangeOffset != 0
// ---------------------------------------------------------------------------

func TestCmapFormat4_IdRangeOffset(t *testing.T) {
	// Build a cmapFormat4 with 2 segments + sentinel.
	// Segment 0: startCode=65 ('A'), endCode=67 ('C'), idRangeOffset != 0
	// Segment 1: sentinel (0xFFFF)
	f4 := &cmapFormat4{
		segCount:       2,
		endCodes:       []uint16{67, 0xFFFF},
		startCodes:     []uint16{65, 0xFFFF},
		idDeltas:       []int16{0, 1},
		idRangeOffsets: []uint16{4, 0}, // idRangeOffset[0] = 4
		glyphIDArray:   []uint16{10, 20, 30},
	}
	// The formula: glyphIdx = idRangeOffset[idx]/2 + (cp - startCode[idx]) - (segCount - idx)
	// For cp=65 (A), idx=0: glyphIdx = 4/2 + (65-65) - (2-0) = 2 + 0 - 2 = 0 => glyphIDArray[0] = 10
	// For cp=66 (B), idx=0: glyphIdx = 4/2 + (66-65) - (2-0) = 2 + 1 - 2 = 1 => glyphIDArray[1] = 20
	// For cp=67 (C), idx=0: glyphIdx = 4/2 + (67-65) - (2-0) = 2 + 2 - 2 = 2 => glyphIDArray[2] = 30

	tests := []struct {
		cp   uint16
		want uint16
	}{
		{65, 10}, // 'A' -> glyphIDArray[0] + idDelta[0] = 10 + 0
		{66, 20}, // 'B' -> glyphIDArray[1] + idDelta[0] = 20 + 0
		{67, 30}, // 'C' -> glyphIDArray[2] + idDelta[0] = 30 + 0
	}
	for _, tc := range tests {
		got := f4.lookup(tc.cp)
		if got != tc.want {
			t.Errorf("lookup(%d) = %d, want %d", tc.cp, got, tc.want)
		}
	}
}

func TestCmapFormat4_IdRangeOffset_OutOfBounds(t *testing.T) {
	// glyphIdx calculation yields index beyond glyphIDArray.
	f4 := &cmapFormat4{
		segCount:       2,
		endCodes:       []uint16{70, 0xFFFF},
		startCodes:     []uint16{65, 0xFFFF},
		idDeltas:       []int16{0, 1},
		idRangeOffsets: []uint16{100, 0}, // large offset, will go out of bounds
		glyphIDArray:   []uint16{10},     // only 1 element
	}
	got := f4.lookup(65)
	if got != 0 {
		t.Errorf("lookup(65) = %d, want 0 for out-of-bounds", got)
	}
}

func TestCmapFormat4_IdRangeOffset_GlyphIDZero(t *testing.T) {
	// glyphIDArray entry is 0, should return 0.
	f4 := &cmapFormat4{
		segCount:       2,
		endCodes:       []uint16{65, 0xFFFF},
		startCodes:     []uint16{65, 0xFFFF},
		idDeltas:       []int16{5, 1},
		idRangeOffsets: []uint16{4, 0},
		glyphIDArray:   []uint16{0}, // glyph ID is 0
	}
	// glyphIdx = 4/2 + (65-65) - (2-0) = 2 + 0 - 2 = 0 => glyphIDArray[0] = 0
	got := f4.lookup(65)
	if got != 0 {
		t.Errorf("lookup(65) = %d, want 0 for zero glyph ID in array", got)
	}
}

func TestCmapFormat4_IdRangeOffset_NegativeIndex(t *testing.T) {
	// Setup where glyphIdx calculation yields a negative index.
	f4 := &cmapFormat4{
		segCount:       2,
		endCodes:       []uint16{65, 0xFFFF},
		startCodes:     []uint16{65, 0xFFFF},
		idDeltas:       []int16{0, 1},
		idRangeOffsets: []uint16{2, 0}, // small offset
		glyphIDArray:   []uint16{10, 20, 30},
	}
	// glyphIdx = 2/2 + (65-65) - (2-0) = 1 + 0 - 2 = -1 => out of bounds
	got := f4.lookup(65)
	if got != 0 {
		t.Errorf("lookup(65) = %d, want 0 for negative index", got)
	}
}

func TestCmapFormat4_CpBelowStartCode(t *testing.T) {
	// cp falls within a segment's endCode range but below startCode.
	f4 := &cmapFormat4{
		segCount:       2,
		endCodes:       []uint16{70, 0xFFFF},
		startCodes:     []uint16{68, 0xFFFF},
		idDeltas:       []int16{0, 1},
		idRangeOffsets: []uint16{0, 0},
		glyphIDArray:   nil,
	}
	// cp=65 < startCodes[0]=68, so segment doesn't match.
	got := f4.lookup(65)
	if got != 0 {
		t.Errorf("lookup(65) = %d, want 0", got)
	}
}

func TestCmapFormat4_CpBeyondAllSegments(t *testing.T) {
	f4 := &cmapFormat4{
		segCount:       1,
		endCodes:       []uint16{50},
		startCodes:     []uint16{40},
		idDeltas:       []int16{0},
		idRangeOffsets: []uint16{0},
		glyphIDArray:   nil,
	}
	// cp=100 > all endCodes
	got := f4.lookup(100)
	if got != 0 {
		t.Errorf("lookup(100) = %d, want 0", got)
	}
}

// ---------------------------------------------------------------------------
// addCompositeComponents tests
// ---------------------------------------------------------------------------

func TestAddCompositeComponents_WithComposite(t *testing.T) {
	// Build a minimal font with a composite glyph that references component glyph IDs.
	numGlyphs := 3

	// Glyph 0: simple (numContours = 1), 12 bytes
	// Glyph 1: composite (numContours = -1), references glyph 2
	//   header: 10 bytes + flags(2) + gid(2) + args(4) = 18 bytes
	// Glyph 2: simple (numContours = 1), 12 bytes

	glyf := make([]byte, 12+18+12) // 42 bytes total

	// Glyph 0: simple at offset 0
	binary.BigEndian.PutUint16(glyf[0:2], 1) // numContours = 1

	// Glyph 1: composite at offset 12
	off := 12
	binary.BigEndian.PutUint16(glyf[off:off+2], uint16(0xFFFF)) // numContours = -1 (composite)
	// Glyph header: 10 bytes (numContours + xMin + yMin + xMax + yMax)
	// Composite data starts at offset 22
	compOff := off + 10
	flags := uint16(0x0001) // ARG_1_AND_2_ARE_WORDS, no MORE_COMPONENTS (bit 5 = 0)
	binary.BigEndian.PutUint16(glyf[compOff:compOff+2], flags)
	binary.BigEndian.PutUint16(glyf[compOff+2:compOff+4], 2) // component glyph ID = 2
	// arg1, arg2: 4 bytes (since ARG_1_AND_2_ARE_WORDS) — already zero

	// Glyph 2: simple at offset 30
	binary.BigEndian.PutUint16(glyf[30:32], 1) // numContours = 1

	// loca table (long format)
	loca := make([]byte, (numGlyphs+1)*4)
	binary.BigEndian.PutUint32(loca[0:4], 0)    // glyph 0 start
	binary.BigEndian.PutUint32(loca[4:8], 12)   // glyph 1 start
	binary.BigEndian.PutUint32(loca[8:12], 30)  // glyph 2 start
	binary.BigEndian.PutUint32(loca[12:16], 42) // end

	// head table (need indexToLocFormat at offset 50)
	head := make([]byte, 54)
	putI16(head[50:52], 1) // long format

	// Assemble data: head at 0, loca at 54, glyf at 54+locaLen
	locaLen := len(loca)
	glyfOff := 54 + locaLen
	totalLen := glyfOff + len(glyf)
	data := make([]byte, totalLen)
	copy(data[0:54], head)
	copy(data[54:54+locaLen], loca)
	copy(data[glyfOff:], glyf)

	glyfRec := subsetTableRecord{tag: tagGlyf, offset: uint32(glyfOff), length: uint32(len(glyf))}
	locaRec := subsetTableRecord{tag: tagLoca, offset: 54, length: uint32(locaLen)}
	headRec := subsetTableRecord{tag: tagHead, offset: 0, length: 54}

	keep := map[uint16]bool{1: true}
	addCompositeComponents(data, glyfRec, locaRec, headRec, keep)

	if !keep[2] {
		t.Error("expected glyph 2 to be added as composite component")
	}
}

func TestAddCompositeComponents_SimpleOnly(t *testing.T) {
	// All glyphs are simple, no composite expansion needed.
	numGlyphs := 2
	glyphSize := 12

	glyf := make([]byte, numGlyphs*glyphSize)
	binary.BigEndian.PutUint16(glyf[0:2], 1)   // glyph 0: simple
	binary.BigEndian.PutUint16(glyf[12:14], 1) // glyph 1: simple

	loca := make([]byte, (numGlyphs+1)*4)
	binary.BigEndian.PutUint32(loca[0:4], 0)
	binary.BigEndian.PutUint32(loca[4:8], 12)
	binary.BigEndian.PutUint32(loca[8:12], 24)

	head := make([]byte, 54)
	putI16(head[50:52], 1)

	locaLen := len(loca)
	glyfOff := 54 + locaLen
	data := make([]byte, glyfOff+len(glyf))
	copy(data[0:54], head)
	copy(data[54:54+locaLen], loca)
	copy(data[glyfOff:], glyf)

	glyfRec := subsetTableRecord{tag: tagGlyf, offset: uint32(glyfOff), length: uint32(len(glyf))}
	locaRec := subsetTableRecord{tag: tagLoca, offset: 54, length: uint32(locaLen)}
	headRec := subsetTableRecord{tag: tagHead, offset: 0, length: 54}

	keep := map[uint16]bool{0: true, 1: true}
	addCompositeComponents(data, glyfRec, locaRec, headRec, keep)

	if len(keep) != 2 {
		t.Errorf("expected 2 glyphs in keep set, got %d", len(keep))
	}
}

// ---------------------------------------------------------------------------
// parseCmap tests — format 12 only, unsupported format
// ---------------------------------------------------------------------------

func TestParseCmap_Format12Only(t *testing.T) {
	// Build a cmap table with only a format 12 subtable (no format 4).
	numGroups := 1
	f12Len := 16 + numGroups*12
	f12 := make([]byte, f12Len)
	binary.BigEndian.PutUint16(f12[0:2], 12)                  // format
	binary.BigEndian.PutUint32(f12[4:8], uint32(f12Len))      // length
	binary.BigEndian.PutUint32(f12[12:16], uint32(numGroups)) // numGroups
	binary.BigEndian.PutUint32(f12[16:20], 0x41)              // startCharCode
	binary.BigEndian.PutUint32(f12[20:24], 0x43)              // endCharCode
	binary.BigEndian.PutUint32(f12[24:28], 1)                 // startGlyphID

	// cmap header: version(2) + numSubtables(2) = 4 bytes
	// Subtable record: platformID(2) + encodingID(2) + offset(4) = 8 bytes
	cmapHeader := make([]byte, 4+8)
	binary.BigEndian.PutUint16(cmapHeader[0:2], 0)                        // version
	binary.BigEndian.PutUint16(cmapHeader[2:4], 1)                        // numSubtables
	binary.BigEndian.PutUint16(cmapHeader[4:6], 3)                        // platformID = 3 (Windows)
	binary.BigEndian.PutUint16(cmapHeader[6:8], 10)                       // encodingID = 10 (Unicode full)
	binary.BigEndian.PutUint32(cmapHeader[8:12], uint32(len(cmapHeader))) // offset

	cmapData := append(cmapHeader, f12...)

	// Build TTF with this cmap
	head := make([]byte, 54)
	binary.BigEndian.PutUint32(head[0:4], 0x00010000)
	binary.BigEndian.PutUint16(head[18:20], 1000)

	hhea := make([]byte, 36)
	binary.BigEndian.PutUint32(hhea[0:4], 0x00010000)
	putI16(hhea[4:6], 800)
	putI16(hhea[6:8], -200)
	binary.BigEndian.PutUint16(hhea[34:36], 3) // numberOfHMetrics

	maxp := make([]byte, 6)
	binary.BigEndian.PutUint32(maxp[0:4], 0x00010000)
	binary.BigEndian.PutUint16(maxp[4:6], 3)

	hmtx := make([]byte, 3*4)
	for i := 0; i < 3; i++ {
		binary.BigEndian.PutUint16(hmtx[i*4:i*4+2], 500)
	}

	nameTable := buildNameTableForTest("TestFont")
	post := make([]byte, 32)
	binary.BigEndian.PutUint32(post[0:4], 0x00030000)

	tables := []tableEntry{
		{tagHead, head},
		{tagHhea, hhea},
		{tagMaxp, maxp},
		{tagHmtx, hmtx},
		{tagCmap, cmapData},
		{tagName, nameTable},
		{tagPost, post},
	}
	data := assembleTTF(tables)

	ttf, err := ParseTrueType(data)
	if err != nil {
		t.Fatalf("ParseTrueType with format 12 only failed: %v", err)
	}

	// Verify lookup works through format 12.
	gid := ttf.GlyphID('A')
	if gid != 1 {
		t.Errorf("GlyphID('A') = %d, want 1", gid)
	}
	gid = ttf.GlyphID('C')
	if gid != 3 {
		t.Errorf("GlyphID('C') = %d, want 3 (via format 12)", gid)
	}
}

func TestParseCmap_UnsupportedFormatOnly(t *testing.T) {
	// Build a cmap with only an unsupported format (e.g., format 6).
	f6 := make([]byte, 20)
	binary.BigEndian.PutUint16(f6[0:2], 6) // format 6
	binary.BigEndian.PutUint16(f6[2:4], 20)

	cmapHeader := make([]byte, 4+8)
	binary.BigEndian.PutUint16(cmapHeader[0:2], 0)
	binary.BigEndian.PutUint16(cmapHeader[2:4], 1)
	binary.BigEndian.PutUint16(cmapHeader[4:6], 3)
	binary.BigEndian.PutUint16(cmapHeader[6:8], 1)
	binary.BigEndian.PutUint32(cmapHeader[8:12], uint32(len(cmapHeader)))

	cmapData := append(cmapHeader, f6...)

	head := make([]byte, 54)
	binary.BigEndian.PutUint32(head[0:4], 0x00010000)
	binary.BigEndian.PutUint16(head[18:20], 1000)

	hhea := make([]byte, 36)
	binary.BigEndian.PutUint32(hhea[0:4], 0x00010000)
	putI16(hhea[4:6], 800)
	putI16(hhea[6:8], -200)
	binary.BigEndian.PutUint16(hhea[34:36], 1)

	maxp := make([]byte, 6)
	binary.BigEndian.PutUint32(maxp[0:4], 0x00010000)
	binary.BigEndian.PutUint16(maxp[4:6], 1)

	hmtx := make([]byte, 4)
	binary.BigEndian.PutUint16(hmtx[0:2], 500)

	nameTable := buildNameTableForTest("TestFont")
	post := make([]byte, 32)
	binary.BigEndian.PutUint32(post[0:4], 0x00030000)

	tables := []tableEntry{
		{tagHead, head},
		{tagHhea, hhea},
		{tagMaxp, maxp},
		{tagHmtx, hmtx},
		{tagCmap, cmapData},
		{tagName, nameTable},
		{tagPost, post},
	}
	data := assembleTTF(tables)

	_, err := ParseTrueType(data)
	if err == nil {
		t.Fatal("expected error for cmap with only unsupported format")
	}
}

// ---------------------------------------------------------------------------
// GlyphWidth / GlyphID bounds tests
// ---------------------------------------------------------------------------

func TestGlyphWidth_BeyondHmtxBounds(t *testing.T) {
	// Build a font where cmap maps to a glyph ID beyond hmtx array.
	// We use format 12 to map U+0041 to glyph 100, but only have 2 glyphs in hmtx.
	numGroups := 1
	f12Len := 16 + numGroups*12
	f12 := make([]byte, f12Len)
	binary.BigEndian.PutUint16(f12[0:2], 12)
	binary.BigEndian.PutUint32(f12[4:8], uint32(f12Len))
	binary.BigEndian.PutUint32(f12[12:16], uint32(numGroups))
	binary.BigEndian.PutUint32(f12[16:20], 0x41) // startCharCode = A
	binary.BigEndian.PutUint32(f12[20:24], 0x41) // endCharCode = A
	binary.BigEndian.PutUint32(f12[24:28], 100)  // startGlyphID = 100

	cmapHeader := make([]byte, 12)
	binary.BigEndian.PutUint16(cmapHeader[0:2], 0)
	binary.BigEndian.PutUint16(cmapHeader[2:4], 1)
	binary.BigEndian.PutUint16(cmapHeader[4:6], 3)
	binary.BigEndian.PutUint16(cmapHeader[6:8], 10)
	binary.BigEndian.PutUint32(cmapHeader[8:12], 12)
	cmapData := append(cmapHeader, f12...)

	head := make([]byte, 54)
	binary.BigEndian.PutUint32(head[0:4], 0x00010000)
	binary.BigEndian.PutUint16(head[18:20], 1000)

	hhea := make([]byte, 36)
	binary.BigEndian.PutUint32(hhea[0:4], 0x00010000)
	putI16(hhea[4:6], 800)
	putI16(hhea[6:8], -200)
	binary.BigEndian.PutUint16(hhea[34:36], 2)

	maxp := make([]byte, 6)
	binary.BigEndian.PutUint32(maxp[0:4], 0x00010000)
	binary.BigEndian.PutUint16(maxp[4:6], 2) // only 2 glyphs

	hmtx := make([]byte, 8)
	binary.BigEndian.PutUint16(hmtx[0:2], 0)
	binary.BigEndian.PutUint16(hmtx[4:6], 600)

	nameTable := buildNameTableForTest("TestFont")
	post := make([]byte, 32)
	binary.BigEndian.PutUint32(post[0:4], 0x00030000)

	tables := []tableEntry{
		{tagHead, head},
		{tagHhea, hhea},
		{tagMaxp, maxp},
		{tagHmtx, hmtx},
		{tagCmap, cmapData},
		{tagName, nameTable},
		{tagPost, post},
	}
	data := assembleTTF(tables)

	ttf, err := ParseTrueType(data)
	if err != nil {
		t.Fatalf("ParseTrueType failed: %v", err)
	}

	// GlyphWidth for 'A' maps to glyph 100, but hmtx only has 2 entries.
	w, ok := ttf.GlyphWidth('A')
	if ok {
		t.Errorf("GlyphWidth('A') ok = true, want false (glyph ID out of bounds)")
	}
	if w != 0 {
		t.Errorf("GlyphWidth('A') = %d, want 0", w)
	}
}

// ---------------------------------------------------------------------------
// cmapTable lookup tests — format 12 preferred over format 4
// ---------------------------------------------------------------------------

func TestCmapTable_Format12Preferred(t *testing.T) {
	f4 := &cmapFormat4{
		segCount:       2,
		endCodes:       []uint16{0x41, 0xFFFF},
		startCodes:     []uint16{0x41, 0xFFFF},
		idDeltas:       []int16{10, 1},
		idRangeOffsets: []uint16{0, 0},
	}
	f12 := &cmapFormat12{
		groups: []cmapFormat12Group{
			{startCharCode: 0x41, endCharCode: 0x41, startGlyphID: 99},
		},
	}
	cmap := &cmapTable{format4: f4, format12: f12}

	gid, ok := cmap.lookup('A')
	if !ok {
		t.Fatal("lookup('A') failed")
	}
	// Format 12 should be preferred, giving glyph 99.
	if gid != 99 {
		t.Errorf("lookup('A') = %d, want 99 (format 12 preferred)", gid)
	}
}

func TestCmapTable_Format12NotFound_FallsBackToFormat4(t *testing.T) {
	f4 := &cmapFormat4{
		segCount:       2,
		endCodes:       []uint16{0x42, 0xFFFF},
		startCodes:     []uint16{0x42, 0xFFFF},
		idDeltas:       []int16{5, 1},
		idRangeOffsets: []uint16{0, 0},
	}
	f12 := &cmapFormat12{
		groups: []cmapFormat12Group{
			{startCharCode: 0x41, endCharCode: 0x41, startGlyphID: 99},
		},
	}
	cmap := &cmapTable{format4: f4, format12: f12}

	// 'B' (0x42) is not in format 12 range, but is in format 4.
	gid, ok := cmap.lookup('B')
	if !ok {
		t.Fatal("lookup('B') failed")
	}
	// format 4: uint16(int16(0x42) + 5) = 71
	if gid != 71 {
		t.Errorf("lookup('B') = %d, want 71 (fallback to format 4)", gid)
	}
}

func TestCmapTable_NeitherFormat(t *testing.T) {
	cmap := &cmapTable{} // no format4, no format12
	_, ok := cmap.lookup('A')
	if ok {
		t.Error("expected lookup to fail with no cmap tables")
	}
}
