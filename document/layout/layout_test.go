package layout

import (
	"fmt"
	"math"
	"strings"
	"testing"

	"github.com/gpdf-dev/gpdf/document"
	"github.com/gpdf-dev/gpdf/pdf"
)

// approxEqual reports whether a and b are within epsilon of each other.
func approxEqual(a, b, epsilon float64) bool {
	return math.Abs(a-b) < epsilon
}

// ---------------------------------------------------------------------------
// Mock FontResolver
// ---------------------------------------------------------------------------

// mockFontResolver returns predictable metrics for testing.
// It measures strings at 0.5 * fontSize per character and breaks lines
// at word boundaries when the line exceeds maxWidth.
type mockFontResolver struct{}

func (m *mockFontResolver) Resolve(family string, weight document.FontWeight, italic bool) ResolvedFont {
	id := family
	if id == "" {
		id = "MockFont"
	}
	if weight == document.WeightBold {
		id += "-Bold"
	}
	if italic {
		id += "-Italic"
	}
	return ResolvedFont{
		ID: id,
		Metrics: FontMetrics{
			Ascender:   0.8,
			Descender:  -0.2,
			LineHeight: 1.0,
			CapHeight:  0.7,
		},
	}
}

func (m *mockFontResolver) MeasureString(_ ResolvedFont, text string, size float64) float64 {
	// Each character is 0.5 * fontSize wide.
	return float64(len([]rune(text))) * size * 0.5
}

func (m *mockFontResolver) LineBreak(_ ResolvedFont, text string, size float64, maxWidth float64) []string {
	avgCharWidth := size * 0.5
	if avgCharWidth <= 0 {
		return []string{text}
	}
	charsPerLine := int(maxWidth / avgCharWidth)
	if charsPerLine <= 0 {
		charsPerLine = 1
	}

	words := strings.Fields(text)
	if len(words) == 0 {
		return nil
	}

	var lines []string
	current := words[0]
	for _, w := range words[1:] {
		if len([]rune(current))+1+len([]rune(w)) <= charsPerLine {
			current += " " + w
		} else {
			lines = append(lines, current)
			current = w
		}
	}
	lines = append(lines, current)
	return lines
}

// ---------------------------------------------------------------------------
// engine.go type tests
// ---------------------------------------------------------------------------

func TestConstraintsFields(t *testing.T) {
	c := Constraints{
		AvailableWidth:  500,
		AvailableHeight: 700,
		FontResolver:    &mockFontResolver{},
	}
	if c.AvailableWidth != 500 {
		t.Errorf("AvailableWidth = %v, want 500", c.AvailableWidth)
	}
	if c.AvailableHeight != 700 {
		t.Errorf("AvailableHeight = %v, want 700", c.AvailableHeight)
	}
	if c.FontResolver == nil {
		t.Error("FontResolver should not be nil")
	}
}

func TestResultFields(t *testing.T) {
	r := Result{
		Bounds:   document.Rectangle{X: 0, Y: 0, Width: 100, Height: 50},
		Children: []PlacedNode{{Node: &document.Text{Content: "x"}}},
		Overflow: nil,
	}
	if r.Bounds.Width != 100 {
		t.Errorf("Bounds.Width = %v, want 100", r.Bounds.Width)
	}
	if len(r.Children) != 1 {
		t.Errorf("len(Children) = %d, want 1", len(r.Children))
	}
	if r.Overflow != nil {
		t.Error("Overflow should be nil")
	}
}

func TestPlacedNodeFields(t *testing.T) {
	pn := PlacedNode{
		Node:     &document.Text{Content: "test"},
		Position: document.Point{X: 10, Y: 20},
		Size:     document.Size{Width: 100, Height: 14},
	}
	if pn.Position.X != 10 || pn.Position.Y != 20 {
		t.Errorf("Position = %+v, unexpected", pn.Position)
	}
	if pn.Size.Width != 100 || pn.Size.Height != 14 {
		t.Errorf("Size = %+v, unexpected", pn.Size)
	}
}

func TestResolvedFontFields(t *testing.T) {
	rf := ResolvedFont{
		ID: "Helvetica",
		Metrics: FontMetrics{
			Ascender:   0.8,
			Descender:  -0.2,
			LineHeight: 1.0,
			CapHeight:  0.7,
		},
	}
	if rf.ID != "Helvetica" {
		t.Errorf("ID = %v, want Helvetica", rf.ID)
	}
	if rf.Metrics.Ascender != 0.8 {
		t.Errorf("Ascender = %v, want 0.8", rf.Metrics.Ascender)
	}
}

// ---------------------------------------------------------------------------
// block.go tests
// ---------------------------------------------------------------------------

func TestNewBlockLayout(t *testing.T) {
	bl := NewBlockLayout()
	if bl == nil {
		t.Fatal("NewBlockLayout() returned nil")
	}
}

func TestBlockLayoutEmptyBox(t *testing.T) {
	bl := NewBlockLayout()
	box := &document.Box{}
	constraints := Constraints{
		AvailableWidth:  500,
		AvailableHeight: 700,
		FontResolver:    &mockFontResolver{},
	}
	result := bl.Layout(box, constraints)
	if result.Overflow != nil {
		t.Error("Empty box should not overflow")
	}
	if len(result.Children) != 0 {
		t.Errorf("Empty box should have 0 children, got %d", len(result.Children))
	}
	if result.Bounds.Width != 500 {
		t.Errorf("Bounds.Width = %v, want 500", result.Bounds.Width)
	}
}

func TestBlockLayoutWithTextChild(t *testing.T) {
	bl := NewBlockLayout()
	txt := &document.Text{
		Content: "Hello world",
		TextStyle: document.Style{
			FontSize:   12,
			LineHeight: 1.2,
		},
	}
	box := &document.Box{Content: []document.DocumentNode{txt}}
	constraints := Constraints{
		AvailableWidth:  500,
		AvailableHeight: 700,
		FontResolver:    &mockFontResolver{},
	}
	result := bl.Layout(box, constraints)
	if result.Overflow != nil {
		t.Error("Short text should not overflow")
	}
	if len(result.Children) != 1 {
		t.Fatalf("Expected 1 child, got %d", len(result.Children))
	}
	// The child should be placed at the content area origin.
	child := result.Children[0]
	if child.Position.X != 0 {
		t.Errorf("Child X = %v, want 0", child.Position.X)
	}
	if child.Position.Y != 0 {
		t.Errorf("Child Y = %v, want 0", child.Position.Y)
	}
}

func TestBlockLayoutWithPaddingAndMargin(t *testing.T) {
	bl := NewBlockLayout()
	txt := &document.Text{
		Content: "Hi",
		TextStyle: document.Style{
			FontSize:   12,
			LineHeight: 1.2,
		},
	}
	box := &document.Box{
		Content: []document.DocumentNode{txt},
		BoxStyle: document.BoxStyle{
			Margin:  document.UniformEdges(document.Pt(10)),
			Padding: document.UniformEdges(document.Pt(5)),
		},
	}
	constraints := Constraints{
		AvailableWidth:  500,
		AvailableHeight: 700,
		FontResolver:    &mockFontResolver{},
	}
	result := bl.Layout(box, constraints)
	if result.Overflow != nil {
		t.Error("Should not overflow")
	}
	if len(result.Children) != 1 {
		t.Fatalf("Expected 1 child, got %d", len(result.Children))
	}
	child := result.Children[0]
	// Content origin = margin.Left + padding.Left = 10 + 5 = 15
	if child.Position.X != 15 {
		t.Errorf("Child X = %v, want 15", child.Position.X)
	}
	if child.Position.Y != 15 {
		t.Errorf("Child Y = %v, want 15", child.Position.Y)
	}
}

func TestBlockLayoutWithBorder(t *testing.T) {
	bl := NewBlockLayout()
	txt := &document.Text{
		Content:   "text",
		TextStyle: document.Style{FontSize: 12, LineHeight: 1.2},
	}
	box := &document.Box{
		Content: []document.DocumentNode{txt},
		BoxStyle: document.BoxStyle{
			Border: document.UniformBorder(document.Pt(2), document.BorderSolid, pdf.Color{}),
		},
	}
	constraints := Constraints{
		AvailableWidth:  500,
		AvailableHeight: 700,
		FontResolver:    &mockFontResolver{},
	}
	result := bl.Layout(box, constraints)
	if len(result.Children) != 1 {
		t.Fatalf("Expected 1 child, got %d", len(result.Children))
	}
	child := result.Children[0]
	// Content origin = borderWidth.Left = 2 (no margin/padding)
	if child.Position.X != 2 {
		t.Errorf("Child X = %v, want 2", child.Position.X)
	}
}

func TestBlockLayoutOverflow(t *testing.T) {
	bl := NewBlockLayout()
	// Create multiple text children that will exhaust vertical space.
	var children []document.DocumentNode
	for i := 0; i < 100; i++ {
		children = append(children, &document.Text{
			Content:   "This is a relatively long line of text that takes some vertical space.",
			TextStyle: document.Style{FontSize: 12, LineHeight: 1.5},
		})
	}
	box := &document.Box{Content: children}
	constraints := Constraints{
		AvailableWidth:  300,
		AvailableHeight: 100, // Very limited vertical space
		FontResolver:    &mockFontResolver{},
	}
	result := bl.Layout(box, constraints)
	if result.Overflow == nil {
		t.Error("Expected overflow with limited height")
	}
}

func TestBlockLayoutNilChildSkipped(t *testing.T) {
	bl := NewBlockLayout()
	txt := &document.Text{
		Content:   "hello",
		TextStyle: document.Style{FontSize: 12, LineHeight: 1.2},
	}
	box := &document.Box{Content: []document.DocumentNode{nil, txt}}
	constraints := Constraints{
		AvailableWidth:  500,
		AvailableHeight: 700,
		FontResolver:    &mockFontResolver{},
	}
	result := bl.Layout(box, constraints)
	if len(result.Children) != 1 {
		t.Errorf("Expected 1 placed child (nil skipped), got %d", len(result.Children))
	}
}

func TestBlockLayoutMultipleTextChildren(t *testing.T) {
	bl := NewBlockLayout()
	txt1 := &document.Text{
		Content:   "First",
		TextStyle: document.Style{FontSize: 12, LineHeight: 1.2},
	}
	txt2 := &document.Text{
		Content:   "Second",
		TextStyle: document.Style{FontSize: 12, LineHeight: 1.2},
	}
	box := &document.Box{Content: []document.DocumentNode{txt1, txt2}}
	constraints := Constraints{
		AvailableWidth:  500,
		AvailableHeight: 700,
		FontResolver:    &mockFontResolver{},
	}
	result := bl.Layout(box, constraints)
	if len(result.Children) != 2 {
		t.Fatalf("Expected 2 children, got %d", len(result.Children))
	}
	// Second child should be below the first.
	if result.Children[1].Position.Y <= result.Children[0].Position.Y {
		t.Error("Second child should be below the first")
	}
}

func TestBlockLayoutNestedBox(t *testing.T) {
	bl := NewBlockLayout()
	inner := &document.Box{
		Content: []document.DocumentNode{
			&document.Text{
				Content:   "inner",
				TextStyle: document.Style{FontSize: 12, LineHeight: 1.2},
			},
		},
	}
	outer := &document.Box{Content: []document.DocumentNode{inner}}
	constraints := Constraints{
		AvailableWidth:  500,
		AvailableHeight: 700,
		FontResolver:    &mockFontResolver{},
	}
	result := bl.Layout(outer, constraints)
	if result.Overflow != nil {
		t.Error("Nested box should not overflow")
	}
	if len(result.Children) != 1 {
		t.Fatalf("Expected 1 child (inner box), got %d", len(result.Children))
	}
}

func TestBlockLayoutZeroAvailableHeight(t *testing.T) {
	bl := NewBlockLayout()
	txt := &document.Text{
		Content:   "text",
		TextStyle: document.Style{FontSize: 12, LineHeight: 1.2},
	}
	box := &document.Box{Content: []document.DocumentNode{txt}}
	constraints := Constraints{
		AvailableWidth:  500,
		AvailableHeight: 0,
		FontResolver:    &mockFontResolver{},
	}
	result := bl.Layout(box, constraints)
	// With 0 height, negative contentHeight gets clamped to 0, and
	// the first child should trigger overflow.
	if result.Overflow == nil {
		t.Error("Expected overflow with zero height")
	}
}

func TestBlockLayoutDefaultFontSize(t *testing.T) {
	bl := NewBlockLayout()
	// Box with no font size set (defaults to 0 in Style).
	box := &document.Box{
		Content: []document.DocumentNode{
			&document.Text{
				Content:   "text",
				TextStyle: document.Style{FontSize: 12, LineHeight: 1.2},
			},
		},
	}
	constraints := Constraints{
		AvailableWidth:  500,
		AvailableHeight: 700,
		FontResolver:    &mockFontResolver{},
	}
	// Should not panic; will use default fontSize=12 for resolving.
	result := bl.Layout(box, constraints)
	if result.Overflow != nil {
		t.Error("Should not overflow")
	}
}

func TestBlockLayoutChildOverflowCarriesRemaining(t *testing.T) {
	bl := NewBlockLayout()
	// A long text that wraps into many lines.
	longText := strings.Repeat("word ", 200)
	txt1 := &document.Text{
		Content:   longText,
		TextStyle: document.Style{FontSize: 12, LineHeight: 1.5},
	}
	txt2 := &document.Text{
		Content:   "after",
		TextStyle: document.Style{FontSize: 12, LineHeight: 1.5},
	}
	box := &document.Box{Content: []document.DocumentNode{txt1, txt2}}
	constraints := Constraints{
		AvailableWidth:  200,
		AvailableHeight: 50, // Very small
		FontResolver:    &mockFontResolver{},
	}
	result := bl.Layout(box, constraints)
	if result.Overflow == nil {
		t.Fatal("Expected overflow")
	}
	// The overflow should be a Box containing both the overflow text and txt2.
	overflowBox, ok := result.Overflow.(*document.Box)
	if !ok {
		t.Fatalf("Overflow is not a *Box, got %T", result.Overflow)
	}
	if len(overflowBox.Content) < 2 {
		t.Errorf("Overflow box should contain at least 2 children (overflow text + remaining), got %d", len(overflowBox.Content))
	}
}

func TestResolveBorderWidths(t *testing.T) {
	border := document.BorderEdges{
		Top:    document.BorderSide{Width: document.Pt(3), Style: document.BorderSolid},
		Right:  document.BorderSide{Width: document.Pt(0), Style: document.BorderNone},
		Bottom: document.BorderSide{Width: document.Pt(2), Style: document.BorderDashed},
		Left:   document.BorderSide{Width: document.Pt(1), Style: document.BorderDotted},
	}
	re := resolveBorderWidths(border, 500, 12)
	if re.Top != 3 {
		t.Errorf("Top = %v, want 3", re.Top)
	}
	if re.Right != 0 {
		t.Errorf("Right = %v, want 0 (BorderNone)", re.Right)
	}
	if re.Bottom != 2 {
		t.Errorf("Bottom = %v, want 2", re.Bottom)
	}
	if re.Left != 1 {
		t.Errorf("Left = %v, want 1", re.Left)
	}
}

func TestCreateOverflowNode_Empty(t *testing.T) {
	parent := &document.Box{}
	result := createOverflowNode(parent, nil)
	if result != nil {
		t.Error("createOverflowNode with nil should return nil")
	}
	result = createOverflowNode(parent, []document.DocumentNode{})
	if result != nil {
		t.Error("createOverflowNode with empty slice should return nil")
	}
}

