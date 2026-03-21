package render

import (
	"bytes"
	"encoding/binary"
	"strings"
	"testing"

	"github.com/gpdf-dev/gpdf/document"
	"github.com/gpdf-dev/gpdf/document/layout"
	"github.com/gpdf-dev/gpdf/pdf"
	"github.com/gpdf-dev/gpdf/pdf/font"
)

// ---------------------------------------------------------------------------
// Minimal TTF builder for testing (no external font file needed)
// ---------------------------------------------------------------------------

// putI16 writes an int16 as big-endian bytes.
func putI16(b []byte, v int16) {
	binary.BigEndian.PutUint16(b, uint16(v))
}

// buildTestTTFData builds a minimal but parseable TrueType font binary with
// the given rune-to-GID mapping and per-glyph widths. Includes loca and glyf
// tables so that Subset works.
func buildTestTTFData(numGlyphs int, widths []uint16, runeMap map[rune]uint16) []byte {
	// head table (54 bytes)
	head := make([]byte, 54)
	binary.BigEndian.PutUint32(head[0:4], 0x00010000)   // version
	binary.BigEndian.PutUint32(head[4:8], 0x00005000)   // fontRevision
	binary.BigEndian.PutUint32(head[12:16], 0x5F0F3CF5) // magicNumber
	binary.BigEndian.PutUint16(head[16:18], 0x000B)     // flags
	binary.BigEndian.PutUint16(head[18:20], 1000)       // unitsPerEm
	putI16(head[36:38], 0)                              // xMin
	putI16(head[38:40], -200)                           // yMin
	putI16(head[40:42], 1000)                           // xMax
	putI16(head[42:44], 800)                            // yMax
	binary.BigEndian.PutUint16(head[44:46], 0)          // macStyle
	binary.BigEndian.PutUint16(head[46:48], 8)          // lowestRecPPEM
	putI16(head[48:50], 2)                              // fontDirectionHint
	putI16(head[50:52], 1)                              // indexToLocFormat (long)
	putI16(head[52:54], 0)                              // glyphDataFormat

	// hhea table (36 bytes)
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

	// maxp table (6 bytes)
	maxp := make([]byte, 6)
	binary.BigEndian.PutUint32(maxp[0:4], 0x00010000)
	binary.BigEndian.PutUint16(maxp[4:6], uint16(numGlyphs))

	// hmtx table
	hmtx := make([]byte, numberOfHMetrics*4)
	for i := 0; i < numberOfHMetrics; i++ {
		w := uint16(0)
		if i < len(widths) {
			w = widths[i]
		}
		binary.BigEndian.PutUint16(hmtx[i*4:i*4+2], w)
		putI16(hmtx[i*4+2:i*4+4], 0) // lsb
	}

	// cmap table (format 4)
	cmap := buildTestCmapTable(runeMap)

	// name table
	nameTable := buildTestNameTable("TestFont")

	// post table (32 bytes)
	post := make([]byte, 32)
	binary.BigEndian.PutUint32(post[0:4], 0x00030000)

	// glyf table: each glyph is a minimal 12-byte simple glyph
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

	// loca table (long format): numGlyphs + 1 entries
	loca := make([]byte, (numGlyphs+1)*4)
	for i := 0; i <= numGlyphs; i++ {
		binary.BigEndian.PutUint32(loca[i*4:i*4+4], uint32(i*glyphSize))
	}

	tables := []struct {
		tag  string
		data []byte
	}{
		{"head", head},
		{"hhea", hhea},
		{"maxp", maxp},
		{"hmtx", hmtx},
		{"cmap", cmap},
		{"loca", loca},
		{"glyf", glyf},
		{"name", nameTable},
		{"post", post},
	}

	return assembleTestTTF(tables)
}

