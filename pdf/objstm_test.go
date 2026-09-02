package pdf

import (
	"bytes"
	"compress/zlib"
	"fmt"
	"strings"
	"testing"
)

// buildObjStmPDF creates a PDF 1.5 whose catalog, page tree and page objects all
// live inside a compressed object stream, referenced by type-2 cross-reference
// entries — the layout most modern producers emit by default.
//
// Object layout:
//
//	1, 2, 3 — catalog / pages / page, stored inside the object stream
//	4       — the /ObjStm itself
//	5       — the /XRef stream
func buildObjStmPDF(t *testing.T, compressStm bool) []byte {
	t.Helper()

	objects := []struct {
		num  int
		body string
	}{
		{1, "<< /Type /Catalog /Pages 2 0 R >>"},
		{2, "<< /Type /Pages /Kids [3 0 R] /Count 1 >>"},
		{3, "<< /Type /Page /Parent 2 0 R /MediaBox [0 0 612 792] >>"},
	}

	// Build the object stream body first so the header offsets are known.
	var header, body strings.Builder
	for _, o := range objects {
		fmt.Fprintf(&header, "%d %d ", o.num, body.Len())
		body.WriteString(o.body)
		body.WriteString("\n")
	}
	first := header.Len()
	objStmContent := []byte(header.String() + body.String())

	objStmFilter := ""
	if compressStm {
		var compressed bytes.Buffer
		zw := zlib.NewWriter(&compressed)
		_, _ = zw.Write(objStmContent)
		_ = zw.Close()
		objStmContent = compressed.Bytes()
		objStmFilter = " /Filter /FlateDecode"
	}

	var pdf bytes.Buffer
	pdf.WriteString("%PDF-1.5\n")

	objStmOffset := pdf.Len()
	fmt.Fprintf(&pdf, "4 0 obj\n<< /Type /ObjStm /N %d /First %d /Length %d%s >>\nstream\n",
		len(objects), first, len(objStmContent), objStmFilter)
	pdf.Write(objStmContent)
	pdf.WriteString("\nendstream\nendobj\n")

	xrefOffset := pdf.Len()

	// /W [1 2 1]: type, 2-byte field, 1-byte field.
	var xrefContent bytes.Buffer
	writeEntry := func(typ byte, f2 int, f3 byte) {
		xrefContent.WriteByte(typ)
		xrefContent.WriteByte(byte(f2 >> 8))
		xrefContent.WriteByte(byte(f2 & 0xFF))
		xrefContent.WriteByte(f3)
	}
	writeEntry(0, 0, 0)            // obj 0: free
	writeEntry(2, 4, 0)            // obj 1: in object stream 4, index 0
	writeEntry(2, 4, 1)            // obj 2: in object stream 4, index 1
	writeEntry(2, 4, 2)            // obj 3: in object stream 4, index 2
	writeEntry(1, objStmOffset, 0) // obj 4: the object stream
	writeEntry(1, xrefOffset, 0)   // obj 5: the xref stream itself

	fmt.Fprintf(&pdf, "5 0 obj\n<< /Type /XRef /Size 6 /W [1 2 1] /Root 1 0 R /Length %d >>\nstream\n", xrefContent.Len())
	pdf.Write(xrefContent.Bytes())
	pdf.WriteString("\nendstream\nendobj\n")

	fmt.Fprintf(&pdf, "startxref\n%d\n%%%%EOF\n", xrefOffset)
	return pdf.Bytes()
}

func TestReaderObjectStream(t *testing.T) {
	for _, tc := range []struct {
		name     string
		compress bool
	}{
		{"FlateDecode", true},
		{"uncompressed", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			r, err := NewReader(buildObjStmPDF(t, tc.compress))
			if err != nil {
				t.Fatalf("NewReader: %v", err)
			}

			// The catalog itself is a type-2 object, so this exercises the
			// compressed path during construction.
			count, err := r.PageCount()
			if err != nil {
				t.Fatalf("PageCount: %v", err)
			}
			if count != 1 {
				t.Fatalf("PageCount = %d, want 1", count)
			}

			page, err := r.PageDict(0)
			if err != nil {
				t.Fatalf("PageDict: %v", err)
			}
			if typ, _ := page[Name("Type")].(Name); typ != Name("Page") {
				t.Errorf("page /Type = %v, want /Page", page[Name("Type")])
			}

			info, err := r.Page(0)
			if err != nil {
				t.Fatalf("Page: %v", err)
			}
			if info.MediaBox.URX != 612 || info.MediaBox.URY != 792 {
				t.Errorf("MediaBox = %+v, want 612x792", info.MediaBox)
			}

			// Object 3 lives in the object stream; object 4 is the stream itself.
			if max := r.MaxObjectNumber(); max != 5 {
				t.Errorf("MaxObjectNumber = %d, want 5", max)
			}
		})
	}
}

func TestReaderObjectStreamCaching(t *testing.T) {
	r, err := NewReader(buildObjStmPDF(t, true))
	if err != nil {
		t.Fatalf("NewReader: %v", err)
	}

	// Resolving three sibling objects must decode the containing stream once.
	for _, num := range []int{1, 2, 3} {
		if _, err := r.GetObject(num); err != nil {
			t.Fatalf("GetObject(%d): %v", num, err)
		}
	}
	if len(r.objStms) != 1 {
		t.Errorf("decoded object streams = %d, want 1", len(r.objStms))
	}
	if _, ok := r.objStms[4]; !ok {
		t.Error("object stream 4 not cached")
	}
}

