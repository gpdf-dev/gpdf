package render

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/gpdf-dev/gpdf/pdf"
)

// svgFormResult holds the output of converting an SVG to a PDF Form XObject.
type svgFormResult struct {
	Content   []byte   // PDF content stream operators
	ViewW     float64  // SVG viewport width in user units (for BBox)
	ViewH     float64  // SVG viewport height in user units (for BBox)
	Resources pdf.Dict // optional resources (e.g., ExtGState for opacity)
}

// svgToFormContent converts SVG data to a PDF Form XObject content stream.
// The resulting Form XObject uses a normalizing matrix so that it maps to the
// [0,1]×[0,1] unit square (Y-flipped), making it compatible with the existing
// RenderImage placement code that uses [width 0 0 height x y cm].
func svgToFormContent(data []byte) (*svgFormResult, error) {
	root, err := parseSVGXML(data)
	if err != nil {
		return nil, fmt.Errorf("svg: parse error: %w", err)
	}
	if root == nil {
		return nil, fmt.Errorf("svg: no root element found")
	}

	viewW, viewH := svgViewBox(root.attrs)
	if viewW <= 0 || viewH <= 0 {
		viewW, viewH = 100, 100
	}

	var b svgBuilder
	b.gsMap = make(map[[2]float64]string)

	for _, child := range root.children {
		b.renderElem(child, svgDefaultStyle, viewW, viewH)
	}

	resources := b.buildResources()
	return &svgFormResult{
		Content:   []byte(b.buf.String()),
		ViewW:     viewW,
		ViewH:     viewH,
		Resources: resources,
	}, nil
}

// parseSVGDimensions returns the SVG viewport dimensions from the root element.
// Called from template/grid.go for setting ImageSource.Width/Height.
func parseSVGDimensions(data []byte) (float64, float64) {
	dec := xml.NewDecoder(bytes.NewReader(data))
	for {
		tok, err := dec.Token()
		if err != nil {
			break
		}
		se, ok := tok.(xml.StartElement)
		if !ok {
			continue
		}
		if strings.ToLower(se.Name.Local) == "svg" {
			attrs := xmlAttrsMap(se.Attr)
			w, h := svgViewBox(attrs)
			if w > 0 && h > 0 {
				return w, h
			}
		}
		break // only inspect the first element
	}
	return 100, 100
}

// svgViewBox extracts viewport dimensions from SVG root attributes.
func svgViewBox(attrs map[string]string) (w, h float64) {
	if vb, ok := attrs["viewbox"]; ok {
		parts := strings.Fields(strings.ReplaceAll(vb, ",", " "))
		if len(parts) == 4 {
			vw, _ := strconv.ParseFloat(parts[2], 64)
			vh, _ := strconv.ParseFloat(parts[3], 64)
			if vw > 0 && vh > 0 {
				return vw, vh
			}
		}
	}
	if ws, ok := attrs["width"]; ok {
		if hs, ok2 := attrs["height"]; ok2 {
			wv := parseSVGLength(ws)
			hv := parseSVGLength(hs)
			if wv > 0 && hv > 0 {
				return wv, hv
			}
		}
	}
	return 0, 0
}

// ============================================================
// XML parser
// ============================================================

type svgElem struct {
	name     string
	attrs    map[string]string
	children []*svgElem
}

func parseSVGXML(data []byte) (*svgElem, error) {
	dec := xml.NewDecoder(bytes.NewReader(data))
	dec.Strict = false
	var stack []*svgElem
	var root *svgElem

	for {
		tok, err := dec.Token()
		if err != nil {
			break
		}
		switch t := tok.(type) {
		case xml.StartElement:
			elem := &svgElem{
				name:  strings.ToLower(t.Name.Local),
				attrs: xmlAttrsMap(t.Attr),
			}
			if len(stack) == 0 {
				root = elem
			} else {
				parent := stack[len(stack)-1]
				parent.children = append(parent.children, elem)
			}
			stack = append(stack, elem)
		case xml.EndElement:
			if len(stack) > 0 {
				stack = stack[:len(stack)-1]
			}
		}
	}
	return root, nil
}

func xmlAttrsMap(attrs []xml.Attr) map[string]string {
	m := make(map[string]string, len(attrs))
	for _, a := range attrs {
		m[strings.ToLower(a.Name.Local)] = a.Value
	}
	return m
}

// ============================================================
// Style
// ============================================================

type svgColor struct {
	R, G, B float64
	None    bool // no paint (fill="none" or stroke="none")
}

type svgPaint struct {
	Color svgColor
	Set   bool
}

type svgStyle struct {
	Fill          svgPaint
	Stroke        svgPaint
	StrokeWidth   float64
	Opacity       float64
	FillOpacity   float64
	StrokeOpacity float64
}

var svgDefaultStyle = svgStyle{
	Fill:          svgPaint{Color: svgColor{0, 0, 0, false}, Set: true},
	Stroke:        svgPaint{Color: svgColor{None: true}, Set: true},
	StrokeWidth:   1,
	Opacity:       1,
	FillOpacity:   1,
	StrokeOpacity: 1,
}

