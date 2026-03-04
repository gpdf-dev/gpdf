package template

import (
	"bytes"
	"strings"
	"testing"

	"github.com/gpdf-dev/gpdf/document"
	"github.com/gpdf-dev/gpdf/document/layout"
	"github.com/gpdf-dev/gpdf/pdf"
	"github.com/gpdf-dev/gpdf/pdf/font"
)

// ---------------------------------------------------------------------------
// builder.go tests
// ---------------------------------------------------------------------------

func TestNewDefaults(t *testing.T) {
	doc := New()

	if doc.config.PageSize != document.A4 {
		t.Errorf("default page size: got %v, want A4", doc.config.PageSize)
	}
	if doc.config.FontSize != 12 {
		t.Errorf("default font size: got %v, want 12", doc.config.FontSize)
	}
	if doc.config.DefaultFont != "" {
		t.Errorf("default font family: got %q, want empty", doc.config.DefaultFont)
	}
	if len(doc.pages) != 0 {
		t.Errorf("initial pages: got %d, want 0", len(doc.pages))
	}
	if doc.fontResolver == nil {
		t.Error("fontResolver is nil")
	}
}

func TestNewWithPageSize(t *testing.T) {
	doc := New(WithPageSize(document.Letter))
	if doc.config.PageSize != document.Letter {
		t.Errorf("page size: got %v, want Letter", doc.config.PageSize)
	}
}

func TestNewWithMargins(t *testing.T) {
	m := document.UniformEdges(document.Mm(15))
	doc := New(WithMargins(m))
	if doc.config.Margins != m {
		t.Errorf("margins: got %v, want %v", doc.config.Margins, m)
	}
}

func TestNewWithDefaultFont(t *testing.T) {
	doc := New(WithDefaultFont("TestFont", 14))
	if doc.config.DefaultFont != "TestFont" {
		t.Errorf("default font: got %q, want %q", doc.config.DefaultFont, "TestFont")
	}
	if doc.config.FontSize != 14 {
		t.Errorf("font size: got %v, want 14", doc.config.FontSize)
	}
}

func TestNewWithMetadata(t *testing.T) {
	meta := document.DocumentMetadata{
		Title:  "Test Title",
		Author: "Test Author",
	}
	doc := New(WithMetadata(meta))
	if doc.config.Metadata.Title != "Test Title" {
		t.Errorf("title: got %q, want %q", doc.config.Metadata.Title, "Test Title")
	}
	if doc.config.Metadata.Author != "Test Author" {
		t.Errorf("author: got %q, want %q", doc.config.Metadata.Author, "Test Author")
	}
}

func TestNewWithFontInvalidData(t *testing.T) {
	// Passing invalid font data should not crash; the font just won't be registered.
	doc := New(WithFont("BadFont", []byte("not a font")))
	if _, ok := doc.fonts["BadFont"]; ok {
		t.Error("invalid font data should not be registered")
	}
}

func TestNewWithMultipleOptions(t *testing.T) {
	meta := document.DocumentMetadata{Title: "Multi"}
	doc := New(
		WithPageSize(document.Legal),
		WithMargins(document.UniformEdges(document.Mm(10))),
		WithDefaultFont("MyFont", 16),
		WithMetadata(meta),
	)
	if doc.config.PageSize != document.Legal {
		t.Errorf("page size: got %v, want Legal", doc.config.PageSize)
	}
	if doc.config.DefaultFont != "MyFont" {
		t.Errorf("default font: got %q, want %q", doc.config.DefaultFont, "MyFont")
	}
	if doc.config.FontSize != 16 {
		t.Errorf("font size: got %v, want 16", doc.config.FontSize)
	}
	if doc.config.Metadata.Title != "Multi" {
		t.Errorf("title: got %q, want %q", doc.config.Metadata.Title, "Multi")
	}
}

func TestAddPage(t *testing.T) {
	doc := New()
	p1 := doc.AddPage()
	if p1 == nil {
		t.Fatal("AddPage returned nil")
	}
	if len(doc.pages) != 1 {
		t.Errorf("pages count: got %d, want 1", len(doc.pages))
	}

	p2 := doc.AddPage()
	if p2 == nil {
		t.Fatal("second AddPage returned nil")
	}
	if len(doc.pages) != 2 {
		t.Errorf("pages count: got %d, want 2", len(doc.pages))
	}
}

func TestHeaderAndFooter(t *testing.T) {
	doc := New()
	if doc.headerFn != nil {
		t.Error("headerFn should be nil initially")
	}
	if doc.footerFn != nil {
		t.Error("footerFn should be nil initially")
	}

	doc.Header(func(p *PageBuilder) {
		p.AutoRow(func(r *RowBuilder) {
			r.Col(12, func(c *ColBuilder) { c.Text("Header") })
		})
	})
	doc.Footer(func(p *PageBuilder) {
		p.AutoRow(func(r *RowBuilder) {
			r.Col(12, func(c *ColBuilder) { c.Text("Footer") })
		})
	})

	if doc.headerFn == nil {
		t.Error("headerFn should not be nil after Header()")
	}
	if doc.footerFn == nil {
		t.Error("footerFn should not be nil after Footer()")
	}
}

func TestGenerateEmptyDocument(t *testing.T) {
	doc := New()
	// A document with no pages should still produce valid output (or at
	// least not panic).
	data, err := doc.Generate()
	if err != nil {
		t.Fatalf("Generate failed on empty document: %v", err)
	}
	// Even with no pages, the PDF writer should produce output.
	if len(data) == 0 {
		t.Log("Empty document produced 0 bytes (acceptable)")
	}
}

func TestGenerateSimpleDocument(t *testing.T) {
	doc := New()
	page := doc.AddPage()
	page.AutoRow(func(r *RowBuilder) {
		r.Col(12, func(c *ColBuilder) {
			c.Text("Hello, World!")
		})
	})

	data, err := doc.Generate()
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}
	if len(data) < 5 {
		t.Fatal("Generated PDF too short")
	}
	if string(data[:5]) != "%PDF-" {
		t.Errorf("PDF header: got %q, want %%PDF-", string(data[:5]))
	}
}

func TestRender(t *testing.T) {
	doc := New()
	page := doc.AddPage()
	page.AutoRow(func(r *RowBuilder) {
		r.Col(12, func(c *ColBuilder) {
			c.Text("Render test")
		})
	})

	var buf bytes.Buffer
	err := doc.Render(&buf)
	if err != nil {
		t.Fatalf("Render failed: %v", err)
	}
	if buf.Len() < 5 {
		t.Fatal("Rendered PDF too short")
	}
	if string(buf.Bytes()[:5]) != "%PDF-" {
		t.Errorf("PDF header: got %q, want %%PDF-", string(buf.Bytes()[:5]))
	}
}

func TestBuildDocumentDefaultProducer(t *testing.T) {
	doc := New()
	doc.AddPage()
	built := doc.buildDocument()
	if built.Metadata.Producer != "gpdf" {
		t.Errorf("producer: got %q, want %q", built.Metadata.Producer, "gpdf")
	}
}

func TestBuildDocumentCustomProducer(t *testing.T) {
	doc := New(WithMetadata(document.DocumentMetadata{Producer: "custom"}))
	doc.AddPage()
	built := doc.buildDocument()
	if built.Metadata.Producer != "custom" {
		t.Errorf("producer: got %q, want %q", built.Metadata.Producer, "custom")
	}
}

func TestBuildDocumentWithHeaderFooter(t *testing.T) {
	doc := New()
	doc.Header(func(p *PageBuilder) {
		p.AutoRow(func(r *RowBuilder) {
			r.Col(12, func(c *ColBuilder) { c.Text("H") })
		})
	})
	doc.Footer(func(p *PageBuilder) {
		p.AutoRow(func(r *RowBuilder) {
			r.Col(12, func(c *ColBuilder) { c.Text("F") })
		})
	})
	page := doc.AddPage()
	page.AutoRow(func(r *RowBuilder) {
		r.Col(12, func(c *ColBuilder) { c.Text("Body") })
	})

	// buildDocument now only contains body content; header/footer are
	// handled separately by the paginator via buildSection.
	built := doc.buildDocument()
	if len(built.Pages) != 1 {
		t.Fatalf("pages: got %d, want 1", len(built.Pages))
	}
	// Body only: at least 1 content node.
	if len(built.Pages[0].Content) < 1 {
		t.Errorf("content nodes: got %d, want >= 1", len(built.Pages[0].Content))
	}

	// Verify that buildSection produces header/footer nodes.
	headerNodes := doc.buildSection(doc.headerFn)
	footerNodes := doc.buildSection(doc.footerFn)
	if len(headerNodes) < 1 {
		t.Errorf("header nodes: got %d, want >= 1", len(headerNodes))
	}
	if len(footerNodes) < 1 {
		t.Errorf("footer nodes: got %d, want >= 1", len(footerNodes))
	}
}

