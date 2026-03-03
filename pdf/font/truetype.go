package font

import (
	"encoding/binary"
	"errors"
	"fmt"
	"sort"
)

// TrueType table tags as 4-byte identifiers.
const (
	tagCmap = "cmap"
	tagHead = "head"
	tagHhea = "hhea"
	tagHmtx = "hmtx"
	tagMaxp = "maxp"
	tagName = "name"
	tagOS2  = "OS/2"
	tagPost = "post"
	tagLoca = "loca"
	tagGlyf = "glyf"
)

// sfntHeader is the TrueType offset table at the start of a font file.
type sfntHeader struct {
	ScalerType    uint32
	NumTables     uint16
	SearchRange   uint16
	EntrySelector uint16
	RangeShift    uint16
}

// tableRecord describes one table in the font file.
type tableRecord struct {
	Tag      [4]byte
	Checksum uint32
	Offset   uint32
	Length   uint32
}

// cmapTable maps Unicode code points to glyph IDs.
type cmapTable struct {
	format4  *cmapFormat4  // BMP mapping (format 4)
	format12 *cmapFormat12 // Full Unicode mapping (format 12)
}

// lookup returns the glyph ID for a Unicode code point.
func (c *cmapTable) lookup(r rune) (uint16, bool) {
	cp := uint32(r)
	// Prefer format 12 for full Unicode coverage.
	if c.format12 != nil {
		for _, g := range c.format12.groups {
			if cp >= g.startCharCode && cp <= g.endCharCode {
				return uint16(g.startGlyphID + (cp - g.startCharCode)), true
			}
		}
	}
	// Fall back to format 4 for BMP.
	if c.format4 != nil && cp <= 0xFFFF {
		return c.format4.lookup(uint16(cp)), true
	}
	return 0, false
}

// cmapFormat4 implements the segment mapping to delta values format
// for the Basic Multilingual Plane.
type cmapFormat4 struct {
	segCount       int
	endCodes       []uint16
	startCodes     []uint16
	idDeltas       []int16
	idRangeOffsets []uint16
	glyphIDArray   []uint16
	// Raw offset where idRangeOffsets begin, used for glyph ID calculation.
	idRangeOffsetBase int
}

func (f4 *cmapFormat4) lookup(cp uint16) uint16 {
	// Binary search for the segment containing cp.
	idx := sort.Search(f4.segCount, func(i int) bool {
		return f4.endCodes[i] >= cp
	})
	if idx >= f4.segCount {
		return 0
	}
	if cp < f4.startCodes[idx] {
		return 0
	}

	if f4.idRangeOffsets[idx] == 0 {
		return uint16(int16(cp) + f4.idDeltas[idx])
	}

	// Use idRangeOffset to index into glyphIDArray.
	// offset = idRangeOffset[idx]/2 + (cp - startCode[idx]) - (segCount - idx)
	glyphIdx := int(f4.idRangeOffsets[idx]/2) + int(cp-f4.startCodes[idx]) - (f4.segCount - idx)
	if glyphIdx < 0 || glyphIdx >= len(f4.glyphIDArray) {
		return 0
	}
	gid := f4.glyphIDArray[glyphIdx]
	if gid == 0 {
		return 0
	}
	return uint16(int16(gid) + f4.idDeltas[idx])
}

// cmapFormat12 implements the segmented coverage format for full Unicode.
type cmapFormat12 struct {
	groups []cmapFormat12Group
}

type cmapFormat12Group struct {
	startCharCode uint32
	endCharCode   uint32
	startGlyphID  uint32
}

// hmtxTable holds horizontal metrics (advance widths) for each glyph.
type hmtxTable struct {
	advanceWidths    []uint16
	leftSideBearings []int16
}

