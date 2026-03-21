package render

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/gpdf-dev/gpdf/document"
	"github.com/gpdf-dev/gpdf/pdf"
	"github.com/gpdf-dev/gpdf/pdf/font"
)

// ---------------------------------------------------------------------------
// RegisterTTFont tests
// ---------------------------------------------------------------------------

func TestRegisterTTFont(t *testing.T) {
	r, _ := newTestRenderer(t)
	ttfPath := os.Getenv("GPDF_TEST_CJK_FONT")
	if ttfPath == "" {
		t.Skip("GPDF_TEST_CJK_FONT not set, skipping TrueType font test")
	}
	rawData, err := os.ReadFile(ttfPath)
	if err != nil {
		t.Fatalf("failed to read font file: %v", err)
	}
	ttf, err := font.ParseTrueType(rawData)
	if err != nil {
		t.Fatalf("failed to parse TrueType font: %v", err)
	}

	r.RegisterTTFont("NotoSansJP", ttf, rawData)

	if _, ok := r.ttFonts["NotoSansJP"]; !ok {
		t.Error("ttFonts should contain NotoSansJP")
	}
	if _, ok := r.ttFontData["NotoSansJP"]; !ok {
		t.Error("ttFontData should contain NotoSansJP")
	}
}

func TestRegisterTTFont_MultipleRegistrations(t *testing.T) {
	r, _ := newTestRenderer(t)
	ttfPath := os.Getenv("GPDF_TEST_CJK_FONT")
	if ttfPath == "" {
		t.Skip("GPDF_TEST_CJK_FONT not set")
	}
	rawData, err := os.ReadFile(ttfPath)
	if err != nil {
		t.Fatalf("failed to read font: %v", err)
	}
	ttf, err := font.ParseTrueType(rawData)
	if err != nil {
		t.Fatalf("failed to parse font: %v", err)
	}

	r.RegisterTTFont("Font1", ttf, rawData)
	r.RegisterTTFont("Font2", ttf, rawData)

	if len(r.ttFonts) != 2 {
		t.Errorf("ttFonts count = %d, want 2", len(r.ttFonts))
	}
}

// ---------------------------------------------------------------------------
// writeTextBoldSetup tests
// ---------------------------------------------------------------------------

func TestWriteTextBoldSetup_SimulateBold(t *testing.T) {
	r, _ := newTestRenderer(t)
	_ = r.BeginPage(document.Size{Width: 595, Height: 842})

	var buf strings.Builder
	style := document.DefaultStyle()
	style.FontSize = 12
	r.writeTextBoldSetup(&buf, true, style, 12)

	content := buf.String()
	// Should set stroke color and stroke width for bold simulation.
	if !strings.Contains(content, "w\n") {
		t.Error("simulateBold should set stroke width")
	}
	// stroke width = fontSize * 0.03 = 12 * 0.03 = 0.36
	if !strings.Contains(content, "0.36 w") {
		t.Errorf("expected 0.36 w for bold setup, got: %s", content)
	}
}

func TestWriteTextBoldSetup_NoBold(t *testing.T) {
	r, _ := newTestRenderer(t)
	var buf strings.Builder
	style := document.DefaultStyle()
	r.writeTextBoldSetup(&buf, false, style, 12)
	if buf.Len() != 0 {
		t.Errorf("no bold should produce empty output, got: %q", buf.String())
	}
}

// ---------------------------------------------------------------------------
// writeTextBegin tests
// ---------------------------------------------------------------------------

func TestWriteTextBegin_Normal(t *testing.T) {
	r, _ := newTestRenderer(t)
	var buf strings.Builder
	r.writeTextBegin(&buf, "F1", 14, false, false, 72, 700)
	content := buf.String()
	if !strings.Contains(content, "BT\n") {
		t.Error("should contain BT")
	}
	if !strings.Contains(content, "/F1 14 Tf") {
		t.Error("should set font")
	}
	if strings.Contains(content, "Tr") {
		t.Error("normal text should not have text rendering mode")
	}
	if strings.Contains(content, "Tm") {
		t.Error("normal text should not have text matrix")
	}
}

