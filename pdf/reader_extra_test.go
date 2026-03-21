package pdf

import (
	"bytes"
	"compress/zlib"
	"fmt"
	"testing"
)

func TestReaderTrailer(t *testing.T) {
	data := buildTestPDF(t, 1)
	r, err := NewReader(data)
	if err != nil {
		t.Fatalf("NewReader: %v", err)
	}
	trailer := r.Trailer()
	if trailer == nil {
		t.Fatal("Trailer() returned nil")
	}
	// Trailer should have /Root and /Size at minimum.
	if _, ok := trailer[Name("Root")]; !ok {
		t.Error("trailer missing /Root")
	}
	if _, ok := trailer[Name("Size")]; !ok {
		t.Error("trailer missing /Size")
	}
}

func TestReaderPageDict(t *testing.T) {
	data := buildTestPDF(t, 2)
	r, err := NewReader(data)
	if err != nil {
		t.Fatalf("NewReader: %v", err)
	}

	// Page 0 should return a valid dict.
	d, err := r.PageDict(0)
	if err != nil {
		t.Fatalf("PageDict(0): %v", err)
	}
	if d == nil {
		t.Fatal("PageDict(0) returned nil")
	}
	// Should have /Type = /Page.
	if typeName, ok := d[Name("Type")].(Name); !ok || typeName != "Page" {
		t.Errorf("page dict /Type = %v, want Page", d[Name("Type")])
	}

	// Page 1 should also work.
	d1, err := r.PageDict(1)
	if err != nil {
		t.Fatalf("PageDict(1): %v", err)
	}
	if d1 == nil {
		t.Fatal("PageDict(1) returned nil")
	}
}

func TestReaderPageDictOutOfRange(t *testing.T) {
	data := buildTestPDF(t, 1)
	r, err := NewReader(data)
	if err != nil {
		t.Fatalf("NewReader: %v", err)
	}
	_, err = r.PageDict(5)
	if err == nil {
		t.Error("expected error for out-of-range page")
	}
}

func TestReaderResolveDictFromStream(t *testing.T) {
	data := buildTestPDF(t, 1)
	r, err := NewReader(data)
	if err != nil {
		t.Fatalf("NewReader: %v", err)
	}
	// Find the content stream object and try to ResolveDict on it.
	// Content stream is referenced from the page. Let's get page dict and resolve its Contents.
	pageDict, err := r.PageDict(0)
	if err != nil {
		t.Fatalf("PageDict: %v", err)
	}
	contents, ok := pageDict[Name("Contents")]
	if !ok {
		t.Fatal("page missing /Contents")
	}
	// Contents could be a ref to a stream; ResolveDict should return the stream's Dict.
	d, err := r.ResolveDict(contents)
	if err != nil {
		t.Fatalf("ResolveDict(contents): %v", err)
	}
	if d == nil {
		t.Fatal("ResolveDict returned nil")
	}
}

func TestReaderResolveDictError(t *testing.T) {
	data := buildTestPDF(t, 1)
	r, err := NewReader(data)
	if err != nil {
		t.Fatalf("NewReader: %v", err)
	}
	// ResolveDict on a non-dict/non-stream should error.
	_, err = r.ResolveDict(Integer(42))
	if err == nil {
		t.Error("expected error for ResolveDict on Integer")
	}
}

func TestReaderGetObjectNotFound(t *testing.T) {
	data := buildTestPDF(t, 1)
	r, err := NewReader(data)
	if err != nil {
		t.Fatalf("NewReader: %v", err)
	}
	_, err = r.GetObject(99999)
	if err == nil {
		t.Error("expected error for non-existent object")
	}
}

func TestReaderStringWithPages(t *testing.T) {
	data := buildTestPDF(t, 3)
	r, err := NewReader(data)
	if err != nil {
		t.Fatalf("NewReader: %v", err)
	}
	// Load pages first.
	_, _ = r.PageCount()
	s := r.String()
	if !bytes.Contains([]byte(s), []byte("3 pages")) {
		t.Errorf("String() = %q, want to contain '3 pages'", s)
	}
}

func TestReaderInvalidPDF(t *testing.T) {
	_, err := NewReader([]byte("not a pdf"))
	if err == nil {
		t.Error("expected error for invalid PDF data")
	}
}

func TestReaderDecodeStreamContent(t *testing.T) {
	// Build a PDF with compression enabled to test decodeStreamContent.
	var buf bytes.Buffer
	w := NewWriter(&buf)
	w.SetCompression(true)

	contentRef := w.AllocObject()
	content := Stream{
		Dict:    Dict{},
		Content: []byte("BT /F1 12 Tf 100 700 Td (Hello) Tj ET"),
	}
	if err := w.WriteObject(contentRef, content); err != nil {
		t.Fatal(err)
	}
	if err := w.AddPage(PageObject{
		MediaBox: Rectangle{LLX: 0, LLY: 0, URX: 595, URY: 842},
		Contents: []ObjectRef{contentRef},
	}); err != nil {
		t.Fatal(err)
	}
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}

	r, err := NewReader(buf.Bytes())
	if err != nil {
		t.Fatalf("NewReader: %v", err)
	}

	// Get the content stream and decode it.
	obj, err := r.GetObject(contentRef.Number)
	if err != nil {
		t.Fatalf("GetObject: %v", err)
	}
	s, ok := obj.(Stream)
	if !ok {
		t.Fatalf("got %T, want Stream", obj)
	}
	decoded, err := r.decodeStreamContent(s)
	if err != nil {
		t.Fatalf("decodeStreamContent: %v", err)
	}
	if !bytes.Contains(decoded, []byte("Hello")) {
		t.Errorf("decoded content = %q, expected to contain 'Hello'", decoded)
	}
}

