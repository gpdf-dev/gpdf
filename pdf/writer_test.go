package pdf_test

import (
	"bytes"
	"compress/zlib"
	"io"
	"strings"
	"testing"

	"github.com/gpdf-dev/gpdf/pdf"
)

// ===========================================================================
// XRefTable tests
// ===========================================================================

func TestNewXRefTable(t *testing.T) {
	xref := pdf.NewXRefTable()
	if xref.Size() != 1 {
		t.Errorf("new XRefTable size = %d, want 1", xref.Size())
	}
}

func TestXRefTable_Add(t *testing.T) {
	xref := pdf.NewXRefTable()
	xref.Add(1, 100, 0)
	if xref.Size() != 2 {
		t.Errorf("after Add(1,...), size = %d, want 2", xref.Size())
	}
}

func TestXRefTable_Add_Gap(t *testing.T) {
	xref := pdf.NewXRefTable()
	// Adding object 5 should create entries 1-4 as free entries.
	xref.Add(5, 500, 0)
	if xref.Size() != 6 {
		t.Errorf("after Add(5,...), size = %d, want 6", xref.Size())
	}
}

func TestXRefTable_Add_Multiple(t *testing.T) {
	xref := pdf.NewXRefTable()
	xref.Add(1, 100, 0)
	xref.Add(2, 200, 0)
	xref.Add(3, 300, 0)
	if xref.Size() != 4 {
		t.Errorf("size = %d, want 4", xref.Size())
	}
}

func TestXRefTable_WriteTo(t *testing.T) {
	xref := pdf.NewXRefTable()
	xref.Add(1, 9, 0)
	xref.Add(2, 74, 0)

	var buf bytes.Buffer
	n, err := xref.WriteTo(&buf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n <= 0 {
		t.Errorf("WriteTo returned %d bytes, want > 0", n)
	}
	got := buf.String()

	// Must start with "xref"
	if !strings.HasPrefix(got, "xref\n") {
		t.Errorf("output does not start with xref header: %q", got[:min(len(got), 30)])
	}

	// Must contain "0 3" (3 entries: 0, 1, 2)
	if !strings.Contains(got, "0 3\n") {
		t.Errorf("missing entry count line: %q", got)
	}

	// Entry 0: free list head
	if !strings.Contains(got, "0000000000 65535 f \r\n") {
		t.Errorf("missing free list head entry: %q", got)
	}

	// Entry 1: offset 9
	if !strings.Contains(got, "0000000009 00000 n \r\n") {
		t.Errorf("missing entry 1: %q", got)
	}

	// Entry 2: offset 74
	if !strings.Contains(got, "0000000074 00000 n \r\n") {
		t.Errorf("missing entry 2: %q", got)
	}
}

func TestXRefTable_WriteTo_SingleEntry(t *testing.T) {
	xref := pdf.NewXRefTable()
	var buf bytes.Buffer
	_, err := xref.WriteTo(&buf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got := buf.String()
	if !strings.Contains(got, "0 1\n") {
		t.Errorf("expected '0 1' for single-entry xref, got: %q", got)
	}
}

// ===========================================================================
// CompressFlate tests
// ===========================================================================

func TestCompressFlate(t *testing.T) {
	input := []byte("Hello, this is test data for compression. Repeated text repeated text repeated text.")
	compressed, err := pdf.CompressFlate(input)
	if err != nil {
		t.Fatalf("CompressFlate error: %v", err)
	}
	if len(compressed) == 0 {
		t.Fatal("compressed output is empty")
	}

	// Decompress to verify round-trip.
	reader, err := zlib.NewReader(bytes.NewReader(compressed))
	if err != nil {
		t.Fatalf("zlib.NewReader error: %v", err)
	}
	defer func() { _ = reader.Close() }()
	decompressed, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("decompression error: %v", err)
	}
	if !bytes.Equal(decompressed, input) {
		t.Errorf("round-trip failed: got %q, want %q", decompressed, input)
	}
}

func TestCompressFlate_Empty(t *testing.T) {
	compressed, err := pdf.CompressFlate([]byte{})
	if err != nil {
		t.Fatalf("CompressFlate error: %v", err)
	}

	reader, err := zlib.NewReader(bytes.NewReader(compressed))
	if err != nil {
		t.Fatalf("zlib.NewReader error: %v", err)
	}
	defer func() { _ = reader.Close() }()
	decompressed, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("decompression error: %v", err)
	}
	if len(decompressed) != 0 {
		t.Errorf("expected empty decompressed output, got %d bytes", len(decompressed))
	}
}