func TestCreateOverflowNode_NonEmpty(t *testing.T) {
	parent := &document.Box{
		BoxStyle: document.BoxStyle{
			Margin:  document.UniformEdges(document.Pt(10)),
			Padding: document.UniformEdges(document.Pt(5)),
		},
	}
	remaining := []document.DocumentNode{
		&document.Text{Content: "remaining"},
	}
	result := createOverflowNode(parent, remaining)
	if result == nil {
		t.Fatal("createOverflowNode should return non-nil")
	}
	box, ok := result.(*document.Box)
	if !ok {
		t.Fatalf("overflow should be *Box, got %T", result)
	}
	if len(box.Content) != 1 {
		t.Errorf("overflow box should have 1 child, got %d", len(box.Content))
	}
	// Should preserve parent's margin and padding.
	if box.BoxStyle.Margin != parent.BoxStyle.Margin {
		t.Error("Overflow box should preserve parent margin")
	}
	if box.BoxStyle.Padding != parent.BoxStyle.Padding {
		t.Error("Overflow box should preserve parent padding")
	}
}

// ---------------------------------------------------------------------------
// flow.go tests
// ---------------------------------------------------------------------------

func TestFlowLayoutEmptyText(t *testing.T) {
	fl := &FlowLayout{}
	style := document.DefaultStyle()
	constraints := Constraints{
		AvailableWidth:  500,
		AvailableHeight: 700,
		FontResolver:    &mockFontResolver{},
	}
	result := fl.LayoutText("", style, constraints)
	if result.Bounds.Height != 0 {
		t.Errorf("Empty text height = %v, want 0", result.Bounds.Height)
	}
	if len(result.Children) != 0 {
		t.Errorf("Empty text children = %d, want 0", len(result.Children))
	}
}

func TestFlowLayoutShortText(t *testing.T) {
	fl := &FlowLayout{}
	style := document.Style{
		FontSize:   12,
		LineHeight: 1.2,
		TextAlign:  document.AlignLeft,
	}
	constraints := Constraints{
		AvailableWidth:  500,
		AvailableHeight: 700,
		FontResolver:    &mockFontResolver{},
	}
	result := fl.LayoutText("Hello", style, constraints)
	if result.Overflow != nil {
		t.Error("Short text should not overflow")
	}
	if len(result.Children) != 1 {
		t.Fatalf("Expected 1 line, got %d", len(result.Children))
	}
	lineSpacing := 12.0 * 1.2
	if !approxEqual(result.Bounds.Height, lineSpacing, 0.001) {
		t.Errorf("Height = %v, want %v", result.Bounds.Height, lineSpacing)
	}
}

func TestFlowLayoutWrapping(t *testing.T) {
	fl := &FlowLayout{}
	style := document.Style{
		FontSize:   12,
		LineHeight: 1.2,
		TextAlign:  document.AlignLeft,
	}
	// With fontSize=12, avgCharWidth=6. Width=60 => 10 chars per line.
	constraints := Constraints{
		AvailableWidth:  60,
		AvailableHeight: 700,
		FontResolver:    &mockFontResolver{},
	}
	result := fl.LayoutText("Hello world this is a test", style, constraints)
	if len(result.Children) < 2 {
		t.Fatalf("Expected multiple lines, got %d", len(result.Children))
	}
}

func TestFlowLayoutTextOverflow(t *testing.T) {
	fl := &FlowLayout{}
	style := document.Style{
		FontSize:   12,
		LineHeight: 1.5,
	}
	lineSpacing := 12.0 * 1.5 // 18
	constraints := Constraints{
		AvailableWidth:  60,
		AvailableHeight: lineSpacing * 1.5, // Enough for about 1 line
		FontResolver:    &mockFontResolver{},
	}
	result := fl.LayoutText("Hello world this is a longer text that wraps", style, constraints)
	if result.Overflow == nil {
		t.Error("Expected overflow with limited height")
	}
	// Overflow should be a Text node with the remaining text.
	overflow, ok := result.Overflow.(*document.Text)
	if !ok {
		t.Fatalf("Overflow is not *Text, got %T", result.Overflow)
	}
	if overflow.Content == "" {
		t.Error("Overflow content should not be empty")
	}
}

func TestFlowLayoutWithoutFontResolver(t *testing.T) {
	fl := &FlowLayout{}
	style := document.Style{
		FontSize:   12,
		LineHeight: 1.2,
	}
	constraints := Constraints{
		AvailableWidth:  60,
		AvailableHeight: 700,
		FontResolver:    nil, // No font resolver
	}
	result := fl.LayoutText("Hello world this is a test", style, constraints)
	// Should use approximateLineBreak fallback.
	if len(result.Children) < 1 {
		t.Error("Expected at least 1 line")
	}
}

func TestFlowLayoutDefaultFontSize(t *testing.T) {
	fl := &FlowLayout{}
	style := document.Style{
		FontSize:   0, // Will default to 12
		LineHeight: 0, // Will default to 1.2
	}
	constraints := Constraints{
		AvailableWidth:  500,
		AvailableHeight: 700,
		FontResolver:    &mockFontResolver{},
	}
	result := fl.LayoutText("Hello", style, constraints)
	// Default: lineSpacing = 12 * 1.2 = 14.4
	expectedHeight := 12.0 * 1.2
	if !approxEqual(result.Bounds.Height, expectedHeight, 0.001) {
		t.Errorf("Height = %v, want %v", result.Bounds.Height, expectedHeight)
	}
}

func TestFlowLayoutAlignCenter(t *testing.T) {
	fl := &FlowLayout{}
	style := document.Style{
		FontSize:   12,
		LineHeight: 1.2,
		TextAlign:  document.AlignCenter,
	}
	constraints := Constraints{
		AvailableWidth:  500,
		AvailableHeight: 700,
		FontResolver:    &mockFontResolver{},
	}
	result := fl.LayoutText("Hi", style, constraints)
	if len(result.Children) != 1 {
		t.Fatalf("Expected 1 line, got %d", len(result.Children))
	}
	// "Hi" = 2 chars * 6pt = 12pt wide. Center = (500-12)/2 = 244
	child := result.Children[0]
	expectedX := (500.0 - 12.0) / 2
	if child.Position.X != expectedX {
		t.Errorf("Center X = %v, want %v", child.Position.X, expectedX)
	}
}

func TestFlowLayoutAlignRight(t *testing.T) {
	fl := &FlowLayout{}
	style := document.Style{
		FontSize:   12,
		LineHeight: 1.2,
		TextAlign:  document.AlignRight,
	}
	constraints := Constraints{
		AvailableWidth:  500,
		AvailableHeight: 700,
		FontResolver:    &mockFontResolver{},
	}
	result := fl.LayoutText("Hi", style, constraints)
	if len(result.Children) != 1 {
		t.Fatalf("Expected 1 line, got %d", len(result.Children))
	}
	// "Hi" = 2 chars * 6pt = 12pt wide. Right = 500-12 = 488
	child := result.Children[0]
	expectedX := 500.0 - 12.0
	if child.Position.X != expectedX {
		t.Errorf("Right X = %v, want %v", child.Position.X, expectedX)
	}
}

func TestFlowLayoutAlignJustify(t *testing.T) {
	fl := &FlowLayout{}
	style := document.Style{
		FontSize:   12,
		LineHeight: 1.2,
		TextAlign:  document.AlignJustify,
	}
	constraints := Constraints{
		AvailableWidth:  500,
		AvailableHeight: 700,
		FontResolver:    &mockFontResolver{},
	}
	result := fl.LayoutText("Hi", style, constraints)
	if len(result.Children) != 1 {
		t.Fatalf("Expected 1 line, got %d", len(result.Children))
	}
	// Justify starts at X=0 (same as left).
	child := result.Children[0]
	if child.Position.X != 0 {
		t.Errorf("Justify X = %v, want 0", child.Position.X)
	}
	// Single line (last line) should have no word spacing.
	textNode := child.Node.(*document.Text)
	if textNode.TextStyle.WordSpacing != 0 {
		t.Errorf("Last line WordSpacing = %v, want 0", textNode.TextStyle.WordSpacing)
	}
}

func TestFlowLayoutJustifyMultiLine(t *testing.T) {
	fl := &FlowLayout{}
	style := document.Style{
		FontSize:   12,
		LineHeight: 1.2,
		TextAlign:  document.AlignJustify,
	}
	// Mock: each char = 0.5*12 = 6pt. Width 60 => 10 chars per line.
	// "aa bb cc dd ee ff" → line 1: "aa bb cc" (8 chars, 2 spaces), line 2: "dd ee ff" (last).
	constraints := Constraints{
		AvailableWidth:  60,
		AvailableHeight: 700,
		FontResolver:    &mockFontResolver{},
	}
	result := fl.LayoutText("aa bb cc dd ee ff", style, constraints)
	if len(result.Children) != 2 {
		t.Fatalf("Expected 2 lines, got %d", len(result.Children))
	}

	// First line: "aa bb cc" → 8 chars → width = 48pt, extra = 12pt, 2 spaces → WordSpacing = 6.
	first := result.Children[0].Node.(*document.Text)
	if !approxEqual(first.TextStyle.WordSpacing, 6, 0.01) {
		t.Errorf("Line 0 WordSpacing = %v, want 6", first.TextStyle.WordSpacing)
	}
	// Justified line should span full available width.
	if !approxEqual(result.Children[0].Size.Width, 60, 0.01) {
		t.Errorf("Line 0 Width = %v, want 60", result.Children[0].Size.Width)
	}

	// Last line should have no word spacing (left-aligned).
	last := result.Children[1].Node.(*document.Text)
	if last.TextStyle.WordSpacing != 0 {
		t.Errorf("Last line WordSpacing = %v, want 0", last.TextStyle.WordSpacing)
	}
}

func TestFlowLayoutJustifyNoSpaces(t *testing.T) {
	fl := &FlowLayout{}
	style := document.Style{
		FontSize:   12,
		LineHeight: 1.2,
		TextAlign:  document.AlignJustify,
	}
	// Single word per line — no spaces to distribute, WordSpacing stays 0.
	constraints := Constraints{
		AvailableWidth:  30, // ~5 chars per line
		AvailableHeight: 700,
		FontResolver:    &mockFontResolver{},
	}
	result := fl.LayoutText("abcdefghij", style, constraints)
	for i, child := range result.Children {
		textNode := child.Node.(*document.Text)
		if textNode.TextStyle.WordSpacing != 0 {
			t.Errorf("Line %d WordSpacing = %v, want 0 (no spaces)", i, textNode.TextStyle.WordSpacing)
		}
	}
}

func TestFlowLayoutItalicFont(t *testing.T) {
	fl := &FlowLayout{}
	style := document.Style{
		FontSize:   12,
		LineHeight: 1.2,
		FontStyle:  document.StyleItalic,
	}
	constraints := Constraints{
		AvailableWidth:  500,
		AvailableHeight: 700,
		FontResolver:    &mockFontResolver{},
	}
	result := fl.LayoutText("italic text", style, constraints)
	if result.Overflow != nil {
		t.Error("Should not overflow")
	}
	if len(result.Children) < 1 {
		t.Error("Expected at least 1 line")
	}
}

// ---------------------------------------------------------------------------
// LetterSpacing tests (WP1)
// ---------------------------------------------------------------------------

func TestFlowLayoutLetterSpacingWidth(t *testing.T) {
	fl := &FlowLayout{}
	// "Hello" = 5 chars, fontSize=12, charWidth=6 each => baseWidth = 30.
	// LetterSpacing=2, 4 gaps => extra = 8.  Total lineWidth = 38.
	style := document.Style{
		FontSize:      12,
		LineHeight:    1.2,
		LetterSpacing: 2,
	}
	constraints := Constraints{
		AvailableWidth:  500,
		AvailableHeight: 700,
		FontResolver:    &mockFontResolver{},
	}
	result := fl.LayoutText("Hello", style, constraints)
	if len(result.Children) != 1 {
		t.Fatalf("Expected 1 line, got %d", len(result.Children))
	}
	// With center alignment we can infer the width indirectly,
	// but let's check the placed node size directly.
	pn := result.Children[0]
	// base = 5 * 12 * 0.5 = 30, letterSpacing = 2*(5-1) = 8, total = 38
	expectedWidth := 38.0
	if !approxEqual(pn.Size.Width, expectedWidth, 0.01) {
		t.Errorf("Line width = %v, want %v", pn.Size.Width, expectedWidth)
	}
}

func TestFlowLayoutLetterSpacingSingleChar(t *testing.T) {
	fl := &FlowLayout{}
	// Single character: no gaps, so letter spacing should not add width.
	// "A" = 1 char, fontSize=12, charWidth=6 => baseWidth = 6.
	// LetterSpacing=10, but only 0 gaps => no extra. Total = 6.
	style := document.Style{
		FontSize:      12,
		LineHeight:    1.2,
		LetterSpacing: 10,
	}
	constraints := Constraints{
		AvailableWidth:  500,
		AvailableHeight: 700,
		FontResolver:    &mockFontResolver{},
	}
	result := fl.LayoutText("A", style, constraints)
	if len(result.Children) != 1 {
		t.Fatalf("Expected 1 line, got %d", len(result.Children))
	}
	pn := result.Children[0]
	expectedWidth := 6.0 // single char, no letter-spacing contribution
	if !approxEqual(pn.Size.Width, expectedWidth, 0.01) {
		t.Errorf("Line width = %v, want %v", pn.Size.Width, expectedWidth)
	}
}

// ---------------------------------------------------------------------------
// TextIndent tests (WP2)
// ---------------------------------------------------------------------------

func TestFlowLayoutTextIndentXOffset(t *testing.T) {
	fl := &FlowLayout{}
	style := document.Style{
		FontSize:   12,
		LineHeight: 1.2,
		TextIndent: document.Pt(24),
	}
	constraints := Constraints{
		AvailableWidth:  500,
		AvailableHeight: 700,
		FontResolver:    &mockFontResolver{},
	}
	// "Hello world" fits in one line; first line should have X offset = 24.
	result := fl.LayoutText("Hello world", style, constraints)
	if len(result.Children) < 1 {
		t.Fatal("Expected at least 1 line")
	}
	if !approxEqual(result.Children[0].Position.X, 24, 0.01) {
		t.Errorf("First line X = %v, want 24", result.Children[0].Position.X)
	}
}

func TestFlowLayoutTextIndentSecondLineNoOffset(t *testing.T) {
	fl := &FlowLayout{}
	// Use narrow width to force wrapping.
	// charWidth=6, "Hello world test" = 16 chars * 6 = 96.
	// Avail=60, indent=12 => first line avail=48 => ~8 chars.
	// Subsequent lines avail=60 => ~10 chars.
	style := document.Style{
		FontSize:   12,
		LineHeight: 1.2,
		TextIndent: document.Pt(12),
	}
	constraints := Constraints{
		AvailableWidth:  60,
		AvailableHeight: 700,
		FontResolver:    &mockFontResolver{},
	}
	result := fl.LayoutText("Hello world test line", style, constraints)
	if len(result.Children) < 2 {
		t.Fatalf("Expected at least 2 lines, got %d", len(result.Children))
	}
	// First line should have indent.
	if !approxEqual(result.Children[0].Position.X, 12, 0.01) {
		t.Errorf("First line X = %v, want 12", result.Children[0].Position.X)
	}
	// Second line should have no indent (X = 0 for left-aligned).
	if !approxEqual(result.Children[1].Position.X, 0, 0.01) {
		t.Errorf("Second line X = %v, want 0", result.Children[1].Position.X)
	}
}