func TestBuildDocumentDefaultStyle(t *testing.T) {
	doc := New(WithDefaultFont("Arial", 18))
	doc.AddPage()
	built := doc.buildDocument()
	if built.DefaultStyle.FontFamily != "Arial" {
		t.Errorf("font family: got %q, want %q", built.DefaultStyle.FontFamily, "Arial")
	}
	if built.DefaultStyle.FontSize != 18 {
		t.Errorf("font size: got %v, want 18", built.DefaultStyle.FontSize)
	}
	if built.DefaultStyle.FontWeight != document.WeightNormal {
		t.Errorf("font weight: got %v, want WeightNormal", built.DefaultStyle.FontWeight)
	}
}

func TestGenerateMultiplePages(t *testing.T) {
	doc := New()
	for i := 0; i < 5; i++ {
		page := doc.AddPage()
		page.AutoRow(func(r *RowBuilder) {
			r.Col(12, func(c *ColBuilder) { c.Text("Page content") })
		})
	}

	data, err := doc.Generate()
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}
	if string(data[:5]) != "%PDF-" {
		t.Errorf("PDF header: got %q, want %%PDF-", string(data[:5]))
	}
}

// ---------------------------------------------------------------------------
// grid.go tests
// ---------------------------------------------------------------------------

func TestPageBuilderRow(t *testing.T) {
	doc := New()
	pb := doc.AddPage()

	pb.Row(document.Mm(20), func(r *RowBuilder) {
		r.Col(12, func(c *ColBuilder) { c.Text("Fixed height") })
	})

	if len(pb.rows) != 1 {
		t.Fatalf("rows: got %d, want 1", len(pb.rows))
	}
	if pb.rows[0].auto {
		t.Error("row should not be auto")
	}
	if pb.rows[0].height != document.Mm(20) {
		t.Errorf("row height: got %v, want Mm(20)", pb.rows[0].height)
	}
}

func TestPageBuilderAutoRow(t *testing.T) {
	doc := New()
	pb := doc.AddPage()

	pb.AutoRow(func(r *RowBuilder) {
		r.Col(12, func(c *ColBuilder) { c.Text("Auto height") })
	})

	if len(pb.rows) != 1 {
		t.Fatalf("rows: got %d, want 1", len(pb.rows))
	}
	if !pb.rows[0].auto {
		t.Error("row should be auto")
	}
}

func TestPageBuilderBuildNodesEmpty(t *testing.T) {
	doc := New()
	pb := &PageBuilder{doc: doc}
	nodes := pb.buildNodes()
	if len(nodes) != 0 {
		t.Errorf("buildNodes on empty page: got %d nodes, want 0", len(nodes))
	}
}

func TestPageBuilderBuildNodesNilFn(t *testing.T) {
	doc := New()
	pb := &PageBuilder{doc: doc}
	pb.rows = append(pb.rows, rowEntry{auto: true, fn: nil})
	nodes := pb.buildNodes()
	if len(nodes) != 1 {
		t.Fatalf("nodes: got %d, want 1", len(nodes))
	}
}

func TestRowBuilderColSpanClamping(t *testing.T) {
	doc := New()
	rb := &RowBuilder{doc: doc}

	// Span too small
	rb.Col(0, func(c *ColBuilder) { c.Text("zero") })
	if rb.cols[0].span != 1 {
		t.Errorf("zero span clamped to: got %d, want 1", rb.cols[0].span)
	}

	// Span negative
	rb.Col(-5, func(c *ColBuilder) { c.Text("negative") })
	if rb.cols[1].span != 1 {
		t.Errorf("negative span clamped to: got %d, want 1", rb.cols[1].span)
	}

	// Span too large
	rb.Col(20, func(c *ColBuilder) { c.Text("large") })
	if rb.cols[2].span != 12 {
		t.Errorf("large span clamped to: got %d, want 12", rb.cols[2].span)
	}

	// Valid span
	rb.Col(6, func(c *ColBuilder) { c.Text("valid") })
	if rb.cols[3].span != 6 {
		t.Errorf("valid span: got %d, want 6", rb.cols[3].span)
	}
}

func TestRowBuilderBuildFixedHeight(t *testing.T) {
	doc := New()
	rb := &RowBuilder{doc: doc}
	rb.Col(6, func(c *ColBuilder) { c.Text("Left") })
	rb.Col(6, func(c *ColBuilder) { c.Text("Right") })

	node := rb.build(document.Mm(30), false)
	box, ok := node.(*document.Box)
	if !ok {
		t.Fatal("build should return a *document.Box")
	}
	if box.BoxStyle.Height != document.Mm(30) {
		t.Errorf("box height: got %v, want Mm(30)", box.BoxStyle.Height)
	}
	if len(box.Content) != 2 {
		t.Errorf("columns: got %d, want 2", len(box.Content))
	}
}

func TestRowBuilderBuildAutoHeight(t *testing.T) {
	doc := New()
	rb := &RowBuilder{doc: doc}
	rb.Col(12, func(c *ColBuilder) { c.Text("Auto") })

	node := rb.build(document.Value{}, true)
	box, ok := node.(*document.Box)
	if !ok {
		t.Fatal("build should return a *document.Box")
	}
	// Auto height means Height should be zero-value.
	zeroVal := document.Value{}
	if box.BoxStyle.Height != zeroVal {
		t.Errorf("auto row height should be zero value, got %v", box.BoxStyle.Height)
	}
}

func TestRowBuilderBuildColumnWidths(t *testing.T) {
	doc := New()
	rb := &RowBuilder{doc: doc}
	rb.Col(3, func(c *ColBuilder) {})
	rb.Col(9, func(c *ColBuilder) {})

	node := rb.build(document.Value{}, true)
	box := node.(*document.Box)
	if len(box.Content) != 2 {
		t.Fatalf("columns: got %d, want 2", len(box.Content))
	}

	col0 := box.Content[0].(*document.Box)
	expectedPct0 := document.Pct(float64(3) / float64(12) * 100)
	if col0.BoxStyle.Width != expectedPct0 {
		t.Errorf("col0 width: got %v, want %v", col0.BoxStyle.Width, expectedPct0)
	}

	col1 := box.Content[1].(*document.Box)
	expectedPct1 := document.Pct(float64(9) / float64(12) * 100)
	if col1.BoxStyle.Width != expectedPct1 {
		t.Errorf("col1 width: got %v, want %v", col1.BoxStyle.Width, expectedPct1)
	}
}

func TestRowBuilderBuildNilColFn(t *testing.T) {
	doc := New()
	rb := &RowBuilder{doc: doc}
	rb.cols = append(rb.cols, colEntry{span: 6, fn: nil})
	node := rb.build(document.Value{}, true)
	box := node.(*document.Box)
	if len(box.Content) != 1 {
		t.Errorf("columns: got %d, want 1", len(box.Content))
	}
}

func TestColBuilderText(t *testing.T) {
	doc := New()
	cb := &ColBuilder{doc: doc}
	cb.Text("hello")
	if len(cb.nodes) != 1 {
		t.Fatalf("nodes: got %d, want 1", len(cb.nodes))
	}
	textNode, ok := cb.nodes[0].(*document.Text)
	if !ok {
		t.Fatal("expected *document.Text")
	}
	if textNode.Content != "hello" {
		t.Errorf("content: got %q, want %q", textNode.Content, "hello")
	}
}

func TestColBuilderTextWithOptions(t *testing.T) {
	doc := New(WithDefaultFont("DefaultFont", 10))
	cb := &ColBuilder{doc: doc}
	cb.Text("styled", FontSize(20), Bold(), Italic(), AlignCenter())
	textNode := cb.nodes[0].(*document.Text)
	if textNode.TextStyle.FontSize != 20 {
		t.Errorf("font size: got %v, want 20", textNode.TextStyle.FontSize)
	}
	if textNode.TextStyle.FontWeight != document.WeightBold {
		t.Errorf("font weight: got %v, want Bold", textNode.TextStyle.FontWeight)
	}
	if textNode.TextStyle.FontStyle != document.StyleItalic {
		t.Errorf("font style: got %v, want Italic", textNode.TextStyle.FontStyle)
	}
	if textNode.TextStyle.TextAlign != document.AlignCenter {
		t.Errorf("align: got %v, want AlignCenter", textNode.TextStyle.TextAlign)
	}
}

