package svg

import (
	"fmt"
	"strconv"
	"strings"
)

// Alpha is a pair of constant alpha values for an ExtGState resource:
// Fill is the non-stroking alpha (/ca) and Stroke the stroking alpha (/CA).
type Alpha struct {
	Fill, Stroke float64
}

// Rendered is the PDF form of an SVG image.
type Rendered struct {
	// Content holds PDF content-stream operators that draw the image into
	// the unit square (0,0)–(1,1), the same coordinate space as an image
	// XObject, so callers place it with "w 0 0 h x y cm".
	Content []byte
	// ExtGStates maps the ExtGState resource names used by Content to the
	// alpha values they must set. Callers must provide them in the
	// resources of the form or page that draws Content.
	ExtGStates map[string]Alpha
}

// skippedElements are never painted directly.
var skippedElements = map[string]bool{
	"defs": true, "symbol": true, "clipPath": true, "mask": true,
	"linearGradient": true, "radialGradient": true, "pattern": true,
	"title": true, "desc": true, "metadata": true, "style": true,
	"script": true, "text": true, "image": true, "filter": true,
	"marker": true, "foreignObject": true,
}

// state is the inherited painting state.
type state struct {
	fill, stroke   string // raw paint values
	color          rgb
	fillOpacity    float64
	strokeOpacity  float64
	opacity        float64 // product of ancestor group opacities
	strokeWidth    float64
	fillRule       string
	lineCap        string
	lineJoin       string
	miterLimit     float64
	dash           []float64
	dashOffset     float64
	visible        bool
	viewportW      float64 // for percentage lengths
	viewportH      float64
	strokeWidthRef float64
}

type renderer struct {
	img    *Image
	buf    strings.Builder
	gsName map[Alpha]string
	gs     map[string]Alpha
}

// Render converts the image to PDF content-stream operators.
func (img *Image) Render() Rendered {
	r := &renderer{img: img, gsName: make(map[Alpha]string), gs: make(map[string]Alpha)}
	vb := img.viewBox
	st := state{
		fill:          "black",
		stroke:        "none",
		color:         rgb{0, 0, 0, 1},
		fillOpacity:   1,
		strokeOpacity: 1,
		opacity:       1,
		strokeWidth:   1,
		fillRule:      "nonzero",
		lineCap:       "butt",
		lineJoin:      "miter",
		miterLimit:    4,
		visible:       true,
		viewportW:     vb[2],
		viewportH:     vb[3],
	}
	st.strokeWidthRef = vb[2]

	r.buf.WriteString("q\n0 0 1 1 re W n\n")
	r.writeMatrix(img.unitMatrix())
	r.renderNode(img.root, st, 0, true)
	r.buf.WriteString("Q\n")
	return Rendered{Content: []byte(r.buf.String()), ExtGStates: r.gs}
}

// unitMatrix maps viewBox coordinates (y down) into the unit square
// (y up), honoring preserveAspectRatio (default xMidYMid meet, or none).
func (img *Image) unitMatrix() matrix {
	vb := img.viewBox
	w, h := img.Width, img.Height
	sx, sy := w/vb[2], h/vb[3]
	var tx, ty float64
	if !img.noAR {
		s := min(sx, sy)
		sx, sy = s, s
		tx = (w - vb[2]*s) / 2
		ty = (h - vb[3]*s) / 2
	}
	// viewport (y down, size w×h) -> unit square (y up).
	return matrix{
		sx / w, 0,
		0, -sy / h,
		(tx - vb[0]*sx) / w, 1 - (ty-vb[1]*sy)/h,
	}
}

func (r *renderer) writeMatrix(m matrix) {
	fmt.Fprintf(&r.buf, "%s %s %s %s %s %s cm\n",
		fmtNum(m[0], 6), fmtNum(m[1], 6), fmtNum(m[2], 6),
		fmtNum(m[3], 6), fmtNum(m[4], 6), fmtNum(m[5], 6))
}

