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
// NewOverlayRenderer tests
// ---------------------------------------------------------------------------

func TestNewOverlayRenderer(t *testing.T) {
	r := NewOverlayRenderer(595, 842, nil, nil)
	if r == nil {
		t.Fatal("NewOverlayRenderer returned nil")
	}
	if r.pageWidth != 595 {
		t.Errorf("pageWidth = %v, want 595", r.pageWidth)
	}
	if r.pageHeight != 842 {
		t.Errorf("pageHeight = %v, want 842", r.pageHeight)
	}
	if r.fontMap == nil {
		t.Error("fontMap should be initialized")
	}
	if r.fontObjects == nil {
		t.Error("fontObjects should be initialized")
	}
	if r.imageMap == nil {
		t.Error("imageMap should be initialized")
	}
	if r.imageObjects == nil {
		t.Error("imageObjects should be initialized")
	}
}

func TestNewOverlayRendererWithFontData(t *testing.T) {
	fontData := map[string][]byte{
		"TestFont": {0x00, 0x01, 0x02},
	}
	r := NewOverlayRenderer(595, 842, nil, fontData)
	if r.fontDataMap == nil {
		t.Error("fontDataMap should be set")
	}
	if _, ok := r.fontDataMap["TestFont"]; !ok {
		t.Error("fontDataMap should contain TestFont")
	}
}

// ---------------------------------------------------------------------------
// overlayImageKey tests
// ---------------------------------------------------------------------------

func TestOverlayImageKey(t *testing.T) {
	data1 := []byte{0x01, 0x02, 0x03}
	data2 := []byte{0x04, 0x05, 0x06}
	key1 := overlayImageKey(data1)
	key2 := overlayImageKey(data2)

	if key1 == "" {
		t.Error("overlayImageKey should not return empty string")
	}
	if !strings.HasPrefix(key1, "ov_") {
		t.Errorf("overlayImageKey should start with 'ov_', got %q", key1)
	}
	if key1 == key2 {
		t.Error("different data should produce different keys")
	}
	// Same data should produce the same key.
	if key1 != overlayImageKey(data1) {
		t.Error("same data should produce the same key")
	}
}

// ---------------------------------------------------------------------------
// ensureFont tests
// ---------------------------------------------------------------------------

func TestOverlayEnsureFont_Default(t *testing.T) {
	r := NewOverlayRenderer(595, 842, nil, nil)
	resName := r.ensureFont("")
	if resName != "OvF1" {
		t.Errorf("ensureFont('') = %q, want 'OvF1'", resName)
	}
	// Should be registered as "Helvetica".
	if _, ok := r.fontMap["Helvetica"]; !ok {
		t.Error("empty family should register as 'Helvetica'")
	}
	if fo, ok := r.fontObjects["Helvetica"]; !ok {
		t.Error("fontObjects should contain Helvetica")
	} else if fo.ResName != "OvF1" {
		t.Errorf("fontObject ResName = %q, want 'OvF1'", fo.ResName)
	}
}

func TestOverlayEnsureFont_Deduplication(t *testing.T) {
	r := NewOverlayRenderer(595, 842, nil, nil)
	res1 := r.ensureFont("Helvetica")
	res2 := r.ensureFont("Helvetica")
	if res1 != res2 {
		t.Errorf("ensureFont should return same name on duplicate: %q vs %q", res1, res2)
	}
	if r.fontCount != 1 {
		t.Errorf("fontCount = %d, want 1", r.fontCount)
	}
}

func TestOverlayEnsureFont_WithFontData(t *testing.T) {
	fontData := map[string][]byte{
		"NotoSansJP": {0xAA, 0xBB},
	}
	r := NewOverlayRenderer(595, 842, nil, fontData)
	resName := r.ensureFont("NotoSansJP")
	if resName == "" {
		t.Error("ensureFont should return non-empty resource name")
	}
	fo, ok := r.fontObjects["NotoSansJP"]
	if !ok {
		t.Fatal("fontObjects should contain NotoSansJP")
	}
	if fo.Data == nil {
		t.Error("fontObject Data should be set from fontDataMap")
	}
}

