package render

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"strings"
	"testing"

	"github.com/gpdf-dev/gpdf/document"
	"github.com/gpdf-dev/gpdf/document/layout"
	"github.com/gpdf-dev/gpdf/pdf"
)

// ---------------------------------------------------------------------------
// renderer.go type tests
// ---------------------------------------------------------------------------

func TestPDFRendererImplementsRenderer(t *testing.T) {
	var _ Renderer = &PDFRenderer{}
}

func TestRectStyleFields(t *testing.T) {
	fill := pdf.Red
	stroke := pdf.Blue
	rs := RectStyle{
		FillColor:   &fill,
		StrokeColor: &stroke,
		StrokeWidth: 2.5,
	}
	if rs.FillColor == nil || *rs.FillColor != fill {
		t.Error("FillColor mismatch")
	}
	if rs.StrokeColor == nil || *rs.StrokeColor != stroke {
		t.Error("StrokeColor mismatch")
	}
	if rs.StrokeWidth != 2.5 {
		t.Errorf("StrokeWidth = %v, want 2.5", rs.StrokeWidth)
	}
}

func TestRectStyleNilColors(t *testing.T) {
	rs := RectStyle{}
	if rs.FillColor != nil {
		t.Error("FillColor should be nil by default")
	}
	if rs.StrokeColor != nil {
		t.Error("StrokeColor should be nil by default")
	}
}

// ---------------------------------------------------------------------------
// pdftarget.go tests
// ---------------------------------------------------------------------------

func newTestRenderer(t *testing.T) (*PDFRenderer, *bytes.Buffer) {
	t.Helper()
	var buf bytes.Buffer
	w := pdf.NewWriter(&buf)
	r := NewPDFRenderer(w)
	return r, &buf
}

func TestNewPDFRenderer(t *testing.T) {
	r, _ := newTestRenderer(t)
	if r == nil {
		t.Fatal("NewPDFRenderer returned nil")
	}
	if r.fontMap == nil {
		t.Error("fontMap should be initialized")
	}
	if r.fontRefs == nil {
		t.Error("fontRefs should be initialized")
	}
	if r.imageMap == nil {
		t.Error("imageMap should be initialized")
	}
	if r.imageRefs == nil {
		t.Error("imageRefs should be initialized")
	}
}

func TestBeginDocument(t *testing.T) {
	r, _ := newTestRenderer(t)
	info := document.DocumentMetadata{
		Title:  "Test",
		Author: "Author",
	}
	err := r.BeginDocument(info)
	if err != nil {
		t.Fatalf("BeginDocument error: %v", err)
	}
}

func TestBeginAndEndPage(t *testing.T) {
	r, buf := newTestRenderer(t)
	err := r.BeginDocument(document.DocumentMetadata{})
	if err != nil {
		t.Fatal(err)
	}
	err = r.BeginPage(document.Size{Width: 595, Height: 842})
	if err != nil {
		t.Fatal(err)
	}
	if r.pageWidth != 595 {
		t.Errorf("pageWidth = %v, want 595", r.pageWidth)
	}
	if r.pageHeight != 842 {
		t.Errorf("pageHeight = %v, want 842", r.pageHeight)
	}
	err = r.EndPage()
	if err != nil {
		t.Fatal(err)
	}
	err = r.EndDocument()
	if err != nil {
		t.Fatal(err)
	}
	output := buf.String()
	if !strings.Contains(output, "%PDF-1.7") {
		t.Error("Output should contain PDF header")
	}
	if !strings.Contains(output, "%%EOF") {
		t.Error("Output should contain EOF marker")
	}
}

func TestRenderTextEmpty(t *testing.T) {
	r, _ := newTestRenderer(t)
	_ = r.BeginPage(document.Size{Width: 595, Height: 842})
	err := r.RenderText("", document.Point{X: 10, Y: 10}, document.DefaultStyle())
	if err != nil {
		t.Fatalf("RenderText empty: %v", err)
	}
	// Empty text should not produce content.
	if len(r.pageContent) != 0 {
		t.Errorf("Empty text should produce no content, got %d bytes", len(r.pageContent))
	}
}

func TestRenderTextContent(t *testing.T) {
	r, _ := newTestRenderer(t)
	_ = r.BeginPage(document.Size{Width: 595, Height: 842})
	style := document.DefaultStyle()
	style.FontSize = 14
	err := r.RenderText("Hello World", document.Point{X: 72, Y: 72}, style)
	if err != nil {
		t.Fatalf("RenderText error: %v", err)
	}
	content := string(r.pageContent)
	if !strings.Contains(content, "BT") {
		t.Error("Content should contain BT (begin text)")
	}
	if !strings.Contains(content, "ET") {
		t.Error("Content should contain ET (end text)")
	}
	if !strings.Contains(content, "Tf") {
		t.Error("Content should contain Tf (set font)")
	}
	if !strings.Contains(content, "Td") {
		t.Error("Content should contain Td (text position)")
	}
	if !strings.Contains(content, "(Hello World) Tj") {
		t.Error("Content should contain the text string with Tj operator")
	}
}

func TestRenderTextDefaultFont(t *testing.T) {
	r, _ := newTestRenderer(t)
	_ = r.BeginPage(document.Size{Width: 595, Height: 842})
	style := document.DefaultStyle()
	style.FontFamily = "" // Should default to "Helvetica"
	err := r.RenderText("test", document.Point{X: 10, Y: 10}, style)
	if err != nil {
		t.Fatalf("RenderText error: %v", err)
	}
	if _, ok := r.fontMap["Helvetica"]; !ok {
		t.Error("Default font 'Helvetica' should be registered")
	}
}

func TestRenderTextCustomFont(t *testing.T) {
	r, _ := newTestRenderer(t)
	_ = r.BeginPage(document.Size{Width: 595, Height: 842})
	style := document.DefaultStyle()
	style.FontFamily = "Times-Roman"
	err := r.RenderText("test", document.Point{X: 10, Y: 10}, style)
	if err != nil {
		t.Fatalf("RenderText error: %v", err)
	}
	if _, ok := r.fontMap["Times-Roman"]; !ok {
		t.Error("Font 'Times-Roman' should be registered")
	}
}

