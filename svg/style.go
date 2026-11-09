package svg

import (
	"strconv"
	"strings"
)

// styleProps are the presentation properties this package understands.
var styleProps = map[string]bool{
	"fill": true, "fill-opacity": true, "fill-rule": true,
	"stroke": true, "stroke-width": true, "stroke-opacity": true,
	"stroke-linecap": true, "stroke-linejoin": true, "stroke-miterlimit": true,
	"stroke-dasharray": true, "stroke-dashoffset": true,
	"opacity": true, "display": true, "visibility": true, "color": true,
	"stop-color": true, "stop-opacity": true,
}

// cssRule is one declaration block of a <style> sheet with a simple
// selector: a tag name, ".class", "#id" or "*".
type cssRule struct {
	selector    string
	specificity int
	order       int
	decls       map[string]string
}

// parseCSS parses a style sheet, keeping only rules whose selectors are a
// single simple selector (combinators and pseudo-classes are skipped).
func parseCSS(src string) []cssRule {
	src = stripCSSComments(src)
	var rules []cssRule
	order := 0
	for {
		open := strings.IndexByte(src, '{')
		if open < 0 {
			return rules
		}
		closing := strings.IndexByte(src[open:], '}')
		if closing < 0 {
			return rules
		}
		selectors := src[:open]
		decls := parseDeclarations(src[open+1 : open+closing])
		src = src[open+closing+1:]
		if strings.HasPrefix(strings.TrimSpace(selectors), "@") {
			continue // at-rules (@media, @font-face …) are not supported
		}
		for _, sel := range strings.Split(selectors, ",") {
			sel = strings.TrimSpace(sel)
			spec, ok := selectorSpecificity(sel)
			if !ok {
				continue
			}
			rules = append(rules, cssRule{selector: sel, specificity: spec, order: order, decls: decls})
			order++
		}
	}
}

func stripCSSComments(s string) string {
	for {
		start := strings.Index(s, "/*")
		if start < 0 {
			return s
		}
		end := strings.Index(s[start+2:], "*/")
		if end < 0 {
			return s[:start]
		}
		s = s[:start] + s[start+2+end+2:]
	}
}

func selectorSpecificity(sel string) (int, bool) {
	if sel == "" || strings.ContainsAny(sel, " >+~:[") {
		return 0, false
	}
	switch {
	case sel == "*":
		return 0, true
	case sel[0] == '#':
		return 100, true
	case sel[0] == '.':
		return 10, true
	default:
		return 1, !strings.ContainsAny(sel, ".#")
	}
}

func (r cssRule) matches(n *node) bool {
	switch {
	case r.selector == "*":
		return true
	case r.selector[0] == '#':
		return n.attrs["id"] == r.selector[1:]
	case r.selector[0] == '.':
		for _, c := range strings.Fields(n.attrs["class"]) {
			if c == r.selector[1:] {
				return true
			}
		}
		return false
	default:
		return n.name == r.selector
	}
}

// parseDeclarations parses "prop: value; prop2: value2".
func parseDeclarations(s string) map[string]string {
	out := make(map[string]string)
	for _, decl := range strings.Split(s, ";") {
		colon := strings.IndexByte(decl, ':')
		if colon < 0 {
			continue
		}
		prop := strings.ToLower(strings.TrimSpace(decl[:colon]))
		val := strings.TrimSpace(decl[colon+1:])
		val = strings.TrimSpace(strings.TrimSuffix(val, "!important"))
		if prop != "" && val != "" {
			out[prop] = val
		}
	}
	return out
}

// declaredProps returns the presentation properties declared on n, in
// cascade order: presentation attributes, then style sheet rules by
// specificity, then the inline style attribute.
func (img *Image) declaredProps(n *node) map[string]string {
	props := make(map[string]string)
	for name, v := range n.attrs {
		if styleProps[name] {
			props[name] = strings.TrimSpace(v)
		}
	}
	var matched []cssRule
	for _, r := range img.css {
		if r.matches(n) {
			matched = append(matched, r)
		}
	}
	// Stable insertion sort by specificity (order is already ascending).
	for i := 1; i < len(matched); i++ {
		for j := i; j > 0 && matched[j-1].specificity > matched[j].specificity; j-- {
			matched[j-1], matched[j] = matched[j], matched[j-1]
		}
	}
	for _, r := range matched {
		for k, v := range r.decls {
			if styleProps[k] {
				props[k] = v
			}
		}
	}
	if inline := n.attrs["style"]; inline != "" {
		for k, v := range parseDeclarations(inline) {
			if styleProps[k] {
				props[k] = v
			}
		}
	}
	return props
}