func TestWriteTextBegin_Bold(t *testing.T) {
	r, _ := newTestRenderer(t)
	var buf strings.Builder
	r.writeTextBegin(&buf, "F1", 14, true, false, 72, 700)
	content := buf.String()
	if !strings.Contains(content, "2 Tr") {
		t.Error("bold should set rendering mode 2 (fill+stroke)")
	}
}

func TestWriteTextBegin_Italic(t *testing.T) {
	r, _ := newTestRenderer(t)
	var buf strings.Builder
	r.writeTextBegin(&buf, "F1", 14, false, true, 72, 700)
	content := buf.String()
	if !strings.Contains(content, "Tm\n") {
		t.Error("italic should set text matrix via Tm operator")
	}
	if !strings.Contains(content, "0.2126") {
		t.Error("italic should use 0.2126 shear factor")
	}
}

func TestWriteTextBegin_BoldItalic(t *testing.T) {
	r, _ := newTestRenderer(t)
	var buf strings.Builder
	r.writeTextBegin(&buf, "F1", 14, true, true, 72, 700)
	content := buf.String()
	if !strings.Contains(content, "2 Tr") {
		t.Error("bold+italic should set rendering mode 2")
	}
	if !strings.Contains(content, "Tm") {
		t.Error("bold+italic should set text matrix")
	}
}

// ---------------------------------------------------------------------------
// writeTextSpacing tests
// ---------------------------------------------------------------------------

func TestWriteTextSpacing_Both(t *testing.T) {
	r, _ := newTestRenderer(t)
	var buf strings.Builder
	style := document.DefaultStyle()
	style.WordSpacing = 5
	style.LetterSpacing = 1.5
	r.writeTextSpacing(&buf, style)
	content := buf.String()
	if !strings.Contains(content, "5 Tw") {
		t.Error("should set word spacing")
	}
	if !strings.Contains(content, "1.5 Tc") {
		t.Error("should set letter spacing")
	}
}

func TestWriteTextSpacing_None(t *testing.T) {
	r, _ := newTestRenderer(t)
	var buf strings.Builder
	style := document.DefaultStyle()
	r.writeTextSpacing(&buf, style)
	if buf.Len() != 0 {
		t.Error("no spacing should produce no output")
	}
}

// ---------------------------------------------------------------------------
// writeTextEnd tests
// ---------------------------------------------------------------------------

func TestWriteTextEnd_Normal(t *testing.T) {
	r, _ := newTestRenderer(t)
	var buf strings.Builder
	style := document.DefaultStyle()
	r.writeTextEnd(&buf, style, false)
	content := buf.String()
	if !strings.Contains(content, "ET\n") {
		t.Error("should contain ET")
	}
	if strings.Contains(content, "Tc") {
		t.Error("should not reset Tc when no letter spacing")
	}
	if strings.Contains(content, "Tw") {
		t.Error("should not reset Tw when no word spacing")
	}
	if strings.Contains(content, "Tr") {
		t.Error("should not reset Tr when not bold")
	}
}

func TestWriteTextEnd_WithSpacingAndBold(t *testing.T) {
	r, _ := newTestRenderer(t)
	var buf strings.Builder
	style := document.DefaultStyle()
	style.WordSpacing = 5
	style.LetterSpacing = 1.5
	r.writeTextEnd(&buf, style, true)
	content := buf.String()
	if !strings.Contains(content, "0 Tc") {
		t.Error("should reset Tc")
	}
	if !strings.Contains(content, "0 Tw") {
		t.Error("should reset Tw")
	}
	if !strings.Contains(content, "0 Tr") {
		t.Error("should reset Tr for bold")
	}
	if !strings.Contains(content, "ET") {
		t.Error("should contain ET")
	}
}

// ---------------------------------------------------------------------------
// buildGlyphWidthArray tests
// ---------------------------------------------------------------------------

