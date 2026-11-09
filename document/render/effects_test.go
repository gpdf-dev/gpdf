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

func effectsPNG(t *testing.T, w, h int) []byte {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.Set(x, y, color.RGBA{R: 200, A: 255})
		}
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

// renderNode renders a single placed node on a fresh A4 page and returns
// the page content stream.
func renderNode(t *testing.T, pn layout.PlacedNode) (*PDFRenderer, string) {
	t.Helper()
	r, _ := newTestRenderer(t)
	if err := r.BeginPage(document.Size{Width: 595, Height: 842}); err != nil {
		t.Fatal(err)
	}
	if err := r.renderPlacedNode(pn, 0, 0); err != nil {
		t.Fatal(err)
	}
	return r, string(r.pageContent)
}

func TestRoundedRectClampsRadius(t *testing.T) {
	path := roundedRect(document.Rectangle{X: 10, Y: 20, Width: 40, Height: 10}, 100)
	start := path.Segments[0].Points[0]
	if start != (document.Point{X: 15, Y: 20}) {
		t.Errorf("start = %+v, want radius clamped to half the height", start)
	}
	if last := path.Segments[len(path.Segments)-2].Points[2]; last != start {
		t.Errorf("path ends at %+v, want %+v", last, start)
	}
	if path.Segments[len(path.Segments)-1].Op != document.PathClose {
		t.Error("path should be closed")
	}
}

func TestBorderRadius(t *testing.T) {
	bg := pdf.RGBHex(0xF5F7F7)
	pngData := effectsPNG(t, 2, 2)
	border := document.UniformBorder(document.Pt(2), document.BorderSolid, pdf.Black)
	mixed := border
	mixed.Left.Color = pdf.Red

	tests := []struct {
		name       string
		node       document.DocumentNode
		want, deny []string
	}{
		{"square box", &document.Box{BoxStyle: document.BoxStyle{Background: &bg}},
			[]string{" re\n"}, []string{" c\n"}},
		{"rounded box", &document.Box{BoxStyle: document.BoxStyle{Background: &bg, BorderRadius: document.Pt(6)}},
			[]string{" c\n", "f\n"}, []string{" re\n"}},
		{"rounded uniform border", &document.Box{BoxStyle: document.BoxStyle{Border: border, BorderRadius: document.Pt(6)}},
			[]string{" c\n", "2 w\n", "S\n"}, []string{" re\n"}},
		{"rounded mixed border falls back to square", &document.Box{BoxStyle: document.BoxStyle{Border: mixed, BorderRadius: document.Pt(6)}},
			[]string{" re\n"}, []string{" c\n"}},
		{"rounded image", &document.Image{
			Source:   document.ImageSource{Data: pngData, Format: document.ImagePNG, Width: 2, Height: 2},
			ImgStyle: document.Style{BorderRadius: document.Pt(6)},
		}, []string{"W n\n", " c\n", "Do\n"}, nil},
		{"square image has no clip", &document.Image{
			Source: document.ImageSource{Data: pngData, Format: document.ImagePNG, Width: 2, Height: 2},
		}, []string{"Do\n"}, []string{"W n\n"}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			pn := layout.PlacedNode{Node: tc.node, Position: document.Point{X: 10, Y: 10}, Size: document.Size{Width: 100, Height: 50}}
			_, content := renderNode(t, pn)
			for _, w := range tc.want {
				if !strings.Contains(content, w) {
					t.Errorf("content missing %q:\n%s", w, content)
				}
			}
			for _, d := range tc.deny {
				if strings.Contains(content, d) {
					t.Errorf("content unexpectedly contains %q:\n%s", d, content)
				}
			}
		})
	}
}

