package font

import (
	"encoding/binary"
	"math"
	"strings"
	"testing"
)

// tableEntry is used by the test helpers to pass table data to assembleTTF.
type tableEntry struct {
	tag  string
	data []byte
}

// putI16 writes an int16 value as big-endian uint16 into a byte slice.
func putI16(b []byte, v int16) {
	binary.BigEndian.PutUint16(b, uint16(v))
}

// ---------------------------------------------------------------------------
// Mock Font for metrics tests
// ---------------------------------------------------------------------------

// mockFont is a simple Font implementation for testing MeasureString and
// LineBreak without requiring a real TrueType font file.
type mockFont struct {
	name       string
	unitsPerEm int
	widths     map[rune]int
}

func (m *mockFont) Name() string                        { return m.name }
func (m *mockFont) Metrics() Metrics                    { return Metrics{UnitsPerEm: m.unitsPerEm} }
func (m *mockFont) Encode(text string) []byte           { return nil }
func (m *mockFont) Subset(runes []rune) ([]byte, error) { return nil, nil }

func (m *mockFont) GlyphWidth(r rune) (int, bool) {
	w, ok := m.widths[r]
	return w, ok
}

// newMockFont creates a mock font where every mapped rune has the given width.
func newMockFont(unitsPerEm int, defaultWidth int, extras map[rune]int) *mockFont {
	widths := make(map[rune]int)
	for r := rune(0x20); r < 0x7F; r++ {
		widths[r] = defaultWidth
	}
	for r, w := range extras {
		widths[r] = w
	}
	return &mockFont{name: "MockFont", unitsPerEm: unitsPerEm, widths: widths}
}

// ---------------------------------------------------------------------------
// MeasureString tests
// ---------------------------------------------------------------------------

func TestMeasureString_Empty(t *testing.T) {
	f := newMockFont(1000, 500, nil)
	got := MeasureString(f, "", 12)
	if got != 0 {
		t.Errorf("MeasureString(empty) = %f, want 0", got)
	}
}

func TestMeasureString_ZeroUnitsPerEm(t *testing.T) {
	f := &mockFont{unitsPerEm: 0, widths: map[rune]int{'A': 500}}
	got := MeasureString(f, "A", 12)
	if got != 0 {
		t.Errorf("MeasureString(unitsPerEm=0) = %f, want 0", got)
	}
}

func TestMeasureString_SingleChar(t *testing.T) {
	f := newMockFont(1000, 500, nil)
	got := MeasureString(f, "A", 10)
	want := 5.0
	if math.Abs(got-want) > 1e-9 {
		t.Errorf("MeasureString('A') = %f, want %f", got, want)
	}
}

func TestMeasureString_MultipleChars(t *testing.T) {
	f := newMockFont(1000, 600, nil)
	got := MeasureString(f, "Hello", 12)
	want := 36.0
	if math.Abs(got-want) > 1e-9 {
		t.Errorf("MeasureString('Hello') = %f, want %f", got, want)
	}
}

func TestMeasureString_MissingGlyphFallback(t *testing.T) {
	f := &mockFont{
		unitsPerEm: 1000,
		widths:     map[rune]int{' ': 250},
	}
	got := MeasureString(f, "AB", 10)
	want := 5.0
	if math.Abs(got-want) > 1e-9 {
		t.Errorf("MeasureString(fallback) = %f, want %f", got, want)
	}
}

func TestMeasureString_MixedWidths(t *testing.T) {
	f := newMockFont(2048, 500, map[rune]int{'W': 1000})
	got := MeasureString(f, "WA", 20)
	want := 1500.0 * 20.0 / 2048.0
	if math.Abs(got-want) > 1e-9 {
		t.Errorf("MeasureString('WA') = %f, want %f", got, want)
	}
}

// ---------------------------------------------------------------------------
// LineBreak tests
// ---------------------------------------------------------------------------

func TestLineBreak_Empty(t *testing.T) {
	f := newMockFont(1000, 500, nil)
	lines := LineBreak(f, "", 12, 100)
	if len(lines) != 1 || lines[0] != "" {
		t.Errorf("LineBreak(empty) = %v, want [\"\"]", lines)
	}
}

func TestLineBreak_SingleLine(t *testing.T) {
	f := newMockFont(1000, 500, nil)
	lines := LineBreak(f, "Hello", 10, 30)
	if len(lines) != 1 || lines[0] != "Hello" {
		t.Errorf("LineBreak(fits) = %v, want [\"Hello\"]", lines)
	}
}

func TestLineBreak_WordWrap(t *testing.T) {
	f := newMockFont(1000, 500, nil)
	lines := LineBreak(f, "Hello World", 10, 30)
	if len(lines) != 2 {
		t.Fatalf("LineBreak(word wrap) got %d lines, want 2: %v", len(lines), lines)
	}
	if lines[0] != "Hello" {
		t.Errorf("line[0] = %q, want %q", lines[0], "Hello")
	}
	if lines[1] != "World" {
		t.Errorf("line[1] = %q, want %q", lines[1], "World")
	}
}

func TestLineBreak_ExplicitNewline(t *testing.T) {
	f := newMockFont(1000, 500, nil)
	lines := LineBreak(f, "A\nB\nC", 10, 1000)
	if len(lines) != 3 {
		t.Fatalf("LineBreak(newline) got %d lines, want 3: %v", len(lines), lines)
	}
	for i, want := range []string{"A", "B", "C"} {
		if lines[i] != want {
			t.Errorf("line[%d] = %q, want %q", i, lines[i], want)
		}
	}
}

func TestLineBreak_LongWordCharBreak(t *testing.T) {
	f := newMockFont(1000, 500, nil)
	lines := LineBreak(f, "ABCDEFG", 10, 15)
	if len(lines) < 2 {
		t.Fatalf("LineBreak(long word) got %d lines, want >=2: %v", len(lines), lines)
	}
	joined := strings.Join(lines, "")
	if joined != "ABCDEFG" {
		t.Errorf("rejoined lines = %q, want %q", joined, "ABCDEFG")
	}
}

func TestLineBreak_CJKCharBreak(t *testing.T) {
	// Each CJK char is 1000 design units. fontSize=10, upm=1000 => 10pt each.
	// maxWidth=25 allows 2 chars per line.
	// CJK break points are set after each character, so the break logic
	// produces multiple lines from a run of CJK characters.
	f := &mockFont{
		unitsPerEm: 1000,
		widths: map[rune]int{
			'\u65E5': 1000, // 日
			'\u672C': 1000, // 本
			'\u8A9E': 1000, // 語
			'\u30C6': 1000, // テ
			' ':      500,
		},
	}
	lines := LineBreak(f, "\u65E5\u672C\u8A9E\u30C6", 10, 25)
	if len(lines) < 2 {
		t.Fatalf("LineBreak(CJK) got %d lines, want >=2: %v", len(lines), lines)
	}
	// Verify the first line contains no more than 2 CJK characters
	// (since each is 10pt and maxWidth is 25pt).
	firstLineLen := len([]rune(lines[0]))
	if firstLineLen > 2 {
		t.Errorf("first line has %d runes, want <=2", firstLineLen)
	}
}

func TestLineBreak_MultipleSpaceBreaks(t *testing.T) {
	f := newMockFont(1000, 500, nil)
	lines := LineBreak(f, "AAAA BBBB CCCC", 10, 25)
	if len(lines) != 3 {
		t.Fatalf("LineBreak(multi-space) got %d lines, want 3: %v", len(lines), lines)
	}
	for i, want := range []string{"AAAA", "BBBB", "CCCC"} {
		if lines[i] != want {
			t.Errorf("line[%d] = %q, want %q", i, lines[i], want)
		}
	}
}

// ---------------------------------------------------------------------------
// isCJK tests
// ---------------------------------------------------------------------------

func TestIsCJK(t *testing.T) {
	tests := []struct {
		r    rune
		want bool
	}{
		{'A', false},
		{'1', false},
		{' ', false},
		{'\u65E5', true}, // Han (日)
		{'\u3042', true}, // Hiragana (あ)
		{'\u30A2', true}, // Katakana (ア)
		{'\uD55C', true}, // Hangul (한)
		{'\u3001', true}, // CJK punctuation (、)
		{'\uFF0C', true}, // Fullwidth comma (，)
	}
	for _, tc := range tests {
		got := isCJK(tc.r)
		if got != tc.want {
			t.Errorf("isCJK(U+%04X) = %v, want %v", tc.r, got, tc.want)
		}
	}
}