// parseElemStyle computes the effective style by merging parent with this element's attributes.
func parseElemStyle(attrs map[string]string, parent svgStyle) svgStyle {
	s := parent

	apply := func(key, val string) {
		key = strings.TrimSpace(key)
		val = strings.TrimSpace(val)
		switch key {
		case "fill":
			if c, ok := parseColor(val); ok {
				s.Fill = svgPaint{Color: c, Set: true}
			}
		case "stroke":
			if c, ok := parseColor(val); ok {
				s.Stroke = svgPaint{Color: c, Set: true}
			}
		case "stroke-width":
			if v, err := strconv.ParseFloat(val, 64); err == nil && v >= 0 {
				s.StrokeWidth = v
			}
		case "opacity":
			if v, err := strconv.ParseFloat(val, 64); err == nil {
				s.Opacity = v
			}
		case "fill-opacity":
			if v, err := strconv.ParseFloat(val, 64); err == nil {
				s.FillOpacity = v
			}
		case "stroke-opacity":
			if v, err := strconv.ParseFloat(val, 64); err == nil {
				s.StrokeOpacity = v
			}
		}
	}

	// Presentation attributes (lower specificity)
	for k, v := range attrs {
		apply(k, v)
	}

	// Inline style attribute (higher specificity)
	if styleStr, ok := attrs["style"]; ok {
		for _, decl := range strings.Split(styleStr, ";") {
			if idx := strings.IndexByte(decl, ':'); idx >= 0 {
				apply(decl[:idx], decl[idx+1:])
			}
		}
	}

	return s
}

// svgNamedColors covers the SVG/CSS named color subset.
var svgNamedColors = map[string]svgColor{
	"black":       {0, 0, 0, false},
	"white":       {1, 1, 1, false},
	"red":         {1, 0, 0, false},
	"green":       {0, 0.50196, 0, false},
	"blue":        {0, 0, 1, false},
	"yellow":      {1, 1, 0, false},
	"cyan":        {0, 1, 1, false},
	"aqua":        {0, 1, 1, false},
	"magenta":     {1, 0, 1, false},
	"fuchsia":     {1, 0, 1, false},
	"orange":      {1, 0.64706, 0, false},
	"purple":      {0.50196, 0, 0.50196, false},
	"pink":        {1, 0.75294, 0.79608, false},
	"brown":       {0.64706, 0.16471, 0.16471, false},
	"gray":        {0.50196, 0.50196, 0.50196, false},
	"grey":        {0.50196, 0.50196, 0.50196, false},
	"silver":      {0.75294, 0.75294, 0.75294, false},
	"lime":        {0, 1, 0, false},
	"navy":        {0, 0, 0.50196, false},
	"teal":        {0, 0.50196, 0.50196, false},
	"maroon":      {0.50196, 0, 0, false},
	"olive":       {0.50196, 0.50196, 0, false},
	"coral":       {1, 0.49804, 0.31373, false},
	"salmon":      {0.98039, 0.50196, 0.44706, false},
	"gold":        {1, 0.84314, 0, false},
	"khaki":       {0.94118, 0.90196, 0.54902, false},
	"violet":      {0.93333, 0.50980, 0.93333, false},
	"indigo":      {0.29412, 0, 0.50980, false},
	"crimson":     {0.86275, 0.07843, 0.23529, false},
	"darkblue":    {0, 0, 0.54510, false},
	"darkgreen":   {0, 0.39216, 0, false},
	"darkred":     {0.54510, 0, 0, false},
	"darkgray":    {0.66275, 0.66275, 0.66275, false},
	"darkgrey":    {0.66275, 0.66275, 0.66275, false},
	"lightblue":   {0.67843, 0.84706, 0.90196, false},
	"lightgreen":  {0.56471, 0.93333, 0.56471, false},
	"lightgray":   {0.82745, 0.82745, 0.82745, false},
	"lightgrey":   {0.82745, 0.82745, 0.82745, false},
	"transparent": {0, 0, 0, true},
	"none":        {0, 0, 0, true},
}

func parseColor(s string) (svgColor, bool) {
	s = strings.TrimSpace(strings.ToLower(s))
	if s == "" {
		return svgColor{}, false
	}
	if c, ok := svgNamedColors[s]; ok {
		return c, true
	}
	if strings.HasPrefix(s, "#") {
		hex := s[1:]
		switch len(hex) {
		case 6:
			r, e1 := strconv.ParseInt(hex[0:2], 16, 32)
			g, e2 := strconv.ParseInt(hex[2:4], 16, 32)
			b, e3 := strconv.ParseInt(hex[4:6], 16, 32)
			if e1 == nil && e2 == nil && e3 == nil {
				return svgColor{float64(r) / 255, float64(g) / 255, float64(b) / 255, false}, true
			}
		case 3:
			r, e1 := strconv.ParseInt(string([]byte{hex[0], hex[0]}), 16, 32)
			g, e2 := strconv.ParseInt(string([]byte{hex[1], hex[1]}), 16, 32)
			b, e3 := strconv.ParseInt(string([]byte{hex[2], hex[2]}), 16, 32)
			if e1 == nil && e2 == nil && e3 == nil {
				return svgColor{float64(r) / 255, float64(g) / 255, float64(b) / 255, false}, true
			}
		}
	}
	if strings.HasPrefix(s, "rgb(") && strings.HasSuffix(s, ")") {
		inner := s[4 : len(s)-1]
		parts := strings.Split(inner, ",")
		if len(parts) == 3 {
			r := parseColorComponent(strings.TrimSpace(parts[0]))
			g := parseColorComponent(strings.TrimSpace(parts[1]))
			b := parseColorComponent(strings.TrimSpace(parts[2]))
			return svgColor{r, g, b, false}, true
		}
	}
	if strings.HasPrefix(s, "rgba(") && strings.HasSuffix(s, ")") {
		inner := s[5 : len(s)-1]
		parts := strings.Split(inner, ",")
		if len(parts) >= 3 {
			r := parseColorComponent(strings.TrimSpace(parts[0]))
			g := parseColorComponent(strings.TrimSpace(parts[1]))
			b := parseColorComponent(strings.TrimSpace(parts[2]))
			return svgColor{r, g, b, false}, true
		}
	}
	return svgColor{}, false
}

