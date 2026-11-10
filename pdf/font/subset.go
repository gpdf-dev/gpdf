package font

import (
	"encoding/binary"
	"errors"
	"fmt"
	"sort"
)

// SubsetTrueType creates a subset of a TrueType font containing only the
// specified glyphs, suitable for embedding as a PDF FontFile2 stream.
//
// Glyph IDs are preserved (CID == GID with Identity-H encoding), but:
//   - outlines of unused glyphs are removed from glyf and loca is rebuilt,
//   - metrics of unused glyphs are zeroed in hmtx,
//   - tables a PDF renderer never consults (OpenType layout, font
//     variations such as gvar/HVAR, vertical metrics, color, ...) are
//     dropped, and
//   - glyph names are stripped from post.
//
// The original cmap is kept as-is; use (*TrueTypeFont).Subset to also
// shrink cmap to the used characters.
//
// The glyphIDs slice should include glyph 0 (.notdef). Composite glyphs
// that reference other glyphs will have their component glyph IDs
// automatically included. Fonts without glyf/loca outlines (e.g. CFF-based
// OpenType) are returned unchanged.
func SubsetTrueType(fontData []byte, glyphIDs []uint16) ([]byte, error) {
	return subsetTrueType(fontData, glyphIDs, nil)
}

// subsetTrueType implements SubsetTrueType. When cmapping is non-nil, the
// cmap table is replaced by one covering exactly those rune → GID pairs.
func subsetTrueType(fontData []byte, glyphIDs []uint16, cmapping map[rune]uint16) ([]byte, error) {
	if len(fontData) < 12 {
		return nil, errors.New("font: data too short for subsetting")
	}

	numTables := int(binary.BigEndian.Uint16(fontData[4:6]))
	tables, err := readTableDirectory(fontData, numTables)
	if err != nil {
		return nil, err
	}
	if err := checkTableBounds(fontData, tables); err != nil {
		return nil, err
	}

	_, hasGlyf := tables[tagGlyf]
	_, hasLoca := tables[tagLoca]
	_, hasHead := tables[tagHead]
	if !hasGlyf || !hasLoca || !hasHead {
		result := make([]byte, len(fontData))
		copy(result, fontData)
		return result, nil
	}

	keep := buildGlyphSet(fontData, tables, glyphIDs)

	out := make(map[string][]byte, len(subsetKeepTables))
	for tag, rec := range tables {
		if subsetKeepTables[tag] {
			out[tag] = fontData[rec.offset : rec.offset+rec.length]
		}
	}

	format := getLocaFormat(fontData, tables[tagHead])
	out[tagGlyf], out[tagLoca] = compactGlyf(fontData, tables, keep, format)
	if hmtx, ok := out[tagHmtx]; ok {
		out[tagHmtx] = zeroUnusedMetrics(hmtx, numberOfHMetrics(out[tagHhea]), keep)
	}
	if post, ok := out[tagPost]; ok && len(post) > postHeaderSize {
		out[tagPost] = stripGlyphNames(post)
	}
	if cmapping != nil {
		out[tagCmap] = buildUnicodeCmap(cmapping)
	}

	sfntVersion := binary.BigEndian.Uint32(fontData[0:4])
	return assembleFont(sfntVersion, out), nil
}

// subsetTableRecord is a table record used during subsetting.
type subsetTableRecord struct {
	tag    string
	offset uint32
	length uint32
}

// readTableDirectory reads the table directory from the font data.
func readTableDirectory(data []byte, numTables int) (map[string]subsetTableRecord, error) {
	tables := make(map[string]subsetTableRecord, numTables)
	offset := 12
	for i := 0; i < numTables; i++ {
		if offset+16 > len(data) {
			return nil, errors.New("font: truncated table directory during subsetting")
		}
		tag := string(data[offset : offset+4])
		tables[tag] = subsetTableRecord{
			tag:    tag,
			offset: binary.BigEndian.Uint32(data[offset+8 : offset+12]),
			length: binary.BigEndian.Uint32(data[offset+12 : offset+16]),
		}
		offset += 16
	}
	return tables, nil
}

// buildGlyphSet builds the complete set of glyph IDs to keep, including
// component glyphs referenced by composite glyphs.
func buildGlyphSet(data []byte, tables map[string]subsetTableRecord, glyphIDs []uint16) map[uint16]bool {
	keep := make(map[uint16]bool, len(glyphIDs))
	for _, gid := range glyphIDs {
		keep[gid] = true
	}

	// Resolve composite glyph dependencies.
	glyfRec, hasGlyf := tables[tagGlyf]
	locaRec, hasLoca := tables[tagLoca]
	headRec, hasHead := tables[tagHead]
	if !hasGlyf || !hasLoca || !hasHead {
		return keep
	}

	addCompositeComponents(data, glyfRec, locaRec, headRec, keep)
	return keep
}