func TestOpacityGroupsAndResources(t *testing.T) {
	bg := pdf.RGBHex(0x336699)
	inner := &document.Text{Content: "hi", TextStyle: document.Style{FontSize: 12, Opacity: 0.8}}
	same := &document.Text{Content: "same", TextStyle: document.Style{FontSize: 12, Opacity: 0.5}}
	box := &document.Box{BoxStyle: document.BoxStyle{Background: &bg, Opacity: 0.5}}
	pn := layout.PlacedNode{
		Node: box,
		Size: document.Size{Width: 100, Height: 50},
		Children: []layout.PlacedNode{
			{Node: inner, Size: document.Size{Width: 20, Height: 14}},
			{Node: same, Size: document.Size{Width: 20, Height: 14}},
		},
	}
	r, content := renderNode(t, pn)

	if got := strings.Count(content, " gs\n"); got != 2 {
		t.Errorf("gs operators = %d, want 2 (box + overriding child; equal alpha is not re-emitted):\n%s", got, content)
	}
	if strings.Count(content, "q\n") != strings.Count(content, "Q\n") {
		t.Errorf("unbalanced q/Q:\n%s", content)
	}
	if len(r.gsDicts) != 2 {
		t.Fatalf("ExtGState count = %d, want 2", len(r.gsDicts))
	}
	if r.alpha != 1 {
		t.Errorf("alpha after render = %g, want 1", r.alpha)
	}

	// Same alpha on a later node reuses the resource.
	if name := r.ensureExtGState(0.5); name != "GS1" {
		t.Errorf("ensureExtGState(0.5) = %q, want GS1", name)
	}

	if err := r.EndPage(); err != nil {
		t.Fatal(err)
	}
}

func TestOpacityUnsetIsNoop(t *testing.T) {
	bg := pdf.RGBHex(0x336699)
	for _, op := range []float64{0, -1, 1} {
		pn := layout.PlacedNode{
			Node: &document.Box{BoxStyle: document.BoxStyle{Background: &bg, Opacity: op}},
			Size: document.Size{Width: 10, Height: 10},
		}
		_, content := renderNode(t, pn)
		if strings.Contains(content, " gs\n") {
			t.Errorf("opacity %g: unexpected gs operator:\n%s", op, content)
		}
	}
}

func TestInheritedBackgroundSuppression(t *testing.T) {
	bg := pdf.RGBHex(0x336699)
	line := func(s document.Style) layout.PlacedNode {
		return layout.PlacedNode{Node: &document.Text{Content: "x", TextStyle: s}, Size: document.Size{Width: 10, Height: 14}}
	}
	base := document.Style{FontSize: 12, Background: &bg}

	tests := []struct {
		name   string
		style  document.Style
		fills  int // number of background fills expected
		filler string
	}{
		{"opaque square repaints (unchanged output)", base, 2, " re\n"},
		{"transparent paints once", func() document.Style { s := base; s.Opacity = 0.5; return s }(), 1, " re\n"},
		{"rounded paints once", func() document.Style { s := base; s.BorderRadius = document.Pt(4); return s }(), 0, " re\n"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			pn := layout.PlacedNode{
				Node:     &document.Text{Content: "x", TextStyle: tc.style},
				Size:     document.Size{Width: 100, Height: 14},
				Children: []layout.PlacedNode{line(tc.style)},
			}
			_, content := renderNode(t, pn)
			if got := strings.Count(content, tc.filler); got != tc.fills {
				t.Errorf("%q count = %d, want %d:\n%s", tc.filler, got, tc.fills, content)
			}
		})
	}
}

func TestLinkAnnotations(t *testing.T) {
	linkStyle := document.Style{FontSize: 12, Link: "https://gpdf.dev/(x)"}
	textNode := &document.Text{Content: "para", TextStyle: linkStyle}
	pn := layout.PlacedNode{
		Node:     textNode,
		Position: document.Point{X: 10, Y: 20},
		Size:     document.Size{Width: 100, Height: 28},
		Children: []layout.PlacedNode{
			{Node: &document.Text{Content: "line 1", TextStyle: linkStyle}, Size: document.Size{Width: 40, Height: 14}},
			{Node: &document.Text{Content: "line 2", TextStyle: linkStyle}, Position: document.Point{Y: 14}, Size: document.Size{Width: 30, Height: 14}},
		},
	}
	r, buf := newTestRenderer(t)
	if err := r.BeginPage(document.Size{Width: 595, Height: 842}); err != nil {
		t.Fatal(err)
	}
	if err := r.renderPlacedNode(pn, 0, 0); err != nil {
		t.Fatal(err)
	}
	if len(r.pageAnnots) != 2 {
		t.Fatalf("annotations = %d, want 2 (one per line, none for the container)", len(r.pageAnnots))
	}
	rect := r.pageAnnots[1][pdf.Name("Rect")].(pdf.Rectangle)
	want := pdf.Rectangle{LLX: 10, LLY: 842 - 34 - 14, URX: 40, URY: 842 - 34}
	if rect != want {
		t.Errorf("second line rect = %+v, want %+v", rect, want)
	}
	if err := r.EndPage(); err != nil {
		t.Fatal(err)
	}
	if err := r.EndDocument(); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	for _, w := range []string{"/Annots [", "/Subtype /Link", "/S /URI", `(https://gpdf.dev/\(x\))`} {
		if !strings.Contains(out, w) {
			t.Errorf("output missing %q", w)
		}
	}
}

