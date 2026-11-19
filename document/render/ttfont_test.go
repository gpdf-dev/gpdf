package render

import (
	"bytes"
	"strings"
	"testing"

	"github.com/gpdf-dev/gpdf/document"
	"github.com/gpdf-dev/gpdf/document/layout"
	"github.com/gpdf-dev/gpdf/internal/testfont"
	"github.com/gpdf-dev/gpdf/pdf"
	"github.com/gpdf-dev/gpdf/pdf/font"
)

// buildAndParseTTF is a test helper that builds a minimal TTF and parses it.
func buildAndParseTTF(t *testing.T) (*font.TrueTypeFont, []byte) {
	t.Helper()
	runeMap := map[rune]uint16{
		'A': 1,
		'B': 2,
		'C': 3,
		' ': 4,
	}
	widths := []uint16{0, 600, 700, 500, 250} // .notdef=0, A=600, B=700, C=500, space=250
	data := testfont.Build(5, widths, runeMap)

	ttf, err := font.ParseTrueType(data)
	if err != nil {
		t.Fatalf("ParseTrueType failed on test font: %v", err)
	}
	return ttf, data
}

// ---------------------------------------------------------------------------
// RegisterTTFont tests (no env var needed)
// ---------------------------------------------------------------------------

func TestRegisterTTFont_InMemory(t *testing.T) {
	ttf, rawData := buildAndParseTTF(t)
	r, _ := newTestRenderer(t)

	r.RegisterTTFont("TestFont", ttf, rawData)

	if _, ok := r.ttFonts["TestFont"]; !ok {
		t.Error("ttFonts should contain TestFont")
	}
	if _, ok := r.ttFontData["TestFont"]; !ok {
		t.Error("ttFontData should contain TestFont")
	}
}

func TestRegisterTTFont_MultipleInMemory(t *testing.T) {
	ttf, rawData := buildAndParseTTF(t)
	r, _ := newTestRenderer(t)

	r.RegisterTTFont("Font1", ttf, rawData)
	r.RegisterTTFont("Font2", ttf, rawData)

	if len(r.ttFonts) != 2 {
		t.Errorf("ttFonts count = %d, want 2", len(r.ttFonts))
	}
	if len(r.ttFontData) != 2 {
		t.Errorf("ttFontData count = %d, want 2", len(r.ttFontData))
	}
}

// ---------------------------------------------------------------------------
// resolveTextFont with TT font tests
// ---------------------------------------------------------------------------

func TestResolveTextFont_TTFont_Bold(t *testing.T) {
	ttf, rawData := buildAndParseTTF(t)
	r, _ := newTestRenderer(t)
	r.RegisterTTFont("MyFont", ttf, rawData)

	style := document.DefaultStyle()
	style.FontFamily = "MyFont"
	style.FontWeight = document.WeightBold

	fontName, ttFontName, simulateBold, simulateItalic := r.resolveTextFont(style)
	if fontName != "MyFont" {
		t.Errorf("fontName = %q, want MyFont", fontName)
	}
	if ttFontName != "MyFont" {
		t.Errorf("ttFontName = %q, want MyFont", ttFontName)
	}
	if !simulateBold {
		t.Error("bold TT font should simulate bold")
	}
	if simulateItalic {
		t.Error("should not simulate italic")
	}
}

func TestResolveTextFont_TTFont_Italic(t *testing.T) {
	ttf, rawData := buildAndParseTTF(t)
	r, _ := newTestRenderer(t)
	r.RegisterTTFont("MyFont", ttf, rawData)

	style := document.DefaultStyle()
	style.FontFamily = "MyFont"
	style.FontStyle = document.StyleItalic

	_, _, simulateBold, simulateItalic := r.resolveTextFont(style)
	if simulateBold {
		t.Error("should not simulate bold")
	}
	if !simulateItalic {
		t.Error("italic TT font should simulate italic")
	}
}