func parseColorComponent(s string) float64 {
	if strings.HasSuffix(s, "%") {
		v, _ := strconv.ParseFloat(s[:len(s)-1], 64)
		return math.Max(0, math.Min(1, v/100))
	}
	v, _ := strconv.ParseFloat(s, 64)
	return math.Max(0, math.Min(1, v/255))
}

// ============================================================
// Transforms
// ============================================================

// matrix6 is a 2D affine transformation in PDF convention: [a b c d e f].
// Applied to point (x,y): x' = a*x + c*y + e, y' = b*x + d*y + f.
type matrix6 [6]float64

func identMatrix() matrix6 { return matrix6{1, 0, 0, 1, 0, 0} }

// multiplyMatrix computes a * b (apply a first, then b).
func multiplyMatrix(a, b matrix6) matrix6 {
	return matrix6{
		a[0]*b[0] + a[2]*b[1],
		a[1]*b[0] + a[3]*b[1],
		a[0]*b[2] + a[2]*b[3],
		a[1]*b[2] + a[3]*b[3],
		a[0]*b[4] + a[2]*b[5] + a[4],
		a[1]*b[4] + a[3]*b[5] + a[5],
	}
}

// parseTransform parses an SVG transform attribute string into a combined matrix.
func parseTransform(s string) matrix6 {
	result := identMatrix()
	for s != "" {
		paren := strings.IndexByte(s, '(')
		if paren < 0 {
			break
		}
		close := strings.IndexByte(s[paren:], ')')
		if close < 0 {
			break
		}
		fnName := strings.ToLower(strings.TrimSpace(s[:paren]))
		args := parseNumberList(s[paren+1 : paren+close])
		s = strings.TrimSpace(s[paren+close+1:])

		var m matrix6
		switch fnName {
		case "translate":
			tx, ty := 0.0, 0.0
			if len(args) >= 1 {
				tx = args[0]
			}
			if len(args) >= 2 {
				ty = args[1]
			}
			m = matrix6{1, 0, 0, 1, tx, ty}
		case "scale":
			sx, sy := 1.0, 1.0
			if len(args) >= 1 {
				sx = args[0]
				sy = sx
			}
			if len(args) >= 2 {
				sy = args[1]
			}
			m = matrix6{sx, 0, 0, sy, 0, 0}
		case "rotate":
			angle := 0.0
			if len(args) >= 1 {
				angle = args[0]
			}
			a := angle * math.Pi / 180
			cosA, sinA := math.Cos(a), math.Sin(a)
			if len(args) == 3 {
				cx, cy := args[1], args[2]
				m = matrix6{cosA, sinA, -sinA, cosA,
					cx - cx*cosA + cy*sinA,
					cy - cx*sinA - cy*cosA}
			} else {
				m = matrix6{cosA, sinA, -sinA, cosA, 0, 0}
			}
		case "skewx":
			if len(args) >= 1 {
				m = matrix6{1, 0, math.Tan(args[0] * math.Pi / 180), 1, 0, 0}
			} else {
				m = identMatrix()
			}
		case "skewy":
			if len(args) >= 1 {
				m = matrix6{1, math.Tan(args[0] * math.Pi / 180), 0, 1, 0, 0}
			} else {
				m = identMatrix()
			}
		case "matrix":
			if len(args) == 6 {
				m = matrix6{args[0], args[1], args[2], args[3], args[4], args[5]}
			} else {
				m = identMatrix()
			}
		default:
			m = identMatrix()
		}
		result = multiplyMatrix(result, m)
	}
	return result
}

func parseNumberList(s string) []float64 {
	s = strings.ReplaceAll(s, ",", " ")
	parts := strings.Fields(s)
	nums := make([]float64, 0, len(parts))
	for _, p := range parts {
		if v, err := strconv.ParseFloat(p, 64); err == nil {
			nums = append(nums, v)
		}
	}
	return nums
}

// ============================================================
// SVG length parsing
// ============================================================

func parseSVGLength(s string) float64 {
	s = strings.TrimSpace(s)
	units := []struct {
		suffix string
		factor float64
	}{
		{"px", 1},
		{"pt", 4.0 / 3}, // 1pt = 4/3 px at 96dpi
		{"mm", 96.0 / 25.4},
		{"cm", 96.0 / 2.54},
		{"in", 96},
		{"rem", 16},
		{"em", 16},
	}
	for _, u := range units {
		if strings.HasSuffix(s, u.suffix) {
			v, err := strconv.ParseFloat(strings.TrimSuffix(s, u.suffix), 64)
			if err == nil {
				return v * u.factor
			}
		}
	}
	v, _ := strconv.ParseFloat(s, 64)
	return v
}

