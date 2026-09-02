package pdf

import (
	"bytes"
	"compress/zlib"
	"fmt"
	"testing"
)

// encodePNGUp applies the PNG "Up" filter to rows of rowLen bytes, producing the
// representation an xref stream with /Predictor 12 carries on disk.
func encodePNGUp(raw []byte, rowLen int) []byte {
	var out bytes.Buffer
	prev := make([]byte, rowLen)
	for start := 0; start < len(raw); start += rowLen {
		end := start + rowLen
		if end > len(raw) {
			end = len(raw)
		}
		row := raw[start:end]
		out.WriteByte(2) // filter type: Up
		for i, b := range row {
			out.WriteByte(b - prev[i])
		}
		prev = make([]byte, rowLen)
		copy(prev, row)
	}
	return out.Bytes()
}

// buildPredictedXRefPDF builds a PDF whose xref stream is Flate-compressed with
// a PNG "Up" predictor — the encoding emitted by nearly every modern producer.
func buildPredictedXRefPDF(t *testing.T) []byte {
	t.Helper()

	var pdf bytes.Buffer
	pdf.WriteString("%PDF-1.5\n")

	obj1 := pdf.Len()
	pdf.WriteString("1 0 obj\n<< /Type /Catalog /Pages 2 0 R >>\nendobj\n")
	obj2 := pdf.Len()
	pdf.WriteString("2 0 obj\n<< /Type /Pages /Kids [3 0 R] /Count 1 >>\nendobj\n")
	obj3 := pdf.Len()
	pdf.WriteString("3 0 obj\n<< /Type /Page /Parent 2 0 R /MediaBox [0 0 595 842] >>\nendobj\n")
	xrefOffset := pdf.Len()

	// /W [1 2 1] → 4-byte rows.
	var raw bytes.Buffer
	writeEntry := func(typ byte, f2 int, f3 byte) {
		raw.WriteByte(typ)
		raw.WriteByte(byte(f2 >> 8))
		raw.WriteByte(byte(f2 & 0xFF))
		raw.WriteByte(f3)
	}
	writeEntry(0, 0, 0)
	writeEntry(1, obj1, 0)
	writeEntry(1, obj2, 0)
	writeEntry(1, obj3, 0)
	writeEntry(1, xrefOffset, 0)

	var compressed bytes.Buffer
	zw := zlib.NewWriter(&compressed)
	_, _ = zw.Write(encodePNGUp(raw.Bytes(), 4))
	_ = zw.Close()

	fmt.Fprintf(&pdf, "4 0 obj\n<< /Type /XRef /Size 5 /W [1 2 1] /Root 1 0 R /Filter /FlateDecode "+
		"/DecodeParms << /Predictor 12 /Columns 4 >> /Length %d >>\nstream\n", compressed.Len())
	pdf.Write(compressed.Bytes())
	pdf.WriteString("\nendstream\nendobj\n")

	fmt.Fprintf(&pdf, "startxref\n%d\n%%%%EOF\n", xrefOffset)
	return pdf.Bytes()
}

func TestReaderXRefStreamWithPredictor(t *testing.T) {
	r, err := NewReader(buildPredictedXRefPDF(t))
	if err != nil {
		t.Fatalf("NewReader: %v", err)
	}

	count, err := r.PageCount()
	if err != nil {
		t.Fatalf("PageCount: %v", err)
	}
	if count != 1 {
		t.Fatalf("PageCount = %d, want 1", count)
	}

	info, err := r.Page(0)
	if err != nil {
		t.Fatalf("Page: %v", err)
	}
	if info.MediaBox.URX != 595 || info.MediaBox.URY != 842 {
		t.Errorf("MediaBox = %+v, want 595x842", info.MediaBox)
	}
}

func TestApplyPNGPredictor(t *testing.T) {
	raw := []byte{
		1, 2, 3, 4,
		5, 7, 9, 11,
		0, 0, 0, 0,
	}
	encoded := encodePNGUp(raw, 4)

	got, err := applyPredictor(encoded, predictorParams{predictor: 12, colors: 1, bitsPerComponent: 8, columns: 4})
	if err != nil {
		t.Fatalf("applyPredictor: %v", err)
	}
	if !bytes.Equal(got, raw) {
		t.Errorf("got %v, want %v", got, raw)
	}
}