// addCompositeComponents finds composite glyphs in the keep set and adds
// their component glyph IDs.
func addCompositeComponents(data []byte, glyfRec, locaRec, headRec subsetTableRecord, keep map[uint16]bool) {
	indexToLocFormat := getLocaFormat(data, headRec)

	// Iterate until no new glyphs are added.
	for {
		added := false
		for gid := range keep {
			glyfOff, glyfEnd := getGlyphOffsets(data, locaRec, indexToLocFormat, gid)
			if glyfOff >= glyfEnd {
				continue
			}

			absOff := int(glyfRec.offset) + glyfOff
			absEnd := int(glyfRec.offset) + glyfEnd
			if absOff+2 > len(data) {
				continue
			}

			numContours := int16(binary.BigEndian.Uint16(data[absOff : absOff+2]))
			if numContours >= 0 {
				continue // simple glyph
			}

			// Composite glyph: extract component glyph IDs.
			components := extractCompositeComponents(data, absOff, absEnd)
			for _, cid := range components {
				if !keep[cid] {
					keep[cid] = true
					added = true
				}
			}
		}
		if !added {
			break
		}
	}
}

// extractCompositeComponents parses a composite glyph and returns the
// glyph IDs of its components.
func extractCompositeComponents(data []byte, absOff, absEnd int) []uint16 {
	var components []uint16
	// Skip glyph header (10 bytes: numContours + xMin + yMin + xMax + yMax).
	pos := absOff + 10

	for pos+4 <= absEnd {
		flags := binary.BigEndian.Uint16(data[pos : pos+2])
		componentGID := binary.BigEndian.Uint16(data[pos+2 : pos+4])
		components = append(components, componentGID)
		pos += 4

		pos += compositeArgSize(flags)
		pos += compositeTransformSize(flags)

		if flags&0x0020 == 0 { // MORE_COMPONENTS flag
			break
		}
	}
	return components
}

// compositeArgSize returns the byte size of the composite glyph arguments
// based on the flags.
func compositeArgSize(flags uint16) int {
	if flags&0x0001 != 0 { // ARG_1_AND_2_ARE_WORDS
		return 4
	}
	return 2
}

// compositeTransformSize returns the byte size of the transform data
// based on the flags.
func compositeTransformSize(flags uint16) int {
	switch {
	case flags&0x0008 != 0: // WE_HAVE_A_SCALE
		return 2
	case flags&0x0040 != 0: // WE_HAVE_AN_X_AND_Y_SCALE
		return 4
	case flags&0x0080 != 0: // WE_HAVE_A_TWO_BY_TWO
		return 8
	default:
		return 0
	}
}

// getLocaFormat reads the indexToLocFormat from the head table.
func getLocaFormat(data []byte, headRec subsetTableRecord) int16 {
	off := int(headRec.offset) + 50
	if off+2 <= len(data) {
		return int16(binary.BigEndian.Uint16(data[off : off+2]))
	}
	return 0
}

// getGlyphOffsets returns the start and end offsets of a glyph within the
// glyf table, using the loca table for lookup.
func getGlyphOffsets(data []byte, locaRec subsetTableRecord, format int16, gid uint16) (int, int) {
	locaOff := int(locaRec.offset)
	if format == 0 {
		// Short format: offsets are uint16 * 2.
		idx := locaOff + int(gid)*2
		if idx+4 > len(data) {
			return 0, 0
		}
		start := int(binary.BigEndian.Uint16(data[idx:idx+2])) * 2
		end := int(binary.BigEndian.Uint16(data[idx+2:idx+4])) * 2
		return start, end
	}
	// Long format: offsets are uint32.
	idx := locaOff + int(gid)*4
	if idx+8 > len(data) {
		return 0, 0
	}
	start := int(binary.BigEndian.Uint32(data[idx : idx+4]))
	end := int(binary.BigEndian.Uint32(data[idx+4 : idx+8]))
	return start, end
}

// recalcTableChecksums recalculates checksums for all tables in the font.
func recalcTableChecksums(data []byte, numTables int) {
	offset := 12
	for i := 0; i < numTables; i++ {
		if offset+16 > len(data) {
			break
		}
		tblOffset := binary.BigEndian.Uint32(data[offset+8 : offset+12])
		tblLength := binary.BigEndian.Uint32(data[offset+12 : offset+16])

		checksum := calcTableChecksum(data, tblOffset, tblLength)
		binary.BigEndian.PutUint32(data[offset+4:offset+8], checksum)
		offset += 16
	}
}

// calcTableChecksum calculates the checksum for a table as defined
// by the TrueType specification.
func calcTableChecksum(data []byte, offset, length uint32) uint32 {
	var sum uint32
	end := offset + ((length + 3) &^ 3) // round up to 4-byte boundary
	if int(end) > len(data) {
		end = uint32(len(data))
	}
	for i := offset; i+4 <= end; i += 4 {
		sum += binary.BigEndian.Uint32(data[i : i+4])
	}
	return sum
}

