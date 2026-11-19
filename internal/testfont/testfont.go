// Package testfont builds minimal TrueType fonts in memory for tests, so
// that tests need no font files.
package testfont

import "encoding/binary"

// ---------------------------------------------------------------------------
// Minimal TTF builder for testing (no external font file needed)
// ---------------------------------------------------------------------------

// putI16 writes an int16 as big-endian bytes.
func putI16(b []byte, v int16) {
	binary.BigEndian.PutUint16(b, uint16(v))
}

// Build builds a minimal but parseable TrueType font binary with
// the given rune-to-GID mapping and per-glyph widths. Includes loca and glyf
// tables so that Subset works.
func Build(numGlyphs int, widths []uint16, runeMap map[rune]uint16) []byte {
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