func TestBuildGlyphWidthArray_Empty(t *testing.T) {
	ttfPath := os.Getenv("GPDF_TEST_CJK_FONT")
	if ttfPath == "" {
		t.Skip("GPDF_TEST_CJK_FONT not set")
	}
	rawData, err := os.ReadFile(ttfPath)
	if err != nil {
		t.Fatal(err)
	}
	ttf, err := font.ParseTrueType(rawData)
	if err != nil {
		t.Fatal(err)
	}
	// Empty runeToGID should produce empty/nil array.
	emptyMap := map[rune]uint16{}
	result := buildGlyphWidthArray(ttf, emptyMap, ttf.Metrics().UnitsPerEm)
	if len(result) != 0 {
		t.Errorf("empty runeToGID should produce empty array, got %d elements", len(result))
	}
}

func TestBuildGlyphWidthArray_WithGlyphs(t *testing.T) {
	ttfPath := os.Getenv("GPDF_TEST_CJK_FONT")
	if ttfPath == "" {
		t.Skip("GPDF_TEST_CJK_FONT not set")
	}
	rawData, err := os.ReadFile(ttfPath)
	if err != nil {
		t.Fatal(err)
	}
	ttf, err := font.ParseTrueType(rawData)
	if err != nil {
		t.Fatal(err)
	}
	// Encode some text to populate usedRunes.
	ttf.Encode("ABC")
	runeToGID := ttf.RuneToGID()
	metrics := ttf.Metrics()
	result := buildGlyphWidthArray(ttf, runeToGID, metrics.UnitsPerEm)
	if len(result) == 0 {
		t.Error("width array should not be empty for encoded glyphs")
	}
}

// ---------------------------------------------------------------------------
// writeCompressedStream tests
// ---------------------------------------------------------------------------

func TestWriteCompressedStream(t *testing.T) {
	var buf bytes.Buffer
	w := pdf.NewWriter(&buf)
	data := []byte("Hello, compressed stream!")
	extraDict := pdf.Dict{
		pdf.Name("Length1"): pdf.Integer(len(data)),
	}
	ref, err := writeCompressedStream(w, data, extraDict)
	if err != nil {
		t.Fatalf("writeCompressedStream error: %v", err)
	}
	if ref.Number == 0 {
		t.Error("should return non-zero ref")
	}
}

func TestWriteCompressedStream_EmptyData(t *testing.T) {
	var buf bytes.Buffer
	w := pdf.NewWriter(&buf)
	ref, err := writeCompressedStream(w, []byte{}, pdf.Dict{})
	if err != nil {
		t.Fatalf("writeCompressedStream error: %v", err)
	}
	if ref.Number == 0 {
		t.Error("should return non-zero ref even for empty data")
	}
}

// ---------------------------------------------------------------------------
// Full TrueType rendering pipeline (writeType0Font, writeFontDescriptor,
// writeCIDFont, writeToUnicodeCMap, subsetFontData)
// ---------------------------------------------------------------------------

func TestFullTTFontRendering(t *testing.T) {
	ttfPath := os.Getenv("GPDF_TEST_CJK_FONT")
	if ttfPath == "" {
		t.Skip("GPDF_TEST_CJK_FONT not set")
	}
	rawData, err := os.ReadFile(ttfPath)
	if err != nil {
		t.Fatalf("failed to read font: %v", err)
	}
	ttf, err := font.ParseTrueType(rawData)
	if err != nil {
		t.Fatalf("failed to parse font: %v", err)
	}

	var buf bytes.Buffer
	w := pdf.NewWriter(&buf)
	r := NewPDFRenderer(w)
	r.RegisterTTFont("TestCJK", ttf, rawData)

	if err := r.BeginDocument(document.DocumentMetadata{Title: "CJK Test"}); err != nil {
		t.Fatal(err)
	}
	if err := r.BeginPage(document.A4); err != nil {
		t.Fatal(err)
	}

	style := document.DefaultStyle()
	style.FontFamily = "TestCJK"
	style.FontSize = 14
	if err := r.RenderText("Hello World", document.Point{X: 72, Y: 72}, style); err != nil {
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
		"%PDF-1.7", "%%EOF", "/Type0", "/CIDFontType2",
		"/ToUnicode", "/FontDescriptor", "/FontFile2",
	} {
		if !strings.Contains(output, want) {
			t.Errorf("output should contain %q", want)
		}
	}
}