func TestRenderTextDefaultFontSize(t *testing.T) {
	r, _ := newTestRenderer(t)
	_ = r.BeginPage(document.Size{Width: 595, Height: 842})
	style := document.DefaultStyle()
	style.FontSize = 0 // Should default to 12
	err := r.RenderText("test", document.Point{X: 10, Y: 10}, style)
	if err != nil {
		t.Fatalf("RenderText error: %v", err)
	}
	content := string(r.pageContent)
	if !strings.Contains(content, "12 Tf") {
		t.Error("Should use default font size 12")
	}
}

func TestRenderTextYConversion(t *testing.T) {
	r, _ := newTestRenderer(t)
	_ = r.BeginPage(document.Size{Width: 595, Height: 842})
	style := document.DefaultStyle()
	style.FontSize = 12
	err := r.RenderText("test", document.Point{X: 0, Y: 100}, style)
	if err != nil {
		t.Fatal(err)
	}
	content := string(r.pageContent)
	// pdfY = 842 - 100 - 12 = 730
	if !strings.Contains(content, "0 730 Td") {
		t.Errorf("Expected Y conversion to PDF coords, got: %s", content)
	}
}

func TestRenderTextSpecialChars(t *testing.T) {
	r, _ := newTestRenderer(t)
	_ = r.BeginPage(document.Size{Width: 595, Height: 842})
	err := r.RenderText("hello (world) \\test\n", document.Point{X: 10, Y: 10}, document.DefaultStyle())
	if err != nil {
		t.Fatal(err)
	}
	content := string(r.pageContent)
	if !strings.Contains(content, `\(`) {
		t.Error("Should escape parentheses")
	}
	if !strings.Contains(content, `\\`) {
		t.Error("Should escape backslash")
	}
	if !strings.Contains(content, `\n`) {
		t.Error("Should escape newline")
	}
}

func TestRenderTextWordSpacing(t *testing.T) {
	r, _ := newTestRenderer(t)
	_ = r.BeginPage(document.Size{Width: 595, Height: 842})
	style := document.DefaultStyle()
	style.WordSpacing = 5.5
	err := r.RenderText("hello world", document.Point{X: 10, Y: 10}, style)
	if err != nil {
		t.Fatal(err)
	}
	content := string(r.pageContent)
	if !strings.Contains(content, "5.5 Tw") {
		t.Error("Should contain Tw operator with word spacing value")
	}
	if !strings.Contains(content, "0 Tw") {
		t.Error("Should reset Tw to 0 after text")
	}
}

func TestRenderTextNoWordSpacing(t *testing.T) {
	r, _ := newTestRenderer(t)
	_ = r.BeginPage(document.Size{Width: 595, Height: 842})
	style := document.DefaultStyle()
	err := r.RenderText("hello world", document.Point{X: 10, Y: 10}, style)
	if err != nil {
		t.Fatal(err)
	}
	content := string(r.pageContent)
	if strings.Contains(content, "Tw") {
		t.Error("Should not contain Tw operator when WordSpacing is 0")
	}
}

func TestRenderRectFillOnly(t *testing.T) {
	r, _ := newTestRenderer(t)
	_ = r.BeginPage(document.Size{Width: 595, Height: 842})
	fill := pdf.Red
	err := r.RenderRect(document.Rectangle{X: 10, Y: 20, Width: 100, Height: 50}, RectStyle{
		FillColor: &fill,
	})
	if err != nil {
		t.Fatal(err)
	}
	content := string(r.pageContent)
	if !strings.Contains(content, "q\n") {
		t.Error("Should save graphics state")
	}
	if !strings.Contains(content, "Q\n") {
		t.Error("Should restore graphics state")
	}
	if !strings.Contains(content, "re\n") {
		t.Error("Should contain rectangle operator")
	}
	if !strings.Contains(content, "f\n") {
		t.Error("Should contain fill operator 'f'")
	}
}

func TestRenderRectStrokeOnly(t *testing.T) {
	r, _ := newTestRenderer(t)
	_ = r.BeginPage(document.Size{Width: 595, Height: 842})
	stroke := pdf.Blue
	err := r.RenderRect(document.Rectangle{X: 10, Y: 20, Width: 100, Height: 50}, RectStyle{
		StrokeColor: &stroke,
		StrokeWidth: 2,
	})
	if err != nil {
		t.Fatal(err)
	}
	content := string(r.pageContent)
	if !strings.Contains(content, "S\n") {
		t.Error("Should contain stroke operator 'S'")
	}
	if !strings.Contains(content, "2 w\n") {
		t.Error("Should set stroke width")
	}
}

func TestRenderRectFillAndStroke(t *testing.T) {
	r, _ := newTestRenderer(t)
	_ = r.BeginPage(document.Size{Width: 595, Height: 842})
	fill := pdf.Red
	stroke := pdf.Blue
	err := r.RenderRect(document.Rectangle{X: 10, Y: 20, Width: 100, Height: 50}, RectStyle{
		FillColor:   &fill,
		StrokeColor: &stroke,
		StrokeWidth: 1,
	})
	if err != nil {
		t.Fatal(err)
	}
	content := string(r.pageContent)
	if !strings.Contains(content, "B\n") {
		t.Error("Should contain fill-and-stroke operator 'B'")
	}
}

func TestRenderRectNoFillNoStroke(t *testing.T) {
	r, _ := newTestRenderer(t)
	_ = r.BeginPage(document.Size{Width: 595, Height: 842})
	err := r.RenderRect(document.Rectangle{X: 10, Y: 20, Width: 100, Height: 50}, RectStyle{})
	if err != nil {
		t.Fatal(err)
	}
	content := string(r.pageContent)
	if !strings.Contains(content, "n\n") {
		t.Error("Should contain no-op path operator 'n'")
	}
}

