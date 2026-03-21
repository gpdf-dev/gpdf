package document

import (
	"math"
	"testing"

	"github.com/gpdf-dev/gpdf/pdf"
)

// ---------------------------------------------------------------------------
// units.go tests
// ---------------------------------------------------------------------------

func TestValueResolve_Pt(t *testing.T) {
	v := Pt(36)
	got := v.Resolve(100, 12)
	if got != 36 {
		t.Errorf("Pt(36).Resolve() = %v, want 36", got)
	}
}

func TestValueResolve_Mm(t *testing.T) {
	v := Mm(10)
	got := v.Resolve(100, 12)
	want := 10 * 2.83465
	if math.Abs(got-want) > 0.001 {
		t.Errorf("Mm(10).Resolve() = %v, want %v", got, want)
	}
}

func TestValueResolve_In(t *testing.T) {
	v := In(1)
	got := v.Resolve(100, 12)
	if got != 72 {
		t.Errorf("In(1).Resolve() = %v, want 72", got)
	}
}

func TestValueResolve_Cm(t *testing.T) {
	v := Cm(1)
	got := v.Resolve(100, 12)
	want := 28.3465
	if math.Abs(got-want) > 0.001 {
		t.Errorf("Cm(1).Resolve() = %v, want %v", got, want)
	}
}

func TestValueResolve_Em(t *testing.T) {
	v := Em(2)
	got := v.Resolve(100, 14)
	want := 2 * 14.0
	if got != want {
		t.Errorf("Em(2).Resolve(_, 14) = %v, want %v", got, want)
	}
}

func TestValueResolve_Pct(t *testing.T) {
	v := Pct(50)
	got := v.Resolve(200, 12)
	want := 100.0
	if got != want {
		t.Errorf("Pct(50).Resolve(200, _) = %v, want %v", got, want)
	}
}

func TestValueResolve_Auto(t *testing.T) {
	got := Auto.Resolve(200, 14)
	if got != 0 {
		t.Errorf("Auto.Resolve() = %v, want 0", got)
	}
}

func TestValueResolve_DefaultFallback(t *testing.T) {
	// Unit value beyond the known constants triggers the default branch.
	v := Value{Amount: 42, Unit: Unit(99)}
	got := v.Resolve(100, 12)
	if got != 42 {
		t.Errorf("Value{42, unknown}.Resolve() = %v, want 42", got)
	}
}

func TestValueIsAuto(t *testing.T) {
	if !Auto.IsAuto() {
		t.Error("Auto.IsAuto() = false, want true")
	}
	if Pt(10).IsAuto() {
		t.Error("Pt(10).IsAuto() = true, want false")
	}
	if Mm(5).IsAuto() {
		t.Error("Mm(5).IsAuto() = true, want false")
	}
}

func TestEdgesResolve(t *testing.T) {
	edges := Edges{
		Top:    Pt(10),
		Right:  Pct(50),
		Bottom: Mm(5),
		Left:   Em(1),
	}
	re := edges.Resolve(200, 400, 12)
	if re.Top != 10 {
		t.Errorf("Top = %v, want 10", re.Top)
	}
	// Right: 50% of parentWidth=200 => 100
	if re.Right != 100 {
		t.Errorf("Right = %v, want 100", re.Right)
	}
	// Bottom: 5mm => 5*2.83465
	wantBottom := 5 * 2.83465
	if math.Abs(re.Bottom-wantBottom) > 0.001 {
		t.Errorf("Bottom = %v, want %v", re.Bottom, wantBottom)
	}
	// Left: 1em => 1*12 = 12
	if re.Left != 12 {
		t.Errorf("Left = %v, want 12", re.Left)
	}
}

func TestUniformEdges(t *testing.T) {
	e := UniformEdges(Pt(5))
	if e.Top != Pt(5) || e.Right != Pt(5) || e.Bottom != Pt(5) || e.Left != Pt(5) {
		t.Errorf("UniformEdges(Pt(5)) not uniform: %+v", e)
	}
}