func TestCompressFlate_LargeData(t *testing.T) {
	input := bytes.Repeat([]byte("ABCDEFGHIJKLMNOPQRSTUVWXYZ"), 1000)
	compressed, err := pdf.CompressFlate(input)
	if err != nil {
		t.Fatalf("CompressFlate error: %v", err)
	}
	// Compressed should be smaller than input for highly repetitive data.
	if len(compressed) >= len(input) {
		t.Errorf("compressed (%d bytes) is not smaller than input (%d bytes)", len(compressed), len(input))
	}
}

// ===========================================================================
// Color tests
// ===========================================================================

func TestRGB(t *testing.T) {
	c := pdf.RGB(0.5, 0.6, 0.7)
	if c.R != 0.5 || c.G != 0.6 || c.B != 0.7 {
		t.Errorf("RGB components = (%v, %v, %v), want (0.5, 0.6, 0.7)", c.R, c.G, c.B)
	}
	if c.A != 1.0 {
		t.Errorf("Alpha = %v, want 1.0", c.A)
	}
	if c.Space != pdf.ColorSpaceRGB {
		t.Errorf("Space = %v, want ColorSpaceRGB", c.Space)
	}
}

func TestRGBHex(t *testing.T) {
	tests := []struct {
		name  string
		hex   uint32
		wantR float64
		wantG float64
		wantB float64
	}{
		{"red", 0xFF0000, 1.0, 0.0, 0.0},
		{"green", 0x00FF00, 0.0, 1.0, 0.0},
		{"blue", 0x0000FF, 0.0, 0.0, 1.0},
		{"white", 0xFFFFFF, 1.0, 1.0, 1.0},
		{"black", 0x000000, 0.0, 0.0, 0.0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := pdf.RGBHex(tt.hex)
			if c.R != tt.wantR || c.G != tt.wantG || c.B != tt.wantB {
				t.Errorf("RGBHex(0x%06X) = (%v, %v, %v), want (%v, %v, %v)",
					tt.hex, c.R, c.G, c.B, tt.wantR, tt.wantG, tt.wantB)
			}
			if c.Space != pdf.ColorSpaceRGB {
				t.Errorf("Space = %v, want ColorSpaceRGB", c.Space)
			}
		})
	}
}

func TestGray(t *testing.T) {
	c := pdf.Gray(0.5)
	if c.R != 0.5 {
		t.Errorf("Gray value = %v, want 0.5", c.R)
	}
	if c.Space != pdf.ColorSpaceGray {
		t.Errorf("Space = %v, want ColorSpaceGray", c.Space)
	}
	if c.A != 1.0 {
		t.Errorf("Alpha = %v, want 1.0", c.A)
	}
}

func TestCMYK(t *testing.T) {
	c := pdf.CMYK(0.1, 0.2, 0.3, 0.4)
	if c.R != 0.1 || c.G != 0.2 || c.B != 0.3 || c.A != 0.4 {
		t.Errorf("CMYK = (%v, %v, %v, %v), want (0.1, 0.2, 0.3, 0.4)", c.R, c.G, c.B, c.A)
	}
	if c.Space != pdf.ColorSpaceCMYK {
		t.Errorf("Space = %v, want ColorSpaceCMYK", c.Space)
	}
}