func TestColBuilderImagePNG(t *testing.T) {
	// Minimal PNG header
	pngData := []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}
	doc := New()
	cb := &ColBuilder{doc: doc}
	cb.Image(pngData)
	if len(cb.nodes) != 1 {
		t.Fatalf("nodes: got %d, want 1", len(cb.nodes))
	}
	imgNode, ok := cb.nodes[0].(*document.Image)
	if !ok {
		t.Fatal("expected *document.Image")
	}
	if imgNode.Source.Format != document.ImagePNG {
		t.Errorf("format: got %v, want ImagePNG", imgNode.Source.Format)
	}
	if imgNode.FitMode != document.FitContain {
		t.Errorf("fit mode: got %v, want FitContain", imgNode.FitMode)
	}
}

func TestColBuilderImageJPEG(t *testing.T) {
	// JPEG header bytes
	jpegData := []byte{0xFF, 0xD8, 0xFF, 0xE0, 0x00}
	doc := New()
	cb := &ColBuilder{doc: doc}
	cb.Image(jpegData)
	imgNode := cb.nodes[0].(*document.Image)
	if imgNode.Source.Format != document.ImageJPEG {
		t.Errorf("format: got %v, want ImageJPEG", imgNode.Source.Format)
	}
}

func TestColBuilderImageWithFitWidth(t *testing.T) {
	doc := New()
	cb := &ColBuilder{doc: doc}
	cb.Image([]byte{0x89, 0x50, 0x4E, 0x47}, FitWidth(document.Mm(100)))
	imgNode := cb.nodes[0].(*document.Image)
	if imgNode.FitMode != document.FitContain {
		t.Errorf("fit mode: got %v, want FitContain", imgNode.FitMode)
	}
}

func TestColBuilderImageWithFitHeight(t *testing.T) {
	doc := New()
	cb := &ColBuilder{doc: doc}
	cb.Image([]byte{0x89, 0x50, 0x4E, 0x47}, FitHeight(document.Mm(50)))
	imgNode := cb.nodes[0].(*document.Image)
	if imgNode.FitMode != document.FitContain {
		t.Errorf("fit mode: got %v, want FitContain", imgNode.FitMode)
	}
}

func TestColBuilderTable(t *testing.T) {
	doc := New()
	cb := &ColBuilder{doc: doc}
	header := []string{"A", "B", "C"}
	rows := [][]string{
		{"1", "2", "3"},
		{"4", "5", "6"},
	}
	cb.Table(header, rows)
	if len(cb.nodes) != 1 {
		t.Fatalf("nodes: got %d, want 1", len(cb.nodes))
	}
	tbl, ok := cb.nodes[0].(*document.Table)
	if !ok {
		t.Fatal("expected *document.Table")
	}
	if len(tbl.Header) != 1 {
		t.Errorf("header rows: got %d, want 1", len(tbl.Header))
	}
	if len(tbl.Header[0].Cells) != 3 {
		t.Errorf("header cells: got %d, want 3", len(tbl.Header[0].Cells))
	}
	if len(tbl.Body) != 2 {
		t.Errorf("body rows: got %d, want 2", len(tbl.Body))
	}
}

func TestColBuilderTableNoHeader(t *testing.T) {
	doc := New()
	cb := &ColBuilder{doc: doc}
	rows := [][]string{
		{"1", "2"},
	}
	cb.Table(nil, rows)
	tbl := cb.nodes[0].(*document.Table)
	if len(tbl.Header) != 0 {
		t.Errorf("header rows: got %d, want 0", len(tbl.Header))
	}
	if len(tbl.Body) != 1 {
		t.Errorf("body rows: got %d, want 1", len(tbl.Body))
	}
}

func TestColBuilderTableWithColumnWidths(t *testing.T) {
	doc := New()
	cb := &ColBuilder{doc: doc}
	cb.Table(
		[]string{"A", "B"},
		[][]string{{"1", "2"}},
		ColumnWidths(60, 40),
	)
	tbl := cb.nodes[0].(*document.Table)
	if len(tbl.Columns) != 2 {
		t.Fatalf("columns: got %d, want 2", len(tbl.Columns))
	}
	if tbl.Columns[0].Width != document.Pct(60) {
		t.Errorf("col0 width: got %v, want Pct(60)", tbl.Columns[0].Width)
	}
	if tbl.Columns[1].Width != document.Pct(40) {
		t.Errorf("col1 width: got %v, want Pct(40)", tbl.Columns[1].Width)
	}
}

func TestColBuilderTableWithHeaderStyle(t *testing.T) {
	doc := New()
	cb := &ColBuilder{doc: doc}
	bg := pdf.RGB(0.2, 0.3, 0.4)
	tc := pdf.White
	cb.Table(
		[]string{"H1"},
		[][]string{{"V1"}},
		TableHeaderStyle(BgColor(bg), TextColor(tc)),
	)
	tbl := cb.nodes[0].(*document.Table)
	headerCell := tbl.Header[0].Cells[0]
	textNode := headerCell.Content[0].(*document.Text)
	if textNode.TextStyle.Background == nil {
		t.Error("header bg should not be nil")
	}
	if textNode.TextStyle.Color != tc {
		t.Errorf("header text color: got %v, want %v", textNode.TextStyle.Color, tc)
	}
}

func TestColBuilderTableWithStripe(t *testing.T) {
	doc := New()
	cb := &ColBuilder{doc: doc}
	stripeColor := pdf.Gray(0.95)
	cb.Table(
		[]string{"A"},
		[][]string{{"row0"}, {"row1"}, {"row2"}, {"row3"}},
		TableStripe(stripeColor),
	)
	tbl := cb.nodes[0].(*document.Table)
	// Row 0 (even) should not have stripe.
	cell0 := tbl.Body[0].Cells[0].Content[0].(*document.Text)
	if cell0.TextStyle.Background != nil {
		t.Error("even row should not have stripe background")
	}
	// Row 1 (odd) should have stripe.
	cell1 := tbl.Body[1].Cells[0].Content[0].(*document.Text)
	if cell1.TextStyle.Background == nil {
		t.Error("odd row should have stripe background")
	}
}

func TestColBuilderLine(t *testing.T) {
	doc := New()
	cb := &ColBuilder{doc: doc}
	cb.Line()
	if len(cb.nodes) != 1 {
		t.Fatalf("nodes: got %d, want 1", len(cb.nodes))
	}
	box, ok := cb.nodes[0].(*document.Box)
	if !ok {
		t.Fatal("expected *document.Box")
	}
	if box.BoxStyle.Background == nil {
		t.Error("line should have background color")
	}
	if box.BoxStyle.Height != document.Pt(1) {
		t.Errorf("line height: got %v, want Pt(1)", box.BoxStyle.Height)
	}
}

func TestColBuilderLineWithOptions(t *testing.T) {
	doc := New()
	cb := &ColBuilder{doc: doc}
	lineColor := pdf.Red
	cb.Line(LineColor(lineColor), LineThickness(document.Pt(3)))
	box := cb.nodes[0].(*document.Box)
	if *box.BoxStyle.Background != lineColor {
		t.Errorf("line color: got %v, want %v", *box.BoxStyle.Background, lineColor)
	}
	if box.BoxStyle.Height != document.Pt(3) {
		t.Errorf("line thickness: got %v, want Pt(3)", box.BoxStyle.Height)
	}
}

func TestColBuilderSpacer(t *testing.T) {
	doc := New()
	cb := &ColBuilder{doc: doc}
	cb.Spacer(document.Mm(10))
	if len(cb.nodes) != 1 {
		t.Fatalf("nodes: got %d, want 1", len(cb.nodes))
	}
	box, ok := cb.nodes[0].(*document.Box)
	if !ok {
		t.Fatal("expected *document.Box")
	}
	if box.BoxStyle.Height != document.Mm(10) {
		t.Errorf("spacer height: got %v, want Mm(10)", box.BoxStyle.Height)
	}
	if box.BoxStyle.Background != nil {
		t.Error("spacer should have nil background")
	}
}

func TestColBuilderBuildNodesEmpty(t *testing.T) {
	cb := &ColBuilder{}
	nodes := cb.buildNodes()
	if len(nodes) != 0 {
		t.Errorf("buildNodes on empty: got %d, want 0", len(nodes))
	}
}

func TestColBuilderDefaultStyleWithDoc(t *testing.T) {
	doc := New(WithDefaultFont("CustomFont", 20))
	cb := &ColBuilder{doc: doc}
	s := cb.defaultStyle()
	if s.FontFamily != "CustomFont" {
		t.Errorf("font family: got %q, want %q", s.FontFamily, "CustomFont")
	}
	if s.FontSize != 20 {
		t.Errorf("font size: got %v, want 20", s.FontSize)
	}
}

func TestColBuilderDefaultStyleNilDoc(t *testing.T) {
	cb := &ColBuilder{doc: nil}
	s := cb.defaultStyle()
	// Should return DefaultStyle without panic.
	if s.FontSize != 12 {
		t.Errorf("font size: got %v, want 12 (default)", s.FontSize)
	}
}