func TestFlowLayoutTextIndentSingleLine(t *testing.T) {
	fl := &FlowLayout{}
	style := document.Style{
		FontSize:   12,
		LineHeight: 1.2,
		TextIndent: document.Pt(30),
	}
	constraints := Constraints{
		AvailableWidth:  500,
		AvailableHeight: 700,
		FontResolver:    &mockFontResolver{},
	}
	// Short text on a single line; indent should still apply.
	result := fl.LayoutText("Hi", style, constraints)
	if len(result.Children) != 1 {
		t.Fatalf("Expected 1 line, got %d", len(result.Children))
	}
	if !approxEqual(result.Children[0].Position.X, 30, 0.01) {
		t.Errorf("Single line X = %v, want 30", result.Children[0].Position.X)
	}
}

func TestApproximateLineBreak(t *testing.T) {
	// fontSize=12 => avgCharWidth=6. maxWidth=60 => 10 chars per line.
	lines := approximateLineBreak("Hello world this is test", 12, 60)
	if len(lines) < 2 {
		t.Fatalf("Expected multiple lines, got %d: %v", len(lines), lines)
	}
	// Verify no line exceeds ~10 characters (except possible mid-word break).
	for i, line := range lines {
		if len([]rune(line)) > 12 { // Allow some slack for mid-word breaks
			t.Errorf("Line %d too long: %q (%d runes)", i, line, len([]rune(line)))
		}
	}
}

func TestApproximateLineBreak_ShortText(t *testing.T) {
	lines := approximateLineBreak("Hi", 12, 500)
	if len(lines) != 1 {
		t.Fatalf("Expected 1 line, got %d", len(lines))
	}
	if lines[0] != "Hi" {
		t.Errorf("Line = %q, want %q", lines[0], "Hi")
	}
}

func TestApproximateLineBreak_ZeroFontSize(t *testing.T) {
	lines := approximateLineBreak("Hello", 0, 500)
	if len(lines) != 1 {
		t.Fatalf("Expected 1 line, got %d", len(lines))
	}
	if lines[0] != "Hello" {
		t.Errorf("Line = %q, want %q", lines[0], "Hello")
	}
}

func TestApproximateLineBreak_VeryNarrow(t *testing.T) {
	// maxWidth=3, fontSize=12, avgCharWidth=6 => charsPerLine = 0 => forced to 1
	lines := approximateLineBreak("abc", 12, 3)
	if len(lines) < 1 {
		t.Fatal("Expected at least 1 line")
	}
}

func TestApproximateLineBreak_ForcedMidWordBreak(t *testing.T) {
	// fontSize=12, avgCharWidth=6, maxWidth=18 => 3 chars per line
	// "abcdef" has no spaces, so must break mid-word.
	lines := approximateLineBreak("abcdef", 12, 18)
	if len(lines) < 2 {
		t.Fatalf("Expected at least 2 lines, got %d: %v", len(lines), lines)
	}
}