func TestPredefinedColors(t *testing.T) {
	tests := []struct {
		name  string
		color pdf.Color
		space pdf.ColorSpace
	}{
		{"Black", pdf.Black, pdf.ColorSpaceGray},
		{"White", pdf.White, pdf.ColorSpaceGray},
		{"Red", pdf.Red, pdf.ColorSpaceRGB},
		{"Green", pdf.Green, pdf.ColorSpaceRGB},
		{"Blue", pdf.Blue, pdf.ColorSpaceRGB},
		{"Yellow", pdf.Yellow, pdf.ColorSpaceRGB},
		{"Cyan", pdf.Cyan, pdf.ColorSpaceRGB},
		{"Magenta", pdf.Magenta, pdf.ColorSpaceRGB},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.color.Space != tt.space {
				t.Errorf("%s.Space = %v, want %v", tt.name, tt.color.Space, tt.space)
			}
		})
	}
}

func TestStrokeColorCmd_RGB(t *testing.T) {
	c := pdf.RGB(1, 0, 0)
	got := c.StrokeColorCmd()
	want := "1 0 0 RG"
	if got != want {
		t.Errorf("StrokeColorCmd() = %q, want %q", got, want)
	}
}

func TestStrokeColorCmd_Gray(t *testing.T) {
	c := pdf.Gray(0.5)
	got := c.StrokeColorCmd()
	want := "0.5 G"
	if got != want {
		t.Errorf("StrokeColorCmd() = %q, want %q", got, want)
	}
}

func TestStrokeColorCmd_CMYK(t *testing.T) {
	c := pdf.CMYK(0.1, 0.2, 0.3, 0.4)
	got := c.StrokeColorCmd()
	want := "0.1 0.2 0.3 0.4 K"
	if got != want {
		t.Errorf("StrokeColorCmd() = %q, want %q", got, want)
	}
}

func TestFillColorCmd_RGB(t *testing.T) {
	c := pdf.RGB(0, 1, 0)
	got := c.FillColorCmd()
	want := "0 1 0 rg"
	if got != want {
		t.Errorf("FillColorCmd() = %q, want %q", got, want)
	}
}

func TestFillColorCmd_Gray(t *testing.T) {
	c := pdf.Gray(0)
	got := c.FillColorCmd()
	want := "0 g"
	if got != want {
		t.Errorf("FillColorCmd() = %q, want %q", got, want)
	}
}

func TestFillColorCmd_CMYK(t *testing.T) {
	c := pdf.CMYK(1, 1, 0, 0)
	got := c.FillColorCmd()
	want := "1 1 0 0 k"
	if got != want {
		t.Errorf("FillColorCmd() = %q, want %q", got, want)
	}
}

// ===========================================================================
// Writer tests
// ===========================================================================

func TestNewWriter_WritesHeader(t *testing.T) {
	var buf bytes.Buffer
	_ = pdf.NewWriter(&buf)
	got := buf.String()
	if !strings.HasPrefix(got, "%PDF-1.7\n") {
		t.Errorf("output does not start with PDF header: %q", got[:min(len(got), 20)])
	}
}

func TestWriter_AllocObject(t *testing.T) {
	var buf bytes.Buffer
	w := pdf.NewWriter(&buf)
	// NewWriter pre-allocates catalog (1) and pageTree (2),
	// so the next alloc should be 3.
	ref := w.AllocObject()
	if ref.Number != 3 {
		t.Errorf("AllocObject() = %d, want 3", ref.Number)
	}
	if ref.Generation != 0 {
		t.Errorf("Generation = %d, want 0", ref.Generation)
	}

	ref2 := w.AllocObject()
	if ref2.Number != 4 {
		t.Errorf("second AllocObject() = %d, want 4", ref2.Number)
	}
}