func TestRenderRectYConversion(t *testing.T) {
	r, _ := newTestRenderer(t)
	_ = r.BeginPage(document.Size{Width: 595, Height: 842})
	fill := pdf.Red
	err := r.RenderRect(document.Rectangle{X: 0, Y: 100, Width: 200, Height: 50}, RectStyle{
		FillColor: &fill,
	})
	if err != nil {
		t.Fatal(err)
	}
	content := string(r.pageContent)
	// pdfY = 842 - 100 - 50 = 692
	if !strings.Contains(content, "0 692 200 50 re") {
		t.Errorf("Expected Y conversion in rect, got: %s", content)
	}
}

func TestRenderImage(t *testing.T) {
	r, _ := newTestRenderer(t)
	_ = r.BeginPage(document.Size{Width: 595, Height: 842})
	src := document.ImageSource{
		Data:   []byte{0xFF, 0xD8, 0xFF, 0xE0, 0x00, 0x10},
		Format: document.ImageJPEG,
		Width:  100,
		Height: 80,
	}
	pos := document.Point{X: 50, Y: 100}
	size := document.Size{Width: 200, Height: 160}
	err := r.RenderImage(src, pos, size)
	if err != nil {
		t.Fatal(err)
	}
	content := string(r.pageContent)
	if !strings.Contains(content, "q\n") {
		t.Error("Should save graphics state")
	}
	if !strings.Contains(content, "cm\n") {
		t.Error("Should contain concat matrix operator")
	}
	if !strings.Contains(content, "Do\n") {
		t.Error("Should contain Do operator for XObject")
	}
	if !strings.Contains(content, "Q\n") {
		t.Error("Should restore graphics state")
	}
}

func TestRenderImageDeduplication(t *testing.T) {
	r, _ := newTestRenderer(t)
	_ = r.BeginPage(document.Size{Width: 595, Height: 842})
	src := document.ImageSource{
		Data:   testPNG(t, 10, 10),
		Format: document.ImagePNG,
		Width:  10,
		Height: 10,
	}
	pos := document.Point{X: 0, Y: 0}
	size := document.Size{Width: 10, Height: 10}

	err := r.RenderImage(src, pos, size)
	if err != nil {
		t.Fatal(err)
	}
	imageCountBefore := len(r.imageMap)

	// Render the same image again.
	err = r.RenderImage(src, document.Point{X: 20, Y: 20}, size)
	if err != nil {
		t.Fatal(err)
	}
	if len(r.imageMap) != imageCountBefore {
		t.Error("Same image data should not create a new image entry")
	}
}

func TestEnsureFontDeduplication(t *testing.T) {
	r, _ := newTestRenderer(t)
	_ = r.BeginPage(document.Size{Width: 595, Height: 842})

	// Register the same font twice.
	style := document.DefaultStyle()
	style.FontFamily = "Helvetica"
	err := r.RenderText("first", document.Point{X: 0, Y: 0}, style)
	if err != nil {
		t.Fatal(err)
	}
	fontCountBefore := len(r.fontMap)

	err = r.RenderText("second", document.Point{X: 0, Y: 20}, style)
	if err != nil {
		t.Fatal(err)
	}
	if len(r.fontMap) != fontCountBefore {
		t.Error("Same font should not be registered twice")
	}
}

func TestEndDocument(t *testing.T) {
	r, buf := newTestRenderer(t)
	_ = r.BeginDocument(document.DocumentMetadata{Title: "Test"})
	_ = r.BeginPage(document.Size{Width: 595, Height: 842})
	_ = r.EndPage()
	err := r.EndDocument()
	if err != nil {
		t.Fatal(err)
	}
	output := buf.String()
	if !strings.Contains(output, "%%EOF") {
		t.Error("Should contain EOF marker")
	}
	if !strings.Contains(output, "xref") {
		t.Error("Should contain xref table")
	}
	if !strings.Contains(output, "trailer") {
		t.Error("Should contain trailer")
	}
}

func TestRenderDocumentEmpty(t *testing.T) {
	r, buf := newTestRenderer(t)
	info := document.DocumentMetadata{Title: "Empty"}
	err := r.RenderDocument(nil, info)
	if err != nil {
		t.Fatal(err)
	}
	output := buf.String()
	if !strings.Contains(output, "%%EOF") {
		t.Error("Should produce valid PDF even with no pages")
	}
}

func TestRenderDocumentWithPages(t *testing.T) {
	r, buf := newTestRenderer(t)
	pages := []layout.PageLayout{
		{
			Size: document.A4,
			Children: []layout.PlacedNode{
				{
					Node: &document.Text{
						Content: "Page 1 text",
						TextStyle: document.Style{
							FontSize:   12,
							LineHeight: 1.2,
							Color:      pdf.Black,
						},
					},
					Position: document.Point{X: 72, Y: 72},
					Size:     document.Size{Width: 451, Height: 14.4},
				},
			},
		},
	}
	info := document.DocumentMetadata{
		Title:  "Test Document",
		Author: "Test Author",
	}
	err := r.RenderDocument(pages, info)
	if err != nil {
		t.Fatal(err)
	}
	output := buf.String()
	if !strings.Contains(output, "%PDF-1.7") {
		t.Error("Should contain PDF header")
	}
	if !strings.Contains(output, "%%EOF") {
		t.Error("Should contain EOF")
	}
	if !strings.Contains(output, "Page 1 text") {
		t.Error("Should contain the rendered text")
	}
}

func TestRenderDocumentWithBackground(t *testing.T) {
	r, _ := newTestRenderer(t)
	bg := pdf.RGB(0.9, 0.9, 0.9)
	pages := []layout.PageLayout{
		{
			Size: document.A4,
			Children: []layout.PlacedNode{
				{
					Node: &document.Box{
						BoxStyle: document.BoxStyle{
							Background: &bg,
						},
					},
					Position: document.Point{X: 72, Y: 72},
					Size:     document.Size{Width: 451, Height: 100},
				},
			},
		},
	}
	info := document.DocumentMetadata{}
	err := r.RenderDocument(pages, info)
	if err != nil {
		t.Fatal(err)
	}
}