// TrueTypeFont is a parsed TrueType (.ttf) font that implements the [Font]
// interface. It provides glyph metrics, Unicode-to-GID mapping (via cmap
// format 4 and 12), horizontal advance widths, PDF text encoding, and font
// subsetting for embedding. CJK (Chinese, Japanese, Korean) characters are
// fully supported through format 12 cmap tables.
//
// Create a TrueTypeFont by calling [ParseTrueType] with the raw font file data.
type TrueTypeFont struct {
	name       string
	data       []byte // original font file data
	metrics    Metrics
	unitsPerEm int
	cmapTbl    cmapTable
	hmtxTbl    hmtxTable
	numGlyphs  int
	usedRunes  map[rune]bool
}

// ParseTrueType parses a TrueType font file from the given data.
// It reads the sfnt offset table and required tables (head, hhea, maxp,
// cmap, hmtx, name, post, OS/2) to extract font metrics and glyph mappings.
func ParseTrueType(data []byte) (*TrueTypeFont, error) {
	if len(data) < 12 {
		return nil, errors.New("font: data too short for sfnt header")
	}

	var hdr sfntHeader
	hdr.ScalerType = binary.BigEndian.Uint32(data[0:4])
	hdr.NumTables = binary.BigEndian.Uint16(data[4:6])

	// Validate scaler type.
	switch hdr.ScalerType {
	case 0x00010000, // TrueType
		0x74727565, // 'true'
		0x4F54544F: // 'OTTO' (CFF, but we handle the container)
		// OK
	default:
		return nil, fmt.Errorf("font: unsupported scaler type 0x%08X", hdr.ScalerType)
	}

	// Read table directory.
	tables := make(map[string]tableRecord)
	offset := 12
	for i := 0; i < int(hdr.NumTables); i++ {
		if offset+16 > len(data) {
			return nil, errors.New("font: truncated table directory")
		}
		var rec tableRecord
		copy(rec.Tag[:], data[offset:offset+4])
		rec.Checksum = binary.BigEndian.Uint32(data[offset+4 : offset+8])
		rec.Offset = binary.BigEndian.Uint32(data[offset+8 : offset+12])
		rec.Length = binary.BigEndian.Uint32(data[offset+12 : offset+16])
		tag := string(rec.Tag[:])
		tables[tag] = rec
		offset += 16
	}

	ttf := &TrueTypeFont{
		data:      data,
		usedRunes: make(map[rune]bool),
	}

	// Parse head table.
	if err := ttf.parseHead(data, tables); err != nil {
		return nil, err
	}

	// Parse hhea table.
	var numberOfHMetrics uint16
	if err := ttf.parseHhea(data, tables, &numberOfHMetrics); err != nil {
		return nil, err
	}

	// Parse maxp table.
	if err := ttf.parseMaxp(data, tables); err != nil {
		return nil, err
	}

	// Parse OS/2 table (optional but common).
	ttf.parseOS2(data, tables)

	// Parse name table for font name.
	ttf.parseName(data, tables)

	// Parse hmtx table.
	if err := ttf.parseHmtx(data, tables, numberOfHMetrics); err != nil {
		return nil, err
	}

	// Parse cmap table.
	if err := ttf.parseCmap(data, tables); err != nil {
		return nil, err
	}

	// Parse post table (optional metrics).
	ttf.parsePost(data, tables)

	return ttf, nil
}

func (ttf *TrueTypeFont) getTable(data []byte, tables map[string]tableRecord, tag string) ([]byte, error) {
	rec, ok := tables[tag]
	if !ok {
		return nil, fmt.Errorf("font: missing required table '%s'", tag)
	}
	end := int(rec.Offset) + int(rec.Length)
	if end > len(data) {
		return nil, fmt.Errorf("font: table '%s' extends beyond data", tag)
	}
	return data[rec.Offset:end], nil
}

func (ttf *TrueTypeFont) parseHead(data []byte, tables map[string]tableRecord) error {
	tbl, err := ttf.getTable(data, tables, tagHead)
	if err != nil {
		return err
	}
	if len(tbl) < 54 {
		return errors.New("font: head table too short")
	}
	ttf.unitsPerEm = int(binary.BigEndian.Uint16(tbl[18:20]))
	ttf.metrics.UnitsPerEm = ttf.unitsPerEm
	return nil
}