// BuildSubsetCmap generates a simple cmap table (format 4) that maps
// a contiguous range of character codes to the given glyph IDs.
// This is useful for creating a cmap for a subsetted font where
// CIDs are assigned sequentially.
func BuildSubsetCmap(glyphIDs []uint16) []byte {
	if len(glyphIDs) == 0 {
		return nil
	}

	// Sort glyph IDs for deterministic output.
	sorted := make([]uint16, len(glyphIDs))
	copy(sorted, glyphIDs)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i] < sorted[j] })

	// Build a format 4 cmap: one segment per glyph mapping CID -> GID.
	// For simplicity, we create one segment covering [1..len] mapping to
	// the sorted glyph IDs, plus the sentinel segment.
	segCount := len(sorted) + 1 // segments + sentinel

	// Calculate table size.
	headerSize := 14              // format(2) + length(2) + language(2) + segCount*2(2) + searchRange(2) + entrySelector(2) + rangeShift(2)
	arraySize := segCount * 2 * 4 // endCode, startCode, idDelta, idRangeOffset (each segCount * 2 bytes)
	paddingSize := 2              // reservedPad
	glyphArraySize := len(sorted) * 2
	totalLength := headerSize + arraySize + paddingSize + glyphArraySize

	buf := make([]byte, totalLength)

	// Header.
	binary.BigEndian.PutUint16(buf[0:2], 4)                   // format
	binary.BigEndian.PutUint16(buf[2:4], uint16(totalLength)) // length
	binary.BigEndian.PutUint16(buf[4:6], 0)                   // language
	binary.BigEndian.PutUint16(buf[6:8], uint16(segCount*2))  // segCountX2

	// searchRange, entrySelector, rangeShift.
	sr := computeSearchRange(segCount)
	binary.BigEndian.PutUint16(buf[8:10], sr.searchRange)
	binary.BigEndian.PutUint16(buf[10:12], sr.entrySelector)
	binary.BigEndian.PutUint16(buf[12:14], sr.rangeShift)

	off := 14

	// endCode array.
	for i := 0; i < len(sorted); i++ {
		binary.BigEndian.PutUint16(buf[off+i*2:off+i*2+2], uint16(i+1))
	}
	binary.BigEndian.PutUint16(buf[off+(segCount-1)*2:off+(segCount-1)*2+2], 0xFFFF)
	off += segCount * 2

	// reservedPad.
	binary.BigEndian.PutUint16(buf[off:off+2], 0)
	off += 2

	// startCode array.
	for i := 0; i < len(sorted); i++ {
		binary.BigEndian.PutUint16(buf[off+i*2:off+i*2+2], uint16(i+1))
	}
	binary.BigEndian.PutUint16(buf[off+(segCount-1)*2:off+(segCount-1)*2+2], 0xFFFF)
	off += segCount * 2

	// idDelta array: map each CID directly via delta.
	for i := 0; i < len(sorted); i++ {
		delta := int16(sorted[i]) - int16(i+1)
		binary.BigEndian.PutUint16(buf[off+i*2:off+i*2+2], uint16(delta))
	}
	binary.BigEndian.PutUint16(buf[off+(segCount-1)*2:off+(segCount-1)*2+2], 1) // sentinel delta
	off += segCount * 2

	// idRangeOffset array: all zeros (using deltas).
	// Already zeroed by make().
	off += segCount * 2

	// glyphIDArray: not strictly needed when using deltas, but included
	// for completeness.
	for i := 0; i < len(sorted); i++ {
		binary.BigEndian.PutUint16(buf[off+i*2:off+i*2+2], sorted[i])
	}

	return buf
}

type searchRangeResult struct {
	searchRange   uint16
	entrySelector uint16
	rangeShift    uint16
}

func computeSearchRange(segCount int) searchRangeResult {
	var entrySelector uint16
	sr := 1
	for sr*2 <= segCount {
		sr *= 2
		entrySelector++
	}
	searchRange := uint16(sr * 2)
	rangeShift := uint16(segCount*2) - searchRange
	return searchRangeResult{
		searchRange:   searchRange,
		entrySelector: entrySelector,
		rangeShift:    rangeShift,
	}
}

// ValidateTrueType performs basic validation of TrueType font data.
// It checks for the presence of required tables and basic structural
// integrity. Returns an error describing any issues found.
func ValidateTrueType(data []byte) error {
	if len(data) < 12 {
		return errors.New("font: data too short for TrueType font")
	}

	scaler := binary.BigEndian.Uint32(data[0:4])
	switch scaler {
	case 0x00010000, 0x74727565, 0x4F54544F:
		// Valid scaler types.
	default:
		return fmt.Errorf("font: unsupported scaler type 0x%08X", scaler)
	}

	numTables := int(binary.BigEndian.Uint16(data[4:6]))
	required := []string{tagHead, tagHhea, tagMaxp, tagCmap, tagHmtx}

	tables, err := readTableDirectory(data, numTables)
	if err != nil {
		return err
	}

	for _, tag := range required {
		if _, ok := tables[tag]; !ok {
			return fmt.Errorf("font: missing required table '%s'", tag)
		}
	}

	return nil
}