// buildTestCmapTable builds a minimal cmap table with format 4 subtable.
func buildTestCmapTable(runeMap map[rune]uint16) []byte {
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
	// Sort by startCode (bubble sort for simplicity).
	for i := 0; i < len(segments); i++ {
		for j := i + 1; j < len(segments); j++ {
			if segments[j].startCode < segments[i].startCode {
				segments[i], segments[j] = segments[j], segments[i]
			}
		}
	}
	// Sentinel segment.
	segments = append(segments, seg{startCode: 0xFFFF, endCode: 0xFFFF, delta: 1})

	segCount := len(segments)
	subtableLen := 14 + segCount*2*4 + 2
	subtable := make([]byte, subtableLen)
	binary.BigEndian.PutUint16(subtable[0:2], 4)                   // format
	binary.BigEndian.PutUint16(subtable[2:4], uint16(subtableLen)) // length
	binary.BigEndian.PutUint16(subtable[6:8], uint16(segCount*2))  // segCountX2

	// searchRange, entrySelector, rangeShift
	sr := testSearchRange(segCount)
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
	// idRangeOffsets all zero (from make).

	cmapHeader := make([]byte, 12)
	binary.BigEndian.PutUint16(cmapHeader[0:2], 0) // version
	binary.BigEndian.PutUint16(cmapHeader[2:4], 1) // numTables
	binary.BigEndian.PutUint16(cmapHeader[4:6], 3) // platformID (Windows)
	binary.BigEndian.PutUint16(cmapHeader[6:8], 1) // encodingID (Unicode BMP)
	binary.BigEndian.PutUint32(cmapHeader[8:12], uint32(len(cmapHeader)))

	result := make([]byte, 0, len(cmapHeader)+len(subtable))
	result = append(result, cmapHeader...)
	result = append(result, subtable...)
	return result
}

// buildTestNameTable builds a minimal name table.
func buildTestNameTable(fontName string) []byte {
	nameBytes := []byte(fontName)
	storageOffset := 6 + 12
	totalLen := storageOffset + len(nameBytes)
	tbl := make([]byte, totalLen)
	binary.BigEndian.PutUint16(tbl[0:2], 0)                     // format
	binary.BigEndian.PutUint16(tbl[2:4], 1)                     // count
	binary.BigEndian.PutUint16(tbl[4:6], uint16(storageOffset)) // stringOffset
	off := 6
	binary.BigEndian.PutUint16(tbl[off:off+2], 1)   // platformID (Mac)
	binary.BigEndian.PutUint16(tbl[off+2:off+4], 0) // encodingID
	binary.BigEndian.PutUint16(tbl[off+4:off+6], 0) // languageID
	binary.BigEndian.PutUint16(tbl[off+6:off+8], 6) // nameID (PostScript)
	binary.BigEndian.PutUint16(tbl[off+8:off+10], uint16(len(nameBytes)))
	binary.BigEndian.PutUint16(tbl[off+10:off+12], 0) // offset
	copy(tbl[storageOffset:], nameBytes)
	return tbl
}

type testSearchRangeResult struct {
	searchRange   uint16
	entrySelector uint16
	rangeShift    uint16
}

func testSearchRange(n int) testSearchRangeResult {
	var r testSearchRangeResult
	if n <= 0 {
		return r
	}
	power := 1
	exp := uint16(0)
	for power*2 <= n {
		power *= 2
		exp++
	}
	r.searchRange = uint16(power * 2)
	r.entrySelector = exp
	r.rangeShift = uint16(n*2) - r.searchRange
	return r
}

// assembleTestTTF assembles table entries into a valid TrueType font binary.
func assembleTestTTF(tables []struct {
	tag  string
	data []byte
}) []byte {
	numTables := len(tables)
	headerSize := 12 + numTables*16

	offsets := make([]int, numTables)
	currentOffset := headerSize
	for i := range tables {
		offsets[i] = currentOffset
		size := len(tables[i].data)
		currentOffset += (size + 3) &^ 3 // align to 4 bytes
	}

	result := make([]byte, currentOffset)
	binary.BigEndian.PutUint32(result[0:4], 0x00010000) // scaler type
	binary.BigEndian.PutUint16(result[4:6], uint16(numTables))

	sr := testSearchRange(numTables)
	binary.BigEndian.PutUint16(result[6:8], sr.searchRange)
	binary.BigEndian.PutUint16(result[8:10], sr.entrySelector)
	binary.BigEndian.PutUint16(result[10:12], sr.rangeShift)

	dirOffset := 12
	for i, tbl := range tables {
		copy(result[dirOffset:dirOffset+4], []byte(tbl.tag))
		binary.BigEndian.PutUint32(result[dirOffset+4:dirOffset+8], 0) // checksum placeholder
		binary.BigEndian.PutUint32(result[dirOffset+8:dirOffset+12], uint32(offsets[i]))
		binary.BigEndian.PutUint32(result[dirOffset+12:dirOffset+16], uint32(len(tbl.data)))
		dirOffset += 16
	}

	for i, tbl := range tables {
		copy(result[offsets[i]:], tbl.data)
	}

	// Recalculate table checksums.
	recalcTestChecksums(result, numTables)

	return result
}