func (ttf *TrueTypeFont) parseHhea(data []byte, tables map[string]tableRecord, numberOfHMetrics *uint16) error {
	tbl, err := ttf.getTable(data, tables, tagHhea)
	if err != nil {
		return err
	}
	if len(tbl) < 36 {
		return errors.New("font: hhea table too short")
	}
	ttf.metrics.Ascender = int(int16(binary.BigEndian.Uint16(tbl[4:6])))
	ttf.metrics.Descender = int(int16(binary.BigEndian.Uint16(tbl[6:8])))
	ttf.metrics.LineGap = int(int16(binary.BigEndian.Uint16(tbl[8:10])))
	*numberOfHMetrics = binary.BigEndian.Uint16(tbl[34:36])
	return nil
}

func (ttf *TrueTypeFont) parseMaxp(data []byte, tables map[string]tableRecord) error {
	tbl, err := ttf.getTable(data, tables, tagMaxp)
	if err != nil {
		return err
	}
	if len(tbl) < 6 {
		return errors.New("font: maxp table too short")
	}
	ttf.numGlyphs = int(binary.BigEndian.Uint16(tbl[4:6]))
	return nil
}

func (ttf *TrueTypeFont) parseOS2(data []byte, tables map[string]tableRecord) {
	tbl, err := ttf.getTable(data, tables, tagOS2)
	if err != nil || len(tbl) < 72 {
		return
	}
	// sCapHeight is at offset 88 in version 2+ of OS/2.
	if len(tbl) >= 90 {
		ttf.metrics.CapHeight = int(int16(binary.BigEndian.Uint16(tbl[88:90])))
	}
	// sxHeight is at offset 86 in version 2+.
	if len(tbl) >= 88 {
		ttf.metrics.XHeight = int(int16(binary.BigEndian.Uint16(tbl[86:88])))
	}
}

// nameRecord represents a single record in the TrueType name table.
type nameRecord struct {
	platformID uint16
	nameID     uint16
	length     uint16
	strOffset  uint16
}

// readNameRecord reads a name record from the given offset in the table data.
func readNameRecord(tbl []byte, offset int) nameRecord {
	return nameRecord{
		platformID: binary.BigEndian.Uint16(tbl[offset : offset+2]),
		nameID:     binary.BigEndian.Uint16(tbl[offset+6 : offset+8]),
		length:     binary.BigEndian.Uint16(tbl[offset+8 : offset+10]),
		strOffset:  binary.BigEndian.Uint16(tbl[offset+10 : offset+12]),
	}
}

// decodeNameString decodes a name string from the table storage area.
func decodeNameString(tbl []byte, storageOffset int, rec nameRecord) string {
	start := storageOffset + int(rec.strOffset)
	end := start + int(rec.length)
	if end > len(tbl) {
		return ""
	}
	if rec.platformID == 3 || rec.platformID == 0 {
		return decodeUTF16BE(tbl[start:end])
	}
	return string(tbl[start:end])
}

func (ttf *TrueTypeFont) parseName(data []byte, tables map[string]tableRecord) {
	tbl, err := ttf.getTable(data, tables, tagName)
	if err != nil || len(tbl) < 6 {
		ttf.name = "Unknown"
		return
	}

	count := int(binary.BigEndian.Uint16(tbl[2:4]))
	storageOffset := int(binary.BigEndian.Uint16(tbl[4:6]))

	var postScriptName, fullName string
	offset := 6
	for i := 0; i < count && offset+12 <= len(tbl); i++ {
		rec := readNameRecord(tbl, offset)
		offset += 12

		if rec.nameID != 6 && rec.nameID != 4 {
			continue
		}
		name := decodeNameString(tbl, storageOffset, rec)
		if rec.nameID == 6 && postScriptName == "" {
			postScriptName = name
		}
		if rec.nameID == 4 && fullName == "" {
			fullName = name
		}
	}

	switch {
	case postScriptName != "":
		ttf.name = postScriptName
	case fullName != "":
		ttf.name = fullName
	default:
		ttf.name = "Unknown"
	}
}