// ============================================================
// PDF content builder
// ============================================================

type svgBuilder struct {
	buf     strings.Builder
	gsMap   map[[2]float64]string // [fillAlpha, strokeAlpha] -> gs name
	gsCount int
}

// buildResources returns a pdf.Dict for the ExtGState resources used in the form.
func (b *svgBuilder) buildResources() pdf.Dict {
	if len(b.gsMap) == 0 {
		return nil
	}
	gsDict := make(pdf.Dict, len(b.gsMap))
	for key, name := range b.gsMap {
		entry := pdf.Dict{}
		if key[0] < 1 {
			entry[pdf.Name("ca")] = pdf.Real(key[0])
		}
		if key[1] < 1 {
			entry[pdf.Name("CA")] = pdf.Real(key[1])
		}
		gsDict[pdf.Name(name)] = entry
	}
	return pdf.Dict{pdf.Name("ExtGState"): gsDict}
}

func (b *svgBuilder) renderElem(elem *svgElem, parent svgStyle, viewW, viewH float64) {
	// Skip non-rendering elements.
	switch elem.name {
	case "defs", "title", "desc", "metadata", "style",
		"lineargradient", "radialgradient", "pattern",
		"filter", "mask", "symbol", "clippath":
		return
	}

	style := parseElemStyle(elem.attrs, parent)

	var hasTransform bool
	if t, ok := elem.attrs["transform"]; ok && t != "" {
		m := parseTransform(t)
		b.buf.WriteString("q\n")
		fmt.Fprintf(&b.buf, "%g %g %g %g %g %g cm\n", m[0], m[1], m[2], m[3], m[4], m[5])
		hasTransform = true
	}

	switch elem.name {
	case "g", "a":
		for _, child := range elem.children {
			b.renderElem(child, style, viewW, viewH)
		}
	case "use":
		// <use> references are complex; skip for now.
	case "path":
		b.renderPath(elem, style)
	case "rect":
		b.renderRect(elem, style)
	case "circle":
		b.renderCircle(elem, style)
	case "ellipse":
		b.renderEllipse(elem, style)
	case "line":
		b.renderLine(elem, style)
	case "polyline":
		b.renderPolyPoints(elem, style, false)
	case "polygon":
		b.renderPolyPoints(elem, style, true)
	}

	if hasTransform {
		b.buf.WriteString("Q\n")
	}
}

// emitFillStroke emits color, ExtGState (if needed), and the paint operator.
func (b *svgBuilder) emitFillStroke(s svgStyle) {
	fillNone := !s.Fill.Set || s.Fill.Color.None
	strokeNone := !s.Stroke.Set || s.Stroke.Color.None

	fillAlpha := s.Opacity * s.FillOpacity
	strokeAlpha := s.Opacity * s.StrokeOpacity

	// Emit ExtGState for opacity < 1.
	if fillAlpha < 1 || strokeAlpha < 1 {
		key := [2]float64{fillAlpha, strokeAlpha}
		if _, ok := b.gsMap[key]; !ok {
			b.gsCount++
			b.gsMap[key] = fmt.Sprintf("gs%d", b.gsCount)
		}
		fmt.Fprintf(&b.buf, "/%s gs\n", b.gsMap[key])
	}

	if !fillNone {
		c := s.Fill.Color
		fmt.Fprintf(&b.buf, "%g %g %g rg\n", c.R, c.G, c.B)
	}
	if !strokeNone {
		c := s.Stroke.Color
		fmt.Fprintf(&b.buf, "%g %g %g RG\n", c.R, c.G, c.B)
		sw := s.StrokeWidth
		if sw <= 0 {
			sw = 1
		}
		fmt.Fprintf(&b.buf, "%g w\n", sw)
	}

	switch {
	case !fillNone && !strokeNone:
		b.buf.WriteString("B\n")
	case !fillNone:
		b.buf.WriteString("f\n")
	case !strokeNone:
		b.buf.WriteString("S\n")
	default:
		b.buf.WriteString("n\n")
	}
}

func (b *svgBuilder) renderPath(elem *svgElem, style svgStyle) {
	d := elem.attrs["d"]
	if d == "" {
		return
	}
	b.buf.WriteString("q\n")
	convertPathData(d, &b.buf)
	b.emitFillStroke(style)
	b.buf.WriteString("Q\n")
}