func TestRenderDocumentWithImage(t *testing.T) {
	r, buf := newTestRenderer(t)
	pages := []layout.PageLayout{
		{
			Size: document.A4,
			Children: []layout.PlacedNode{
				{
					Node: &document.Image{
						Source: document.ImageSource{
							Data:   []byte{0xFF, 0xD8, 0xFF},
							Format: document.ImageJPEG,
							Width:  100,
							Height: 80,
						},
					},
					Position: document.Point{X: 72, Y: 72},
					Size:     document.Size{Width: 200, Height: 160},
				},
			},
		},
	}
	info := document.DocumentMetadata{}
	err := r.RenderDocument(pages, info)
	if err != nil {
		t.Fatal(err)
	}
	output := buf.String()
	if !strings.Contains(output, "XObject") {
		t.Error("Should contain XObject for image")
	}
}

func TestRenderDocumentWithBorders(t *testing.T) {
	r, _ := newTestRenderer(t)
	border := document.UniformBorder(document.Pt(1), document.BorderSolid, pdf.Black)
	pages := []layout.PageLayout{
		{
			Size: document.A4,
			Children: []layout.PlacedNode{
				{
					Node: &document.Box{
						BoxStyle: document.BoxStyle{
							Border: border,
						},
					},
					Position: document.Point{X: 72, Y: 72},
					Size:     document.Size{Width: 200, Height: 100},
				},
			},
		},
	}
	info := document.DocumentMetadata{}
	err := r.RenderDocument(pages, info)
	if err != nil {
		t.Fatal(err)
	}
}

func TestRenderDocumentNilNode(t *testing.T) {
	r, _ := newTestRenderer(t)
	pages := []layout.PageLayout{
		{
			Size: document.A4,
			Children: []layout.PlacedNode{
				{Node: nil}, // nil node should be skipped
			},
		},
	}
	info := document.DocumentMetadata{}
	err := r.RenderDocument(pages, info)
	if err != nil {
		t.Fatal(err)
	}
}

func TestRenderDocumentNestedChildren(t *testing.T) {
	r, buf := newTestRenderer(t)
	pages := []layout.PageLayout{
		{
			Size: document.A4,
			Children: []layout.PlacedNode{
				{
					Node:     &document.Box{},
					Position: document.Point{X: 72, Y: 72},
					Size:     document.Size{Width: 451, Height: 100},
					Children: []layout.PlacedNode{
						{
							Node: &document.Text{
								Content:   "Nested text",
								TextStyle: document.DefaultStyle(),
							},
							Position: document.Point{X: 72, Y: 72},
							Size:     document.Size{Width: 100, Height: 14},
						},
					},
				},
			},
		},
	}
	info := document.DocumentMetadata{}
	err := r.RenderDocument(pages, info)
	if err != nil {
		t.Fatal(err)
	}
	output := buf.String()
	if !strings.Contains(output, "Nested text") {
		t.Error("Should render nested text")
	}
}

func TestEscapeStringPDF(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"hello", "hello"},
		{"hello(world)", `hello\(world\)`},
		{`back\slash`, `back\\slash`},
		{"line\nbreak", `line\nbreak`},
		{"carriage\rreturn", `carriage\rreturn`},
		{"", ""},
		{"(())", `\(\(\)\)`},
	}
	for _, tt := range tests {
		got := escapeStringPDF(tt.input)
		if got != tt.want {
			t.Errorf("escapeStringPDF(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestEscapeStringPDF_WinAnsi(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  byte // expected single-byte output for a one-character input
	}{
		{"bullet", "\u2022", 0x95},
		{"endash", "\u2013", 0x96},
		{"emdash", "\u2014", 0x97},
		{"euro", "\u20AC", 0x80},
		{"ellipsis", "\u2026", 0x85},
		{"trademark", "\u2122", 0x99},
		{"leftdoublequote", "\u201C", 0x93},
		{"rightdoublequote", "\u201D", 0x94},
		{"leftsinglequote", "\u2018", 0x91},
		{"rightsinglequote", "\u2019", 0x92},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := escapeStringPDF(tt.input)
			if len(got) != 1 || got[0] != tt.want {
				t.Errorf("escapeStringPDF(%q) = %x, want single byte %x", tt.input, []byte(got), tt.want)
			}
		})
	}
}

func TestEscapeStringPDF_Latin1(t *testing.T) {
	// Latin-1 supplement characters (0xA0–0xFF) should map directly.
	tests := []struct {
		name  string
		input rune
		want  byte
	}{
		{"copyright", '©', 0xA9},
		{"registered", '®', 0xAE},
		{"degree", '°', 0xB0},
		{"umlaut_u", 'ü', 0xFC},
		{"ntilde", 'ñ', 0xF1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := escapeStringPDF(string(tt.input))
			if len(got) != 1 || got[0] != tt.want {
				t.Errorf("escapeStringPDF(%q) = %x, want single byte %x", string(tt.input), []byte(got), tt.want)
			}
		})
	}
}

func TestEscapeStringPDF_UnmappedRune(t *testing.T) {
	// CJK characters are outside WinAnsiEncoding; should be replaced with '?'.
	got := escapeStringPDF("日本語")
	if got != "???" {
		t.Errorf("escapeStringPDF(CJK) = %q, want %q", got, "???")
	}
}

func TestRuneToWinAnsi(t *testing.T) {
	// ASCII
	if b := runeToWinAnsi('A'); b != 'A' {
		t.Errorf("runeToWinAnsi('A') = %x, want %x", b, 'A')
	}
	// Bullet
	if b := runeToWinAnsi('\u2022'); b != 0x95 {
		t.Errorf("runeToWinAnsi(bullet) = %x, want 0x95", b)
	}
	// Unmapped
	if b := runeToWinAnsi('日'); b != '?' {
		t.Errorf("runeToWinAnsi(CJK) = %x, want '?'", b)
	}
}