func TestResolvedEdgesHorizontal(t *testing.T) {
	re := ResolvedEdges{Top: 1, Right: 20, Bottom: 3, Left: 10}
	got := re.Horizontal()
	if got != 30 {
		t.Errorf("Horizontal() = %v, want 30", got)
	}
}

func TestResolvedEdgesVertical(t *testing.T) {
	re := ResolvedEdges{Top: 5, Right: 20, Bottom: 15, Left: 10}
	got := re.Vertical()
	if got != 20 {
		t.Errorf("Vertical() = %v, want 20", got)
	}
}

func TestPredefinedSizes(t *testing.T) {
	tests := []struct {
		name  string
		size  Size
		wantW float64
		wantH float64
	}{
		{"A4", A4, 595.28, 841.89},
		{"A3", A3, 841.89, 1190.55},
		{"Letter", Letter, 612, 792},
		{"Legal", Legal, 612, 1008},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.size.Width != tt.wantW {
				t.Errorf("Width = %v, want %v", tt.size.Width, tt.wantW)
			}
			if tt.size.Height != tt.wantH {
				t.Errorf("Height = %v, want %v", tt.size.Height, tt.wantH)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// style.go tests
// ---------------------------------------------------------------------------

func TestDefaultStyle(t *testing.T) {
	s := DefaultStyle()
	if s.FontSize != 12 {
		t.Errorf("FontSize = %v, want 12", s.FontSize)
	}
	if s.FontWeight != WeightNormal {
		t.Errorf("FontWeight = %v, want %v", s.FontWeight, WeightNormal)
	}
	if s.FontStyle != StyleNormal {
		t.Errorf("FontStyle = %v, want %v", s.FontStyle, StyleNormal)
	}
	if s.Color != pdf.Black {
		t.Errorf("Color = %+v, want Black", s.Color)
	}
	if s.TextAlign != AlignLeft {
		t.Errorf("TextAlign = %v, want %v", s.TextAlign, AlignLeft)
	}
	if s.LineHeight != 1.2 {
		t.Errorf("LineHeight = %v, want 1.2", s.LineHeight)
	}
	if s.Background != nil {
		t.Error("Background should be nil by default")
	}
}

func TestUniformBorder(t *testing.T) {
	width := Pt(2)
	style := BorderSolid
	color := pdf.Red
	b := UniformBorder(width, style, color)

	sides := []struct {
		name string
		side BorderSide
	}{
		{"Top", b.Top},
		{"Right", b.Right},
		{"Bottom", b.Bottom},
		{"Left", b.Left},
	}
	for _, s := range sides {
		if s.side.Width != width {
			t.Errorf("%s.Width = %v, want %v", s.name, s.side.Width, width)
		}
		if s.side.Style != style {
			t.Errorf("%s.Style = %v, want %v", s.name, s.side.Style, style)
		}
		if s.side.Color != color {
			t.Errorf("%s.Color = %+v, want %+v", s.name, s.side.Color, color)
		}
	}
}

func TestFontWeightConstants(t *testing.T) {
	if WeightNormal != 400 {
		t.Errorf("WeightNormal = %v, want 400", WeightNormal)
	}
	if WeightBold != 700 {
		t.Errorf("WeightBold = %v, want 700", WeightBold)
	}
}

func TestFontStyleConstants(t *testing.T) {
	if StyleNormal != 0 {
		t.Errorf("StyleNormal = %v, want 0", StyleNormal)
	}
	if StyleItalic != 1 {
		t.Errorf("StyleItalic = %v, want 1", StyleItalic)
	}
}

func TestTextAlignConstants(t *testing.T) {
	if AlignLeft != 0 {
		t.Errorf("AlignLeft = %v, want 0", AlignLeft)
	}
	if AlignCenter != 1 {
		t.Errorf("AlignCenter = %v, want 1", AlignCenter)
	}
	if AlignRight != 2 {
		t.Errorf("AlignRight = %v, want 2", AlignRight)
	}
	if AlignJustify != 3 {
		t.Errorf("AlignJustify = %v, want 3", AlignJustify)
	}
}

func TestBorderStyleConstants(t *testing.T) {
	if BorderNone != 0 {
		t.Errorf("BorderNone = %v, want 0", BorderNone)
	}
	if BorderSolid != 1 {
		t.Errorf("BorderSolid = %v, want 1", BorderSolid)
	}
	if BorderDashed != 2 {
		t.Errorf("BorderDashed = %v, want 2", BorderDashed)
	}
	if BorderDotted != 3 {
		t.Errorf("BorderDotted = %v, want 3", BorderDotted)
	}
}

// ---------------------------------------------------------------------------
// node.go tests
// ---------------------------------------------------------------------------

func TestNodeTypeConstants(t *testing.T) {
	tests := []struct {
		name string
		nt   NodeType
		want int
	}{
		{"NodeDocument", NodeDocument, 0},
		{"NodePage", NodePage, 1},
		{"NodeBox", NodeBox, 2},
		{"NodeText", NodeText, 3},
		{"NodeImage", NodeImage, 4},
		{"NodeTable", NodeTable, 5},
		{"NodeTableRow", NodeTableRow, 6},
		{"NodeTableCell", NodeTableCell, 7},
		{"NodeList", NodeList, 8},
		{"NodeListItem", NodeListItem, 9},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if int(tt.nt) != tt.want {
				t.Errorf("%s = %v, want %v", tt.name, int(tt.nt), tt.want)
			}
		})
	}
}

func TestBreakValueConstants(t *testing.T) {
	if BreakAuto != 0 {
		t.Errorf("BreakAuto = %v, want 0", BreakAuto)
	}
	if BreakAvoid != 1 {
		t.Errorf("BreakAvoid = %v, want 1", BreakAvoid)
	}
	if BreakAlways != 2 {
		t.Errorf("BreakAlways = %v, want 2", BreakAlways)
	}
	if BreakPage != 3 {
		t.Errorf("BreakPage = %v, want 3", BreakPage)
	}
}

// ---------------------------------------------------------------------------
// box.go tests
// ---------------------------------------------------------------------------

func TestBoxImplementsDocumentNode(t *testing.T) {
	var _ DocumentNode = &Box{}
}

func TestBoxNodeType(t *testing.T) {
	b := &Box{}
	if b.NodeType() != NodeBox {
		t.Errorf("Box.NodeType() = %v, want %v", b.NodeType(), NodeBox)
	}
}

func TestBoxChildren(t *testing.T) {
	child1 := &Text{Content: "hello"}
	child2 := &Text{Content: "world"}
	b := &Box{Content: []DocumentNode{child1, child2}}
	children := b.Children()
	if len(children) != 2 {
		t.Fatalf("len(Children) = %d, want 2", len(children))
	}
	if children[0] != child1 || children[1] != child2 {
		t.Error("Children do not match expected nodes")
	}
}

func TestBoxChildrenNil(t *testing.T) {
	b := &Box{}
	if children := b.Children(); children != nil {
		t.Errorf("Children() = %v, want nil", children)
	}
}

func TestBoxStyle(t *testing.T) {
	bg := pdf.Red
	b := &Box{
		BoxStyle: BoxStyle{
			Margin:     UniformEdges(Pt(10)),
			Padding:    UniformEdges(Pt(5)),
			Border:     UniformBorder(Pt(1), BorderSolid, pdf.Black),
			Background: &bg,
		},
	}
	s := b.Style()
	if s.Margin != UniformEdges(Pt(10)) {
		t.Error("Style().Margin mismatch")
	}
	if s.Padding != UniformEdges(Pt(5)) {
		t.Error("Style().Padding mismatch")
	}
	if s.Background == nil || *s.Background != bg {
		t.Error("Style().Background mismatch")
	}
}

// ---------------------------------------------------------------------------
// text.go tests
// ---------------------------------------------------------------------------

func TestTextImplementsDocumentNode(t *testing.T) {
	var _ DocumentNode = &Text{}
}

func TestTextNodeType(t *testing.T) {
	txt := &Text{Content: "hello"}
	if txt.NodeType() != NodeText {
		t.Errorf("Text.NodeType() = %v, want %v", txt.NodeType(), NodeText)
	}
}

func TestTextChildren(t *testing.T) {
	txt := &Text{Content: "hello"}
	if txt.Children() != nil {
		t.Error("Text.Children() should be nil")
	}
}

func TestTextStyle(t *testing.T) {
	s := DefaultStyle()
	s.FontSize = 24
	txt := &Text{Content: "hello", TextStyle: s}
	got := txt.Style()
	if got.FontSize != 24 {
		t.Errorf("Style().FontSize = %v, want 24", got.FontSize)
	}
}

// ---------------------------------------------------------------------------
// image.go tests
// ---------------------------------------------------------------------------

func TestImageImplementsDocumentNode(t *testing.T) {
	var _ DocumentNode = &Image{}
}

func TestImageNodeType(t *testing.T) {
	img := &Image{}
	if img.NodeType() != NodeImage {
		t.Errorf("Image.NodeType() = %v, want %v", img.NodeType(), NodeImage)
	}
}

func TestImageChildren(t *testing.T) {
	img := &Image{}
	if img.Children() != nil {
		t.Error("Image.Children() should be nil")
	}
}

func TestImageStyle(t *testing.T) {
	s := DefaultStyle()
	s.FontSize = 16
	img := &Image{ImgStyle: s}
	got := img.Style()
	if got.FontSize != 16 {
		t.Errorf("Style().FontSize = %v, want 16", got.FontSize)
	}
}

func TestImageFormatConstants(t *testing.T) {
	if ImageJPEG != 0 {
		t.Errorf("ImageJPEG = %v, want 0", ImageJPEG)
	}
	if ImagePNG != 1 {
		t.Errorf("ImagePNG = %v, want 1", ImagePNG)
	}
}

func TestImageFitModeConstants(t *testing.T) {
	if FitContain != 0 {
		t.Errorf("FitContain = %v, want 0", FitContain)
	}
	if FitCover != 1 {
		t.Errorf("FitCover = %v, want 1", FitCover)
	}
	if FitStretch != 2 {
		t.Errorf("FitStretch = %v, want 2", FitStretch)
	}
	if FitOriginal != 3 {
		t.Errorf("FitOriginal = %v, want 3", FitOriginal)
	}
}

func TestImageSource(t *testing.T) {
	src := ImageSource{
		Data:   []byte{0xFF, 0xD8, 0xFF},
		Format: ImageJPEG,
		Width:  640,
		Height: 480,
	}
	img := &Image{Source: src}
	if img.Source.Width != 640 {
		t.Errorf("Source.Width = %d, want 640", img.Source.Width)
	}
	if img.Source.Height != 480 {
		t.Errorf("Source.Height = %d, want 480", img.Source.Height)
	}
	if img.Source.Format != ImageJPEG {
		t.Errorf("Source.Format = %v, want ImageJPEG", img.Source.Format)
	}
}

// ---------------------------------------------------------------------------
// table.go tests
// ---------------------------------------------------------------------------

func TestTableImplementsDocumentNode(t *testing.T) {
	var _ DocumentNode = &Table{}
}

func TestTableNodeType(t *testing.T) {
	tbl := &Table{}
	if tbl.NodeType() != NodeTable {
		t.Errorf("Table.NodeType() = %v, want %v", tbl.NodeType(), NodeTable)
	}
}

func TestTableChildrenEmpty(t *testing.T) {
	tbl := &Table{}
	children := tbl.Children()
	if len(children) != 0 {
		t.Errorf("Empty Table.Children() length = %d, want 0", len(children))
	}
}

func TestTableChildren(t *testing.T) {
	tbl := &Table{
		Header: []TableRow{
			{Cells: []TableCell{{Content: []DocumentNode{&Text{Content: "H1"}}}}},
		},
		Body: []TableRow{
			{Cells: []TableCell{
				{Content: []DocumentNode{&Text{Content: "B1"}}},
				{Content: []DocumentNode{&Text{Content: "B2"}}},
			}},
		},
		Footer: []TableRow{
			{Cells: []TableCell{{Content: []DocumentNode{&Text{Content: "F1"}}}}},
		},
	}
	children := tbl.Children()
	// 1 header cell + 2 body cells + 1 footer cell = 4
	if len(children) != 4 {
		t.Fatalf("Table.Children() length = %d, want 4", len(children))
	}
	// Each child should be a cellNode implementing DocumentNode with NodeTableCell type.
	for i, child := range children {
		if child.NodeType() != NodeTableCell {
			t.Errorf("child[%d].NodeType() = %v, want NodeTableCell", i, child.NodeType())
		}
	}
}

func TestTableStyle(t *testing.T) {
	bg := pdf.Blue
	tbl := &Table{
		TableStyle: TableStyle{
			BoxStyle: BoxStyle{
				Margin:     UniformEdges(Pt(5)),
				Padding:    UniformEdges(Pt(3)),
				Background: &bg,
			},
		},
	}
	s := tbl.Style()
	if s.Margin != UniformEdges(Pt(5)) {
		t.Error("Table.Style().Margin mismatch")
	}
	if s.Background == nil || *s.Background != bg {
		t.Error("Table.Style().Background mismatch")
	}
}

func TestTableCellNodeChildren(t *testing.T) {
	txt := &Text{Content: "cell"}
	tbl := &Table{
		Body: []TableRow{
			{Cells: []TableCell{{Content: []DocumentNode{txt}}}},
		},
	}
	children := tbl.Children()
	if len(children) != 1 {
		t.Fatalf("Expected 1 child, got %d", len(children))
	}
	cellChildren := children[0].Children()
	if len(cellChildren) != 1 {
		t.Fatalf("Expected 1 cell child, got %d", len(cellChildren))
	}
	if cellChildren[0] != txt {
		t.Error("Cell child does not match expected text node")
	}
}

func TestTableCellNodeStyle(t *testing.T) {
	s := DefaultStyle()
	s.FontSize = 14
	tbl := &Table{
		Body: []TableRow{
			{Cells: []TableCell{{CellStyle: s}}},
		},
	}
	children := tbl.Children()
	if len(children) != 1 {
		t.Fatal("Expected 1 child")
	}
	got := children[0].Style()
	if got.FontSize != 14 {
		t.Errorf("CellNode.Style().FontSize = %v, want 14", got.FontSize)
	}
}

// ---------------------------------------------------------------------------
// page.go tests
// ---------------------------------------------------------------------------

func TestPageImplementsDocumentNode(t *testing.T) {
	var _ DocumentNode = &Page{}
}

func TestPageNodeType(t *testing.T) {
	p := &Page{}
	if p.NodeType() != NodePage {
		t.Errorf("Page.NodeType() = %v, want %v", p.NodeType(), NodePage)
	}
}

func TestPageChildren(t *testing.T) {
	txt := &Text{Content: "page content"}
	p := &Page{Content: []DocumentNode{txt}}
	children := p.Children()
	if len(children) != 1 {
		t.Fatalf("Page.Children() length = %d, want 1", len(children))
	}
	if children[0] != txt {
		t.Error("Page child does not match")
	}
}

func TestPageStyle(t *testing.T) {
	s := DefaultStyle()
	s.FontSize = 10
	p := &Page{PageStyle: s}
	got := p.Style()
	if got.FontSize != 10 {
		t.Errorf("Page.Style().FontSize = %v, want 10", got.FontSize)
	}
}

func TestDocumentImplementsDocumentNode(t *testing.T) {
	var _ DocumentNode = &Document{}
}

func TestDocumentNodeType(t *testing.T) {
	doc := &Document{}
	if doc.NodeType() != NodeDocument {
		t.Errorf("Document.NodeType() = %v, want %v", doc.NodeType(), NodeDocument)
	}
}

func TestDocumentChildren(t *testing.T) {
	p1 := &Page{Size: A4}
	p2 := &Page{Size: Letter}
	doc := &Document{Pages: []*Page{p1, p2}}
	children := doc.Children()
	if len(children) != 2 {
		t.Fatalf("Document.Children() length = %d, want 2", len(children))
	}
	if children[0] != p1 {
		t.Error("Document child[0] mismatch")
	}
	if children[1] != p2 {
		t.Error("Document child[1] mismatch")
	}
}

func TestDocumentChildrenEmpty(t *testing.T) {
	doc := &Document{}
	children := doc.Children()
	if len(children) != 0 {
		t.Errorf("Empty Document.Children() length = %d, want 0", len(children))
	}
}

func TestDocumentStyle(t *testing.T) {
	s := DefaultStyle()
	doc := &Document{DefaultStyle: s}
	got := doc.Style()
	if got.FontSize != 12 {
		t.Errorf("Document.Style().FontSize = %v, want 12", got.FontSize)
	}
}

func TestDocumentMetadata(t *testing.T) {
	meta := DocumentMetadata{
		Title:    "Test Title",
		Author:   "Test Author",
		Subject:  "Test Subject",
		Creator:  "Test Creator",
		Producer: "Test Producer",
	}
	doc := &Document{Metadata: meta}
	if doc.Metadata.Title != "Test Title" {
		t.Error("Metadata.Title mismatch")
	}
	if doc.Metadata.Author != "Test Author" {
		t.Error("Metadata.Author mismatch")
	}
}

// ---------------------------------------------------------------------------
// Convenience constructor coverage
// ---------------------------------------------------------------------------

func TestValueConstructors(t *testing.T) {
	tests := []struct {
		name string
		val  Value
		unit Unit
		amt  float64
	}{
		{"Pt", Pt(10), UnitPt, 10},
		{"Mm", Mm(25.4), UnitMm, 25.4},
		{"In", In(1), UnitIn, 1},
		{"Cm", Cm(2.54), UnitCm, 2.54},
		{"Em", Em(1.5), UnitEm, 1.5},
		{"Pct", Pct(100), UnitPct, 100},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.val.Unit != tt.unit {
				t.Errorf("Unit = %v, want %v", tt.val.Unit, tt.unit)
			}
			if tt.val.Amount != tt.amt {
				t.Errorf("Amount = %v, want %v", tt.val.Amount, tt.amt)
			}
		})
	}
}

