package font

import (
	"encoding/binary"
	"fmt"
	"sort"
)

// subsetKeepTables lists the tables retained in a subset font. These are
// the tables required for a PDF FontFile2 (ISO 32000-2 §9.9) plus name,
// post and OS/2, which some consumers expect. Everything else — GSUB/GPOS/
// GDEF, variation tables (fvar/gvar/HVAR/...), vertical metrics, color
// tables, kerning — is never consulted when rendering a CIDFontType2 with
// explicit /W widths and is dropped.
var subsetKeepTables = map[string]bool{
	tagHead: true,
	tagHhea: true,
	tagMaxp: true,
	tagLoca: true,
	tagGlyf: true,
	tagHmtx: true,
	tagCmap: true,
	tagPost: true,
	tagName: true,
	"OS/2":  true,
	"cvt ":  true,
	"fpgm":  true,
	"prep":  true,
}

// postHeaderSize is the size of a version 3.0 post table (no glyph names).
const postHeaderSize = 32

// checkTableBounds verifies that every table lies within the font data.
func checkTableBounds(data []byte, tables map[string]subsetTableRecord) error {
	for tag, rec := range tables {
		if uint64(rec.offset)+uint64(rec.length) > uint64(len(data)) {
			return fmt.Errorf("font: table '%s' extends past end of data", tag)
		}
	}
	return nil
}

// compactGlyf builds new glyf and loca tables that contain only the
// outlines of kept glyphs. Glyph IDs are unchanged: unused glyphs become
// empty (zero-length) entries. The loca format of the source font is
// preserved; since the new glyf is never larger than the original, short
// offsets cannot overflow.
func compactGlyf(data []byte, tables map[string]subsetTableRecord, keep map[uint16]bool, format int16) (glyf, loca []byte) {
	glyfRec, locaRec := tables[tagGlyf], tables[tagLoca]
	entry, align := 4, 4
	if format == 0 {
		entry, align = 2, 2
	}
	numGlyphs := int(locaRec.length)/entry - 1
	if numGlyphs < 0 {
		numGlyphs = 0
	}

	loca = make([]byte, (numGlyphs+1)*entry)
	for gid := 0; gid < numGlyphs; gid++ {
		putLocaEntry(loca, format, gid, len(glyf))
		if !keep[uint16(gid)] {
			continue
		}
		start, end := getGlyphOffsets(data, locaRec, format, uint16(gid))
		if start >= end || uint32(end) > glyfRec.length {
			continue
		}
		glyf = append(glyf, data[int(glyfRec.offset)+start:int(glyfRec.offset)+end]...)
		for len(glyf)%align != 0 {
			glyf = append(glyf, 0)
		}
	}
	putLocaEntry(loca, format, numGlyphs, len(glyf))
	return glyf, loca
}

func putLocaEntry(loca []byte, format int16, idx, offset int) {
	if format == 0 {
		binary.BigEndian.PutUint16(loca[idx*2:], uint16(offset/2))
		return
	}
	binary.BigEndian.PutUint32(loca[idx*4:], uint32(offset))
}

// numberOfHMetrics reads hhea.numberOfHMetrics.
func numberOfHMetrics(hhea []byte) int {
	if len(hhea) < 36 {
		return 0
	}
	return int(binary.BigEndian.Uint16(hhea[34:36]))
}

// zeroUnusedMetrics returns a copy of hmtx with the entries of unused
// glyphs zeroed. PDF viewers take advances from the /W array, so the
// values only matter for kept glyphs; zeroing the rest lets Flate compress
// the table to almost nothing.
func zeroUnusedMetrics(hmtx []byte, numHMetrics int, keep map[uint16]bool) []byte {
	out := make([]byte, len(hmtx))
	for gid := range keep {
		g := int(gid)
		if g < numHMetrics {
			if off := g * 4; off+4 <= len(hmtx) {
				copy(out[off:off+4], hmtx[off:off+4])
			}
			continue
		}
		// Glyphs past numberOfHMetrics only store a left side bearing.
		if off := numHMetrics*4 + (g-numHMetrics)*2; off+2 <= len(hmtx) {
			copy(out[off:off+2], hmtx[off:off+2])
		}
	}
	// The last full metric's advance applies to all following glyphs, so
	// keep it intact when any such glyph is used.
	if numHMetrics > 0 && numHMetrics*4 <= len(hmtx) {
		last := (numHMetrics - 1) * 4
		copy(out[last:last+2], hmtx[last:last+2])
	}
	return out
}