func TestOverlayEnsureFont_MultipleFonts(t *testing.T) {
	r := NewOverlayRenderer(595, 842, nil, nil)
	res1 := r.ensureFont("Helvetica")
	res2 := r.ensureFont("Times-Roman")
	res3 := r.ensureFont("Courier")
	if res1 == res2 || res2 == res3 || res1 == res3 {
		t.Error("different fonts should have different resource names")
	}
	if r.fontCount != 3 {
		t.Errorf("fontCount = %d, want 3", r.fontCount)
	}
}

// ---------------------------------------------------------------------------
// ensureImage tests
// ---------------------------------------------------------------------------

func TestOverlayEnsureImage_JPEG(t *testing.T) {
	r := NewOverlayRenderer(595, 842, nil, nil)
	src := document.ImageSource{
		Data:   []byte{0xFF, 0xD8, 0xFF, 0xE0},
		Format: document.ImageJPEG,
		Width:  100,
		Height: 80,
	}
	key := overlayImageKey(src.Data)
	resName := r.ensureImage(key, src)
	if resName != "OvIm1" {
		t.Errorf("ensureImage = %q, want 'OvIm1'", resName)
	}
	io, ok := r.imageObjects[key]
	if !ok {
		t.Fatal("imageObjects should contain the image")
	}
	if io.Filter != "DCTDecode" {
		t.Errorf("JPEG filter = %q, want 'DCTDecode'", io.Filter)
	}
	if io.ColorSpace != "DeviceRGB" {
		t.Errorf("colorSpace = %q, want 'DeviceRGB'", io.ColorSpace)
	}
	if io.Width != 100 || io.Height != 80 {
		t.Errorf("dimensions = %dx%d, want 100x80", io.Width, io.Height)
	}
}

func TestOverlayEnsureImage_PNG(t *testing.T) {
	r := NewOverlayRenderer(595, 842, nil, nil)
	pngData := testPNG(t, 4, 4)
	src := document.ImageSource{
		Data:   pngData,
		Format: document.ImagePNG,
		Width:  4,
		Height: 4,
	}
	key := overlayImageKey(src.Data)
	resName := r.ensureImage(key, src)
	if resName == "" {
		t.Error("ensureImage should return non-empty resource name")
	}
	io := r.imageObjects[key]
	if io.ColorSpace != "DeviceRGB" {
		t.Errorf("PNG colorSpace = %q, want 'DeviceRGB'", io.ColorSpace)
	}
	if io.Width != 4 || io.Height != 4 {
		t.Errorf("PNG dimensions = %dx%d, want 4x4", io.Width, io.Height)
	}
}

func TestOverlayEnsureImage_PNGWithAlpha(t *testing.T) {
	r := NewOverlayRenderer(595, 842, nil, nil)
	// Create a transparent PNG.
	img := image.NewRGBA(image.Rect(0, 0, 2, 2))
	img.Set(0, 0, color.RGBA{R: 255, A: 128})
	img.Set(1, 0, color.RGBA{G: 255, A: 200})
	img.Set(0, 1, color.RGBA{B: 255, A: 0})
	img.Set(1, 1, color.RGBA{R: 128, A: 100})
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatal(err)
	}

	src := document.ImageSource{
		Data:   buf.Bytes(),
		Format: document.ImagePNG,
		Width:  2,
		Height: 2,
	}
	key := overlayImageKey(src.Data)
	r.ensureImage(key, src)
	io := r.imageObjects[key]
	if io.SmaskData == nil {
		t.Error("transparent PNG should have SmaskData")
	}
}

func TestOverlayEnsureImage_Deduplication(t *testing.T) {
	r := NewOverlayRenderer(595, 842, nil, nil)
	src := document.ImageSource{
		Data:   []byte{0xFF, 0xD8},
		Format: document.ImageJPEG,
		Width:  10,
		Height: 10,
	}
	key := overlayImageKey(src.Data)
	res1 := r.ensureImage(key, src)
	res2 := r.ensureImage(key, src)
	if res1 != res2 {
		t.Error("same image key should return same resource name")
	}
	if r.imageCount != 1 {
		t.Errorf("imageCount = %d, want 1", r.imageCount)
	}
}

func TestOverlayEnsureImage_DefaultFormat(t *testing.T) {
	r := NewOverlayRenderer(595, 842, nil, nil)
	src := document.ImageSource{
		Data:   []byte{0x01, 0x02, 0x03},
		Format: 99, // unknown format
		Width:  10,
		Height: 10,
	}
	key := overlayImageKey(src.Data)
	r.ensureImage(key, src)
	io := r.imageObjects[key]
	if io.Filter != "" {
		t.Errorf("unknown format filter = %q, want empty", io.Filter)
	}
	if io.ColorSpace != "DeviceRGB" {
		t.Errorf("colorSpace = %q, want 'DeviceRGB'", io.ColorSpace)
	}
}

