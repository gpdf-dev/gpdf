package document

import "github.com/gpdf-dev/gpdf/pdf"

// Style holds the complete set of visual properties that can be applied to
// a document node. It follows the CSS box model conventions: font, color,
// text alignment, margins, padding, and borders.
type Style struct {
	// Font properties
	FontFamily string
	FontSize   float64    // in points
	FontWeight FontWeight // e.g., WeightNormal (400), WeightBold (700)
	FontStyle  FontStyle  // Normal or Italic

	// Color properties
	Color      pdf.Color  // text (foreground) color
	Background *pdf.Color // background fill color; nil means transparent

	// Text properties
	TextAlign      TextAlign      // horizontal text alignment
	LineHeight     float64        // line height multiplier (e.g., 1.5 for 150%)
	WordSpacing    float64        // extra space added to each word gap in points (used by justify)
	LetterSpacing  float64        // extra space between characters in points (PDF Tc operator)
	TextIndent     Value          // first-line indentation
	TextDecoration TextDecoration // bitmask: underline, strikethrough, overline
	VerticalAlign  VerticalAlign  // vertical alignment in table cells

	// Box model
	Margin  Edges
	Padding Edges
	Border  BorderEdges

	// BorderRadius rounds the corners of the background and of a uniform
	// border, and clips images to the rounded shape. The radius is clamped
	// to half the width and height. The zero value draws square corners.
	BorderRadius Value

	// Opacity is the constant alpha (0.0–1.0) applied to the node and its
	// descendants, for fills, strokes, text and images alike. The zero
	// value means "unset" (fully opaque). A descendant with its own
	// Opacity replaces the inherited value rather than multiplying it.
	// Note that PDF/A-1 forbids transparency.
	Opacity float64

	// Link is a URI (e.g. "https://gpdf.dev") that the node's area links
	// to. For wrapped text, each line becomes its own clickable area.
	Link string
}

// DefaultStyle returns a Style with sensible defaults: 12pt black text,
// left-aligned with a 1.2 line height.
func DefaultStyle() Style {
	return Style{
		FontSize:   12,
		FontWeight: WeightNormal,
		FontStyle:  StyleNormal,
		Color:      pdf.Black,
		TextAlign:  AlignLeft,
		LineHeight: 1.2,
	}
}

// FontWeight represents the weight (boldness) of a font, using the
// standard CSS numeric scale where 400 is normal and 700 is bold.
type FontWeight int

const (
	// WeightNormal is the standard font weight (CSS 400).
	WeightNormal FontWeight = 400
	// WeightBold is the bold font weight (CSS 700).
	WeightBold FontWeight = 700
)

// FontStyle selects between normal (upright) and italic typeface variants.
type FontStyle int

const (
	// StyleNormal selects the upright typeface variant.
	StyleNormal FontStyle = iota
	// StyleItalic selects the italic typeface variant.
	StyleItalic
)

// TextAlign specifies horizontal alignment for text content.
type TextAlign int

const (
	// AlignLeft aligns text to the left edge.
	AlignLeft TextAlign = iota
	// AlignCenter centers text horizontally.
	AlignCenter
	// AlignRight aligns text to the right edge.
	AlignRight
	// AlignJustify stretches text to fill the full width.
	AlignJustify
)

// TextDecoration is a bitmask specifying text decoration lines.
type TextDecoration uint8

const (
	// DecorationNone indicates no text decoration.
	DecorationNone TextDecoration = 0
	// DecorationUnderline draws a line below the text baseline.
	DecorationUnderline TextDecoration = 1
	// DecorationStrikethrough draws a line through the middle of the text.
	DecorationStrikethrough TextDecoration = 2
	// DecorationOverline draws a line above the text.
	DecorationOverline TextDecoration = 4
)

// VerticalAlign specifies vertical alignment within a container (e.g., table cell).
type VerticalAlign int

const (
	// VAlignTop aligns content to the top (default).
	VAlignTop VerticalAlign = iota
	// VAlignMiddle centers content vertically.
	VAlignMiddle
	// VAlignBottom aligns content to the bottom.
	VAlignBottom
)

// BorderEdges defines border styling for each side of a box.
type BorderEdges struct {
	Top, Right, Bottom, Left BorderSide
}

// BorderSide defines the visual properties of a single border edge.
type BorderSide struct {
	Width Value
	Style BorderStyle
	Color pdf.Color
}

// BorderStyle specifies the line style for a border edge.
type BorderStyle int

const (
	// BorderNone indicates no border is drawn.
	BorderNone BorderStyle = iota
	// BorderSolid draws a continuous solid line.
	BorderSolid
	// BorderDashed draws a dashed line.
	BorderDashed
	// BorderDotted draws a dotted line.
	BorderDotted
)

// UniformBorder creates a BorderEdges with the same width, style, and
// color applied to all four sides.
func UniformBorder(width Value, style BorderStyle, color pdf.Color) BorderEdges {
	side := BorderSide{Width: width, Style: style, Color: color}
	return BorderEdges{
		Top:    side,
		Right:  side,
		Bottom: side,
		Left:   side,
	}
}
