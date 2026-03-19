package render

import (
	"bytes"
	"encoding/hex"
	"os"
	"strings"
	"testing"

	"github.com/gpdf-dev/gpdf/document"
	"github.com/gpdf-dev/gpdf/document/layout"
	"github.com/gpdf-dev/gpdf/pdf"
	"github.com/gpdf-dev/gpdf/pdf/font"
)

// TestTrueTypeFontHexEncoding verifies that RenderText uses hex-encoded
// glyph IDs (<XXXX> Tj) for TrueType fonts instead of WinAnsi literal strings.
func TestTrueTypeFontHexEncoding(t *testing.T) {
	fontPath := os.Getenv("GPDF_TEST_CJK_FONT")
	if fontPath == "" {
		t.Skip("set GPDF_TEST_CJK_FONT to a CJK TTF path")
	}

	fontData, err := os.ReadFile(fontPath)
	if err != nil {
		t.Fatalf("failed to read font file: %v", err)
	}

	ttf, err := font.ParseTrueType(fontData)
	if err != nil {
		t.Fatalf("failed to parse font: %v", err)
	}

	var buf bytes.Buffer
	pw := pdf.NewWriter(&buf)
	renderer := NewPDFRenderer(pw)
	renderer.RegisterTTFont("TestCJK", ttf, fontData)

	// Begin a page to set up pageHeight for Y conversion.
	if err = renderer.BeginPage(document.Size{Width: 595, Height: 842}); err != nil {
		t.Fatalf("BeginPage failed: %v", err)
	}

	// Render Japanese text.
	text := "日本語テスト"
	style := document.Style{
		FontFamily: "TestCJK",
		FontSize:   12,
		Color:      pdf.Black,
	}
	err = renderer.RenderText(text, document.Point{X: 50, Y: 50}, style)
	if err != nil {
		t.Fatalf("RenderText failed: %v", err)
	}

	// The page content should contain hex-encoded glyph IDs, not "(???)" literal.
	content := string(renderer.pageContent)

	if strings.Contains(content, "(") && strings.Contains(content, ") Tj") {
		t.Error("TrueType font text should use hex encoding <XX> Tj, not literal string () Tj")
	}
	if !strings.Contains(content, "> Tj") {
		t.Error("expected hex-encoded text with > Tj suffix")
	}

	// Verify the encoded bytes match what the font's Encode method produces.
	encoded := ttf.Encode(text)
	hexStr := hex.EncodeToString(encoded)
	if !strings.Contains(content, "<"+hexStr+"> Tj") {
		t.Errorf("content does not contain expected hex string <%s> Tj", hexStr)
	}
}

// TestCJKFullPDFGeneration verifies that a complete PDF with CJK text
// can be generated without errors and contains the Type0 font structure.
func TestCJKFullPDFGeneration(t *testing.T) {
	fontPath := os.Getenv("GPDF_TEST_CJK_FONT")
	if fontPath == "" {
		t.Skip("set GPDF_TEST_CJK_FONT to a CJK TTF path")
	}

	fontData, err := os.ReadFile(fontPath)
	if err != nil {
		t.Fatalf("failed to read font file: %v", err)
	}

	ttf, err := font.ParseTrueType(fontData)
	if err != nil {
		t.Fatalf("failed to parse font: %v", err)
	}

	var buf bytes.Buffer
	pw := pdf.NewWriter(&buf)
	renderer := NewPDFRenderer(pw)
	renderer.RegisterTTFont("TestCJK", ttf, fontData)

	// Build a simple page with CJK text.
	textNode := &document.Text{
		Content: "こんにちは世界",
		TextStyle: document.Style{
			FontFamily: "TestCJK",
			FontSize:   14,
			Color:      pdf.Black,
		},
	}

	pages := []layout.PageLayout{
		{
			Size: document.Size{Width: 595, Height: 842},
			Children: []layout.PlacedNode{
				{
					Node:     textNode,
					Position: document.Point{X: 72, Y: 72},
					Size:     document.Size{Width: 200, Height: 20},
				},
			},
		},
	}

	err = renderer.RenderDocument(pages, document.DocumentMetadata{Title: "CJK Test"})
	if err != nil {
		t.Fatalf("RenderDocument failed: %v", err)
	}

	// Verify the PDF output contains Type0 font markers.
	pdfBytes := buf.String()

	if !strings.Contains(pdfBytes, "/Subtype /Type0") {
		t.Error("PDF should contain Type0 font subtype")
	}
	if !strings.Contains(pdfBytes, "/Encoding /Identity-H") {
		t.Error("PDF should contain Identity-H encoding")
	}
	if !strings.Contains(pdfBytes, "/Subtype /CIDFontType2") {
		t.Error("PDF should contain CIDFontType2 descendant font")
	}
	if !strings.Contains(pdfBytes, "/ToUnicode") {
		t.Error("PDF should contain ToUnicode reference")
	}
}

// TestStandardFontNotAffected verifies that standard (non-TrueType) fonts
// still use the WinAnsi literal string encoding path.
func TestStandardFontNotAffected(t *testing.T) {
	var buf bytes.Buffer
	pw := pdf.NewWriter(&buf)
	renderer := NewPDFRenderer(pw)

	if err := renderer.BeginPage(document.Size{Width: 595, Height: 842}); err != nil {
		t.Fatalf("BeginPage failed: %v", err)
	}

	style := document.Style{
		FontFamily: "Helvetica",
		FontSize:   12,
		Color:      pdf.Black,
	}
	err := renderer.RenderText("Hello World", document.Point{X: 50, Y: 50}, style)
	if err != nil {
		t.Fatalf("RenderText failed: %v", err)
	}

	content := string(renderer.pageContent)
	if !strings.Contains(content, "(Hello World) Tj") {
		t.Error("standard font should use literal string encoding")
	}
}