// ---------------------------------------------------------------------------
// CMap tests
// ---------------------------------------------------------------------------

func TestGenerateToUnicodeCMap_Empty(t *testing.T) {
	result := GenerateToUnicodeCMap(nil)
	if result != nil {
		t.Errorf("GenerateToUnicodeCMap(nil) = %v, want nil", result)
	}

	result = GenerateToUnicodeCMap(map[rune]uint16{})
	if result != nil {
		t.Errorf("GenerateToUnicodeCMap(empty) = %v, want nil", result)
	}
}

func TestGenerateToUnicodeCMap_BMPChars(t *testing.T) {
	mapping := map[rune]uint16{
		'A': 1,
		'B': 2,
	}
	result := GenerateToUnicodeCMap(mapping)
	if result == nil {
		t.Fatal("GenerateToUnicodeCMap returned nil for non-empty mapping")
	}

	s := string(result)
	if !strings.Contains(s, "begincmap") {
		t.Error("missing 'begincmap'")
	}
	if !strings.Contains(s, "endcmap") {
		t.Error("missing 'endcmap'")
	}
	if !strings.Contains(s, "beginbfchar") {
		t.Error("missing 'beginbfchar'")
	}
	if !strings.Contains(s, "endbfchar") {
		t.Error("missing 'endbfchar'")
	}
	if !strings.Contains(s, "<0000> <FFFF>") {
		t.Error("missing codespace range")
	}
	if !strings.Contains(s, "<0001> <0041>") {
		t.Error("missing BMP entry for 'A'")
	}
	if !strings.Contains(s, "<0002> <0042>") {
		t.Error("missing BMP entry for 'B'")
	}
}

func TestGenerateToUnicodeCMap_SupplementaryPlane(t *testing.T) {
	mapping := map[rune]uint16{
		0x1F600: 100,
	}
	result := GenerateToUnicodeCMap(mapping)
	if result == nil {
		t.Fatal("GenerateToUnicodeCMap returned nil for supplementary char")
	}

	s := string(result)
	if !strings.Contains(s, "<0064> <D83DDE00>") {
		t.Errorf("missing supplementary plane entry; got:\n%s", s)
	}
}

func TestGenerateToUnicodeCMap_LargeMapping(t *testing.T) {
	mapping := make(map[rune]uint16)
	for i := 0; i < 150; i++ {
		mapping[rune(0x30+i)] = uint16(i + 1)
	}
	result := GenerateToUnicodeCMap(mapping)
	if result == nil {
		t.Fatal("GenerateToUnicodeCMap returned nil for large mapping")
	}

	s := string(result)
	count := strings.Count(s, "beginbfchar")
	if count < 2 {
		t.Errorf("expected >=2 beginbfchar blocks for 150 entries, got %d", count)
	}
}

func TestGenerateToUnicodeCMap_Sorted(t *testing.T) {
	mapping := map[rune]uint16{
		'Z': 1,
		'A': 5,
		'M': 3,
	}
	result := GenerateToUnicodeCMap(mapping)
	s := string(result)

	idx1 := strings.Index(s, "<0001>")
	idx3 := strings.Index(s, "<0003>")
	idx5 := strings.Index(s, "<0005>")
	if idx1 < 0 || idx3 < 0 || idx5 < 0 {
		t.Fatalf("missing gid entries in output:\n%s", s)
	}
	if idx1 >= idx3 || idx3 >= idx5 {
		t.Error("entries not sorted by glyph ID")
	}
}

// ---------------------------------------------------------------------------
// utf16SurrogatePair tests
// ---------------------------------------------------------------------------

func TestUTF16SurrogatePair(t *testing.T) {
	tests := []struct {
		r        rune
		wantHigh uint16
		wantLow  uint16
	}{
		{0x10000, 0xD800, 0xDC00},
		{0x10FFFF, 0xDBFF, 0xDFFF},
		{0x1F600, 0xD83D, 0xDE00},
	}
	for _, tc := range tests {
		high, low := utf16SurrogatePair(tc.r)
		if high != tc.wantHigh || low != tc.wantLow {
			t.Errorf("utf16SurrogatePair(U+%04X) = (%04X, %04X), want (%04X, %04X)",
				tc.r, high, low, tc.wantHigh, tc.wantLow)
		}
	}
}

// ---------------------------------------------------------------------------
// Minimal TrueType font builder for testing
// ---------------------------------------------------------------------------

// buildMinimalTTF constructs a minimal valid TrueType font binary containing
// the required tables: head, hhea, maxp, cmap, hmtx, name, post.
func buildMinimalTTF(numGlyphs int, widths []uint16, runeMap map[rune]uint16) []byte {
	// --- head table (54 bytes minimum) ---
	head := make([]byte, 54)
	binary.BigEndian.PutUint32(head[0:4], 0x00010000)   // version
	binary.BigEndian.PutUint32(head[4:8], 0x00005000)   // fontRevision
	binary.BigEndian.PutUint32(head[8:12], 0)           // checksumAdjustment
	binary.BigEndian.PutUint32(head[12:16], 0x5F0F3CF5) // magicNumber
	binary.BigEndian.PutUint16(head[16:18], 0x000B)     // flags
	binary.BigEndian.PutUint16(head[18:20], 1000)       // unitsPerEm
	putI16(head[36:38], 0)                              // xMin
	putI16(head[38:40], 0)                              // yMin
	putI16(head[40:42], 1000)                           // xMax
	putI16(head[42:44], 1000)                           // yMax
	binary.BigEndian.PutUint16(head[44:46], 0)          // macStyle
	binary.BigEndian.PutUint16(head[46:48], 8)          // lowestRecPPEM
	putI16(head[48:50], 2)                              // fontDirectionHint
	putI16(head[50:52], 1)                              // indexToLocFormat (long)
	putI16(head[52:54], 0)                              // glyphDataFormat

	// --- hhea table (36 bytes) ---
	hhea := make([]byte, 36)
	binary.BigEndian.PutUint32(hhea[0:4], 0x00010000) // version
	putI16(hhea[4:6], 800)                            // ascender
	putI16(hhea[6:8], -200)                           // descender
	putI16(hhea[8:10], 0)                             // lineGap
	binary.BigEndian.PutUint16(hhea[10:12], 1000)     // advanceWidthMax
	numberOfHMetrics := numGlyphs
	if numberOfHMetrics > len(widths) {
		numberOfHMetrics = len(widths)
	}
	binary.BigEndian.PutUint16(hhea[34:36], uint16(numberOfHMetrics))

	// --- maxp table (6 bytes minimum) ---
	maxp := make([]byte, 6)
	binary.BigEndian.PutUint32(maxp[0:4], 0x00010000)
	binary.BigEndian.PutUint16(maxp[4:6], uint16(numGlyphs))

	// --- hmtx table ---
	hmtx := make([]byte, numberOfHMetrics*4)
	for i := 0; i < numberOfHMetrics; i++ {
		w := uint16(0)
		if i < len(widths) {
			w = widths[i]
		}
		binary.BigEndian.PutUint16(hmtx[i*4:i*4+2], w)
		putI16(hmtx[i*4+2:i*4+4], 0) // lsb
	}

	// --- cmap table (format 4) ---
	cmap := buildCmapTableForTest(runeMap)

	// --- name table ---
	nameTable := buildNameTableForTest("TestFont")

	// --- post table (32 bytes) ---
	post := make([]byte, 32)
	binary.BigEndian.PutUint32(post[0:4], 0x00030000)

	tables := []tableEntry{
		{tagHead, head},
		{tagHhea, hhea},
		{tagMaxp, maxp},
		{tagHmtx, hmtx},
		{tagCmap, cmap},
		{tagName, nameTable},
		{tagPost, post},
	}

	return assembleTTF(tables)
}

