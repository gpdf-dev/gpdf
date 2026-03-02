package css

import (
	"strings"
)

// ExpandShorthands expands shorthand properties into their longhand equivalents.
// e.g., "margin: 10px 20px" → margin-top:10px, margin-right:20px, margin-bottom:10px, margin-left:20px
func ExpandShorthands(decls []Declaration) []Declaration {
	var result []Declaration
	for _, d := range decls {
		expanded := expandSingle(d)
		result = append(result, expanded...)
	}
	return result
}

func expandSingle(d Declaration) []Declaration {
	switch d.Property {
	case "margin":
		return expandEdges(d, "margin-top", "margin-right", "margin-bottom", "margin-left")
	case "padding":
		return expandEdges(d, "padding-top", "padding-right", "padding-bottom", "padding-left")
	case "border":
		return expandBorder(d, []string{"top", "right", "bottom", "left"})
	case "border-top":
		return expandBorderSide(d, "top")
	case "border-right":
		return expandBorderSide(d, "right")
	case "border-bottom":
		return expandBorderSide(d, "bottom")
	case "border-left":
		return expandBorderSide(d, "left")
	case "border-width":
		return expandEdges(d, "border-top-width", "border-right-width", "border-bottom-width", "border-left-width")
	case "border-style":
		return expandEdges(d, "border-top-style", "border-right-style", "border-bottom-style", "border-left-style")
	case "border-color":
		return expandEdges(d, "border-top-color", "border-right-color", "border-bottom-color", "border-left-color")
	case "font":
		return expandFont(d)
	case "background":
		return expandBackground(d)
	case "text-decoration":
		// Map to text-decoration-line (we treat it as text-decoration for simplicity)
		return []Declaration{d}
	default:
		return []Declaration{d}
	}
}

// expandEdges expands a 1-4 value shorthand into four longhand properties.
// 1 value: all sides
// 2 values: vertical horizontal
// 3 values: top horizontal bottom
// 4 values: top right bottom left
func expandEdges(d Declaration, top, right, bottom, left string) []Declaration {
	parts := splitValues(d.Value)
	var t, r, b, l string

	switch len(parts) {
	case 1:
		t, r, b, l = parts[0], parts[0], parts[0], parts[0]
	case 2:
		t, b = parts[0], parts[0]
		r, l = parts[1], parts[1]
	case 3:
		t = parts[0]
		r, l = parts[1], parts[1]
		b = parts[2]
	case 4:
		t, r, b, l = parts[0], parts[1], parts[2], parts[3]
	default:
		return []Declaration{d}
	}

	return []Declaration{
		{Property: top, Value: t, Important: d.Important},
		{Property: right, Value: r, Important: d.Important},
		{Property: bottom, Value: b, Important: d.Important},
		{Property: left, Value: l, Important: d.Important},
	}
}

// expandBorder expands "border: 2px solid #333" into width/style/color for all sides.
func expandBorder(d Declaration, sides []string) []Declaration {
	width, style, color := parseBorderValue(d.Value)
	var result []Declaration
	for _, side := range sides {
		if width != "" {
			result = append(result, Declaration{Property: "border-" + side + "-width", Value: width, Important: d.Important})
		}
		if style != "" {
			result = append(result, Declaration{Property: "border-" + side + "-style", Value: style, Important: d.Important})
		}
		if color != "" {
			result = append(result, Declaration{Property: "border-" + side + "-color", Value: color, Important: d.Important})
		}
	}
	return result
}

// expandBorderSide expands "border-top: 2px solid red" into width/style/color.
func expandBorderSide(d Declaration, side string) []Declaration {
	width, style, color := parseBorderValue(d.Value)
	var result []Declaration
	if width != "" {
		result = append(result, Declaration{Property: "border-" + side + "-width", Value: width, Important: d.Important})
	}
	if style != "" {
		result = append(result, Declaration{Property: "border-" + side + "-style", Value: style, Important: d.Important})
	}
	if color != "" {
		result = append(result, Declaration{Property: "border-" + side + "-color", Value: color, Important: d.Important})
	}
	return result
}