// recalcTestChecksums recalculates checksums for each table in the font binary.
func recalcTestChecksums(data []byte, numTables int) {
	for i := 0; i < numTables; i++ {
		dirOff := 12 + i*16
		tblOffset := binary.BigEndian.Uint32(data[dirOff+8 : dirOff+12])
		tblLength := binary.BigEndian.Uint32(data[dirOff+12 : dirOff+16])
		var sum uint32
		end := tblOffset + tblLength
		// Pad to 4-byte boundary.
		paddedEnd := (end + 3) &^ 3
		if paddedEnd > uint32(len(data)) {
			paddedEnd = uint32(len(data))
		}
		for j := tblOffset; j+4 <= paddedEnd; j += 4 {
			sum += binary.BigEndian.Uint32(data[j : j+4])
		}
		binary.BigEndian.PutUint32(data[dirOff+4:dirOff+8], sum)
	}
}

// buildAndParseTTF is a test helper that builds a minimal TTF and parses it.
func buildAndParseTTF(t *testing.T) (*font.TrueTypeFont, []byte) {
	t.Helper()
	runeMap := map[rune]uint16{
		'A': 1,
		'B': 2,
		'C': 3,
		' ': 4,
	}
	widths := []uint16{0, 600, 700, 500, 250} // .notdef=0, A=600, B=700, C=500, space=250
	data := buildTestTTFData(5, widths, runeMap)

	ttf, err := font.ParseTrueType(data)
	if err != nil {
		t.Fatalf("ParseTrueType failed on test font: %v", err)
	}
	return ttf, data
}

// ---------------------------------------------------------------------------
// RegisterTTFont tests (no env var needed)
// ---------------------------------------------------------------------------

func TestRegisterTTFont_InMemory(t *testing.T) {
	ttf, rawData := buildAndParseTTF(t)
	r, _ := newTestRenderer(t)

	r.RegisterTTFont("TestFont", ttf, rawData)

	if _, ok := r.ttFonts["TestFont"]; !ok {
		t.Error("ttFonts should contain TestFont")
	}
	if _, ok := r.ttFontData["TestFont"]; !ok {
		t.Error("ttFontData should contain TestFont")
	}
}

func TestRegisterTTFont_MultipleInMemory(t *testing.T) {
	ttf, rawData := buildAndParseTTF(t)
	r, _ := newTestRenderer(t)

	r.RegisterTTFont("Font1", ttf, rawData)
	r.RegisterTTFont("Font2", ttf, rawData)

	if len(r.ttFonts) != 2 {
		t.Errorf("ttFonts count = %d, want 2", len(r.ttFonts))
	}
	if len(r.ttFontData) != 2 {
		t.Errorf("ttFontData count = %d, want 2", len(r.ttFontData))
	}
}

// ---------------------------------------------------------------------------
// resolveTextFont with TT font tests
// ---------------------------------------------------------------------------

func TestResolveTextFont_TTFont_Bold(t *testing.T) {
	ttf, rawData := buildAndParseTTF(t)
	r, _ := newTestRenderer(t)
	r.RegisterTTFont("MyFont", ttf, rawData)

	style := document.DefaultStyle()
	style.FontFamily = "MyFont"
	style.FontWeight = document.WeightBold

	fontName, ttFontName, simulateBold, simulateItalic := r.resolveTextFont(style)
	if fontName != "MyFont" {
		t.Errorf("fontName = %q, want MyFont", fontName)
	}
	if ttFontName != "MyFont" {
		t.Errorf("ttFontName = %q, want MyFont", ttFontName)
	}
	if !simulateBold {
		t.Error("bold TT font should simulate bold")
	}
	if simulateItalic {
		t.Error("should not simulate italic")
	}
}