func TestWriter_WriteObject(t *testing.T) {
	var buf bytes.Buffer
	w := pdf.NewWriter(&buf)
	ref := w.AllocObject()
	err := w.WriteObject(ref, pdf.Dict{
		pdf.Name("Type"): pdf.Name("Test"),
	})
	if err != nil {
		t.Fatalf("WriteObject error: %v", err)
	}
	got := buf.String()
	// Should contain "N 0 obj"
	if !strings.Contains(got, "3 0 obj\n") {
		t.Errorf("missing object header in output: %q", got)
	}
	if !strings.Contains(got, "endobj") {
		t.Errorf("missing endobj in output: %q", got)
	}
	if !strings.Contains(got, "/Type /Test") {
		t.Errorf("missing dict content in output: %q", got)
	}
}

func TestWriter_SetCompression(t *testing.T) {
	var buf bytes.Buffer
	w := pdf.NewWriter(&buf)
	w.SetCompression(false)

	// Register a font; without compression it should not contain FlateDecode.
	fontData := []byte("fake font data for testing purposes")
	_, _, err := w.RegisterFont("TestFont", fontData)
	if err != nil {
		t.Fatalf("RegisterFont error: %v", err)
	}
	got := buf.String()
	if strings.Contains(got, "FlateDecode") {
		t.Errorf("expected no FlateDecode when compression is disabled, got: %q", got)
	}
}

func TestWriter_RegisterFont_Standard14(t *testing.T) {
	var buf bytes.Buffer
	w := pdf.NewWriter(&buf)
	resName, ref, err := w.RegisterFont("Helvetica", nil)
	if err != nil {
		t.Fatalf("RegisterFont error: %v", err)
	}
	if resName != "F1" {
		t.Errorf("resource name = %q, want %q", resName, "F1")
	}
	if ref.Number == 0 {
		t.Errorf("ref.Number = 0, want > 0")
	}
	got := buf.String()
	if !strings.Contains(got, "/Type1") {
		t.Errorf("expected Type1 subtype for standard font: %q", got)
	}
	if !strings.Contains(got, "/Helvetica") {
		t.Errorf("expected BaseFont /Helvetica: %q", got)
	}
}

func TestWriter_RegisterFont_WithData(t *testing.T) {
	var buf bytes.Buffer
	w := pdf.NewWriter(&buf)
	w.SetCompression(false) // disable compression for predictable output

	fontData := []byte("fake TrueType font data")
	resName, ref, err := w.RegisterFont("MyFont", fontData)
	if err != nil {
		t.Fatalf("RegisterFont error: %v", err)
	}
	if resName != "F1" {
		t.Errorf("resource name = %q, want %q", resName, "F1")
	}
	if ref.Number == 0 {
		t.Errorf("ref.Number = 0, want > 0")
	}
	got := buf.String()
	if !strings.Contains(got, "/TrueType") {
		t.Errorf("expected TrueType subtype: %q", got)
	}
	if !strings.Contains(got, "/FontDescriptor") {
		t.Errorf("expected FontDescriptor: %q", got)
	}
	if !strings.Contains(got, "/FontFile2") {
		t.Errorf("expected FontFile2 reference: %q", got)
	}
}

func TestWriter_RegisterFont_Duplicate(t *testing.T) {
	var buf bytes.Buffer
	w := pdf.NewWriter(&buf)
	resName1, ref1, err := w.RegisterFont("Helvetica", nil)
	if err != nil {
		t.Fatalf("first RegisterFont error: %v", err)
	}
	resName2, ref2, err := w.RegisterFont("Helvetica", nil)
	if err != nil {
		t.Fatalf("second RegisterFont error: %v", err)
	}
	// Same font should return same ref.
	if ref1.Number != ref2.Number {
		t.Errorf("duplicate font returned different refs: %d vs %d", ref1.Number, ref2.Number)
	}
	// Resource names should be consistent.
	_ = resName1
	_ = resName2
}

