// Package layout provides layout engines that calculate the position and
// size of document nodes within the available space. The package converts
// the abstract document tree into a flat list of placed nodes that a
// renderer can draw.
package layout

import "github.com/gpdf-dev/gpdf/document"

// Engine performs layout calculation on a document tree, converting logical
// nodes into positioned, sized outputs within the given constraints.
type Engine interface {
	// Layout calculates the position and size of the given node and its
	// children, subject to the provided constraints. If the content does
	// not fit, the Overflow field in the Result carries the remaining
	// content for placement on subsequent pages.
	Layout(node document.DocumentNode, constraints Constraints) Result
}

// Constraints define the available space and resources for layout.
type Constraints struct {
	// AvailableWidth is the maximum horizontal space in points.
	AvailableWidth float64
	// AvailableHeight is the maximum vertical space in points.
	AvailableHeight float64
	// PageWidth is the physical page width in points, used to resolve
	// absolute positions with OriginPage. Zero means fall back to
	// AvailableWidth (single-box layout outside a paginator).
	PageWidth float64
	// PageHeight is the physical page height in points, used to resolve
	// absolute positions with OriginPage. Zero means fall back to
	// AvailableHeight.
	PageHeight float64
	// FontResolver provides font metrics and text measurement.
	FontResolver FontResolver
}

// Result is the output of a layout calculation.
type Result struct {
	// Bounds is the rectangle occupied by the laid-out node.
	Bounds document.Rectangle
	// Children holds the positioned child nodes.
	Children []PlacedNode
	// Overflow holds content that did not fit within the constraints.
	// A nil value indicates that all content was placed.
	Overflow document.DocumentNode
}

// PlacedNode is a document node together with its calculated position,
// size, and recursively placed children.
type PlacedNode struct {
	// Node is the original document node.
	Node document.DocumentNode
	// Position is the top-left corner of the node in layout coordinates.
	Position document.Point
	// Size is the calculated width and height of the node.
	Size document.Size
	// Children holds the placed children of this node.
	Children []PlacedNode
}

// FontResolver resolves font families and weights to concrete font
// metrics, and provides text measurement and line-breaking services.
// This abstraction keeps the layout package independent of the low-level
// pdf/font package.
type FontResolver interface {
	// Resolve looks up a font by family name, weight, and italic flag,
	// returning a resolved font handle suitable for measurement.
	Resolve(family string, weight document.FontWeight, italic bool) ResolvedFont
	// MeasureString returns the width in points of the given text at the
	// specified font size.
	MeasureString(font ResolvedFont, text string, size float64) float64
	// LineBreak splits text into lines that fit within maxWidth at the
	// given font size.
	LineBreak(font ResolvedFont, text string, size float64, maxWidth float64) []string
}

// ResolvedFont is a handle to a concrete font, carrying an identifier
// and pre-resolved metrics.
type ResolvedFont struct {
	// ID is a unique identifier for this font variant (e.g., "Helvetica-Bold").
	ID string
	// Metrics holds the font's typographic measurements.
	Metrics FontMetrics
}

// FontMetrics holds typographic measurements for a resolved font,
// expressed as proportions or absolute values that can be scaled by
// the desired font size.
type FontMetrics struct {
	// Ascender is the distance from the baseline to the top of tall
	// characters, in points per unit of font size.
	Ascender float64
	// Descender is the distance from the baseline to the bottom of
	// descending characters (typically negative), in points per unit.
	Descender float64
	// LineHeight is the recommended distance between baselines, in
	// points per unit of font size.
	LineHeight float64
	// CapHeight is the height of capital letters, in points per unit.
	CapHeight float64
}