// --- Colors ---------------------------------------------------------------

// rgb is a color with components in 0–1 and its own alpha.
type rgb struct {
	r, g, b, a float64
}

// parseColor parses a CSS color. currentColor resolves to cur.
func parseColor(s string, cur rgb) (rgb, bool) {
	s = strings.TrimSpace(s)
	lower := strings.ToLower(s)
	switch {
	case lower == "currentcolor":
		return cur, true
	case lower == "transparent":
		return rgb{}, true
	case strings.HasPrefix(s, "#"):
		return parseHexColor(s[1:])
	case strings.HasPrefix(lower, "rgb"):
		return parseRGBFunc(lower)
	}
	if v, ok := namedColors[lower]; ok {
		return rgb{float64(v>>16&0xFF) / 255, float64(v>>8&0xFF) / 255, float64(v&0xFF) / 255, 1}, true
	}
	return rgb{}, false
}

func parseHexColor(h string) (rgb, bool) {
	expand := func(s string) string {
		var b strings.Builder
		for i := 0; i < len(s); i++ {
			b.WriteByte(s[i])
			b.WriteByte(s[i])
		}
		return b.String()
	}
	switch len(h) {
	case 3, 4:
		h = expand(h)
	case 6, 8:
	default:
		return rgb{}, false
	}
	v, err := strconv.ParseUint(h, 16, 32)
	if err != nil {
		return rgb{}, false
	}
	a := 1.0
	if len(h) == 8 {
		a = float64(v&0xFF) / 255
		v >>= 8
	}
	return rgb{float64(v>>16&0xFF) / 255, float64(v>>8&0xFF) / 255, float64(v&0xFF) / 255, a}, true
}

func parseRGBFunc(s string) (rgb, bool) {
	open := strings.IndexByte(s, '(')
	closing := strings.LastIndexByte(s, ')')
	if open < 0 || closing < open {
		return rgb{}, false
	}
	parts := strings.FieldsFunc(s[open+1:closing], func(r rune) bool { return r == ',' || r == ' ' || r == '/' })
	if len(parts) < 3 {
		return rgb{}, false
	}
	var c [4]float64
	c[3] = 1
	for i := 0; i < len(parts) && i < 4; i++ {
		p := parts[i]
		pct := strings.HasSuffix(p, "%")
		v, err := strconv.ParseFloat(strings.TrimSuffix(p, "%"), 64)
		if err != nil {
			return rgb{}, false
		}
		switch {
		case pct:
			v /= 100
		case i < 3:
			v /= 255
		}
		c[i] = clamp01(v)
	}
	return rgb{c[0], c[1], c[2], c[3]}, true
}

func clamp01(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}

