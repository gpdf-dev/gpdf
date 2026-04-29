package template

import (
	"github.com/gpdf-dev/gpdf/document"
	"github.com/gpdf-dev/gpdf/pdf"
)

// BorderSpec describes how to draw a border around a box-like element such
// as a Table, Image, or generic Box. Construct it with [Border] and one or
// more BorderOption values, then apply with [WithTableBorder],
// [WithImageBorder], [WithBoxBorder], or [WithTextBorder].
//
// The zero value of BorderSpec disables the border on all four edges.
type BorderSpec struct {
	top, right, bottom, left document.Value
	color                    pdf.Color
	style                    document.BorderStyle
	hasWidth                 bool
}

// BorderOption configures a [BorderSpec].
type BorderOption func(*BorderSpec)

// Border builds a [BorderSpec] from one or more [BorderOption] values.
//
// Defaults when no width is supplied: 1 pt uniform width, solid line, black.
// If width is supplied via [BorderWidth] or [BorderWidths], those values win.
func Border(opts ...BorderOption) BorderSpec {
	spec := BorderSpec{
		color: pdf.Black,
		style: document.BorderSolid,
	}
	for _, opt := range opts {
		opt(&spec)
	}
	if !spec.hasWidth {
		w := document.Pt(1)
		spec.top, spec.right, spec.bottom, spec.left = w, w, w, w
	}
	return spec
}

// BorderWidth sets the same width for all four edges.
func BorderWidth(v document.Value) BorderOption {
	return func(s *BorderSpec) {
		s.top, s.right, s.bottom, s.left = v, v, v, v
		s.hasWidth = true
	}
}

// BorderWidths sets per-edge widths in CSS order: top, right, bottom, left.
func BorderWidths(top, right, bottom, left document.Value) BorderOption {
	return func(s *BorderSpec) {
		s.top, s.right, s.bottom, s.left = top, right, bottom, left
		s.hasWidth = true
	}
}

// BorderColor sets the color used for all edges.
func BorderColor(c pdf.Color) BorderOption {
	return func(s *BorderSpec) { s.color = c }
}

// BorderLine sets the line style used for all edges.
// Use [document.BorderSolid], [document.BorderDashed], or [document.BorderDotted].
func BorderLine(style document.BorderStyle) BorderOption {
	return func(s *BorderSpec) { s.style = style }
}

// toEdges converts the spec to a [document.BorderEdges] value, applying the
// uniform color and line style to every side. Edges with zero width fall back
// to [document.BorderNone] so the renderer skips them.
func (s BorderSpec) toEdges() document.BorderEdges {
	side := func(w document.Value) document.BorderSide {
		st := s.style
		if w.Amount == 0 {
			st = document.BorderNone
		}
		return document.BorderSide{Width: w, Style: st, Color: s.color}
	}
	return document.BorderEdges{
		Top:    side(s.top),
		Right:  side(s.right),
		Bottom: side(s.bottom),
		Left:   side(s.left),
	}
}