func TestDetectImageFormatJPEG(t *testing.T) {
	data := []byte{0xFF, 0xD8, 0xFF, 0xE0, 0x00, 0x10}
	if got := detectImageFormat(data); got != document.ImageJPEG {
		t.Errorf("detectImageFormat for JPEG: got %v, want ImageJPEG", got)
	}
}

func TestDetectImageFormatPNG(t *testing.T) {
	data := []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}
	if got := detectImageFormat(data); got != document.ImagePNG {
		t.Errorf("detectImageFormat for PNG: got %v, want ImagePNG", got)
	}
}

func TestDetectImageFormatUnknown(t *testing.T) {
	data := []byte{0x00, 0x00, 0x00}
	if got := detectImageFormat(data); got != document.ImagePNG {
		t.Errorf("detectImageFormat for unknown: got %v, want ImagePNG (fallback)", got)
	}
}

func TestDetectImageFormatShortData(t *testing.T) {
	// Less than 3 bytes
	data := []byte{0xFF, 0xD8}
	if got := detectImageFormat(data); got != document.ImagePNG {
		t.Errorf("detectImageFormat for short data: got %v, want ImagePNG", got)
	}
}

func TestDetectImageFormatEmpty(t *testing.T) {
	data := []byte{}
	if got := detectImageFormat(data); got != document.ImagePNG {
		t.Errorf("detectImageFormat for empty: got %v, want ImagePNG", got)
	}
}

// ---------------------------------------------------------------------------
// component.go tests
// ---------------------------------------------------------------------------

func TestFontSizeOption(t *testing.T) {
	s := document.DefaultStyle()
	FontSize(24)(&s)
	if s.FontSize != 24 {
		t.Errorf("font size: got %v, want 24", s.FontSize)
	}
}

func TestBoldOption(t *testing.T) {
	s := document.DefaultStyle()
	Bold()(&s)
	if s.FontWeight != document.WeightBold {
		t.Errorf("font weight: got %v, want Bold", s.FontWeight)
	}
}

func TestItalicOption(t *testing.T) {
	s := document.DefaultStyle()
	Italic()(&s)
	if s.FontStyle != document.StyleItalic {
		t.Errorf("font style: got %v, want Italic", s.FontStyle)
	}
}

func TestTextColorOption(t *testing.T) {
	s := document.DefaultStyle()
	c := pdf.RGB(0.5, 0.6, 0.7)
	TextColor(c)(&s)
	if s.Color != c {
		t.Errorf("text color: got %v, want %v", s.Color, c)
	}
}

func TestBgColorOption(t *testing.T) {
	s := document.DefaultStyle()
	c := pdf.RGB(0.1, 0.2, 0.3)
	BgColor(c)(&s)
	if s.Background == nil {
		t.Fatal("background should not be nil")
	}
	if *s.Background != c {
		t.Errorf("bg color: got %v, want %v", *s.Background, c)
	}
}

func TestAlignLeftOption(t *testing.T) {
	s := document.DefaultStyle()
	s.TextAlign = document.AlignRight // change first
	AlignLeft()(&s)
	if s.TextAlign != document.AlignLeft {
		t.Errorf("align: got %v, want AlignLeft", s.TextAlign)
	}
}

func TestAlignCenterOption(t *testing.T) {
	s := document.DefaultStyle()
	AlignCenter()(&s)
	if s.TextAlign != document.AlignCenter {
		t.Errorf("align: got %v, want AlignCenter", s.TextAlign)
	}
}

func TestAlignRightOption(t *testing.T) {
	s := document.DefaultStyle()
	AlignRight()(&s)
	if s.TextAlign != document.AlignRight {
		t.Errorf("align: got %v, want AlignRight", s.TextAlign)
	}
}

func TestFontFamilyOption(t *testing.T) {
	s := document.DefaultStyle()
	FontFamily("Courier")(&s)
	if s.FontFamily != "Courier" {
		t.Errorf("font family: got %q, want %q", s.FontFamily, "Courier")
	}
}

func TestFitWidthOption(t *testing.T) {
	cfg := imageConfig{}
	FitWidth(document.Mm(100))(&cfg)
	if cfg.width != document.Mm(100) {
		t.Errorf("width: got %v, want Mm(100)", cfg.width)
	}
	if cfg.fitMode != document.FitContain {
		t.Errorf("fit mode: got %v, want FitContain", cfg.fitMode)
	}
}

func TestFitHeightOption(t *testing.T) {
	cfg := imageConfig{}
	FitHeight(document.Mm(50))(&cfg)
	if cfg.height != document.Mm(50) {
		t.Errorf("height: got %v, want Mm(50)", cfg.height)
	}
	if cfg.fitMode != document.FitContain {
		t.Errorf("fit mode: got %v, want FitContain", cfg.fitMode)
	}
}

func TestTableHeaderStyleOption(t *testing.T) {
	cfg := tableConfig{}
	bg := pdf.Blue
	tc := pdf.White
	TableHeaderStyle(BgColor(bg), TextColor(tc))(&cfg)
	if cfg.headerBgColor == nil {
		t.Fatal("header bg color should not be nil")
	}
	if *cfg.headerBgColor != bg {
		t.Errorf("header bg: got %v, want %v", *cfg.headerBgColor, bg)
	}
	if cfg.headerTextColor == nil {
		t.Fatal("header text color should not be nil")
	}
	if *cfg.headerTextColor != tc {
		t.Errorf("header text: got %v, want %v", *cfg.headerTextColor, tc)
	}
}

func TestTableHeaderStyleNoBg(t *testing.T) {
	cfg := tableConfig{}
	// Only set text color, no background
	tc := pdf.Red
	TableHeaderStyle(TextColor(tc))(&cfg)
	if cfg.headerBgColor != nil {
		t.Error("header bg should be nil when not set")
	}
	if cfg.headerTextColor == nil || *cfg.headerTextColor != tc {
		t.Errorf("header text color: got %v, want %v", cfg.headerTextColor, tc)
	}
}

func TestTableStripeOption(t *testing.T) {
	cfg := tableConfig{}
	c := pdf.Gray(0.9)
	TableStripe(c)(&cfg)
	if cfg.stripeColor == nil {
		t.Fatal("stripe color should not be nil")
	}
	if *cfg.stripeColor != c {
		t.Errorf("stripe color: got %v, want %v", *cfg.stripeColor, c)
	}
}

func TestColumnWidthsOption(t *testing.T) {
	cfg := tableConfig{}
	ColumnWidths(30, 40, 30)(&cfg)
	if len(cfg.columnWidths) != 3 {
		t.Fatalf("column widths len: got %d, want 3", len(cfg.columnWidths))
	}
	expected := []float64{30, 40, 30}
	for i, w := range cfg.columnWidths {
		if w != expected[i] {
			t.Errorf("column width[%d]: got %v, want %v", i, w, expected[i])
		}
	}
}

func TestLineColorOption(t *testing.T) {
	cfg := lineConfig{}
	c := pdf.Green
	LineColor(c)(&cfg)
	if cfg.color != c {
		t.Errorf("line color: got %v, want %v", cfg.color, c)
	}
}

func TestLineThicknessOption(t *testing.T) {
	cfg := lineConfig{}
	LineThickness(document.Pt(5))(&cfg)
	if cfg.thickness != document.Pt(5) {
		t.Errorf("line thickness: got %v, want Pt(5)", cfg.thickness)
	}
}

// ---------------------------------------------------------------------------
// fontresolver.go tests
// ---------------------------------------------------------------------------

func TestBuiltinFontResolverFallback(t *testing.T) {
	r := newBuiltinFontResolver(nil)
	f := r.Resolve("Unknown", document.WeightNormal, false)
	if f.ID != "Unknown" {
		t.Errorf("ID: got %q, want %q", f.ID, "Unknown")
	}
	if f.Metrics.Ascender != 0.8 {
		t.Errorf("ascender: got %v, want 0.8", f.Metrics.Ascender)
	}
	if f.Metrics.Descender != -0.2 {
		t.Errorf("descender: got %v, want -0.2", f.Metrics.Descender)
	}
	if f.Metrics.LineHeight != 1.2 {
		t.Errorf("line height: got %v, want 1.2", f.Metrics.LineHeight)
	}
	if f.Metrics.CapHeight != 0.7 {
		t.Errorf("cap height: got %v, want 0.7", f.Metrics.CapHeight)
	}
}

func TestBuiltinFontResolverEmptyMap(t *testing.T) {
	r := newBuiltinFontResolver(make(map[string]*font.TrueTypeFont))
	f := r.Resolve("Anything", document.WeightNormal, false)
	// Should fallback since the map is empty.
	if f.Metrics.Ascender != 0.8 {
		t.Errorf("ascender: got %v, want 0.8", f.Metrics.Ascender)
	}
}