func TestResolveTextFont_TTFont_Italic(t *testing.T) {
	ttf, rawData := buildAndParseTTF(t)
	r, _ := newTestRenderer(t)
	r.RegisterTTFont("MyFont", ttf, rawData)

	style := document.DefaultStyle()
	style.FontFamily = "MyFont"
	style.FontStyle = document.StyleItalic

	_, _, simulateBold, simulateItalic := r.resolveTextFont(style)
	if simulateBold {
		t.Error("should not simulate bold")
	}
	if !simulateItalic {
		t.Error("italic TT font should simulate italic")
	}
}

func TestResolveTextFont_TTFont_BoldItalic(t *testing.T) {
	ttf, rawData := buildAndParseTTF(t)
	r, _ := newTestRenderer(t)
	r.RegisterTTFont("MyFont", ttf, rawData)

	style := document.DefaultStyle()
	style.FontFamily = "MyFont"
	style.FontWeight = document.WeightBold
	style.FontStyle = document.StyleItalic

	_, _, simulateBold, simulateItalic := r.resolveTextFont(style)
	if !simulateBold {
		t.Error("bold+italic TT font should simulate bold")
	}
	if !simulateItalic {
		t.Error("bold+italic TT font should simulate italic")
	}
}

func TestResolveTextFont_TTFont_Normal(t *testing.T) {
	ttf, rawData := buildAndParseTTF(t)
	r, _ := newTestRenderer(t)
	r.RegisterTTFont("MyFont", ttf, rawData)

	style := document.DefaultStyle()
	style.FontFamily = "MyFont"

	fontName, ttFontName, simulateBold, simulateItalic := r.resolveTextFont(style)
	if fontName != "MyFont" {
		t.Errorf("fontName = %q, want MyFont", fontName)
	}
	if ttFontName != "MyFont" {
		t.Errorf("ttFontName = %q, want MyFont", ttFontName)
	}
	if simulateBold {
		t.Error("normal should not simulate bold")
	}
	if simulateItalic {
		t.Error("normal should not simulate italic")
	}
}

// ---------------------------------------------------------------------------
// ensureFont with TT font (triggers OnBeforeClose hook registration)
// ---------------------------------------------------------------------------

func TestEnsureFont_TTFont(t *testing.T) {
	ttf, rawData := buildAndParseTTF(t)
	r, _ := newTestRenderer(t)
	r.RegisterTTFont("MyTTFont", ttf, rawData)

	resName, err := r.ensureFont("MyTTFont")
	if err != nil {
		t.Fatalf("ensureFont error: %v", err)
	}
	if resName == "" {
		t.Error("resName should not be empty")
	}
	if _, ok := r.fontMap["MyTTFont"]; !ok {
		t.Error("fontMap should contain MyTTFont")
	}
	if _, ok := r.fontRefs["MyTTFont"]; !ok {
		t.Error("fontRefs should contain MyTTFont")
	}
}

func TestEnsureFont_TTFont_CachedSecondCall(t *testing.T) {
	ttf, rawData := buildAndParseTTF(t)
	r, _ := newTestRenderer(t)
	r.RegisterTTFont("MyTTFont", ttf, rawData)

	resName1, err := r.ensureFont("MyTTFont")
	if err != nil {
		t.Fatal(err)
	}
	resName2, err := r.ensureFont("MyTTFont")
	if err != nil {
		t.Fatal(err)
	}
	if resName1 != resName2 {
		t.Errorf("second call should return cached name: %q vs %q", resName1, resName2)
	}
}

// ---------------------------------------------------------------------------
// buildGlyphWidthArray tests (in-memory font)
// ---------------------------------------------------------------------------

func TestBuildGlyphWidthArray_InMemory_Empty(t *testing.T) {
	ttf, _ := buildAndParseTTF(t)
	emptyMap := map[rune]uint16{}
	result := buildGlyphWidthArray(ttf, emptyMap, ttf.Metrics().UnitsPerEm)
	if len(result) != 0 {
		t.Errorf("empty runeToGID should produce empty array, got %d elements", len(result))
	}
}

func TestBuildGlyphWidthArray_InMemory_WithGlyphs(t *testing.T) {
	ttf, _ := buildAndParseTTF(t)
	// Encode some text to populate usedRunes.
	ttf.Encode("AB")
	runeToGID := ttf.RuneToGID()
	metrics := ttf.Metrics()
	result := buildGlyphWidthArray(ttf, runeToGID, metrics.UnitsPerEm)
	if len(result) == 0 {
		t.Error("width array should not be empty for encoded glyphs")
	}
}

