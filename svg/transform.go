package svg

import (
	"math"
	"strings"
)

// matrix is an affine transform [a b c d e f] mapping (x, y) to
// (a*x + c*y + e, b*x + d*y + f), as in SVG and PDF.
type matrix [6]float64

var identity = matrix{1, 0, 0, 1, 0, 0}

// mul returns m × n: n is applied first, then m.
func (m matrix) mul(n matrix) matrix {
	return matrix{
		m[0]*n[0] + m[2]*n[1],
		m[1]*n[0] + m[3]*n[1],
		m[0]*n[2] + m[2]*n[3],
		m[1]*n[2] + m[3]*n[3],
		m[0]*n[4] + m[2]*n[5] + m[4],
		m[1]*n[4] + m[3]*n[5] + m[5],
	}
}

// parseTransform parses an SVG transform list such as
// "translate(10 20) rotate(45)". Unknown or malformed entries end parsing;
// the transforms read so far are kept.
func parseTransform(s string) matrix {
	m := identity
	for {
		s = strings.TrimLeft(s, " \t\r\n,")
		open := strings.IndexByte(s, '(')
		if open < 0 {
			return m
		}
		closing := strings.IndexByte(s[open:], ')')
		if closing < 0 {
			return m
		}
		name := strings.TrimSpace(s[:open])
		args := parseNumberList(s[open+1 : open+closing])
		t, ok := transformFunc(name, args)
		if !ok {
			return m
		}
		m = m.mul(t)
		s = s[open+closing+1:]
	}
}

func transformFunc(name string, a []float64) (matrix, bool) {
	switch name {
	case "matrix":
		if len(a) == 6 {
			return matrix{a[0], a[1], a[2], a[3], a[4], a[5]}, true
		}
	case "translate":
		switch len(a) {
		case 1:
			return matrix{1, 0, 0, 1, a[0], 0}, true
		case 2:
			return matrix{1, 0, 0, 1, a[0], a[1]}, true
		}
	case "scale":
		switch len(a) {
		case 1:
			return matrix{a[0], 0, 0, a[0], 0, 0}, true
		case 2:
			return matrix{a[0], 0, 0, a[1], 0, 0}, true
		}
	case "rotate":
		return rotateTransform(a)
	case "skewX":
		if len(a) == 1 {
			return matrix{1, 0, math.Tan(a[0] * math.Pi / 180), 1, 0, 0}, true
		}
	case "skewY":
		if len(a) == 1 {
			return matrix{1, math.Tan(a[0] * math.Pi / 180), 0, 1, 0, 0}, true
		}
	}
	return identity, false
}

func rotateTransform(a []float64) (matrix, bool) {
	if len(a) != 1 && len(a) != 3 {
		return identity, false
	}
	rad := a[0] * math.Pi / 180
	cos, sin := math.Cos(rad), math.Sin(rad)
	r := matrix{cos, sin, -sin, cos, 0, 0}
	if len(a) == 3 {
		cx, cy := a[1], a[2]
		return matrix{1, 0, 0, 1, cx, cy}.mul(r).mul(matrix{1, 0, 0, 1, -cx, -cy}), true
	}
	return r, true
}