// ---------------------------------------------------------------------------
// renderText tests (via RenderOverlay)
// ---------------------------------------------------------------------------

func TestOverlayRenderText(t *testing.T) {
	r := NewOverlayRenderer(595, 842, nil, nil)
	style := document.DefaultStyle()
	style.FontSize = 14
	textNode := &document.Text{
		Content:   "Hello Overlay",
		TextStyle: style,
	}
	nodes := []layout.PlacedNode{
		{
			Node:     textNode,
			Position: document.Point{X: 72, Y: 100},
			Size:     document.Size{Width: 200, Height: 16.8},
		},
	}

	result, err := r.RenderOverlay(nodes)
	if err != nil {
		t.Fatalf("RenderOverlay error: %v", err)
	}
	content := string(result.Content)
	if !strings.Contains(content, "BT") {
		t.Error("should contain BT")
	}
	if !strings.Contains(content, "ET") {
		t.Error("should contain ET")
	}
	if !strings.Contains(content, "Tf") {
		t.Error("should contain Tf")
	}
	if !strings.Contains(content, "(Hello Overlay) Tj") {
		t.Error("should contain the text string")
	}
}

func TestOverlayRenderTextEmpty(t *testing.T) {
	r := NewOverlayRenderer(595, 842, nil, nil)
	style := document.DefaultStyle()
	textNode := &document.Text{
		Content:   "",
		TextStyle: style,
	}
	nodes := []layout.PlacedNode{
		{
			Node:     textNode,
			Position: document.Point{X: 10, Y: 10},
			Size:     document.Size{Width: 100, Height: 14},
		},
	}
	result, err := r.RenderOverlay(nodes)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(result.Content), "BT") {
		t.Error("empty text should not produce BT operator")
	}
}

func TestOverlayRenderTextDefaultFontSize(t *testing.T) {
	r := NewOverlayRenderer(595, 842, nil, nil)
	style := document.DefaultStyle()
	style.FontSize = 0 // should default to 12
	textNode := &document.Text{
		Content:   "test",
		TextStyle: style,
	}
	nodes := []layout.PlacedNode{
		{
			Node:     textNode,
			Position: document.Point{X: 10, Y: 10},
			Size:     document.Size{Width: 100, Height: 14},
		},
	}
	result, err := r.RenderOverlay(nodes)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(result.Content), "12 Tf") {
		t.Error("should use default font size 12")
	}
}

func TestOverlayRenderTextWordSpacing(t *testing.T) {
	r := NewOverlayRenderer(595, 842, nil, nil)
	style := document.DefaultStyle()
	style.WordSpacing = 3.5
	textNode := &document.Text{
		Content:   "hello world",
		TextStyle: style,
	}
	nodes := []layout.PlacedNode{
		{
			Node:     textNode,
			Position: document.Point{X: 10, Y: 10},
			Size:     document.Size{Width: 100, Height: 14},
		},
	}
	result, err := r.RenderOverlay(nodes)
	if err != nil {
		t.Fatal(err)
	}
	content := string(result.Content)
	if !strings.Contains(content, "3.5 Tw") {
		t.Error("should contain Tw operator")
	}
	if !strings.Contains(content, "0 Tw") {
		t.Error("should reset Tw to 0")
	}
}

func TestOverlayRenderTextLetterSpacing(t *testing.T) {
	r := NewOverlayRenderer(595, 842, nil, nil)
	style := document.DefaultStyle()
	style.LetterSpacing = 2.0
	textNode := &document.Text{
		Content:   "hello",
		TextStyle: style,
	}
	nodes := []layout.PlacedNode{
		{
			Node:     textNode,
			Position: document.Point{X: 10, Y: 10},
			Size:     document.Size{Width: 100, Height: 14},
		},
	}
	result, err := r.RenderOverlay(nodes)
	if err != nil {
		t.Fatal(err)
	}
	content := string(result.Content)
	if !strings.Contains(content, "2 Tc") {
		t.Error("should contain Tc operator")
	}
	if !strings.Contains(content, "0 Tc") {
		t.Error("should reset Tc to 0")
	}
}

