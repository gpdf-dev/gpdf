package layout

import (
	"testing"

	"github.com/gpdf-dev/gpdf/document"
)

func TestAbsolutePositioning_BasicPlacement(t *testing.T) {
	block := NewBlockLayout()

	// Container with one flow child and one absolute child.
	container := &document.Box{
		Content: []document.DocumentNode{
			&document.Text{
				Content:   "Flow text",
				TextStyle: document.Style{FontSize: 12},
			},
			&document.Box{
				Content: []document.DocumentNode{
					&document.Text{
						Content:   "Absolute text",
						TextStyle: document.Style{FontSize: 12},
					},
				},
				BoxStyle: document.BoxStyle{
					Position: document.Position{
						Mode: document.PositionAbsolute,
						X:    document.Mm(50),
						Y:    document.Mm(100),
					},
				},
			},
		},
	}

	constraints := Constraints{
		AvailableWidth:  595.28, // A4 width
		AvailableHeight: 841.89,
		FontResolver:    &mockFontResolver{},
	}

	result := block.Layout(container, constraints)

	// Should have 2 placed nodes: flow text + absolute box.
	if len(result.Children) < 2 {
		t.Fatalf("expected at least 2 placed children, got %d", len(result.Children))
	}

	// The absolute node should be last (rendered on top).
	absNode := result.Children[len(result.Children)-1]
	box, ok := absNode.Node.(*document.Box)
	if !ok {
		t.Fatal("last child should be a Box")
	}
	if box.BoxStyle.Position.Mode != document.PositionAbsolute {
		t.Error("last child should be absolute-positioned")
	}

	// Check that the absolute node is at roughly (50mm, 100mm) in pt.
	expectedX := document.Mm(50).Resolve(595.28, 12)
	expectedY := document.Mm(100).Resolve(841.89, 12)
	if !approxEqual(absNode.Position.X, expectedX, 1.0) {
		t.Errorf("absolute X: got %g, want ~%g", absNode.Position.X, expectedX)
	}
	if !approxEqual(absNode.Position.Y, expectedY, 1.0) {
		t.Errorf("absolute Y: got %g, want ~%g", absNode.Position.Y, expectedY)
	}
}

func TestAbsolutePositioning_DoesNotAffectFlow(t *testing.T) {
	block := NewBlockLayout()

	flowText1 := &document.Text{
		Content:   "First",
		TextStyle: document.Style{FontSize: 12},
	}
	flowText2 := &document.Text{
		Content:   "Second",
		TextStyle: document.Style{FontSize: 12},
	}

	// Layout without absolute node.
	containerWithout := &document.Box{
		Content: []document.DocumentNode{flowText1, flowText2},
	}
	constraints := Constraints{
		AvailableWidth:  500,
		AvailableHeight: 800,
		FontResolver:    &mockFontResolver{},
	}
	resultWithout := block.Layout(containerWithout, constraints)

	// Layout with absolute node inserted between the two flow nodes.
	flowText1b := &document.Text{
		Content:   "First",
		TextStyle: document.Style{FontSize: 12},
	}
	flowText2b := &document.Text{
		Content:   "Second",
		TextStyle: document.Style{FontSize: 12},
	}
	containerWith := &document.Box{
		Content: []document.DocumentNode{
			flowText1b,
			&document.Box{
				Content: []document.DocumentNode{
					&document.Text{
						Content:   "Abs",
						TextStyle: document.Style{FontSize: 12},
					},
				},
				BoxStyle: document.BoxStyle{
					Position: document.Position{
						Mode: document.PositionAbsolute,
						X:    document.Pt(100),
						Y:    document.Pt(100),
					},
				},
			},
			flowText2b,
		},
	}
	resultWith := block.Layout(containerWith, constraints)

	// Flow positions of "First" and "Second" should be the same
	// regardless of the absolute node.
	if len(resultWithout.Children) < 2 {
		t.Fatalf("without: expected at least 2 children, got %d", len(resultWithout.Children))
	}
	// With absolute, should have 3: flow1, flow2, absolute.
	if len(resultWith.Children) < 3 {
		t.Fatalf("with: expected at least 3 children, got %d", len(resultWith.Children))
	}

	// Compare flow node positions.
	for i := 0; i < 2; i++ {
		posWithout := resultWithout.Children[i].Position
		posWidth := resultWith.Children[i].Position
		if !approxEqual(posWithout.X, posWidth.X, 0.01) {
			t.Errorf("child %d X differs: without=%g, with=%g", i, posWithout.X, posWidth.X)
		}
		if !approxEqual(posWithout.Y, posWidth.Y, 0.01) {
			t.Errorf("child %d Y differs: without=%g, with=%g", i, posWithout.Y, posWidth.Y)
		}
	}
}