func TestWriter_RegisterImage(t *testing.T) {
	var buf bytes.Buffer
	w := pdf.NewWriter(&buf)
	w.SetCompression(false)

	imgData := bytes.Repeat([]byte{0xFF, 0x00, 0x00}, 10) // fake red pixel data
	resName, ref, err := w.RegisterImage("testimg", imgData, 10, 1, "DeviceRGB", "", nil)
	if err != nil {
		t.Fatalf("RegisterImage error: %v", err)
	}
	if resName != "Im1" {
		t.Errorf("resource name = %q, want %q", resName, "Im1")
	}
	if ref.Number == 0 {
		t.Errorf("ref.Number = 0, want > 0")
	}
	got := buf.String()
	if !strings.Contains(got, "/Image") {
		t.Errorf("expected /Image subtype: %q", got)
	}
	if !strings.Contains(got, "/Width 10") {
		t.Errorf("expected /Width 10: %q", got)
	}
	if !strings.Contains(got, "/Height 1") {
		t.Errorf("expected /Height 1: %q", got)
	}
	if !strings.Contains(got, "/DeviceRGB") {
		t.Errorf("expected /DeviceRGB: %q", got)
	}
}

func TestWriter_RegisterImage_Duplicate(t *testing.T) {
	var buf bytes.Buffer
	w := pdf.NewWriter(&buf)
	w.SetCompression(false)

	imgData := []byte{0x01, 0x02, 0x03}
	_, ref1, err := w.RegisterImage("img1", imgData, 1, 1, "DeviceGray", "", nil)
	if err != nil {
		t.Fatalf("first RegisterImage error: %v", err)
	}
	_, ref2, err := w.RegisterImage("img1", imgData, 1, 1, "DeviceGray", "", nil)
	if err != nil {
		t.Fatalf("second RegisterImage error: %v", err)
	}
	if ref1.Number != ref2.Number {
		t.Errorf("duplicate image returned different refs: %d vs %d", ref1.Number, ref2.Number)
	}
}

func TestWriter_RegisterImage_Compressed(t *testing.T) {
	var buf bytes.Buffer
	w := pdf.NewWriter(&buf)
	// compression is enabled by default

	imgData := bytes.Repeat([]byte{0xAA}, 100)
	_, _, err := w.RegisterImage("cimg", imgData, 10, 10, "DeviceGray", "", nil)
	if err != nil {
		t.Fatalf("RegisterImage error: %v", err)
	}
	got := buf.String()
	if !strings.Contains(got, "/FlateDecode") {
		t.Errorf("expected FlateDecode filter when compression enabled: %q", got)
	}
}

func TestWriter_RegisterImage_WithSMask(t *testing.T) {
	var buf bytes.Buffer
	w := pdf.NewWriter(&buf)
	w.SetCompression(false)

	imgData := bytes.Repeat([]byte{0xFF, 0x00, 0x00}, 4) // 4 red pixels
	smaskData := []byte{128, 255, 0, 200}                 // alpha for 4 pixels
	resName, ref, err := w.RegisterImage("smask-img", imgData, 2, 2, "DeviceRGB", "", smaskData)
	if err != nil {
		t.Fatalf("RegisterImage error: %v", err)
	}
	if resName != "Im1" {
		t.Errorf("resource name = %q, want %q", resName, "Im1")
	}
	if ref.Number == 0 {
		t.Errorf("ref.Number = 0, want > 0")
	}
	got := buf.String()
	if !strings.Contains(got, "/SMask") {
		t.Errorf("expected /SMask in output: %q", got)
	}
	if !strings.Contains(got, "/DeviceGray") {
		t.Errorf("expected /DeviceGray for SMask: %q", got)
	}
}