// renderNode paints n and its descendants. root marks the outermost <svg>,
// whose size and viewBox are already applied.
func (r *renderer) renderNode(n *node, parent state, depth int, root bool) {
	if depth > maxDepth || skippedElements[n.name] {
		return
	}
	st, display := r.img.applyProps(n, parent)
	if !display {
		return
	}

	t, hasT := n.attrs["transform"]
	if hasT {
		r.buf.WriteString("q\n")
		r.writeMatrix(parseTransform(t))
		defer r.buf.WriteString("Q\n")
	}

	switch n.name {
	case "svg":
		if !root {
			r.renderTranslated(n, st, depth)
			return
		}
		r.renderChildren(n.children, st, depth)
	case "g", "a":
		r.renderChildren(n.children, st, depth)
	case "switch":
		if len(n.children) > 0 {
			r.renderNode(n.children[0], st, depth+1, false)
		}
	case "use":
		r.renderUse(n, st, depth)
	default:
		r.drawShape(n, st)
	}
}

func (r *renderer) renderChildren(children []*node, st state, depth int) {
	for _, c := range children {
		r.renderNode(c, st, depth+1, false)
	}
}

// renderTranslated renders a nested <svg> offset by its x/y attributes.
func (r *renderer) renderTranslated(n *node, st state, depth int) {
	x := lengthAttr(n, "x", st.viewportW, 0)
	y := lengthAttr(n, "y", st.viewportH, 0)
	r.buf.WriteString("q\n")
	r.writeMatrix(matrix{1, 0, 0, 1, x, y})
	r.renderChildren(n.children, st, depth)
	r.buf.WriteString("Q\n")
}

func (r *renderer) renderUse(n *node, st state, depth int) {
	href := n.attrs["href"]
	if !strings.HasPrefix(href, "#") {
		return
	}
	ref, ok := r.img.ids[href[1:]]
	if !ok || ref == n {
		return
	}
	x := lengthAttr(n, "x", st.viewportW, 0)
	y := lengthAttr(n, "y", st.viewportH, 0)
	r.buf.WriteString("q\n")
	if x != 0 || y != 0 {
		r.writeMatrix(matrix{1, 0, 0, 1, x, y})
	}
	if ref.name == "symbol" {
		// A referenced symbol renders like a group.
		if sst, display := r.img.applyProps(ref, st); display {
			r.renderChildren(ref.children, sst, depth+1)
		}
	} else {
		r.renderNode(ref, st, depth+1, false)
	}
	r.buf.WriteString("Q\n")
}

// applyProps returns the state for n after applying its declared
// properties, and false when n is not displayed.
func (img *Image) applyProps(n *node, parent state) (state, bool) {
	st := parent
	st.dash = parent.dash
	props := img.declaredProps(n)
	if strings.TrimSpace(props["display"]) == "none" {
		return st, false
	}
	for k, v := range props {
		st.apply(k, v)
	}
	return st, true
}

func (st *state) apply(prop, v string) {
	if v == "inherit" {
		return
	}
	if strings.HasPrefix(prop, "stroke") {
		st.applyStroke(prop, v)
		return
	}
	switch prop {
	case "fill":
		st.fill = v
	case "color":
		if c, ok := parseColor(v, st.color); ok {
			st.color = c
		}
	case "fill-opacity":
		st.fillOpacity = parseOpacity(v, st.fillOpacity)
	case "opacity":
		st.opacity *= parseOpacity(v, 1)
	case "fill-rule":
		st.fillRule = v
	case "visibility":
		st.visible = v != "hidden" && v != "collapse"
	}
}

func (st *state) applyStroke(prop, v string) {
	switch prop {
	case "stroke":
		st.stroke = v
	case "stroke-opacity":
		st.strokeOpacity = parseOpacity(v, st.strokeOpacity)
	case "stroke-width":
		if w, ok := parseLength(v, st.strokeWidthRef); ok && w >= 0 {
			st.strokeWidth = w
		}
	case "stroke-linecap":
		st.lineCap = v
	case "stroke-linejoin":
		st.lineJoin = v
	case "stroke-miterlimit":
		if m, err := strconv.ParseFloat(v, 64); err == nil && m >= 1 {
			st.miterLimit = m
		}
	case "stroke-dasharray":
		st.dash = parseDashArray(v)
	case "stroke-dashoffset":
		if o, ok := parseLength(v, st.strokeWidthRef); ok {
			st.dashOffset = o
		}
	}
}

func parseOpacity(v string, def float64) float64 {
	v = strings.TrimSpace(v)
	pct := strings.HasSuffix(v, "%")
	f, err := strconv.ParseFloat(strings.TrimSuffix(v, "%"), 64)
	if err != nil {
		return def
	}
	if pct {
		f /= 100
	}
	return clamp01(f)
}