// namedColors are the CSS Color Module Level 4 named colors.
var namedColors = map[string]uint32{
	"aliceblue": 0xF0F8FF, "antiquewhite": 0xFAEBD7, "aqua": 0x00FFFF, "aquamarine": 0x7FFFD4,
	"azure": 0xF0FFFF, "beige": 0xF5F5DC, "bisque": 0xFFE4C4, "black": 0x000000,
	"blanchedalmond": 0xFFEBCD, "blue": 0x0000FF, "blueviolet": 0x8A2BE2, "brown": 0xA52A2A,
	"burlywood": 0xDEB887, "cadetblue": 0x5F9EA0, "chartreuse": 0x7FFF00, "chocolate": 0xD2691E,
	"coral": 0xFF7F50, "cornflowerblue": 0x6495ED, "cornsilk": 0xFFF8DC, "crimson": 0xDC143C,
	"cyan": 0x00FFFF, "darkblue": 0x00008B, "darkcyan": 0x008B8B, "darkgoldenrod": 0xB8860B,
	"darkgray": 0xA9A9A9, "darkgreen": 0x006400, "darkgrey": 0xA9A9A9, "darkkhaki": 0xBDB76B,
	"darkmagenta": 0x8B008B, "darkolivegreen": 0x556B2F, "darkorange": 0xFF8C00, "darkorchid": 0x9932CC,
	"darkred": 0x8B0000, "darksalmon": 0xE9967A, "darkseagreen": 0x8FBC8F, "darkslateblue": 0x483D8B,
	"darkslategray": 0x2F4F4F, "darkslategrey": 0x2F4F4F, "darkturquoise": 0x00CED1, "darkviolet": 0x9400D3,
	"deeppink": 0xFF1493, "deepskyblue": 0x00BFFF, "dimgray": 0x696969, "dimgrey": 0x696969,
	"dodgerblue": 0x1E90FF, "firebrick": 0xB22222, "floralwhite": 0xFFFAF0, "forestgreen": 0x228B22,
	"fuchsia": 0xFF00FF, "gainsboro": 0xDCDCDC, "ghostwhite": 0xF8F8FF, "gold": 0xFFD700,
	"goldenrod": 0xDAA520, "gray": 0x808080, "green": 0x008000, "greenyellow": 0xADFF2F,
	"grey": 0x808080, "honeydew": 0xF0FFF0, "hotpink": 0xFF69B4, "indianred": 0xCD5C5C,
	"indigo": 0x4B0082, "ivory": 0xFFFFF0, "khaki": 0xF0E68C, "lavender": 0xE6E6FA,
	"lavenderblush": 0xFFF0F5, "lawngreen": 0x7CFC00, "lemonchiffon": 0xFFFACD, "lightblue": 0xADD8E6,
	"lightcoral": 0xF08080, "lightcyan": 0xE0FFFF, "lightgoldenrodyellow": 0xFAFAD2, "lightgray": 0xD3D3D3,
	"lightgreen": 0x90EE90, "lightgrey": 0xD3D3D3, "lightpink": 0xFFB6C1, "lightsalmon": 0xFFA07A,
	"lightseagreen": 0x20B2AA, "lightskyblue": 0x87CEFA, "lightslategray": 0x778899, "lightslategrey": 0x778899,
	"lightsteelblue": 0xB0C4DE, "lightyellow": 0xFFFFE0, "lime": 0x00FF00, "limegreen": 0x32CD32,
	"linen": 0xFAF0E6, "magenta": 0xFF00FF, "maroon": 0x800000, "mediumaquamarine": 0x66CDAA,
	"mediumblue": 0x0000CD, "mediumorchid": 0xBA55D3, "mediumpurple": 0x9370DB, "mediumseagreen": 0x3CB371,
	"mediumslateblue": 0x7B68EE, "mediumspringgreen": 0x00FA9A, "mediumturquoise": 0x48D1CC, "mediumvioletred": 0xC71585,
	"midnightblue": 0x191970, "mintcream": 0xF5FFFA, "mistyrose": 0xFFE4E1, "moccasin": 0xFFE4B5,
	"navajowhite": 0xFFDEAD, "navy": 0x000080, "oldlace": 0xFDF5E6, "olive": 0x808000,
	"olivedrab": 0x6B8E23, "orange": 0xFFA500, "orangered": 0xFF4500, "orchid": 0xDA70D6,
	"palegoldenrod": 0xEEE8AA, "palegreen": 0x98FB98, "paleturquoise": 0xAFEEEE, "palevioletred": 0xDB7093,
	"papayawhip": 0xFFEFD5, "peachpuff": 0xFFDAB9, "peru": 0xCD853F, "pink": 0xFFC0CB,
	"plum": 0xDDA0DD, "powderblue": 0xB0E0E6, "purple": 0x800080, "rebeccapurple": 0x663399,
	"red": 0xFF0000, "rosybrown": 0xBC8F8F, "royalblue": 0x4169E1, "saddlebrown": 0x8B4513,
	"salmon": 0xFA8072, "sandybrown": 0xF4A460, "seagreen": 0x2E8B57, "seashell": 0xFFF5EE,
	"sienna": 0xA0522D, "silver": 0xC0C0C0, "skyblue": 0x87CEEB, "slateblue": 0x6A5ACD,
	"slategray": 0x708090, "slategrey": 0x708090, "snow": 0xFFFAFA, "springgreen": 0x00FF7F,
	"steelblue": 0x4682B4, "tan": 0xD2B48C, "teal": 0x008080, "thistle": 0xD8BFD8,
	"tomato": 0xFF6347, "turquoise": 0x40E0D0, "violet": 0xEE82EE, "wheat": 0xF5DEB3,
	"white": 0xFFFFFF, "whitesmoke": 0xF5F5F5, "yellow": 0xFFFF00, "yellowgreen": 0x9ACD32,
}