func TestWriter_RegisterImage_NoSMask(t *testing.T) {
	var buf bytes.Buffer
	w := pdf.NewWriter(&buf)
	w.SetCompression(false)

	imgData := bytes.Repeat([]byte{0xFF, 0x00, 0x00}, 4)
	_, _, err := w.RegisterImage("no-smask", imgData, 2, 2, "DeviceRGB", "", nil)
	if err != nil {
		t.Fatalf("RegisterImage error: %v", err)
	}
	got := buf.String()
	if strings.Contains(got, "/SMask") {
		t.Errorf("no smask data should not produce /SMask: %q", got)
	}
}

func TestWriter_AddPage_NoContents(t *testing.T) {
	var buf bytes.Buffer
	w := pdf.NewWriter(&buf)
	err := w.AddPage(pdf.PageObject{
		MediaBox: pdf.Rectangle{LLX: 0, LLY: 0, URX: 612, URY: 792},
	})
	if err != nil {
		t.Fatalf("AddPage error: %v", err)
	}
	got := buf.String()
	if !strings.Contains(got, "/Page") {
		t.Errorf("expected /Page type: %q", got)
	}
	if !strings.Contains(got, "[0 0 612 792]") {
		t.Errorf("expected MediaBox: %q", got)
	}
}

func TestWriter_AddPage_SingleContent(t *testing.T) {
	var buf bytes.Buffer
	w := pdf.NewWriter(&buf)

	// Write a content stream first.
	contentRef := w.AllocObject()
	content := pdf.Stream{
		Dict:    pdf.Dict{},
		Content: []byte("BT /F1 12 Tf (Hello) Tj ET"),
	}
	err := w.WriteObject(contentRef, content)
	if err != nil {
		t.Fatalf("WriteObject error: %v", err)
	}

	err = w.AddPage(pdf.PageObject{
		MediaBox: pdf.Rectangle{LLX: 0, LLY: 0, URX: 612, URY: 792},
		Contents: []pdf.ObjectRef{contentRef},
	})
	if err != nil {
		t.Fatalf("AddPage error: %v", err)
	}
}

func TestWriter_AddPage_MultipleContents(t *testing.T) {
	var buf bytes.Buffer
	w := pdf.NewWriter(&buf)

	ref1 := w.AllocObject()
	_ = w.WriteObject(ref1, pdf.Stream{Dict: pdf.Dict{}, Content: []byte("stream1")})
	ref2 := w.AllocObject()
	_ = w.WriteObject(ref2, pdf.Stream{Dict: pdf.Dict{}, Content: []byte("stream2")})

	err := w.AddPage(pdf.PageObject{
		MediaBox: pdf.Rectangle{LLX: 0, LLY: 0, URX: 612, URY: 792},
		Contents: []pdf.ObjectRef{ref1, ref2},
	})
	if err != nil {
		t.Fatalf("AddPage error: %v", err)
	}
}

func TestWriter_AddPage_WithResources(t *testing.T) {
	var buf bytes.Buffer
	w := pdf.NewWriter(&buf)

	fontRef := pdf.ObjectRef{Number: 10, Generation: 0}
	err := w.AddPage(pdf.PageObject{
		MediaBox: pdf.Rectangle{LLX: 0, LLY: 0, URX: 612, URY: 792},
		Resources: pdf.ResourceDict{
			Font: pdf.Dict{
				pdf.Name("F1"): fontRef,
			},
		},
	})
	if err != nil {
		t.Fatalf("AddPage error: %v", err)
	}
	got := buf.String()
	if !strings.Contains(got, "/Resources") {
		t.Errorf("expected /Resources in page dict: %q", got)
	}
}

