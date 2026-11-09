package document

import "github.com/gpdf-dev/gpdf/pdf"

// Box is the fundamental layout container, implementing the CSS-like box
// model with margins, padding, borders, and size constraints. Boxes stack
// their children vertically by default.
type Box struct {
	// Content holds the child nodes rendered inside this box.
	Content []DocumentNode
	// BoxStyle controls the dimensions, spacing, and decoration of the box.
	BoxStyle BoxStyle
	// BreakPolicy controls page break behavior around and within this box.
	BreakPolicy BreakPolicy
}

// Direction controls how a Box arranges its children.
type Direction int

const (
	// DirectionVertical arranges children top-to-bottom (default).
	DirectionVertical Direction = iota
	// DirectionHorizontal arranges children left-to-right.
	DirectionHorizontal
)

// BoxStyle defines the dimensional and decorative properties of a Box.
type BoxStyle struct {
	Width      Value
	Height     Value
	MinWidth   Value
	MaxWidth   Value
	MinHeight  Value
	MaxHeight  Value
	Margin     Edges
	Padding    Edges
	Border     BorderEdges
	Background *pdf.Color
	Direction  Direction
	Position   Position
	// BorderRadius rounds the corners of the background and of a uniform
	// border. See [Style.BorderRadius].
	BorderRadius Value
	// Opacity is the constant alpha applied to the box and its content.
	// See [Style.Opacity].
	Opacity float64
	// Link is a URI that the whole box area links to. See [Style.Link].
	Link string
}

// NodeType returns NodeBox.
func (b *Box) NodeType() NodeType { return NodeBox }

// Children returns the box's content nodes.
func (b *Box) Children() []DocumentNode { return b.Content }

// Style constructs a Style from the BoxStyle's spacing and decoration fields.
func (b *Box) Style() Style {
	return Style{
		Margin:       b.BoxStyle.Margin,
		Padding:      b.BoxStyle.Padding,
		Border:       b.BoxStyle.Border,
		Background:   b.BoxStyle.Background,
		BorderRadius: b.BoxStyle.BorderRadius,
		Opacity:      b.BoxStyle.Opacity,
		Link:         b.BoxStyle.Link,
	}
}