func (b *svgBuilder) renderRect(elem *svgElem, style svgStyle) {
	attrs := elem.attrs
	x := attrFloat(attrs, "x", 0)
	y := attrFloat(attrs, "y", 0)
	w := attrFloat(attrs, "width", 0)
	h := attrFloat(attrs, "height", 0)
	if w <= 0 || h <= 0 {
		return
	}
	rx := attrFloat(attrs, "rx", -1)
	ry := attrFloat(attrs, "ry", -1)

	// Resolve rx/ry defaults per SVG spec.
	if rx < 0 && ry < 0 {
		rx, ry = 0, 0
	} else if rx < 0 {
		rx = ry
	} else if ry < 0 {
		ry = rx
	}
	rx = math.Min(rx, w/2)
	ry = math.Min(ry, h/2)

	b.buf.WriteString("q\n")
	if rx == 0 && ry == 0 {
		fmt.Fprintf(&b.buf, "%g %g %g %g re\n", x, y, w, h)
	} else {
		// Rounded rectangle via cubic Bézier curves.
		// κ ≈ 0.5523 gives best circular approximation.
		const kappa = 0.5523
		// Start at top-left corner, after the left arc.
		fmt.Fprintf(&b.buf, "%g %g m\n", x+rx, y)
		fmt.Fprintf(&b.buf, "%g %g l\n", x+w-rx, y)
		fmt.Fprintf(&b.buf, "%g %g %g %g %g %g c\n",
			x+w-rx+kappa*rx, y, x+w, y+kappa*ry, x+w, y+ry)
		fmt.Fprintf(&b.buf, "%g %g l\n", x+w, y+h-ry)
		fmt.Fprintf(&b.buf, "%g %g %g %g %g %g c\n",
			x+w, y+h-ry+kappa*ry, x+w-rx+kappa*rx, y+h, x+w-rx, y+h)
		fmt.Fprintf(&b.buf, "%g %g l\n", x+rx, y+h)
		fmt.Fprintf(&b.buf, "%g %g %g %g %g %g c\n",
			x+rx-kappa*rx, y+h, x, y+h-ry+kappa*ry, x, y+h-ry)
		fmt.Fprintf(&b.buf, "%g %g l\n", x, y+ry)
		fmt.Fprintf(&b.buf, "%g %g %g %g %g %g c\n",
			x, y+ry-kappa*ry, x+rx-kappa*rx, y, x+rx, y)
		b.buf.WriteString("h\n")
	}
	b.emitFillStroke(style)
	b.buf.WriteString("Q\n")
}

func (b *svgBuilder) renderCircle(elem *svgElem, style svgStyle) {
	cx := attrFloat(elem.attrs, "cx", 0)
	cy := attrFloat(elem.attrs, "cy", 0)
	r := attrFloat(elem.attrs, "r", 0)
	if r <= 0 {
		return
	}
	b.renderEllipseShape(cx, cy, r, r, style)
}

func (b *svgBuilder) renderEllipse(elem *svgElem, style svgStyle) {
	cx := attrFloat(elem.attrs, "cx", 0)
	cy := attrFloat(elem.attrs, "cy", 0)
	rx := attrFloat(elem.attrs, "rx", 0)
	ry := attrFloat(elem.attrs, "ry", 0)
	if rx <= 0 || ry <= 0 {
		return
	}
	b.renderEllipseShape(cx, cy, rx, ry, style)
}

func (b *svgBuilder) renderEllipseShape(cx, cy, rx, ry float64, style svgStyle) {
	const kappa = 0.5523
	b.buf.WriteString("q\n")
	fmt.Fprintf(&b.buf, "%g %g m\n", cx+rx, cy)
	fmt.Fprintf(&b.buf, "%g %g %g %g %g %g c\n",
		cx+rx, cy+kappa*ry, cx+kappa*rx, cy+ry, cx, cy+ry)
	fmt.Fprintf(&b.buf, "%g %g %g %g %g %g c\n",
		cx-kappa*rx, cy+ry, cx-rx, cy+kappa*ry, cx-rx, cy)
	fmt.Fprintf(&b.buf, "%g %g %g %g %g %g c\n",
		cx-rx, cy-kappa*ry, cx-kappa*rx, cy-ry, cx, cy-ry)
	fmt.Fprintf(&b.buf, "%g %g %g %g %g %g c\n",
		cx+kappa*rx, cy-ry, cx+rx, cy-kappa*ry, cx+rx, cy)
	b.buf.WriteString("h\n")
	b.emitFillStroke(style)
	b.buf.WriteString("Q\n")
}

func (b *svgBuilder) renderLine(elem *svgElem, style svgStyle) {
	x1 := attrFloat(elem.attrs, "x1", 0)
	y1 := attrFloat(elem.attrs, "y1", 0)
	x2 := attrFloat(elem.attrs, "x2", 0)
	y2 := attrFloat(elem.attrs, "y2", 0)
	b.buf.WriteString("q\n")
	fmt.Fprintf(&b.buf, "%g %g m\n", x1, y1)
	fmt.Fprintf(&b.buf, "%g %g l\n", x2, y2)
	// Lines have no fill.
	ls := style
	ls.Fill = svgPaint{Color: svgColor{None: true}, Set: true}
	b.emitFillStroke(ls)
	b.buf.WriteString("Q\n")
}

func (b *svgBuilder) renderPolyPoints(elem *svgElem, style svgStyle, close bool) {
	pts := parsePointsList(elem.attrs["points"])
	if len(pts) < 4 {
		return
	}
	b.buf.WriteString("q\n")
	fmt.Fprintf(&b.buf, "%g %g m\n", pts[0], pts[1])
	for i := 2; i+1 < len(pts); i += 2 {
		fmt.Fprintf(&b.buf, "%g %g l\n", pts[i], pts[i+1])
	}
	if close {
		b.buf.WriteString("h\n")
		b.emitFillStroke(style)
	} else {
		ls := style
		ls.Fill = svgPaint{Color: svgColor{None: true}, Set: true}
		b.emitFillStroke(ls)
	}
	b.buf.WriteString("Q\n")
}

