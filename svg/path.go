package svg

import (
	"math"
	"strings"
)

// segment is a normalized path segment: move, line, cubic curve or close.
type segment struct {
	op byte       // 'M', 'L', 'C' or 'Z'
	p  [6]float64 // M/L: p[0:2]; C: c1, c2, end
}

// pathBuilder accumulates normalized segments while tracking the state
// needed by relative and smooth commands.
type pathBuilder struct {
	segs         []segment
	cx, cy       float64 // current point
	sx, sy       float64 // start of current subpath
	lastCtrl     [2]float64
	lastCmd      byte // previous command (upper case) for smooth curves
	hasSubpath   bool
	pendingClose bool
}

func (b *pathBuilder) moveTo(x, y float64) {
	b.segs = append(b.segs, segment{op: 'M', p: [6]float64{x, y}})
	b.cx, b.cy, b.sx, b.sy = x, y, x, y
	b.hasSubpath = true
	b.pendingClose = false
}

func (b *pathBuilder) lineTo(x, y float64) {
	b.ensureSubpath()
	b.segs = append(b.segs, segment{op: 'L', p: [6]float64{x, y}})
	b.cx, b.cy = x, y
}

func (b *pathBuilder) curveTo(x1, y1, x2, y2, x, y float64) {
	b.ensureSubpath()
	b.segs = append(b.segs, segment{op: 'C', p: [6]float64{x1, y1, x2, y2, x, y}})
	b.lastCtrl = [2]float64{x2, y2}
	b.cx, b.cy = x, y
}

func (b *pathBuilder) closePath() {
	if !b.hasSubpath {
		return
	}
	b.segs = append(b.segs, segment{op: 'Z'})
	b.cx, b.cy = b.sx, b.sy
	b.pendingClose = true
}

// ensureSubpath starts a new subpath at the current point after a close,
// as drawing commands following Z continue from the subpath start.
func (b *pathBuilder) ensureSubpath() {
	if !b.hasSubpath || b.pendingClose {
		b.segs = append(b.segs, segment{op: 'M', p: [6]float64{b.cx, b.cy}})
		b.sx, b.sy = b.cx, b.cy
		b.hasSubpath = true
		b.pendingClose = false
	}
}

// quadTo appends a quadratic Bézier as the equivalent cubic.
func (b *pathBuilder) quadTo(qx, qy, x, y float64) {
	x0, y0 := b.cx, b.cy
	b.curveTo(x0+2.0/3*(qx-x0), y0+2.0/3*(qy-y0), x+2.0/3*(qx-x), y+2.0/3*(qy-y), x, y)
	b.lastCtrl = [2]float64{qx, qy}
}

// parsePathData converts SVG path data into normalized segments. Parsing
// stops at the first error, keeping everything read so far (as browsers do).
func parsePathData(d string) []segment {
	b := &pathBuilder{}
	sc := numScanner{s: d}
	var cmd byte
	for {
		sc.skipSep()
		if sc.pos >= len(sc.s) {
			break
		}
		c := sc.s[sc.pos]
		if isPathCommand(c) {
			cmd = c
			sc.pos++
		} else if cmd == 0 {
			break
		}
		if !b.command(cmd, &sc) {
			break
		}
		// An implicit repeat of M is treated as L.
		switch cmd {
		case 'M':
			cmd = 'L'
		case 'm':
			cmd = 'l'
		case 'Z', 'z':
			cmd = 0
		}
	}
	return b.segs
}

func isPathCommand(c byte) bool {
	return strings.IndexByte("MmLlHhVvCcSsQqTtAaZz", c) >= 0
}

// command executes one command with its arguments, reporting false when
// the arguments are missing or malformed.
func (b *pathBuilder) command(cmd byte, sc *numScanner) bool {
	rel := cmd >= 'a'
	upper := cmd &^ 0x20
	var ox, oy float64
	if rel {
		ox, oy = b.cx, b.cy
	}
	ok := true
	switch upper {
	case 'Z':
		b.closePath()
	case 'M', 'L', 'T':
		ok = b.pointCommand(upper, sc, ox, oy)
	case 'H':
		var x float64
		if x, ok = sc.number(); ok {
			b.lineTo(x+ox, b.cy)
		}
	case 'V':
		var y float64
		if y, ok = sc.number(); ok {
			b.lineTo(b.cx, y+oy)
		}
	case 'C', 'S', 'Q':
		ok = b.curveCommand(upper, sc, ox, oy)
	case 'A':
		ok = b.arcCommand(sc, ox, oy)
	}
	b.lastCmd = upper
	return ok
}