func TestNodeEmitsLink(t *testing.T) {
	tests := []struct {
		t    document.NodeType
		leaf bool
		want bool
	}{
		{document.NodeText, true, true},
		{document.NodeText, false, false},
		{document.NodeRichText, false, false},
		{document.NodeBox, false, true},
		{document.NodeImage, true, true},
	}
	for _, tc := range tests {
		if got := nodeEmitsLink(tc.t, tc.leaf); got != tc.want {
			t.Errorf("nodeEmitsLink(%v, %v) = %v, want %v", tc.t, tc.leaf, got, tc.want)
		}
	}
}

func TestFitRect(t *testing.T) {
	page := document.Rectangle{Width: 200, Height: 100}
	tests := []struct {
		mode document.ImageFitMode
		want document.Rectangle
	}{
		{document.FitStretch, page},
		{document.FitContain, document.Rectangle{X: 50, Y: 0, Width: 100, Height: 100}},
		{document.FitCover, document.Rectangle{X: 0, Y: -50, Width: 200, Height: 200}},
		{document.FitOriginal, document.Rectangle{X: 90, Y: 40, Width: 20, Height: 20}},
	}
	for _, tc := range tests {
		if got := fitRect(20, 20, page, tc.mode); got != tc.want {
			t.Errorf("fitRect(mode %d) = %+v, want %+v", tc.mode, got, tc.want)
		}
	}
	if got := fitRect(0, 0, page, document.FitCover); got != page {
		t.Errorf("fitRect with unknown size = %+v, want page", got)
	}
}

func TestPageBackground(t *testing.T) {
	bgColor := pdf.RGBHex(0xFFFAF0)
	pngData := effectsPNG(t, 4, 2)
	pages := []layout.PageLayout{
		{
			Size: document.Size{Width: 200, Height: 100},
			Background: &document.PageBackground{
				Color:   &bgColor,
				Image:   &document.ImageSource{Data: pngData, Format: document.ImagePNG, Width: 4, Height: 2},
				FitMode: document.FitCover,
				Opacity: 0.3,
			},
		},
	}
	r, buf := newTestRenderer(t)
	r.writer.SetCompression(false)
	if err := r.RenderDocument(pages, document.DocumentMetadata{}); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	for _, w := range []string{"/GS1 gs", "0 0 200 100 re", "W n", "Do", "/ExtGState", "/ca 0.3"} {
		if !strings.Contains(out, w) {
			t.Errorf("output missing %q", w)
		}
	}
}