func TestMergeObjectStreamPDF(t *testing.T) {
	data := buildObjStmPDF(t, true)

	merged, err := MergePDFs([]MergeSource{{Data: data}, {Data: data}}, MergeConfig{})
	if err != nil {
		t.Fatalf("MergePDFs: %v", err)
	}

	r, err := NewReader(merged)
	if err != nil {
		t.Fatalf("NewReader(merged): %v", err)
	}
	count, err := r.PageCount()
	if err != nil {
		t.Fatalf("PageCount: %v", err)
	}
	if count != 2 {
		t.Errorf("merged PageCount = %d, want 2", count)
	}
}

func TestReaderObjectStreamStaleIndex(t *testing.T) {
	r, err := NewReader(buildObjStmPDF(t, true))
	if err != nil {
		t.Fatalf("NewReader: %v", err)
	}

	// Point object 3 at an index that belongs to a different object. The lookup
	// must fall back to matching by object number rather than returning object 1.
	r.compressed[3] = compressedObjRef{streamNum: 4, index: 0}
	obj, err := r.GetObject(3)
	if err != nil {
		t.Fatalf("GetObject(3): %v", err)
	}
	d, ok := obj.(Dict)
	if !ok {
		t.Fatalf("GetObject(3) = %T, want Dict", obj)
	}
	if typ, _ := d[Name("Type")].(Name); typ != Name("Page") {
		t.Errorf("/Type = %v, want /Page", d[Name("Type")])
	}
}

func TestReaderObjectStreamMissingObject(t *testing.T) {
	r, err := NewReader(buildObjStmPDF(t, true))
	if err != nil {
		t.Fatalf("NewReader: %v", err)
	}

	r.compressed[99] = compressedObjRef{streamNum: 4, index: 0}
	if _, err := r.GetObject(99); err == nil {
		t.Fatal("GetObject(99) succeeded, want error")
	}
}

func TestReaderObjectStreamNotAStream(t *testing.T) {
	r, err := NewReader(buildObjStmPDF(t, true))
	if err != nil {
		t.Fatalf("NewReader: %v", err)
	}

	// Object 1 is a plain dict, not an /ObjStm.
	r.compressed[99] = compressedObjRef{streamNum: 1, index: 0}
	_, err = r.GetObject(99)
	if err == nil || !strings.Contains(err.Error(), "not a stream") {
		t.Fatalf("err = %v, want 'not a stream'", err)
	}
}

func TestParseXRefEntriesType2(t *testing.T) {
	r := &Reader{
		xref:       make(map[int]int64),
		compressed: make(map[int]compressedObjRef),
	}

	// w=[1,2,1], entrySize=4.
	content := []byte{
		2, 0x00, 0x0A, 3, // obj 5: type=2, stream 10, index 3
		1, 0x01, 0x00, 0, // obj 6: type=1, offset 256
		0, 0x00, 0x00, 0, // obj 7: free
	}
	r.parseXRefEntries(content, []int{5, 3}, [3]int{1, 2, 1}, 4)

	want := compressedObjRef{streamNum: 10, index: 3}
	if got := r.compressed[5]; got != want {
		t.Errorf("compressed[5] = %+v, want %+v", got, want)
	}
	if _, ok := r.xref[5]; ok {
		t.Error("xref[5] should not be set for a type-2 entry")
	}
	if r.xref[6] != 256 {
		t.Errorf("xref[6] = %d, want 256", r.xref[6])
	}
	if _, ok := r.compressed[7]; ok {
		t.Error("compressed[7] should not exist (free entry)")
	}
}

func TestParseXRefEntriesNewestWins(t *testing.T) {
	r := &Reader{
		xref:       map[int]int64{5: 100},
		compressed: map[int]compressedObjRef{6: {streamNum: 1, index: 0}},
	}

	// An older /Prev section must not override entries already recorded.
	content := []byte{
		2, 0x00, 0x0A, 3, // obj 5: type=2 — must not displace xref[5]
		2, 0x00, 0x0B, 7, // obj 6: type=2 — must not displace compressed[6]
	}
	r.parseXRefEntries(content, []int{5, 2}, [3]int{1, 2, 1}, 4)

	if r.xref[5] != 100 {
		t.Errorf("xref[5] = %d, want 100", r.xref[5])
	}
	if _, ok := r.compressed[5]; ok {
		t.Error("compressed[5] should not have been set")
	}
	if got := r.compressed[6]; got.streamNum != 1 || got.index != 0 {
		t.Errorf("compressed[6] = %+v, want {1 0}", got)
	}
}

func TestParseObjStmHeader(t *testing.T) {
	header := []byte("1 0 2 10 3 25 ")
	entries, err := parseObjStmHeader(header, 3, 20, 60)
	if err != nil {
		t.Fatalf("parseObjStmHeader: %v", err)
	}

	want := []objStmEntry{
		{num: 1, start: 20, end: 30},
		{num: 2, start: 30, end: 45},
		{num: 3, start: 45, end: 60},
	}
	for i, w := range want {
		if entries[i] != w {
			t.Errorf("entries[%d] = %+v, want %+v", i, entries[i], w)
		}
	}
}

func TestParseObjStmHeaderErrors(t *testing.T) {
	tests := []struct {
		name    string
		header  string
		n       int
		first   int
		bodyLen int
	}{
		{"negative N", "", -1, 0, 0},
		{"truncated header", "1 0 2", 3, 10, 40},
		{"offset past body", "1 999", 1, 10, 40},
		{"non-integer offset", "1 1.5", 1, 10, 40},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := parseObjStmHeader([]byte(tc.header), tc.n, tc.first, tc.bodyLen); err == nil {
				t.Fatal("expected error, got nil")
			}
		})
	}
}