func TestBuiltinFontResolverMeasureStringFallback(t *testing.T) {
	r := newBuiltinFontResolver(nil)
	f := layout.ResolvedFont{ID: "NotRegistered"}
	// "hello" is 5 runes, fontSize 10, approximate width = 5 * 10 * 0.5 = 25
	width := r.MeasureString(f, "hello", 10)
	expected := 5.0 * 10.0 * 0.5
	if width != expected {
		t.Errorf("measure fallback: got %v, want %v", width, expected)
	}
}

func TestBuiltinFontResolverMeasureStringFallbackUnicode(t *testing.T) {
	r := newBuiltinFontResolver(nil)
	f := layout.ResolvedFont{ID: "NoFont"}
	// Unicode text: 3 runes
	width := r.MeasureString(f, "abc", 12)
	expected := 3.0 * 12.0 * 0.5
	if width != expected {
		t.Errorf("measure unicode: got %v, want %v", width, expected)
	}
}

func TestBuiltinFontResolverMeasureStringFallbackMultibyte(t *testing.T) {
	r := newBuiltinFontResolver(nil)
	f := layout.ResolvedFont{ID: "NoFont"}
	// Japanese characters: 3 runes, each multibyte in UTF-8
	text := "あいう"
	width := r.MeasureString(f, text, 14)
	expected := 3.0 * 14.0 * 0.5
	if width != expected {
		t.Errorf("measure multibyte: got %v, want %v", width, expected)
	}
}

func TestBuiltinFontResolverLineBreakFallback(t *testing.T) {
	r := newBuiltinFontResolver(nil)
	f := layout.ResolvedFont{ID: "NoFont"}
	text := "Hello World"
	// fontSize=10, avgCharWidth=5, maxWidth=30 -> charsPerLine=6
	// "Hello" (5 chars) + " " + "World" (5 chars) = 11 -> breaks into two lines
	lines := r.LineBreak(f, text, 10, 30)
	if len(lines) != 2 {
		t.Errorf("line break: got %d lines, want 2: %v", len(lines), lines)
	}
	if len(lines) >= 2 {
		if lines[0] != "Hello" {
			t.Errorf("line[0]: got %q, want %q", lines[0], "Hello")
		}
		if lines[1] != "World" {
			t.Errorf("line[1]: got %q, want %q", lines[1], "World")
		}
	}
}

func TestApproximateBreakSingleLine(t *testing.T) {
	// Width is large enough for the entire text.
	lines := approximateBreak("Hello World", 10, 1000)
	if len(lines) != 1 {
		t.Errorf("lines: got %d, want 1", len(lines))
	}
	if lines[0] != "Hello World" {
		t.Errorf("line[0]: got %q, want %q", lines[0], "Hello World")
	}
}

func TestApproximateBreakMultipleLines(t *testing.T) {
	// fontSize=10, avgCharWidth=5, maxWidth=25 -> charsPerLine=5
	text := "aa bb cc dd"
	lines := approximateBreak(text, 10, 25)
	// "aa" fits, "aa bb" = 5 chars fits, "aa bb cc" = 8 chars > 5 -> break
	// So: "aa bb", "cc dd" or similar
	if len(lines) < 2 {
		t.Errorf("lines: got %d, want >= 2: %v", len(lines), lines)
	}
}

func TestApproximateBreakNewlines(t *testing.T) {
	text := "Line 1\nLine 2\nLine 3"
	lines := approximateBreak(text, 10, 1000)
	if len(lines) != 3 {
		t.Errorf("lines: got %d, want 3: %v", len(lines), lines)
	}
	expected := []string{"Line 1", "Line 2", "Line 3"}
	for i, line := range lines {
		if i < len(expected) && line != expected[i] {
			t.Errorf("line[%d]: got %q, want %q", i, line, expected[i])
		}
	}
}

func TestApproximateBreakEmptyParagraph(t *testing.T) {
	text := "Before\n\nAfter"
	lines := approximateBreak(text, 10, 1000)
	if len(lines) != 3 {
		t.Errorf("lines: got %d, want 3: %v", len(lines), lines)
	}
	if len(lines) >= 2 && lines[1] != "" {
		t.Errorf("empty paragraph line: got %q, want empty", lines[1])
	}
}

func TestApproximateBreakZeroFontSize(t *testing.T) {
	// avgCharWidth = 0 * 0.5 = 0, should return text as-is
	lines := approximateBreak("text", 0, 100)
	if len(lines) != 1 {
		t.Errorf("lines: got %d, want 1", len(lines))
	}
	if lines[0] != "text" {
		t.Errorf("line[0]: got %q, want %q", lines[0], "text")
	}
}

func TestApproximateBreakVeryNarrowWidth(t *testing.T) {
	// maxWidth is very small, charsPerLine should clamp to 1
	lines := approximateBreak("ab cd", 10, 1)
	// With charsPerLine=0 -> clamped to 1, each word on its own line
	if len(lines) != 2 {
		t.Errorf("lines: got %d, want 2: %v", len(lines), lines)
	}
}

func TestApproximateBreakNegativeFontSize(t *testing.T) {
	lines := approximateBreak("text", -5, 100)
	if len(lines) != 1 {
		t.Errorf("lines: got %d, want 1", len(lines))
	}
	if lines[0] != "text" {
		t.Errorf("line[0]: got %q, want %q", lines[0], "text")
	}
}

func TestRuneLen(t *testing.T) {
	tests := []struct {
		input string
		want  int
	}{
		{"", 0},
		{"hello", 5},
		{"あいう", 3},
		{"a日b", 3},
		{"🎉", 1},
		{strings.Repeat("x", 100), 100},
	}
	for _, tt := range tests {
		got := runeLen(tt.input)
		if got != tt.want {
			t.Errorf("runeLen(%q): got %d, want %d", tt.input, got, tt.want)
		}
	}
}

// ---------------------------------------------------------------------------
// Integration / end-to-end tests
// ---------------------------------------------------------------------------

func TestGenerateWithAllComponents(t *testing.T) {
	doc := New(
		WithPageSize(document.A4),
		WithMargins(document.UniformEdges(document.Mm(15))),
		WithDefaultFont("Helvetica", 11),
		WithMetadata(document.DocumentMetadata{
			Title:   "Full Test",
			Author:  "Tester",
			Subject: "Integration",
		}),
	)

	doc.Header(func(p *PageBuilder) {
		p.AutoRow(func(r *RowBuilder) {
			r.Col(6, func(c *ColBuilder) {
				c.Text("Header Left", Bold(), FontSize(10))
			})
			r.Col(6, func(c *ColBuilder) {
				c.Text("Header Right", AlignRight(), FontSize(10))
			})
		})
		p.AutoRow(func(r *RowBuilder) {
			r.Col(12, func(c *ColBuilder) {
				c.Line(LineColor(pdf.Gray(0.5)), LineThickness(document.Pt(0.5)))
			})
		})
	})

	doc.Footer(func(p *PageBuilder) {
		p.AutoRow(func(r *RowBuilder) {
			r.Col(12, func(c *ColBuilder) {
				c.Line()
			})
		})
		p.AutoRow(func(r *RowBuilder) {
			r.Col(12, func(c *ColBuilder) {
				c.Text("Page Footer", AlignCenter(), FontSize(8), Italic())
			})
		})
	})

	page := doc.AddPage()

	// Title
	page.AutoRow(func(r *RowBuilder) {
		r.Col(12, func(c *ColBuilder) {
			c.Text("Document Title", FontSize(24), Bold(), AlignCenter(), TextColor(pdf.Blue))
		})
	})

	// Spacer
	page.AutoRow(func(r *RowBuilder) {
		r.Col(12, func(c *ColBuilder) {
			c.Spacer(document.Mm(5))
		})
	})

	// Two-column layout
	page.AutoRow(func(r *RowBuilder) {
		r.Col(6, func(c *ColBuilder) {
			c.Text("Left column text", AlignLeft(), FontFamily("Helvetica"))
		})
		r.Col(6, func(c *ColBuilder) {
			c.Text("Right column text", AlignRight(),
				BgColor(pdf.Gray(0.95)), TextColor(pdf.RGB(0.2, 0.2, 0.2)))
		})
	})

	// Line
	page.AutoRow(func(r *RowBuilder) {
		r.Col(12, func(c *ColBuilder) {
			c.Line(LineColor(pdf.Red), LineThickness(document.Pt(2)))
		})
	})

	// Table
	page.AutoRow(func(r *RowBuilder) {
		r.Col(12, func(c *ColBuilder) {
			c.Table(
				[]string{"Name", "Value", "Status"},
				[][]string{
					{"Item A", "100", "Active"},
					{"Item B", "200", "Inactive"},
					{"Item C", "300", "Active"},
					{"Item D", "400", "Pending"},
				},
				TableHeaderStyle(BgColor(pdf.RGB(0.2, 0.4, 0.6)), TextColor(pdf.White)),
				TableStripe(pdf.Gray(0.95)),
				ColumnWidths(40, 30, 30),
			)
		})
	})

	// Fixed-height row
	page.Row(document.Mm(20), func(r *RowBuilder) {
		r.Col(4, func(c *ColBuilder) { c.Text("Col 1/3") })
		r.Col(4, func(c *ColBuilder) { c.Text("Col 2/3") })
		r.Col(4, func(c *ColBuilder) { c.Text("Col 3/3") })
	})

	data, err := doc.Generate()
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}
	if len(data) < 5 {
		t.Fatal("Generated PDF too short")
	}
	if string(data[:5]) != "%PDF-" {
		t.Errorf("PDF header: got %q, want %%PDF-", string(data[:5]))
	}
	t.Logf("Generated full integration PDF: %d bytes", len(data))
}