func TestPointAndRectangle(t *testing.T) {
	pt := Point{X: 10, Y: 20}
	if pt.X != 10 || pt.Y != 20 {
		t.Errorf("Point = %+v, want {10, 20}", pt)
	}

	rect := Rectangle{X: 0, Y: 0, Width: 100, Height: 200}
	if rect.Width != 100 || rect.Height != 200 {
		t.Errorf("Rectangle = %+v, unexpected", rect)
	}
}

func TestBreakPolicy(t *testing.T) {
	bp := BreakPolicy{
		BreakBefore: BreakAlways,
		BreakAfter:  BreakAuto,
		BreakInside: BreakAvoid,
	}
	if bp.BreakBefore != BreakAlways {
		t.Error("BreakBefore mismatch")
	}
	if bp.BreakAfter != BreakAuto {
		t.Error("BreakAfter mismatch")
	}
	if bp.BreakInside != BreakAvoid {
		t.Error("BreakInside mismatch")
	}
}

func TestBoxWithBreakPolicy(t *testing.T) {
	b := &Box{
		BreakPolicy: BreakPolicy{
			BreakBefore: BreakPage,
		},
	}
	if b.BreakPolicy.BreakBefore != BreakPage {
		t.Error("Box.BreakPolicy.BreakBefore mismatch")
	}
}