func parsePointsList(s string) []float64 {
	s = strings.ReplaceAll(s, ",", " ")
	parts := strings.Fields(s)
	nums := make([]float64, 0, len(parts))
	for _, p := range parts {
		if v, err := strconv.ParseFloat(p, 64); err == nil {
			nums = append(nums, v)
		}
	}
	return nums
}

func attrFloat(attrs map[string]string, key string, def float64) float64 {
	if s, ok := attrs[key]; ok {
		if v, err := strconv.ParseFloat(strings.TrimSpace(s), 64); err == nil {
			return v
		}
	}
	return def
}

// ============================================================
// SVG path data → PDF path operators
// ============================================================

// convertPathData parses an SVG path data string and writes the equivalent
// PDF path construction operators to buf.
func convertPathData(d string, buf *strings.Builder) {
	p := &pathParser{data: d, buf: buf}
	p.parse()
}

type pathParser struct {
	data     string
	pos      int
	buf      *strings.Builder
	curX     float64
	curY     float64
	startX   float64
	startY   float64
	lastCPX  float64 // last (smooth) control point X
	lastCPY  float64 // last (smooth) control point Y
}

func (p *pathParser) parse() {
	var cmd byte
	for {
		p.skipSep()
		if p.pos >= len(p.data) {
			break
		}
		c := p.data[p.pos]
		if isPathLetter(c) {
			cmd = c
			p.pos++
			p.skipSep()
		}
		if cmd == 0 {
			break
		}
		if cmd == 'Z' || cmd == 'z' {
			p.buf.WriteString("h\n")
			p.curX, p.curY = p.startX, p.startY
			p.lastCPX, p.lastCPY = p.curX, p.curY
			cmd = 0
			continue
		}
		if p.pos >= len(p.data) || isPathLetter(p.data[p.pos]) {
			continue
		}
		p.execOne(cmd)
		// After M/m, implicit repetitions use L/l.
		if cmd == 'M' {
			cmd = 'L'
		} else if cmd == 'm' {
			cmd = 'l'
		}
	}
}