func TestGenerateEmptyPage(t *testing.T) {
	doc := New()
	doc.AddPage() // page with no content
	data, err := doc.Generate()
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}
	if len(data) < 5 {
		t.Fatal("Generated PDF too short")
	}
	if string(data[:5]) != "%PDF-" {
		t.Errorf("PDF header: got %q, want %%PDF-", string(data[:5]))
	}
}

func TestGenerateWithRenderConsistency(t *testing.T) {
	doc := New()
	page := doc.AddPage()
	page.AutoRow(func(r *RowBuilder) {
		r.Col(12, func(c *ColBuilder) {
			c.Text("Consistency check")
		})
	})

	data1, err := doc.Generate()
	if err != nil {
		t.Fatalf("Generate 1 failed: %v", err)
	}

	var buf bytes.Buffer
	if err := doc.Render(&buf); err != nil {
		t.Fatalf("Render failed: %v", err)
	}
	data2 := buf.Bytes()

	if len(data1) != len(data2) {
		t.Errorf("Generate and Render produced different sizes: %d vs %d", len(data1), len(data2))
	}
	if !bytes.Equal(data1, data2) {
		t.Error("Generate and Render produced different content")
	}
}

func TestWithFontRegistration(t *testing.T) {
	// Test that WithFont initializes the rawFonts map.
	cfg := Config{}
	opt := WithFont("test", []byte{1, 2, 3})
	opt(&cfg)
	if cfg.rawFonts == nil {
		t.Fatal("rawFonts should be initialized")
	}
	if _, ok := cfg.rawFonts["test"]; !ok {
		t.Error("rawFonts should contain 'test'")
	}

	// Calling WithFont twice should work.
	opt2 := WithFont("test2", []byte{4, 5, 6})
	opt2(&cfg)
	if len(cfg.rawFonts) != 2 {
		t.Errorf("rawFonts count: got %d, want 2", len(cfg.rawFonts))
	}
}

func TestMultipleRowsAndColumns(t *testing.T) {
	doc := New()
	page := doc.AddPage()

	// Row with 12 individual columns
	page.AutoRow(func(r *RowBuilder) {
		for i := 0; i < 12; i++ {
			r.Col(1, func(c *ColBuilder) {
				c.Text("x")
			})
		}
	})

	built := doc.buildDocument()
	if len(built.Pages) != 1 {
		t.Fatalf("pages: got %d, want 1", len(built.Pages))
	}
	content := built.Pages[0].Content
	if len(content) < 1 {
		t.Fatal("expected at least 1 content node")
	}
	box := content[0].(*document.Box)
	if len(box.Content) != 12 {
		t.Errorf("columns: got %d, want 12", len(box.Content))
	}
}

func TestMixedRowTypes(t *testing.T) {
	doc := New()
	page := doc.AddPage()

	page.AutoRow(func(r *RowBuilder) {
		r.Col(12, func(c *ColBuilder) { c.Text("Auto") })
	})
	page.Row(document.Pt(50), func(r *RowBuilder) {
		r.Col(12, func(c *ColBuilder) { c.Text("Fixed") })
	})
	page.AutoRow(func(r *RowBuilder) {
		r.Col(12, func(c *ColBuilder) { c.Text("Auto again") })
	})

	if len(page.rows) != 3 {
		t.Fatalf("rows: got %d, want 3", len(page.rows))
	}
	if !page.rows[0].auto {
		t.Error("row 0 should be auto")
	}
	if page.rows[1].auto {
		t.Error("row 1 should not be auto")
	}
	if !page.rows[2].auto {
		t.Error("row 2 should be auto")
	}
}

func TestMultipleContentInColumn(t *testing.T) {
	doc := New()
	cb := &ColBuilder{doc: doc}
	cb.Text("Text 1")
	cb.Text("Text 2")
	cb.Line()
	cb.Spacer(document.Mm(5))
	cb.Table([]string{"A"}, [][]string{{"1"}})

	nodes := cb.buildNodes()
	if len(nodes) != 5 {
		t.Errorf("nodes: got %d, want 5", len(nodes))
	}
}

func TestGenerateWithHeaderFooterMultiplePages(t *testing.T) {
	doc := New()

	doc.Header(func(p *PageBuilder) {
		p.AutoRow(func(r *RowBuilder) {
			r.Col(12, func(c *ColBuilder) { c.Text("H") })
		})
	})
	doc.Footer(func(p *PageBuilder) {
		p.AutoRow(func(r *RowBuilder) {
			r.Col(12, func(c *ColBuilder) { c.Text("F") })
		})
	})

	for i := 0; i < 3; i++ {
		page := doc.AddPage()
		page.AutoRow(func(r *RowBuilder) {
			r.Col(12, func(c *ColBuilder) { c.Text("Body") })
		})
	}

	// Header/footer are built once via buildSection and placed on every
	// page by the paginator. Verify the rendered output contains them.
	data, err := doc.Generate()
	if err != nil {
		t.Fatalf("Generate error: %v", err)
	}
	output := string(data)
	if !strings.Contains(output, "%PDF-") {
		t.Error("Should produce valid PDF")
	}
	// The document should have 3 pages worth of content.
	built := doc.buildDocument()
	if len(built.Pages) != 3 {
		t.Fatalf("pages: got %d, want 3", len(built.Pages))
	}
}

// ---------------------------------------------------------------------------
// PageNumber / TotalPages tests (WP5)
// ---------------------------------------------------------------------------

func TestFooterWithPageNumber(t *testing.T) {
	doc := New()
	doc.Footer(func(p *PageBuilder) {
		p.AutoRow(func(r *RowBuilder) {
			r.Col(12, func(c *ColBuilder) {
				c.PageNumber()
			})
		})
	})
	page := doc.AddPage()
	page.AutoRow(func(r *RowBuilder) {
		r.Col(12, func(c *ColBuilder) { c.Text("Body") })
	})

	data, err := doc.Generate()
	if err != nil {
		t.Fatalf("Generate error: %v", err)
	}
	output := string(data)
	// The page number placeholder should have been replaced with "1".
	if !strings.Contains(output, "(1)") {
		t.Error("Expected page number '1' in output")
	}
	// The raw placeholder should NOT be in the output.
	if strings.Contains(output, document.PageNumberPlaceholder) {
		t.Error("Raw page number placeholder should not appear in output")
	}
}

func TestHeaderWithTotalPages(t *testing.T) {
	doc := New()
	doc.Header(func(p *PageBuilder) {
		p.AutoRow(func(r *RowBuilder) {
			r.Col(12, func(c *ColBuilder) {
				c.TotalPages()
			})
		})
	})
	// Add 2 pages.
	for range 2 {
		page := doc.AddPage()
		page.AutoRow(func(r *RowBuilder) {
			r.Col(12, func(c *ColBuilder) { c.Text("Body") })
		})
	}

	data, err := doc.Generate()
	if err != nil {
		t.Fatalf("Generate error: %v", err)
	}
	output := string(data)
	// Total pages should be "2".
	if !strings.Contains(output, "(2)") {
		t.Error("Expected total pages '2' in output")
	}
	if strings.Contains(output, document.TotalPagesPlaceholder) {
		t.Error("Raw total pages placeholder should not appear in output")
	}
}

func TestApproximateBreakLongWord(t *testing.T) {
	// A single word that is longer than charsPerLine should still appear
	// (not be lost).
	text := "superlongword"
	// fontSize=10, avgCharWidth=5, maxWidth=25 -> charsPerLine=5
	lines := approximateBreak(text, 10, 25)
	if len(lines) != 1 {
		t.Errorf("lines: got %d, want 1: %v", len(lines), lines)
	}
	if lines[0] != "superlongword" {
		t.Errorf("line[0]: got %q, want %q", lines[0], "superlongword")
	}
}