func TestTableColumns(t *testing.T) {
	tbl := &Table{
		Columns: []TableColumn{
			{Width: Pct(50)},
			{Width: Pct(50)},
		},
	}
	if len(tbl.Columns) != 2 {
		t.Fatalf("Columns count = %d, want 2", len(tbl.Columns))
	}
	if tbl.Columns[0].Width != Pct(50) {
		t.Error("Column 0 width mismatch")
	}
}

func TestTableCellSpan(t *testing.T) {
	cell := TableCell{
		ColSpan: 2,
		RowSpan: 3,
	}
	if cell.ColSpan != 2 {
		t.Errorf("ColSpan = %d, want 2", cell.ColSpan)
	}
	if cell.RowSpan != 3 {
		t.Errorf("RowSpan = %d, want 3", cell.RowSpan)
	}
}

func TestTableStyleBorderCollapse(t *testing.T) {
	ts := TableStyle{BorderCollapse: true}
	if !ts.BorderCollapse {
		t.Error("BorderCollapse should be true")
	}
}

// ---------------------------------------------------------------------------
// list.go tests
// ---------------------------------------------------------------------------

func TestListImplementsDocumentNode(t *testing.T) {
	var _ DocumentNode = &List{}
}

func TestListNodeType(t *testing.T) {
	l := &List{}
	if l.NodeType() != NodeList {
		t.Errorf("List.NodeType() = %v, want %v", l.NodeType(), NodeList)
	}
}