func decodeUTF16BE(b []byte) string {
	if len(b)%2 != 0 {
		b = b[:len(b)-1]
	}
	runes := make([]rune, 0, len(b)/2)
	for i := 0; i+1 < len(b); i += 2 {
		r := rune(binary.BigEndian.Uint16(b[i : i+2]))
		runes = append(runes, r)
	}
	return string(runes)
}

func (ttf *TrueTypeFont) parseHmtx(data []byte, tables map[string]tableRecord, numberOfHMetrics uint16) error {
	tbl, err := ttf.getTable(data, tables, tagHmtx)
	if err != nil {
		return err
	}

	numH := int(numberOfHMetrics)
	if len(tbl) < numH*4 {
		return errors.New("font: hmtx table too short")
	}

	ttf.hmtxTbl.advanceWidths = make([]uint16, ttf.numGlyphs)
	ttf.hmtxTbl.leftSideBearings = make([]int16, ttf.numGlyphs)

	// Read longHorMetric entries.
	for i := 0; i < numH; i++ {
		off := i * 4
		ttf.hmtxTbl.advanceWidths[i] = binary.BigEndian.Uint16(tbl[off : off+2])
		ttf.hmtxTbl.leftSideBearings[i] = int16(binary.BigEndian.Uint16(tbl[off+2 : off+4]))
	}

	// Remaining glyphs share the last advance width.
	lastWidth := uint16(0)
	if numH > 0 {
		lastWidth = ttf.hmtxTbl.advanceWidths[numH-1]
	}

	remainingOffset := numH * 4
	for i := numH; i < ttf.numGlyphs; i++ {
		ttf.hmtxTbl.advanceWidths[i] = lastWidth
		lsbOff := remainingOffset + (i-numH)*2
		if lsbOff+2 <= len(tbl) {
			ttf.hmtxTbl.leftSideBearings[i] = int16(binary.BigEndian.Uint16(tbl[lsbOff : lsbOff+2]))
		}
	}

	return nil
}

func (ttf *TrueTypeFont) parseCmap(data []byte, tables map[string]tableRecord) error {
	tbl, err := ttf.getTable(data, tables, tagCmap)
	if err != nil {
		return err
	}
	if len(tbl) < 4 {
		return errors.New("font: cmap table too short")
	}

	numSubtables := int(binary.BigEndian.Uint16(tbl[2:4]))

	// Find the best subtable: prefer (3,10) for format 12, then (3,1) or (0,3) for format 4.
	type subtableInfo struct {
		platformID uint16
		encodingID uint16
		offset     uint32
	}
	var subtables []subtableInfo

	off := 4
	for i := 0; i < numSubtables; i++ {
		if off+8 > len(tbl) {
			break
		}
		si := subtableInfo{
			platformID: binary.BigEndian.Uint16(tbl[off : off+2]),
			encodingID: binary.BigEndian.Uint16(tbl[off+2 : off+4]),
			offset:     binary.BigEndian.Uint32(tbl[off+4 : off+8]),
		}
		subtables = append(subtables, si)
		off += 8
	}

	// Parse subtables, looking for format 4 and format 12.
	for _, st := range subtables {
		if int(st.offset)+2 > len(tbl) {
			continue
		}
		format := binary.BigEndian.Uint16(tbl[st.offset : st.offset+2])
		switch format {
		case 4:
			if ttf.cmapTbl.format4 == nil {
				f4, err := parseCmapFormat4(tbl, int(st.offset))
				if err == nil {
					ttf.cmapTbl.format4 = f4
				}
			}
		case 12:
			if ttf.cmapTbl.format12 == nil {
				f12, err := parseCmapFormat12(tbl, int(st.offset))
				if err == nil {
					ttf.cmapTbl.format12 = f12
				}
			}
		}
	}

	if ttf.cmapTbl.format4 == nil && ttf.cmapTbl.format12 == nil {
		return errors.New("font: no supported cmap subtable found (need format 4 or 12)")
	}

	return nil
}

