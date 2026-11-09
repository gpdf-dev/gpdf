package template

import (
	"bytes"
	"encoding/base64"
	"image"
	"image/color"
	"image/png"
	"strings"
	"testing"

	"github.com/gpdf-dev/gpdf/document"
	"github.com/gpdf-dev/gpdf/pdf"
)

func effectsPNG(t *testing.T) []byte {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, 8, 4))
	for y := 0; y < 4; y++ {
		for x := 0; x < 8; x++ {
			img.Set(x, y, color.RGBA{B: 200, A: 255})
		}
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

// pageDict returns the page dictionary of the given page index.
func pageDict(t *testing.T, data []byte, index int) (*pdf.Reader, pdf.Dict) {
	t.Helper()
	r, err := pdf.NewReader(data)
	if err != nil {
		t.Fatalf("NewReader: %v", err)
	}
	info, err := r.Page(index)
	if err != nil {
		t.Fatalf("Page(%d): %v", index, err)
	}
	d, err := r.ResolveDict(info.Ref)
	if err != nil {
		t.Fatalf("ResolveDict: %v", err)
	}
	return r, d
}

func annotCount(t *testing.T, r *pdf.Reader, page pdf.Dict) int {
	t.Helper()
	obj, ok := page[pdf.Name("Annots")]
	if !ok {
		return 0
	}
	resolved, err := r.Resolve(obj)
	if err != nil {
		t.Fatalf("resolve Annots: %v", err)
	}
	arr, ok := resolved.(pdf.Array)
	if !ok {
		t.Fatalf("Annots is %T, want Array", resolved)
	}
	return len(arr)
}

func TestEffectOptions(t *testing.T) {
	var s document.Style
	for _, opt := range []TextOption{WithTextRadius(document.Pt(3)), WithTextOpacity(0.4), WithTextLink("https://a")} {
		opt(&s)
	}
	if s.BorderRadius != document.Pt(3) || s.Opacity != 0.4 || s.Link != "https://a" {
		t.Errorf("text options not applied: %+v", s)
	}

	var ic imageConfig
	for _, opt := range []ImageOption{WithImageRadius(document.Pt(4)), WithImageOpacity(0.5), WithImageLink("https://b")} {
		opt(&ic)
	}
	if ic.radius != document.Pt(4) || ic.opacity != 0.5 || ic.link != "https://b" {
		t.Errorf("image options not applied: %+v", ic)
	}

	var bc boxConfig
	for _, opt := range []BoxOption{WithBoxRadius(document.Pt(5)), WithBoxOpacity(0.6), WithBoxLink("https://c")} {
		opt(&bc)
	}
	if bc.radius != document.Pt(5) || bc.opacity != 0.6 || bc.link != "https://c" {
		t.Errorf("box options not applied: %+v", bc)
	}
}

func TestColBuilderLink(t *testing.T) {
	c := &ColBuilder{}
	c.Link("gpdf", "https://gpdf.dev", TextColor(pdf.Red))
	txt := c.nodes[0].(*document.Text)
	if txt.TextStyle.Link != "https://gpdf.dev" {
		t.Errorf("Link = %q", txt.TextStyle.Link)
	}
	if txt.TextStyle.TextDecoration&document.DecorationUnderline == 0 {
		t.Error("link should be underlined by default")
	}
	if txt.TextStyle.Color != pdf.Red {
		t.Error("opts should override the default link color")
	}
}

func TestGenerateWithLinksRadiusOpacity(t *testing.T) {
	img := effectsPNG(t)
	doc := New()
	page := doc.AddPage()
	page.AutoRow(func(r *RowBuilder) {
		r.Col(12, func(c *ColBuilder) {
			c.Link("gpdf.dev", "https://gpdf.dev")
			c.RichText(func(rt *RichTextBuilder) {
				rt.Span("Read the ")
				rt.Span("docs", WithTextLink("https://gpdf.dev/docs"), Underline())
			})
			c.Box(func(c *ColBuilder) {
				c.Text("card")
			}, WithBoxRadius(document.Pt(6)), WithBoxBackground(pdf.Gray(0.9)), WithBoxOpacity(0.5), WithBoxLink("https://example.com"))
			c.Image(img, FitWidth(document.Mm(20)), WithImageRadius(document.Mm(2)), WithImageLink("https://example.org"))
		})
	})
	data, err := doc.Generate()
	if err != nil {
		t.Fatal(err)
	}
	r, pd := pageDict(t, data, 0)
	if got := annotCount(t, r, pd); got != 4 {
		t.Errorf("annotations = %d, want 4 (link, rich-text span, box, image)", got)
	}
	res, err := r.ResolveDict(pd[pdf.Name("Resources")])
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := res[pdf.Name("ExtGState")]; !ok {
		t.Error("page resources missing ExtGState")
	}
}

func TestBackgroundOptions(t *testing.T) {
	img := effectsPNG(t)
	cfg := Config{}
	WithBackgroundColor(pdf.White)(&cfg)
	WithBackgroundImage(img, BackgroundOpacity(0.2))(&cfg)
	bg := cfg.Background
	if bg == nil || bg.Color == nil || bg.Image == nil {
		t.Fatalf("background not configured: %+v", bg)
	}
	if bg.FitMode != document.FitCover {
		t.Errorf("default fit = %d, want FitCover", bg.FitMode)
	}
	if bg.Opacity != 0.2 || bg.Image.Width != 8 || bg.Image.Height != 4 {
		t.Errorf("unexpected background: %+v / %dx%d", bg, bg.Image.Width, bg.Image.Height)
	}

	// Options build a fresh copy; a previously captured background is untouched.
	before := cfg.Background
	BackgroundFit(document.FitStretch)(cfg.ensureBackground())
	if before.FitMode != document.FitCover {
		t.Error("ensureBackground mutated the previous background")
	}
}

func TestBackgroundOnEveryPage(t *testing.T) {
	doc := New(WithBackgroundImage(effectsPNG(t)))
	page := doc.AddPage()
	for i := 0; i < 120; i++ {
		page.AutoRow(func(r *RowBuilder) {
			r.Col(12, func(c *ColBuilder) { c.Text("row") })
		})
	}
	data, err := doc.Generate()
	if err != nil {
		t.Fatal(err)
	}
	r, err := pdf.NewReader(data)
	if err != nil {
		t.Fatal(err)
	}
	n, _ := r.PageCount()
	if n < 2 {
		t.Fatalf("expected overflow pages, got %d", n)
	}
	for i := 0; i < n; i++ {
		_, pd := pageDict(t, data, i)
		res, err := r.ResolveDict(pd[pdf.Name("Resources")])
		if err != nil {
			t.Fatal(err)
		}
		if _, ok := res[pdf.Name("XObject")]; !ok {
			t.Errorf("page %d has no XObject (background image) resource", i)
		}
	}
}

func TestOverlayLinkAndOpacity(t *testing.T) {
	data := generateTestPDF(t, 1)
	doc, err := OpenExisting(data)
	if err != nil {
		t.Fatal(err)
	}
	err = doc.Overlay(0, func(p *PageBuilder) {
		p.Absolute(document.Mm(20), document.Mm(100), func(c *ColBuilder) {
			c.Text("CONFIDENTIAL", FontSize(48), WithTextOpacity(0.2))
			c.Link("verify", "https://gpdf.dev/verify")
		})
	})
	if err != nil {
		t.Fatal(err)
	}
	out, err := doc.Save()
	if err != nil {
		t.Fatal(err)
	}
	r, pd := pageDict(t, out, 0)
	if got := annotCount(t, r, pd); got != 1 {
		t.Errorf("annotations = %d, want 1", got)
	}
	contents, ok := pd[pdf.Name("Contents")].(pdf.Array)
	if !ok || len(contents) < 3 {
		t.Errorf("overlay content lost after adding annotations: %v", pd[pdf.Name("Contents")])
	}
	res, err := r.ResolveDict(pd[pdf.Name("Resources")])
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := res[pdf.Name("ExtGState")]; !ok {
		t.Error("overlay resources missing ExtGState")
	}
}

func TestSchemaEffects(t *testing.T) {
	img := base64.StdEncoding.EncodeToString(effectsPNG(t))
	schema := []byte(`{
		"page": {"size": "A4", "background": {"color": "#FFFAF0", "image": "` + img + `", "fit": "contain", "opacity": 0.3}},
		"body": [{"row": {"cols": [{"span": 12, "elements": [
			{"type": "link", "content": "gpdf.dev", "style": {"link": "https://gpdf.dev"}},
			{"type": "text", "content": "pill", "style": {"background": "#EEEEEE", "borderRadius": "4pt", "opacity": 0.5}},
			{"type": "image", "image": {"src": "` + img + `", "borderRadius": "2pt", "opacity": 0.7, "link": "https://example.org"}},
			{"type": "box", "box": {"background": "#EEEEEE", "borderRadius": "6pt", "opacity": 0.8, "link": "https://example.com",
				"elements": [{"type": "text", "content": "card"}]}}
		]}]}}]
	}`)
	doc, err := FromJSON(schema, nil)
	if err != nil {
		t.Fatal(err)
	}
	bg := doc.config.Background
	if bg == nil || bg.Color == nil || bg.Image == nil || bg.FitMode != document.FitContain || bg.Opacity != 0.3 {
		t.Fatalf("schema background not applied: %+v", bg)
	}
	data, err := doc.Generate()
	if err != nil {
		t.Fatal(err)
	}
	r, pd := pageDict(t, data, 0)
	if got := annotCount(t, r, pd); got != 3 {
		t.Errorf("annotations = %d, want 3", got)
	}
}

func TestSchemaBackgroundErrors(t *testing.T) {
	for _, bg := range []string{
		`{"color": "not-a-color"}`,
		`{"image": "/no/such/file.png"}`,
		`{"color": "#FFFFFF", "fit": "sideways"}`,
	} {
		_, err := FromJSON([]byte(`{"page": {"background": `+bg+`}, "body": []}`), nil)
		if err == nil || !strings.Contains(err.Error(), "page background") {
			t.Errorf("background %s: err = %v, want page background error", bg, err)
		}
	}

	doc, err := FromJSON([]byte(`{"page": {"background": {"color": "#000000", "opacity": 0.5}}, "body": []}`), nil)
	if err != nil {
		t.Fatal(err)
	}
	if doc.config.Background.Opacity != 0.5 {
		t.Errorf("color-only background opacity = %g, want 0.5", doc.config.Background.Opacity)
	}
}

const templateSVG = `<?xml version="1.0"?>
<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 30 15"><rect width="30" height="15" fill="teal"/></svg>`

func TestSVGDetectionAndSize(t *testing.T) {
	if f := detectImageFormat([]byte("\xEF\xBB\xBF  " + templateSVG)); f != document.ImageSVG {
		t.Errorf("format = %d, want ImageSVG", f)
	}
	if f := detectImageFormat([]byte("<html>no svg</html>")); f != document.ImagePNG {
		t.Errorf("non-SVG markup format = %d, want PNG fallback", f)
	}
	if w, h := extractImageDimensions([]byte(templateSVG), document.ImageSVG); w != 30 || h != 15 {
		t.Errorf("size = %dx%d, want 30x15", w, h)
	}
	if w, h := extractImageDimensions([]byte("<svg"), document.ImageSVG); w != 0 || h != 0 {
		t.Errorf("invalid SVG size = %dx%d, want 0x0", w, h)
	}
}

func TestSVGImageGenerateAndOverlay(t *testing.T) {
	doc := New(WithBackgroundImage([]byte(templateSVG), BackgroundFit(document.FitContain)))
	page := doc.AddPage()
	page.AutoRow(func(r *RowBuilder) {
		r.Col(12, func(c *ColBuilder) {
			c.Image([]byte(templateSVG), FitWidth(document.Mm(40)))
			c.Image([]byte("<svg broken"), FitWidth(document.Mm(40)))
		})
	})
	data, err := doc.Generate()
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(data, []byte("/Subtype /Form")) {
		t.Error("SVG not embedded as a Form XObject")
	}

	ex, err := OpenExisting(data)
	if err != nil {
		t.Fatal(err)
	}
	if err := ex.Overlay(0, func(p *PageBuilder) {
		p.Absolute(document.Mm(10), document.Mm(10), func(c *ColBuilder) {
			c.Image([]byte(templateSVG), FitWidth(document.Mm(20)))
		})
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := ex.Save(); err != nil {
		t.Fatal(err)
	}
}

func TestSchemaSVGSources(t *testing.T) {
	for _, src := range []string{
		`"<svg viewBox='0 0 30 15'><rect width='30' height='15'/></svg>"`,
		`"data:image/svg+xml;utf8,%3Csvg viewBox='0 0 30 15'%3E%3Crect width='30' height='15'/%3E%3C/svg%3E"`,
		`"data:image/svg+xml;base64,` + base64.StdEncoding.EncodeToString([]byte(templateSVG)) + `"`,
	} {
		schema := `{"page": {"size": "A4"}, "body": [{"row": {"cols": [{"span": 12, "image": {"src": ` + src + `, "width": "30mm"}}]}}]}`
		doc, err := FromJSON([]byte(schema), nil)
		if err != nil {
			t.Fatalf("%s: %v", src, err)
		}
		data, err := doc.Generate()
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Contains(data, []byte("/Subtype /Form")) {
			t.Errorf("%s: SVG not embedded", src)
		}
	}
	if _, err := decodeBase64Image("data:image/svg+xml,%zz"); err == nil {
		t.Error("expected error for bad URL escape")
	}
}