func TestImageKey(t *testing.T) {
	data1 := []byte{0x01, 0x02, 0x03}
	data2 := []byte{0x04, 0x05, 0x06}
	key1 := imageKey(data1)
	key2 := imageKey(data2)
	if key1 == "" {
		t.Error("imageKey should not be empty")
	}
	if key1 == key2 {
		t.Error("Different data should produce different keys")
	}
	// Same data should produce the same key.
	key1Again := imageKey(data1)
	if key1 != key1Again {
		t.Error("Same data should produce the same key")
	}
}

func TestDecodePNGToRaw_OpaqueImage(t *testing.T) {
	// Create a fully opaque 2x2 PNG.
	img := image.NewRGBA(image.Rect(0, 0, 2, 2))
	img.Set(0, 0, color.RGBA{R: 255, G: 0, B: 0, A: 255})
	img.Set(1, 0, color.RGBA{R: 0, G: 255, B: 0, A: 255})
	img.Set(0, 1, color.RGBA{R: 0, G: 0, B: 255, A: 255})
	img.Set(1, 1, color.RGBA{R: 255, G: 255, B: 255, A: 255})

	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatal(err)
	}

	rgb, alpha, w, h, err := decodePNGToRaw(buf.Bytes())
	if err != nil {
		t.Fatalf("decodePNGToRaw error: %v", err)
	}
	if w != 2 || h != 2 {
		t.Errorf("dimensions = %dx%d, want 2x2", w, h)
	}
	if len(rgb) != 2*2*3 {
		t.Errorf("RGB data length = %d, want %d", len(rgb), 2*2*3)
	}
	if alpha != nil {
		t.Error("fully opaque image should return nil alpha")
	}
}

func TestDecodePNGToRaw_TransparentImage(t *testing.T) {
	// Create a 2x2 PNG with alpha.
	img := image.NewRGBA(image.Rect(0, 0, 2, 2))
	img.Set(0, 0, color.RGBA{R: 255, G: 0, B: 0, A: 128})
	img.Set(1, 0, color.RGBA{R: 0, G: 255, B: 0, A: 255})
	img.Set(0, 1, color.RGBA{R: 0, G: 0, B: 255, A: 0})
	img.Set(1, 1, color.RGBA{R: 255, G: 255, B: 255, A: 200})

	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatal(err)
	}

	rgb, alpha, w, h, err := decodePNGToRaw(buf.Bytes())
	if err != nil {
		t.Fatalf("decodePNGToRaw error: %v", err)
	}
	if w != 2 || h != 2 {
		t.Errorf("dimensions = %dx%d, want 2x2", w, h)
	}
	if len(rgb) != 2*2*3 {
		t.Errorf("RGB data length = %d, want %d", len(rgb), 2*2*3)
	}
	if alpha == nil {
		t.Fatal("expected non-nil alpha for transparent image")
	}
	if len(alpha) != 2*2 {
		t.Errorf("alpha data length = %d, want %d", len(alpha), 2*2)
	}
	// Check alpha values.
	if alpha[0] != 128 {
		t.Errorf("alpha[0] = %d, want 128", alpha[0])
	}
	if alpha[2] != 0 {
		t.Errorf("alpha[2] = %d, want 0", alpha[2])
	}
}

func TestEnsureImage_PNGWithAlpha(t *testing.T) {
	r, buf := newTestRenderer(t)

	// Create a small transparent PNG.
	img := image.NewRGBA(image.Rect(0, 0, 4, 4))
	for y := 0; y < 4; y++ {
		for x := 0; x < 4; x++ {
			img.Set(x, y, color.RGBA{R: 255, G: 0, B: 0, A: 128})
		}
	}
	var pngBuf bytes.Buffer
	if err := png.Encode(&pngBuf, img); err != nil {
		t.Fatal(err)
	}

	if err := r.BeginDocument(document.DocumentMetadata{}); err != nil {
		t.Fatal(err)
	}
	if err := r.BeginPage(document.A4); err != nil {
		t.Fatal(err)
	}

	src := document.ImageSource{
		Data:   pngBuf.Bytes(),
		Format: document.ImagePNG,
		Width:  4,
		Height: 4,
	}
	resName, err := r.ensureImage("alpha-test", src)
	if err != nil {
		t.Fatalf("ensureImage error: %v", err)
	}
	if resName == "" {
		t.Error("expected non-empty resource name")
	}

	if err := r.EndPage(); err != nil {
		t.Fatal(err)
	}
	if err := r.EndDocument(); err != nil {
		t.Fatal(err)
	}

	output := buf.String()
	// The output should contain an SMask reference.
	if !strings.Contains(output, "/SMask") {
		t.Error("expected /SMask in PDF output for transparent PNG")
	}
	// The output should contain DeviceGray (for the SMask image).
	if !strings.Contains(output, "/DeviceGray") {
		t.Error("expected /DeviceGray for SMask image in PDF output")
	}
}

func TestEnsureImage_OpaquePN_NoSMask(t *testing.T) {
	r, buf := newTestRenderer(t)

	// Create a fully opaque PNG.
	img := image.NewRGBA(image.Rect(0, 0, 4, 4))
	for y := 0; y < 4; y++ {
		for x := 0; x < 4; x++ {
			img.Set(x, y, color.RGBA{R: 255, G: 0, B: 0, A: 255})
		}
	}
	var pngBuf bytes.Buffer
	if err := png.Encode(&pngBuf, img); err != nil {
		t.Fatal(err)
	}

	if err := r.BeginDocument(document.DocumentMetadata{}); err != nil {
		t.Fatal(err)
	}
	if err := r.BeginPage(document.A4); err != nil {
		t.Fatal(err)
	}

	src := document.ImageSource{
		Data:   pngBuf.Bytes(),
		Format: document.ImagePNG,
		Width:  4,
		Height: 4,
	}
	_, err := r.ensureImage("opaque-test", src)
	if err != nil {
		t.Fatalf("ensureImage error: %v", err)
	}

	if err := r.EndPage(); err != nil {
		t.Fatal(err)
	}
	if err := r.EndDocument(); err != nil {
		t.Fatal(err)
	}

	output := buf.String()
	// Fully opaque PNG should NOT have SMask.
	if strings.Contains(output, "/SMask") {
		t.Error("opaque PNG should not have /SMask")
	}
}