func TestBuildGlyphWidthArray_InMemory_ConsecutiveGIDs(t *testing.T) {
	ttf, _ := buildAndParseTTF(t)
	ttf.Encode("ABC")
	runeToGID := ttf.RuneToGID()
	metrics := ttf.Metrics()
	result := buildGlyphWidthArray(ttf, runeToGID, metrics.UnitsPerEm)
	// GIDs 1, 2, 3 are consecutive, should be grouped in one W entry: [startGID [w1 w2 w3]]
	if len(result) < 2 {
		t.Fatalf("expected at least 2 elements in W array, got %d", len(result))
	}
}

// ---------------------------------------------------------------------------
// subsetFontData tests (in-memory font)
// ---------------------------------------------------------------------------

func TestSubsetFontData_InMemory(t *testing.T) {
	ttf, rawData := buildAndParseTTF(t)
	r, _ := newTestRenderer(t)
	r.RegisterTTFont("TestFont", ttf, rawData)

	// Encode characters to populate usedRunes.
	ttf.Encode("ABC")

	subsetData := r.subsetFontData(ttf, rawData)
	if len(subsetData) == 0 {
		t.Error("subset data should not be empty")
	}
}

func TestSubsetFontData_InMemory_NoUsedRunes(t *testing.T) {
	ttf, rawData := buildAndParseTTF(t)
	r, _ := newTestRenderer(t)
	// No characters encoded.
	subsetData := r.subsetFontData(ttf, rawData)
	if subsetData == nil {
		t.Error("subset data should not be nil (fallback to raw data)")
	}
}

// ---------------------------------------------------------------------------
// writeType0Font integration test (via EndDocument triggering beforeClose)
// ---------------------------------------------------------------------------

func TestWriteType0Font_Integration(t *testing.T) {
	ttf, rawData := buildAndParseTTF(t)

	var buf bytes.Buffer
	w := pdf.NewWriter(&buf)
	r := NewPDFRenderer(w)
	r.RegisterTTFont("InMemFont", ttf, rawData)

	if err := r.BeginDocument(document.DocumentMetadata{Title: "TTF Test"}); err != nil {
		t.Fatal(err)
	}
	if err := r.BeginPage(document.A4); err != nil {
		t.Fatal(err)
	}

	style := document.DefaultStyle()
	style.FontFamily = "InMemFont"
	style.FontSize = 14
	if err := r.RenderText("AB C", document.Point{X: 72, Y: 72}, style); err != nil {
		t.Fatal(err)
	}

	if err := r.EndPage(); err != nil {
		t.Fatal(err)
	}
	if err := r.EndDocument(); err != nil {
		t.Fatal(err)
	}

	output := buf.String()
	for _, want := range []string{
		"%PDF-1.7",
		"%%EOF",
		"/Type0",
		"/CIDFontType2",
		"/ToUnicode",
		"/FontDescriptor",
		"/FontFile2",
		"/Identity-H",
	} {
		if !strings.Contains(output, want) {
			t.Errorf("output should contain %q", want)
		}
	}
}

// ---------------------------------------------------------------------------
// writeFontDescriptor test (in-memory)
// ---------------------------------------------------------------------------

func TestWriteFontDescriptor_InMemory(t *testing.T) {
	ttf, _ := buildAndParseTTF(t)

	var buf bytes.Buffer
	w := pdf.NewWriter(&buf)
	r := NewPDFRenderer(w)
	metrics := ttf.Metrics()

	fontFileRef := w.AllocObject()
	descRef, err := r.writeFontDescriptor(w, "TestFont", metrics, fontFileRef)
	if err != nil {
		t.Fatalf("writeFontDescriptor error: %v", err)
	}
	if descRef.Number == 0 {
		t.Error("descriptor ref should be non-zero")
	}
}

// ---------------------------------------------------------------------------
// RenderText with TT font (hex encoding path)
// ---------------------------------------------------------------------------