func parseCmapFormat4(tbl []byte, offset int) (*cmapFormat4, error) {
	if offset+14 > len(tbl) {
		return nil, errors.New("font: cmap format 4 header too short")
	}

	// format := binary.BigEndian.Uint16(tbl[offset : offset+2]) // already known to be 4
	length := int(binary.BigEndian.Uint16(tbl[offset+2 : offset+4]))
	if offset+length > len(tbl) {
		return nil, errors.New("font: cmap format 4 extends beyond table")
	}

	segCount := int(binary.BigEndian.Uint16(tbl[offset+6:offset+8])) / 2

	f4 := &cmapFormat4{
		segCount: segCount,
	}

	// The arrays start at offset+14.
	arrStart := offset + 14

	// endCodes
	f4.endCodes = make([]uint16, segCount)
	for i := 0; i < segCount; i++ {
		pos := arrStart + i*2
		if pos+2 > len(tbl) {
			return nil, errors.New("font: cmap format 4 endCode array truncated")
		}
		f4.endCodes[i] = binary.BigEndian.Uint16(tbl[pos : pos+2])
	}

	// reservedPad (2 bytes).
	arrStart += segCount*2 + 2

	// startCodes
	f4.startCodes = make([]uint16, segCount)
	for i := 0; i < segCount; i++ {
		pos := arrStart + i*2
		if pos+2 > len(tbl) {
			return nil, errors.New("font: cmap format 4 startCode array truncated")
		}
		f4.startCodes[i] = binary.BigEndian.Uint16(tbl[pos : pos+2])
	}
	arrStart += segCount * 2

	// idDeltas
	f4.idDeltas = make([]int16, segCount)
	for i := 0; i < segCount; i++ {
		pos := arrStart + i*2
		if pos+2 > len(tbl) {
			return nil, errors.New("font: cmap format 4 idDelta array truncated")
		}
		f4.idDeltas[i] = int16(binary.BigEndian.Uint16(tbl[pos : pos+2]))
	}
	arrStart += segCount * 2

	// idRangeOffsets
	f4.idRangeOffsets = make([]uint16, segCount)
	f4.idRangeOffsetBase = arrStart
	for i := 0; i < segCount; i++ {
		pos := arrStart + i*2
		if pos+2 > len(tbl) {
			return nil, errors.New("font: cmap format 4 idRangeOffset array truncated")
		}
		f4.idRangeOffsets[i] = binary.BigEndian.Uint16(tbl[pos : pos+2])
	}
	arrStart += segCount * 2

	// glyphIDArray: remaining data in the subtable.
	glyphIDEnd := offset + length
	if arrStart < glyphIDEnd {
		numGlyphIDs := (glyphIDEnd - arrStart) / 2
		f4.glyphIDArray = make([]uint16, numGlyphIDs)
		for i := 0; i < numGlyphIDs; i++ {
			pos := arrStart + i*2
			if pos+2 > len(tbl) {
				break
			}
			f4.glyphIDArray[i] = binary.BigEndian.Uint16(tbl[pos : pos+2])
		}
	}

	return f4, nil
}

func parseCmapFormat12(tbl []byte, offset int) (*cmapFormat12, error) {
	if offset+16 > len(tbl) {
		return nil, errors.New("font: cmap format 12 header too short")
	}

	// Bytes 0-1: format (12), 2-3: reserved, 4-7: length, 8-11: language, 12-15: numGroups.
	numGroups := int(binary.BigEndian.Uint32(tbl[offset+12 : offset+16]))

	f12 := &cmapFormat12{
		groups: make([]cmapFormat12Group, numGroups),
	}

	groupStart := offset + 16
	for i := 0; i < numGroups; i++ {
		pos := groupStart + i*12
		if pos+12 > len(tbl) {
			return nil, errors.New("font: cmap format 12 group array truncated")
		}
		f12.groups[i] = cmapFormat12Group{
			startCharCode: binary.BigEndian.Uint32(tbl[pos : pos+4]),
			endCharCode:   binary.BigEndian.Uint32(tbl[pos+4 : pos+8]),
			startGlyphID:  binary.BigEndian.Uint32(tbl[pos+8 : pos+12]),
		}
	}

	return f12, nil
}