func TestTTFontRendering_CJKText(t *testing.T) {
	ttfPath := os.Getenv("GPDF_TEST_CJK_FONT")
	if ttfPath == "" {
		t.Skip("GPDF_TEST_CJK_FONT not set")
	}
	rawData, err := os.ReadFile(ttfPath)
	if err != nil {
		t.Fatal(err)
	}
	ttf, err := font.ParseTrueType(rawData)
	if err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer
	w := pdf.NewWriter(&buf)
	r := NewPDFRenderer(w)
	r.RegisterTTFont("CJKFont", ttf, rawData)

	if err := r.BeginDocument(document.DocumentMetadata{}); err != nil {
		t.Fatal(err)
	}
	if err := r.BeginPage(document.A4); err != nil {
		t.Fatal(err)
	}

	style := document.DefaultStyle()
	style.FontFamily = "CJKFont"
	style.FontSize = 12
	// Render Japanese text.
	if err := r.RenderText("こんにちは世界", document.Point{X: 72, Y: 72}, style); err != nil {
		t.Fatal(err)
	}

	if err := r.EndPage(); err != nil {
		t.Fatal(err)
	}
	if err := r.EndDocument(); err != nil {
		t.Fatal(err)
	}

	output := buf.String()
	// CJK text should be encoded as hex glyph IDs, not literal strings.
	if strings.Contains(output, "(こんにちは世界)") {
		t.Error("CJK text should be encoded as hex glyph IDs, not literal")
	}
	if !strings.Contains(output, "/Identity-H") {
		t.Error("should use Identity-H encoding")
	}
}

func TestTTFontRendering_BoldSimulation(t *testing.T) {
	ttfPath := os.Getenv("GPDF_TEST_CJK_FONT")
	if ttfPath == "" {
		t.Skip("GPDF_TEST_CJK_FONT not set")
	}
	rawData, err := os.ReadFile(ttfPath)
	if err != nil {
		t.Fatal(err)
	}
	ttf, err := font.ParseTrueType(rawData)
	if err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer
	w := pdf.NewWriter(&buf)
	r := NewPDFRenderer(w)
	r.RegisterTTFont("CJKFont", ttf, rawData)

	if err := r.BeginDocument(document.DocumentMetadata{}); err != nil {
		t.Fatal(err)
	}
	if err := r.BeginPage(document.A4); err != nil {
		t.Fatal(err)
	}

	style := document.DefaultStyle()
	style.FontFamily = "CJKFont"
	style.FontSize = 12
	style.FontWeight = document.WeightBold
	if err := r.RenderText("Bold CJK", document.Point{X: 72, Y: 72}, style); err != nil {
		t.Fatal(err)
	}

	content := string(r.pageContent)
	// Bold simulation should use text rendering mode 2 (fill+stroke).
	if !strings.Contains(content, "2 Tr") {
		t.Error("bold CJK should use rendering mode 2")
	}
	// Should reset rendering mode after text.
	if !strings.Contains(content, "0 Tr") {
		t.Error("should reset rendering mode after bold text")
	}
}

func TestTTFontRendering_ItalicSimulation(t *testing.T) {
	ttfPath := os.Getenv("GPDF_TEST_CJK_FONT")
	if ttfPath == "" {
		t.Skip("GPDF_TEST_CJK_FONT not set")
	}
	rawData, err := os.ReadFile(ttfPath)
	if err != nil {
		t.Fatal(err)
	}
	ttf, err := font.ParseTrueType(rawData)
	if err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer
	w := pdf.NewWriter(&buf)
	r := NewPDFRenderer(w)
	r.RegisterTTFont("CJKFont", ttf, rawData)

	if err := r.BeginDocument(document.DocumentMetadata{}); err != nil {
		t.Fatal(err)
	}
	if err := r.BeginPage(document.A4); err != nil {
		t.Fatal(err)
	}

	style := document.DefaultStyle()
	style.FontFamily = "CJKFont"
	style.FontSize = 12
	style.FontStyle = document.StyleItalic
	if err := r.RenderText("Italic CJK", document.Point{X: 72, Y: 72}, style); err != nil {
		t.Fatal(err)
	}

	content := string(r.pageContent)
	// Italic simulation should use Tm operator with shear.
	if !strings.Contains(content, "Tm") {
		t.Error("italic CJK should use Tm operator for shear")
	}
	if !strings.Contains(content, "0.2126") {
		t.Error("italic should use 0.2126 shear factor")
	}
}