func TestResolveTextFont_TTFont_BoldItalic(t *testing.T) {
	ttf, rawData := buildAndParseTTF(t)
	r, _ := newTestRenderer(t)
	r.RegisterTTFont("MyFont", ttf, rawData)

	style := document.DefaultStyle()
	style.FontFamily = "MyFont"
	style.FontWeight = document.WeightBold
	style.FontStyle = document.StyleItalic

	_, _, simulateBold, simulateItalic := r.resolveTextFont(style)
	if !simulateBold {
		t.Error("bold+italic TT font should simulate bold")
	}
	if !simulateItalic {
		t.Error("bold+italic TT font should simulate italic")
	}
}

func TestResolveTextFont_TTFont_Normal(t *testing.T) {
	ttf, rawData := buildAndParseTTF(t)
	r, _ := newTestRenderer(t)
	r.RegisterTTFont("MyFont", ttf, rawData)

	style := document.DefaultStyle()
	style.FontFamily = "MyFont"

	fontName, ttFontName, simulateBold, simulateItalic := r.resolveTextFont(style)
	if fontName != "MyFont" {
		t.Errorf("fontName = %q, want MyFont", fontName)
	}
	if ttFontName != "MyFont" {
		t.Errorf("ttFontName = %q, want MyFont", ttFontName)
	}
	if simulateBold {
		t.Error("normal should not simulate bold")
	}
	if simulateItalic {
		t.Error("normal should not simulate italic")
	}
}

// ---------------------------------------------------------------------------
// ensureFont with TT font (triggers OnBeforeClose hook registration)
// ---------------------------------------------------------------------------

func TestEnsureFont_TTFont(t *testing.T) {
	ttf, rawData := buildAndParseTTF(t)
	r, _ := newTestRenderer(t)
	r.RegisterTTFont("MyTTFont", ttf, rawData)

	resName, err := r.ensureFont("MyTTFont")
	if err != nil {
		t.Fatalf("ensureFont error: %v", err)
	}
	if resName == "" {
		t.Error("resName should not be empty")
	}
	if _, ok := r.fontMap["MyTTFont"]; !ok {
		t.Error("fontMap should contain MyTTFont")
	}
	if _, ok := r.fontRefs["MyTTFont"]; !ok {
		t.Error("fontRefs should contain MyTTFont")
	}
}

func TestEnsureFont_TTFont_CachedSecondCall(t *testing.T) {
	ttf, rawData := buildAndParseTTF(t)
	r, _ := newTestRenderer(t)
	r.RegisterTTFont("MyTTFont", ttf, rawData)

	resName1, err := r.ensureFont("MyTTFont")
	if err != nil {
		t.Fatal(err)
	}
	resName2, err := r.ensureFont("MyTTFont")
	if err != nil {
		t.Fatal(err)
	}
	if resName1 != resName2 {
		t.Errorf("second call should return cached name: %q vs %q", resName1, resName2)
	}
}

// ---------------------------------------------------------------------------
// buildGlyphWidthArray tests (in-memory font)
// ---------------------------------------------------------------------------

func TestBuildGlyphWidthArray_InMemory_Empty(t *testing.T) {
	ttf, _ := buildAndParseTTF(t)
	emptyMap := map[rune]uint16{}
	result := buildGlyphWidthArray(ttf, emptyMap, nil, ttf.Metrics().UnitsPerEm)
	if len(result) != 0 {
		t.Errorf("empty runeToGID should produce empty array, got %d elements", len(result))
	}
}

func TestBuildGlyphWidthArray_InMemory_WithGlyphs(t *testing.T) {
	ttf, _ := buildAndParseTTF(t)
	// Encode some text to populate usedRunes.
	ttf.Encode("AB")
	runeToGID := ttf.RuneToGID()
	metrics := ttf.Metrics()
	result := buildGlyphWidthArray(ttf, runeToGID, nil, metrics.UnitsPerEm)
	if len(result) == 0 {
		t.Error("width array should not be empty for encoded glyphs")
	}
}

