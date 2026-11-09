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

// Landscape returns the size in landscape orientation (width >= height).
// Sizes that are already landscape or square are returned unchanged, so
// the call is idempotent.
//
//	template.WithPageSize(document.A4.Landscape())
func (s Size) Landscape() Size {
	if s.Width >= s.Height {
		return s
	}
	return Size{Width: s.Height, Height: s.Width}
}

// Rectangle represents an axis-aligned rectangle defined by its origin
// (top-left corner in layout coordinates) and dimensions.
type Rectangle struct {
	X, Y, Width, Height float64
}

// Predefined page sizes expressed in PDF points (1/72 inch), rounded to
// two decimal places. All sizes are portrait; use [Size.Landscape] for
// the landscape orientation.
//
// The A, B and C series follow ISO 216 / ISO 269. Note that the ISO B
// series differs from the JIS B series commonly used in Japan (e.g. JIS
// B5 is 182mm x 257mm); use [Size] with [Mm] for JIS sizes.
var (
	// A0 is the ISO A0 page size (841mm x 1189mm).
	A0 = Size{Width: 2383.94, Height: 3370.39}
	// A1 is the ISO A1 page size (594mm x 841mm).
	A1 = Size{Width: 1683.78, Height: 2383.94}
	// A2 is the ISO A2 page size (420mm x 594mm).
	A2 = Size{Width: 1190.55, Height: 1683.78}
	// A3 is the ISO A3 page size (297mm x 420mm).
	A3 = Size{Width: 841.89, Height: 1190.55}
	// A4 is the ISO A4 page size (210mm x 297mm).
	A4 = Size{Width: 595.28, Height: 841.89}
	// A5 is the ISO A5 page size (148mm x 210mm).
	A5 = Size{Width: 419.53, Height: 595.28}
	// A6 is the ISO A6 page size (105mm x 148mm).
	A6 = Size{Width: 297.64, Height: 419.53}
	// A7 is the ISO A7 page size (74mm x 105mm).
	A7 = Size{Width: 209.76, Height: 297.64}
	// A8 is the ISO A8 page size (52mm x 74mm).
	A8 = Size{Width: 147.40, Height: 209.76}
	// A9 is the ISO A9 page size (37mm x 52mm).
	A9 = Size{Width: 104.88, Height: 147.40}
	// A10 is the ISO A10 page size (26mm x 37mm).
	A10 = Size{Width: 73.70, Height: 104.88}

	// B0 is the ISO B0 page size (1000mm x 1414mm).
	B0 = Size{Width: 2834.65, Height: 4008.19}
	// B1 is the ISO B1 page size (707mm x 1000mm).
	B1 = Size{Width: 2004.09, Height: 2834.65}
	// B2 is the ISO B2 page size (500mm x 707mm).
	B2 = Size{Width: 1417.32, Height: 2004.09}
	// B3 is the ISO B3 page size (353mm x 500mm).
	B3 = Size{Width: 1000.63, Height: 1417.32}
	// B4 is the ISO B4 page size (250mm x 353mm).
	B4 = Size{Width: 708.66, Height: 1000.63}
	// B5 is the ISO B5 page size (176mm x 250mm).
	B5 = Size{Width: 498.90, Height: 708.66}
	// B6 is the ISO B6 page size (125mm x 176mm).
	B6 = Size{Width: 354.33, Height: 498.90}
	// B7 is the ISO B7 page size (88mm x 125mm).
	B7 = Size{Width: 249.45, Height: 354.33}
	// B8 is the ISO B8 page size (62mm x 88mm).
	B8 = Size{Width: 175.75, Height: 249.45}
	// B9 is the ISO B9 page size (44mm x 62mm).
	B9 = Size{Width: 124.72, Height: 175.75}
	// B10 is the ISO B10 page size (31mm x 44mm).
	B10 = Size{Width: 87.87, Height: 124.72}

	// C0 is the ISO C0 envelope size (917mm x 1297mm).
	C0 = Size{Width: 2599.37, Height: 3676.54}
	// C1 is the ISO C1 envelope size (648mm x 917mm).
	C1 = Size{Width: 1836.85, Height: 2599.37}
	// C2 is the ISO C2 envelope size (458mm x 648mm).
	C2 = Size{Width: 1298.27, Height: 1836.85}
	// C3 is the ISO C3 envelope size (324mm x 458mm).
	C3 = Size{Width: 918.43, Height: 1298.27}
	// C4 is the ISO C4 envelope size (229mm x 324mm).
	C4 = Size{Width: 649.13, Height: 918.43}
	// C5 is the ISO C5 envelope size (162mm x 229mm).
	C5 = Size{Width: 459.21, Height: 649.13}
	// C6 is the ISO C6 envelope size (114mm x 162mm).
	C6 = Size{Width: 323.15, Height: 459.21}
	// C7 is the ISO C7 envelope size (81mm x 114mm).
	C7 = Size{Width: 229.61, Height: 323.15}
	// C8 is the ISO C8 envelope size (57mm x 81mm).
	C8 = Size{Width: 161.57, Height: 229.61}
	// C9 is the ISO C9 envelope size (40mm x 57mm).
	C9 = Size{Width: 113.39, Height: 161.57}
	// C10 is the ISO C10 envelope size (28mm x 40mm).
	C10 = Size{Width: 79.37, Height: 113.39}

	// DL is the ISO DL envelope size (110mm x 220mm).
	DL = Size{Width: 311.81, Height: 623.62}
	// Letter is the US Letter page size (8.5" x 11").
	Letter = Size{Width: 612, Height: 792}
	// Legal is the US Legal page size (8.5" x 14").
	Legal = Size{Width: 612, Height: 1008}
)