func TestOverlayRenderTextYConversion(t *testing.T) {
	r := NewOverlayRenderer(595, 842, nil, nil)
	style := document.DefaultStyle()
	style.FontSize = 12
	textNode := &document.Text{
		Content:   "test",
		TextStyle: style,
	}
	nodes := []layout.PlacedNode{
		{
			Node:     textNode,
			Position: document.Point{X: 0, Y: 100},
			Size:     document.Size{Width: 100, Height: 14},
		},
	}
	result, err := r.RenderOverlay(nodes)
	if err != nil {
		t.Fatal(err)
	}
	// pdfY = 842 - 100 - 12 = 730
	if !strings.Contains(string(result.Content), "0 730 Td") {
		t.Errorf("expected Y conversion, got: %s", string(result.Content))
	}
}

// ---------------------------------------------------------------------------
// renderRect tests (via RenderOverlay with background)
// ---------------------------------------------------------------------------

func TestOverlayRenderRectFillOnly(t *testing.T) {
	r := NewOverlayRenderer(595, 842, nil, nil)
	bg := pdf.Red
	boxNode := &document.Box{
		BoxStyle: document.BoxStyle{
			Background: &bg,
		},
	}
	nodes := []layout.PlacedNode{
		{
			Node:     boxNode,
			Position: document.Point{X: 10, Y: 20},
			Size:     document.Size{Width: 100, Height: 50},
		},
	}
	result, err := r.RenderOverlay(nodes)
	if err != nil {
		t.Fatal(err)
	}
	content := string(result.Content)
	if !strings.Contains(content, "q\n") {
		t.Error("should save graphics state")
	}
	if !strings.Contains(content, "Q\n") {
		t.Error("should restore graphics state")
	}
	if !strings.Contains(content, "re\n") {
		t.Error("should contain rectangle operator")
	}
	if !strings.Contains(content, "f\n") {
		t.Error("should contain fill operator 'f'")
	}
}

func TestOverlayRenderRectStrokeOnly(t *testing.T) {
	r := NewOverlayRenderer(595, 842, nil, nil)
	stroke := pdf.Blue
	r.renderRect(document.Rectangle{X: 10, Y: 20, Width: 100, Height: 50}, RectStyle{
		StrokeColor: &stroke,
		StrokeWidth: 2,
	})
	content := string(r.content)
	if !strings.Contains(content, "S\n") {
		t.Error("should contain stroke operator 'S'")
	}
	if !strings.Contains(content, "2 w\n") {
		t.Error("should set stroke width")
	}
}

func TestOverlayRenderRectFillAndStroke(t *testing.T) {
	r := NewOverlayRenderer(595, 842, nil, nil)
	fill := pdf.Red
	stroke := pdf.Blue
	r.renderRect(document.Rectangle{X: 10, Y: 20, Width: 100, Height: 50}, RectStyle{
		FillColor:   &fill,
		StrokeColor: &stroke,
		StrokeWidth: 1,
	})
	content := string(r.content)
	if !strings.Contains(content, "B\n") {
		t.Error("should contain fill-and-stroke operator 'B'")
	}
}

func TestOverlayRenderRectNoFillNoStroke(t *testing.T) {
	r := NewOverlayRenderer(595, 842, nil, nil)
	r.renderRect(document.Rectangle{X: 10, Y: 20, Width: 100, Height: 50}, RectStyle{})
	content := string(r.content)
	if !strings.Contains(content, "n\n") {
		t.Error("should contain no-op path operator 'n'")
	}
}

func TestOverlayRenderRectYConversion(t *testing.T) {
	r := NewOverlayRenderer(595, 842, nil, nil)
	fill := pdf.Red
	r.renderRect(document.Rectangle{X: 0, Y: 100, Width: 200, Height: 50}, RectStyle{
		FillColor: &fill,
	})
	content := string(r.content)
	// pdfY = 842 - 100 - 50 = 692
	if !strings.Contains(content, "0 692 200 50 re") {
		t.Errorf("expected Y conversion in rect, got: %s", content)
	}
}

// ---------------------------------------------------------------------------
// renderImage tests (via RenderOverlay)
// ---------------------------------------------------------------------------