func readNums(sc *numScanner, n int) ([6]float64, bool) {
	var v [6]float64
	for i := 0; i < n; i++ {
		x, ok := sc.number()
		if !ok {
			return v, false
		}
		v[i] = x
	}
	return v, true
}

func (b *pathBuilder) pointCommand(upper byte, sc *numScanner, ox, oy float64) bool {
	v, ok := readNums(sc, 2)
	if !ok {
		return false
	}
	x, y := v[0]+ox, v[1]+oy
	switch upper {
	case 'M':
		b.moveTo(x, y)
	case 'L':
		b.lineTo(x, y)
	case 'T':
		qx, qy := b.cx, b.cy
		if b.lastCmd == 'Q' || b.lastCmd == 'T' {
			qx, qy = 2*b.cx-b.lastCtrl[0], 2*b.cy-b.lastCtrl[1]
		}
		b.quadTo(qx, qy, x, y)
	}
	return true
}

func (b *pathBuilder) curveCommand(upper byte, sc *numScanner, ox, oy float64) bool {
	switch upper {
	case 'C':
		v, ok := readNums(sc, 6)
		if ok {
			b.curveTo(v[0]+ox, v[1]+oy, v[2]+ox, v[3]+oy, v[4]+ox, v[5]+oy)
		}
		return ok
	case 'S':
		v, ok := readNums(sc, 4)
		if ok {
			x1, y1 := b.cx, b.cy
			if b.lastCmd == 'C' || b.lastCmd == 'S' {
				x1, y1 = 2*b.cx-b.lastCtrl[0], 2*b.cy-b.lastCtrl[1]
			}
			b.curveTo(x1, y1, v[0]+ox, v[1]+oy, v[2]+ox, v[3]+oy)
		}
		return ok
	default: // 'Q'
		v, ok := readNums(sc, 4)
		if ok {
			b.quadTo(v[0]+ox, v[1]+oy, v[2]+ox, v[3]+oy)
		}
		return ok
	}
}

func (b *pathBuilder) arcCommand(sc *numScanner, ox, oy float64) bool {
	v, ok := readNums(sc, 3)
	if !ok {
		return false
	}
	large, ok1 := sc.flag()
	sweep, ok2 := sc.flag()
	end, ok3 := readNums(sc, 2)
	if !ok1 || !ok2 || !ok3 {
		return false
	}
	b.arcTo(v[0], v[1], v[2], large, sweep, end[0]+ox, end[1]+oy)
	return true
}

// arcTo appends an elliptical arc as cubic Béziers, following the
// endpoint-to-center conversion in the SVG 1.1 implementation notes (F.6).
func (b *pathBuilder) arcTo(rx, ry, angle float64, large, sweep bool, x, y float64) {
	x1, y1 := b.cx, b.cy
	if x1 == x && y1 == y {
		return
	}
	rx, ry = math.Abs(rx), math.Abs(ry)
	if rx == 0 || ry == 0 {
		b.lineTo(x, y)
		return
	}
	a := centerArc(x1, y1, rx, ry, angle*math.Pi/180, large, sweep, x, y)

	n := max(int(math.Ceil(math.Abs(a.dtheta)/(math.Pi/2))), 1)
	delta := a.dtheta / float64(n)
	k := 4.0 / 3 * math.Tan(delta/4)

	t := a.theta1
	px, py := x1, y1
	for i := 0; i < n; i++ {
		t2 := t + delta
		ex, ey := a.point(t2)
		if i == n-1 {
			ex, ey = x, y
		}
		d1x, d1y := a.deriv(t)
		d2x, d2y := a.deriv(t2)
		b.curveTo(px+k*d1x, py+k*d1y, ex-k*d2x, ey-k*d2y, ex, ey)
		px, py = ex, ey
		t = t2
	}
}

// arc is an elliptical arc in center parameterization.
type arc struct {
	cx, cy, rx, ry float64
	cosPhi, sinPhi float64
	theta1, dtheta float64
}

func (a arc) point(t float64) (float64, float64) {
	ex, ey := a.rx*math.Cos(t), a.ry*math.Sin(t)
	return a.cosPhi*ex - a.sinPhi*ey + a.cx, a.sinPhi*ex + a.cosPhi*ey + a.cy
}

func (a arc) deriv(t float64) (float64, float64) {
	ex, ey := -a.rx*math.Sin(t), a.ry*math.Cos(t)
	return a.cosPhi*ex - a.sinPhi*ey, a.sinPhi*ex + a.cosPhi*ey
}