// stripGlyphNames converts a post table to version 3.0 (no glyph names),
// keeping the header fields (italic angle, underline metrics, ...).
func stripGlyphNames(post []byte) []byte {
	out := make([]byte, postHeaderSize)
	copy(out, post[:postHeaderSize])
	binary.BigEndian.PutUint32(out[0:4], 0x00030000)
	return out
}

// cmapRange is a run of consecutive characters mapped to consecutive GIDs.
type cmapRange struct {
	start, end rune
	startGID   uint16
}

// cmapRanges groups a rune → GID mapping into maximal consecutive runs,
// keeping only runes accepted by include.
func cmapRanges(m map[rune]uint16, include func(rune) bool) []cmapRange {
	runes := make([]rune, 0, len(m))
	for r := range m {
		if include(r) {
			runes = append(runes, r)
		}
	}
	sort.Slice(runes, func(i, j int) bool { return runes[i] < runes[j] })

	var ranges []cmapRange
	for _, r := range runes {
		gid := m[r]
		if n := len(ranges); n > 0 {
			last := &ranges[n-1]
			if r == last.end+1 && int(gid) == int(last.startGID)+int(r-last.start) {
				last.end = r
				continue
			}
		}
		ranges = append(ranges, cmapRange{start: r, end: r, startGID: gid})
	}
	return ranges
}

// buildUnicodeCmap builds a cmap table mapping exactly the given runes.
// It contains a (3,1) format 4 subtable for BMP characters and, when any
// supplementary-plane character is present, a (3,10) format 12 subtable
// covering all characters.
func buildUnicodeCmap(m map[rune]uint16) []byte {
	bmp := buildCmapFormat4(cmapRanges(m, func(r rune) bool { return r < 0xFFFF }))
	subtables := [][]byte{bmp}
	encodings := []uint16{1}
	if full := cmapRanges(m, func(rune) bool { return true }); len(full) > 0 && full[len(full)-1].end > 0xFFFF {
		subtables = append(subtables, buildCmapFormat12(full))
		encodings = append(encodings, 10)
	}

	header := 4 + 8*len(subtables)
	buf := make([]byte, header)
	binary.BigEndian.PutUint16(buf[2:4], uint16(len(subtables)))
	offset := header
	for i, st := range subtables {
		rec := buf[4+8*i:]
		binary.BigEndian.PutUint16(rec[0:2], 3) // platform: Windows
		binary.BigEndian.PutUint16(rec[2:4], encodings[i])
		binary.BigEndian.PutUint32(rec[4:8], uint32(offset))
		offset += len(st)
	}
	for _, st := range subtables {
		buf = append(buf, st...)
	}
	return buf
}