func TestOverlayRenderImage(t *testing.T) {
	r := NewOverlayRenderer(595, 842, nil, nil)
	imgData := []byte{0xFF, 0xD8, 0xFF, 0xE0, 0x00, 0x10}
	imgNode := &document.Image{
		Source: document.ImageSource{
			Data:   imgData,
			Format: document.ImageJPEG,
			Width:  100,
			Height: 80,
		},
	}
	nodes := []layout.PlacedNode{
		{
			Node:     imgNode,
			Position: document.Point{X: 50, Y: 100},
			Size:     document.Size{Width: 200, Height: 160},
		},
	}
	result, err := r.RenderOverlay(nodes)
	if err != nil {
		t.Fatal(err)
	}
	content := string(result.Content)
	if !strings.Contains(content, "q\n") {
		t.Error("should save graphics state")
	}
	if !strings.Contains(content, "cm\n") {
		t.Error("should contain cm operator")
	}
	if !strings.Contains(content, "Do\n") {
		t.Error("should contain Do operator")
	}
	if !strings.Contains(content, "Q\n") {
		t.Error("should restore graphics state")
	}
	// Should have registered an image.
	if len(r.imageObjects) != 1 {
		t.Errorf("imageObjects count = %d, want 1", len(r.imageObjects))
	}
}

// ---------------------------------------------------------------------------
// RenderOverlay comprehensive tests
// ---------------------------------------------------------------------------

func TestOverlayRenderNilNode(t *testing.T) {
	r := NewOverlayRenderer(595, 842, nil, nil)
	nodes := []layout.PlacedNode{
		{Node: nil},
	}
	result, err := r.RenderOverlay(nodes)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Content) != 0 {
		t.Error("nil node should produce no content")
	}
}

func TestOverlayRenderEmptyNodes(t *testing.T) {
	r := NewOverlayRenderer(595, 842, nil, nil)
	result, err := r.RenderOverlay(nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Content) != 0 {
		t.Error("empty nodes should produce no content")
	}
	if len(result.FontObjects) != 0 {
		t.Error("empty nodes should have no font objects")
	}
}

func TestOverlayRenderResources(t *testing.T) {
	r := NewOverlayRenderer(595, 842, nil, nil)
	style := document.DefaultStyle()
	style.FontSize = 12
	textNode := &document.Text{
		Content:   "test",
		TextStyle: style,
	}
	nodes := []layout.PlacedNode{
		{
			Node:     textNode,
			Position: document.Point{X: 10, Y: 10},
			Size:     document.Size{Width: 100, Height: 14},
		},
	}
	result, err := r.RenderOverlay(nodes)
	if err != nil {
		t.Fatal(err)
	}
	// Should have a Font entry in Resources.
	if result.Resources == nil {
		t.Fatal("Resources should not be nil")
	}
	if _, ok := result.Resources[pdf.Name("Font")]; !ok {
		t.Error("Resources should contain Font dict")
	}
}

func TestOverlayRenderNestedChildren(t *testing.T) {
	r := NewOverlayRenderer(595, 842, nil, nil)
	childText := &document.Text{
		Content:   "Child",
		TextStyle: document.DefaultStyle(),
	}
	parentBox := &document.Box{}
	nodes := []layout.PlacedNode{
		{
			Node:     parentBox,
			Position: document.Point{X: 72, Y: 72},
			Size:     document.Size{Width: 200, Height: 100},
			Children: []layout.PlacedNode{
				{
					Node:     childText,
					Position: document.Point{X: 10, Y: 10},
					Size:     document.Size{Width: 100, Height: 14},
				},
			},
		},
	}
	result, err := r.RenderOverlay(nodes)
	if err != nil {
		t.Fatal(err)
	}
	content := string(result.Content)
	if !strings.Contains(content, "(Child) Tj") {
		t.Error("should render nested child text")
	}
}

func TestOverlayRenderTextWithParentChildren(t *testing.T) {
	// Parent text with children should NOT render its own content.
	r := NewOverlayRenderer(595, 842, nil, nil)
	childText := &document.Text{
		Content:   "Line1",
		TextStyle: document.DefaultStyle(),
	}
	parentText := &document.Text{
		Content:   "ParentContent",
		TextStyle: document.DefaultStyle(),
	}
	nodes := []layout.PlacedNode{
		{
			Node:     parentText,
			Position: document.Point{X: 10, Y: 10},
			Size:     document.Size{Width: 200, Height: 28},
			Children: []layout.PlacedNode{
				{
					Node:     childText,
					Position: document.Point{X: 0, Y: 0},
					Size:     document.Size{Width: 100, Height: 14},
				},
			},
		},
	}
	result, err := r.RenderOverlay(nodes)
	if err != nil {
		t.Fatal(err)
	}
	content := string(result.Content)
	if strings.Contains(content, "(ParentContent)") {
		t.Error("parent text with children should not render its own content")
	}
	if !strings.Contains(content, "(Line1)") {
		t.Error("should render child text")
	}
}