func TestBuildGlyphWidthArray_InMemory_ConsecutiveGIDs(t *testing.T) {
	ttf, _ := buildAndParseTTF(t)
	ttf.Encode("ABC")
	runeToGID := ttf.RuneToGID()
	metrics := ttf.Metrics()
	result := buildGlyphWidthArray(ttf, runeToGID, nil, metrics.UnitsPerEm)
	// GIDs 1, 2, 3 are consecutive, should be grouped in one W entry: [startGID [w1 w2 w3]]
	if len(result) < 2 {
		t.Fatalf("expected at least 2 elements in W array, got %d", len(result))
	}
}

// ---------------------------------------------------------------------------
// subsetFontData tests (in-memory font)
// ---------------------------------------------------------------------------

func TestSubsetFontData_InMemory(t *testing.T) {
	ttf, rawData := buildAndParseTTF(t)
	r, _ := newTestRenderer(t)
	r.RegisterTTFont("TestFont", ttf, rawData)

	// Encode characters to populate usedRunes.
	ttf.Encode("ABC")

	subsetData := subsetFontData(ttf, rawData)
	if len(subsetData) == 0 {
		t.Error("subset data should not be empty")
	}
}

func TestSubsetFontData_InMemory_NoUsedRunes(t *testing.T) {
	ttf, rawData := buildAndParseTTF(t)
	// No characters encoded.
	subsetData := subsetFontData(ttf, rawData)
	if subsetData == nil {
		t.Error("subset data should not be nil (fallback to raw data)")
	}
}

// ---------------------------------------------------------------------------
// writeType0Font integration test (via EndDocument triggering beforeClose)
// ---------------------------------------------------------------------------

func TestWriteType0Font_Integration(t *testing.T) {
	ttf, rawData := buildAndParseTTF(t)

	var buf bytes.Buffer
	w := pdf.NewWriter(&buf)
	r := NewPDFRenderer(w)
	r.RegisterTTFont("InMemFont", ttf, rawData)

	if err := r.BeginDocument(document.DocumentMetadata{Title: "TTF Test"}); err != nil {
		t.Fatal(err)
	}
	if err := r.BeginPage(document.A4); err != nil {
		t.Fatal(err)
	}

	style := document.DefaultStyle()
	style.FontFamily = "InMemFont"
	style.FontSize = 14
	if err := r.RenderText("AB C", document.Point{X: 72, Y: 72}, style); err != nil {
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
		"%PDF-1.7",
		"%%EOF",
		"/Type0",
		"/CIDFontType2",
		"/ToUnicode",
		"/FontDescriptor",
		"/FontFile2",
		"/Identity-H",
	} {
		if !strings.Contains(output, want) {
			t.Errorf("output should contain %q", want)
		}
	}
}

// ---------------------------------------------------------------------------
// writeFontDescriptor test (in-memory)
// ---------------------------------------------------------------------------

func TestWriteFontDescriptor_InMemory(t *testing.T) {
	ttf, _ := buildAndParseTTF(t)

	var buf bytes.Buffer
	w := pdf.NewWriter(&buf)
	metrics := ttf.Metrics()

	fontFileRef := w.AllocObject()
	descRef, err := writeFontDescriptor(w, "TestFont", metrics, fontFileRef)
	if err != nil {
		t.Fatalf("writeFontDescriptor error: %v", err)
	}
	if descRef.Number == 0 {
		t.Error("descriptor ref should be non-zero")
	}
}

// ---------------------------------------------------------------------------
// RenderText with TT font (hex encoding path)
// ---------------------------------------------------------------------------

func TestRenderText_TTFont_HexEncoding(t *testing.T) {
	ttf, rawData := buildAndParseTTF(t)

	var buf bytes.Buffer
	w := pdf.NewWriter(&buf)
	r := NewPDFRenderer(w)
	r.RegisterTTFont("TestTT", ttf, rawData)

	if err := r.BeginPage(document.Size{Width: 595, Height: 842}); err != nil {
		t.Fatal(err)
	}

	style := document.DefaultStyle()
	style.FontFamily = "TestTT"
	style.FontSize = 12
	if err := r.RenderText("ABC", document.Point{X: 50, Y: 50}, style); err != nil {
		t.Fatal(err)
	}

	content := string(r.pageContent)
	// TT font text should use hex encoding <XX> Tj, not literal string () Tj.
	if strings.Contains(content, "(ABC) Tj") {
		t.Error("TrueType font should not use literal string encoding")
	}
	if !strings.Contains(content, "> Tj") {
		t.Error("expected hex-encoded text with > Tj suffix")
	}
}