func parseDashArray(v string) []float64 {
	if strings.TrimSpace(v) == "none" {
		return nil
	}
	nums := parseNumberList(v)
	sum := 0.0
	for _, n := range nums {
		if n < 0 {
			return nil
		}
		sum += n
	}
	if sum == 0 {
		return nil
	}
	if len(nums)%2 == 1 {
		nums = append(nums, nums...)
	}
	return nums
}

// shapeSegments returns the geometry of a basic shape element.
func shapeSegments(n *node, st state) (segs []segment, fillable bool) {
	w, h := st.viewportW, st.viewportH
	switch n.name {
	case "path":
		return parsePathData(n.attrs["d"]), true
	case "rect":
		rx, hasRX := parseLength(n.attrs["rx"], w)
		ry, hasRY := parseLength(n.attrs["ry"], h)
		if !hasRX {
			rx = ry
		}
		if !hasRY {
			ry = rx
		}
		return rectSegments(lengthAttr(n, "x", w, 0), lengthAttr(n, "y", h, 0),
			lengthAttr(n, "width", w, 0), lengthAttr(n, "height", h, 0), rx, ry), true
	case "circle":
		r := lengthAttr(n, "r", w, 0)
		return ellipseSegments(lengthAttr(n, "cx", w, 0), lengthAttr(n, "cy", h, 0), r, r), true
	case "ellipse":
		return ellipseSegments(lengthAttr(n, "cx", w, 0), lengthAttr(n, "cy", h, 0),
			lengthAttr(n, "rx", w, 0), lengthAttr(n, "ry", h, 0)), true
	case "line":
		b := &pathBuilder{}
		b.moveTo(lengthAttr(n, "x1", w, 0), lengthAttr(n, "y1", h, 0))
		b.lineTo(lengthAttr(n, "x2", w, 0), lengthAttr(n, "y2", h, 0))
		return b.segs, false
	case "polyline":
		return polySegments(n.attrs["points"], false), true
	case "polygon":
		return polySegments(n.attrs["points"], true), true
	}
	return nil, false
}

// shapePaint is the resolved painting of one shape.
type shapePaint struct {
	fill, stroke       rgb
	hasFill, hasStroke bool
	alpha              Alpha
}

// resolveShapePaint resolves fill and stroke for a shape, reporting false
// when nothing visible would be painted.
func (r *renderer) resolveShapePaint(st state, fillable bool) (shapePaint, bool) {
	var p shapePaint
	p.fill, p.hasFill = r.resolvePaint(st.fill, st)
	p.stroke, p.hasStroke = r.resolvePaint(st.stroke, st)
	p.hasFill = p.hasFill && fillable
	p.hasStroke = p.hasStroke && st.strokeWidth > 0
	p.alpha = Alpha{Fill: 1, Stroke: 1}
	if p.hasFill {
		p.alpha.Fill = p.fill.a * st.fillOpacity * st.opacity
		p.hasFill = p.alpha.Fill > 0
	}
	if p.hasStroke {
		p.alpha.Stroke = p.stroke.a * st.strokeOpacity * st.opacity
		p.hasStroke = p.alpha.Stroke > 0
	}
	return p, p.hasFill || p.hasStroke
}

// drawShape paints a basic shape with the current state.
func (r *renderer) drawShape(n *node, st state) {
	if !st.visible {
		return
	}
	segs, fillable := shapeSegments(n, st)
	if len(segs) == 0 {
		return
	}
	p, ok := r.resolveShapePaint(st, fillable)
	if !ok {
		return
	}

	r.buf.WriteString("q\n")
	if p.alpha.Fill < 1 || p.alpha.Stroke < 1 {
		fmt.Fprintf(&r.buf, "/%s gs\n", r.extGState(p.alpha))
	}
	if p.hasFill {
		fmt.Fprintf(&r.buf, "%s %s %s rg\n", fmtNum(p.fill.r, 4), fmtNum(p.fill.g, 4), fmtNum(p.fill.b, 4))
	}
	if p.hasStroke {
		r.writeStrokeState(p.stroke, st)
	}
	r.writeSegments(segs)
	r.buf.WriteString(paintOp(p.hasFill, p.hasStroke, st.fillRule == "evenodd"))
	r.buf.WriteString("Q\n")
}