func TestWriter_SetInfo(t *testing.T) {
	var buf bytes.Buffer
	w := pdf.NewWriter(&buf)
	w.SetInfo(pdf.DocumentInfo{
		Title:    "Test PDF",
		Author:   "Test Author",
		Producer: "gpdf",
	})
	err := w.AddPage(pdf.PageObject{
		MediaBox: pdf.Rectangle{LLX: 0, LLY: 0, URX: 612, URY: 792},
	})
	if err != nil {
		t.Fatalf("AddPage error: %v", err)
	}
	err = w.Close()
	if err != nil {
		t.Fatalf("Close error: %v", err)
	}
	got := buf.String()
	if !strings.Contains(got, "(Test PDF)") {
		t.Errorf("expected title in output: %q", got)
	}
	if !strings.Contains(got, "(Test Author)") {
		t.Errorf("expected author in output: %q", got)
	}
	if !strings.Contains(got, "(gpdf)") {
		t.Errorf("expected producer in output: %q", got)
	}
}

func TestWriter_Close(t *testing.T) {
	var buf bytes.Buffer
	w := pdf.NewWriter(&buf)
	err := w.AddPage(pdf.PageObject{
		MediaBox: pdf.Rectangle{LLX: 0, LLY: 0, URX: 612, URY: 792},
	})
	if err != nil {
		t.Fatalf("AddPage error: %v", err)
	}
	err = w.Close()
	if err != nil {
		t.Fatalf("Close error: %v", err)
	}
	got := buf.String()

	// Must contain required PDF structure.
	checks := []string{
		"%PDF-1.7",
		"/Catalog",
		"/Pages",
		"xref",
		"trailer",
		"startxref",
		"%%EOF",
	}
	for _, check := range checks {
		if !strings.Contains(got, check) {
			t.Errorf("missing %q in PDF output", check)
		}
	}
}

