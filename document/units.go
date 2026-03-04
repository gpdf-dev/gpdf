package document

// Value is a dimension expressed as a numeric amount with an associated unit.
// Use the convenience constructors (Pt, Mm, In, Cm, Em, Pct) to create values.
type Value struct {
	Amount float64
	Unit   Unit
}

// Unit identifies the measurement system for a Value.
type Unit int

const (
	// UnitPt is the PDF point (1/72 inch), the native PDF coordinate unit.
	UnitPt Unit = iota
	// UnitMm is the millimeter.
	UnitMm
	// UnitIn is the inch.
	UnitIn
	// UnitCm is the centimeter.
	UnitCm
	// UnitEm is relative to the current font size.
	UnitEm
	// UnitPct is a percentage relative to the parent dimension.
	UnitPct
	// UnitAuto indicates that the value should be calculated automatically.
	UnitAuto
)

// Pt creates a Value in PDF points (1/72 inch).
func Pt(v float64) Value { return Value{Amount: v, Unit: UnitPt} }

// Mm creates a Value in millimeters.
func Mm(v float64) Value { return Value{Amount: v, Unit: UnitMm} }

// In creates a Value in inches.
func In(v float64) Value { return Value{Amount: v, Unit: UnitIn} }

// Cm creates a Value in centimeters.
func Cm(v float64) Value { return Value{Amount: v, Unit: UnitCm} }

// Em creates a Value relative to the current font size.
func Em(v float64) Value { return Value{Amount: v, Unit: UnitEm} }

// Pct creates a Value as a percentage of the parent dimension.
func Pct(v float64) Value { return Value{Amount: v, Unit: UnitPct} }

// Auto is a sentinel Value indicating that the dimension should be
// calculated automatically by the layout engine.
var Auto = Value{Unit: UnitAuto}

// Resolve converts a Value to PDF points using the given parent size
// (for percentage calculations) and font size (for em calculations).
//
// Conversion factors:
//   - Pt:  amount (identity)
//   - Mm:  amount * 2.83465
//   - In:  amount * 72
//   - Cm:  amount * 28.3465
//   - Em:  amount * fontSize
//   - Pct: amount / 100 * parentSize
//   - Auto: 0
func (v Value) Resolve(parentSize, fontSize float64) float64 {
	switch v.Unit {
	case UnitPt:
		return v.Amount
	case UnitMm:
		return v.Amount * 2.83465
	case UnitIn:
		return v.Amount * 72
	case UnitCm:
		return v.Amount * 28.3465
	case UnitEm:
		return v.Amount * fontSize
	case UnitPct:
		return v.Amount / 100 * parentSize
	case UnitAuto:
		return 0
	default:
		return v.Amount
	}
}

// IsAuto reports whether the value uses the Auto unit, indicating the
// dimension should be determined by the layout engine.
func (v Value) IsAuto() bool { return v.Unit == UnitAuto }

// Edges represents four-sided dimension values, following the CSS box model
// convention of Top, Right, Bottom, Left.
type Edges struct {
	Top, Right, Bottom, Left Value
}

// UniformEdges creates Edges with the same value applied to all four sides.
func UniformEdges(v Value) Edges {
	return Edges{Top: v, Right: v, Bottom: v, Left: v}
}

// Resolve converts all edge values to points. Horizontal edges (Left, Right)
// are resolved against parentWidth; vertical edges (Top, Bottom) are resolved
// against parentHeight.
func (e Edges) Resolve(parentWidth, parentHeight, fontSize float64) ResolvedEdges {
	return ResolvedEdges{
		Top:    e.Top.Resolve(parentHeight, fontSize),
		Right:  e.Right.Resolve(parentWidth, fontSize),
		Bottom: e.Bottom.Resolve(parentHeight, fontSize),
		Left:   e.Left.Resolve(parentWidth, fontSize),
	}
}

// ResolvedEdges holds four-sided dimensions that have been resolved to
// PDF points.
type ResolvedEdges struct {
	Top, Right, Bottom, Left float64
}

// Horizontal returns the sum of the left and right edges.
func (re ResolvedEdges) Horizontal() float64 {
	return re.Left + re.Right
}

// Vertical returns the sum of the top and bottom edges.
func (re ResolvedEdges) Vertical() float64 {
	return re.Top + re.Bottom
}

// Point represents a position in 2D space, expressed in PDF points.
type Point struct {
	X, Y float64
}

// Size represents a 2D dimension (width and height) in PDF points.
type Size struct {
	Width, Height float64
}

// Rectangle represents an axis-aligned rectangle defined by its origin
// (top-left corner in layout coordinates) and dimensions.
type Rectangle struct {
	X, Y, Width, Height float64
}

// Predefined page sizes expressed in PDF points (1/72 inch).
var (
	// A4 is the ISO A4 page size (210mm x 297mm).
	A4 = Size{Width: 595.28, Height: 841.89}
	// A3 is the ISO A3 page size (297mm x 420mm).
	A3 = Size{Width: 841.89, Height: 1190.55}
	// Letter is the US Letter page size (8.5" x 11").
	Letter = Size{Width: 612, Height: 792}
	// Legal is the US Legal page size (8.5" x 14").
	Legal = Size{Width: 612, Height: 1008}
)