func TestListChildren(t *testing.T) {
	l := &List{
		Items: []ListItem{
			{Content: []DocumentNode{&Text{Content: "A"}}},
			{Content: []DocumentNode{&Text{Content: "B"}}},
		},
	}
	children := l.Children()
	if len(children) != 2 {
		t.Fatalf("List.Children() length = %d, want 2", len(children))
	}
	for i, child := range children {
		if child.NodeType() != NodeListItem {
			t.Errorf("child[%d].NodeType() = %v, want NodeListItem", i, child.NodeType())
		}
	}
}

func TestListItemNodeChildren(t *testing.T) {
	txt := &Text{Content: "item"}
	l := &List{
		Items: []ListItem{
			{Content: []DocumentNode{txt}},
		},
	}
	children := l.Children()
	if len(children) != 1 {
		t.Fatalf("Expected 1 child, got %d", len(children))
	}
	itemChildren := children[0].Children()
	if len(itemChildren) != 1 {
		t.Fatalf("Expected 1 item child, got %d", len(itemChildren))
	}
	if itemChildren[0] != txt {
		t.Error("Item child does not match expected text node")
	}
}

func TestListTypeConstants(t *testing.T) {
	if Unordered != 0 {
		t.Errorf("Unordered = %v, want 0", Unordered)
	}
	if Ordered != 1 {
		t.Errorf("Ordered = %v, want 1", Ordered)
	}
}