func (ttf *TrueTypeFont) parsePost(data []byte, tables map[string]tableRecord) {
	tbl, err := ttf.getTable(data, tables, tagPost)
	if err != nil || len(tbl) < 32 {
		return
	}
	// italicAngle is a Fixed (16.16) at offset 4.
	intPart := int16(binary.BigEndian.Uint16(tbl[4:6]))
	fracPart := binary.BigEndian.Uint16(tbl[6:8])
	ttf.metrics.ItalicAngle = float64(intPart) + float64(fracPart)/65536.0
}

// ---------------------------------------------------------------------------
// Font interface implementation
// ---------------------------------------------------------------------------

// Name returns the PostScript name of the font.
func (ttf *TrueTypeFont) Name() string {
	return ttf.name
}

// Metrics returns the font's metric information.
func (ttf *TrueTypeFont) Metrics() Metrics {
	return ttf.metrics
}

// GlyphWidth returns the advance width of the glyph for the given rune
// in font design units.
func (ttf *TrueTypeFont) GlyphWidth(r rune) (int, bool) {
	gid, ok := ttf.cmapTbl.lookup(r)
	if !ok || int(gid) >= len(ttf.hmtxTbl.advanceWidths) {
		return 0, false
	}
	return int(ttf.hmtxTbl.advanceWidths[gid]), true
}

// GlyphID returns the glyph ID for a rune, or 0 if not found.
func (ttf *TrueTypeFont) GlyphID(r rune) uint16 {
	gid, ok := ttf.cmapTbl.lookup(r)
	if !ok {
		return 0
	}
	return gid
}

// Encode encodes text into a byte sequence for PDF content streams.
// For identity-encoded TrueType fonts, each character is mapped to its
// glyph ID and encoded as a big-endian uint16.
func (ttf *TrueTypeFont) Encode(text string) []byte {
	var result []byte
	for _, r := range text {
		ttf.usedRunes[r] = true
		gid, ok := ttf.cmapTbl.lookup(r)
		if !ok {
			gid = 0
		}
		result = append(result, byte(gid>>8), byte(gid&0xFF))
	}
	return result
}

// Subset creates a subsetted font containing only the specified runes.
// It delegates to SubsetTrueType with the appropriate glyph IDs.
func (ttf *TrueTypeFont) Subset(runes []rune) ([]byte, error) {
	glyphIDs := make([]uint16, 0, len(runes)+1)
	glyphIDs = append(glyphIDs, 0) // always include .notdef

	seen := make(map[uint16]bool)
	seen[0] = true
	for _, r := range runes {
		gid, ok := ttf.cmapTbl.lookup(r)
		if ok && !seen[gid] {
			glyphIDs = append(glyphIDs, gid)
			seen[gid] = true
		}
	}

	return SubsetTrueType(ttf.data, glyphIDs)
}

// UsedRunes returns the set of runes that have been encoded so far.
func (ttf *TrueTypeFont) UsedRunes() map[rune]bool {
	result := make(map[rune]bool, len(ttf.usedRunes))
	for r := range ttf.usedRunes {
		result[r] = true
	}
	return result
}

// RuneToGID returns a mapping from rune to glyph ID for all used runes.
func (ttf *TrueTypeFont) RuneToGID() map[rune]uint16 {
	result := make(map[rune]uint16, len(ttf.usedRunes))
	for r := range ttf.usedRunes {
		gid, ok := ttf.cmapTbl.lookup(r)
		if ok {
			result[r] = gid
		}
	}
	return result
}

// NumGlyphs returns the total number of glyphs in the font.
func (ttf *TrueTypeFont) NumGlyphs() int {
	return ttf.numGlyphs
}

// Data returns the original font file data.
func (ttf *TrueTypeFont) Data() []byte {
	return ttf.data
}
