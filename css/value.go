package css

import (
	"fmt"
	"math"
	"strings"
)

// CSSValue represents a parsed CSS value.
type CSSValue struct {
	Type     CSSValueType
	Str      string    // for Ident, String, custom strings
	Num      float64   // for Number, Dimension, Percentage
	Unit     string    // for Dimension (e.g. "px", "pt", "em")
	Color    *Color    // for Color values
	Values   []CSSValue // for multi-value (e.g. margin: 1px 2px 3px 4px)
	Function string    // function name for Function values (e.g. "rgb")
}

// CSSValueType classifies a CSS value.
type CSSValueType int

const (
	ValueIdent      CSSValueType = iota // keyword (e.g. "auto", "bold", "inherit")
	ValueString                         // quoted string
	ValueNumber                         // bare number (e.g. line-height: 1.5)
	ValueDimension                      // number + unit (e.g. 12px, 2em)
	ValuePercentage                     // number + % (e.g. 50%)
	ValueColor                          // color value (#hex, rgb(), named)
	ValueFunction                       // function call (e.g. rgb(255,0,0))
)

// String returns a debug representation of the value.
func (v CSSValue) String() string {
	switch v.Type {
	case ValueIdent:
		return v.Str
	case ValueString:
		return fmt.Sprintf("%q", v.Str)
	case ValueNumber:
		return fmt.Sprintf("%g", v.Num)
	case ValueDimension:
		return fmt.Sprintf("%g%s", v.Num, v.Unit)
	case ValuePercentage:
		return fmt.Sprintf("%g%%", v.Num)
	case ValueColor:
		if v.Color != nil {
			return v.Color.String()
		}
		return v.Str
	case ValueFunction:
		return fmt.Sprintf("%s(...)", v.Function)
	}
	return v.Str
}

// IsZero reports whether the value is numerically zero.
func (v CSSValue) IsZero() bool {
	switch v.Type {
	case ValueNumber, ValueDimension, ValuePercentage:
		return v.Num == 0
	}
	return false
}

// Color represents an RGBA color.
type Color struct {
	R, G, B, A float64 // 0.0–1.0
}

// String returns the CSS representation of the color.
func (c *Color) String() string {
	if c.A < 1.0 {
		return fmt.Sprintf("rgba(%.0f,%.0f,%.0f,%.2f)", c.R*255, c.G*255, c.B*255, c.A)
	}
	return fmt.Sprintf("rgb(%.0f,%.0f,%.0f)", c.R*255, c.G*255, c.B*255)
}

// ─── Named CSS colors ─────────────────────────────────────────────

