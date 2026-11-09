package render

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/gpdf-dev/gpdf/document"
	"github.com/gpdf-dev/gpdf/pdf"
	"github.com/gpdf-dev/gpdf/svg"
)

// bezierCircle is the control-point distance (as a fraction of the radius)
// that approximates a quarter circle with a single cubic Bézier curve.
const bezierCircle = 0.5522847498

// roundedRect returns a closed path tracing rect with each corner replaced
// by a quarter circle of radius r. The radius is clamped to half the width
// and height. Coordinates are in layout space (origin top-left).
func roundedRect(rect document.Rectangle, r float64) document.Path {
	r = min(r, rect.Width/2, rect.Height/2)
	if r < 0 {
		r = 0
	}
	k := r * bezierCircle
	x0, y0, x1, y1 := rect.X, rect.Y, rect.X+rect.Width, rect.Y+rect.Height
	pt := func(x, y float64) document.Point { return document.Point{X: x, Y: y} }
	curve := func(c1, c2, end document.Point) document.PathSegment {
		return document.PathSegment{Op: document.PathCurveTo, Points: []document.Point{c1, c2, end}}
	}
	line := func(p document.Point) document.PathSegment {
		return document.PathSegment{Op: document.PathLineTo, Points: []document.Point{p}}
	}
	return document.Path{Segments: []document.PathSegment{
		{Op: document.PathMoveTo, Points: []document.Point{pt(x0+r, y0)}},
		line(pt(x1-r, y0)),
		curve(pt(x1-r+k, y0), pt(x1, y0+r-k), pt(x1, y0+r)),
		line(pt(x1, y1-r)),
		curve(pt(x1, y1-r+k), pt(x1-r+k, y1), pt(x1-r, y1)),
		line(pt(x0+r, y1)),
		curve(pt(x0+r-k, y1), pt(x0, y1-r+k), pt(x0, y1-r)),
		line(pt(x0, y0+r)),
		curve(pt(x0, y0+r-k), pt(x0+r-k, y0), pt(x0+r, y0)),
		{Op: document.PathClose},
	}}
}

// writePathOps emits PDF path-construction operators (m, l, c, h) for each
// segment, converting Y coordinates from layout space to PDF space.
func writePathOps(buf *strings.Builder, path document.Path, pageHeight float64) {
	for _, seg := range path.Segments {
		switch seg.Op {
		case document.PathMoveTo:
			if len(seg.Points) >= 1 {
				p := seg.Points[0]
				fmt.Fprintf(buf, "%g %g m\n", p.X, pageHeight-p.Y)
			}
		case document.PathLineTo:
			if len(seg.Points) >= 1 {
				p := seg.Points[0]
				fmt.Fprintf(buf, "%g %g l\n", p.X, pageHeight-p.Y)
			}
		case document.PathCurveTo:
			if len(seg.Points) >= 3 {
				c1, c2, ep := seg.Points[0], seg.Points[1], seg.Points[2]
				fmt.Fprintf(buf, "%g %g %g %g %g %g c\n",
					c1.X, pageHeight-c1.Y,
					c2.X, pageHeight-c2.Y,
					ep.X, pageHeight-ep.Y)
			}
		case document.PathClose:
			buf.WriteString("h\n")
		}
	}
}

// writeRoundedFill emits a self-contained (q … Q) fill of a rounded
// rectangle in the given color.
func writeRoundedFill(buf *strings.Builder, rect document.Rectangle, radius float64, color pdf.Color, pageHeight float64) {
	buf.WriteString("q\n")
	buf.WriteString(color.FillColorCmd())
	buf.WriteByte('\n')
	writePathOps(buf, roundedRect(rect, radius), pageHeight)
	buf.WriteString("f\nQ\n")
}

// writeRoundedClip begins a graphics-state group (q) clipped to a rounded
// rectangle. The caller must close it with "Q".
func writeRoundedClip(buf *strings.Builder, rect document.Rectangle, radius float64, pageHeight float64) {
	buf.WriteString("q\n")
	writePathOps(buf, roundedRect(rect, radius), pageHeight)
	buf.WriteString("W n\n")
}

// uniformBorder reports whether all four sides of b are drawn with the same
// width, style and color, returning that side. Rounded corners are only
// stroked for uniform borders; mixed borders keep square corners.
func uniformBorder(b document.BorderEdges) (document.BorderSide, bool) {
	if b.Top.Style == document.BorderNone {
		return document.BorderSide{}, false
	}
	if b.Top != b.Right || b.Top != b.Bottom || b.Top != b.Left {
		return document.BorderSide{}, false
	}
	return b.Top, true
}

// writeRoundedBorder strokes a uniform border of width bw along the inside
// of rect, following its rounded corners.
func writeRoundedBorder(buf *strings.Builder, rect document.Rectangle, radius, bw float64, color pdf.Color, pageHeight float64) {
	half := bw / 2
	inset := document.Rectangle{
		X:      rect.X + half,
		Y:      rect.Y + half,
		Width:  rect.Width - bw,
		Height: rect.Height - bw,
	}
	if inset.Width < 0 || inset.Height < 0 {
		return
	}
	buf.WriteString("q\n")
	buf.WriteString(color.StrokeColorCmd())
	buf.WriteByte('\n')
	fmt.Fprintf(buf, "%g w\n", bw)
	writePathOps(buf, roundedRect(inset, max(radius-half, 0)), pageHeight)
	buf.WriteString("S\nQ\n")
}

