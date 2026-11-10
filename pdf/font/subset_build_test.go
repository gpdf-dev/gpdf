package font

import (
	"bytes"
	"encoding/binary"
	"testing"
)

// tableMap parses the table directory of a font for assertions.
func tableMap(t *testing.T, data []byte) map[string][]byte {
	t.Helper()
	n := int(binary.BigEndian.Uint16(data[4:6]))
	recs, err := readTableDirectory(data, n)
	if err != nil {
		t.Fatalf("readTableDirectory: %v", err)
	}
	out := make(map[string][]byte, n)
	for tag, r := range recs {
		out[tag] = data[r.offset : r.offset+r.length]
	}
	return out
}

// withExtraTables rebuilds a test font with additional tables appended.
func withExtraTables(t *testing.T, data []byte, extra ...tableEntry) []byte {
	t.Helper()
	var entries []tableEntry
	for tag, b := range tableMap(t, data) {
		entries = append(entries, tableEntry{tag, b})
	}
	return assembleTTF(append(entries, extra...))
}

func TestSubsetTrueType_DropsUnneededTables(t *testing.T) {
	base := buildMinimalTTFWithLocaGlyf(4, []uint16{0, 600, 700, 800}, map[rune]uint16{'A': 1, 'B': 2, 'C': 3})
	data := withExtraTables(t, base,
		tableEntry{"gvar", bytes.Repeat([]byte{0xAB}, 4096)},
		tableEntry{"GSUB", bytes.Repeat([]byte{0xCD}, 256)},
		tableEntry{"vmtx", bytes.Repeat([]byte{0x01}, 16)},
		tableEntry{"OS/2", make([]byte, 96)},
		tableEntry{"prep", []byte{0xB0, 0x00}},
	)

	result, err := SubsetTrueType(data, []uint16{0, 1})
	if err != nil {
		t.Fatalf("SubsetTrueType: %v", err)
	}
	tables := tableMap(t, result)
	for _, tag := range []string{"gvar", "GSUB", "vmtx"} {
		if _, ok := tables[tag]; ok {
			t.Errorf("table %q should have been dropped", tag)
		}
	}
	for _, tag := range []string{tagHead, tagHhea, tagMaxp, tagLoca, tagGlyf, tagHmtx, tagCmap, tagPost, tagName, "OS/2", "prep"} {
		if _, ok := tables[tag]; !ok {
			t.Errorf("table %q should have been kept", tag)
		}
	}
	if _, err := ParseTrueType(result); err != nil {
		t.Errorf("subset font does not parse: %v", err)
	}
}

func TestSubsetTrueType_CompactsGlyfAndLoca(t *testing.T) {
	data := buildMinimalTTFWithLocaGlyf(4, []uint16{0, 600, 700, 800}, map[rune]uint16{'A': 1, 'B': 2, 'C': 3})
	orig := tableMap(t, data)

	result, err := SubsetTrueType(data, []uint16{0, 2})
	if err != nil {
		t.Fatalf("SubsetTrueType: %v", err)
	}
	tables := tableMap(t, result)

	if got, want := len(tables[tagGlyf]), 24; got != want {
		t.Errorf("glyf length = %d, want %d (two 12-byte glyphs)", got, want)
	}
	loca := tables[tagLoca]
	wantLoca := []uint32{0, 12, 12, 24, 24} // gid0, gid1 (empty), gid2, gid3 (empty), end
	for i, w := range wantLoca {
		if got := binary.BigEndian.Uint32(loca[i*4:]); got != w {
			t.Errorf("loca[%d] = %d, want %d", i, got, w)
		}
	}
	if !bytes.Equal(tables[tagGlyf][12:24], orig[tagGlyf][24:36]) {
		t.Error("kept glyph 2 outline was not copied verbatim")
	}

	hmtx := tables[tagHmtx]
	if got := binary.BigEndian.Uint16(hmtx[2*4:]); got != 700 {
		t.Errorf("kept glyph advance = %d, want 700", got)
	}
	if got := binary.BigEndian.Uint16(hmtx[1*4:]); got != 0 {
		t.Errorf("unused glyph advance = %d, want 0", got)
	}
}

func TestSubsetTrueType_ShortLoca(t *testing.T) {
	data := buildMinimalTTFWithLocaGlyf(3, []uint16{0, 600, 700}, map[rune]uint16{'A': 1, 'B': 2})
	tables := tableMap(t, data)
	head := append([]byte(nil), tables[tagHead]...)
	putI16(head[50:52], 0)
	loca := make([]byte, 4*2)
	for i := 0; i <= 3; i++ {
		binary.BigEndian.PutUint16(loca[i*2:], uint16(i*12/2))
	}
	var entries []tableEntry
	for tag, b := range tables {
		switch tag {
		case tagHead:
			b = head
		case tagLoca:
			b = loca
		}
		entries = append(entries, tableEntry{tag, b})
	}
	data = assembleTTF(entries)

	result, err := SubsetTrueType(data, []uint16{0, 2})
	if err != nil {
		t.Fatalf("SubsetTrueType: %v", err)
	}
	got := tableMap(t, result)[tagLoca]
	want := []uint16{0, 6, 6, 12}
	for i, w := range want {
		if v := binary.BigEndian.Uint16(got[i*2:]); v != w {
			t.Errorf("short loca[%d] = %d, want %d", i, v, w)
		}
	}
}