// namedColors maps CSS color keywords to RGB hex values.
var namedColors = map[string]uint32{
	"black":                0x000000,
	"silver":               0xC0C0C0,
	"gray":                 0x808080,
	"grey":                 0x808080,
	"white":                0xFFFFFF,
	"maroon":               0x800000,
	"red":                  0xFF0000,
	"purple":               0x800080,
	"fuchsia":              0xFF00FF,
	"green":                0x008000,
	"lime":                 0x00FF00,
	"olive":                0x808000,
	"yellow":               0xFFFF00,
	"navy":                 0x000080,
	"blue":                 0x0000FF,
	"teal":                 0x008080,
	"aqua":                 0x00FFFF,
	"orange":               0xFFA500,
	"aliceblue":            0xF0F8FF,
	"antiquewhite":         0xFAEBD7,
	"aquamarine":           0x7FFFD4,
	"azure":                0xF0FFFF,
	"beige":                0xF5F5DC,
	"bisque":               0xFFE4C4,
	"blanchedalmond":       0xFFEBCD,
	"blueviolet":           0x8A2BE2,
	"brown":                0xA52A2A,
	"burlywood":            0xDEB887,
	"cadetblue":            0x5F9EA0,
	"chartreuse":           0x7FFF00,
	"chocolate":            0xD2691E,
	"coral":                0xFF7F50,
	"cornflowerblue":       0x6495ED,
	"cornsilk":             0xFFF8DC,
	"crimson":              0xDC143C,
	"cyan":                 0x00FFFF,
	"darkblue":             0x00008B,
	"darkcyan":             0x008B8B,
	"darkgoldenrod":        0xB8860B,
	"darkgray":             0xA9A9A9,
	"darkgrey":             0xA9A9A9,
	"darkgreen":            0x006400,
	"darkkhaki":            0xBDB76B,
	"darkmagenta":          0x8B008B,
	"darkolivegreen":       0x556B2F,
	"darkorange":           0xFF8C00,
	"darkorchid":           0x9932CC,
	"darkred":              0x8B0000,
	"darksalmon":           0xE9967A,
	"darkseagreen":         0x8FBC8F,
	"darkslateblue":        0x483D8B,
	"darkslategray":        0x2F4F4F,
	"darkslategrey":        0x2F4F4F,
	"darkturquoise":        0x00CED1,
	"darkviolet":           0x9400D3,
	"deeppink":             0xFF1493,
	"deepskyblue":          0x00BFFF,
	"dimgray":              0x696969,
	"dimgrey":              0x696969,
	"dodgerblue":           0x1E90FF,
	"firebrick":            0xB22222,
	"floralwhite":          0xFFFAF0,
	"forestgreen":          0x228B22,
	"gainsboro":            0xDCDCDC,
	"ghostwhite":           0xF8F8FF,
	"gold":                 0xFFD700,
	"goldenrod":            0xDAA520,
	"greenyellow":          0xADFF2F,
	"honeydew":             0xF0FFF0,
	"hotpink":              0xFF69B4,
	"indianred":            0xCD5C5C,
	"indigo":               0x4B0082,
	"ivory":                0xFFFFF0,
	"khaki":                0xF0E68C,
	"lavender":             0xE6E6FA,
	"lavenderblush":        0xFFF0F5,
	"lawngreen":            0x7CFC00,
	"lemonchiffon":         0xFFFACD,
	"lightblue":            0xADD8E6,
	"lightcoral":           0xF08080,
	"lightcyan":            0xE0FFFF,
	"lightgoldenrodyellow": 0xFAFAD2,
	"lightgray":            0xD3D3D3,
	"lightgrey":            0xD3D3D3,
	"lightgreen":           0x90EE90,
	"lightpink":            0xFFB6C1,
	"lightsalmon":          0xFFA07A,
	"lightseagreen":        0x20B2AA,
	"lightskyblue":         0x87CEFA,
	"lightslategray":       0x778899,
	"lightslategrey":       0x778899,
	"lightsteelblue":       0xB0C4DE,
	"lightyellow":          0xFFFFE0,
	"limegreen":            0x32CD32,
	"linen":                0xFAF0E6,
	"magenta":              0xFF00FF,
	"mediumaquamarine":     0x66CDAA,
	"mediumblue":           0x0000CD,
	"mediumorchid":         0xBA55D3,
	"mediumpurple":         0x9370DB,
	"mediumseagreen":       0x3CB371,
	"mediumslateblue":      0x7B68EE,
	"mediumspringgreen":    0x00FA9A,
	"mediumturquoise":      0x48D1CC,
	"mediumvioletred":      0xC71585,
	"midnightblue":         0x191970,
	"mintcream":            0xF5FFFA,
	"mistyrose":            0xFFE4E1,
	"moccasin":             0xFFE4B5,
	"navajowhite":          0xFFDEAD,
	"oldlace":              0xFDF5E6,
	"olivedrab":            0x6B8E23,
	"orangered":            0xFF4500,
	"orchid":               0xDA70D6,
	"palegoldenrod":        0xEEE8AA,
	"palegreen":            0x98FB98,
	"paleturquoise":        0xAFEEEE,
	"palevioletred":        0xDB7093,
	"papayawhip":           0xFFEFD5,
	"peachpuff":            0xFFDAB9,
	"peru":                 0xCD853F,
	"pink":                 0xFFC0CB,
	"plum":                 0xDDA0DD,
	"powderblue":           0xB0E0E6,
	"rebeccapurple":        0x663399,
	"rosybrown":            0xBC8F8F,
	"royalblue":            0x4169E1,
	"saddlebrown":          0x8B4513,
	"salmon":               0xFA8072,
	"sandybrown":           0xF4A460,
	"seagreen":             0x2E8B57,
	"seashell":             0xFFF5EE,
	"sienna":               0xA0522D,
	"skyblue":              0x87CEEB,
	"slateblue":            0x6A5ACD,
	"slategray":            0x708090,
	"slategrey":            0x708090,
	"snow":                 0xFFFAFA,
	"springgreen":          0x00FF7F,
	"steelblue":            0x4682B4,
	"tan":                  0xD2B48C,
	"thistle":              0xD8BFD8,
	"tomato":               0xFF6347,
	"turquoise":            0x40E0D0,
	"violet":               0xEE82EE,
	"wheat":                0xF5DEB3,
	"whitesmoke":           0xF5F5F5,
	"yellowgreen":          0x9ACD32,
	"transparent":          0x000000, // special: alpha=0
}

// ParseColor parses a CSS color value from tokens.
// Supports: #hex, rgb(), rgba(), named colors.
func ParseColor(s string) (*Color, bool) {
	s = strings.TrimSpace(s)
	lower := strings.ToLower(s)

	// Transparent
	if lower == "transparent" {
		return &Color{0, 0, 0, 0}, true
	}

	// Named color
	if hex, ok := namedColors[lower]; ok {
		return hexToColor(hex, 1.0), true
	}

	// Hex color
	if len(s) > 0 && s[0] == '#' {
		return parseHexColor(s[1:])
	}

	return nil, false
}