func TestApplyPNGPredictorFilterTypes(t *testing.T) {
	p := predictorParams{predictor: 15, colors: 1, bitsPerComponent: 8, columns: 3}

	tests := []struct {
		name    string
		encoded []byte
		want    []byte
	}{
		// filter 0 (None): bytes pass through.
		{"none", []byte{0, 10, 20, 30}, []byte{10, 20, 30}},
		// filter 1 (Sub): each byte adds the one bpp to its left.
		{"sub", []byte{1, 10, 5, 5}, []byte{10, 15, 20}},
		// filter 3 (Average) on the first row: prev row is zero, so each byte
		// adds half of the byte to its left.
		{"average", []byte{3, 10, 5, 5}, []byte{10, 10, 10}},
		// filter 4 (Paeth) on the first row degenerates to Sub.
		{"paeth", []byte{4, 10, 5, 5}, []byte{10, 15, 20}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := applyPredictor(tc.encoded, p)
			if err != nil {
				t.Fatalf("applyPredictor: %v", err)
			}
			if !bytes.Equal(got, tc.want) {
				t.Errorf("got %v, want %v", got, tc.want)
			}
		})
	}
}

func TestApplyPredictorPassthroughAndErrors(t *testing.T) {
	data := []byte{1, 2, 3, 4}

	got, err := applyPredictor(data, predictorParams{predictor: 1, colors: 1, bitsPerComponent: 8, columns: 4})
	if err != nil {
		t.Fatalf("predictor 1: %v", err)
	}
	if !bytes.Equal(got, data) {
		t.Errorf("predictor 1 modified data: %v", got)
	}

	if _, err := applyPredictor(data, predictorParams{predictor: 5, colors: 1, bitsPerComponent: 8, columns: 4}); err == nil {
		t.Error("predictor 5: expected error")
	}

	// Unknown PNG row filter type.
	if _, err := applyPredictor([]byte{9, 1, 2, 3}, predictorParams{predictor: 12, colors: 1, bitsPerComponent: 8, columns: 3}); err == nil {
		t.Error("filter type 9: expected error")
	}
}

func TestApplyTIFFPredictor(t *testing.T) {
	// Two rows of 3 single-byte components, stored as deltas.
	encoded := []byte{
		10, 5, 5,
		1, 1, 1,
	}
	p := predictorParams{predictor: 2, colors: 1, bitsPerComponent: 8, columns: 3}

	got, err := applyPredictor(encoded, p)
	if err != nil {
		t.Fatalf("applyPredictor: %v", err)
	}
	want := []byte{10, 15, 20, 1, 2, 3}
	if !bytes.Equal(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}

	p.bitsPerComponent = 4
	if _, err := applyPredictor(encoded, p); err == nil {
		t.Error("4 bits per component: expected error")
	}
}

func TestReadPredictorParams(t *testing.T) {
	r := &Reader{}

	p, err := r.readPredictorParams(nil)
	if err != nil {
		t.Fatalf("nil dict: %v", err)
	}
	if p.predictor != 1 || p.colors != 1 || p.bitsPerComponent != 8 || p.columns != 1 {
		t.Errorf("defaults = %+v", p)
	}

	p, err = r.readPredictorParams(Dict{
		Name("Predictor"):        Integer(12),
		Name("Colors"):           Integer(3),
		Name("BitsPerComponent"): Integer(8),
		Name("Columns"):          Integer(4),
	})
	if err != nil {
		t.Fatalf("full dict: %v", err)
	}
	if p.predictor != 12 || p.colors != 3 || p.columns != 4 {
		t.Errorf("params = %+v", p)
	}
	if got := p.pixelBytes(); got != 3 {
		t.Errorf("pixelBytes = %d, want 3", got)
	}
	if got := p.rowLength(); got != 12 {
		t.Errorf("rowLength = %d, want 12", got)
	}

	if _, err := r.readPredictorParams(Dict{Name("Columns"): Integer(0)}); err == nil {
		t.Error("Columns=0: expected error")
	}
	if _, err := r.readPredictorParams(Dict{Name("Predictor"): Name("bogus")}); err == nil {
		t.Error("non-integer Predictor: expected error")
	}
}

func TestDecodeParmsList(t *testing.T) {
	r := &Reader{}

	// Single dict form.
	parms, err := r.decodeParmsList(Dict{
		Name("DecodeParms"): Dict{Name("Predictor"): Integer(12)},
	}, 1)
	if err != nil {
		t.Fatalf("single dict: %v", err)
	}
	if parms[0] == nil {
		t.Fatal("parms[0] is nil")
	}

	// Array form with a null for the filter that takes no parameters.
	parms, err = r.decodeParmsList(Dict{
		Name("DecodeParms"): Array{Null{}, Dict{Name("Predictor"): Integer(12)}},
	}, 2)
	if err != nil {
		t.Fatalf("array: %v", err)
	}
	if parms[0] != nil {
		t.Error("parms[0] should be nil for a null entry")
	}
	if parms[1] == nil {
		t.Error("parms[1] should be set")
	}

	// Absent.
	parms, err = r.decodeParmsList(Dict{}, 1)
	if err != nil {
		t.Fatalf("absent: %v", err)
	}
	if parms[0] != nil {
		t.Error("parms[0] should be nil when /DecodeParms is absent")
	}
}