// ---------------------------------------------------------------------------
// subsetFontData tests
// ---------------------------------------------------------------------------

func TestSubsetFontData(t *testing.T) {
	ttfPath := os.Getenv("GPDF_TEST_CJK_FONT")
	if ttfPath == "" {
		t.Skip("GPDF_TEST_CJK_FONT not set")
	}
	rawData, err := os.ReadFile(ttfPath)
	if err != nil {
		t.Fatal(err)
	}
	ttf, err := font.ParseTrueType(rawData)
	if err != nil {
		t.Fatal(err)
	}

	r, _ := newTestRenderer(t)
	r.RegisterTTFont("TestFont", ttf, rawData)

	// Encode a few characters to populate usedRunes.
	ttf.Encode("ABC")

	subsetData := r.subsetFontData(ttf, rawData)
	if len(subsetData) == 0 {
		t.Error("subset data should not be empty")
	}
	// Subset should be smaller than original for a large font.
	if len(subsetData) >= len(rawData) {
		t.Logf("subset (%d bytes) not smaller than original (%d bytes) - may be expected for small fonts", len(subsetData), len(rawData))
	}
}

func TestSubsetFontData_NoUsedRunes(t *testing.T) {
	ttfPath := os.Getenv("GPDF_TEST_CJK_FONT")
	if ttfPath == "" {
		t.Skip("GPDF_TEST_CJK_FONT not set")
	}
	rawData, err := os.ReadFile(ttfPath)
	if err != nil {
		t.Fatal(err)
	}
	ttf, err := font.ParseTrueType(rawData)
	if err != nil {
		t.Fatal(err)
	}

	r, _ := newTestRenderer(t)
	// No characters encoded, usedRunes is empty.
	subsetData := r.subsetFontData(ttf, rawData)
	if subsetData == nil {
		t.Error("subset data should not be nil")
	}
}

// ---------------------------------------------------------------------------
// writeFontDescriptor tests
// ---------------------------------------------------------------------------