// buildCmapFormat4 encodes BMP ranges as a format 4 subtable using idDelta.
func buildCmapFormat4(ranges []cmapRange) []byte {
	segCount := len(ranges) + 1 // + 0xFFFF sentinel
	length := 16 + segCount*8
	buf := make([]byte, length)
	binary.BigEndian.PutUint16(buf[0:2], 4)
	binary.BigEndian.PutUint16(buf[2:4], uint16(length))
	binary.BigEndian.PutUint16(buf[6:8], uint16(segCount*2))
	sr := computeSearchRange(segCount)
	binary.BigEndian.PutUint16(buf[8:10], sr.searchRange)
	binary.BigEndian.PutUint16(buf[10:12], sr.entrySelector)
	binary.BigEndian.PutUint16(buf[12:14], sr.rangeShift)

	endCodes := buf[14:]
	startCodes := buf[16+segCount*2:]
	deltas := buf[16+segCount*4:]
	// idRangeOffset (buf[16+segCount*6:]) stays zero.
	for i, rg := range ranges {
		binary.BigEndian.PutUint16(endCodes[i*2:], uint16(rg.end))
		binary.BigEndian.PutUint16(startCodes[i*2:], uint16(rg.start))
		binary.BigEndian.PutUint16(deltas[i*2:], rg.startGID-uint16(rg.start))
	}
	s := segCount - 1
	binary.BigEndian.PutUint16(endCodes[s*2:], 0xFFFF)
	binary.BigEndian.PutUint16(startCodes[s*2:], 0xFFFF)
	binary.BigEndian.PutUint16(deltas[s*2:], 1)
	return buf
}

// buildCmapFormat12 encodes ranges as a format 12 (segmented coverage)
// subtable.
func buildCmapFormat12(ranges []cmapRange) []byte {
	length := 16 + 12*len(ranges)
	buf := make([]byte, length)
	binary.BigEndian.PutUint16(buf[0:2], 12)
	binary.BigEndian.PutUint32(buf[4:8], uint32(length))
	binary.BigEndian.PutUint32(buf[12:16], uint32(len(ranges)))
	for i, rg := range ranges {
		g := buf[16+12*i:]
		binary.BigEndian.PutUint32(g[0:4], uint32(rg.start))
		binary.BigEndian.PutUint32(g[4:8], uint32(rg.end))
		binary.BigEndian.PutUint32(g[8:12], uint32(rg.startGID))
	}
	return buf
}

// assembleFont serializes tables into an sfnt file with a sorted table
// directory, 4-byte aligned table data, correct table checksums and
// head.checkSumAdjustment.
func assembleFont(sfntVersion uint32, tables map[string][]byte) []byte {
	tags := make([]string, 0, len(tables))
	for tag := range tables {
		tags = append(tags, tag)
	}
	sort.Strings(tags)

	if head, ok := tables[tagHead]; ok && len(head) >= 12 {
		h := make([]byte, len(head))
		copy(h, head)
		binary.BigEndian.PutUint32(h[8:12], 0) // checkSumAdjustment
		tables[tagHead] = h
	}

	n := len(tags)
	dirSize := 12 + 16*n
	size := dirSize
	for _, tag := range tags {
		size += (len(tables[tag]) + 3) &^ 3
	}
	buf := make([]byte, dirSize, size)

	binary.BigEndian.PutUint32(buf[0:4], sfntVersion)
	binary.BigEndian.PutUint16(buf[4:6], uint16(n))
	entrySelector := 0
	for 1<<(entrySelector+1) <= n {
		entrySelector++
	}
	searchRange := (1 << entrySelector) * 16
	binary.BigEndian.PutUint16(buf[6:8], uint16(searchRange))
	binary.BigEndian.PutUint16(buf[8:10], uint16(entrySelector))
	binary.BigEndian.PutUint16(buf[10:12], uint16(n*16-searchRange))

	headOffset := -1
	for i, tag := range tags {
		data := tables[tag]
		offset := len(buf)
		if tag == tagHead {
			headOffset = offset
		}
		buf = append(buf, data...)
		for len(buf)%4 != 0 {
			buf = append(buf, 0)
		}
		rec := buf[12+16*i:]
		copy(rec[0:4], tag)
		binary.BigEndian.PutUint32(rec[4:8], calcTableChecksum(buf, uint32(offset), uint32(len(data))))
		binary.BigEndian.PutUint32(rec[8:12], uint32(offset))
		binary.BigEndian.PutUint32(rec[12:16], uint32(len(data)))
	}

	if headOffset >= 0 {
		total := calcTableChecksum(buf, 0, uint32(len(buf)))
		binary.BigEndian.PutUint32(buf[headOffset+8:headOffset+12], 0xB1B0AFBA-total)
	}
	return buf
}