// centerArc converts an endpoint-parameterized arc to center form,
// scaling up radii that are too small to reach the end point.
func centerArc(x1, y1, rx, ry, phi float64, large, sweep bool, x, y float64) arc {
	cosPhi, sinPhi := math.Cos(phi), math.Sin(phi)
	dx, dy := (x1-x)/2, (y1-y)/2
	x1p := cosPhi*dx + sinPhi*dy
	y1p := -sinPhi*dx + cosPhi*dy

	if l := x1p*x1p/(rx*rx) + y1p*y1p/(ry*ry); l > 1 {
		s := math.Sqrt(l)
		rx, ry = rx*s, ry*s
	}

	num := rx*rx*ry*ry - rx*rx*y1p*y1p - ry*ry*x1p*x1p
	den := rx*rx*y1p*y1p + ry*ry*x1p*x1p
	coef := 0.0
	if den != 0 && num > 0 {
		coef = math.Sqrt(num / den)
	}
	if large == sweep {
		coef = -coef
	}
	cxp := coef * rx * y1p / ry
	cyp := -coef * ry * x1p / rx

	theta1 := vecAngle(1, 0, (x1p-cxp)/rx, (y1p-cyp)/ry)
	dtheta := vecAngle((x1p-cxp)/rx, (y1p-cyp)/ry, (-x1p-cxp)/rx, (-y1p-cyp)/ry)
	if !sweep && dtheta > 0 {
		dtheta -= 2 * math.Pi
	} else if sweep && dtheta < 0 {
		dtheta += 2 * math.Pi
	}
	return arc{
		cx: cosPhi*cxp - sinPhi*cyp + (x1+x)/2,
		cy: sinPhi*cxp + cosPhi*cyp + (y1+y)/2,
		rx: rx, ry: ry, cosPhi: cosPhi, sinPhi: sinPhi,
		theta1: theta1, dtheta: dtheta,
	}
}

// vecAngle returns the signed angle from vector u to vector v.
func vecAngle(ux, uy, vx, vy float64) float64 {
	return math.Atan2(ux*vy-uy*vx, ux*vx+uy*vy)
}

// --- Basic shapes -------------------------------------------------------

// bezierCircle is the control distance (fraction of the radius) that best
// approximates a quarter circle with one cubic Bézier.
const bezierCircle = 0.5522847498

func ellipseSegments(cx, cy, rx, ry float64) []segment {
	if rx <= 0 || ry <= 0 {
		return nil
	}
	kx, ky := rx*bezierCircle, ry*bezierCircle
	b := &pathBuilder{}
	b.moveTo(cx+rx, cy)
	b.curveTo(cx+rx, cy+ky, cx+kx, cy+ry, cx, cy+ry)
	b.curveTo(cx-kx, cy+ry, cx-rx, cy+ky, cx-rx, cy)
	b.curveTo(cx-rx, cy-ky, cx-kx, cy-ry, cx, cy-ry)
	b.curveTo(cx+kx, cy-ry, cx+rx, cy-ky, cx+rx, cy)
	b.closePath()
	return b.segs
}

func rectSegments(x, y, w, h, rx, ry float64) []segment {
	if w <= 0 || h <= 0 {
		return nil
	}
	rx, ry = math.Min(rx, w/2), math.Min(ry, h/2)
	b := &pathBuilder{}
	if rx <= 0 || ry <= 0 {
		b.moveTo(x, y)
		b.lineTo(x+w, y)
		b.lineTo(x+w, y+h)
		b.lineTo(x, y+h)
		b.closePath()
		return b.segs
	}
	kx, ky := rx*bezierCircle, ry*bezierCircle
	b.moveTo(x+rx, y)
	b.lineTo(x+w-rx, y)
	b.curveTo(x+w-rx+kx, y, x+w, y+ry-ky, x+w, y+ry)
	b.lineTo(x+w, y+h-ry)
	b.curveTo(x+w, y+h-ry+ky, x+w-rx+kx, y+h, x+w-rx, y+h)
	b.lineTo(x+rx, y+h)
	b.curveTo(x+rx-kx, y+h, x, y+h-ry+ky, x, y+h-ry)
	b.lineTo(x, y+ry)
	b.curveTo(x, y+ry-ky, x+rx-kx, y, x+rx, y)
	b.closePath()
	return b.segs
}

func polySegments(points string, closed bool) []segment {
	nums := parsePointsList(points)
	if len(nums) < 4 {
		return nil
	}
	b := &pathBuilder{}
	b.moveTo(nums[0], nums[1])
	for i := 2; i+1 < len(nums); i += 2 {
		b.lineTo(nums[i], nums[i+1])
	}
	if closed {
		b.closePath()
	}
	return b.segs
}

// parsePointsList parses a points attribute, dropping an odd trailing
// coordinate.
func parsePointsList(s string) []float64 {
	nums := parseNumberList(s)
	return nums[:len(nums)&^1]
}