// normalizeOpacity converts a Style.Opacity value into a constant alpha.
// It reports false when the value means "unset" (zero or negative).
func normalizeOpacity(v float64) (float64, bool) {
	if v <= 0 {
		return 0, false
	}
	if v > 1 {
		v = 1
	}
	return v, true
}

// alphaKey returns a stable string key for an alpha value, used to
// deduplicate ExtGState resources.
func alphaKey(a float64) string {
	return strconv.FormatFloat(a, 'f', 4, 64)
}

// extGStateDict builds an ExtGState dictionary that sets both the
// stroking (CA) and non-stroking (ca) constant alpha.
func extGStateDict(a float64) pdf.Dict {
	return pdf.Dict{
		pdf.Name("Type"): pdf.Name("ExtGState"),
		pdf.Name("CA"):   pdf.Real(a),
		pdf.Name("ca"):   pdf.Real(a),
	}
}

// linkAnnotation builds a URI link annotation covering the layout-space
// rectangle rect on a page of the given height.
func linkAnnotation(rect document.Rectangle, uri string, pageHeight float64) pdf.Dict {
	return pdf.Dict{
		pdf.Name("Type"):    pdf.Name("Annot"),
		pdf.Name("Subtype"): pdf.Name("Link"),
		pdf.Name("Rect"): pdf.Rectangle{
			LLX: rect.X,
			LLY: pageHeight - rect.Y - rect.Height,
			URX: rect.X + rect.Width,
			URY: pageHeight - rect.Y,
		},
		pdf.Name("Border"): pdf.Array{pdf.Integer(0), pdf.Integer(0), pdf.Integer(0)},
		pdf.Name("F"):      pdf.Integer(4), // Print flag (required by PDF/A)
		pdf.Name("A"): pdf.Dict{
			pdf.Name("Type"): pdf.Name("Action"),
			pdf.Name("S"):    pdf.Name("URI"),
			pdf.Name("URI"):  pdf.LiteralString(uri),
		},
	}
}

// nodeEmitsLink reports whether a placed node should produce its own link
// annotation. Text containers and rich-text lines are skipped because
// their leaf children (one per line or run) carry the same link and give
// tighter clickable areas.
func nodeEmitsLink(t document.NodeType, leaf bool) bool {
	switch t {
	case document.NodeText:
		return leaf
	case document.NodeRichText:
		return false
	default:
		return true
	}
}

// fitRect computes where an image of intrinsic size (iw, ih) is drawn
// inside box for the given fit mode. The image is centered on the box.
func fitRect(iw, ih float64, box document.Rectangle, mode document.ImageFitMode) document.Rectangle {
	if iw <= 0 || ih <= 0 {
		return box
	}
	var w, h float64
	switch mode {
	case document.FitStretch:
		return box
	case document.FitOriginal:
		w, h = iw, ih
	case document.FitCover:
		s := max(box.Width/iw, box.Height/ih)
		w, h = iw*s, ih*s
	default: // FitContain
		s := min(box.Width/iw, box.Height/ih)
		w, h = iw*s, ih*s
	}
	return document.Rectangle{
		X:      box.X + (box.Width-w)/2,
		Y:      box.Y + (box.Height-h)/2,
		Width:  w,
		Height: h,
	}
}

// svgFormContent converts SVG data into Form XObject content drawn in the
// unit square, plus the form's resource dictionary.
func svgFormContent(data []byte) ([]byte, pdf.Dict, error) {
	img, err := svg.Parse(data)
	if err != nil {
		return nil, nil, fmt.Errorf("render: failed to parse SVG: %w", err)
	}
	out := img.Render()
	resources := pdf.Dict{}
	if len(out.ExtGStates) > 0 {
		gs := make(pdf.Dict, len(out.ExtGStates))
		for name, a := range out.ExtGStates {
			gs[pdf.Name(name)] = pdf.Dict{
				pdf.Name("Type"): pdf.Name("ExtGState"),
				pdf.Name("ca"):   pdf.Real(a.Fill),
				pdf.Name("CA"):   pdf.Real(a.Stroke),
			}
		}
		resources[pdf.Name("ExtGState")] = gs
	}
	return out.Content, resources, nil
}

// svgFormDict returns the stream dictionary of an SVG Form XObject.
func svgFormDict(resources pdf.Dict) pdf.Dict {
	return pdf.Dict{
		pdf.Name("Type"):      pdf.Name("XObject"),
		pdf.Name("Subtype"):   pdf.Name("Form"),
		pdf.Name("BBox"):      pdf.Rectangle{URX: 1, URY: 1},
		pdf.Name("Resources"): resources,
	}
}