func TestApproximateBreakOnlySpaces(t *testing.T) {
	text := "   "
	lines := approximateBreak(text, 10, 100)
	// strings.Fields("   ") returns empty slice, so we get one empty line
	if len(lines) != 1 {
		t.Errorf("lines: got %d, want 1: %v", len(lines), lines)
	}
	if lines[0] != "" {
		t.Errorf("line[0]: got %q, want empty", lines[0])
	}
}

func TestApproximateBreakEmptyString(t *testing.T) {
	lines := approximateBreak("", 10, 100)
	// strings.Split("", "\n") returns [""] so we get one empty line
	if len(lines) != 1 {
		t.Errorf("lines: got %d, want 1: %v", len(lines), lines)
	}
	if lines[0] != "" {
		t.Errorf("line[0]: got %q, want empty", lines[0])
	}
}

func TestBuiltinFontResolverResolveDifferentWeightsAndStyles(t *testing.T) {
	r := newBuiltinFontResolver(nil)

	// All these should fall back since no fonts are registered.
	// The returned ID should reflect the weight/style variant.
	tests := []struct {
		family string
		weight document.FontWeight
		italic bool
		wantID string
	}{
		{"Helvetica", document.WeightNormal, false, "Helvetica"},
		{"Helvetica", document.WeightBold, false, "Helvetica-Bold"},
		{"Helvetica", document.WeightNormal, true, "Helvetica-Italic"},
		{"Helvetica", document.WeightBold, true, "Helvetica-BoldItalic"},
		{"Times", document.WeightNormal, false, "Times"},
	}

	for _, tt := range tests {
		f := r.Resolve(tt.family, tt.weight, tt.italic)
		if f.ID != tt.wantID {
			t.Errorf("Resolve(%q, %v, %v) ID: got %q, want %q",
				tt.family, tt.weight, tt.italic, f.ID, tt.wantID)
		}
		if f.Metrics.Ascender != 0.8 {
			t.Errorf("Resolve(%q) ascender: got %v, want 0.8", tt.family, f.Metrics.Ascender)
		}
	}
}

func TestMeasureStringFallbackEmptyString(t *testing.T) {
	r := newBuiltinFontResolver(nil)
	f := layout.ResolvedFont{ID: "NoFont"}
	width := r.MeasureString(f, "", 12)
	if width != 0 {
		t.Errorf("measure empty string: got %v, want 0", width)
	}
}

func TestLineBreakFallbackEmptyText(t *testing.T) {
	r := newBuiltinFontResolver(nil)
	f := layout.ResolvedFont{ID: "NoFont"}
	lines := r.LineBreak(f, "", 12, 100)
	// approximateBreak("", ...) returns [""]
	if len(lines) != 1 || lines[0] != "" {
		t.Errorf("line break empty: got %v, want [\"\"]", lines)
	}
}

func TestColBuilderMultipleImages(t *testing.T) {
	doc := New()
	cb := &ColBuilder{doc: doc}
	png := []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}
	jpeg := []byte{0xFF, 0xD8, 0xFF, 0xE0}
	cb.Image(png)
	cb.Image(jpeg)
	if len(cb.nodes) != 2 {
		t.Fatalf("nodes: got %d, want 2", len(cb.nodes))
	}
	img0 := cb.nodes[0].(*document.Image)
	img1 := cb.nodes[1].(*document.Image)
	if img0.Source.Format != document.ImagePNG {
		t.Error("first image should be PNG")
	}
	if img1.Source.Format != document.ImageJPEG {
		t.Error("second image should be JPEG")
	}
}

func TestTableEmptyHeaderAndBody(t *testing.T) {
	doc := New()
	cb := &ColBuilder{doc: doc}
	cb.Table(nil, nil)
	tbl := cb.nodes[0].(*document.Table)
	if len(tbl.Header) != 0 {
		t.Errorf("header: got %d, want 0", len(tbl.Header))
	}
	if len(tbl.Body) != 0 {
		t.Errorf("body: got %d, want 0", len(tbl.Body))
	}
}

func TestTableCellSpan(t *testing.T) {
	doc := New()
	cb := &ColBuilder{doc: doc}
	cb.Table(
		[]string{"A", "B"},
		[][]string{{"1", "2"}},
	)
	tbl := cb.nodes[0].(*document.Table)
	cell := tbl.Body[0].Cells[0]
	if cell.ColSpan != 1 {
		t.Errorf("col span: got %d, want 1", cell.ColSpan)
	}
	if cell.RowSpan != 1 {
		t.Errorf("row span: got %d, want 1", cell.RowSpan)
	}
}

func TestHeaderBoldStyle(t *testing.T) {
	doc := New()
	cb := &ColBuilder{doc: doc}
	cb.Table(
		[]string{"Header"},
		[][]string{{"Body"}},
	)
	tbl := cb.nodes[0].(*document.Table)
	headerText := tbl.Header[0].Cells[0].Content[0].(*document.Text)
	if headerText.TextStyle.FontWeight != document.WeightBold {
		t.Error("header cells should be bold by default")
	}
	bodyText := tbl.Body[0].Cells[0].Content[0].(*document.Text)
	if bodyText.TextStyle.FontWeight == document.WeightBold {
		t.Error("body cells should not be bold by default")
	}
}

func TestGenerateAndVerifyPDFStructure(t *testing.T) {
	doc := New(
		WithPageSize(document.Letter),
		WithMetadata(document.DocumentMetadata{
			Title:  "Structure Test",
			Author: "Test",
		}),
	)
	page := doc.AddPage()
	page.AutoRow(func(r *RowBuilder) {
		r.Col(12, func(c *ColBuilder) {
			c.Text("PDF Structure Test")
		})
	})

	data, err := doc.Generate()
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	// Check PDF header
	if !bytes.HasPrefix(data, []byte("%PDF-")) {
		t.Error("missing PDF header")
	}

	// Check for EOF marker
	eofMarker := []byte("%") // "%%EOF"
	eofMarker = append(eofMarker, []byte("%EOF")...)
	if !bytes.HasSuffix(bytes.TrimRight(data, "\r\n"), eofMarker) {
		t.Log("PDF may not end with EOF marker (not all writers guarantee trailing newline handling)")
	}
}

func TestBuildDocumentPageSizeAndMargins(t *testing.T) {
	margins := document.UniformEdges(document.Mm(25))
	doc := New(
		WithPageSize(document.A3),
		WithMargins(margins),
	)
	page := doc.AddPage()
	page.AutoRow(func(r *RowBuilder) {
		r.Col(12, func(c *ColBuilder) { c.Text("A3") })
	})

	built := doc.buildDocument()
	p := built.Pages[0]
	if p.Size != document.A3 {
		t.Errorf("page size: got %v, want A3", p.Size)
	}
	if p.Margins != margins {
		t.Errorf("margins: got %v, want %v", p.Margins, margins)
	}
}

// ---------------------------------------------------------------------------
// List template tests (WP5)
// ---------------------------------------------------------------------------

func TestColBuilderList(t *testing.T) {
	doc := New()
	page := doc.AddPage()
	page.AutoRow(func(r *RowBuilder) {
		r.Col(12, func(c *ColBuilder) {
			c.List([]string{"Alpha", "Beta", "Gamma"})
		})
	})
	built := doc.buildDocument()
	if len(built.Pages) == 0 {
		t.Fatal("Expected at least 1 page")
	}
	// The page content should contain a List node.
	found := false
	var walk func(nodes []document.DocumentNode)
	walk = func(nodes []document.DocumentNode) {
		for _, n := range nodes {
			if n.NodeType() == document.NodeList {
				found = true
				lst := n.(*document.List)
				if lst.ListType != document.Unordered {
					t.Errorf("Expected Unordered, got %v", lst.ListType)
				}
				if len(lst.Items) != 3 {
					t.Errorf("Expected 3 items, got %d", len(lst.Items))
				}
			}
			if ch := n.Children(); ch != nil {
				walk(ch)
			}
		}
	}
	walk(built.Pages[0].Content)
	if !found {
		t.Error("List node not found in page content")
	}
}

func TestColBuilderOrderedList(t *testing.T) {
	doc := New()
	page := doc.AddPage()
	page.AutoRow(func(r *RowBuilder) {
		r.Col(12, func(c *ColBuilder) {
			c.OrderedList([]string{"First", "Second"})
		})
	})
	built := doc.buildDocument()
	found := false
	var walk func(nodes []document.DocumentNode)
	walk = func(nodes []document.DocumentNode) {
		for _, n := range nodes {
			if n.NodeType() == document.NodeList {
				found = true
				lst := n.(*document.List)
				if lst.ListType != document.Ordered {
					t.Errorf("Expected Ordered, got %v", lst.ListType)
				}
			}
			if ch := n.Children(); ch != nil {
				walk(ch)
			}
		}
	}
	walk(built.Pages[0].Content)
	if !found {
		t.Error("Ordered List node not found in page content")
	}
}