func TestReaderDecodeStreamNoFilter(t *testing.T) {
	data := buildTestPDF(t, 1)
	r, err := NewReader(data)
	if err != nil {
		t.Fatalf("NewReader: %v", err)
	}
	// Stream without filter should return content as-is.
	s := Stream{
		Dict:    Dict{},
		Content: []byte("raw content"),
	}
	result, err := r.decodeStreamContent(s)
	if err != nil {
		t.Fatalf("decodeStreamContent: %v", err)
	}
	if string(result) != "raw content" {
		t.Errorf("got %q, want %q", result, "raw content")
	}
}

func TestReaderDecodeStreamUnsupportedFilter(t *testing.T) {
	data := buildTestPDF(t, 1)
	r, err := NewReader(data)
	if err != nil {
		t.Fatalf("NewReader: %v", err)
	}
	s := Stream{
		Dict: Dict{
			Name("Filter"): Name("LZWDecode"),
		},
		Content: []byte("data"),
	}
	_, err = r.decodeStreamContent(s)
	if err == nil {
		t.Error("expected error for unsupported filter")
	}
}

func TestReaderDecodeStreamFilterArray(t *testing.T) {
	data := buildTestPDF(t, 1)
	r, err := NewReader(data)
	if err != nil {
		t.Fatalf("NewReader: %v", err)
	}

	// Compress some data.
	original := []byte("test data for array filter")
	var zbuf bytes.Buffer
	zw := zlib.NewWriter(&zbuf)
	_, _ = zw.Write(original)
	_ = zw.Close()

	s := Stream{
		Dict: Dict{
			Name("Filter"): Array{Name("FlateDecode")},
		},
		Content: zbuf.Bytes(),
	}
	result, err := r.decodeStreamContent(s)
	if err != nil {
		t.Fatalf("decodeStreamContent: %v", err)
	}
	if !bytes.Equal(result, original) {
		t.Errorf("got %q, want %q", result, original)
	}
}