// ---------------------------------------------------------------------------
// RenderOverlayContent tests
// ---------------------------------------------------------------------------

func TestRenderOverlayContent_EmptyNodes(t *testing.T) {
	result, err := RenderOverlayContent(
		nil,
		document.A4,
		document.Edges{},
		nil, nil, nil,
	)
	if err != nil {
		t.Fatal(err)
	}
	if result == nil {
		t.Fatal("result should not be nil")
	}
}

func TestRenderOverlayContent_WithText(t *testing.T) {
	nodes := []document.DocumentNode{
		&document.Text{
			Content: "Overlay Text",
			TextStyle: document.Style{
				FontSize:   12,
				LineHeight: 1.2,
				Color:      pdf.Black,
			},
		},
	}
	result, err := RenderOverlayContent(
		nodes,
		document.A4,
		document.UniformEdges(document.Pt(72)),
		nil, nil, nil,
	)
	if err != nil {
		t.Fatal(err)
	}
	if result == nil {
		t.Fatal("result should not be nil")
	}
	// With no font resolver, layout may or may not produce content,
	// but the function should not panic.
}

// ---------------------------------------------------------------------------
// WriteOverlayToModifier tests
// ---------------------------------------------------------------------------

func newTestModifier(t *testing.T) *pdf.Modifier {
	t.Helper()
	// Create a minimal valid PDF so we can create a Reader then Modifier.
	var buf bytes.Buffer
	w := pdf.NewWriter(&buf)
	w.SetInfo(pdf.DocumentInfo{Title: "test"})
	page := pdf.PageObject{
		MediaBox:  pdf.Rectangle{LLX: 0, LLY: 0, URX: 595, URY: 842},
		Resources: pdf.ResourceDict{},
		Contents:  nil,
	}
	if err := w.AddPage(page); err != nil {
		t.Fatal(err)
	}
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}

	r, err := pdf.NewReader(buf.Bytes())
	if err != nil {
		t.Fatal(err)
	}
	return pdf.NewModifier(r)
}

func TestWriteOverlayToModifier_Nil(t *testing.T) {
	m := newTestModifier(t)
	content, res, err := WriteOverlayToModifier(nil, m)
	if err != nil {
		t.Fatal(err)
	}
	if content != nil {
		t.Error("nil result should return nil content")
	}
	if res != nil {
		t.Error("nil result should return nil resources")
	}
}

func TestWriteOverlayToModifier_EmptyContent(t *testing.T) {
	m := newTestModifier(t)
	result := &OverlayResult{
		Content: nil,
	}
	content, res, err := WriteOverlayToModifier(result, m)
	if err != nil {
		t.Fatal(err)
	}
	if content != nil {
		t.Error("empty content should return nil")
	}
	if res != nil {
		t.Error("empty content should return nil resources")
	}
}

func TestWriteOverlayToModifier_StandardFont(t *testing.T) {
	m := newTestModifier(t)
	result := &OverlayResult{
		Content: []byte("BT /OvF1 12 Tf (test) Tj ET"),
		FontObjects: map[string]fontObject{
			"Helvetica": {
				ResName: "OvF1",
				Family:  "Helvetica",
				Data:    nil, // standard font
			},
		},
		ImageObjects: map[string]imageObject{},
	}
	content, res, err := WriteOverlayToModifier(result, m)
	if err != nil {
		t.Fatal(err)
	}
	if content == nil {
		t.Error("content should not be nil")
	}
	if res == nil {
		t.Fatal("resources should not be nil for font")
	}
	fontDict, ok := (*res)[pdf.Name("Font")]
	if !ok {
		t.Fatal("resources should contain Font")
	}
	if fontDict == nil {
		t.Error("font dict should not be nil")
	}
}