func TestSubsetTrueType_Checksums(t *testing.T) {
	data := buildMinimalTTFWithLocaGlyf(4, []uint16{0, 600, 700, 800}, map[rune]uint16{'A': 1, 'B': 2, 'C': 3})
	result, err := SubsetTrueType(data, []uint16{0, 1})
	if err != nil {
		t.Fatalf("SubsetTrueType: %v", err)
	}
	if len(result)%4 != 0 {
		t.Fatalf("font length %d not 4-byte aligned", len(result))
	}
	if sum := calcTableChecksum(result, 0, uint32(len(result))); sum != 0xB1B0AFBA {
		t.Errorf("whole-font checksum = 0x%08X, want 0xB1B0AFBA", sum)
	}
	n := int(binary.BigEndian.Uint16(result[4:6]))
	for i := 0; i < n; i++ {
		rec := result[12+16*i:]
		tag := string(rec[0:4])
		if tag == tagHead {
			continue // head checksum is computed with checkSumAdjustment = 0
		}
		want := binary.BigEndian.Uint32(rec[4:8])
		got := calcTableChecksum(result, binary.BigEndian.Uint32(rec[8:12]), binary.BigEndian.Uint32(rec[12:16]))
		if got != want {
			t.Errorf("table %q checksum = 0x%08X, want 0x%08X", tag, got, want)
		}
	}
}

func TestSubsetTrueType_StripsGlyphNames(t *testing.T) {
	base := buildMinimalTTFWithLocaGlyf(2, []uint16{0, 600}, map[rune]uint16{'A': 1})
	tables := tableMap(t, base)
	post := make([]byte, 64)
	binary.BigEndian.PutUint32(post[0:4], 0x00020000)
	binary.BigEndian.PutUint32(post[4:8], 0xFFF40000) // italicAngle -12
	var entries []tableEntry
	for tag, b := range tables {
		if tag == tagPost {
			b = post
		}
		entries = append(entries, tableEntry{tag, b})
	}

	result, err := SubsetTrueType(assembleTTF(entries), []uint16{0, 1})
	if err != nil {
		t.Fatalf("SubsetTrueType: %v", err)
	}
	got := tableMap(t, result)[tagPost]
	if len(got) != postHeaderSize {
		t.Fatalf("post length = %d, want %d", len(got), postHeaderSize)
	}
	if v := binary.BigEndian.Uint32(got[0:4]); v != 0x00030000 {
		t.Errorf("post version = 0x%08X, want 0x00030000", v)
	}
	if v := binary.BigEndian.Uint32(got[4:8]); v != 0xFFF40000 {
		t.Errorf("italicAngle not preserved: 0x%08X", v)
	}
}

func TestTrueTypeFont_Subset_RebuildsCmap(t *testing.T) {
	data := buildMinimalTTFWithLocaGlyf(4, []uint16{0, 600, 700, 800}, map[rune]uint16{'A': 1, 'B': 2, 'C': 3})
	ttf, err := ParseTrueType(data)
	if err != nil {
		t.Fatal(err)
	}
	result, err := ttf.Subset([]rune{'A', 'C'})
	if err != nil {
		t.Fatalf("Subset: %v", err)
	}
	sub, err := ParseTrueType(result)
	if err != nil {
		t.Fatalf("subset font does not parse: %v", err)
	}
	for r, want := range map[rune]uint16{'A': 1, 'C': 3, 'B': 0} {
		if got, _ := sub.cmapTbl.lookup(r); got != want {
			t.Errorf("cmap[%q] = %d, want %d", r, got, want)
		}
	}
}

func TestBuildUnicodeCmap_Supplementary(t *testing.T) {
	m := map[rune]uint16{'a': 10, 'b': 11, 'd': 20, 0x20000: 30, 0x20001: 31}
	ttf := &TrueTypeFont{}
	cmap := buildUnicodeCmap(m)
	tables := map[string]tableRecord{tagCmap: {Offset: 0, Length: uint32(len(cmap))}}
	if err := ttf.parseCmap(cmap, tables); err != nil {
		t.Fatalf("parseCmap: %v", err)
	}
	if ttf.cmapTbl.format4 == nil || ttf.cmapTbl.format12 == nil {
		t.Fatal("expected both format 4 and format 12 subtables")
	}
	for r, want := range m {
		if got, _ := ttf.cmapTbl.lookup(r); got != want {
			t.Errorf("lookup(%U) = %d, want %d", r, got, want)
		}
	}
	if got := ttf.cmapTbl.format4.lookup('c'); got != 0 {
		t.Errorf("unmapped 'c' -> %d, want 0", got)
	}
}

func TestCmapRanges_Grouping(t *testing.T) {
	m := map[rune]uint16{'a': 1, 'b': 2, 'c': 3, 'e': 5, 'f': 9}
	got := cmapRanges(m, func(rune) bool { return true })
	want := []cmapRange{{'a', 'c', 1}, {'e', 'e', 5}, {'f', 'f', 9}}
	if len(got) != len(want) {
		t.Fatalf("ranges = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("range[%d] = %v, want %v", i, got[i], want[i])
		}
	}
}

func TestSubsetTrueType_TableOutOfBounds(t *testing.T) {
	data := buildMinimalTTFWithLocaGlyf(2, []uint16{0, 600}, map[rune]uint16{'A': 1})
	// Corrupt the first table's length.
	binary.BigEndian.PutUint32(data[12+12:12+16], 0xFFFFFF)
	if _, err := SubsetTrueType(data, []uint16{0}); err == nil {
		t.Error("expected error for out-of-bounds table")
	}
}
