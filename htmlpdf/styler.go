// Package htmlpdf converts HTML+CSS into PDF documents using the gpdf
// document model. It provides a Layer 4 CSS rendering engine that sits
// on top of gpdf's existing layout and rendering pipeline.
package htmlpdf

import (
	"strings"

	"github.com/gpdf-dev/gpdf/css"
	"github.com/gpdf-dev/gpdf/document"
	"github.com/gpdf-dev/gpdf/pdf"
)

// applyStyle converts CSS computed styles to a document.Style.
func applyStyle(computed css.ComputedStyles) document.Style {
	s := document.DefaultStyle()

	if v, ok := computed["font-family"]; ok && v != "" {
		s.FontFamily = unquote(v)
	}
	if v, ok := computed["font-size"]; ok {
		s.FontSize = css.ParsePt(v, s.FontSize)
	}
	if v, ok := computed["font-weight"]; ok {
		s.FontWeight = parseFontWeight(v)
	}
	if v, ok := computed["font-style"]; ok {
		s.FontStyle = parseFontStyle(v)
	}
	if v, ok := computed["color"]; ok {
		if c, cok := css.ParseColor(v); cok {
			s.Color = cssColorToPDF(c)
		}
	}
	if v, ok := computed["background-color"]; ok && v != "transparent" {
		if c, cok := css.ParseColor(v); cok {
			col := cssColorToPDF(c)
			s.Background = &col
		}
	}
	if v, ok := computed["text-align"]; ok {
		s.TextAlign = parseTextAlign(v)
	}
	if v, ok := computed["line-height"]; ok {
		s.LineHeight = css.ParsePt(v, s.FontSize*1.2)
	}
	if v, ok := computed["letter-spacing"]; ok && v != "normal" {
		s.LetterSpacing = css.ParsePt(v, 0)
	}
	if v, ok := computed["word-spacing"]; ok && v != "normal" {
		s.WordSpacing = css.ParsePt(v, 0)
	}
	if v, ok := computed["text-indent"]; ok {
		s.TextIndent = parseCSSValue(v)
	}
	if v, ok := computed["text-decoration"]; ok {
		s.TextDecoration = parseTextDecoration(v)
	}

	// Box model on Style
	s.Margin = parseEdges(computed, "margin")
	s.Padding = parseEdges(computed, "padding")
	s.Border = parseBorderEdges(computed)

	return s
}

// applyBoxStyle converts CSS computed styles to a document.BoxStyle.
func applyBoxStyle(computed css.ComputedStyles) document.BoxStyle {
	bs := document.BoxStyle{
		Direction: document.DirectionVertical,
	}

	bs.Width = parseCSSValue(computed["width"])
	bs.Height = parseCSSValue(computed["height"])
	bs.MinWidth = parseCSSValue(computed["min-width"])
	bs.MaxWidth = parseCSSValue(computed["max-width"])
	bs.MinHeight = parseCSSValue(computed["min-height"])
	bs.MaxHeight = parseCSSValue(computed["max-height"])
	bs.Margin = parseEdges(computed, "margin")
	bs.Padding = parseEdges(computed, "padding")
	bs.Border = parseBorderEdges(computed)

	if v, ok := computed["background-color"]; ok && v != "transparent" {
		if c, cok := css.ParseColor(v); cok {
			col := cssColorToPDF(c)
			bs.Background = &col
		}
	}

	return bs
}

// ─── Value parsing helpers ────────────────────────────────────────

func parseCSSValue(s string) document.Value {
	s = strings.TrimSpace(s)
	if s == "" || s == "auto" {
		return document.Auto
	}
	if s == "none" || s == "0" {
		return document.Pt(0)
	}

	tokens := css.ParseValue(s)
	if len(tokens) != 1 {
		return document.Auto
	}

	tok := tokens[0]
	switch tok.Type {
	case css.TokenDimension:
		return dimensionToValue(tok.Num, tok.Unit)
	case css.TokenPercentage:
		return document.Pct(tok.Num)
	case css.TokenNumber:
		if tok.Num == 0 {
			return document.Pt(0)
		}
		return document.Pt(tok.Num)
	}
	return document.Auto
}

func dimensionToValue(num float64, unit string) document.Value {
	switch strings.ToLower(unit) {
	case "pt":
		return document.Pt(num)
	case "px":
		return document.Pt(num * 0.75)
	case "mm":
		return document.Mm(num)
	case "cm":
		return document.Cm(num)
	case "in":
		return document.In(num)
	case "em":
		return document.Em(num)
	default:
		return document.Pt(num)
	}
}

func parseEdges(computed css.ComputedStyles, prefix string) document.Edges {
	return document.Edges{
		Top:    parseCSSValue(computed[prefix+"-top"]),
		Right:  parseCSSValue(computed[prefix+"-right"]),
		Bottom: parseCSSValue(computed[prefix+"-bottom"]),
		Left:   parseCSSValue(computed[prefix+"-left"]),
	}
}

func parseBorderEdges(computed css.ComputedStyles) document.BorderEdges {
	return document.BorderEdges{
		Top:    parseBorderSide(computed, "top"),
		Right:  parseBorderSide(computed, "right"),
		Bottom: parseBorderSide(computed, "bottom"),
		Left:   parseBorderSide(computed, "left"),
	}
}

func parseBorderSide(computed css.ComputedStyles, side string) document.BorderSide {
	style := parseBorderStyle(computed["border-"+side+"-style"])
	if style == document.BorderNone {
		return document.BorderSide{}
	}
	bs := document.BorderSide{
		Width: parseCSSValue(computed["border-"+side+"-width"]),
		Style: style,
	}
	if v, ok := computed["border-"+side+"-color"]; ok {
		if c, cok := css.ParseColor(v); cok {
			bs.Color = cssColorToPDF(c)
		}
	}
	return bs
}

func parseFontWeight(s string) document.FontWeight {
	s = strings.TrimSpace(s)
	switch strings.ToLower(s) {
	case "bold", "700", "800", "900":
		return document.WeightBold
	default:
		return document.WeightNormal
	}
}

func parseFontStyle(s string) document.FontStyle {
	if strings.EqualFold(strings.TrimSpace(s), "italic") || strings.EqualFold(strings.TrimSpace(s), "oblique") {
		return document.StyleItalic
	}
	return document.StyleNormal
}

func parseTextAlign(s string) document.TextAlign {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "center":
		return document.AlignCenter
	case "right":
		return document.AlignRight
	case "justify":
		return document.AlignJustify
	default:
		return document.AlignLeft
	}
}

func parseTextDecoration(s string) document.TextDecoration {
	var d document.TextDecoration
	lower := strings.ToLower(s)
	if strings.Contains(lower, "underline") {
		d |= document.DecorationUnderline
	}
	if strings.Contains(lower, "line-through") || strings.Contains(lower, "strikethrough") {
		d |= document.DecorationStrikethrough
	}
	if strings.Contains(lower, "overline") {
		d |= document.DecorationOverline
	}
	return d
}

func parseBorderStyle(s string) document.BorderStyle {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "solid":
		return document.BorderSolid
	case "dashed":
		return document.BorderDashed
	case "dotted":
		return document.BorderDotted
	default:
		return document.BorderNone
	}
}

func cssColorToPDF(c *css.Color) pdf.Color {
	return pdf.RGB(c.R, c.G, c.B)
}

func unquote(s string) string {
	s = strings.TrimSpace(s)
	if len(s) >= 2 {
		if (s[0] == '"' && s[len(s)-1] == '"') || (s[0] == '\'' && s[len(s)-1] == '\'') {
			return s[1 : len(s)-1]
		}
	}
	return s
}