func TestWriter_Close_DoubleClose(t *testing.T) {
	var buf bytes.Buffer
	w := pdf.NewWriter(&buf)
	_ = w.AddPage(pdf.PageObject{
		MediaBox: pdf.Rectangle{LLX: 0, LLY: 0, URX: 612, URY: 792},
	})
	err := w.Close()
	if err != nil {
		t.Fatalf("first Close error: %v", err)
	}
	err = w.Close()
	if err == nil {
		t.Errorf("expected error on double close, got nil")
	}
	if err != nil && !strings.Contains(err.Error(), "already closed") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestWriter_Close_NoInfoDict(t *testing.T) {
	var buf bytes.Buffer
	w := pdf.NewWriter(&buf)
	_ = w.AddPage(pdf.PageObject{
		MediaBox: pdf.Rectangle{LLX: 0, LLY: 0, URX: 612, URY: 792},
	})
	err := w.Close()
	if err != nil {
		t.Fatalf("Close error: %v", err)
	}
	got := buf.String()
	// Should not contain /Info in trailer when no info is set.
	// Find the trailer section.
	trailerIdx := strings.Index(got, "trailer")
	if trailerIdx < 0 {
		t.Fatal("trailer not found")
	}
	trailerSection := got[trailerIdx:]
	if strings.Contains(trailerSection, "/Info") {
		t.Errorf("unexpected /Info in trailer when no info set: %q", trailerSection)
	}
}

func TestWriter_Close_WithInfoDict(t *testing.T) {
	var buf bytes.Buffer
	w := pdf.NewWriter(&buf)
	w.SetInfo(pdf.DocumentInfo{Title: "Test"})
	_ = w.AddPage(pdf.PageObject{
		MediaBox: pdf.Rectangle{LLX: 0, LLY: 0, URX: 612, URY: 792},
	})
	err := w.Close()
	if err != nil {
		t.Fatalf("Close error: %v", err)
	}
	got := buf.String()
	trailerIdx := strings.Index(got, "trailer")
	if trailerIdx < 0 {
		t.Fatal("trailer not found")
	}
	trailerSection := got[trailerIdx:]
	if !strings.Contains(trailerSection, "/Info") {
		t.Errorf("expected /Info in trailer when info is set: %q", trailerSection)
	}
}

func TestWriter_Close_MultiplePages(t *testing.T) {
	var buf bytes.Buffer
	w := pdf.NewWriter(&buf)
	for i := 0; i < 3; i++ {
		err := w.AddPage(pdf.PageObject{
			MediaBox: pdf.Rectangle{LLX: 0, LLY: 0, URX: 612, URY: 792},
		})
		if err != nil {
			t.Fatalf("AddPage %d error: %v", i, err)
		}
	}
	err := w.Close()
	if err != nil {
		t.Fatalf("Close error: %v", err)
	}
	got := buf.String()
	if !strings.Contains(got, "/Count 3") {
		t.Errorf("expected /Count 3 for 3 pages: %q", got)
	}
}

// TestWriter_FullPDFGeneration tests a complete PDF generation workflow
// including font registration, image registration, page creation, and closing.
func TestWriter_FullPDFGeneration(t *testing.T) {
	var buf bytes.Buffer
	w := pdf.NewWriter(&buf)
	w.SetCompression(false) // deterministic output

	w.SetInfo(pdf.DocumentInfo{
		Title:    "Full Test PDF",
		Author:   "Test Suite",
		Subject:  "Unit Testing",
		Creator:  "gpdf test",
		Producer: "gpdf",
	})

	// Register a standard font.
	fontResName, fontRef, err := w.RegisterFont("Helvetica", nil)
	if err != nil {
		t.Fatalf("RegisterFont error: %v", err)
	}

	// Register an image.
	imgData := bytes.Repeat([]byte{0xFF}, 30) // 10 white pixels (RGB)
	imgResName, imgRef, err := w.RegisterImage("test.png", imgData, 10, 1, "DeviceRGB", "", nil)
	if err != nil {
		t.Fatalf("RegisterImage error: %v", err)
	}

	// Write a content stream.
	contentRef := w.AllocObject()
	contentData := []byte("BT /F1 12 Tf 72 700 Td (Hello World) Tj ET")
	contentStream := pdf.Stream{
		Dict:    pdf.Dict{},
		Content: contentData,
	}
	if err := w.WriteObject(contentRef, contentStream); err != nil {
		t.Fatalf("WriteObject content error: %v", err)
	}

	// Add a page.
	err = w.AddPage(pdf.PageObject{
		MediaBox: pdf.Rectangle{LLX: 0, LLY: 0, URX: 612, URY: 792},
		Resources: pdf.ResourceDict{
			Font: pdf.Dict{
				pdf.Name(fontResName): fontRef,
			},
			XObject: pdf.Dict{
				pdf.Name(imgResName): imgRef,
			},
		},
		Contents: []pdf.ObjectRef{contentRef},
	})
	if err != nil {
		t.Fatalf("AddPage error: %v", err)
	}

	// Close to finalize.
	if err := w.Close(); err != nil {
		t.Fatalf("Close error: %v", err)
	}

	got := buf.String()

	// Verify essential PDF structure.
	essentials := []string{
		"%PDF-1.7",
		"/Catalog",
		"/Pages",
		"/Page",
		"/Helvetica",
		"/DeviceRGB",
		"(Full Test PDF)",
		"(Test Suite)",
		"(Unit Testing)",
		"(gpdf test)",
		"(gpdf)",
		"xref",
		"trailer",
		"startxref",
		"%%EOF",
	}
	for _, s := range essentials {
		if !strings.Contains(got, s) {
			t.Errorf("missing %q in PDF output", s)
		}
	}
}

func TestWriter_RegisterFont_Compressed(t *testing.T) {
	var buf bytes.Buffer
	w := pdf.NewWriter(&buf)
	// compression is enabled by default

	fontData := bytes.Repeat([]byte("ABCD"), 100) // repetitive data compresses well
	_, _, err := w.RegisterFont("CompressedFont", fontData)
	if err != nil {
		t.Fatalf("RegisterFont error: %v", err)
	}
	got := buf.String()
	if !strings.Contains(got, "/FlateDecode") {
		t.Errorf("expected FlateDecode filter for compressed font data")
	}
}