func TestJoinLines(t *testing.T) {
	tests := []struct {
		input []string
		want  string
	}{
		{nil, ""},
		{[]string{}, ""},
		{[]string{"hello"}, "hello"},
		{[]string{"hello", "world"}, "hello world"},
		{[]string{"a", "b", "c"}, "a b c"},
	}
	for _, tt := range tests {
		got := joinLines(tt.input)
		if got != tt.want {
			t.Errorf("joinLines(%v) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestAlignTextX(t *testing.T) {
	tests := []struct {
		align     document.TextAlign
		lineWidth float64
		available float64
		want      float64
	}{
		{document.AlignLeft, 100, 500, 0},
		{document.AlignCenter, 100, 500, 200},
		{document.AlignRight, 100, 500, 400},
		{document.AlignJustify, 100, 500, 0},
	}
	for _, tt := range tests {
		got := alignTextX(tt.align, tt.lineWidth, tt.available)
		if got != tt.want {
			t.Errorf("alignTextX(%v, %v, %v) = %v, want %v", tt.align, tt.lineWidth, tt.available, got, tt.want)
		}
	}
}

// ---------------------------------------------------------------------------
// paging.go tests
// ---------------------------------------------------------------------------

func TestNewPaginator(t *testing.T) {
	p := NewPaginator(document.A4, document.UniformEdges(document.Pt(72)), &mockFontResolver{})
	if p == nil {
		t.Fatal("NewPaginator returned nil")
	}
}

func TestPaginatorNilDocument(t *testing.T) {
	p := NewPaginator(document.A4, document.UniformEdges(document.Pt(72)), &mockFontResolver{})
	pages := p.Paginate(nil)
	if pages != nil {
		t.Errorf("Paginate(nil) = %v, want nil", pages)
	}
}

func TestPaginatorEmptyDocument(t *testing.T) {
	p := NewPaginator(document.A4, document.UniformEdges(document.Pt(72)), &mockFontResolver{})
	doc := &document.Document{}
	pages := p.Paginate(doc)
	if len(pages) != 0 {
		t.Errorf("Paginate empty doc = %d pages, want 0", len(pages))
	}
}

func TestPaginatorSinglePage(t *testing.T) {
	p := NewPaginator(document.A4, document.UniformEdges(document.Pt(72)), &mockFontResolver{})
	doc := &document.Document{
		Pages: []*document.Page{
			{
				Size: document.A4,
				Content: []document.DocumentNode{
					&document.Text{
						Content:   "Hello World",
						TextStyle: document.Style{FontSize: 12, LineHeight: 1.2},
					},
				},
			},
		},
	}
	pages := p.Paginate(doc)
	if len(pages) < 1 {
		t.Fatalf("Expected at least 1 page, got %d", len(pages))
	}
	if pages[0].Size != document.A4 {
		t.Errorf("Page size = %+v, want A4", pages[0].Size)
	}
}

func TestPaginatorPageMarginOffset(t *testing.T) {
	margin := document.UniformEdges(document.Pt(50))
	p := NewPaginator(document.A4, margin, &mockFontResolver{})
	doc := &document.Document{
		Pages: []*document.Page{
			{
				Size: document.A4,
				Content: []document.DocumentNode{
					&document.Text{
						Content:   "Hello",
						TextStyle: document.Style{FontSize: 12, LineHeight: 1.2},
					},
				},
			},
		},
	}
	pages := p.Paginate(doc)
	if len(pages) < 1 {
		t.Fatal("Expected at least 1 page")
	}
	// All children should be offset by the margin.
	for _, child := range pages[0].Children {
		if child.Position.X < 50 {
			t.Errorf("Child X = %v, want >= 50 (margin offset)", child.Position.X)
		}
		if child.Position.Y < 50 {
			t.Errorf("Child Y = %v, want >= 50 (margin offset)", child.Position.Y)
		}
	}
}

func TestPaginatorDefaultPageSize(t *testing.T) {
	p := NewPaginator(document.Letter, document.UniformEdges(document.Pt(36)), &mockFontResolver{})
	doc := &document.Document{
		Pages: []*document.Page{
			{
				// Size is zero-valued, so paginator should use default.
				Content: []document.DocumentNode{
					&document.Text{
						Content:   "test",
						TextStyle: document.Style{FontSize: 12, LineHeight: 1.2},
					},
				},
			},
		},
	}
	pages := p.Paginate(doc)
	if len(pages) < 1 {
		t.Fatal("Expected at least 1 page")
	}
	if pages[0].Size != document.Letter {
		t.Errorf("Page size = %+v, want Letter", pages[0].Size)
	}
}

func TestPaginatorDefaultMargins(t *testing.T) {
	defaultMargins := document.UniformEdges(document.Pt(72))
	p := NewPaginator(document.A4, defaultMargins, &mockFontResolver{})
	doc := &document.Document{
		Pages: []*document.Page{
			{
				Size: document.A4,
				// Margins zero-valued, so paginator should use default.
				Content: []document.DocumentNode{
					&document.Text{
						Content:   "test",
						TextStyle: document.Style{FontSize: 12, LineHeight: 1.2},
					},
				},
			},
		},
	}
	pages := p.Paginate(doc)
	if len(pages) < 1 {
		t.Fatal("Expected at least 1 page")
	}
	// Children should be offset by the default margin (72 pt).
	for _, child := range pages[0].Children {
		if child.Position.X < 72 {
			t.Errorf("Child X = %v, want >= 72 (default margin)", child.Position.X)
		}
	}
}

func TestPaginatorOverflowCreatesMultiplePages(t *testing.T) {
	p := NewPaginator(document.A4, document.UniformEdges(document.Pt(72)), &mockFontResolver{})
	// Create enough text to overflow onto multiple pages.
	var content []document.DocumentNode
	for i := 0; i < 200; i++ {
		content = append(content, &document.Text{
			Content:   "This is a line of text that is repeated many times to test pagination overflow behavior.",
			TextStyle: document.Style{FontSize: 14, LineHeight: 1.5},
		})
	}
	doc := &document.Document{
		Pages: []*document.Page{
			{
				Size:    document.A4,
				Content: content,
			},
		},
	}
	pages := p.Paginate(doc)
	if len(pages) < 2 {
		t.Errorf("Expected at least 2 pages from overflow, got %d", len(pages))
	}
}

func TestOffsetNodes(t *testing.T) {
	nodes := []PlacedNode{
		{
			Node:     &document.Text{Content: "a"},
			Position: document.Point{X: 10, Y: 20},
			Size:     document.Size{Width: 100, Height: 14},
		},
		{
			Node:     &document.Text{Content: "b"},
			Position: document.Point{X: 10, Y: 34},
			Size:     document.Size{Width: 100, Height: 14},
		},
	}
	shifted := offsetNodes(nodes, 50, 72)
	if shifted[0].Position.X != 60 {
		t.Errorf("shifted[0].X = %v, want 60", shifted[0].Position.X)
	}
	if shifted[0].Position.Y != 92 {
		t.Errorf("shifted[0].Y = %v, want 92", shifted[0].Position.Y)
	}
	if shifted[1].Position.X != 60 {
		t.Errorf("shifted[1].X = %v, want 60", shifted[1].Position.X)
	}
	if shifted[1].Position.Y != 106 {
		t.Errorf("shifted[1].Y = %v, want 106", shifted[1].Position.Y)
	}
}

func TestOffsetNodesEmpty(t *testing.T) {
	result := offsetNodes(nil, 50, 72)
	if result != nil {
		t.Errorf("offsetNodes(nil) = %v, want nil", result)
	}
	result = offsetNodes([]PlacedNode{}, 50, 72)
	if len(result) != 0 {
		t.Errorf("offsetNodes([]) length = %d, want 0", len(result))
	}
}

func TestPageLayoutFields(t *testing.T) {
	pl := PageLayout{
		Size: document.A4,
		Children: []PlacedNode{
			{Node: &document.Text{Content: "test"}},
		},
	}
	if pl.Size != document.A4 {
		t.Error("PageLayout.Size mismatch")
	}
	if len(pl.Children) != 1 {
		t.Errorf("PageLayout.Children length = %d, want 1", len(pl.Children))
	}
}

func TestBlockLayoutEngineInterface(t *testing.T) {
	// Verify BlockLayout implements Engine.
	var _ Engine = &BlockLayout{}
}

func TestMockFontResolverResolve(t *testing.T) {
	m := &mockFontResolver{}
	rf := m.Resolve("Arial", document.WeightBold, true)
	if rf.ID != "Arial-Bold-Italic" {
		t.Errorf("ID = %v, want Arial-Bold-Italic", rf.ID)
	}
	rf = m.Resolve("", document.WeightNormal, false)
	if rf.ID != "MockFont" {
		t.Errorf("ID = %v, want MockFont", rf.ID)
	}
}

func TestPaginatorMultipleInputPages(t *testing.T) {
	p := NewPaginator(document.A4, document.UniformEdges(document.Pt(36)), &mockFontResolver{})
	doc := &document.Document{
		Pages: []*document.Page{
			{
				Size: document.A4,
				Content: []document.DocumentNode{
					&document.Text{
						Content:   "Page 1",
						TextStyle: document.Style{FontSize: 12, LineHeight: 1.2},
					},
				},
			},
			{
				Size: document.Letter,
				Content: []document.DocumentNode{
					&document.Text{
						Content:   "Page 2",
						TextStyle: document.Style{FontSize: 12, LineHeight: 1.2},
					},
				},
			},
		},
	}
	pages := p.Paginate(doc)
	if len(pages) < 2 {
		t.Fatalf("Expected at least 2 pages, got %d", len(pages))
	}
	if pages[0].Size != document.A4 {
		t.Errorf("Page 0 size = %+v, want A4", pages[0].Size)
	}
	if pages[1].Size != document.Letter {
		t.Errorf("Page 1 size = %+v, want Letter", pages[1].Size)
	}
}

// ---------------------------------------------------------------------------
// Horizontal layout tests (layoutHorizontal, resolveChildWidths,
// resolveFixedHeight, stretchPlacedNodes)
// ---------------------------------------------------------------------------

func TestLayoutHorizontal_BasicTwoColumns(t *testing.T) {
	bl := NewBlockLayout()
	col1 := &document.Box{
		Content: []document.DocumentNode{
			&document.Text{Content: "Left", TextStyle: document.Style{FontSize: 12, LineHeight: 1.2}},
		},
		BoxStyle: document.BoxStyle{Width: document.Pct(50)},
	}
	col2 := &document.Box{
		Content: []document.DocumentNode{
			&document.Text{Content: "Right", TextStyle: document.Style{FontSize: 12, LineHeight: 1.2}},
		},
		BoxStyle: document.BoxStyle{Width: document.Pct(50)},
	}
	row := &document.Box{
		Content:  []document.DocumentNode{col1, col2},
		BoxStyle: document.BoxStyle{Direction: document.DirectionHorizontal},
	}
	constraints := Constraints{
		AvailableWidth:  400,
		AvailableHeight: 700,
		FontResolver:    &mockFontResolver{},
	}
	result := bl.Layout(row, constraints)
	if result.Overflow != nil {
		t.Error("Should not overflow")
	}
	if len(result.Children) != 2 {
		t.Fatalf("Expected 2 children, got %d", len(result.Children))
	}
	// Each column should be 200pt wide (50% of 400).
	if !approxEqual(result.Children[0].Size.Width, 200, 0.1) {
		t.Errorf("col1 width = %v, want 200", result.Children[0].Size.Width)
	}
	if !approxEqual(result.Children[1].Size.Width, 200, 0.1) {
		t.Errorf("col2 width = %v, want 200", result.Children[1].Size.Width)
	}
	// Second column should be offset to the right.
	if !approxEqual(result.Children[1].Position.X, result.Children[0].Position.X+200, 0.1) {
		t.Errorf("col2 X = %v, want %v", result.Children[1].Position.X, result.Children[0].Position.X+200)
	}
}

func TestLayoutHorizontal_AutoWidthDistribution(t *testing.T) {
	bl := NewBlockLayout()
	// No explicit widths: all 3 children should share equally.
	children := make([]document.DocumentNode, 3)
	for i := range children {
		children[i] = &document.Box{
			Content: []document.DocumentNode{
				&document.Text{Content: "col", TextStyle: document.Style{FontSize: 12, LineHeight: 1.2}},
			},
		}
	}
	row := &document.Box{
		Content:  children,
		BoxStyle: document.BoxStyle{Direction: document.DirectionHorizontal},
	}
	constraints := Constraints{
		AvailableWidth:  300,
		AvailableHeight: 700,
		FontResolver:    &mockFontResolver{},
	}
	result := bl.Layout(row, constraints)
	if len(result.Children) != 3 {
		t.Fatalf("Expected 3 children, got %d", len(result.Children))
	}
	for i, child := range result.Children {
		if !approxEqual(child.Size.Width, 100, 0.1) {
			t.Errorf("child[%d] width = %v, want 100", i, child.Size.Width)
		}
	}
}

func TestLayoutHorizontal_MixedWidths(t *testing.T) {
	bl := NewBlockLayout()
	// First child: fixed 100pt, second: auto (gets remaining 200pt).
	col1 := &document.Box{
		Content: []document.DocumentNode{
			&document.Text{Content: "A", TextStyle: document.Style{FontSize: 12, LineHeight: 1.2}},
		},
		BoxStyle: document.BoxStyle{Width: document.Pt(100)},
	}
	col2 := &document.Box{
		Content: []document.DocumentNode{
			&document.Text{Content: "B", TextStyle: document.Style{FontSize: 12, LineHeight: 1.2}},
		},
	}
	row := &document.Box{
		Content:  []document.DocumentNode{col1, col2},
		BoxStyle: document.BoxStyle{Direction: document.DirectionHorizontal},
	}
	constraints := Constraints{
		AvailableWidth:  300,
		AvailableHeight: 700,
		FontResolver:    &mockFontResolver{},
	}
	result := bl.Layout(row, constraints)
	if len(result.Children) != 2 {
		t.Fatalf("Expected 2 children, got %d", len(result.Children))
	}
	if !approxEqual(result.Children[0].Size.Width, 100, 0.1) {
		t.Errorf("col1 width = %v, want 100", result.Children[0].Size.Width)
	}
	if !approxEqual(result.Children[1].Size.Width, 200, 0.1) {
		t.Errorf("col2 width = %v, want 200", result.Children[1].Size.Width)
	}
}

func TestLayoutHorizontal_WithPaddingAndMargin(t *testing.T) {
	bl := NewBlockLayout()
	col := &document.Box{
		Content: []document.DocumentNode{
			&document.Text{Content: "X", TextStyle: document.Style{FontSize: 12, LineHeight: 1.2}},
		},
	}
	row := &document.Box{
		Content: []document.DocumentNode{col},
		BoxStyle: document.BoxStyle{
			Direction: document.DirectionHorizontal,
			Margin:    document.UniformEdges(document.Pt(10)),
			Padding:   document.UniformEdges(document.Pt(5)),
		},
	}
	constraints := Constraints{
		AvailableWidth:  400,
		AvailableHeight: 700,
		FontResolver:    &mockFontResolver{},
	}
	result := bl.Layout(row, constraints)
	if len(result.Children) != 1 {
		t.Fatalf("Expected 1 child, got %d", len(result.Children))
	}
	// Content origin: margin(10) + padding(5) = 15
	if result.Children[0].Position.X != 15 {
		t.Errorf("child X = %v, want 15", result.Children[0].Position.X)
	}
	if result.Children[0].Position.Y != 15 {
		t.Errorf("child Y = %v, want 15", result.Children[0].Position.Y)
	}
}

func TestLayoutHorizontal_StretchesToTallestChild(t *testing.T) {
	bl := NewBlockLayout()
	// col1 has 1 line, col2 has 2 lines => col2 is taller.
	col1 := &document.Box{
		Content: []document.DocumentNode{
			&document.Text{Content: "Short", TextStyle: document.Style{FontSize: 12, LineHeight: 1.2}},
		},
		BoxStyle: document.BoxStyle{Width: document.Pct(50)},
	}
	col2 := &document.Box{
		Content: []document.DocumentNode{
			&document.Text{Content: "Line1", TextStyle: document.Style{FontSize: 12, LineHeight: 1.2}},
			&document.Text{Content: "Line2", TextStyle: document.Style{FontSize: 12, LineHeight: 1.2}},
		},
		BoxStyle: document.BoxStyle{Width: document.Pct(50)},
	}
	row := &document.Box{
		Content:  []document.DocumentNode{col1, col2},
		BoxStyle: document.BoxStyle{Direction: document.DirectionHorizontal},
	}
	constraints := Constraints{
		AvailableWidth:  400,
		AvailableHeight: 700,
		FontResolver:    &mockFontResolver{},
	}
	result := bl.Layout(row, constraints)
	if len(result.Children) != 2 {
		t.Fatalf("Expected 2 children, got %d", len(result.Children))
	}
	// Both columns should have the same height (stretched to tallest).
	if result.Children[0].Size.Height != result.Children[1].Size.Height {
		t.Errorf("heights should match: col1=%v, col2=%v", result.Children[0].Size.Height, result.Children[1].Size.Height)
	}
}

func TestLayoutHorizontal_NilChildSkipped(t *testing.T) {
	bl := NewBlockLayout()
	col := &document.Box{
		Content: []document.DocumentNode{
			&document.Text{Content: "ok", TextStyle: document.Style{FontSize: 12, LineHeight: 1.2}},
		},
	}
	row := &document.Box{
		Content:  []document.DocumentNode{nil, col},
		BoxStyle: document.BoxStyle{Direction: document.DirectionHorizontal},
	}
	constraints := Constraints{
		AvailableWidth:  300,
		AvailableHeight: 700,
		FontResolver:    &mockFontResolver{},
	}
	result := bl.Layout(row, constraints)
	if len(result.Children) != 1 {
		t.Errorf("Expected 1 child (nil skipped), got %d", len(result.Children))
	}
}

func TestLayoutHorizontal_WithFixedHeight(t *testing.T) {
	bl := NewBlockLayout()
	col := &document.Box{
		Content: []document.DocumentNode{
			&document.Text{Content: "A", TextStyle: document.Style{FontSize: 12, LineHeight: 1.2}},
		},
	}
	row := &document.Box{
		Content: []document.DocumentNode{col},
		BoxStyle: document.BoxStyle{
			Direction: document.DirectionHorizontal,
			Height:    document.Pt(100),
		},
	}
	constraints := Constraints{
		AvailableWidth:  400,
		AvailableHeight: 700,
		FontResolver:    &mockFontResolver{},
	}
	result := bl.Layout(row, constraints)
	// Total height should incorporate the fixed height.
	if result.Bounds.Height < 100 {
		t.Errorf("Bounds.Height = %v, want >= 100", result.Bounds.Height)
	}
}

func TestLayoutHorizontal_DefaultFontSize(t *testing.T) {
	bl := NewBlockLayout()
	col := &document.Box{
		Content: []document.DocumentNode{
			&document.Text{Content: "A", TextStyle: document.Style{FontSize: 12, LineHeight: 1.2}},
		},
	}
	row := &document.Box{
		Content: []document.DocumentNode{col},
		BoxStyle: document.BoxStyle{
			Direction: document.DirectionHorizontal,
		},
	}
	// FontSize in the box style is 0, should default to 12.
	constraints := Constraints{
		AvailableWidth:  400,
		AvailableHeight: 700,
		FontResolver:    &mockFontResolver{},
	}
	result := bl.Layout(row, constraints)
	if result.Bounds.Width != 400 {
		t.Errorf("Bounds.Width = %v, want 400", result.Bounds.Width)
	}
}

func TestResolveChildWidths_AllExplicit(t *testing.T) {
	children := []document.DocumentNode{
		&document.Box{BoxStyle: document.BoxStyle{Width: document.Pt(100)}},
		&document.Box{BoxStyle: document.BoxStyle{Width: document.Pt(200)}},
	}
	widths := resolveChildWidths(children, 500, 12)
	if !approxEqual(widths[0], 100, 0.1) {
		t.Errorf("widths[0] = %v, want 100", widths[0])
	}
	if !approxEqual(widths[1], 200, 0.1) {
		t.Errorf("widths[1] = %v, want 200", widths[1])
	}
}

func TestResolveChildWidths_AllAuto(t *testing.T) {
	children := []document.DocumentNode{
		&document.Box{},
		&document.Box{},
	}
	widths := resolveChildWidths(children, 400, 12)
	if !approxEqual(widths[0], 200, 0.1) {
		t.Errorf("widths[0] = %v, want 200", widths[0])
	}
	if !approxEqual(widths[1], 200, 0.1) {
		t.Errorf("widths[1] = %v, want 200", widths[1])
	}
}

func TestResolveChildWidths_NilChild(t *testing.T) {
	children := []document.DocumentNode{
		nil,
		&document.Box{},
	}
	widths := resolveChildWidths(children, 300, 12)
	if widths[0] != 0 {
		t.Errorf("nil child width = %v, want 0", widths[0])
	}
	if !approxEqual(widths[1], 300, 0.1) {
		t.Errorf("widths[1] = %v, want 300", widths[1])
	}
}

func TestResolveChildWidths_ExplicitExceedsParent(t *testing.T) {
	children := []document.DocumentNode{
		&document.Box{BoxStyle: document.BoxStyle{Width: document.Pt(400)}},
		&document.Box{}, // auto: remaining = max(0, 300-400) = 0
	}
	widths := resolveChildWidths(children, 300, 12)
	if !approxEqual(widths[0], 400, 0.1) {
		t.Errorf("widths[0] = %v, want 400", widths[0])
	}
	if widths[1] != 0 {
		t.Errorf("widths[1] = %v, want 0 (no remaining space)", widths[1])
	}
}

func TestResolveChildWidths_TextChild(t *testing.T) {
	// Non-Box children are treated as auto-width.
	children := []document.DocumentNode{
		&document.Text{Content: "txt", TextStyle: document.Style{FontSize: 12}},
	}
	widths := resolveChildWidths(children, 400, 12)
	if !approxEqual(widths[0], 400, 0.1) {
		t.Errorf("widths[0] = %v, want 400", widths[0])
	}
}

func TestResolveFixedHeight_BoxWithFixedHeight(t *testing.T) {
	box := &document.Box{
		BoxStyle: document.BoxStyle{Height: document.Pt(200)},
	}
	margin := document.ResolvedEdges{Top: 10, Bottom: 10}
	padding := document.ResolvedEdges{Top: 5, Bottom: 5}
	border := document.ResolvedEdges{Top: 1, Bottom: 1}
	h := resolveFixedHeight(box, 500, 12, margin, padding, border)
	// 200 - (10+10) - (5+5) - (1+1) = 168
	if !approxEqual(h, 168, 0.1) {
		t.Errorf("fixedHeight = %v, want 168", h)
	}
}

func TestResolveFixedHeight_NoHeight(t *testing.T) {
	box := &document.Box{} // Height is Auto by default.
	h := resolveFixedHeight(box, 500, 12, document.ResolvedEdges{}, document.ResolvedEdges{}, document.ResolvedEdges{})
	if h != -1 {
		t.Errorf("fixedHeight = %v, want -1", h)
	}
}

func TestResolveFixedHeight_NonBox(t *testing.T) {
	txt := &document.Text{Content: "text"}
	h := resolveFixedHeight(txt, 500, 12, document.ResolvedEdges{}, document.ResolvedEdges{}, document.ResolvedEdges{})
	if h != -1 {
		t.Errorf("fixedHeight for non-Box = %v, want -1", h)
	}
}

func TestResolveFixedHeight_HeightSmallerThanSpacing(t *testing.T) {
	box := &document.Box{
		BoxStyle: document.BoxStyle{Height: document.Pt(10)},
	}
	margin := document.ResolvedEdges{Top: 10, Bottom: 10}
	padding := document.ResolvedEdges{Top: 5, Bottom: 5}
	border := document.ResolvedEdges{}
	// 10 - 20 - 10 = -20 => clamped to 0
	h := resolveFixedHeight(box, 500, 12, margin, padding, border)
	if h != 0 {
		t.Errorf("fixedHeight = %v, want 0", h)
	}
}

func TestResolveFixedHeight_ZeroAmount(t *testing.T) {
	box := &document.Box{
		BoxStyle: document.BoxStyle{Height: document.Pt(0)},
	}
	h := resolveFixedHeight(box, 500, 12, document.ResolvedEdges{}, document.ResolvedEdges{}, document.ResolvedEdges{})
	if h != -1 {
		t.Errorf("fixedHeight for zero amount = %v, want -1", h)
	}
}

func TestStretchPlacedNodes(t *testing.T) {
	nodes := []PlacedNode{
		{Size: document.Size{Width: 100, Height: 10}},
		{Size: document.Size{Width: 100, Height: 20}},
	}
	stretchPlacedNodes(nodes, 50)
	for i, n := range nodes {
		if n.Size.Height != 50 {
			t.Errorf("nodes[%d].Height = %v, want 50", i, n.Size.Height)
		}
	}
}

func TestStretchPlacedNodes_Empty(t *testing.T) {
	// Should not panic on empty/nil.
	stretchPlacedNodes(nil, 50)
	stretchPlacedNodes([]PlacedNode{}, 50)
}

// ---------------------------------------------------------------------------
// Image layout tests
// ---------------------------------------------------------------------------

func TestLayoutImage_Basic(t *testing.T) {
	bl := NewBlockLayout()
	img := &document.Image{
		Source: document.ImageSource{Width: 200, Height: 100},
	}
	constraints := Constraints{
		AvailableWidth:  500,
		AvailableHeight: 700,
	}
	result := bl.layoutImage(img, constraints)
	// Intrinsic size fits: 200x100.
	if !approxEqual(result.Bounds.Width, 200, 0.1) {
		t.Errorf("Width = %v, want 200", result.Bounds.Width)
	}
	if !approxEqual(result.Bounds.Height, 100, 0.1) {
		t.Errorf("Height = %v, want 100", result.Bounds.Height)
	}
}

func TestLayoutImage_ScaleDownToFitWidth(t *testing.T) {
	bl := NewBlockLayout()
	img := &document.Image{
		Source: document.ImageSource{Width: 1000, Height: 500},
	}
	constraints := Constraints{
		AvailableWidth:  200,
		AvailableHeight: 700,
	}
	result := bl.layoutImage(img, constraints)
	// Intrinsic 1000x500 scaled down to 200x100.
	if !approxEqual(result.Bounds.Width, 200, 0.1) {
		t.Errorf("Width = %v, want 200", result.Bounds.Width)
	}
	if !approxEqual(result.Bounds.Height, 100, 0.1) {
		t.Errorf("Height = %v, want 100", result.Bounds.Height)
	}
}

func TestLayoutImage_ExplicitWidth(t *testing.T) {
	bl := NewBlockLayout()
	img := &document.Image{
		Source:       document.ImageSource{Width: 200, Height: 100},
		DisplayWidth: document.Pt(100),
	}
	constraints := Constraints{
		AvailableWidth:  500,
		AvailableHeight: 700,
	}
	result := bl.layoutImage(img, constraints)
	// Width=100, Height=100/2=50 (aspect ratio 2:1).
	if !approxEqual(result.Bounds.Width, 100, 0.1) {
		t.Errorf("Width = %v, want 100", result.Bounds.Width)
	}
	if !approxEqual(result.Bounds.Height, 50, 0.1) {
		t.Errorf("Height = %v, want 50", result.Bounds.Height)
	}
}

func TestLayoutImage_ExplicitHeight(t *testing.T) {
	bl := NewBlockLayout()
	img := &document.Image{
		Source:        document.ImageSource{Width: 200, Height: 100},
		DisplayHeight: document.Pt(50),
	}
	constraints := Constraints{
		AvailableWidth:  500,
		AvailableHeight: 700,
	}
	result := bl.layoutImage(img, constraints)
	// Height=50, Width=50*2=100 (aspect ratio 2:1).
	if !approxEqual(result.Bounds.Width, 100, 0.1) {
		t.Errorf("Width = %v, want 100", result.Bounds.Width)
	}
	if !approxEqual(result.Bounds.Height, 50, 0.1) {
		t.Errorf("Height = %v, want 50", result.Bounds.Height)
	}
}

func TestLayoutImage_BothExplicit(t *testing.T) {
	bl := NewBlockLayout()
	img := &document.Image{
		Source:        document.ImageSource{Width: 200, Height: 100},
		DisplayWidth:  document.Pt(150),
		DisplayHeight: document.Pt(75),
	}
	constraints := Constraints{
		AvailableWidth:  500,
		AvailableHeight: 700,
	}
	result := bl.layoutImage(img, constraints)
	if !approxEqual(result.Bounds.Width, 150, 0.1) {
		t.Errorf("Width = %v, want 150", result.Bounds.Width)
	}
	if !approxEqual(result.Bounds.Height, 75, 0.1) {
		t.Errorf("Height = %v, want 75", result.Bounds.Height)
	}
}

func TestLayoutImage_ClampToAvailableHeight(t *testing.T) {
	bl := NewBlockLayout()
	img := &document.Image{
		Source: document.ImageSource{Width: 100, Height: 1000},
	}
	constraints := Constraints{
		AvailableWidth:  500,
		AvailableHeight: 200,
	}
	result := bl.layoutImage(img, constraints)
	// Height clamped to 200, width recalculated: 200 * (100/1000) = 20.
	if !approxEqual(result.Bounds.Height, 200, 0.1) {
		t.Errorf("Height = %v, want 200", result.Bounds.Height)
	}
	if !approxEqual(result.Bounds.Width, 20, 0.1) {
		t.Errorf("Width = %v, want 20", result.Bounds.Width)
	}
}

func TestLayoutImage_ZeroIntrinsic(t *testing.T) {
	bl := NewBlockLayout()
	img := &document.Image{
		Source: document.ImageSource{Width: 0, Height: 0},
	}
	constraints := Constraints{AvailableWidth: 500, AvailableHeight: 700}
	result := bl.layoutImage(img, constraints)
	if result.Bounds.Width != 0 || result.Bounds.Height != 0 {
		t.Errorf("Zero intrinsic image should produce zero bounds, got %+v", result.Bounds)
	}
}

func TestLayoutImage_NonImageNode(t *testing.T) {
	bl := NewBlockLayout()
	txt := &document.Text{Content: "not an image"}
	result := bl.layoutImage(txt, Constraints{AvailableWidth: 500, AvailableHeight: 700})
	if result.Bounds.Width != 0 || result.Bounds.Height != 0 {
		t.Errorf("Non-image should produce zero bounds, got %+v", result.Bounds)
	}
}

func TestLayoutImage_FitCover(t *testing.T) {
	bl := NewBlockLayout()
	// 200x100 image (2:1), bounds 100x100 → cover uses max scale → 100x50
	// but cover must fill both dimensions, so scale = max(100/200, 100/100) = 1.0
	// display = 200x100, then clamped to available width 100 → 100x50? No.
	// Actually: bounds 100x100, scaleW=100/200=0.5, scaleH=100/100=1.0
	// scale = max(0.5, 1.0) = 1.0 → 200x100, clamped to width 100 → 100x50
	img := &document.Image{
		Source:  document.ImageSource{Width: 200, Height: 100},
		FitMode: document.FitCover,
	}
	constraints := Constraints{
		AvailableWidth:  100,
		AvailableHeight: 100,
	}
	result := bl.layoutImage(img, constraints)
	// FitCover scales to cover bounds (100x100). Scale = max(0.5, 1.0) = 1.0
	// → 200x100, then clamped to available width 100 → 100x50.
	if !approxEqual(result.Bounds.Width, 100, 0.1) {
		t.Errorf("FitCover Width = %v, want 100", result.Bounds.Width)
	}
	if !approxEqual(result.Bounds.Height, 50, 0.1) {
		t.Errorf("FitCover Height = %v, want 50", result.Bounds.Height)
	}
}

func TestLayoutImage_FitCover_TallImage(t *testing.T) {
	bl := NewBlockLayout()
	// 100x200 image (0.5:1), bounds 100x100
	// scaleW = 100/100 = 1.0, scaleH = 100/200 = 0.5
	// scale = max(1.0, 0.5) = 1.0 → 100x200, clamped to height 100 → 50x100
	img := &document.Image{
		Source:  document.ImageSource{Width: 100, Height: 200},
		FitMode: document.FitCover,
	}
	constraints := Constraints{
		AvailableWidth:  100,
		AvailableHeight: 100,
	}
	result := bl.layoutImage(img, constraints)
	if !approxEqual(result.Bounds.Height, 100, 0.1) {
		t.Errorf("FitCover Height = %v, want 100", result.Bounds.Height)
	}
	if !approxEqual(result.Bounds.Width, 50, 0.1) {
		t.Errorf("FitCover Width = %v, want 50", result.Bounds.Width)
	}
}

func TestLayoutImage_FitStretch(t *testing.T) {
	bl := NewBlockLayout()
	img := &document.Image{
		Source:        document.ImageSource{Width: 200, Height: 100},
		FitMode:       document.FitStretch,
		DisplayWidth:  document.Pt(150),
		DisplayHeight: document.Pt(200),
	}
	constraints := Constraints{
		AvailableWidth:  500,
		AvailableHeight: 700,
	}
	result := bl.layoutImage(img, constraints)
	// Stretch ignores aspect ratio, uses explicit dimensions.
	if !approxEqual(result.Bounds.Width, 150, 0.1) {
		t.Errorf("FitStretch Width = %v, want 150", result.Bounds.Width)
	}
	if !approxEqual(result.Bounds.Height, 200, 0.1) {
		t.Errorf("FitStretch Height = %v, want 200", result.Bounds.Height)
	}
}

func TestLayoutImage_FitStretch_NoExplicit(t *testing.T) {
	bl := NewBlockLayout()
	img := &document.Image{
		Source:  document.ImageSource{Width: 200, Height: 100},
		FitMode: document.FitStretch,
	}
	constraints := Constraints{
		AvailableWidth:  300,
		AvailableHeight: 700,
	}
	result := bl.layoutImage(img, constraints)
	// No explicit dimensions: fills available width, aspect ratio preserved.
	if !approxEqual(result.Bounds.Width, 300, 0.1) {
		t.Errorf("FitStretch Width = %v, want 300", result.Bounds.Width)
	}
	if !approxEqual(result.Bounds.Height, 150, 0.1) {
		t.Errorf("FitStretch Height = %v, want 150", result.Bounds.Height)
	}
}

func TestLayoutImage_FitOriginal(t *testing.T) {
	bl := NewBlockLayout()
	img := &document.Image{
		Source:  document.ImageSource{Width: 200, Height: 100},
		FitMode: document.FitOriginal,
	}
	constraints := Constraints{
		AvailableWidth:  500,
		AvailableHeight: 700,
	}
	result := bl.layoutImage(img, constraints)
	// Original: intrinsic dimensions (200x100).
	if !approxEqual(result.Bounds.Width, 200, 0.1) {
		t.Errorf("FitOriginal Width = %v, want 200", result.Bounds.Width)
	}
	if !approxEqual(result.Bounds.Height, 100, 0.1) {
		t.Errorf("FitOriginal Height = %v, want 100", result.Bounds.Height)
	}
}

func TestLayoutImage_FitOriginal_Clamped(t *testing.T) {
	bl := NewBlockLayout()
	img := &document.Image{
		Source:  document.ImageSource{Width: 1000, Height: 500},
		FitMode: document.FitOriginal,
	}
	constraints := Constraints{
		AvailableWidth:  200,
		AvailableHeight: 700,
	}
	result := bl.layoutImage(img, constraints)
	// Original 1000x500 clamped to available width 200 → 200x100.
	if !approxEqual(result.Bounds.Width, 200, 0.1) {
		t.Errorf("FitOriginal Width = %v, want 200", result.Bounds.Width)
	}
	if !approxEqual(result.Bounds.Height, 100, 0.1) {
		t.Errorf("FitOriginal Height = %v, want 100", result.Bounds.Height)
	}
}

// ---------------------------------------------------------------------------
// layoutChild dispatch tests
// ---------------------------------------------------------------------------

func TestLayoutChild_TextNode(t *testing.T) {
	bl := NewBlockLayout()
	txt := &document.Text{
		Content:   "hello",
		TextStyle: document.Style{FontSize: 12, LineHeight: 1.2},
	}
	constraints := Constraints{
		AvailableWidth:  500,
		AvailableHeight: 700,
		FontResolver:    &mockFontResolver{},
	}
	result := bl.layoutChild(txt, constraints)
	if result.Bounds.Height <= 0 {
		t.Error("Text should have positive height")
	}
}

func TestLayoutChild_ImageNode(t *testing.T) {
	bl := NewBlockLayout()
	img := &document.Image{
		Source: document.ImageSource{Width: 100, Height: 50},
	}
	constraints := Constraints{
		AvailableWidth:  500,
		AvailableHeight: 700,
	}
	result := bl.layoutChild(img, constraints)
	if !approxEqual(result.Bounds.Width, 100, 0.1) {
		t.Errorf("Image width = %v, want 100", result.Bounds.Width)
	}
}

func TestLayoutChild_TableNode(t *testing.T) {
	bl := NewBlockLayout()
	tbl := &document.Table{
		Columns: []document.TableColumn{
			{Width: document.Pct(50)},
			{Width: document.Pct(50)},
		},
		Body: []document.TableRow{
			{Cells: []document.TableCell{
				{Content: []document.DocumentNode{&document.Text{Content: "A", TextStyle: document.Style{FontSize: 12, LineHeight: 1.2}}}},
				{Content: []document.DocumentNode{&document.Text{Content: "B", TextStyle: document.Style{FontSize: 12, LineHeight: 1.2}}}},
			}},
		},
	}
	constraints := Constraints{
		AvailableWidth:  400,
		AvailableHeight: 700,
		FontResolver:    &mockFontResolver{},
	}
	result := bl.layoutChild(tbl, constraints)
	if result.Bounds.Height <= 0 {
		t.Error("Table should have positive height")
	}
}

func TestLayoutChild_BoxNode(t *testing.T) {
	bl := NewBlockLayout()
	box := &document.Box{
		Content: []document.DocumentNode{
			&document.Text{Content: "nested", TextStyle: document.Style{FontSize: 12, LineHeight: 1.2}},
		},
	}
	constraints := Constraints{
		AvailableWidth:  500,
		AvailableHeight: 700,
		FontResolver:    &mockFontResolver{},
	}
	result := bl.layoutChild(box, constraints)
	if result.Bounds.Height <= 0 {
		t.Error("Box should have positive height")
	}
}

// ---------------------------------------------------------------------------
// Table layout tests
// ---------------------------------------------------------------------------

func TestTableColumnCount_FromColumns(t *testing.T) {
	tbl := &document.Table{
		Columns: []document.TableColumn{
			{Width: document.Pt(100)},
			{Width: document.Pt(200)},
			{Width: document.Pt(100)},
		},
	}
	if n := tableColumnCount(tbl); n != 3 {
		t.Errorf("tableColumnCount = %d, want 3", n)
	}
}

func TestTableColumnCount_FromHeader(t *testing.T) {
	tbl := &document.Table{
		Header: []document.TableRow{
			{Cells: []document.TableCell{{}, {}, {}}},
		},
	}
	if n := tableColumnCount(tbl); n != 3 {
		t.Errorf("tableColumnCount = %d, want 3", n)
	}
}

func TestTableColumnCount_FromBody(t *testing.T) {
	tbl := &document.Table{
		Body: []document.TableRow{
			{Cells: []document.TableCell{{}, {}}},
		},
	}
	if n := tableColumnCount(tbl); n != 2 {
		t.Errorf("tableColumnCount = %d, want 2", n)
	}
}

func TestTableColumnCount_Empty(t *testing.T) {
	tbl := &document.Table{}
	if n := tableColumnCount(tbl); n != 0 {
		t.Errorf("tableColumnCount = %d, want 0", n)
	}
}

func TestCountRowColumns_Simple(t *testing.T) {
	row := document.TableRow{
		Cells: []document.TableCell{{}, {}, {}},
	}
	if n := countRowColumns(row); n != 3 {
		t.Errorf("countRowColumns = %d, want 3", n)
	}
}

func TestCountRowColumns_WithColSpan(t *testing.T) {
	row := document.TableRow{
		Cells: []document.TableCell{
			{ColSpan: 2},
			{ColSpan: 1},
		},
	}
	if n := countRowColumns(row); n != 3 {
		t.Errorf("countRowColumns = %d, want 3", n)
	}
}

func TestCountRowColumns_ZeroColSpanTreatedAsOne(t *testing.T) {
	row := document.TableRow{
		Cells: []document.TableCell{
			{ColSpan: 0},
			{ColSpan: 0},
		},
	}
	if n := countRowColumns(row); n != 2 {
		t.Errorf("countRowColumns = %d, want 2 (zero ColSpan treated as 1)", n)
	}
}

func TestResolveTableColumnWidths_NoColumns(t *testing.T) {
	widths := resolveTableColumnWidths(nil, 4, 400, 12)
	for i, w := range widths {
		if !approxEqual(w, 100, 0.1) {
			t.Errorf("widths[%d] = %v, want 100", i, w)
		}
	}
}

func TestResolveTableColumnWidths_AllExplicit(t *testing.T) {
	cols := []document.TableColumn{
		{Width: document.Pt(100)},
		{Width: document.Pt(200)},
		{Width: document.Pt(100)},
	}
	widths := resolveTableColumnWidths(cols, 3, 400, 12)
	if !approxEqual(widths[0], 100, 0.1) {
		t.Errorf("widths[0] = %v, want 100", widths[0])
	}
	if !approxEqual(widths[1], 200, 0.1) {
		t.Errorf("widths[1] = %v, want 200", widths[1])
	}
	if !approxEqual(widths[2], 100, 0.1) {
		t.Errorf("widths[2] = %v, want 100", widths[2])
	}
}

func TestResolveTableColumnWidths_MixedExplicitAuto(t *testing.T) {
	cols := []document.TableColumn{
		{Width: document.Pt(100)},
		{}, // auto
	}
	widths := resolveTableColumnWidths(cols, 2, 400, 12)
	if !approxEqual(widths[0], 100, 0.1) {
		t.Errorf("widths[0] = %v, want 100", widths[0])
	}
	if !approxEqual(widths[1], 300, 0.1) {
		t.Errorf("widths[1] = %v, want 300", widths[1])
	}
}

func TestResolveTableColumnWidths_MoreColsThanDefinitions(t *testing.T) {
	cols := []document.TableColumn{
		{Width: document.Pt(100)},
	}
	// numCols=3 but only 1 column defined: cols beyond len(cols) are auto.
	widths := resolveTableColumnWidths(cols, 3, 400, 12)
	if !approxEqual(widths[0], 100, 0.1) {
		t.Errorf("widths[0] = %v, want 100", widths[0])
	}
	// Remaining 300 split between 2 auto columns.
	if !approxEqual(widths[1], 150, 0.1) {
		t.Errorf("widths[1] = %v, want 150", widths[1])
	}
	if !approxEqual(widths[2], 150, 0.1) {
		t.Errorf("widths[2] = %v, want 150", widths[2])
	}
}

func TestLayoutTable_SimpleGrid(t *testing.T) {
	bl := NewBlockLayout()
	tbl := &document.Table{
		Columns: []document.TableColumn{
			{Width: document.Pct(50)},
			{Width: document.Pct(50)},
		},
		Body: []document.TableRow{
			{Cells: []document.TableCell{
				{Content: []document.DocumentNode{&document.Text{Content: "A1", TextStyle: document.Style{FontSize: 12, LineHeight: 1.2}}}},
				{Content: []document.DocumentNode{&document.Text{Content: "B1", TextStyle: document.Style{FontSize: 12, LineHeight: 1.2}}}},
			}},
			{Cells: []document.TableCell{
				{Content: []document.DocumentNode{&document.Text{Content: "A2", TextStyle: document.Style{FontSize: 12, LineHeight: 1.2}}}},
				{Content: []document.DocumentNode{&document.Text{Content: "B2", TextStyle: document.Style{FontSize: 12, LineHeight: 1.2}}}},
			}},
		},
	}
	constraints := Constraints{
		AvailableWidth:  400,
		AvailableHeight: 700,
		FontResolver:    &mockFontResolver{},
	}
	result := bl.layoutTable(tbl, constraints)
	// 2 rows * 2 cells = 4 placed nodes.
	if len(result.Children) != 4 {
		t.Fatalf("Expected 4 children, got %d", len(result.Children))
	}
	if result.Bounds.Height <= 0 {
		t.Error("Table should have positive height")
	}
}

func TestLayoutTable_HeaderBodyFooter(t *testing.T) {
	bl := NewBlockLayout()
	mkRow := func(content string) document.TableRow {
		return document.TableRow{
			Cells: []document.TableCell{
				{Content: []document.DocumentNode{&document.Text{Content: content, TextStyle: document.Style{FontSize: 12, LineHeight: 1.2}}}},
			},
		}
	}
	tbl := &document.Table{
		Columns: []document.TableColumn{{Width: document.Pct(100)}},
		Header:  []document.TableRow{mkRow("Header")},
		Body:    []document.TableRow{mkRow("Body1"), mkRow("Body2")},
		Footer:  []document.TableRow{mkRow("Footer")},
	}
	constraints := Constraints{
		AvailableWidth:  400,
		AvailableHeight: 700,
		FontResolver:    &mockFontResolver{},
	}
	result := bl.layoutTable(tbl, constraints)
	// 1 header + 2 body + 1 footer = 4 cells.
	if len(result.Children) != 4 {
		t.Fatalf("Expected 4 children, got %d", len(result.Children))
	}
}

func TestLayoutTable_ZeroColumns(t *testing.T) {
	bl := NewBlockLayout()
	tbl := &document.Table{} // No columns, no rows.
	constraints := Constraints{
		AvailableWidth:  400,
		AvailableHeight: 700,
		FontResolver:    &mockFontResolver{},
	}
	result := bl.layoutTable(tbl, constraints)
	if len(result.Children) != 0 {
		t.Errorf("Expected 0 children for empty table, got %d", len(result.Children))
	}
}

func TestLayoutTable_WithMarginAndPadding(t *testing.T) {
	bl := NewBlockLayout()
	tbl := &document.Table{
		Columns: []document.TableColumn{{Width: document.Pct(100)}},
		Body: []document.TableRow{
			{Cells: []document.TableCell{
				{Content: []document.DocumentNode{&document.Text{Content: "X", TextStyle: document.Style{FontSize: 12, LineHeight: 1.2}}}},
			}},
		},
		TableStyle: document.TableStyle{
			BoxStyle: document.BoxStyle{
				Margin:  document.UniformEdges(document.Pt(10)),
				Padding: document.UniformEdges(document.Pt(5)),
			},
		},
	}
	constraints := Constraints{
		AvailableWidth:  400,
		AvailableHeight: 700,
		FontResolver:    &mockFontResolver{},
	}
	result := bl.layoutTable(tbl, constraints)
	if len(result.Children) != 1 {
		t.Fatalf("Expected 1 child, got %d", len(result.Children))
	}
	// Cell should be offset by margin + padding = 15.
	if result.Children[0].Position.X < 15 {
		t.Errorf("cell X = %v, want >= 15", result.Children[0].Position.X)
	}
}

func TestLayoutTableRow_ColSpan(t *testing.T) {
	bl := NewBlockLayout()
	row := document.TableRow{
		Cells: []document.TableCell{
			{
				Content: []document.DocumentNode{&document.Text{Content: "Wide", TextStyle: document.Style{FontSize: 12, LineHeight: 1.2}}},
				ColSpan: 2,
			},
			{
				Content: []document.DocumentNode{&document.Text{Content: "Narrow", TextStyle: document.Style{FontSize: 12, LineHeight: 1.2}}},
			},
		},
	}
	colWidths := []float64{100, 100, 100}
	colOffsets := []float64{0, 100, 200}
	constraints := Constraints{
		AvailableWidth:  300,
		AvailableHeight: 700,
		FontResolver:    &mockFontResolver{},
	}
	placed, rowHeight := bl.layoutTableRow(row, colWidths, colOffsets, 3, constraints)
	if len(placed) != 2 {
		t.Fatalf("Expected 2 placed cells, got %d", len(placed))
	}
	// First cell spans 2 columns => width = 200.
	if !approxEqual(placed[0].Size.Width, 200, 0.1) {
		t.Errorf("cell[0] width = %v, want 200", placed[0].Size.Width)
	}
	// Second cell is at column 2 => width = 100.
	if !approxEqual(placed[1].Size.Width, 100, 0.1) {
		t.Errorf("cell[1] width = %v, want 100", placed[1].Size.Width)
	}
	if rowHeight <= 0 {
		t.Error("Row height should be positive")
	}
	// All cells stretched to same height.
	if placed[0].Size.Height != placed[1].Size.Height {
		t.Errorf("Cell heights should match: %v vs %v", placed[0].Size.Height, placed[1].Size.Height)
	}
}

func TestLayoutTableRow_ColSpanExceedsColumns(t *testing.T) {
	bl := NewBlockLayout()
	row := document.TableRow{
		Cells: []document.TableCell{
			{
				Content: []document.DocumentNode{&document.Text{Content: "X", TextStyle: document.Style{FontSize: 12, LineHeight: 1.2}}},
				ColSpan: 10, // Exceeds numCols=2, should be clamped.
			},
		},
	}
	colWidths := []float64{150, 150}
	colOffsets := []float64{0, 150}
	constraints := Constraints{
		AvailableWidth:  300,
		AvailableHeight: 700,
		FontResolver:    &mockFontResolver{},
	}
	placed, _ := bl.layoutTableRow(row, colWidths, colOffsets, 2, constraints)
	if len(placed) != 1 {
		t.Fatalf("Expected 1 placed cell, got %d", len(placed))
	}
	// ColSpan clamped to 2 => width = 300.
	if !approxEqual(placed[0].Size.Width, 300, 0.1) {
		t.Errorf("cell width = %v, want 300", placed[0].Size.Width)
	}
}

func TestLayoutTableRow_ZeroColSpanTreatedAsOne(t *testing.T) {
	bl := NewBlockLayout()
	row := document.TableRow{
		Cells: []document.TableCell{
			{
				Content: []document.DocumentNode{&document.Text{Content: "A", TextStyle: document.Style{FontSize: 12, LineHeight: 1.2}}},
				ColSpan: 0,
			},
			{
				Content: []document.DocumentNode{&document.Text{Content: "B", TextStyle: document.Style{FontSize: 12, LineHeight: 1.2}}},
				ColSpan: 0,
			},
		},
	}
	colWidths := []float64{200, 200}
	colOffsets := []float64{0, 200}
	constraints := Constraints{
		AvailableWidth:  400,
		AvailableHeight: 700,
		FontResolver:    &mockFontResolver{},
	}
	placed, _ := bl.layoutTableRow(row, colWidths, colOffsets, 2, constraints)
	if len(placed) != 2 {
		t.Fatalf("Expected 2 placed cells, got %d", len(placed))
	}
}

func TestPaginatorPageWithCustomMargins(t *testing.T) {
	defaultMargins := document.UniformEdges(document.Pt(36))
	customMargins := document.UniformEdges(document.Pt(100))
	p := NewPaginator(document.A4, defaultMargins, &mockFontResolver{})
	doc := &document.Document{
		Pages: []*document.Page{
			{
				Size:    document.A4,
				Margins: customMargins,
				Content: []document.DocumentNode{
					&document.Text{
						Content:   "test",
						TextStyle: document.Style{FontSize: 12, LineHeight: 1.2},
					},
				},
			},
		},
	}
	pages := p.Paginate(doc)
	if len(pages) < 1 {
		t.Fatal("Expected at least 1 page")
	}
	// Children should be offset by the custom margin (100 pt).
	for _, child := range pages[0].Children {
		if child.Position.X < 100 {
			t.Errorf("Child X = %v, want >= 100 (custom margin)", child.Position.X)
		}
	}
}

// ---------------------------------------------------------------------------
// BreakPolicy tests (WP2)
// ---------------------------------------------------------------------------

func TestBreakBeforeAlways(t *testing.T) {
	// Two children; second has BreakBefore=BreakAlways.
	// The second child should be pushed to overflow.
	bl := NewBlockLayout()
	parent := &document.Box{
		Content: []document.DocumentNode{
			&document.Box{
				Content: []document.DocumentNode{
					&document.Text{Content: "first", TextStyle: document.DefaultStyle()},
				},
			},
			&document.Box{
				BreakPolicy: document.BreakPolicy{BreakBefore: document.BreakAlways},
				Content: []document.DocumentNode{
					&document.Text{Content: "second", TextStyle: document.DefaultStyle()},
				},
			},
		},
	}
	constraints := Constraints{
		AvailableWidth:  500,
		AvailableHeight: 1000,
		FontResolver:    &mockFontResolver{},
	}
	result := bl.Layout(parent, constraints)
	if result.Overflow == nil {
		t.Fatal("Expected overflow due to BreakBefore=BreakAlways")
	}
	if len(result.Children) != 1 {
		t.Errorf("Expected 1 placed child, got %d", len(result.Children))
	}
}

func TestBreakBeforeFirstChild(t *testing.T) {
	// First child has BreakBefore=BreakAlways; should be ignored (no break
	// before the very first child).
	bl := NewBlockLayout()
	parent := &document.Box{
		Content: []document.DocumentNode{
			&document.Box{
				BreakPolicy: document.BreakPolicy{BreakBefore: document.BreakAlways},
				Content: []document.DocumentNode{
					&document.Text{Content: "only child", TextStyle: document.DefaultStyle()},
				},
			},
		},
	}
	constraints := Constraints{
		AvailableWidth:  500,
		AvailableHeight: 1000,
		FontResolver:    &mockFontResolver{},
	}
	result := bl.Layout(parent, constraints)
	if result.Overflow != nil {
		t.Error("BreakBefore on first child should not cause overflow")
	}
}

func TestBreakAfterAlways(t *testing.T) {
	// Three children; first has BreakAfter=BreakAlways.
	// Children 2 and 3 should overflow.
	bl := NewBlockLayout()
	parent := &document.Box{
		Content: []document.DocumentNode{
			&document.Box{
				BreakPolicy: document.BreakPolicy{BreakAfter: document.BreakAlways},
				Content: []document.DocumentNode{
					&document.Text{Content: "A", TextStyle: document.DefaultStyle()},
				},
			},
			&document.Box{
				Content: []document.DocumentNode{
					&document.Text{Content: "B", TextStyle: document.DefaultStyle()},
				},
			},
			&document.Box{
				Content: []document.DocumentNode{
					&document.Text{Content: "C", TextStyle: document.DefaultStyle()},
				},
			},
		},
	}
	constraints := Constraints{
		AvailableWidth:  500,
		AvailableHeight: 1000,
		FontResolver:    &mockFontResolver{},
	}
	result := bl.Layout(parent, constraints)
	if result.Overflow == nil {
		t.Fatal("Expected overflow due to BreakAfter=BreakAlways")
	}
	if len(result.Children) != 1 {
		t.Errorf("Expected 1 placed child, got %d", len(result.Children))
	}
}

func TestBreakInsideAvoid(t *testing.T) {
	// A child with BreakInside=BreakAvoid that would overflow should be
	// pushed entirely to the next page.
	bl := NewBlockLayout()

	// First child takes up most of the space.
	bigChild := &document.Box{
		BoxStyle: document.BoxStyle{Height: document.Pt(90)},
		Content: []document.DocumentNode{
			&document.Text{Content: "big", TextStyle: document.DefaultStyle()},
		},
	}
	// Second child is too tall to fit but has BreakInside=BreakAvoid.
	avoidChild := &document.Box{
		BoxStyle:    document.BoxStyle{Height: document.Pt(50)},
		BreakPolicy: document.BreakPolicy{BreakInside: document.BreakAvoid},
		Content: []document.DocumentNode{
			&document.Text{Content: strings.Repeat("x ", 100), TextStyle: document.DefaultStyle()},
		},
	}

	parent := &document.Box{Content: []document.DocumentNode{bigChild, avoidChild}}
	constraints := Constraints{
		AvailableWidth:  200,
		AvailableHeight: 100, // only 10pt left after bigChild
		FontResolver:    &mockFontResolver{},
	}
	result := bl.Layout(parent, constraints)
	if result.Overflow == nil {
		t.Fatal("Expected overflow for BreakInside=BreakAvoid child")
	}
	// The big child was placed; avoidChild went entirely to overflow.
	if len(result.Children) != 1 {
		t.Errorf("Expected 1 placed child (big), got %d", len(result.Children))
	}
}

func TestBreakInsideAvoidFits(t *testing.T) {
	// A child with BreakInside=BreakAvoid that fits entirely should not overflow.
	bl := NewBlockLayout()
	child := &document.Box{
		BreakPolicy: document.BreakPolicy{BreakInside: document.BreakAvoid},
		Content: []document.DocumentNode{
			&document.Text{Content: "short", TextStyle: document.DefaultStyle()},
		},
	}
	parent := &document.Box{Content: []document.DocumentNode{child}}
	constraints := Constraints{
		AvailableWidth:  500,
		AvailableHeight: 1000,
		FontResolver:    &mockFontResolver{},
	}
	result := bl.Layout(parent, constraints)
	if result.Overflow != nil {
		t.Error("BreakInside=BreakAvoid should not overflow when content fits")
	}
}

func TestBreakAutoDefault(t *testing.T) {
	// Default BreakPolicy (all BreakAuto) should behave normally.
	bl := NewBlockLayout()
	parent := &document.Box{
		Content: []document.DocumentNode{
			&document.Box{
				Content: []document.DocumentNode{
					&document.Text{Content: "A", TextStyle: document.DefaultStyle()},
				},
			},
			&document.Box{
				Content: []document.DocumentNode{
					&document.Text{Content: "B", TextStyle: document.DefaultStyle()},
				},
			},
		},
	}
	constraints := Constraints{
		AvailableWidth:  500,
		AvailableHeight: 1000,
		FontResolver:    &mockFontResolver{},
	}
	result := bl.Layout(parent, constraints)
	if result.Overflow != nil {
		t.Error("Default BreakAuto should not cause overflow")
	}
	if len(result.Children) != 2 {
		t.Errorf("Expected 2 placed children, got %d", len(result.Children))
	}
}

// ---------------------------------------------------------------------------
// List layout tests (WP5)
// ---------------------------------------------------------------------------

func TestListLayoutUnordered(t *testing.T) {
	bl := NewBlockLayout()
	lst := &document.List{
		ListType:  document.Unordered,
		ListStyle: document.DefaultStyle(),
		Items: []document.ListItem{
			{Content: []document.DocumentNode{&document.Text{Content: "Item A", TextStyle: document.DefaultStyle()}}},
			{Content: []document.DocumentNode{&document.Text{Content: "Item B", TextStyle: document.DefaultStyle()}}},
		},
	}
	constraints := Constraints{
		AvailableWidth:  500,
		AvailableHeight: 1000,
		FontResolver:    &mockFontResolver{},
	}
	result := bl.layoutList(lst, constraints)
	// 2 items → each produces 2 placed nodes (marker + content) = 4 total.
	if len(result.Children) != 4 {
		t.Errorf("Expected 4 placed children, got %d", len(result.Children))
	}
	if result.Overflow != nil {
		t.Error("List should fit without overflow")
	}
	if result.Bounds.Height <= 0 {
		t.Error("List should have positive height")
	}
}

func TestListLayoutOrdered(t *testing.T) {
	bl := NewBlockLayout()
	lst := &document.List{
		ListType:  document.Ordered,
		ListStyle: document.DefaultStyle(),
		Items: []document.ListItem{
			{Content: []document.DocumentNode{&document.Text{Content: "First", TextStyle: document.DefaultStyle()}}},
			{Content: []document.DocumentNode{&document.Text{Content: "Second", TextStyle: document.DefaultStyle()}}},
			{Content: []document.DocumentNode{&document.Text{Content: "Third", TextStyle: document.DefaultStyle()}}},
		},
	}
	constraints := Constraints{
		AvailableWidth:  500,
		AvailableHeight: 1000,
		FontResolver:    &mockFontResolver{},
	}
	result := bl.layoutList(lst, constraints)
	// 3 items → 6 placed nodes.
	if len(result.Children) != 6 {
		t.Errorf("Expected 6 placed children, got %d", len(result.Children))
	}
}

func TestListLayoutViaBlockChild(t *testing.T) {
	// Verify that a List is dispatched correctly through layoutChild.
	bl := NewBlockLayout()
	lst := &document.List{
		ListType:  document.Unordered,
		ListStyle: document.DefaultStyle(),
		Items: []document.ListItem{
			{Content: []document.DocumentNode{&document.Text{Content: "X", TextStyle: document.DefaultStyle()}}},
		},
	}
	parent := &document.Box{Content: []document.DocumentNode{lst}}
	constraints := Constraints{
		AvailableWidth:  500,
		AvailableHeight: 1000,
		FontResolver:    &mockFontResolver{},
	}
	result := bl.Layout(parent, constraints)
	if result.Overflow != nil {
		t.Error("Single-item list in box should not overflow")
	}
	if len(result.Children) < 1 {
		t.Error("Expected at least 1 placed child")
	}
}

func TestListLayoutCustomIndent(t *testing.T) {
	bl := NewBlockLayout()
	lst := &document.List{
		ListType:     document.Unordered,
		ListStyle:    document.DefaultStyle(),
		MarkerIndent: 40,
		Items: []document.ListItem{
			{Content: []document.DocumentNode{&document.Text{Content: "Wide indent", TextStyle: document.DefaultStyle()}}},
		},
	}
	constraints := Constraints{
		AvailableWidth:  500,
		AvailableHeight: 1000,
		FontResolver:    &mockFontResolver{},
	}
	result := bl.layoutList(lst, constraints)
	// Content should be placed at X=40 (the custom indent).
	if len(result.Children) >= 2 {
		contentNode := result.Children[1]
		if contentNode.Position.X != 40 {
			t.Errorf("Content X = %v, want 40 (custom indent)", contentNode.Position.X)
		}
	}
}

// ---------------------------------------------------------------------------
// Table header repeat on overflow tests (WP6)
// ---------------------------------------------------------------------------

func makeTestRow(text string) document.TableRow {
	return document.TableRow{
		Cells: []document.TableCell{
			{
				Content: []document.DocumentNode{&document.Text{Content: text, TextStyle: document.DefaultStyle()}},
				ColSpan: 1,
				RowSpan: 1,
			},
		},
	}
}

func TestTableHeaderRepeatOnOverflow(t *testing.T) {
	bl := NewBlockLayout()
	tbl := &document.Table{
		Header: []document.TableRow{makeTestRow("Header")},
		Body: []document.TableRow{
			makeTestRow("Row 1"),
			makeTestRow("Row 2"),
			makeTestRow("Row 3"),
			makeTestRow("Row 4"),
			makeTestRow("Row 5"),
		},
	}
	// Height only enough for header + ~2 body rows (each row ~14.4pt with 12pt font * 1.2 line height).
	constraints := Constraints{
		AvailableWidth:  200,
		AvailableHeight: 50,
		FontResolver:    &mockFontResolver{},
	}
	result := bl.layoutTable(tbl, constraints)
	if result.Overflow == nil {
		t.Fatal("Expected overflow when table doesn't fit")
	}
	// The overflow should be a Table with the same header.
	overflowTbl, ok := result.Overflow.(*document.Table)
	if !ok {
		t.Fatal("Overflow should be a *Table")
	}
	if len(overflowTbl.Header) != 1 {
		t.Errorf("Overflow table Header rows = %d, want 1", len(overflowTbl.Header))
	}
	if len(overflowTbl.Body) == 0 {
		t.Error("Overflow table should have remaining body rows")
	}
}

func TestTableFitsNoOverflow(t *testing.T) {
	bl := NewBlockLayout()
	tbl := &document.Table{
		Header: []document.TableRow{makeTestRow("Header")},
		Body: []document.TableRow{
			makeTestRow("Row 1"),
			makeTestRow("Row 2"),
		},
		Footer: []document.TableRow{makeTestRow("Footer")},
	}
	constraints := Constraints{
		AvailableWidth:  200,
		AvailableHeight: 1000, // plenty of space
		FontResolver:    &mockFontResolver{},
	}
	result := bl.layoutTable(tbl, constraints)
	if result.Overflow != nil {
		t.Error("Table should fit without overflow")
	}
	if result.Bounds.Height <= 0 {
		t.Error("Table should have positive height")
	}
}

func TestTableOverflowPreservesFooter(t *testing.T) {
	bl := NewBlockLayout()
	tbl := &document.Table{
		Header: []document.TableRow{makeTestRow("H")},
		Body: []document.TableRow{
			makeTestRow("B1"),
			makeTestRow("B2"),
			makeTestRow("B3"),
		},
		Footer: []document.TableRow{makeTestRow("F")},
	}
	constraints := Constraints{
		AvailableWidth:  200,
		AvailableHeight: 40, // tight: header + ~1 body row
		FontResolver:    &mockFontResolver{},
	}
	result := bl.layoutTable(tbl, constraints)
	if result.Overflow == nil {
		t.Fatal("Expected overflow")
	}
	overflowTbl, ok := result.Overflow.(*document.Table)
	if !ok {
		t.Fatal("Overflow should be *Table")
	}
	if len(overflowTbl.Footer) != 1 {
		t.Errorf("Overflow table Footer rows = %d, want 1", len(overflowTbl.Footer))
	}
}

// ---------------------------------------------------------------------------
// PageNumber tests (WP5)
// ---------------------------------------------------------------------------

func TestResolvePageNumbers_SinglePage(t *testing.T) {
	textNode := &document.Text{
		Content:   "Page " + document.PageNumberPlaceholder + " of " + document.TotalPagesPlaceholder,
		TextStyle: document.DefaultStyle(),
	}
	pages := []PageLayout{
		{
			Size: document.A4,
			Children: []PlacedNode{
				{Node: textNode, Position: document.Point{X: 0, Y: 0}, Size: document.Size{Width: 100, Height: 14}},
			},
		},
	}
	ResolvePageNumbers(pages)
	if textNode.Content != "Page 1 of 1" {
		t.Errorf("Content = %q, want %q", textNode.Content, "Page 1 of 1")
	}
}

func TestResolvePageNumbers_MultiplePages(t *testing.T) {
	textNodes := make([]*document.Text, 3)
	pages := make([]PageLayout, 3)
	for i := range 3 {
		textNodes[i] = &document.Text{
			Content:   document.PageNumberPlaceholder + "/" + document.TotalPagesPlaceholder,
			TextStyle: document.DefaultStyle(),
		}
		pages[i] = PageLayout{
			Size: document.A4,
			Children: []PlacedNode{
				{Node: textNodes[i], Position: document.Point{X: 0, Y: 0}, Size: document.Size{Width: 100, Height: 14}},
			},
		}
	}
	ResolvePageNumbers(pages)
	for i, tn := range textNodes {
		wantStr := fmt.Sprintf("%d/3", i+1)
		if tn.Content != wantStr {
			t.Errorf("Page %d: Content = %q, want %q", i+1, tn.Content, wantStr)
		}
	}
}

func TestResolvePageNumbers_NestedChildren(t *testing.T) {
	textNode := &document.Text{
		Content:   "Page " + document.PageNumberPlaceholder,
		TextStyle: document.DefaultStyle(),
	}
	pages := []PageLayout{
		{
			Size: document.A4,
			Children: []PlacedNode{
				{
					Node: &document.Box{},
					Children: []PlacedNode{
						{Node: textNode, Position: document.Point{X: 0, Y: 0}, Size: document.Size{Width: 100, Height: 14}},
					},
				},
			},
		},
	}
	ResolvePageNumbers(pages)
	if textNode.Content != "Page 1" {
		t.Errorf("Content = %q, want %q", textNode.Content, "Page 1")
	}
}

func TestPaginatorWithHeaderFooter(t *testing.T) {
	paginator := NewPaginator(document.A4, document.UniformEdges(document.Pt(50)), &mockFontResolver{})
	headerNodes := []document.DocumentNode{
		&document.Text{Content: "Header", TextStyle: document.Style{FontSize: 12, LineHeight: 1.2}},
	}
	footerNodes := []document.DocumentNode{
		&document.Text{Content: "Footer " + document.PageNumberPlaceholder, TextStyle: document.Style{FontSize: 12, LineHeight: 1.2}},
	}
	paginator.SetHeaderFooter(headerNodes, footerNodes)

	doc := &document.Document{
		Pages: []*document.Page{
			{
				Size:    document.A4,
				Margins: document.UniformEdges(document.Pt(50)),
				Content: []document.DocumentNode{
					&document.Text{Content: "Body content", TextStyle: document.Style{FontSize: 12, LineHeight: 1.2}},
				},
			},
		},
	}
	pages := paginator.Paginate(doc)
	if len(pages) < 1 {
		t.Fatal("Expected at least 1 page")
	}
	// The page should have header, body, and footer nodes.
	if len(pages[0].Children) < 3 {
		t.Errorf("Expected at least 3 placed node groups (header+body+footer), got %d", len(pages[0].Children))
	}

	// Resolve page numbers.
	ResolvePageNumbers(pages)

	// Verify the footer text was resolved.
	found := false
	var walkNodes func([]PlacedNode)
	walkNodes = func(nodes []PlacedNode) {
		for _, pn := range nodes {
			if tn, ok := pn.Node.(*document.Text); ok {
				if tn.Content == "Footer 1" {
					found = true
				}
			}
			walkNodes(pn.Children)
		}
	}
	walkNodes(pages[0].Children)
	if !found {
		t.Error("Expected footer text to be resolved to 'Footer 1'")
	}
}

// ---------------------------------------------------------------------------
// VerticalAlign tests (WP4)
// ---------------------------------------------------------------------------

func TestVerticalAlignTop(t *testing.T) {
	bl := NewBlockLayout()
	// Two cells: one tall (2 lines), one short (1 line). Default is VAlignTop.
	tbl := &document.Table{
		Body: []document.TableRow{
			{Cells: []document.TableCell{
				{Content: []document.DocumentNode{
					&document.Text{Content: "tall cell with long text that wraps to second line", TextStyle: document.Style{FontSize: 12, LineHeight: 1.2}},
				}, ColSpan: 1, RowSpan: 1},
				{Content: []document.DocumentNode{
					&document.Text{Content: "short", TextStyle: document.Style{FontSize: 12, LineHeight: 1.2, VerticalAlign: document.VAlignTop}},
				}, ColSpan: 1, RowSpan: 1},
			}},
		},
	}
	constraints := Constraints{
		AvailableWidth:  100,
		AvailableHeight: 500,
		FontResolver:    &mockFontResolver{},
	}
	result := bl.layoutTable(tbl, constraints)
	// Find the second cell.
	if len(result.Children) < 2 {
		t.Fatalf("Expected at least 2 placed cells, got %d", len(result.Children))
	}
	cell2 := result.Children[1]
	// With VAlignTop, the content should start at Y=0 (relative to cell).
	if len(cell2.Children) > 0 {
		firstChild := cell2.Children[0]
		if firstChild.Position.Y != 0 {
			t.Errorf("VAlignTop: first child Y = %v, want 0", firstChild.Position.Y)
		}
	}
}

func TestVerticalAlignMiddle(t *testing.T) {
	bl := NewBlockLayout()
	tbl := &document.Table{
		Body: []document.TableRow{
			{Cells: []document.TableCell{
				{Content: []document.DocumentNode{
					&document.Text{Content: "tall cell with long text that wraps to second line", TextStyle: document.Style{FontSize: 12, LineHeight: 1.2}},
				}, ColSpan: 1, RowSpan: 1},
				{Content: []document.DocumentNode{
					&document.Text{Content: "short", TextStyle: document.Style{FontSize: 12, LineHeight: 1.2, VerticalAlign: document.VAlignMiddle}},
				}, ColSpan: 1, RowSpan: 1, CellStyle: document.Style{VerticalAlign: document.VAlignMiddle}},
			}},
		},
	}
	constraints := Constraints{
		AvailableWidth:  100,
		AvailableHeight: 500,
		FontResolver:    &mockFontResolver{},
	}
	result := bl.layoutTable(tbl, constraints)
	if len(result.Children) < 2 {
		t.Fatalf("Expected at least 2 placed cells, got %d", len(result.Children))
	}
	cell2 := result.Children[1]
	// The second cell should have content offset downward.
	if len(cell2.Children) > 0 {
		firstChild := cell2.Children[0]
		if firstChild.Position.Y <= 0 {
			t.Errorf("VAlignMiddle: first child Y = %v, want > 0", firstChild.Position.Y)
		}
	}
}

func TestVerticalAlignBottom(t *testing.T) {
	bl := NewBlockLayout()
	tbl := &document.Table{
		Body: []document.TableRow{
			{Cells: []document.TableCell{
				{Content: []document.DocumentNode{
					&document.Text{Content: "tall cell with long text that wraps to second line", TextStyle: document.Style{FontSize: 12, LineHeight: 1.2}},
				}, ColSpan: 1, RowSpan: 1},
				{Content: []document.DocumentNode{
					&document.Text{Content: "short", TextStyle: document.Style{FontSize: 12, LineHeight: 1.2, VerticalAlign: document.VAlignBottom}},
				}, ColSpan: 1, RowSpan: 1, CellStyle: document.Style{VerticalAlign: document.VAlignBottom}},
			}},
		},
	}
	constraints := Constraints{
		AvailableWidth:  100,
		AvailableHeight: 500,
		FontResolver:    &mockFontResolver{},
	}
	result := bl.layoutTable(tbl, constraints)
	if len(result.Children) < 2 {
		t.Fatalf("Expected at least 2 placed cells, got %d", len(result.Children))
	}
	cell2 := result.Children[1]
	// The second cell should have content pushed to the bottom.
	if len(cell2.Children) > 0 {
		firstChild := cell2.Children[0]
		if firstChild.Position.Y <= 0 {
			t.Errorf("VAlignBottom: first child Y = %v, want > 0", firstChild.Position.Y)
		}
		// Bottom alignment should give a larger offset than middle.
	}
}

// ---------------------------------------------------------------------------
// RichText layout tests
// ---------------------------------------------------------------------------

func TestLayoutRichText_SingleFragment(t *testing.T) {
	c := Constraints{
		AvailableWidth:  500,
		AvailableHeight: 700,
		FontResolver:    &mockFontResolver{},
	}

	rt := &document.RichText{
		Fragments: []document.RichTextFragment{
			{Content: "Hello", FragmentStyle: document.Style{FontSize: 12}},
		},
		BlockStyle: document.Style{FontSize: 12, LineHeight: 1.2},
	}

	fl := &FlowLayout{}
	result := fl.LayoutRichText(rt, c)

	// Should produce one line.
	if len(result.Children) != 1 {
		t.Fatalf("expected 1 line, got %d", len(result.Children))
	}

	line := result.Children[0]
	if line.Node.NodeType() != document.NodeRichText {
		t.Errorf("line NodeType = %v, want NodeRichText", line.Node.NodeType())
	}

	// The line should have one child (the word "Hello").
	if len(line.Children) != 1 {
		t.Fatalf("expected 1 child in line, got %d", len(line.Children))
	}

	child := line.Children[0]
	textNode, ok := child.Node.(*document.Text)
	if !ok {
		t.Fatal("child is not *document.Text")
	}
	if textNode.Content != "Hello" {
		t.Errorf("text content = %q, want %q", textNode.Content, "Hello")
	}

	if result.Overflow != nil {
		t.Errorf("unexpected overflow")
	}
}

func TestLayoutRichText_TwoFragmentsDifferentStyles(t *testing.T) {
	c := Constraints{
		AvailableWidth:  500,
		AvailableHeight: 700,
		FontResolver:    &mockFontResolver{},
	}

	rt := &document.RichText{
		Fragments: []document.RichTextFragment{
			{Content: "Hello ", FragmentStyle: document.Style{FontSize: 12}},
			{Content: "world", FragmentStyle: document.Style{FontSize: 12, FontWeight: document.WeightBold}},
		},
		BlockStyle: document.Style{FontSize: 12, LineHeight: 1.2},
	}

	fl := &FlowLayout{}
	result := fl.LayoutRichText(rt, c)

	if len(result.Children) != 1 {
		t.Fatalf("expected 1 line, got %d", len(result.Children))
	}

	line := result.Children[0]
	// "Hello " splits into "Hello" (word) + " " (space), "world" (word).
	// Space runs are not placed as children, only words.
	if len(line.Children) != 2 {
		t.Fatalf("expected 2 children (Hello, world), got %d", len(line.Children))
	}

	first := line.Children[0].Node.(*document.Text)
	second := line.Children[1].Node.(*document.Text)
	if first.Content != "Hello" {
		t.Errorf("first word = %q, want %q", first.Content, "Hello")
	}
	if second.Content != "world" {
		t.Errorf("second word = %q, want %q", second.Content, "world")
	}
	if second.TextStyle.FontWeight != document.WeightBold {
		t.Errorf("second word FontWeight = %v, want Bold", second.TextStyle.FontWeight)
	}
}

func TestLayoutRichText_LineWrapping(t *testing.T) {
	c := Constraints{
		AvailableWidth:  60, // At 0.5*12=6 per char, fits 10 chars per line.
		AvailableHeight: 700,
		FontResolver:    &mockFontResolver{},
	}

	rt := &document.RichText{
		Fragments: []document.RichTextFragment{
			{Content: "Hello ", FragmentStyle: document.Style{FontSize: 12}},
			{Content: "beautiful world", FragmentStyle: document.Style{FontSize: 12, FontWeight: document.WeightBold}},
		},
		BlockStyle: document.Style{FontSize: 12, LineHeight: 1.2},
	}

	fl := &FlowLayout{}
	result := fl.LayoutRichText(rt, c)

	// "Hello" (5), " " (1), "beautiful" (9), " " (1), "world" (5)
	// Line 1: "Hello" + " " + ... "beautiful" = 5+1+9 = 15 chars = 90 > 60, so
	// Line 1: "Hello" (5 chars * 6 = 30), then space + "beautiful" won't fit
	// Line 1: "Hello" (30)
	// Line 2: "beautiful" (54)
	// Line 3: "world" (30)
	if len(result.Children) < 2 {
		t.Fatalf("expected at least 2 lines, got %d", len(result.Children))
	}
}

func TestLayoutRichText_DifferentFontSizes(t *testing.T) {
	c := Constraints{
		AvailableWidth:  500,
		AvailableHeight: 700,
		FontResolver:    &mockFontResolver{},
	}

	rt := &document.RichText{
		Fragments: []document.RichTextFragment{
			{Content: "Small ", FragmentStyle: document.Style{FontSize: 10}},
			{Content: "BIG", FragmentStyle: document.Style{FontSize: 24}},
		},
		BlockStyle: document.Style{FontSize: 12, LineHeight: 1.2},
	}

	fl := &FlowLayout{}
	result := fl.LayoutRichText(rt, c)

	if len(result.Children) != 1 {
		t.Fatalf("expected 1 line, got %d", len(result.Children))
	}

	line := result.Children[0]
	if len(line.Children) != 2 {
		t.Fatalf("expected 2 children, got %d", len(line.Children))
	}

	// The smaller text should have a larger Y offset for baseline alignment.
	smallY := line.Children[0].Position.Y
	bigY := line.Children[1].Position.Y

	if smallY <= bigY {
		t.Errorf("small text Y=%v should be > big text Y=%v for baseline alignment", smallY, bigY)
	}

	// Line spacing should be based on the max font size (24).
	expectedSpacing := 24 * 1.2
	if !approxEqual(line.Size.Height, expectedSpacing, 0.01) {
		t.Errorf("line height = %v, want %v", line.Size.Height, expectedSpacing)
	}
}

func TestLayoutRichText_TextAlign(t *testing.T) {
	c := Constraints{
		AvailableWidth:  500,
		AvailableHeight: 700,
		FontResolver:    &mockFontResolver{},
	}

	tests := []struct {
		name  string
		align document.TextAlign
		wantX func(lineWidth float64) float64
	}{
		{"Left", document.AlignLeft, func(float64) float64 { return 0 }},
		{"Center", document.AlignCenter, func(w float64) float64 { return (500 - w) / 2 }},
		{"Right", document.AlignRight, func(w float64) float64 { return 500 - w }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rt := &document.RichText{
				Fragments: []document.RichTextFragment{
					{Content: "Test", FragmentStyle: document.Style{FontSize: 12}},
				},
				BlockStyle: document.Style{FontSize: 12, LineHeight: 1.2, TextAlign: tt.align},
			}

			fl := &FlowLayout{}
			result := fl.LayoutRichText(rt, c)

			if len(result.Children) == 0 {
				t.Fatal("no lines")
			}
			line := result.Children[0]
			if len(line.Children) == 0 {
				t.Fatal("no children in line")
			}

			// "Test" = 4 chars * 0.5 * 12 = 24
			textWidth := 24.0
			expectedX := tt.wantX(textWidth)
			gotX := line.Children[0].Position.X

			if !approxEqual(gotX, expectedX, 0.01) {
				t.Errorf("X position = %v, want %v", gotX, expectedX)
			}
		})
	}
}

func TestLayoutRichText_TextIndent(t *testing.T) {
	c := Constraints{
		AvailableWidth:  500,
		AvailableHeight: 700,
		FontResolver:    &mockFontResolver{},
	}

	rt := &document.RichText{
		Fragments: []document.RichTextFragment{
			{Content: "Hello", FragmentStyle: document.Style{FontSize: 12}},
		},
		BlockStyle: document.Style{FontSize: 12, LineHeight: 1.2, TextIndent: document.Pt(30)},
	}

	fl := &FlowLayout{}
	result := fl.LayoutRichText(rt, c)

	if len(result.Children) == 0 {
		t.Fatal("no lines")
	}
	line := result.Children[0]
	if len(line.Children) == 0 {
		t.Fatal("no children")
	}

	// First line should be indented by 30pt.
	if line.Children[0].Position.X < 30 {
		t.Errorf("first child X = %v, want >= 30 (indent)", line.Children[0].Position.X)
	}
}

func TestLayoutRichText_Overflow(t *testing.T) {
	// At 6px/char, 60px width fits 10 chars per line.
	// The text "Hello world more text here" needs multiple lines.
	// 14.4 lineSpacing fits 1 line in height=15.
	c := Constraints{
		AvailableWidth:  60,
		AvailableHeight: 15, // Only enough for ~1 line at 12*1.2=14.4.
		FontResolver:    &mockFontResolver{},
	}

	rt := &document.RichText{
		Fragments: []document.RichTextFragment{
			{Content: "Hello world more text here", FragmentStyle: document.Style{FontSize: 12}},
		},
		BlockStyle: document.Style{FontSize: 12, LineHeight: 1.2},
	}

	fl := &FlowLayout{}
	result := fl.LayoutRichText(rt, c)

	if len(result.Children) == 0 {
		t.Fatal("expected at least one line placed")
	}
	if result.Overflow == nil {
		t.Fatal("expected overflow, got nil")
	}
	overflowRT, ok := result.Overflow.(*document.RichText)
	if !ok {
		t.Fatalf("overflow type = %T, want *document.RichText", result.Overflow)
	}
	if len(overflowRT.Fragments) == 0 {
		t.Error("overflow RichText has no fragments")
	}
}

func TestLayoutRichText_EmptyFragments(t *testing.T) {
	c := Constraints{
		AvailableWidth:  500,
		AvailableHeight: 700,
		FontResolver:    &mockFontResolver{},
	}

	rt := &document.RichText{
		BlockStyle: document.Style{FontSize: 12, LineHeight: 1.2},
	}

	fl := &FlowLayout{}
	result := fl.LayoutRichText(rt, c)

	if result.Bounds.Height != 0 {
		t.Errorf("height = %v, want 0 for empty", result.Bounds.Height)
	}
	if len(result.Children) != 0 {
		t.Errorf("children = %d, want 0 for empty", len(result.Children))
	}
}

func TestLayoutRichText_Justify(t *testing.T) {
	// Width fits exactly ~10 chars at 6px each = 60.
	c := Constraints{
		AvailableWidth:  60,
		AvailableHeight: 700,
		FontResolver:    &mockFontResolver{},
	}

	// "aa bb cc dd" should wrap. With justify, spaces on non-last lines expand.
	rt := &document.RichText{
		Fragments: []document.RichTextFragment{
			{Content: "aa bb cc dd ee", FragmentStyle: document.Style{FontSize: 12}},
		},
		BlockStyle: document.Style{FontSize: 12, LineHeight: 1.2, TextAlign: document.AlignJustify},
	}

	fl := &FlowLayout{}
	result := fl.LayoutRichText(rt, c)

	if len(result.Children) < 2 {
		t.Fatalf("expected at least 2 lines for justify test, got %d", len(result.Children))
	}

	// Non-last lines should have children spanning the full width.
	// Just verify that the line was produced without error.
}

func TestLayoutRichText_BlockDispatch(t *testing.T) {
	c := Constraints{
		AvailableWidth:  500,
		AvailableHeight: 700,
		FontResolver:    &mockFontResolver{},
	}

	rt := &document.RichText{
		Fragments: []document.RichTextFragment{
			{Content: "via block", FragmentStyle: document.Style{FontSize: 12}},
		},
		BlockStyle: document.Style{FontSize: 12, LineHeight: 1.2},
	}

	box := &document.Box{
		Content: []document.DocumentNode{rt},
	}

	bl := NewBlockLayout()
	result := bl.Layout(box, c)

	if len(result.Children) == 0 {
		t.Fatal("expected placed children from block layout with RichText")
	}
}

func TestEffectiveFontSize_Positive(t *testing.T) {
	if got := effectiveFontSize(14); got != 14 {
		t.Errorf("effectiveFontSize(14) = %g, want 14", got)
	}
}

func TestEffectiveFontSize_Zero(t *testing.T) {
	if got := effectiveFontSize(0); got != 12 {
		t.Errorf("effectiveFontSize(0) = %g, want 12", got)
	}
}

func TestEffectiveFontSize_Negative(t *testing.T) {
	if got := effectiveFontSize(-5); got != 12 {
		t.Errorf("effectiveFontSize(-5) = %g, want 12", got)
	}
}

func TestMeasureRunWidth_NilResolver(t *testing.T) {
	style := document.Style{FontSize: 10}
	constraints := Constraints{FontResolver: nil}
	w := measureRunWidth("Hello", style, 10, constraints)
	// Without resolver: len("Hello")=5, 5*10*0.5 = 25
	if !approxEqual(w, 25, 0.1) {
		t.Errorf("measureRunWidth(nil resolver) = %g, want 25", w)
	}
}

func TestMeasureRunWidth_WithResolver(t *testing.T) {
	style := document.Style{FontSize: 10, FontFamily: "Test"}
	constraints := Constraints{FontResolver: &mockFontResolver{}}
	w := measureRunWidth("Hello", style, 10, constraints)
	// With mock resolver: 5 chars * 10 * 0.5 = 25
	if !approxEqual(w, 25, 0.1) {
		t.Errorf("measureRunWidth(mock resolver) = %g, want 25", w)
	}
}

func TestMeasureRunWidth_WithLetterSpacing(t *testing.T) {
	style := document.Style{FontSize: 10, FontFamily: "Test", LetterSpacing: 2.0}
	constraints := Constraints{FontResolver: &mockFontResolver{}}
	w := measureRunWidth("Hello", style, 10, constraints)
	// Mock: 5*10*0.5=25, plus letter spacing: 2.0 * (5-1) = 8 → total 33
	if !approxEqual(w, 33, 0.1) {
		t.Errorf("measureRunWidth(letter spacing) = %g, want 33", w)
	}
}

func TestListItemNode_StyleViaList(t *testing.T) {
	// Test ListItemNode Style fallback via List.Children().
	lst := &document.List{
		ListStyle: document.Style{FontSize: 12},
		Items: []document.ListItem{
			{ItemStyle: document.Style{}},             // No font size → falls back to list
			{ItemStyle: document.Style{FontSize: 16}}, // Has font size → overrides
		},
	}
	children := lst.Children()
	if len(children) != 2 {
		t.Fatalf("expected 2 children, got %d", len(children))
	}
	// First item falls back to list style.
	s0 := children[0].Style()
	if s0.FontSize != 12 {
		t.Errorf("item 0: expected list style fallback (12), got FontSize=%g", s0.FontSize)
	}
	// Second item overrides with its own style.
	s1 := children[1].Style()
	if s1.FontSize != 16 {
		t.Errorf("item 1: expected item style override (16), got FontSize=%g", s1.FontSize)
	}
}