func TestWriteOverlayToModifier_TrueTypeFont(t *testing.T) {
	m := newTestModifier(t)
	result := &OverlayResult{
		Content: []byte("BT /OvF1 12 Tf <0048> Tj ET"),
		FontObjects: map[string]fontObject{
			"NotoSansJP": {
				ResName: "OvF1",
				Family:  "NotoSansJP",
				Data:    []byte{0x00, 0x01, 0x02, 0x03}, // dummy font data
			},
		},
		ImageObjects: map[string]imageObject{},
	}
	content, res, err := WriteOverlayToModifier(result, m)
	if err != nil {
		t.Fatal(err)
	}
	if content == nil {
		t.Error("content should not be nil")
	}
	if res == nil {
		t.Fatal("resources should not be nil")
	}
}

func TestWriteOverlayToModifier_JPEGImage(t *testing.T) {
	m := newTestModifier(t)
	result := &OverlayResult{
		Content:     []byte("q 100 0 0 80 50 700 cm /OvIm1 Do Q"),
		FontObjects: map[string]fontObject{},
		ImageObjects: map[string]imageObject{
			"img1": {
				ResName:    "OvIm1",
				Data:       []byte{0xFF, 0xD8, 0xFF},
				Width:      100,
				Height:     80,
				ColorSpace: "DeviceRGB",
				Filter:     "DCTDecode",
			},
		},
	}
	content, res, err := WriteOverlayToModifier(result, m)
	if err != nil {
		t.Fatal(err)
	}
	if content == nil {
		t.Error("content should not be nil")
	}
	if res == nil {
		t.Fatal("resources should not be nil")
	}
	if _, ok := (*res)[pdf.Name("XObject")]; !ok {
		t.Error("resources should contain XObject")
	}
}

func TestWriteOverlayToModifier_PNGImageWithSmask(t *testing.T) {
	m := newTestModifier(t)
	result := &OverlayResult{
		Content:     []byte("q 10 0 0 10 0 830 cm /OvIm1 Do Q"),
		FontObjects: map[string]fontObject{},
		ImageObjects: map[string]imageObject{
			"img1": {
				ResName:    "OvIm1",
				Data:       []byte{0xFF, 0x00, 0x00, 0x00, 0xFF, 0x00, 0x00, 0x00, 0xFF},
				SmaskData:  []byte{0x80, 0xFF, 0x00},
				Width:      3,
				Height:     1,
				ColorSpace: "DeviceRGB",
				Filter:     "",
			},
		},
	}
	content, res, err := WriteOverlayToModifier(result, m)
	if err != nil {
		t.Fatal(err)
	}
	if content == nil {
		t.Error("content should not be nil")
	}
	if res == nil {
		t.Fatal("resources should not be nil")
	}
}

func TestWriteOverlayToModifier_NoResources(t *testing.T) {
	m := newTestModifier(t)
	result := &OverlayResult{
		Content:      []byte("q Q"),
		FontObjects:  map[string]fontObject{},
		ImageObjects: map[string]imageObject{},
	}
	content, res, err := WriteOverlayToModifier(result, m)
	if err != nil {
		t.Fatal(err)
	}
	if content == nil {
		t.Error("content should not be nil")
	}
	if res != nil {
		t.Error("no resources should return nil")
	}
}

func TestWriteOverlayToModifier_FontsAndImages(t *testing.T) {
	m := newTestModifier(t)
	result := &OverlayResult{
		Content: []byte("BT /OvF1 12 Tf (hi) Tj ET q /OvIm1 Do Q"),
		FontObjects: map[string]fontObject{
			"Helvetica": {ResName: "OvF1", Family: "Helvetica"},
		},
		ImageObjects: map[string]imageObject{
			"img1": {
				ResName:    "OvIm1",
				Data:       []byte{0xFF, 0xD8},
				Width:      10,
				Height:     10,
				ColorSpace: "DeviceRGB",
				Filter:     "DCTDecode",
			},
		},
	}
	content, res, err := WriteOverlayToModifier(result, m)
	if err != nil {
		t.Fatal(err)
	}
	if content == nil {
		t.Error("content should not be nil")
	}
	if res == nil {
		t.Fatal("resources should not be nil")
	}
	if _, ok := (*res)[pdf.Name("Font")]; !ok {
		t.Error("should have Font in resources")
	}
	if _, ok := (*res)[pdf.Name("XObject")]; !ok {
		t.Error("should have XObject in resources")
	}
}