// ParseColorFromTokens parses a color from CSS tokens including function calls.
func ParseColorFromTokens(tokens []Token) (*Color, bool) {
	if len(tokens) == 0 {
		return nil, false
	}

	// Single token: ident or hash
	if len(tokens) == 1 {
		tok := tokens[0]
		if tok.Type == TokenHash {
			return parseHexColor(tok.Value)
		}
		if tok.Type == TokenIdent {
			return ParseColor(tok.Value)
		}
	}

	// Function: rgb(...) or rgba(...)
	if tokens[0].Type == TokenFunction {
		fname := strings.ToLower(tokens[0].Value)
		if fname == "rgb" || fname == "rgba" {
			return parseRGBFunction(tokens)
		}
	}

	return nil, false
}

// parseHexColor parses a hex color string (without the '#').
func parseHexColor(hex string) (*Color, bool) {
	switch len(hex) {
	case 3: // #RGB → #RRGGBB
		r := hexDigit(hex[0])
		g := hexDigit(hex[1])
		b := hexDigit(hex[2])
		if r < 0 || g < 0 || b < 0 {
			return nil, false
		}
		return &Color{
			R: float64(r*17) / 255,
			G: float64(g*17) / 255,
			B: float64(b*17) / 255,
			A: 1.0,
		}, true
	case 4: // #RGBA
		r := hexDigit(hex[0])
		g := hexDigit(hex[1])
		b := hexDigit(hex[2])
		a := hexDigit(hex[3])
		if r < 0 || g < 0 || b < 0 || a < 0 {
			return nil, false
		}
		return &Color{
			R: float64(r*17) / 255,
			G: float64(g*17) / 255,
			B: float64(b*17) / 255,
			A: float64(a*17) / 255,
		}, true
	case 6: // #RRGGBB
		r := hexByte(hex[0], hex[1])
		g := hexByte(hex[2], hex[3])
		b := hexByte(hex[4], hex[5])
		if r < 0 || g < 0 || b < 0 {
			return nil, false
		}
		return &Color{
			R: float64(r) / 255,
			G: float64(g) / 255,
			B: float64(b) / 255,
			A: 1.0,
		}, true
	case 8: // #RRGGBBAA
		r := hexByte(hex[0], hex[1])
		g := hexByte(hex[2], hex[3])
		b := hexByte(hex[4], hex[5])
		a := hexByte(hex[6], hex[7])
		if r < 0 || g < 0 || b < 0 || a < 0 {
			return nil, false
		}
		return &Color{
			R: float64(r) / 255,
			G: float64(g) / 255,
			B: float64(b) / 255,
			A: float64(a) / 255,
		}, true
	}
	return nil, false
}

// parseRGBFunction parses rgb() or rgba() function tokens.
func parseRGBFunction(tokens []Token) (*Color, bool) {
	// Collect numeric values from the token list
	var nums []float64
	for _, tok := range tokens[1:] { // skip function token
		switch tok.Type {
		case TokenNumber:
			nums = append(nums, tok.Num)
		case TokenPercentage:
			nums = append(nums, tok.Num*255/100)
		case TokenWhitespace, TokenComma, TokenCloseParen:
			continue
		case TokenDelim:
			if tok.Value == "/" {
				continue // alpha separator in modern syntax
			}
		}
	}

	switch len(nums) {
	case 3:
		return &Color{
			R: clamp01(nums[0] / 255),
			G: clamp01(nums[1] / 255),
			B: clamp01(nums[2] / 255),
			A: 1.0,
		}, true
	case 4:
		a := nums[3]
		if a > 1 {
			a /= 255 // normalize if given as 0-255
		}
		return &Color{
			R: clamp01(nums[0] / 255),
			G: clamp01(nums[1] / 255),
			B: clamp01(nums[2] / 255),
			A: clamp01(a),
		}, true
	}
	return nil, false
}

func hexToColor(hex uint32, alpha float64) *Color {
	return &Color{
		R: float64((hex>>16)&0xFF) / 255,
		G: float64((hex>>8)&0xFF) / 255,
		B: float64(hex&0xFF) / 255,
		A: alpha,
	}
}

func hexDigit(c byte) int {
	switch {
	case c >= '0' && c <= '9':
		return int(c - '0')
	case c >= 'a' && c <= 'f':
		return int(c-'a') + 10
	case c >= 'A' && c <= 'F':
		return int(c-'A') + 10
	}
	return -1
}

func hexByte(hi, lo byte) int {
	h := hexDigit(hi)
	l := hexDigit(lo)
	if h < 0 || l < 0 {
		return -1
	}
	return h*16 + l
}

func clamp01(v float64) float64 {
	return math.Max(0, math.Min(1, v))
}