// parseBorderValue parses "2px solid red" into (width, style, color).
func parseBorderValue(value string) (width, style, color string) {
	parts := splitValues(value)
	for _, p := range parts {
		lower := strings.ToLower(p)
		if isBorderStyle(lower) {
			style = lower
		} else if isLengthOrZero(lower) {
			width = p
		} else {
			color = p
		}
	}
	return
}

// expandFont expands the font shorthand (simplified: font-size and font-family required).
func expandFont(d Declaration) []Declaration {
	// Very simplified: just pass through as-is for now
	// Full font shorthand parsing is complex and not critical for Phase 4-A
	parts := splitValues(d.Value)
	var result []Declaration

	for _, p := range parts {
		lower := strings.ToLower(p)
		switch lower {
		case "bold":
			result = append(result, Declaration{Property: "font-weight", Value: "bold", Important: d.Important})
		case "normal":
			// Could be weight or style; skip ambiguity
		case "italic", "oblique":
			result = append(result, Declaration{Property: "font-style", Value: lower, Important: d.Important})
		default:
			if isLengthOrZero(lower) || strings.HasSuffix(lower, "%") {
				// Check for font-size/line-height (e.g., "16px/1.5")
				if idx := strings.Index(p, "/"); idx >= 0 {
					result = append(result, Declaration{Property: "font-size", Value: p[:idx], Important: d.Important})
					result = append(result, Declaration{Property: "line-height", Value: p[idx+1:], Important: d.Important})
				} else {
					result = append(result, Declaration{Property: "font-size", Value: p, Important: d.Important})
				}
			} else {
				// Assume font-family (last value typically)
				result = append(result, Declaration{Property: "font-family", Value: p, Important: d.Important})
			}
		}
	}

	if len(result) == 0 {
		return []Declaration{d}
	}
	return result
}

// expandBackground expands background shorthand (only background-color supported).
func expandBackground(d Declaration) []Declaration {
	// Phase 4-A: only extract background-color
	value := strings.TrimSpace(d.Value)
	// If it looks like a color (hex, named, rgb), use it as background-color
	if _, ok := ParseColor(value); ok {
		return []Declaration{{Property: "background-color", Value: value, Important: d.Important}}
	}
	// Otherwise pass through as background-color
	return []Declaration{{Property: "background-color", Value: value, Important: d.Important}}
}

// splitValues splits a CSS value string into whitespace-separated parts,
// respecting parentheses (e.g., "rgb(1, 2, 3)" stays as one part).
func splitValues(value string) []string {
	value = strings.TrimSpace(value)
	var parts []string
	var current strings.Builder
	depth := 0

	for _, r := range value {
		switch {
		case r == '(':
			depth++
			current.WriteRune(r)
		case r == ')':
			if depth > 0 {
				depth--
			}
			current.WriteRune(r)
		case (r == ' ' || r == '\t' || r == '\n') && depth == 0:
			if current.Len() > 0 {
				parts = append(parts, current.String())
				current.Reset()
			}
		default:
			current.WriteRune(r)
		}
	}
	if current.Len() > 0 {
		parts = append(parts, current.String())
	}
	return parts
}

var borderStyles = map[string]bool{
	"none": true, "solid": true, "dashed": true, "dotted": true,
	"double": true, "groove": true, "ridge": true, "inset": true,
	"outset": true, "hidden": true,
}

func isBorderStyle(s string) bool {
	return borderStyles[s]
}

func isLengthOrZero(s string) bool {
	if s == "0" {
		return true
	}
	// Check for number followed by unit
	units := []string{"px", "pt", "em", "rem", "mm", "cm", "in", "ex", "ch", "vw", "vh"}
	for _, u := range units {
		if strings.HasSuffix(s, u) {
			return true
		}
	}
	return false
}