func TestAbsolutePositioning_WithExplicitSize(t *testing.T) {
	block := NewBlockLayout()

	container := &document.Box{
		Content: []document.DocumentNode{
			&document.Box{
				Content: []document.DocumentNode{
					&document.Text{
						Content:   "Sized",
						TextStyle: document.Style{FontSize: 12},
					},
				},
				BoxStyle: document.BoxStyle{
					Width:  document.Pt(200),
					Height: document.Pt(100),
					Position: document.Position{
						Mode: document.PositionAbsolute,
						X:    document.Pt(50),
						Y:    document.Pt(50),
					},
				},
			},
		},
	}

	constraints := Constraints{
		AvailableWidth:  595.28,
		AvailableHeight: 841.89,
		FontResolver:    &mockFontResolver{},
	}

	result := block.Layout(container, constraints)

	if len(result.Children) != 1 {
		t.Fatalf("expected 1 child, got %d", len(result.Children))
	}

	absNode := result.Children[0]
	if !approxEqual(absNode.Size.Width, 200, 1.0) {
		t.Errorf("width: got %g, want 200", absNode.Size.Width)
	}
}

func TestAbsolutePositioning_OriginPage(t *testing.T) {
	// Test that OriginPage is stored correctly (the coordinate
	// adjustment happens in the paginator, tested via integration).
	box := &document.Box{
		BoxStyle: document.BoxStyle{
			Position: document.Position{
				Mode:   document.PositionAbsolute,
				X:      document.Mm(10),
				Y:      document.Mm(10),
				Origin: document.OriginPage,
			},
		},
	}

	if box.BoxStyle.Position.Origin != document.OriginPage {
		t.Error("expected OriginPage")
	}
}

func TestAdjustAbsoluteOrigins(t *testing.T) {
	marginX := 56.69 // 20mm
	marginY := 56.69

	absBox := &document.Box{
		BoxStyle: document.BoxStyle{
			Position: document.Position{
				Mode:   document.PositionAbsolute,
				X:      document.Mm(10),
				Y:      document.Mm(10),
				Origin: document.OriginPage,
			},
		},
	}

	normalBox := &document.Box{
		BoxStyle: document.BoxStyle{
			Position: document.Position{
				Mode: document.PositionAbsolute,
				X:    document.Mm(10),
				Y:    document.Mm(10),
				// OriginContent (default, no OriginPage)
			},
		},
	}

	textNode := &document.Text{
		Content:   "Not a box",
		TextStyle: document.Style{FontSize: 12},
	}

	nodes := []PlacedNode{
		{Node: absBox, Position: document.Point{X: 100, Y: 200}},
		{Node: normalBox, Position: document.Point{X: 100, Y: 200}},
		{Node: textNode, Position: document.Point{X: 50, Y: 50}},
	}

	adjustAbsoluteOrigins(nodes, marginX, marginY)

	// OriginPage node should have margins subtracted.
	if !approxEqual(nodes[0].Position.X, 100-marginX, 0.01) {
		t.Errorf("OriginPage X: got %g, want %g", nodes[0].Position.X, 100-marginX)
	}
	if !approxEqual(nodes[0].Position.Y, 200-marginY, 0.01) {
		t.Errorf("OriginPage Y: got %g, want %g", nodes[0].Position.Y, 200-marginY)
	}

	// OriginContent node should be unchanged.
	if !approxEqual(nodes[1].Position.X, 100, 0.01) {
		t.Errorf("OriginContent X: got %g, want 100", nodes[1].Position.X)
	}
	if !approxEqual(nodes[1].Position.Y, 200, 0.01) {
		t.Errorf("OriginContent Y: got %g, want 200", nodes[1].Position.Y)
	}

	// Text node should be unchanged.
	if !approxEqual(nodes[2].Position.X, 50, 0.01) {
		t.Errorf("Text X: got %g, want 50", nodes[2].Position.X)
	}
}