func TestRenderText_TTFont_HexEncoding(t *testing.T) {
	ttf, rawData := buildAndParseTTF(t)

	var buf bytes.Buffer
	w := pdf.NewWriter(&buf)
	r := NewPDFRenderer(w)
	r.RegisterTTFont("TestTT", ttf, rawData)

	if err := r.BeginPage(document.Size{Width: 595, Height: 842}); err != nil {
		t.Fatal(err)
	}

	style := document.DefaultStyle()
	style.FontFamily = "TestTT"
	style.FontSize = 12
	if err := r.RenderText("ABC", document.Point{X: 50, Y: 50}, style); err != nil {
		t.Fatal(err)
	}

	content := string(r.pageContent)
	// TT font text should use hex encoding <XX> Tj, not literal string () Tj.
	if strings.Contains(content, "(ABC) Tj") {
		t.Error("TrueType font should not use literal string encoding")
	}
	if !strings.Contains(content, "> Tj") {
		t.Error("expected hex-encoded text with > Tj suffix")
	}
}

// ---------------------------------------------------------------------------
// RenderText with TT font + bold simulation
// ---------------------------------------------------------------------------

func TestRenderText_TTFont_BoldSimulation(t *testing.T) {
	ttf, rawData := buildAndParseTTF(t)

	var buf bytes.Buffer
	w := pdf.NewWriter(&buf)
	r := NewPDFRenderer(w)
	r.RegisterTTFont("TestTT", ttf, rawData)

	if err := r.BeginPage(document.Size{Width: 595, Height: 842}); err != nil {
		t.Fatal(err)
	}

	style := document.DefaultStyle()
	style.FontFamily = "TestTT"
	style.FontSize = 12
	style.FontWeight = document.WeightBold
	if err := r.RenderText("AB", document.Point{X: 50, Y: 50}, style); err != nil {
		t.Fatal(err)
	}

	content := string(r.pageContent)
	if !strings.Contains(content, "2 Tr") {
		t.Error("bold TT should use rendering mode 2")
	}
	if !strings.Contains(content, "0 Tr") {
		t.Error("should reset rendering mode after bold text")
	}
}

// ---------------------------------------------------------------------------
// RenderText with TT font + italic simulation
// ---------------------------------------------------------------------------

func TestRenderText_TTFont_ItalicSimulation(t *testing.T) {
	ttf, rawData := buildAndParseTTF(t)

	var buf bytes.Buffer
	w := pdf.NewWriter(&buf)
	r := NewPDFRenderer(w)
	r.RegisterTTFont("TestTT", ttf, rawData)

	if err := r.BeginPage(document.Size{Width: 595, Height: 842}); err != nil {
		t.Fatal(err)
	}

	style := document.DefaultStyle()
	style.FontFamily = "TestTT"
	style.FontSize = 12
	style.FontStyle = document.StyleItalic
	if err := r.RenderText("AB", document.Point{X: 50, Y: 50}, style); err != nil {
		t.Fatal(err)
	}

	content := string(r.pageContent)
	if !strings.Contains(content, "Tm") {
		t.Error("italic TT should use Tm operator")
	}
	if !strings.Contains(content, "0.2126") {
		t.Error("italic should use 0.2126 shear factor")
	}
}

// ---------------------------------------------------------------------------
// RenderDocument with TT font (multi-page, full pipeline)
// ---------------------------------------------------------------------------