func (r *renderer) writeStrokeState(c rgb, st state) {
	fmt.Fprintf(&r.buf, "%s %s %s RG\n%s w\n", fmtNum(c.r, 4), fmtNum(c.g, 4), fmtNum(c.b, 4), fmtNum(st.strokeWidth, 4))
	switch st.lineCap {
	case "round":
		r.buf.WriteString("1 J\n")
	case "square":
		r.buf.WriteString("2 J\n")
	}
	switch st.lineJoin {
	case "round":
		r.buf.WriteString("1 j\n")
	case "bevel":
		r.buf.WriteString("2 j\n")
	}
	if st.miterLimit != 10 {
		fmt.Fprintf(&r.buf, "%s M\n", fmtNum(st.miterLimit, 4))
	}
	if len(st.dash) > 0 {
		r.buf.WriteByte('[')
		for i, d := range st.dash {
			if i > 0 {
				r.buf.WriteByte(' ')
			}
			r.buf.WriteString(fmtNum(d, 4))
		}
		fmt.Fprintf(&r.buf, "] %s d\n", fmtNum(st.dashOffset, 4))
	}
}

func (r *renderer) writeSegments(segs []segment) {
	for _, s := range segs {
		switch s.op {
		case 'M':
			fmt.Fprintf(&r.buf, "%s %s m\n", fmtNum(s.p[0], 4), fmtNum(s.p[1], 4))
		case 'L':
			fmt.Fprintf(&r.buf, "%s %s l\n", fmtNum(s.p[0], 4), fmtNum(s.p[1], 4))
		case 'C':
			fmt.Fprintf(&r.buf, "%s %s %s %s %s %s c\n",
				fmtNum(s.p[0], 4), fmtNum(s.p[1], 4), fmtNum(s.p[2], 4),
				fmtNum(s.p[3], 4), fmtNum(s.p[4], 4), fmtNum(s.p[5], 4))
		case 'Z':
			r.buf.WriteString("h\n")
		}
	}
}

func paintOp(fill, stroke, evenOdd bool) string {
	switch {
	case fill && stroke && evenOdd:
		return "B*\n"
	case fill && stroke:
		return "B\n"
	case fill && evenOdd:
		return "f*\n"
	case fill:
		return "f\n"
	default:
		return "S\n"
	}
}

func (r *renderer) extGState(a Alpha) string {
	a.Fill = roundAlpha(a.Fill)
	a.Stroke = roundAlpha(a.Stroke)
	if name, ok := r.gsName[a]; ok {
		return name
	}
	name := "SG" + strconv.Itoa(len(r.gsName)+1)
	r.gsName[a] = name
	r.gs[name] = a
	return name
}

func roundAlpha(v float64) float64 {
	return float64(int(v*1000+0.5)) / 1000
}

// resolvePaint resolves a fill/stroke value to a color. Gradients are
// approximated by the color of their first stop.
func (r *renderer) resolvePaint(v string, st state) (rgb, bool) {
	v = strings.TrimSpace(v)
	if v == "" || v == "none" {
		return rgb{}, false
	}
	if strings.HasPrefix(v, "url(") {
		end := strings.IndexByte(v, ')')
		if end < 0 {
			return rgb{}, false
		}
		id := strings.Trim(strings.TrimSpace(v[4:end]), `'"`)
		if c, ok := r.gradientColor(strings.TrimPrefix(id, "#"), st, 0); ok {
			return c, true
		}
		// Fall back to the color given after the reference, if any.
		return r.resolvePaint(v[end+1:], st)
	}
	c, ok := parseColor(v, st.color)
	return c, ok && c.a > 0
}

func (r *renderer) gradientColor(id string, st state, depth int) (rgb, bool) {
	g, ok := r.img.ids[id]
	if !ok || depth > 8 || (g.name != "linearGradient" && g.name != "radialGradient") {
		return rgb{}, false
	}
	for _, c := range g.children {
		if c.name != "stop" {
			continue
		}
		props := r.img.declaredProps(c)
		col := rgb{0, 0, 0, 1}
		if v, ok := props["stop-color"]; ok {
			if pc, ok := parseColor(v, st.color); ok {
				col = pc
			}
		}
		col.a *= parseOpacity(props["stop-opacity"], 1)
		return col, col.a > 0
	}
	// No stops of its own: follow the template reference.
	if href := g.attrs["href"]; strings.HasPrefix(href, "#") {
		return r.gradientColor(href[1:], st, depth+1)
	}
	return rgb{}, false
}
