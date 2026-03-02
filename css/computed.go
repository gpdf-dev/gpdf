package css

import (
	"strings"
)

// ResolveComputed resolves relative values to absolute values.
// parentFontSize is the parent's computed font-size in pt.
// containerWidth is the containing block width in pt (for percentage resolution).
func ResolveComputed(styles ComputedStyles, parentFontSize, containerWidth float64) ComputedStyles {
	resolved := make(ComputedStyles, len(styles))
	for k, v := range styles {
		resolved[k] = v
	}

	// Resolve font-size first (other em/% values depend on it)
	if fs, ok := resolved["font-size"]; ok {
		resolved["font-size"] = resolveLength(fs, parentFontSize, containerWidth)
	}

	fontSize := parsePt(resolved["font-size"], parentFontSize)

	// Resolve all length/percentage properties
	lengthProps := []string{
		"margin-top", "margin-right", "margin-bottom", "margin-left",
		"padding-top", "padding-right", "padding-bottom", "padding-left",
		"border-top-width", "border-right-width", "border-bottom-width", "border-left-width",
		"width", "height", "min-width", "max-width", "min-height", "max-height",
		"text-indent",
	}

	for _, prop := range lengthProps {
		if val, ok := resolved[prop]; ok {
			resolved[prop] = resolveLength(val, fontSize, containerWidth)
		}
	}

	// Resolve border-spacing
	if val, ok := resolved["border-spacing"]; ok && val != "0" {
		resolved["border-spacing"] = resolveLength(val, fontSize, containerWidth)
	}

	// Resolve line-height
	if lh, ok := resolved["line-height"]; ok {
		resolved["line-height"] = resolveLineHeight(lh, fontSize)
	}

	// Resolve letter-spacing and word-spacing
	for _, prop := range []string{"letter-spacing", "word-spacing"} {
		if val, ok := resolved[prop]; ok && val != "normal" {
			resolved[prop] = resolveLength(val, fontSize, containerWidth)
		}
	}

	// Resolve border widths named values
	for _, side := range []string{"top", "right", "bottom", "left"} {
		prop := "border-" + side + "-width"
		if val, ok := resolved[prop]; ok {
			resolved[prop] = resolveBorderWidth(val)
		}
	}

	// Resolve font-weight keywords
	if fw, ok := resolved["font-weight"]; ok {
		resolved["font-weight"] = resolveFontWeight(fw)
	}

	return resolved
}

// resolveLength converts a CSS length value to a pt string.
func resolveLength(value string, fontSize, containerWidth float64) string {
	value = strings.TrimSpace(value)
	lower := strings.ToLower(value)

	if lower == "auto" || lower == "none" || lower == "normal" {
		return value
	}

	tokens := ParseValue(value)
	if len(tokens) != 1 {
		return value
	}

	tok := tokens[0]
	switch tok.Type {
	case TokenNumber:
		if tok.Num == 0 {
			return "0pt"
		}
		return value

	case TokenDimension:
		pt := dimensionToPt(tok.Num, tok.Unit, fontSize)
		return formatPt(pt)

	case TokenPercentage:
		pt := tok.Num / 100 * containerWidth
		return formatPt(pt)
	}

	return value
}

// resolveLineHeight resolves line-height values.
func resolveLineHeight(value string, fontSize float64) string {
	value = strings.TrimSpace(value)
	if value == "normal" {
		return formatPt(fontSize * 1.2)
	}

	tokens := ParseValue(value)
	if len(tokens) != 1 {
		return value
	}

	tok := tokens[0]
	switch tok.Type {
	case TokenNumber:
		// Unitless number: multiply by font-size
		return formatPt(tok.Num * fontSize)
	case TokenDimension:
		return formatPt(dimensionToPt(tok.Num, tok.Unit, fontSize))
	case TokenPercentage:
		return formatPt(tok.Num / 100 * fontSize)
	}

	return value
}

// resolveBorderWidth maps named border widths to pt values.
func resolveBorderWidth(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "thin":
		return "0.75pt"
	case "medium":
		return "1.5pt"
	case "thick":
		return "3pt"
	case "0", "none":
		return "0pt"
	}
	// Try to resolve as a length
	tokens := ParseValue(value)
	if len(tokens) == 1 && tokens[0].Type == TokenDimension {
		return formatPt(dimensionToPt(tokens[0].Num, tokens[0].Unit, 12))
	}
	return value
}

// resolveFontWeight normalizes font-weight keywords to numeric values.
func resolveFontWeight(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "normal":
		return "400"
	case "bold":
		return "700"
	case "lighter":
		return "100"
	case "bolder":
		return "900"
	}
	return value
}

// dimensionToPt converts a numeric value with a CSS unit to points.
func dimensionToPt(num float64, unit string, fontSize float64) float64 {
	switch strings.ToLower(unit) {
	case "pt":
		return num
	case "px":
		return num * 0.75 // 1px = 0.75pt (96dpi vs 72dpi)
	case "em":
		return num * fontSize
	case "rem":
		return num * 12 // assume root font-size 12pt
	case "mm":
		return num * 72 / 25.4
	case "cm":
		return num * 72 / 2.54
	case "in":
		return num * 72
	case "pc":
		return num * 12 // 1 pica = 12pt
	case "ex":
		return num * fontSize * 0.5 // approximate
	case "ch":
		return num * fontSize * 0.5 // approximate
	default:
		return num
	}
}

// parsePt extracts a numeric pt value from a resolved string.
func parsePt(value string, fallback float64) float64 {
	value = strings.TrimSpace(value)
	tokens := ParseValue(value)
	if len(tokens) == 1 {
		switch tokens[0].Type {
		case TokenDimension:
			if strings.ToLower(tokens[0].Unit) == "pt" {
				return tokens[0].Num
			}
			return dimensionToPt(tokens[0].Num, tokens[0].Unit, fallback)
		case TokenNumber:
			return tokens[0].Num
		}
	}
	return fallback
}

// formatPt formats a float64 as a pt value string.
func formatPt(pt float64) string {
	s := formatFloat(pt)
	// Only trim trailing zeros after a decimal point
	if strings.Contains(s, ".") {
		s = strings.TrimRight(s, "0")
		s = strings.TrimRight(s, ".")
	}
	return s + "pt"
}

// formatFloat formats a float64 without importing fmt.
func formatFloat(f float64) string {
	if f < 0 {
		return "-" + formatFloat(-f)
	}
	intPart := int(f)
	fracPart := f - float64(intPart)

	s := formatInt(intPart)
	if fracPart > 0.0001 {
		s += "."
		// Up to 4 decimal places
		for range 4 {
			fracPart *= 10
			digit := int(fracPart)
			s += string(rune('0' + digit))
			fracPart -= float64(digit)
			if fracPart < 0.0001 {
				break
			}
		}
	}
	return s
}

func formatInt(n int) string {
	if n == 0 {
		return "0"
	}
	s := ""
	for n > 0 {
		s = string(rune('0'+n%10)) + s
		n /= 10
	}
	return s
}

// ParsePt is a public wrapper for parsePt.
func ParsePt(value string, fallback float64) float64 {
	return parsePt(value, fallback)
}

// DimensionToPt is a public wrapper for dimensionToPt.
func DimensionToPt(num float64, unit string, fontSize float64) float64 {
	return dimensionToPt(num, unit, fontSize)
}