func (p *pathParser) execOne(cmd byte) {
	switch cmd {
	case 'M':
		x, y := p.readXY()
		p.curX, p.curY = x, y
		p.startX, p.startY = x, y
		p.lastCPX, p.lastCPY = x, y
		fmt.Fprintf(p.buf, "%g %g m\n", x, y)
	case 'm':
		x, y := p.readXY()
		p.curX += x
		p.curY += y
		p.startX, p.startY = p.curX, p.curY
		p.lastCPX, p.lastCPY = p.curX, p.curY
		fmt.Fprintf(p.buf, "%g %g m\n", p.curX, p.curY)
	case 'L':
		x, y := p.readXY()
		p.curX, p.curY = x, y
		p.lastCPX, p.lastCPY = x, y
		fmt.Fprintf(p.buf, "%g %g l\n", x, y)
	case 'l':
		x, y := p.readXY()
		p.curX += x
		p.curY += y
		p.lastCPX, p.lastCPY = p.curX, p.curY
		fmt.Fprintf(p.buf, "%g %g l\n", p.curX, p.curY)
	case 'H':
		x := p.readFloat()
		p.curX = x
		p.lastCPX = x
		fmt.Fprintf(p.buf, "%g %g l\n", x, p.curY)
	case 'h':
		x := p.readFloat()
		p.curX += x
		p.lastCPX = p.curX
		fmt.Fprintf(p.buf, "%g %g l\n", p.curX, p.curY)
	case 'V':
		y := p.readFloat()
		p.curY = y
		p.lastCPY = y
		fmt.Fprintf(p.buf, "%g %g l\n", p.curX, y)
	case 'v':
		y := p.readFloat()
		p.curY += y
		p.lastCPY = p.curY
		fmt.Fprintf(p.buf, "%g %g l\n", p.curX, p.curY)
	case 'C':
		x1, y1 := p.readXY()
		x2, y2 := p.readXY()
		x, y := p.readXY()
		fmt.Fprintf(p.buf, "%g %g %g %g %g %g c\n", x1, y1, x2, y2, x, y)
		p.lastCPX, p.lastCPY = x2, y2
		p.curX, p.curY = x, y
	case 'c':
		x1, y1 := p.readXY()
		x2, y2 := p.readXY()
		dx, dy := p.readXY()
		ax1, ay1 := p.curX+x1, p.curY+y1
		ax2, ay2 := p.curX+x2, p.curY+y2
		ax, ay := p.curX+dx, p.curY+dy
		fmt.Fprintf(p.buf, "%g %g %g %g %g %g c\n", ax1, ay1, ax2, ay2, ax, ay)
		p.lastCPX, p.lastCPY = ax2, ay2
		p.curX, p.curY = ax, ay
	case 'S':
		x2, y2 := p.readXY()
		x, y := p.readXY()
		x1 := 2*p.curX - p.lastCPX
		y1 := 2*p.curY - p.lastCPY
		fmt.Fprintf(p.buf, "%g %g %g %g %g %g c\n", x1, y1, x2, y2, x, y)
		p.lastCPX, p.lastCPY = x2, y2
		p.curX, p.curY = x, y
	case 's':
		x2, y2 := p.readXY()
		dx, dy := p.readXY()
		ax2, ay2 := p.curX+x2, p.curY+y2
		ax, ay := p.curX+dx, p.curY+dy
		x1 := 2*p.curX - p.lastCPX
		y1 := 2*p.curY - p.lastCPY
		fmt.Fprintf(p.buf, "%g %g %g %g %g %g c\n", x1, y1, ax2, ay2, ax, ay)
		p.lastCPX, p.lastCPY = ax2, ay2
		p.curX, p.curY = ax, ay
	case 'Q':
		qx1, qy1 := p.readXY()
		x, y := p.readXY()
		cp1x := p.curX + 2.0/3*(qx1-p.curX)
		cp1y := p.curY + 2.0/3*(qy1-p.curY)
		cp2x := x + 2.0/3*(qx1-x)
		cp2y := y + 2.0/3*(qy1-y)
		fmt.Fprintf(p.buf, "%g %g %g %g %g %g c\n", cp1x, cp1y, cp2x, cp2y, x, y)
		p.lastCPX, p.lastCPY = qx1, qy1
		p.curX, p.curY = x, y
	case 'q':
		qx1, qy1 := p.readXY()
		dx, dy := p.readXY()
		aqx1, aqy1 := p.curX+qx1, p.curY+qy1
		ax, ay := p.curX+dx, p.curY+dy
		cp1x := p.curX + 2.0/3*(aqx1-p.curX)
		cp1y := p.curY + 2.0/3*(aqy1-p.curY)
		cp2x := ax + 2.0/3*(aqx1-ax)
		cp2y := ay + 2.0/3*(aqy1-ay)
		fmt.Fprintf(p.buf, "%g %g %g %g %g %g c\n", cp1x, cp1y, cp2x, cp2y, ax, ay)
		p.lastCPX, p.lastCPY = aqx1, aqy1
		p.curX, p.curY = ax, ay
	case 'T':
		x, y := p.readXY()
		qx1 := 2*p.curX - p.lastCPX
		qy1 := 2*p.curY - p.lastCPY
		cp1x := p.curX + 2.0/3*(qx1-p.curX)
		cp1y := p.curY + 2.0/3*(qy1-p.curY)
		cp2x := x + 2.0/3*(qx1-x)
		cp2y := y + 2.0/3*(qy1-y)
		fmt.Fprintf(p.buf, "%g %g %g %g %g %g c\n", cp1x, cp1y, cp2x, cp2y, x, y)
		p.lastCPX, p.lastCPY = qx1, qy1
		p.curX, p.curY = x, y
	case 't':
		dx, dy := p.readXY()
		ax, ay := p.curX+dx, p.curY+dy
		qx1 := 2*p.curX - p.lastCPX
		qy1 := 2*p.curY - p.lastCPY
		cp1x := p.curX + 2.0/3*(qx1-p.curX)
		cp1y := p.curY + 2.0/3*(qy1-p.curY)
		cp2x := ax + 2.0/3*(qx1-ax)
		cp2y := ay + 2.0/3*(qy1-ay)
		fmt.Fprintf(p.buf, "%g %g %g %g %g %g c\n", cp1x, cp1y, cp2x, cp2y, ax, ay)
		p.lastCPX, p.lastCPY = qx1, qy1
		p.curX, p.curY = ax, ay
	case 'A', 'a':
		rx := math.Abs(p.readFloat())
		ry := math.Abs(p.readFloat())
		xRot := p.readFloat()
		largeArc := p.readFloat() != 0
		sweep := p.readFloat() != 0
		x, y := p.readXY()
		if cmd == 'a' {
			x += p.curX
			y += p.curY
		}
		arcToBezier(p.curX, p.curY, rx, ry, xRot, largeArc, sweep, x, y, p.buf)
		p.lastCPX, p.lastCPY = x, y
		p.curX, p.curY = x, y
	}
}

func (p *pathParser) skipSep() {
	for p.pos < len(p.data) {
		c := p.data[p.pos]
		if c == ' ' || c == '\t' || c == '\n' || c == '\r' || c == ',' {
			p.pos++
		} else {
			break
		}
	}
}

func (p *pathParser) readFloat() float64 {
	p.skipSep()
	if p.pos >= len(p.data) {
		return 0
	}
	start := p.pos
	c := p.data[p.pos]
	if c == '+' || c == '-' {
		p.pos++
	}
	for p.pos < len(p.data) {
		c = p.data[p.pos]
		if (c >= '0' && c <= '9') || c == '.' {
			p.pos++
		} else {
			break
		}
	}
	// Scientific notation
	if p.pos < len(p.data) && (p.data[p.pos] == 'e' || p.data[p.pos] == 'E') {
		p.pos++
		if p.pos < len(p.data) && (p.data[p.pos] == '+' || p.data[p.pos] == '-') {
			p.pos++
		}
		for p.pos < len(p.data) && p.data[p.pos] >= '0' && p.data[p.pos] <= '9' {
			p.pos++
		}
	}
	v, _ := strconv.ParseFloat(p.data[start:p.pos], 64)
	return v
}