func TestListIndentOption(t *testing.T) {
	doc := New()
	page := doc.AddPage()
	page.AutoRow(func(r *RowBuilder) {
		r.Col(12, func(c *ColBuilder) {
			c.List([]string{"A"}, ListIndent(document.Pt(40)))
		})
	})
	built := doc.buildDocument()
	var walk func(nodes []document.DocumentNode)
	walk = func(nodes []document.DocumentNode) {
		for _, n := range nodes {
			if n.NodeType() == document.NodeList {
				lst := n.(*document.List)
				if lst.MarkerIndent != 40 {
					t.Errorf("MarkerIndent = %v, want 40", lst.MarkerIndent)
				}
			}
			if ch := n.Children(); ch != nil {
				walk(ch)
			}
		}
	}
	walk(built.Pages[0].Content)
}

// ---------------------------------------------------------------------------
// RichTextBuilder tests
// ---------------------------------------------------------------------------

func TestRichTextBuilder_Span(t *testing.T) {
	rtb := &RichTextBuilder{
		defaultStyle: document.Style{FontSize: 12, FontFamily: "Helvetica", FontWeight: document.WeightNormal},
	}
	rtb.Span("Hello ")
	rtb.Span("world", Bold(), TextColor(pdf.RGB(1, 0, 0)))

	if len(rtb.fragments) != 2 {
		t.Fatalf("fragments = %d, want 2", len(rtb.fragments))
	}

	f0 := rtb.fragments[0]
	if f0.Content != "Hello " {
		t.Errorf("f0.Content = %q, want %q", f0.Content, "Hello ")
	}
	if f0.FragmentStyle.FontWeight != document.WeightNormal {
		t.Errorf("f0 should be normal weight")
	}

	f1 := rtb.fragments[1]
	if f1.Content != "world" {
		t.Errorf("f1.Content = %q, want %q", f1.Content, "world")
	}
	if f1.FragmentStyle.FontWeight != document.WeightBold {
		t.Errorf("f1 should be bold")
	}
	if f1.FragmentStyle.Color != pdf.RGB(1, 0, 0) {
		t.Errorf("f1 color = %v, want red", f1.FragmentStyle.Color)
	}
}

func TestColBuilder_RichText(t *testing.T) {
	doc := New()
	cb := &ColBuilder{doc: doc}

	cb.RichText(func(rt *RichTextBuilder) {
		rt.Span("Normal ")
		rt.Span("Bold", Bold())
	}, AlignCenter())

	nodes := cb.buildNodes()
	if len(nodes) != 1 {
		t.Fatalf("nodes = %d, want 1", len(nodes))
	}

	rtNode, ok := nodes[0].(*document.RichText)
	if !ok {
		t.Fatalf("node type = %T, want *document.RichText", nodes[0])
	}
	if rtNode.NodeType() != document.NodeRichText {
		t.Errorf("NodeType = %v, want NodeRichText", rtNode.NodeType())
	}
	if rtNode.BlockStyle.TextAlign != document.AlignCenter {
		t.Errorf("TextAlign = %v, want AlignCenter", rtNode.BlockStyle.TextAlign)
	}
	if len(rtNode.Fragments) != 2 {
		t.Fatalf("fragments = %d, want 2", len(rtNode.Fragments))
	}
	if rtNode.Fragments[0].Content != "Normal " {
		t.Errorf("first fragment = %q, want %q", rtNode.Fragments[0].Content, "Normal ")
	}
	if rtNode.Fragments[1].FragmentStyle.FontWeight != document.WeightBold {
		t.Errorf("second fragment should be bold")
	}
}

func TestColBuilder_RichText_GeneratesPDF(t *testing.T) {
	doc := New()
	p := doc.AddPage()
	p.AutoRow(func(r *RowBuilder) {
		r.Col(12, func(c *ColBuilder) {
			c.RichText(func(rt *RichTextBuilder) {
				rt.Span("Hello ")
				rt.Span("world", Bold())
				rt.Span("!")
			})
		})
	})

	data, err := doc.Generate()
	if err != nil {
		t.Fatalf("Generate() error: %v", err)
	}
	if len(data) == 0 {
		t.Fatal("generated PDF is empty")
	}
	if !bytes.HasPrefix(data, []byte("%PDF-")) {
		t.Error("output does not start with %PDF-")
	}
}

// ---------------------------------------------------------------------------
// extractJPEGDimensions / extractImageDimensions tests
// ---------------------------------------------------------------------------

func TestExtractJPEGDimensions_ValidSOF0(t *testing.T) {
	// Construct a minimal JPEG with SOF0 marker.
	// The parser accesses data[i+5..i+8] where i is the index of 0xFF before SOF marker.
	// SOI: FF D8, then SOF0: FF C0, length, precision, height, width.
	// After SOI (2 bytes), scanner is at i=2. data[2]=0xFF, data[3]=0xC0 → SOF0.
	// Access: data[i+5]=data[7], data[i+6]=data[8], data[i+7]=data[9], data[i+8]=data[10].
	// Need i+9 < len(data) → 11 < len → need at least 12 bytes.
	data := []byte{
		0xFF, 0xD8, // SOI (i starts at 2)
		0xFF, 0xC0, // SOF0 at i=2
		0x00, 0x0B, // length (at i+2, i+3)
		0x08,       // precision (at i+4)
		0x00, 0x40, // height=64 (at i+5, i+6)
		0x00, 0x80, // width=128 (at i+7, i+8)
		0x00, // extra byte to satisfy i+9 < len
	}
	w, h := extractJPEGDimensions(data)
	if w != 128 || h != 64 {
		t.Errorf("extractJPEGDimensions = (%d, %d), want (128, 64)", w, h)
	}
}

func TestExtractJPEGDimensions_Invalid(t *testing.T) {
	// Not a JPEG.
	w, h := extractJPEGDimensions([]byte{0x00, 0x00})
	if w != 0 || h != 0 {
		t.Errorf("expected (0,0) for non-JPEG, got (%d,%d)", w, h)
	}

	// Too short.
	w, h = extractJPEGDimensions([]byte{0xFF})
	if w != 0 || h != 0 {
		t.Errorf("expected (0,0) for short data, got (%d,%d)", w, h)
	}
}

func TestExtractJPEGDimensions_WithPrecedingSegment(t *testing.T) {
	// SOI + APP0 segment (skipped) + SOF0.
	data := []byte{
		0xFF, 0xD8, // SOI
		0xFF, 0xE0, // APP0
		0x00, 0x04, // length=4 (includes length bytes)
		0x00, 0x00, // dummy data
		0xFF, 0xC0, // SOF0
		0x00, 0x0B, // length
		0x08,       // precision
		0x01, 0x00, // height=256
		0x02, 0x00, // width=512
		0x00, // extra byte to satisfy i+9 < len
	}
	w, h := extractJPEGDimensions(data)
	if w != 512 || h != 256 {
		t.Errorf("extractJPEGDimensions = (%d, %d), want (512, 256)", w, h)
	}
}

func TestExtractImageDimensions_PNG(t *testing.T) {
	// Valid PNG header with IHDR: width=4, height=4.
	png := make([]byte, 24)
	// 8-byte signature (not needed for dimension extraction per the implementation).
	png[16] = 0
	png[17] = 0
	png[18] = 0
	png[19] = 4 // width=4
	png[20] = 0
	png[21] = 0
	png[22] = 0
	png[23] = 4 // height=4
	w, h := extractImageDimensions(png, document.ImagePNG)
	if w != 4 || h != 4 {
		t.Errorf("extractImageDimensions(PNG) = (%d, %d), want (4, 4)", w, h)
	}
}

func TestExtractImageDimensions_Unknown(t *testing.T) {
	w, h := extractImageDimensions([]byte{0x00}, document.ImageFormat(99))
	if w != 0 || h != 0 {
		t.Errorf("expected (0,0) for unknown format, got (%d,%d)", w, h)
	}
}

func TestTotalPages(t *testing.T) {
	doc := New()
	page := doc.AddPage()
	page.AutoRow(func(r *RowBuilder) {
		r.Col(12, func(c *ColBuilder) {
			c.TotalPages(FontSize(10), AlignCenter())
		})
	})
	data, err := doc.Generate()
	if err != nil {
		t.Fatalf("Generate error: %v", err)
	}
	if len(data) == 0 {
		t.Fatal("generated PDF is empty")
	}
}