// renderPipelineTextPage renders a page with text and a filled rectangle.
func renderPipelineTextPage(t *testing.T, r *PDFRenderer) {
	t.Helper()
	if err := r.BeginPage(document.A4); err != nil {
		t.Fatal(err)
	}
	style := document.DefaultStyle()
	style.FontSize = 24
	if err := r.RenderText("Hello World", document.Point{X: 72, Y: 72}, style); err != nil {
		t.Fatal(err)
	}
	fill := pdf.RGB(0.9, 0.9, 0.9)
	if err := r.RenderRect(document.Rectangle{X: 72, Y: 120, Width: 451, Height: 100}, RectStyle{
		FillColor: &fill,
	}); err != nil {
		t.Fatal(err)
	}
	if err := r.EndPage(); err != nil {
		t.Fatal(err)
	}
}

// renderPipelineImagePage renders a page with an image.
func renderPipelineImagePage(t *testing.T, r *PDFRenderer) {
	t.Helper()
	if err := r.BeginPage(document.A4); err != nil {
		t.Fatal(err)
	}
	imgData := make([]byte, 300)
	for i := range imgData {
		imgData[i] = byte(i % 256)
	}
	src := document.ImageSource{
		Data:   imgData,
		Format: document.ImageJPEG,
		Width:  100,
		Height: 100,
	}
	if err := r.RenderImage(src, document.Point{X: 72, Y: 72}, document.Size{Width: 200, Height: 200}); err != nil {
		t.Fatal(err)
	}
	if err := r.EndPage(); err != nil {
		t.Fatal(err)
	}
}

// assertPipelineOutput verifies the final PDF output contains expected elements.
func assertPipelineOutput(t *testing.T, output string) {
	t.Helper()
	for _, want := range []string{"%PDF-1.7", "%%EOF", "Hello World", "xref", "trailer", "/Font"} {
		if !strings.Contains(output, want) {
			t.Errorf("Missing expected element in output: %s", want)
		}
	}
}

func TestFullRenderingPipeline(t *testing.T) {
	var buf bytes.Buffer
	w := pdf.NewWriter(&buf)
	r := NewPDFRenderer(w)

	info := document.DocumentMetadata{
		Title:    "Full Pipeline Test",
		Author:   "Test Suite",
		Subject:  "Testing",
		Creator:  "gpdf",
		Producer: "gpdf test",
	}
	if err := r.BeginDocument(info); err != nil {
		t.Fatal(err)
	}

	renderPipelineTextPage(t, r)
	renderPipelineImagePage(t, r)

	if err := r.EndDocument(); err != nil {
		t.Fatal(err)
	}

	assertPipelineOutput(t, buf.String())
}

func TestRenderDocumentMultiplePages(t *testing.T) {
	r, buf := newTestRenderer(t)
	pages := []layout.PageLayout{
		{
			Size: document.A4,
			Children: []layout.PlacedNode{
				{
					Node: &document.Text{
						Content:   "Page 1",
						TextStyle: document.DefaultStyle(),
					},
					Position: document.Point{X: 72, Y: 72},
					Size:     document.Size{Width: 100, Height: 14},
				},
			},
		},
		{
			Size: document.Letter,
			Children: []layout.PlacedNode{
				{
					Node: &document.Text{
						Content:   "Page 2",
						TextStyle: document.DefaultStyle(),
					},
					Position: document.Point{X: 72, Y: 72},
					Size:     document.Size{Width: 100, Height: 14},
				},
			},
		},
	}
	info := document.DocumentMetadata{Title: "Multi-page"}
	err := r.RenderDocument(pages, info)
	if err != nil {
		t.Fatal(err)
	}
	output := buf.String()
	if !strings.Contains(output, "Page 1") {
		t.Error("Missing page 1 text")
	}
	if !strings.Contains(output, "Page 2") {
		t.Error("Missing page 2 text")
	}
}

func TestRenderRectStrokeWithZeroWidth(t *testing.T) {
	r, _ := newTestRenderer(t)
	_ = r.BeginPage(document.Size{Width: 595, Height: 842})
	stroke := pdf.Black
	err := r.RenderRect(document.Rectangle{X: 10, Y: 20, Width: 100, Height: 50}, RectStyle{
		StrokeColor: &stroke,
		StrokeWidth: 0, // Zero width should not write a "w" command
	})
	if err != nil {
		t.Fatal(err)
	}
	content := string(r.pageContent)
	if strings.Contains(content, " w\n") {
		t.Error("Zero stroke width should not produce 'w' operator")
	}
}

func TestRenderBordersSelectiveSides(t *testing.T) {
	r, _ := newTestRenderer(t)
	_ = r.BeginPage(document.Size{Width: 595, Height: 842})

	pn := layout.PlacedNode{
		Node: &document.Box{
			BoxStyle: document.BoxStyle{
				Border: document.BorderEdges{
					Top:    document.BorderSide{Width: document.Pt(1), Style: document.BorderSolid, Color: pdf.Black},
					Right:  document.BorderSide{Width: document.Pt(0), Style: document.BorderNone},
					Bottom: document.BorderSide{Width: document.Pt(2), Style: document.BorderDashed, Color: pdf.Red},
					Left:   document.BorderSide{Width: document.Pt(0), Style: document.BorderNone},
				},
			},
		},
		Position: document.Point{X: 50, Y: 50},
		Size:     document.Size{Width: 200, Height: 100},
	}
	// Render the placed node to test border rendering.
	err := r.renderPlacedNode(pn, 0, 0)
	if err != nil {
		t.Fatal(err)
	}
	// Should have rendered top and bottom borders but not right and left.
	content := string(r.pageContent)
	if !strings.Contains(content, "re") {
		t.Error("Should contain rectangle operations for borders")
	}
}