// ---------------------------------------------------------------------------
// RenderText with TT font + bold simulation
// ---------------------------------------------------------------------------

func TestRenderText_TTFont_BoldSimulation(t *testing.T) {
	ttf, rawData := buildAndParseTTF(t)

	var buf bytes.Buffer
	w := pdf.NewWriter(&buf)
	r := NewPDFRenderer(w)
	r.RegisterTTFont("TestTT", ttf, rawData)

	if err := r.BeginPage(document.Size{Width: 595, Height: 842}); err != nil {
		t.Fatal(err)
	}

	style := document.DefaultStyle()
	style.FontFamily = "TestTT"
	style.FontSize = 12
	style.FontWeight = document.WeightBold
	if err := r.RenderText("AB", document.Point{X: 50, Y: 50}, style); err != nil {
		t.Fatal(err)
	}

	content := string(r.pageContent)
	if !strings.Contains(content, "2 Tr") {
		t.Error("bold TT should use rendering mode 2")
	}
	if !strings.Contains(content, "0 Tr") {
		t.Error("should reset rendering mode after bold text")
	}
}

// ---------------------------------------------------------------------------
// RenderText with TT font + italic simulation
// ---------------------------------------------------------------------------

func TestRenderText_TTFont_ItalicSimulation(t *testing.T) {
	ttf, rawData := buildAndParseTTF(t)

	var buf bytes.Buffer
	w := pdf.NewWriter(&buf)
	r := NewPDFRenderer(w)
	r.RegisterTTFont("TestTT", ttf, rawData)

	if err := r.BeginPage(document.Size{Width: 595, Height: 842}); err != nil {
		t.Fatal(err)
	}

	style := document.DefaultStyle()
	style.FontFamily = "TestTT"
	style.FontSize = 12
	style.FontStyle = document.StyleItalic
	if err := r.RenderText("AB", document.Point{X: 50, Y: 50}, style); err != nil {
		t.Fatal(err)
	}

	content := string(r.pageContent)
	if !strings.Contains(content, "Tm") {
		t.Error("italic TT should use Tm operator")
	}
	if !strings.Contains(content, "0.2126") {
		t.Error("italic should use 0.2126 shear factor")
	}
}

// ---------------------------------------------------------------------------
// RenderDocument with TT font (multi-page, full pipeline)
// ---------------------------------------------------------------------------