// TestParseXRefFieldWidths tests the parseXRefFieldWidths function directly.
func TestParseXRefFieldWidths(t *testing.T) {
	tests := []struct {
		name    string
		dict    Dict
		want    [3]int
		wantErr bool
	}{
		{
			name: "valid W array",
			dict: Dict{
				Name("W"): Array{Integer(1), Integer(2), Integer(1)},
			},
			want: [3]int{1, 2, 1},
		},
		{
			name:    "missing W",
			dict:    Dict{},
			wantErr: true,
		},
		{
			name: "W not array",
			dict: Dict{
				Name("W"): Integer(42),
			},
			wantErr: true,
		},
		{
			name: "W wrong length",
			dict: Dict{
				Name("W"): Array{Integer(1), Integer(2)},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseXRefFieldWidths(tt.dict)
			if tt.wantErr {
				if err == nil {
					t.Error("expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}

// TestParseXRefIndices tests the parseXRefIndices function directly.
func TestParseXRefIndices(t *testing.T) {
	tests := []struct {
		name string
		dict Dict
		want []int
	}{
		{
			name: "default (no Index)",
			dict: Dict{
				Name("Size"): Integer(10),
			},
			want: []int{0, 10},
		},
		{
			name: "with Index array",
			dict: Dict{
				Name("Size"):  Integer(10),
				Name("Index"): Array{Integer(5), Integer(3), Integer(20), Integer(2)},
			},
			want: []int{5, 3, 20, 2},
		},
		{
			name: "no Size",
			dict: Dict{},
			want: []int{0, 0},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseXRefIndices(tt.dict)
			if len(got) != len(tt.want) {
				t.Fatalf("len = %d, want %d", len(got), len(tt.want))
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("index[%d] = %d, want %d", i, got[i], tt.want[i])
				}
			}
		})
	}
}

// TestReadXRefFields tests the readXRefFields function directly.
func TestReadXRefFields(t *testing.T) {
	tests := []struct {
		name string
		data []byte
		w    [3]int
		want [3]int64
	}{
		{
			name: "1-2-1 widths",
			data: []byte{1, 0, 100, 0},
			w:    [3]int{1, 2, 1},
			want: [3]int64{1, 100, 0},
		},
		{
			name: "type 0 width defaults to 1",
			data: []byte{0, 200, 5},
			w:    [3]int{0, 2, 1},
			want: [3]int64{1, 200, 5},
		},
		{
			name: "1-3-0 widths",
			data: []byte{1, 0, 1, 0},
			w:    [3]int{1, 3, 0},
			want: [3]int64{1, 256, 0},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := readXRefFields(tt.data, tt.w)
			if got != tt.want {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}

// TestParseXRefEntries tests parseXRefEntries populating the xref table.
func TestParseXRefEntries(t *testing.T) {
	r := &Reader{
		xref: make(map[int]int64),
	}

	// Build content: 3 entries with w=[1,2,1], entrySize=4
	// Entry 0: type=1, offset=256 (0x0100), gen=0 → in use
	// Entry 1: type=0, offset=0, gen=0 → free (not added)
	// Entry 2: type=1, offset=512 (0x0200), gen=0 → in use
	content := []byte{
		1, 0x01, 0x00, 0, // type=1, offset=256, gen=0
		0, 0x00, 0x00, 0, // type=0, free
		1, 0x02, 0x00, 0, // type=1, offset=512, gen=0
	}
	w := [3]int{1, 2, 1}
	indices := []int{5, 3} // start at obj 5, count 3

	r.parseXRefEntries(content, indices, w, 4)

	// Should have objects 5 and 7 (type=1), not 6 (type=0).
	if r.xref[5] != 256 {
		t.Errorf("xref[5] = %d, want 256", r.xref[5])
	}
	if _, ok := r.xref[6]; ok {
		t.Error("xref[6] should not exist (free entry)")
	}
	if r.xref[7] != 512 {
		t.Errorf("xref[7] = %d, want 512", r.xref[7])
	}
}

// TestBuildXRefStreamPDF tests parsing a PDF with xref stream (PDF 1.5+).
func TestBuildXRefStreamPDF(t *testing.T) {
	// Build a minimal PDF 1.5 with cross-reference stream manually.
	// This is complex but tests the xref stream parsing path.

	// We'll build a simple PDF with xref stream.
	var pdf bytes.Buffer

	// Header
	pdf.WriteString("%PDF-1.5\n")

	// Object 1: Catalog
	obj1Offset := pdf.Len()
	pdf.WriteString("1 0 obj\n<< /Type /Catalog /Pages 2 0 R >>\nendobj\n")

	// Object 2: Pages
	obj2Offset := pdf.Len()
	pdf.WriteString("2 0 obj\n<< /Type /Pages /Kids [3 0 R] /Count 1 >>\nendobj\n")

	// Object 3: Page
	obj3Offset := pdf.Len()
	pdf.WriteString("3 0 obj\n<< /Type /Page /Parent 2 0 R /MediaBox [0 0 612 792] >>\nendobj\n")

	// Object 4: XRef stream
	xrefStreamOffset := pdf.Len()

	// Build xref stream content: w=[1, 2, 1]
	// Entry for obj 0: free (type=0)
	// Entry for obj 1: in-use at obj1Offset
	// Entry for obj 2: in-use at obj2Offset
	// Entry for obj 3: in-use at obj3Offset
	// Entry for obj 4: in-use at xrefStreamOffset (self)
	var streamContent bytes.Buffer
	writeXRefEntry := func(typ byte, offset int, gen byte) {
		streamContent.WriteByte(typ)
		streamContent.WriteByte(byte(offset >> 8))
		streamContent.WriteByte(byte(offset & 0xFF))
		streamContent.WriteByte(gen)
	}
	writeXRefEntry(0, 0, 0)                // obj 0: free
	writeXRefEntry(1, obj1Offset, 0)       // obj 1
	writeXRefEntry(1, obj2Offset, 0)       // obj 2
	writeXRefEntry(1, obj3Offset, 0)       // obj 3
	writeXRefEntry(1, xrefStreamOffset, 0) // obj 4 (xref stream itself)

	// Compress the stream content.
	var compressedContent bytes.Buffer
	zw := zlib.NewWriter(&compressedContent)
	_, _ = zw.Write(streamContent.Bytes())
	_ = zw.Close()

	fmt.Fprintf(&pdf, "4 0 obj\n")
	fmt.Fprintf(&pdf, "<< /Type /XRef /Size 5 /W [1 2 1] /Root 1 0 R /Length %d /Filter /FlateDecode >>\n", compressedContent.Len())
	fmt.Fprintf(&pdf, "stream\n")
	pdf.Write(compressedContent.Bytes())
	fmt.Fprintf(&pdf, "\nendstream\nendobj\n")

	fmt.Fprintf(&pdf, "startxref\n%d\n%%%%EOF\n", xrefStreamOffset)

	r, err := NewReader(pdf.Bytes())
	if err != nil {
		t.Fatalf("NewReader with xref stream: %v", err)
	}

	count, err := r.PageCount()
	if err != nil {
		t.Fatalf("PageCount: %v", err)
	}
	if count != 1 {
		t.Errorf("PageCount = %d, want 1", count)
	}

	// Verify trailer has /Root.
	trailer := r.Trailer()
	if _, ok := trailer[Name("Root")]; !ok {
		t.Error("trailer missing /Root")
	}
}