func (p *pathParser) readXY() (float64, float64) {
	x := p.readFloat()
	y := p.readFloat()
	return x, y
}

func isPathLetter(c byte) bool {
	return (c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z')
}

// ============================================================
// Arc to cubic Bézier conversion
// ============================================================

// arcToBezier converts an SVG elliptical arc to cubic Bézier curves.
// Based on the SVG specification's endpoint-to-center parameterization.
func arcToBezier(x1, y1, rx, ry, phi float64, largeArc, sweep bool, x2, y2 float64, buf *strings.Builder) {
	if rx == 0 || ry == 0 {
		fmt.Fprintf(buf, "%g %g l\n", x2, y2)
		return
	}
	if x1 == x2 && y1 == y2 {
		return
	}

	phiRad := phi * math.Pi / 180
	cosPhi := math.Cos(phiRad)
	sinPhi := math.Sin(phiRad)

	// Step 1: midpoint.
	dx := (x1 - x2) / 2
	dy := (y1 - y2) / 2
	x1p := cosPhi*dx + sinPhi*dy
	y1p := -sinPhi*dx + cosPhi*dy

	// Ensure radii are large enough.
	rx2, ry2 := rx*rx, ry*ry
	x1p2, y1p2 := x1p*x1p, y1p*y1p
	if lambda := x1p2/rx2 + y1p2/ry2; lambda > 1 {
		sqrtL := math.Sqrt(lambda)
		rx *= sqrtL
		ry *= sqrtL
		rx2 = rx * rx
		ry2 = ry * ry
	}

	// Step 2: center in rotated frame.
	num := rx2*ry2 - rx2*y1p2 - ry2*x1p2
	den := rx2*y1p2 + ry2*x1p2
	var sq float64
	if den > 0 {
		sq = math.Sqrt(math.Max(0, num/den))
	}
	if largeArc == sweep {
		sq = -sq
	}
	cxp := sq * rx * y1p / ry
	cyp := -sq * ry * x1p / rx

	// Step 3: actual center.
	cx := cosPhi*cxp - sinPhi*cyp + (x1+x2)/2
	cy := sinPhi*cxp + cosPhi*cyp + (y1+y2)/2

	// Step 4: angles.
	ux := (x1p - cxp) / rx
	uy := (y1p - cyp) / ry
	vx := (-x1p - cxp) / rx
	vy := (-y1p - cyp) / ry

	theta1 := svgVecAngle(1, 0, ux, uy)
	dTheta := svgVecAngle(ux, uy, vx, vy)
	if !sweep && dTheta > 0 {
		dTheta -= 2 * math.Pi
	} else if sweep && dTheta < 0 {
		dTheta += 2 * math.Pi
	}

	// Split into ≤90° segments.
	nSegs := int(math.Ceil(math.Abs(dTheta) / (math.Pi / 2)))
	if nSegs < 1 {
		nSegs = 1
	}
	dSeg := dTheta / float64(nSegs)
	for i := 0; i < nSegs; i++ {
		t1 := theta1 + float64(i)*dSeg
		t2 := t1 + dSeg
		arcSegToBezier(cx, cy, rx, ry, phiRad, t1, t2, buf)
	}
}

func arcSegToBezier(cx, cy, rx, ry, phi, t1, t2 float64, buf *strings.Builder) {
	halfDt := (t2 - t1) / 2
	alpha := math.Sin(t2-t1) * (math.Sqrt(4+3*math.Tan(halfDt)*math.Tan(halfDt)) - 1) / 3

	cosPhi, sinPhi := math.Cos(phi), math.Sin(phi)
	cosT1, sinT1 := math.Cos(t1), math.Sin(t1)
	cosT2, sinT2 := math.Cos(t2), math.Sin(t2)

	p1x := cx + cosPhi*rx*cosT1 - sinPhi*ry*sinT1
	p1y := cy + sinPhi*rx*cosT1 + cosPhi*ry*sinT1
	p2x := cx + cosPhi*rx*cosT2 - sinPhi*ry*sinT2
	p2y := cy + sinPhi*rx*cosT2 + cosPhi*ry*sinT2

	dx1 := -cosPhi*rx*sinT1 - sinPhi*ry*cosT1
	dy1 := -sinPhi*rx*sinT1 + cosPhi*ry*cosT1
	dx2 := -cosPhi*rx*sinT2 - sinPhi*ry*cosT2
	dy2 := -sinPhi*rx*sinT2 + cosPhi*ry*cosT2

	cp1x := p1x + alpha*dx1
	cp1y := p1y + alpha*dy1
	cp2x := p2x - alpha*dx2
	cp2y := p2y - alpha*dy2

	fmt.Fprintf(buf, "%g %g %g %g %g %g c\n", cp1x, cp1y, cp2x, cp2y, p2x, p2y)
}

func svgVecAngle(ux, uy, vx, vy float64) float64 {
	dot := ux*vx + uy*vy
	lenU := math.Sqrt(ux*ux + uy*uy)
	lenV := math.Sqrt(vx*vx + vy*vy)
	if lenU == 0 || lenV == 0 {
		return 0
	}
	cosA := math.Max(-1, math.Min(1, dot/(lenU*lenV)))
	angle := math.Acos(cosA)
	if ux*vy-uy*vx < 0 {
		angle = -angle
	}
	return angle
}