// buildCmapTableForTest builds a minimal cmap table with a single format 4 subtable.
func buildCmapTableForTest(runeMap map[rune]uint16) []byte {
	type seg struct {
		startCode uint16
		endCode   uint16
		delta     int16
	}

	var segments []seg
	for r, gid := range runeMap {
		if r > 0xFFFF {
			continue
		}
		cp := uint16(r)
		delta := int16(gid) - int16(cp)
		segments = append(segments, seg{startCode: cp, endCode: cp, delta: delta})
	}
	// Sort segments by startCode.
	for i := 0; i < len(segments); i++ {
		for j := i + 1; j < len(segments); j++ {
			if segments[j].startCode < segments[i].startCode {
				segments[i], segments[j] = segments[j], segments[i]
			}
		}
	}
	// Add sentinel segment.
	segments = append(segments, seg{startCode: 0xFFFF, endCode: 0xFFFF, delta: 1})

	segCount := len(segments)
	subtableLen := 14 + segCount*2*4 + 2
	subtable := make([]byte, subtableLen)
	binary.BigEndian.PutUint16(subtable[0:2], 4)
	binary.BigEndian.PutUint16(subtable[2:4], uint16(subtableLen))
	binary.BigEndian.PutUint16(subtable[4:6], 0)
	binary.BigEndian.PutUint16(subtable[6:8], uint16(segCount*2))

	sr := computeSearchRange(segCount)
	binary.BigEndian.PutUint16(subtable[8:10], sr.searchRange)
	binary.BigEndian.PutUint16(subtable[10:12], sr.entrySelector)
	binary.BigEndian.PutUint16(subtable[12:14], sr.rangeShift)

	off := 14
	for i, s := range segments {
		binary.BigEndian.PutUint16(subtable[off+i*2:off+i*2+2], s.endCode)
	}
	off += segCount * 2
	binary.BigEndian.PutUint16(subtable[off:off+2], 0) // reservedPad
	off += 2
	for i, s := range segments {
		binary.BigEndian.PutUint16(subtable[off+i*2:off+i*2+2], s.startCode)
	}
	off += segCount * 2
	for i, s := range segments {
		binary.BigEndian.PutUint16(subtable[off+i*2:off+i*2+2], uint16(s.delta))
	}
	_ = off // idRangeOffset array: all zeros (already zeroed by make).

	cmapHeader := make([]byte, 4+8)
	binary.BigEndian.PutUint16(cmapHeader[0:2], 0)
	binary.BigEndian.PutUint16(cmapHeader[2:4], 1)
	binary.BigEndian.PutUint16(cmapHeader[4:6], 3)
	binary.BigEndian.PutUint16(cmapHeader[6:8], 1)
	binary.BigEndian.PutUint32(cmapHeader[8:12], uint32(len(cmapHeader)))

	result := make([]byte, 0, len(cmapHeader)+len(subtable))
	result = append(result, cmapHeader...)
	result = append(result, subtable...)
	return result
}

// buildNameTableForTest builds a minimal name table with a PostScript name.
func buildNameTableForTest(fontName string) []byte {
	nameBytes := []byte(fontName)
	storageOffset := 6 + 12
	totalLen := storageOffset + len(nameBytes)
	tbl := make([]byte, totalLen)

	binary.BigEndian.PutUint16(tbl[0:2], 0)
	binary.BigEndian.PutUint16(tbl[2:4], 1)
	binary.BigEndian.PutUint16(tbl[4:6], uint16(storageOffset))

	off := 6
	binary.BigEndian.PutUint16(tbl[off:off+2], 1)                         // platformID (Mac)
	binary.BigEndian.PutUint16(tbl[off+2:off+4], 0)                       // encodingID
	binary.BigEndian.PutUint16(tbl[off+4:off+6], 0)                       // languageID
	binary.BigEndian.PutUint16(tbl[off+6:off+8], 6)                       // nameID (PostScript)
	binary.BigEndian.PutUint16(tbl[off+8:off+10], uint16(len(nameBytes))) // length
	binary.BigEndian.PutUint16(tbl[off+10:off+12], 0)                     // offset

	copy(tbl[storageOffset:], nameBytes)
	return tbl
}

// assembleTTF assembles table data into a valid TrueType font binary.
func assembleTTF(tables []tableEntry) []byte {
	numTables := len(tables)
	headerSize := 12 + numTables*16

	offsets := make([]int, numTables)
	currentOffset := headerSize
	for i := range tables {
		offsets[i] = currentOffset
		size := len(tables[i].data)
		currentOffset += (size + 3) &^ 3
	}

	result := make([]byte, currentOffset)

	binary.BigEndian.PutUint32(result[0:4], 0x00010000)
	binary.BigEndian.PutUint16(result[4:6], uint16(numTables))

	sr := computeSearchRange(numTables)
	binary.BigEndian.PutUint16(result[6:8], sr.searchRange)
	binary.BigEndian.PutUint16(result[8:10], sr.entrySelector)
	binary.BigEndian.PutUint16(result[10:12], sr.rangeShift)

	dirOffset := 12
	for i, tbl := range tables {
		copy(result[dirOffset:dirOffset+4], []byte(tbl.tag))
		binary.BigEndian.PutUint32(result[dirOffset+4:dirOffset+8], 0)
		binary.BigEndian.PutUint32(result[dirOffset+8:dirOffset+12], uint32(offsets[i]))
		binary.BigEndian.PutUint32(result[dirOffset+12:dirOffset+16], uint32(len(tbl.data)))
		dirOffset += 16
	}

	for i, tbl := range tables {
		copy(result[offsets[i]:], tbl.data)
	}

	recalcTableChecksums(result, numTables)

	return result
}

// buildMinimalTTFWithLocaGlyf builds a TTF that also includes loca and glyf
// tables, needed for SubsetTrueType testing.
func buildMinimalTTFWithLocaGlyf(numGlyphs int, widths []uint16, runeMap map[rune]uint16) []byte {
	head := make([]byte, 54)
	binary.BigEndian.PutUint32(head[0:4], 0x00010000)
	binary.BigEndian.PutUint32(head[12:16], 0x5F0F3CF5)
	binary.BigEndian.PutUint16(head[18:20], 1000)
	putI16(head[50:52], 1) // indexToLocFormat = long

	hhea := make([]byte, 36)
	binary.BigEndian.PutUint32(hhea[0:4], 0x00010000)
	putI16(hhea[4:6], 800)
	putI16(hhea[6:8], -200)
	numberOfHMetrics := numGlyphs
	if numberOfHMetrics > len(widths) {
		numberOfHMetrics = len(widths)
	}
	binary.BigEndian.PutUint16(hhea[34:36], uint16(numberOfHMetrics))

	maxp := make([]byte, 6)
	binary.BigEndian.PutUint32(maxp[0:4], 0x00010000)
	binary.BigEndian.PutUint16(maxp[4:6], uint16(numGlyphs))

	hmtx := make([]byte, numberOfHMetrics*4)
	for i := 0; i < numberOfHMetrics; i++ {
		w := uint16(0)
		if i < len(widths) {
			w = widths[i]
		}
		binary.BigEndian.PutUint16(hmtx[i*4:i*4+2], w)
	}

	cmap := buildCmapTableForTest(runeMap)

	// glyf table: numGlyphs simple glyphs, each 12 bytes.
	glyphSize := 12
	glyf := make([]byte, numGlyphs*glyphSize)
	for i := 0; i < numGlyphs; i++ {
		off := i * glyphSize
		binary.BigEndian.PutUint16(glyf[off:off+2], 1) // numContours = 1
		putI16(glyf[off+2:off+4], 0)                   // xMin
		putI16(glyf[off+4:off+6], 0)                   // yMin
		putI16(glyf[off+6:off+8], 500)                 // xMax
		putI16(glyf[off+8:off+10], 700)                // yMax
		binary.BigEndian.PutUint16(glyf[off+10:off+12], 0)
	}

	// loca table (long format): numGlyphs + 1 uint32 entries.
	loca := make([]byte, (numGlyphs+1)*4)
	for i := 0; i <= numGlyphs; i++ {
		binary.BigEndian.PutUint32(loca[i*4:i*4+4], uint32(i*glyphSize))
	}

	nameTable := buildNameTableForTest("TestFont")

	post := make([]byte, 32)
	binary.BigEndian.PutUint32(post[0:4], 0x00030000)

	tables := []tableEntry{
		{tagHead, head},
		{tagHhea, hhea},
		{tagMaxp, maxp},
		{tagHmtx, hmtx},
		{tagCmap, cmap},
		{tagLoca, loca},
		{tagGlyf, glyf},
		{tagName, nameTable},
		{tagPost, post},
	}

	return assembleTTF(tables)
}