func TestWriteFontDescriptor(t *testing.T) {
	ttfPath := os.Getenv("GPDF_TEST_CJK_FONT")
	if ttfPath == "" {
		t.Skip("GPDF_TEST_CJK_FONT not set")
	}
	rawData, err := os.ReadFile(ttfPath)
	if err != nil {
		t.Fatal(err)
	}
	ttf, err := font.ParseTrueType(rawData)
	if err != nil {
		t.Fatal(err)
	}

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
// writeCIDFont tests
// ---------------------------------------------------------------------------

func TestWriteCIDFont(t *testing.T) {
	var buf bytes.Buffer
	w := pdf.NewWriter(&buf)
	r := NewPDFRenderer(w)

	descRef := w.AllocObject()
	wArray := pdf.Array{pdf.Integer(1), pdf.Array{pdf.Integer(500)}}

	cidRef, err := r.writeCIDFont(w, "TestFont", descRef, 1000, wArray)
	if err != nil {
		t.Fatalf("writeCIDFont error: %v", err)
	}
	if cidRef.Number == 0 {
		t.Error("CIDFont ref should be non-zero")
	}
}

func TestWriteCIDFont_EmptyWidths(t *testing.T) {
	var buf bytes.Buffer
	w := pdf.NewWriter(&buf)
	r := NewPDFRenderer(w)

	descRef := w.AllocObject()

	cidRef, err := r.writeCIDFont(w, "TestFont", descRef, 1000, nil)
	if err != nil {
		t.Fatalf("writeCIDFont error: %v", err)
	}
	if cidRef.Number == 0 {
		t.Error("CIDFont ref should be non-zero")
	}
}

// ---------------------------------------------------------------------------
// writeToUnicodeCMap tests
// ---------------------------------------------------------------------------

func TestWriteToUnicodeCMap(t *testing.T) {
	var buf bytes.Buffer
	w := pdf.NewWriter(&buf)

	runeToGID := map[rune]uint16{
		'A': 1,
		'B': 2,
		'C': 3,
	}
	ref, err := writeToUnicodeCMap(w, runeToGID)
	if err != nil {
		t.Fatalf("writeToUnicodeCMap error: %v", err)
	}
	if ref.Number == 0 {
		t.Error("ToUnicode ref should be non-zero")
	}
}

func TestWriteToUnicodeCMap_Empty(t *testing.T) {
	var buf bytes.Buffer
	w := pdf.NewWriter(&buf)

	ref, err := writeToUnicodeCMap(w, map[rune]uint16{})
	if err != nil {
		t.Fatalf("writeToUnicodeCMap error: %v", err)
	}
	if ref.Number == 0 {
		t.Error("ToUnicode ref should be non-zero even for empty map")
	}
}

// ---------------------------------------------------------------------------
// resolveTextFont tests
// ---------------------------------------------------------------------------

func TestResolveTextFont_StandardFont(t *testing.T) {
	r, _ := newTestRenderer(t)
	style := document.DefaultStyle()
	style.FontFamily = "Helvetica"
	style.FontWeight = document.WeightBold

	fontName, _, simulateBold, simulateItalic := r.resolveTextFont(style)
	if fontName != "Helvetica-Bold" {
		t.Errorf("fontName = %q, want 'Helvetica-Bold'", fontName)
	}
	if simulateBold {
		t.Error("standard bold variant should not simulate bold")
	}
	if simulateItalic {
		t.Error("should not simulate italic")
	}
}

func TestResolveTextFont_TTFont_BoldSimulation(t *testing.T) {
	ttfPath := os.Getenv("GPDF_TEST_CJK_FONT")
	if ttfPath == "" {
		t.Skip("GPDF_TEST_CJK_FONT not set")
	}
	rawData, err := os.ReadFile(ttfPath)
	if err != nil {
		t.Fatal(err)
	}
	ttf, err := font.ParseTrueType(rawData)
	if err != nil {
		t.Fatal(err)
	}

	r, _ := newTestRenderer(t)
	r.RegisterTTFont("CJKFont", ttf, rawData)

	style := document.DefaultStyle()
	style.FontFamily = "CJKFont"
	style.FontWeight = document.WeightBold

	fontName, ttFontName, simulateBold, simulateItalic := r.resolveTextFont(style)
	if fontName != "CJKFont" {
		t.Errorf("fontName = %q, want 'CJKFont'", fontName)
	}
	if ttFontName != "CJKFont" {
		t.Errorf("ttFontName = %q, want 'CJKFont'", ttFontName)
	}
	if !simulateBold {
		t.Error("CJK bold should use simulation")
	}
	if simulateItalic {
		t.Error("should not simulate italic")
	}
}

func TestResolveTextFont_TTFont_ItalicSimulation(t *testing.T) {
	ttfPath := os.Getenv("GPDF_TEST_CJK_FONT")
	if ttfPath == "" {
		t.Skip("GPDF_TEST_CJK_FONT not set")
	}
	rawData, err := os.ReadFile(ttfPath)
	if err != nil {
		t.Fatal(err)
	}
	ttf, err := font.ParseTrueType(rawData)
	if err != nil {
		t.Fatal(err)
	}

	r, _ := newTestRenderer(t)
	r.RegisterTTFont("CJKFont", ttf, rawData)

	style := document.DefaultStyle()
	style.FontFamily = "CJKFont"
	style.FontStyle = document.StyleItalic

	_, _, simulateBold, simulateItalic := r.resolveTextFont(style)
	if simulateBold {
		t.Error("should not simulate bold")
	}
	if !simulateItalic {
		t.Error("CJK italic should use simulation")
	}
}