func TestOverlayEffects(t *testing.T) {
	bg := pdf.RGBHex(0x336699)
	pngData := effectsPNG(t, 2, 2)
	nodes := []layout.PlacedNode{
		{
			Node:     &document.Text{Content: "WATERMARK", TextStyle: document.Style{FontSize: 40, Opacity: 0.25, Link: "https://gpdf.dev"}},
			Size:     document.Size{Width: 200, Height: 48},
			Children: []layout.PlacedNode{{Node: &document.Text{Content: "WATERMARK", TextStyle: document.Style{FontSize: 40, Opacity: 0.25, Link: "https://gpdf.dev"}}, Size: document.Size{Width: 200, Height: 48}}},
		},
		{
			Node: &document.Image{
				Source:   document.ImageSource{Data: pngData, Format: document.ImagePNG, Width: 2, Height: 2},
				ImgStyle: document.Style{BorderRadius: document.Pt(4), Background: &bg},
			},
			Position: document.Point{Y: 60},
			Size:     document.Size{Width: 20, Height: 20},
		},
	}
	r := NewOverlayRenderer(595, 842, nil, nil)
	res, err := r.RenderOverlay(nodes)
	if err != nil {
		t.Fatal(err)
	}
	content := string(res.Content)
	for _, w := range []string{"/OvGS1 gs", "W n", " c\n"} {
		if !strings.Contains(content, w) {
			t.Errorf("overlay content missing %q:\n%s", w, content)
		}
	}
	if len(res.ExtGStates) != 1 {
		t.Errorf("ExtGStates = %d, want 1", len(res.ExtGStates))
	}
	if len(res.Annots) != 1 {
		t.Errorf("Annots = %d, want 1", len(res.Annots))
	}
	if _, ok := res.Resources[pdf.Name("ExtGState")]; !ok {
		t.Error("overlay resources missing ExtGState")
	}
}

const testSVG = `<svg xmlns="http://www.w3.org/2000/svg" width="40" height="20" viewBox="0 0 40 20">` +
	`<rect width="40" height="20" rx="4" fill="#1A237E"/><circle cx="20" cy="10" r="6" fill="white" fill-opacity="0.5"/></svg>`

func TestSVGImageForm(t *testing.T) {
	src := document.ImageSource{Data: []byte(testSVG), Format: document.ImageSVG, Width: 40, Height: 20}
	r, buf := newTestRenderer(t)
	r.writer.SetCompression(false)
	pages := []layout.PageLayout{{
		Size: document.Size{Width: 200, Height: 100},
		Children: []layout.PlacedNode{
			{Node: &document.Image{Source: src}, Size: document.Size{Width: 80, Height: 40}},
			{Node: &document.Image{Source: src}, Position: document.Point{Y: 50}, Size: document.Size{Width: 40, Height: 20}},
		},
	}}
	if err := r.RenderDocument(pages, document.DocumentMetadata{}); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	for _, w := range []string{"/Subtype /Form", "/BBox [0 0 1 1]", "/Fm1 Do", "/ExtGState", "/SG1"} {
		if !strings.Contains(out, w) {
			t.Errorf("output missing %q", w)
		}
	}
	if n := strings.Count(out, "/Subtype /Form"); n != 1 {
		t.Errorf("form XObjects = %d, want 1 (deduplicated)", n)
	}
}

func TestSVGImageInvalid(t *testing.T) {
	r, _ := newTestRenderer(t)
	if err := r.BeginPage(document.Size{Width: 100, Height: 100}); err != nil {
		t.Fatal(err)
	}
	err := r.RenderImage(document.ImageSource{Data: []byte("<html/>"), Format: document.ImageSVG}, document.Point{}, document.Size{Width: 10, Height: 10})
	if err == nil {
		t.Error("expected error for invalid SVG")
	}
}

func TestOverlaySVG(t *testing.T) {
	r := NewOverlayRenderer(595, 842, nil, nil)
	nodes := []layout.PlacedNode{
		{Node: &document.Image{Source: document.ImageSource{Data: []byte(testSVG), Format: document.ImageSVG, Width: 40, Height: 20}}, Size: document.Size{Width: 40, Height: 20}},
		{Node: &document.Image{Source: document.ImageSource{Data: []byte("<bad"), Format: document.ImageSVG, Width: 4, Height: 2}}, Size: document.Size{Width: 4, Height: 2}},
	}
	res, err := r.RenderOverlay(nodes)
	if err != nil {
		t.Fatal(err)
	}
	if n := strings.Count(string(res.Content), " Do\n"); n != 1 {
		t.Errorf("Do operators = %d, want 1 (invalid SVG skipped)", n)
	}
	var forms int
	for _, io := range res.ImageObjects {
		if io.IsForm {
			forms++
		}
	}
	if forms != 1 {
		t.Errorf("form objects = %d, want 1", forms)
	}
}