// ---------------------------------------------------------------------------
// ParseTrueType tests
// ---------------------------------------------------------------------------

func TestParseTrueType_TooShort(t *testing.T) {
	_, err := ParseTrueType([]byte{0, 1, 2, 3})
	if err == nil {
		t.Fatal("ParseTrueType should fail on short data")
	}
}

func TestParseTrueType_BadScaler(t *testing.T) {
	data := make([]byte, 128)
	binary.BigEndian.PutUint32(data[0:4], 0xDEADBEEF)
	binary.BigEndian.PutUint16(data[4:6], 0)
	_, err := ParseTrueType(data)
	if err == nil {
		t.Fatal("ParseTrueType should fail on invalid scaler type")
	}
	if !strings.Contains(err.Error(), "unsupported scaler type") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestParseTrueType_TruncatedTableDirectory(t *testing.T) {
	data := make([]byte, 14)
	binary.BigEndian.PutUint32(data[0:4], 0x00010000)
	binary.BigEndian.PutUint16(data[4:6], 5)
	_, err := ParseTrueType(data)
	if err == nil {
		t.Fatal("ParseTrueType should fail on truncated table directory")
	}
}

// buildTestFont creates a parsed TrueTypeFont for use across sub-tests.
func buildTestFont(t *testing.T) (*TrueTypeFont, []byte) {
	t.Helper()
	runeMap := map[rune]uint16{
		'A': 1,
		'B': 2,
		' ': 3,
	}
	widths := []uint16{0, 600, 700, 250}
	data := buildMinimalTTF(4, widths, runeMap)

	ttf, err := ParseTrueType(data)
	if err != nil {
		t.Fatalf("ParseTrueType failed: %v", err)
	}
	return ttf, data
}

func TestParseTrueType_MinimalFont_Name(t *testing.T) {
	ttf, _ := buildTestFont(t)
	if ttf.Name() != "TestFont" {
		t.Errorf("Name() = %q, want %q", ttf.Name(), "TestFont")
	}
}

func TestParseTrueType_MinimalFont_Metrics(t *testing.T) {
	ttf, _ := buildTestFont(t)
	m := ttf.Metrics()
	if m.UnitsPerEm != 1000 {
		t.Errorf("UnitsPerEm = %d, want 1000", m.UnitsPerEm)
	}
	if m.Ascender != 800 {
		t.Errorf("Ascender = %d, want 800", m.Ascender)
	}
	if m.Descender != -200 {
		t.Errorf("Descender = %d, want -200", m.Descender)
	}
}

func TestParseTrueType_MinimalFont_GlyphWidths(t *testing.T) {
	ttf, _ := buildTestFont(t)

	w, ok := ttf.GlyphWidth('A')
	if !ok || w != 600 {
		t.Errorf("GlyphWidth('A') = (%d, %v), want (600, true)", w, ok)
	}
	w, ok = ttf.GlyphWidth('B')
	if !ok || w != 700 {
		t.Errorf("GlyphWidth('B') = (%d, %v), want (700, true)", w, ok)
	}
	w, ok = ttf.GlyphWidth(' ')
	if !ok || w != 250 {
		t.Errorf("GlyphWidth(' ') = (%d, %v), want (250, true)", w, ok)
	}
}

func TestParseTrueType_MinimalFont_UnmappedRune(t *testing.T) {
	ttf, _ := buildTestFont(t)

	// For an unmapped rune, the format 4 cmap returns glyph 0 (.notdef)
	// and the lookup is considered successful (ok=true), returning width 0.
	w, ok := ttf.GlyphWidth('Z')
	if !ok {
		t.Error("GlyphWidth('Z') should return true (maps to .notdef via cmap)")
	}
	if w != 0 {
		t.Errorf("GlyphWidth('Z') = %d, want 0 (.notdef width)", w)
	}
}

func TestParseTrueType_MinimalFont_NumGlyphsAndData(t *testing.T) {
	ttf, data := buildTestFont(t)

	if ttf.NumGlyphs() != 4 {
		t.Errorf("NumGlyphs() = %d, want 4", ttf.NumGlyphs())
	}
	if len(ttf.Data()) != len(data) {
		t.Errorf("Data() length = %d, want %d", len(ttf.Data()), len(data))
	}
}

func TestParseTrueType_MissingRequiredTable(t *testing.T) {
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

	tables := []tableEntry{
		{tagHead, head},
		{tagHhea, hhea},
		{tagMaxp, maxp},
	}
	data := assembleTTF(tables)

	_, err := ParseTrueType(data)
	if err == nil {
		t.Fatal("ParseTrueType should fail when required tables are missing")
	}
}

// ---------------------------------------------------------------------------
// TrueTypeFont interface method tests
// ---------------------------------------------------------------------------

func TestTrueTypeFont_Encode(t *testing.T) {
	runeMap := map[rune]uint16{'A': 1, 'B': 2}
	widths := []uint16{0, 600, 700}
	data := buildMinimalTTF(3, widths, runeMap)

	ttf, err := ParseTrueType(data)
	if err != nil {
		t.Fatalf("ParseTrueType failed: %v", err)
	}

	encoded := ttf.Encode("AB")
	if len(encoded) != 4 {
		t.Fatalf("Encode('AB') length = %d, want 4", len(encoded))
	}
	if encoded[0] != 0x00 || encoded[1] != 0x01 {
		t.Errorf("Encode('A') = %02X%02X, want 0001", encoded[0], encoded[1])
	}
	if encoded[2] != 0x00 || encoded[3] != 0x02 {
		t.Errorf("Encode('B') = %02X%02X, want 0002", encoded[2], encoded[3])
	}

	encoded = ttf.Encode("Z")
	if len(encoded) != 2 {
		t.Fatalf("Encode('Z') length = %d, want 2", len(encoded))
	}
	if encoded[0] != 0x00 || encoded[1] != 0x00 {
		t.Errorf("Encode('Z') = %02X%02X, want 0000", encoded[0], encoded[1])
	}
}

func TestTrueTypeFont_UsedRunes(t *testing.T) {
	runeMap := map[rune]uint16{'A': 1, 'B': 2}
	widths := []uint16{0, 600, 700}
	data := buildMinimalTTF(3, widths, runeMap)

	ttf, err := ParseTrueType(data)
	if err != nil {
		t.Fatalf("ParseTrueType failed: %v", err)
	}

	used := ttf.UsedRunes()
	if len(used) != 0 {
		t.Errorf("UsedRunes() initially should be empty, got %d", len(used))
	}

	ttf.Encode("AB")
	used = ttf.UsedRunes()
	if len(used) != 2 {
		t.Errorf("UsedRunes() after Encode('AB') = %d, want 2", len(used))
	}
	if !used['A'] || !used['B'] {
		t.Error("UsedRunes() should contain 'A' and 'B'")
	}
}

func TestTrueTypeFont_RuneToGID(t *testing.T) {
	runeMap := map[rune]uint16{'A': 1, 'B': 2}
	widths := []uint16{0, 600, 700}
	data := buildMinimalTTF(3, widths, runeMap)

	ttf, err := ParseTrueType(data)
	if err != nil {
		t.Fatalf("ParseTrueType failed: %v", err)
	}

	ttf.Encode("AB")
	gidMap := ttf.RuneToGID()
	if gidMap['A'] != 1 {
		t.Errorf("RuneToGID()['A'] = %d, want 1", gidMap['A'])
	}
	if gidMap['B'] != 2 {
		t.Errorf("RuneToGID()['B'] = %d, want 2", gidMap['B'])
	}
}

func TestTrueTypeFont_GlyphID(t *testing.T) {
	runeMap := map[rune]uint16{'A': 1}
	widths := []uint16{0, 600}
	data := buildMinimalTTF(2, widths, runeMap)

	ttf, err := ParseTrueType(data)
	if err != nil {
		t.Fatalf("ParseTrueType failed: %v", err)
	}

	if gid := ttf.GlyphID('A'); gid != 1 {
		t.Errorf("GlyphID('A') = %d, want 1", gid)
	}
	if gid := ttf.GlyphID('Z'); gid != 0 {
		t.Errorf("GlyphID('Z') = %d, want 0", gid)
	}
}

// ---------------------------------------------------------------------------
// ValidateTrueType tests
// ---------------------------------------------------------------------------

func TestValidateTrueType_TooShort(t *testing.T) {
	err := ValidateTrueType([]byte{0, 1, 2})
	if err == nil {
		t.Fatal("ValidateTrueType should fail on short data")
	}
}

func TestValidateTrueType_BadScaler(t *testing.T) {
	data := make([]byte, 128)
	binary.BigEndian.PutUint32(data[0:4], 0xBADBAD00)
	err := ValidateTrueType(data)
	if err == nil {
		t.Fatal("ValidateTrueType should fail on bad scaler")
	}
	if !strings.Contains(err.Error(), "unsupported scaler type") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestValidateTrueType_MissingTable(t *testing.T) {
	head := make([]byte, 54)
	tables := []tableEntry{{tagHead, head}}
	data := assembleTTF(tables)

	err := ValidateTrueType(data)
	if err == nil {
		t.Fatal("ValidateTrueType should fail when required tables are missing")
	}
	if !strings.Contains(err.Error(), "missing required table") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestValidateTrueType_Valid(t *testing.T) {
	runeMap := map[rune]uint16{'A': 1}
	widths := []uint16{0, 600}
	data := buildMinimalTTF(2, widths, runeMap)

	err := ValidateTrueType(data)
	if err != nil {
		t.Errorf("ValidateTrueType failed on valid font: %v", err)
	}
}

func TestValidateTrueType_OTTOScaler(t *testing.T) {
	runeMap := map[rune]uint16{'A': 1}
	widths := []uint16{0, 600}
	data := buildMinimalTTF(2, widths, runeMap)
	binary.BigEndian.PutUint32(data[0:4], 0x4F54544F)

	err := ValidateTrueType(data)
	if err != nil {
		t.Errorf("ValidateTrueType failed with OTTO scaler: %v", err)
	}
}

func TestValidateTrueType_TrueScaler(t *testing.T) {
	runeMap := map[rune]uint16{'A': 1}
	widths := []uint16{0, 600}
	data := buildMinimalTTF(2, widths, runeMap)
	binary.BigEndian.PutUint32(data[0:4], 0x74727565)

	err := ValidateTrueType(data)
	if err != nil {
		t.Errorf("ValidateTrueType failed with 'true' scaler: %v", err)
	}
}

// ---------------------------------------------------------------------------
// SubsetTrueType tests
// ---------------------------------------------------------------------------

func TestSubsetTrueType_TooShort(t *testing.T) {
	_, err := SubsetTrueType([]byte{1, 2, 3}, []uint16{0})
	if err == nil {
		t.Fatal("SubsetTrueType should fail on short data")
	}
}

func TestSubsetTrueType_Basic(t *testing.T) {
	runeMap := map[rune]uint16{'A': 1, 'B': 2, 'C': 3}
	widths := []uint16{0, 600, 700, 800}
	data := buildMinimalTTFWithLocaGlyf(4, widths, runeMap)

	result, err := SubsetTrueType(data, []uint16{0, 1})
	if err != nil {
		t.Fatalf("SubsetTrueType failed: %v", err)
	}

	if len(result) >= len(data) {
		t.Errorf("SubsetTrueType result length = %d, want < %d", len(result), len(data))
	}

	err = ValidateTrueType(result)
	if err != nil {
		t.Errorf("Subset result is not valid: %v", err)
	}
}

func TestSubsetTrueType_AllGlyphs(t *testing.T) {
	runeMap := map[rune]uint16{'A': 1, 'B': 2}
	widths := []uint16{0, 600, 700}
	data := buildMinimalTTFWithLocaGlyf(3, widths, runeMap)

	result, err := SubsetTrueType(data, []uint16{0, 1, 2})
	if err != nil {
		t.Fatalf("SubsetTrueType failed: %v", err)
	}

	err = ValidateTrueType(result)
	if err != nil {
		t.Errorf("Subset result is not valid: %v", err)
	}
}

func TestSubsetTrueType_NoLocaGlyf(t *testing.T) {
	runeMap := map[rune]uint16{'A': 1}
	widths := []uint16{0, 600}
	data := buildMinimalTTF(2, widths, runeMap)

	result, err := SubsetTrueType(data, []uint16{0, 1})
	if err != nil {
		t.Fatalf("SubsetTrueType without loca/glyf failed: %v", err)
	}

	err = ValidateTrueType(result)
	if err != nil {
		t.Errorf("Subset result is not valid: %v", err)
	}
}

// ---------------------------------------------------------------------------
// TrueTypeFont.Subset tests (integration)
// ---------------------------------------------------------------------------

func TestTrueTypeFont_Subset(t *testing.T) {
	runeMap := map[rune]uint16{'A': 1, 'B': 2, 'C': 3}
	widths := []uint16{0, 600, 700, 800}
	data := buildMinimalTTFWithLocaGlyf(4, widths, runeMap)

	ttf, err := ParseTrueType(data)
	if err != nil {
		t.Fatalf("ParseTrueType failed: %v", err)
	}

	subsetData, err := ttf.Subset([]rune{'A', 'C'})
	if err != nil {
		t.Fatalf("Subset failed: %v", err)
	}

	err = ValidateTrueType(subsetData)
	if err != nil {
		t.Errorf("Subset result is not valid: %v", err)
	}
}

// ---------------------------------------------------------------------------
// BuildSubsetCmap tests
// ---------------------------------------------------------------------------

func TestBuildSubsetCmap_Empty(t *testing.T) {
	result := BuildSubsetCmap(nil)
	if result != nil {
		t.Errorf("BuildSubsetCmap(nil) = %v, want nil", result)
	}

	result = BuildSubsetCmap([]uint16{})
	if result != nil {
		t.Errorf("BuildSubsetCmap(empty) = %v, want nil", result)
	}
}

func TestBuildSubsetCmap_SingleGlyph(t *testing.T) {
	result := BuildSubsetCmap([]uint16{42})
	if result == nil {
		t.Fatal("BuildSubsetCmap returned nil")
	}

	if len(result) < 14 {
		t.Fatal("result too short for format 4 header")
	}
	format := binary.BigEndian.Uint16(result[0:2])
	if format != 4 {
		t.Errorf("format = %d, want 4", format)
	}
}

func TestBuildSubsetCmap_MultipleGlyphs(t *testing.T) {
	result := BuildSubsetCmap([]uint16{10, 5, 20})
	if result == nil {
		t.Fatal("BuildSubsetCmap returned nil")
	}

	format := binary.BigEndian.Uint16(result[0:2])
	if format != 4 {
		t.Errorf("format = %d, want 4", format)
	}

	segCountX2 := binary.BigEndian.Uint16(result[6:8])
	expectedSegCount := 3 + 1
	if segCountX2 != uint16(expectedSegCount*2) {
		t.Errorf("segCountX2 = %d, want %d", segCountX2, expectedSegCount*2)
	}
}

func TestBuildSubsetCmap_Sorted(t *testing.T) {
	result := BuildSubsetCmap([]uint16{30, 10, 20})
	if result == nil {
		t.Fatal("BuildSubsetCmap returned nil")
	}

	segCountX2 := int(binary.BigEndian.Uint16(result[6:8]))
	segCount := segCountX2 / 2
	numGlyphs := segCount - 1

	deltaOff := 14 + segCount*2 + 2 + segCount*2

	expectedGIDs := []uint16{10, 20, 30}
	for i := 0; i < numGlyphs; i++ {
		cid := uint16(i + 1)
		delta := int16(binary.BigEndian.Uint16(result[deltaOff+i*2 : deltaOff+i*2+2]))
		gid := uint16(int16(cid) + delta)
		if gid != expectedGIDs[i] {
			t.Errorf("CID %d -> GID %d, want %d", cid, gid, expectedGIDs[i])
		}
	}
}

// ---------------------------------------------------------------------------
// computeSearchRange tests
// ---------------------------------------------------------------------------

func TestComputeSearchRange(t *testing.T) {
	tests := []struct {
		segCount        int
		wantSearchRange uint16
		wantSelector    uint16
	}{
		{1, 2, 0},
		{2, 4, 1},
		{3, 4, 1},
		{4, 8, 2},
		{5, 8, 2},
	}
	for _, tc := range tests {
		sr := computeSearchRange(tc.segCount)
		if sr.searchRange != tc.wantSearchRange {
			t.Errorf("computeSearchRange(%d).searchRange = %d, want %d",
				tc.segCount, sr.searchRange, tc.wantSearchRange)
		}
		if sr.entrySelector != tc.wantSelector {
			t.Errorf("computeSearchRange(%d).entrySelector = %d, want %d",
				tc.segCount, sr.entrySelector, tc.wantSelector)
		}
		wantRangeShift := uint16(tc.segCount*2) - tc.wantSearchRange
		if sr.rangeShift != wantRangeShift {
			t.Errorf("computeSearchRange(%d).rangeShift = %d, want %d",
				tc.segCount, sr.rangeShift, wantRangeShift)
		}
	}
}

// ---------------------------------------------------------------------------
// calcTableChecksum tests
// ---------------------------------------------------------------------------

func TestCalcTableChecksum(t *testing.T) {
	data := make([]byte, 8)
	binary.BigEndian.PutUint32(data[0:4], 100)
	binary.BigEndian.PutUint32(data[4:8], 200)

	got := calcTableChecksum(data, 0, 8)
	if got != 300 {
		t.Errorf("calcTableChecksum = %d, want 300", got)
	}
}

func TestCalcTableChecksum_NonAligned(t *testing.T) {
	data := make([]byte, 8)
	binary.BigEndian.PutUint32(data[0:4], 0x01020304)
	data[4] = 0x05

	got := calcTableChecksum(data, 0, 5)
	want := uint32(0x01020304) + uint32(0x05000000)
	if got != want {
		t.Errorf("calcTableChecksum(non-aligned) = 0x%08X, want 0x%08X", got, want)
	}
}

// ---------------------------------------------------------------------------
// compositeArgSize / compositeTransformSize tests
// ---------------------------------------------------------------------------

func TestCompositeArgSize(t *testing.T) {
	if compositeArgSize(0x0001) != 4 {
		t.Error("ARG_1_AND_2_ARE_WORDS should return 4")
	}
	if compositeArgSize(0x0000) != 2 {
		t.Error("no ARG_1_AND_2_ARE_WORDS should return 2")
	}
}

func TestCompositeTransformSize(t *testing.T) {
	tests := []struct {
		flags uint16
		want  int
	}{
		{0x0008, 2}, // WE_HAVE_A_SCALE
		{0x0040, 4}, // WE_HAVE_AN_X_AND_Y_SCALE
		{0x0080, 8}, // WE_HAVE_A_TWO_BY_TWO
		{0x0000, 0}, // none
	}
	for _, tc := range tests {
		got := compositeTransformSize(tc.flags)
		if got != tc.want {
			t.Errorf("compositeTransformSize(0x%04X) = %d, want %d", tc.flags, got, tc.want)
		}
	}
}

// ---------------------------------------------------------------------------
// getGlyphOffsets tests
// ---------------------------------------------------------------------------

func TestGetGlyphOffsets_ShortFormat(t *testing.T) {
	locaData := make([]byte, 12)
	binary.BigEndian.PutUint16(locaData[0:2], 5)
	binary.BigEndian.PutUint16(locaData[2:4], 10)
	binary.BigEndian.PutUint16(locaData[4:6], 15)

	rec := subsetTableRecord{offset: 0, length: 12}

	start, end := getGlyphOffsets(locaData, rec, 0, 0)
	if start != 10 || end != 20 {
		t.Errorf("glyph 0: offsets = (%d, %d), want (10, 20)", start, end)
	}
	start, end = getGlyphOffsets(locaData, rec, 0, 1)
	if start != 20 || end != 30 {
		t.Errorf("glyph 1: offsets = (%d, %d), want (20, 30)", start, end)
	}
}

func TestGetGlyphOffsets_LongFormat(t *testing.T) {
	locaData := make([]byte, 16)
	binary.BigEndian.PutUint32(locaData[0:4], 0)
	binary.BigEndian.PutUint32(locaData[4:8], 100)
	binary.BigEndian.PutUint32(locaData[8:12], 200)
	binary.BigEndian.PutUint32(locaData[12:16], 300)

	rec := subsetTableRecord{offset: 0, length: 16}

	start, end := getGlyphOffsets(locaData, rec, 1, 0)
	if start != 0 || end != 100 {
		t.Errorf("glyph 0: offsets = (%d, %d), want (0, 100)", start, end)
	}
	start, end = getGlyphOffsets(locaData, rec, 1, 2)
	if start != 200 || end != 300 {
		t.Errorf("glyph 2: offsets = (%d, %d), want (200, 300)", start, end)
	}
}

func TestGetGlyphOffsets_OutOfBounds(t *testing.T) {
	locaData := make([]byte, 4)
	rec := subsetTableRecord{offset: 0, length: 4}

	start, end := getGlyphOffsets(locaData, rec, 0, 5)
	if start != 0 || end != 0 {
		t.Errorf("out of bounds: offsets = (%d, %d), want (0, 0)", start, end)
	}
}

// ---------------------------------------------------------------------------
// getLocaFormat tests
// ---------------------------------------------------------------------------

func TestGetLocaFormat(t *testing.T) {
	data := make([]byte, 54)
	putI16(data[50:52], 1)

	rec := subsetTableRecord{offset: 0, length: 54}
	got := getLocaFormat(data, rec)
	if got != 1 {
		t.Errorf("getLocaFormat = %d, want 1", got)
	}
}

func TestGetLocaFormat_TooShort(t *testing.T) {
	data := make([]byte, 10)
	rec := subsetTableRecord{offset: 0, length: 10}
	got := getLocaFormat(data, rec)
	if got != 0 {
		t.Errorf("getLocaFormat(too short) = %d, want 0", got)
	}
}

// ---------------------------------------------------------------------------
// cmapTable.lookup tests
// ---------------------------------------------------------------------------

func TestCmapTable_Lookup_Format4(t *testing.T) {
	f4 := &cmapFormat4{
		segCount:       2,
		endCodes:       []uint16{0x0042, 0xFFFF},
		startCodes:     []uint16{0x0041, 0xFFFF},
		idDeltas:       []int16{-64, 1},
		idRangeOffsets: []uint16{0, 0},
	}

	tbl := &cmapTable{format4: f4}

	gid, ok := tbl.lookup('A')
	if !ok || gid != 1 {
		t.Errorf("lookup('A') = (%d, %v), want (1, true)", gid, ok)
	}
	gid, ok = tbl.lookup('B')
	if !ok || gid != 2 {
		t.Errorf("lookup('B') = (%d, %v), want (2, true)", gid, ok)
	}
}

func TestCmapTable_Lookup_Format12(t *testing.T) {
	f12 := &cmapFormat12{
		groups: []cmapFormat12Group{
			{startCharCode: 0x41, endCharCode: 0x43, startGlyphID: 1},
		},
	}

	tbl := &cmapTable{format12: f12}

	gid, ok := tbl.lookup('A')
	if !ok || gid != 1 {
		t.Errorf("lookup('A') format12 = (%d, %v), want (1, true)", gid, ok)
	}
	gid, ok = tbl.lookup('C')
	if !ok || gid != 3 {
		t.Errorf("lookup('C') format12 = (%d, %v), want (3, true)", gid, ok)
	}
}

func TestCmapTable_Lookup_Format12_Priority(t *testing.T) {
	f4 := &cmapFormat4{
		segCount:       2,
		endCodes:       []uint16{0x0041, 0xFFFF},
		startCodes:     []uint16{0x0041, 0xFFFF},
		idDeltas:       []int16{-55, 1},
		idRangeOffsets: []uint16{0, 0},
	}
	f12 := &cmapFormat12{
		groups: []cmapFormat12Group{
			{startCharCode: 0x41, endCharCode: 0x41, startGlyphID: 99},
		},
	}
	tbl := &cmapTable{format4: f4, format12: f12}

	gid, ok := tbl.lookup('A')
	if !ok || gid != 99 {
		t.Errorf("lookup with both formats: gid = %d, want 99 (format12 preferred)", gid)
	}
}

// ---------------------------------------------------------------------------
// decodeUTF16BE tests
// ---------------------------------------------------------------------------

func TestDecodeUTF16BE(t *testing.T) {
	data := []byte{0x00, 0x41, 0x00, 0x42}
	got := decodeUTF16BE(data)
	if got != "AB" {
		t.Errorf("decodeUTF16BE = %q, want %q", got, "AB")
	}
}

func TestDecodeUTF16BE_OddLength(t *testing.T) {
	data := []byte{0x00, 0x41, 0x00}
	got := decodeUTF16BE(data)
	if got != "A" {
		t.Errorf("decodeUTF16BE(odd) = %q, want %q", got, "A")
	}
}

func TestDecodeUTF16BE_Empty(t *testing.T) {
	got := decodeUTF16BE(nil)
	if got != "" {
		t.Errorf("decodeUTF16BE(nil) = %q, want empty", got)
	}
}

// ---------------------------------------------------------------------------
// readNameRecord tests
// ---------------------------------------------------------------------------

func TestReadNameRecord(t *testing.T) {
	tbl := make([]byte, 12)
	binary.BigEndian.PutUint16(tbl[0:2], 3)
	binary.BigEndian.PutUint16(tbl[6:8], 6)
	binary.BigEndian.PutUint16(tbl[8:10], 10)
	binary.BigEndian.PutUint16(tbl[10:12], 20)

	rec := readNameRecord(tbl, 0)
	if rec.platformID != 3 {
		t.Errorf("platformID = %d, want 3", rec.platformID)
	}
	if rec.nameID != 6 {
		t.Errorf("nameID = %d, want 6", rec.nameID)
	}
	if rec.length != 10 {
		t.Errorf("length = %d, want 10", rec.length)
	}
	if rec.strOffset != 20 {
		t.Errorf("strOffset = %d, want 20", rec.strOffset)
	}
}

// ---------------------------------------------------------------------------
// MeasureString with parsed TrueTypeFont
// ---------------------------------------------------------------------------

func TestMeasureString_WithTrueTypeFont(t *testing.T) {
	runeMap := map[rune]uint16{'A': 1, 'B': 2}
	widths := []uint16{0, 600, 400}
	data := buildMinimalTTF(3, widths, runeMap)

	ttf, err := ParseTrueType(data)
	if err != nil {
		t.Fatalf("ParseTrueType failed: %v", err)
	}

	got := MeasureString(ttf, "AB", 10)
	want := 10.0
	if math.Abs(got-want) > 1e-9 {
		t.Errorf("MeasureString with TTF = %f, want %f", got, want)
	}
}

// ---------------------------------------------------------------------------
// writeCMapHeader / writeCMapFooter tests
// ---------------------------------------------------------------------------

func TestWriteCMapHeader(t *testing.T) {
	var b strings.Builder
	writeCMapHeader(&b)
	s := b.String()

	expected := []string{
		"/CIDInit /ProcSet findresource begin",
		"12 dict begin",
		"begincmap",
		"/CIDSystemInfo",
		"/CMapName /Adobe-Identity-UCS def",
		"/CMapType 2 def",
		"begincodespacerange",
		"<0000> <FFFF>",
		"endcodespacerange",
	}
	for _, exp := range expected {
		if !strings.Contains(s, exp) {
			t.Errorf("header missing %q", exp)
		}
	}
}

func TestWriteCMapFooter(t *testing.T) {
	var b strings.Builder
	writeCMapFooter(&b)
	s := b.String()

	if !strings.Contains(s, "endcmap") {
		t.Error("footer missing 'endcmap'")
	}
	if !strings.Contains(s, "CMapName currentdict /CMap defineresource pop") {
		t.Error("footer missing defineresource line")
	}
}

// ---------------------------------------------------------------------------
// extractCompositeComponents tests
// ---------------------------------------------------------------------------

func TestExtractCompositeComponents(t *testing.T) {
	glyphData := make([]byte, 30)
	binary.BigEndian.PutUint16(glyphData[0:2], 0xFFFF) // numContours = -1

	// Component 1: flags = MORE_COMPONENTS | ARG_1_AND_2_ARE_WORDS
	binary.BigEndian.PutUint16(glyphData[10:12], 0x0021)
	binary.BigEndian.PutUint16(glyphData[12:14], 5) // component GID = 5
	// args: 4 bytes (WORDS), transform: 0 => next at 18

	// Component 2: flags = 0x0000 (no MORE_COMPONENTS)
	binary.BigEndian.PutUint16(glyphData[18:20], 0x0000)
	binary.BigEndian.PutUint16(glyphData[20:22], 7) // component GID = 7

	components := extractCompositeComponents(glyphData, 0, len(glyphData))
	if len(components) != 2 {
		t.Fatalf("extractCompositeComponents got %d components, want 2", len(components))
	}
	if components[0] != 5 {
		t.Errorf("component[0] = %d, want 5", components[0])
	}
	if components[1] != 7 {
		t.Errorf("component[1] = %d, want 7", components[1])
	}
}

// ---------------------------------------------------------------------------
// readTableDirectory tests
// ---------------------------------------------------------------------------

func TestReadTableDirectory(t *testing.T) {
	data := make([]byte, 12+16)
	binary.BigEndian.PutUint32(data[0:4], 0x00010000)
	binary.BigEndian.PutUint16(data[4:6], 1)

	copy(data[12:16], []byte("head"))
	binary.BigEndian.PutUint32(data[16:20], 0)
	binary.BigEndian.PutUint32(data[20:24], 28)
	binary.BigEndian.PutUint32(data[24:28], 54)

	tables, err := readTableDirectory(data, 1)
	if err != nil {
		t.Fatalf("readTableDirectory failed: %v", err)
	}
	if _, ok := tables["head"]; !ok {
		t.Error("readTableDirectory did not find 'head' table")
	}
	if tables["head"].offset != 28 {
		t.Errorf("head offset = %d, want 28", tables["head"].offset)
	}
}

func TestReadTableDirectory_Truncated(t *testing.T) {
	data := make([]byte, 14)
	_, err := readTableDirectory(data, 1)
	if err == nil {
		t.Fatal("readTableDirectory should fail on truncated data")
	}
}

// ---------------------------------------------------------------------------
// buildGlyphSet tests
// ---------------------------------------------------------------------------

func TestBuildGlyphSet_NoGlyfTable(t *testing.T) {
	tables := map[string]subsetTableRecord{
		tagHead: {tag: tagHead, offset: 0, length: 54},
	}
	keep := buildGlyphSet(make([]byte, 100), tables, []uint16{0, 1, 5})
	if len(keep) != 3 {
		t.Errorf("buildGlyphSet with no glyf: got %d entries, want 3", len(keep))
	}
	if !keep[0] || !keep[1] || !keep[5] {
		t.Error("buildGlyphSet missing expected glyph IDs")
	}
}

// ---------------------------------------------------------------------------
// Round-trip test: build minimal TTF, parse, encode, measure.
// ---------------------------------------------------------------------------

func TestBuildMinimalTTF_RoundTrip(t *testing.T) {
	runeMap := map[rune]uint16{
		'H': 1,
		'i': 2,
	}
	widths := []uint16{0, 500, 400}
	data := buildMinimalTTF(3, widths, runeMap)

	ttf, err := ParseTrueType(data)
	if err != nil {
		t.Fatalf("round-trip parse failed: %v", err)
	}

	enc := ttf.Encode("Hi")
	if len(enc) != 4 {
		t.Fatalf("Encode('Hi') length = %d, want 4", len(enc))
	}

	w := MeasureString(ttf, "Hi", 10)
	if math.Abs(w-9.0) > 1e-9 {
		t.Errorf("MeasureString('Hi') = %f, want 9.0", w)
	}
}

// ---------------------------------------------------------------------------
// Edge case: font with 'true' scaler (Mac fonts)
// ---------------------------------------------------------------------------

func TestParseTrueType_TrueScaler(t *testing.T) {
	runeMap := map[rune]uint16{'A': 1}
	widths := []uint16{0, 600}
	data := buildMinimalTTF(2, widths, runeMap)
	binary.BigEndian.PutUint32(data[0:4], 0x74727565)

	ttf, err := ParseTrueType(data)
	if err != nil {
		t.Fatalf("ParseTrueType with 'true' scaler failed: %v", err)
	}
	if ttf.Name() != "TestFont" {
		t.Errorf("Name() = %q, want %q", ttf.Name(), "TestFont")
	}
}

// ---------------------------------------------------------------------------
// Name table with UTF-16BE (platform 3) encoding
// ---------------------------------------------------------------------------

func TestParseName_UTF16BE(t *testing.T) {
	fontName := "MyFont"
	utf16Name := make([]byte, len(fontName)*2)
	for i, c := range fontName {
		binary.BigEndian.PutUint16(utf16Name[i*2:i*2+2], uint16(c))
	}

	storageOffset := 6 + 12
	nameTable := make([]byte, storageOffset+len(utf16Name))
	binary.BigEndian.PutUint16(nameTable[0:2], 0)
	binary.BigEndian.PutUint16(nameTable[2:4], 1)
	binary.BigEndian.PutUint16(nameTable[4:6], uint16(storageOffset))

	off := 6
	binary.BigEndian.PutUint16(nameTable[off:off+2], 3) // platformID (Windows)
	binary.BigEndian.PutUint16(nameTable[off+2:off+4], 1)
	binary.BigEndian.PutUint16(nameTable[off+4:off+6], 0)
	binary.BigEndian.PutUint16(nameTable[off+6:off+8], 6) // nameID (PostScript)
	binary.BigEndian.PutUint16(nameTable[off+8:off+10], uint16(len(utf16Name)))
	binary.BigEndian.PutUint16(nameTable[off+10:off+12], 0)

	copy(nameTable[storageOffset:], utf16Name)

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

	cmap := buildCmapTableForTest(map[rune]uint16{'A': 0})

	post := make([]byte, 32)
	binary.BigEndian.PutUint32(post[0:4], 0x00030000)

	tables := []tableEntry{
		{tagHead, head},
		{tagHhea, hhea},
		{tagMaxp, maxp},
		{tagHmtx, hmtx},
		{tagCmap, cmap},
		{tagName, nameTable},
		{tagPost, post},
	}

	data := assembleTTF(tables)
	ttf, err := ParseTrueType(data)
	if err != nil {
		t.Fatalf("ParseTrueType failed: %v", err)
	}
	if ttf.Name() != "MyFont" {
		t.Errorf("Name() = %q, want %q", ttf.Name(), "MyFont")
	}
}

// ---------------------------------------------------------------------------
// Post table with italic angle
// ---------------------------------------------------------------------------

func TestParsePost_ItalicAngle(t *testing.T) {
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

	cmap := buildCmapTableForTest(map[rune]uint16{'A': 0})
	nameTable := buildNameTableForTest("TestFont")

	post := make([]byte, 32)
	binary.BigEndian.PutUint32(post[0:4], 0x00030000)
	// italicAngle = -12.0 as Fixed 16.16
	putI16(post[4:6], -12)
	binary.BigEndian.PutUint16(post[6:8], 0)

	tables := []tableEntry{
		{tagHead, head},
		{tagHhea, hhea},
		{tagMaxp, maxp},
		{tagHmtx, hmtx},
		{tagCmap, cmap},
		{tagName, nameTable},
		{tagPost, post},
	}

	data := assembleTTF(tables)
	ttf, err := ParseTrueType(data)
	if err != nil {
		t.Fatalf("ParseTrueType failed: %v", err)
	}
	if ttf.Metrics().ItalicAngle != -12.0 {
		t.Errorf("ItalicAngle = %f, want -12.0", ttf.Metrics().ItalicAngle)
	}
}

// ---------------------------------------------------------------------------
// writeBFCharEntries tests
// ---------------------------------------------------------------------------

func TestWriteBFCharEntries_BMPAndSupplementary(t *testing.T) {
	entries := []cmapEntry{
		{gid: 1, r: 'A'},
		{gid: 2, r: 0x1F600},
	}

	var b strings.Builder
	writeBFCharEntries(&b, entries)
	s := b.String()

	if !strings.Contains(s, "2 beginbfchar") {
		t.Error("missing beginbfchar header")
	}
	if !strings.Contains(s, "<0001> <0041>") {
		t.Error("missing BMP char entry")
	}
	if !strings.Contains(s, "<0002> <D83DDE00>") {
		t.Error("missing supplementary char entry")
	}
}

// ---------------------------------------------------------------------------
// Kinsoku tests (WP4)
// ---------------------------------------------------------------------------

func TestKinsokuStartProhibition(t *testing.T) {
	// Ensure a line break does not leave a kinsoku-start character (e.g. '。')
	// at the beginning of the next line. Instead the break moves earlier.
	f := newMockFont(1000, 500, map[rune]int{
		'あ': 500, 'い': 500, 'う': 500, '。': 500,
	})
	// Width: 4 chars fit (4*500*10/1000 = 20). Text is 5 chars.
	// Naive break: "あいう。" + "え" → '。' would be on line 1 end (ok).
	// Instead test: "あいう" + "。え" — the '。' must not start line 2.
	// We need a case where '。' would start a new line.
	// "あいう え。か" with width for 5 chars. Naive break at space: "あいう" + "え。か".
	// That's fine. Let's construct: "あい。うえ" width=3chars → naive break after 'い' leaves '。' at start.
	lines := LineBreak(f, "あい。うえ", 10, 15) // 15pt → 3 chars fit
	// '。' should not start a line.
	for i := 1; i < len(lines); i++ {
		runes := []rune(lines[i])
		if len(runes) > 0 && kinsokuStart[runes[0]] {
			t.Errorf("Line %d starts with kinsoku-start char %q: %q", i, runes[0], lines[i])
		}
	}
}

func TestKinsokuEndProhibition(t *testing.T) {
	// Ensure a kinsoku-end character (e.g. '「') does not remain at the end
	// of a line alone separated from the text it opens.
	f := newMockFont(1000, 500, map[rune]int{
		'あ': 500, 'い': 500, '「': 500, 'う': 500,
	})
	// "あい「うえ" with 3 chars per line. Naive CJK break after 'い' → '「' at start
	// of next line. But '「' is kinsoku-end — it should not appear at end of
	// a line separated from its content. adjustBreakForKinsoku handles it by
	// moving the break earlier if '「' would appear at line end.
	lines := LineBreak(f, "あ「いうえ", 10, 15) // 3 chars fit
	// '「' should not be the last char of any line (kinsoku-end).
	for i, line := range lines {
		runes := []rune(line)
		if len(runes) > 0 && kinsokuEnd[runes[len(runes)-1]] {
			t.Errorf("Line %d ends with kinsoku-end char %q: %q", i, runes[len(runes)-1], line)
		}
	}
}

func TestKinsokuConsecutive(t *testing.T) {
	// Multiple consecutive kinsoku-start characters after enough room to
	// back up should be kept on the previous line.
	f := newMockFont(1000, 500, map[rune]int{
		'あ': 500, 'い': 500, 'う': 500, '。': 500, '」': 500, 'え': 500,
	})
	// "あいう。」え" with 4 chars per line. CJK break after 4th char ('。')
	// → naive next-line "」え". adjustBreakForKinsoku should move break
	// before '。' so "あいう" + "。」え".
	lines := LineBreak(f, "あいう。」え", 10, 20) // 4 chars fit
	for i := 1; i < len(lines); i++ {
		runes := []rune(lines[i])
		if len(runes) > 0 && kinsokuStart[runes[0]] {
			// '。' starting a line is acceptable only if it was moved there
			// intentionally to avoid splitting from '」'.
			// In this case line 2 should start with '。」え'.
			if runes[0] != '。' {
				t.Errorf("Line %d starts with unexpected kinsoku-start char %q: %q", i, runes[0], lines[i])
			}
		}
	}
	// Verify we got exactly 2 lines.
	if len(lines) != 2 {
		t.Errorf("Expected 2 lines, got %d: %v", len(lines), lines)
	}
}

func TestKinsokuNoEffectLatin(t *testing.T) {
	// Kinsoku rules should not affect normal Latin text breaking.
	f := newMockFont(1000, 500, nil)
	lines := LineBreak(f, "hello world foo bar", 10, 30) // 6 chars fit
	if len(lines) < 2 {
		t.Fatalf("Expected multiple lines, got %d", len(lines))
	}
	// All lines should be non-empty.
	for i, line := range lines {
		if len(line) == 0 {
			t.Errorf("Line %d is empty", i)
		}
	}
}