func TestEndPageWithFontsAndImages(t *testing.T) {
	r, buf := newTestRenderer(t)
	_ = r.BeginDocument(document.DocumentMetadata{})
	_ = r.BeginPage(document.Size{Width: 595, Height: 842})

	// Add text (registers font).
	_ = r.RenderText("test", document.Point{X: 10, Y: 10}, document.DefaultStyle())

	// Add image (registers image).
	src := document.ImageSource{
		Data:   []byte{0x01, 0x02},
		Format: document.ImageJPEG,
		Width:  10,
		Height: 10,
	}
	_ = r.RenderImage(src, document.Point{X: 50, Y: 50}, document.Size{Width: 20, Height: 20})

	_ = r.EndPage()
	_ = r.EndDocument()

	output := buf.String()
	if !strings.Contains(output, "/Font") {
		t.Error("Page should reference fonts")
	}
	if !strings.Contains(output, "/XObject") {
		t.Error("Page should reference XObjects for images")
	}
}

func TestBeginPageResetsContent(t *testing.T) {
	r, _ := newTestRenderer(t)
	_ = r.BeginPage(document.Size{Width: 595, Height: 842})
	_ = r.RenderText("page 1 text", document.Point{X: 10, Y: 10}, document.DefaultStyle())

	// Begin a new page should reset content.
	_ = r.BeginPage(document.Size{Width: 612, Height: 792})
	if len(r.pageContent) != 0 {
		t.Error("BeginPage should reset pageContent")
	}
}

// testPNG creates a small valid PNG image for testing.
func testPNG(t *testing.T, w, h int) []byte {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := range h {
		for x := range w {
			img.Set(x, y, color.RGBA{R: 255, A: 255})
		}
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatalf("failed to create test PNG: %v", err)
	}
	return buf.Bytes()
}

// ---------------------------------------------------------------------------
// RenderPath / RenderLine tests (WP3)
// ---------------------------------------------------------------------------

func TestRenderPathStrokeOnly(t *testing.T) {
	r, _ := newTestRenderer(t)
	_ = r.BeginPage(document.Size{Width: 595, Height: 842})
	stroke := pdf.Black
	path := document.Path{
		Segments: []document.PathSegment{
			{Op: document.PathMoveTo, Points: []document.Point{{X: 10, Y: 20}}},
			{Op: document.PathLineTo, Points: []document.Point{{X: 100, Y: 20}}},
		},
	}
	err := r.RenderPath(path, PathStyle{StrokeColor: &stroke, StrokeWidth: 2})
	if err != nil {
		t.Fatal(err)
	}
	content := string(r.pageContent)
	if !strings.Contains(content, "m\n") {
		t.Error("Should contain moveto operator")
	}
	if !strings.Contains(content, "l\n") {
		t.Error("Should contain lineto operator")
	}
	if !strings.Contains(content, "S\n") {
		t.Error("Should contain stroke operator 'S'")
	}
}

func TestRenderPathFillOnly(t *testing.T) {
	r, _ := newTestRenderer(t)
	_ = r.BeginPage(document.Size{Width: 595, Height: 842})
	fill := pdf.Red
	path := document.Path{
		Segments: []document.PathSegment{
			{Op: document.PathMoveTo, Points: []document.Point{{X: 0, Y: 0}}},
			{Op: document.PathLineTo, Points: []document.Point{{X: 100, Y: 0}}},
			{Op: document.PathLineTo, Points: []document.Point{{X: 50, Y: 80}}},
			{Op: document.PathClose},
		},
	}
	err := r.RenderPath(path, PathStyle{FillColor: &fill})
	if err != nil {
		t.Fatal(err)
	}
	content := string(r.pageContent)
	if !strings.Contains(content, "h\n") {
		t.Error("Should contain close operator 'h'")
	}
	if !strings.Contains(content, "f\n") {
		t.Error("Should contain fill operator 'f'")
	}
}

func TestRenderPathFillAndStroke(t *testing.T) {
	r, _ := newTestRenderer(t)
	_ = r.BeginPage(document.Size{Width: 595, Height: 842})
	fill := pdf.Red
	stroke := pdf.Blue
	path := document.Path{
		Segments: []document.PathSegment{
			{Op: document.PathMoveTo, Points: []document.Point{{X: 0, Y: 0}}},
			{Op: document.PathLineTo, Points: []document.Point{{X: 100, Y: 0}}},
			{Op: document.PathClose},
		},
	}
	err := r.RenderPath(path, PathStyle{FillColor: &fill, StrokeColor: &stroke, StrokeWidth: 1})
	if err != nil {
		t.Fatal(err)
	}
	content := string(r.pageContent)
	if !strings.Contains(content, "B\n") {
		t.Error("Should contain fill-and-stroke operator 'B'")
	}
}

func TestRenderPathCurveTo(t *testing.T) {
	r, _ := newTestRenderer(t)
	_ = r.BeginPage(document.Size{Width: 595, Height: 842})
	stroke := pdf.Black
	path := document.Path{
		Segments: []document.PathSegment{
			{Op: document.PathMoveTo, Points: []document.Point{{X: 10, Y: 10}}},
			{Op: document.PathCurveTo, Points: []document.Point{
				{X: 30, Y: 50}, {X: 70, Y: 50}, {X: 90, Y: 10},
			}},
		},
	}
	err := r.RenderPath(path, PathStyle{StrokeColor: &stroke, StrokeWidth: 1})
	if err != nil {
		t.Fatal(err)
	}
	content := string(r.pageContent)
	if !strings.Contains(content, "c\n") {
		t.Error("Should contain curve operator 'c'")
	}
}

func TestRenderPathDashPattern(t *testing.T) {
	r, _ := newTestRenderer(t)
	_ = r.BeginPage(document.Size{Width: 595, Height: 842})
	stroke := pdf.Black
	path := document.Path{
		Segments: []document.PathSegment{
			{Op: document.PathMoveTo, Points: []document.Point{{X: 0, Y: 0}}},
			{Op: document.PathLineTo, Points: []document.Point{{X: 100, Y: 0}}},
		},
	}
	err := r.RenderPath(path, PathStyle{
		StrokeColor: &stroke,
		StrokeWidth: 1,
		DashPattern: []float64{3, 2},
		DashPhase:   0,
	})
	if err != nil {
		t.Fatal(err)
	}
	content := string(r.pageContent)
	if !strings.Contains(content, "[3 2] 0 d") {
		t.Errorf("Should contain dash pattern, got: %s", content)
	}
}