func TestRenderDocument_TTFont_MultiPage(t *testing.T) {
	ttf, rawData := buildAndParseTTF(t)

	var buf bytes.Buffer
	w := pdf.NewWriter(&buf)
	r := NewPDFRenderer(w)
	r.RegisterTTFont("TestTT", ttf, rawData)

	textNode1 := &document.Text{
		Content: "Page one",
		TextStyle: document.Style{
			FontFamily: "TestTT",
			FontSize:   12,
			Color:      pdf.Black,
		},
	}
	textNode2 := &document.Text{
		Content: "ABC",
		TextStyle: document.Style{
			FontFamily: "TestTT",
			FontSize:   14,
			Color:      pdf.Black,
		},
	}

	pages := []layout.PageLayout{
		{
			Size: document.Size{Width: 595, Height: 842},
			Children: []layout.PlacedNode{
				{
					Node:     textNode1,
					Position: document.Point{X: 72, Y: 72},
					Size:     document.Size{Width: 200, Height: 20},
				},
			},
		},
		{
			Size: document.Size{Width: 595, Height: 842},
			Children: []layout.PlacedNode{
				{
					Node:     textNode2,
					Position: document.Point{X: 72, Y: 100},
					Size:     document.Size{Width: 200, Height: 20},
				},
			},
		},
	}

	if err := r.RenderDocument(pages, document.DocumentMetadata{Title: "Multi-page TT"}); err != nil {
		t.Fatalf("RenderDocument failed: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "/Type0") {
		t.Error("output should contain Type0 font")
	}
	if !strings.Contains(output, "%PDF-1.7") {
		t.Error("output should contain PDF header")
	}
	if !strings.Contains(output, "%%EOF") {
		t.Error("output should contain EOF marker")
	}
}

// ---------------------------------------------------------------------------
// RenderDocument with empty pages list
// ---------------------------------------------------------------------------

func TestRenderDocument_EmptyPages(t *testing.T) {
	var buf bytes.Buffer
	w := pdf.NewWriter(&buf)
	r := NewPDFRenderer(w)

	err := r.RenderDocument(nil, document.DocumentMetadata{Title: "Empty"})
	if err != nil {
		t.Fatalf("RenderDocument with no pages should not error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "%PDF-1.7") {
		t.Error("should produce valid PDF even with no pages")
	}
}

// ---------------------------------------------------------------------------
// RenderDocument with mixed standard + TT font nodes
// ---------------------------------------------------------------------------

func TestRenderDocument_MixedFonts(t *testing.T) {
	ttf, rawData := buildAndParseTTF(t)

	var buf bytes.Buffer
	w := pdf.NewWriter(&buf)
	r := NewPDFRenderer(w)
	r.RegisterTTFont("TestTT", ttf, rawData)

	stdTextNode := &document.Text{
		Content: "Standard text",
		TextStyle: document.Style{
			FontFamily: "Helvetica",
			FontSize:   12,
			Color:      pdf.Black,
		},
	}
	ttTextNode := &document.Text{
		Content: "ABC",
		TextStyle: document.Style{
			FontFamily: "TestTT",
			FontSize:   12,
			Color:      pdf.Black,
		},
	}

	pages := []layout.PageLayout{
		{
			Size: document.Size{Width: 595, Height: 842},
			Children: []layout.PlacedNode{
				{
					Node:     stdTextNode,
					Position: document.Point{X: 72, Y: 72},
					Size:     document.Size{Width: 200, Height: 20},
				},
				{
					Node:     ttTextNode,
					Position: document.Point{X: 72, Y: 100},
					Size:     document.Size{Width: 200, Height: 20},
				},
			},
		},
	}

	if err := r.RenderDocument(pages, document.DocumentMetadata{Title: "Mixed"}); err != nil {
		t.Fatalf("RenderDocument failed: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "/Type0") {
		t.Error("should contain Type0 for TT font")
	}
}

// ---------------------------------------------------------------------------
// RenderDocument with background and border nodes
// ---------------------------------------------------------------------------

func TestRenderDocument_WithBackgroundAndBorders(t *testing.T) {
	var buf bytes.Buffer
	w := pdf.NewWriter(&buf)
	r := NewPDFRenderer(w)

	bg := pdf.Color{R: 0.9, G: 0.9, B: 0.9}
	borderColor := pdf.Red
	textNode := &document.Text{
		Content: "Styled",
		TextStyle: document.Style{
			FontFamily: "Helvetica",
			FontSize:   12,
			Color:      pdf.Black,
			Background: &bg,
			Border: document.BorderEdges{
				Top: document.BorderSide{
					Style: document.BorderSolid,
					Width: document.Pt(1),
					Color: borderColor,
				},
				Bottom: document.BorderSide{
					Style: document.BorderSolid,
					Width: document.Pt(1),
					Color: borderColor,
				},
				Left: document.BorderSide{
					Style: document.BorderSolid,
					Width: document.Pt(1),
					Color: borderColor,
				},
				Right: document.BorderSide{
					Style: document.BorderSolid,
					Width: document.Pt(1),
					Color: borderColor,
				},
			},
		},
	}

	pages := []layout.PageLayout{
		{
			Size: document.Size{Width: 595, Height: 842},
			Children: []layout.PlacedNode{
				{
					Node:     textNode,
					Position: document.Point{X: 50, Y: 50},
					Size:     document.Size{Width: 100, Height: 20},
				},
			},
		},
	}

	if err := r.RenderDocument(pages, document.DocumentMetadata{}); err != nil {
		t.Fatalf("RenderDocument failed: %v", err)
	}
}

// ---------------------------------------------------------------------------
// RenderDocument with nil node (should be handled gracefully)
// ---------------------------------------------------------------------------

func TestRenderDocument_NilNode(t *testing.T) {
	var buf bytes.Buffer
	w := pdf.NewWriter(&buf)
	r := NewPDFRenderer(w)

	pages := []layout.PageLayout{
		{
			Size: document.Size{Width: 595, Height: 842},
			Children: []layout.PlacedNode{
				{
					Node:     nil,
					Position: document.Point{X: 50, Y: 50},
					Size:     document.Size{Width: 100, Height: 20},
				},
			},
		},
	}

	if err := r.RenderDocument(pages, document.DocumentMetadata{}); err != nil {
		t.Fatalf("RenderDocument with nil node should not error: %v", err)
	}
}

// ---------------------------------------------------------------------------
// RenderDocument with nested children
// ---------------------------------------------------------------------------

func TestRenderDocument_NestedChildren(t *testing.T) {
	var buf bytes.Buffer
	w := pdf.NewWriter(&buf)
	r := NewPDFRenderer(w)

	childText := &document.Text{
		Content: "Child text",
		TextStyle: document.Style{
			FontFamily: "Helvetica",
			FontSize:   10,
			Color:      pdf.Black,
		},
	}

	parentText := &document.Text{
		Content: "Parent",
		TextStyle: document.Style{
			FontFamily: "Helvetica",
			FontSize:   12,
			Color:      pdf.Black,
		},
	}

	pages := []layout.PageLayout{
		{
			Size: document.Size{Width: 595, Height: 842},
			Children: []layout.PlacedNode{
				{
					Node:     parentText,
					Position: document.Point{X: 50, Y: 50},
					Size:     document.Size{Width: 200, Height: 40},
					Children: []layout.PlacedNode{
						{
							Node:     childText,
							Position: document.Point{X: 10, Y: 20},
							Size:     document.Size{Width: 100, Height: 15},
						},
					},
				},
			},
		},
	}

	if err := r.RenderDocument(pages, document.DocumentMetadata{}); err != nil {
		t.Fatalf("RenderDocument with nested children failed: %v", err)
	}
}

// ---------------------------------------------------------------------------
// RenderDocument with text decoration
// ---------------------------------------------------------------------------

func TestRenderDocument_TextDecoration(t *testing.T) {
	var buf bytes.Buffer
	w := pdf.NewWriter(&buf)
	r := NewPDFRenderer(w)

	textNode := &document.Text{
		Content: "Underlined",
		TextStyle: document.Style{
			FontFamily:     "Helvetica",
			FontSize:       12,
			Color:          pdf.Black,
			TextDecoration: document.DecorationUnderline | document.DecorationStrikethrough | document.DecorationOverline,
		},
	}

	pages := []layout.PageLayout{
		{
			Size: document.Size{Width: 595, Height: 842},
			Children: []layout.PlacedNode{
				{
					Node:     textNode,
					Position: document.Point{X: 50, Y: 50},
					Size:     document.Size{Width: 100, Height: 20},
				},
			},
		},
	}

	if err := r.RenderDocument(pages, document.DocumentMetadata{}); err != nil {
		t.Fatalf("RenderDocument failed: %v", err)
	}
}

// ---------------------------------------------------------------------------
// writePaintOp coverage (no-fill-no-stroke path)
// ---------------------------------------------------------------------------

func TestWritePaintOp_NoFillNoStroke(t *testing.T) {
	var buf strings.Builder
	writePaintOp(&buf, false, false)
	if !strings.Contains(buf.String(), "n\n") {
		t.Error("no fill + no stroke should produce 'n' operator")
	}
}

func TestWritePaintOp_StrokeOnly(t *testing.T) {
	var buf strings.Builder
	writePaintOp(&buf, false, true)
	if !strings.Contains(buf.String(), "S\n") {
		t.Error("stroke only should produce 'S' operator")
	}
}