func TestListStyle(t *testing.T) {
	s := Style{FontSize: 14, Color: pdf.RGB(0, 0, 0)}
	l := &List{
		ListStyle: s,
		Items: []ListItem{
			{Content: []DocumentNode{&Text{Content: "A"}}},
		},
	}
	got := l.Style()
	if got.FontSize != 14 {
		t.Errorf("List.Style().FontSize = %v, want 14", got.FontSize)
	}
}

func TestListItemNodeStyle_FallbackToListStyle(t *testing.T) {
	listStyle := Style{FontSize: 12}
	l := &List{
		ListStyle: listStyle,
		Items: []ListItem{
			{Content: []DocumentNode{&Text{Content: "item"}}},
		},
	}
	children := l.Children()
	got := children[0].Style()
	if got.FontSize != 12 {
		t.Errorf("ListItemNode.Style().FontSize = %v, want 12 (list fallback)", got.FontSize)
	}
}

func TestListItemNodeStyle_ExplicitItemStyle(t *testing.T) {
	listStyle := Style{FontSize: 12}
	itemStyle := Style{FontSize: 18}
	l := &List{
		ListStyle: listStyle,
		Items: []ListItem{
			{Content: []DocumentNode{&Text{Content: "item"}}, ItemStyle: itemStyle},
		},
	}
	children := l.Children()
	got := children[0].Style()
	if got.FontSize != 18 {
		t.Errorf("ListItemNode.Style().FontSize = %v, want 18 (explicit item style)", got.FontSize)
	}
}