func TestRenderPathYConversion(t *testing.T) {
	r, _ := newTestRenderer(t)
	_ = r.BeginPage(document.Size{Width: 595, Height: 842})
	stroke := pdf.Black
	path := document.Path{
		Segments: []document.PathSegment{
			{Op: document.PathMoveTo, Points: []document.Point{{X: 10, Y: 100}}},
		},
	}
	err := r.RenderPath(path, PathStyle{StrokeColor: &stroke})
	if err != nil {
		t.Fatal(err)
	}
	content := string(r.pageContent)
	// pdfY = 842 - 100 = 742
	if !strings.Contains(content, "10 742 m") {
		t.Errorf("Expected Y conversion in path moveto, got: %s", content)
	}
}

// ---------------------------------------------------------------------------
// LetterSpacing (Tc operator) tests (WP1)
// ---------------------------------------------------------------------------

func TestRenderTextLetterSpacing(t *testing.T) {
	r, _ := newTestRenderer(t)
	_ = r.BeginPage(document.Size{Width: 595, Height: 842})
	style := document.DefaultStyle()
	style.LetterSpacing = 1.5
	err := r.RenderText("hello", document.Point{X: 10, Y: 10}, style)
	if err != nil {
		t.Fatal(err)
	}
	content := string(r.pageContent)
	if !strings.Contains(content, "1.5 Tc") {
		t.Error("Should contain Tc operator with letter spacing value")
	}
	if !strings.Contains(content, "0 Tc") {
		t.Error("Should reset Tc to 0 after text")
	}
}

func TestRenderTextNoLetterSpacing(t *testing.T) {
	r, _ := newTestRenderer(t)
	_ = r.BeginPage(document.Size{Width: 595, Height: 842})
	style := document.DefaultStyle()
	err := r.RenderText("hello", document.Point{X: 10, Y: 10}, style)
	if err != nil {
		t.Fatal(err)
	}
	content := string(r.pageContent)
	if strings.Contains(content, "Tc") {
		t.Error("Should not contain Tc operator when LetterSpacing is 0")
	}
}

// ---------------------------------------------------------------------------
// TextDecoration tests (WP3)
// ---------------------------------------------------------------------------

func TestRenderTextDecorationUnderline(t *testing.T) {
	r, _ := newTestRenderer(t)
	_ = r.BeginPage(document.Size{Width: 595, Height: 842})

	style := document.DefaultStyle()
	style.FontSize = 12
	style.TextDecoration = document.DecorationUnderline
	textNode := &document.Text{Content: "Hello", TextStyle: style}
	pn := layout.PlacedNode{
		Node:     textNode,
		Position: document.Point{X: 10, Y: 20},
		Size:     document.Size{Width: 30, Height: 14.4},
	}
	err := r.renderPlacedNode(pn, 0, 0)
	if err != nil {
		t.Fatal(err)
	}
	content := string(r.pageContent)
	// Should contain line drawing for underline (moveto + lineto).
	if !strings.Contains(content, "m\n") || !strings.Contains(content, "l\n") {
		t.Error("Underline should produce line drawing operators")
	}
}

func TestRenderTextDecorationStrikethrough(t *testing.T) {
	r, _ := newTestRenderer(t)
	_ = r.BeginPage(document.Size{Width: 595, Height: 842})

	style := document.DefaultStyle()
	style.FontSize = 12
	style.TextDecoration = document.DecorationStrikethrough
	textNode := &document.Text{Content: "Hello", TextStyle: style}
	pn := layout.PlacedNode{
		Node:     textNode,
		Position: document.Point{X: 10, Y: 20},
		Size:     document.Size{Width: 30, Height: 14.4},
	}
	err := r.renderPlacedNode(pn, 0, 0)
	if err != nil {
		t.Fatal(err)
	}
	content := string(r.pageContent)
	if !strings.Contains(content, "m\n") || !strings.Contains(content, "l\n") {
		t.Error("Strikethrough should produce line drawing operators")
	}
}

func TestRenderTextDecorationCombined(t *testing.T) {
	r, _ := newTestRenderer(t)
	_ = r.BeginPage(document.Size{Width: 595, Height: 842})

	style := document.DefaultStyle()
	style.FontSize = 12
	style.TextDecoration = document.DecorationUnderline | document.DecorationStrikethrough
	textNode := &document.Text{Content: "Hello", TextStyle: style}
	pn := layout.PlacedNode{
		Node:     textNode,
		Position: document.Point{X: 10, Y: 20},
		Size:     document.Size{Width: 30, Height: 14.4},
	}
	err := r.renderPlacedNode(pn, 0, 0)
	if err != nil {
		t.Fatal(err)
	}
	content := string(r.pageContent)
	// Count the number of line-drawing operations (moveto/lineto pairs).
	// Each decoration produces one m + one l, so we expect at least 2 m operators.
	mCount := strings.Count(content, " m\n")
	if mCount < 2 {
		t.Errorf("Combined decoration should produce at least 2 line drawings, got %d moveto operators", mCount)
	}
}

func TestRenderLine(t *testing.T) {
	r, _ := newTestRenderer(t)
	_ = r.BeginPage(document.Size{Width: 595, Height: 842})
	err := r.RenderLine(
		document.Point{X: 10, Y: 20},
		document.Point{X: 200, Y: 20},
		LineStyle{Color: pdf.Black, Width: 1.5},
	)
	if err != nil {
		t.Fatal(err)
	}
	content := string(r.pageContent)
	if !strings.Contains(content, "m\n") {
		t.Error("RenderLine should produce moveto")
	}
	if !strings.Contains(content, "l\n") {
		t.Error("RenderLine should produce lineto")
	}
	if !strings.Contains(content, "1.5 w") {
		t.Error("RenderLine should set stroke width")
	}
	if !strings.Contains(content, "S\n") {
		t.Error("RenderLine should stroke")
	}
}