func TestRenderDocument_TTFont_MultiPage(t *testing.T) {
	ttf, rawData := buildAndParseTTF(t)

	var buf bytes.Buffer
	w := pdf.NewWriter(&buf)
	r := NewPDFRenderer(w)
	r.RegisterTTFont("TestTT", ttf, rawData)

	textNode1 := &document.Text{
		Content: "Page one",
		TextStyle: document.Style{
			FontFamily: "TestTT",
			FontSize:   12,
			Color:      pdf.Black,
		},
	}
	textNode2 := &document.Text{
		Content: "ABC",
		TextStyle: document.Style{
			FontFamily: "TestTT",
			FontSize:   14,
			Color:      pdf.Black,
		},
	}

	pages := []layout.PageLayout{
		{
			Size: document.Size{Width: 595, Height: 842},
			Children: []layout.PlacedNode{
				{
					Node:     textNode1,
					Position: document.Point{X: 72, Y: 72},
					Size:     document.Size{Width: 200, Height: 20},
				},
			},
		},
		{
			Size: document.Size{Width: 595, Height: 842},
			Children: []layout.PlacedNode{
				{
					Node:     textNode2,
					Position: document.Point{X: 72, Y: 100},
					Size:     document.Size{Width: 200, Height: 20},
				},
			},
		},
	}

	if err := r.RenderDocument(pages, document.DocumentMetadata{Title: "Multi-page TT"}); err != nil {
		t.Fatalf("RenderDocument failed: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "/Type0") {
		t.Error("output should contain Type0 font")
	}
	if !strings.Contains(output, "%PDF-1.7") {
		t.Error("output should contain PDF header")
	}
	if !strings.Contains(output, "%%EOF") {
		t.Error("output should contain EOF marker")
	}
}

// ---------------------------------------------------------------------------
// RenderDocument with empty pages list
// ---------------------------------------------------------------------------

func TestRenderDocument_EmptyPages(t *testing.T) {
	var buf bytes.Buffer
	w := pdf.NewWriter(&buf)
	r := NewPDFRenderer(w)

	err := r.RenderDocument(nil, document.DocumentMetadata{Title: "Empty"})
	if err != nil {
		t.Fatalf("RenderDocument with no pages should not error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "%PDF-1.7") {
		t.Error("should produce valid PDF even with no pages")
	}
}

// ---------------------------------------------------------------------------
// RenderDocument with mixed standard + TT font nodes
// ---------------------------------------------------------------------------

func TestRenderDocument_MixedFonts(t *testing.T) {
	ttf, rawData := buildAndParseTTF(t)

	var buf bytes.Buffer
	w := pdf.NewWriter(&buf)
	r := NewPDFRenderer(w)
	r.RegisterTTFont("TestTT", ttf, rawData)

	stdTextNode := &document.Text{
		Content: "Standard text",
		TextStyle: document.Style{
			FontFamily: "Helvetica",
			FontSize:   12,
			Color:      pdf.Black,
		},
	}
	ttTextNode := &document.Text{
		Content: "ABC",
		TextStyle: document.Style{
			FontFamily: "TestTT",
			FontSize:   12,
			Color:      pdf.Black,
		},
	}

	pages := []layout.PageLayout{
		{
			Size: document.Size{Width: 595, Height: 842},
			Children: []layout.PlacedNode{
				{
					Node:     stdTextNode,
					Position: document.Point{X: 72, Y: 72},
					Size:     document.Size{Width: 200, Height: 20},
				},
				{
					Node:     ttTextNode,
					Position: document.Point{X: 72, Y: 100},
					Size:     document.Size{Width: 200, Height: 20},
				},
			},
		},
	}

	if err := r.RenderDocument(pages, document.DocumentMetadata{Title: "Mixed"}); err != nil {
		t.Fatalf("RenderDocument failed: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "/Type0") {
		t.Error("should contain Type0 for TT font")
	}
}

// ---------------------------------------------------------------------------
// RenderDocument with background and border nodes
// ---------------------------------------------------------------------------

func TestRenderDocument_WithBackgroundAndBorders(t *testing.T) {
	var buf bytes.Buffer
	w := pdf.NewWriter(&buf)
	r := NewPDFRenderer(w)

	bg := pdf.Color{R: 0.9, G: 0.9, B: 0.9}
	borderColor := pdf.Red
	textNode := &document.Text{
		Content: "Styled",
		TextStyle: document.Style{
			FontFamily: "Helvetica",
			FontSize:   12,
			Color:      pdf.Black,
			Background: &bg,
			Border: document.BorderEdges{
				Top: document.BorderSide{
					Style: document.BorderSolid,
					Width: document.Pt(1),
					Color: borderColor,
				},
				Bottom: document.BorderSide{
					Style: document.BorderSolid,
					Width: document.Pt(1),
					Color: borderColor,
				},
				Left: document.BorderSide{
					Style: document.BorderSolid,
					Width: document.Pt(1),
					Color: borderColor,
				},
				Right: document.BorderSide{
					Style: document.BorderSolid,
					Width: document.Pt(1),
					Color: borderColor,
				},
			},
		},
	}

	pages := []layout.PageLayout{
		{
			Size: document.Size{Width: 595, Height: 842},
			Children: []layout.PlacedNode{
				{
					Node:     textNode,
					Position: document.Point{X: 50, Y: 50},
					Size:     document.Size{Width: 100, Height: 20},
				},
			},
		},
	}

	if err := r.RenderDocument(pages, document.DocumentMetadata{}); err != nil {
		t.Fatalf("RenderDocument failed: %v", err)
	}
}

// ---------------------------------------------------------------------------
// RenderDocument with nil node (should be handled gracefully)
// ---------------------------------------------------------------------------

func TestRenderDocument_NilNode(t *testing.T) {
	var buf bytes.Buffer
	w := pdf.NewWriter(&buf)
	r := NewPDFRenderer(w)

	pages := []layout.PageLayout{
		{
			Size: document.Size{Width: 595, Height: 842},
			Children: []layout.PlacedNode{
				{
					Node:     nil,
					Position: document.Point{X: 50, Y: 50},
					Size:     document.Size{Width: 100, Height: 20},
				},
			},
		},
	}

	if err := r.RenderDocument(pages, document.DocumentMetadata{}); err != nil {
		t.Fatalf("RenderDocument with nil node should not error: %v", err)
	}
}

// ---------------------------------------------------------------------------
// RenderDocument with nested children
// ---------------------------------------------------------------------------

func TestRenderDocument_NestedChildren(t *testing.T) {
	var buf bytes.Buffer
	w := pdf.NewWriter(&buf)
	r := NewPDFRenderer(w)

	childText := &document.Text{
		Content: "Child text",
		TextStyle: document.Style{
			FontFamily: "Helvetica",
			FontSize:   10,
			Color:      pdf.Black,
		},
	}

	parentText := &document.Text{
		Content: "Parent",
		TextStyle: document.Style{
			FontFamily: "Helvetica",
			FontSize:   12,
			Color:      pdf.Black,
		},
	}

	pages := []layout.PageLayout{
		{
			Size: document.Size{Width: 595, Height: 842},
			Children: []layout.PlacedNode{
				{
					Node:     parentText,
					Position: document.Point{X: 50, Y: 50},
					Size:     document.Size{Width: 200, Height: 40},
					Children: []layout.PlacedNode{
						{
							Node:     childText,
							Position: document.Point{X: 10, Y: 20},
							Size:     document.Size{Width: 100, Height: 15},
						},
					},
				},
			},
		},
	}

	if err := r.RenderDocument(pages, document.DocumentMetadata{}); err != nil {
		t.Fatalf("RenderDocument with nested children failed: %v", err)
	}
}

// ---------------------------------------------------------------------------
// RenderDocument with text decoration
// ---------------------------------------------------------------------------

func TestRenderDocument_TextDecoration(t *testing.T) {
	var buf bytes.Buffer
	w := pdf.NewWriter(&buf)
	r := NewPDFRenderer(w)

	textNode := &document.Text{
		Content: "Underlined",
		TextStyle: document.Style{
			FontFamily:     "Helvetica",
			FontSize:       12,
			Color:          pdf.Black,
			TextDecoration: document.DecorationUnderline | document.DecorationStrikethrough | document.DecorationOverline,
		},
	}

	pages := []layout.PageLayout{
		{
			Size: document.Size{Width: 595, Height: 842},
			Children: []layout.PlacedNode{
				{
					Node:     textNode,
					Position: document.Point{X: 50, Y: 50},
					Size:     document.Size{Width: 100, Height: 20},
				},
			},
		},
	}

	if err := r.RenderDocument(pages, document.DocumentMetadata{}); err != nil {
		t.Fatalf("RenderDocument failed: %v", err)
	}
}

// ---------------------------------------------------------------------------
// writePaintOp coverage (no-fill-no-stroke path)
// ---------------------------------------------------------------------------

func TestWritePaintOp_NoFillNoStroke(t *testing.T) {
	var buf strings.Builder
	writePaintOp(&buf, false, false)
	if !strings.Contains(buf.String(), "n\n") {
		t.Error("no fill + no stroke should produce 'n' operator")
	}
}

func TestWritePaintOp_StrokeOnly(t *testing.T) {
	var buf strings.Builder
	writePaintOp(&buf, false, true)
	if !strings.Contains(buf.String(), "S\n") {
		t.Error("stroke only should produce 'S' operator")
	}
}
