package template

import (
	"strings"

	"github.com/gpdf-dev/gpdf/document"
	"github.com/gpdf-dev/gpdf/document/layout"
	"github.com/gpdf-dev/gpdf/pdf/font"
)

// builtinFontResolver implements layout.FontResolver using registered
// TrueType fonts. For unregistered font families it falls back to
// approximate metrics so that layout can proceed without embedded fonts.
type builtinFontResolver struct {
	fonts map[string]*font.TrueTypeFont
}

// newBuiltinFontResolver creates a FontResolver backed by the given fonts.
func newBuiltinFontResolver(fonts map[string]*font.TrueTypeFont) *builtinFontResolver {
	return &builtinFontResolver{fonts: fonts}
}

// Resolve returns a ResolvedFont for the requested family and style.
// It tries variant-specific fonts first (e.g. "MyFont-Bold"), then
// falls back to the base family for metrics.
func (r *builtinFontResolver) Resolve(family string, weight document.FontWeight, italic bool) layout.ResolvedFont {
	bold := weight >= document.WeightBold
	// Empty family means "the PDF renderer default" — which is Helvetica
	// (see document/render/pdftarget.go:resolvePDFFontName). Mirror that
	// here so layout uses the same metrics the viewer will.
	if family == "" {
		family = font.Helvetica
	}
	variantID := buildFontVariantID(family, bold, italic)

	// Try variant-specific font first, then fall back to base family.
	ttf, ok := r.fonts[variantID]
	if !ok {
		ttf, ok = r.fonts[family]
	}

	if !ok {
		// If the requested font is one of the Adobe 14 standard fonts, use
		// its AFM-derived metrics so that layout matches what PDF viewers
		// render from non-embedded Type1 entries.
		if m, std := font.Standard14Metrics(variantID); std {
			scale := 1.0 / float64(m.UnitsPerEm)
			return layout.ResolvedFont{
				ID: variantID,
				Metrics: layout.FontMetrics{
					Ascender:   float64(m.Ascender) * scale,
					Descender:  float64(m.Descender) * scale,
					LineHeight: float64(m.Ascender-m.Descender+m.LineGap) * scale,
					CapHeight:  float64(m.CapHeight) * scale,
				},
			}
		}
		// Return approximate metrics for non-standard unregistered fonts.
		return layout.ResolvedFont{
			ID: variantID,
			Metrics: layout.FontMetrics{
				Ascender:   0.8,
				Descender:  -0.2,
				LineHeight: 1.2,
				CapHeight:  0.7,
			},
		}
	}

	m := ttf.Metrics()
	scale := 1.0 / float64(m.UnitsPerEm)
	return layout.ResolvedFont{
		ID: ttf.Name(),
		Metrics: layout.FontMetrics{
			Ascender:   float64(m.Ascender) * scale,
			Descender:  float64(m.Descender) * scale,
			LineHeight: float64(m.Ascender-m.Descender+m.LineGap) * scale,
			CapHeight:  float64(m.CapHeight) * scale,
		},
	}
}

// buildFontVariantID appends bold/italic suffixes to a font family name.
func buildFontVariantID(family string, bold, italic bool) string {
	switch {
	case bold && italic:
		return family + "-BoldItalic"
	case bold:
		return family + "-Bold"
	case italic:
		return family + "-Italic"
	default:
		return family
	}
}

// MeasureString measures the width of text at the given font size.
func (r *builtinFontResolver) MeasureString(f layout.ResolvedFont, text string, size float64) float64 {
	if ttf, ok := r.fonts[f.ID]; ok {
		return font.MeasureString(ttf, text, size)
	}
	// Non-embedded Adobe 14 fonts: use AFM widths so layout matches viewer.
	if std, ok := font.NewStandard14Font(f.ID); ok {
		return font.MeasureString(std, text, size)
	}
	// Fallback to approximate: average char width ≈ 0.5 * fontSize.
	return float64(len([]rune(text))) * size * 0.5
}

// LineBreak splits text into lines fitting within maxWidth.
func (r *builtinFontResolver) LineBreak(f layout.ResolvedFont, text string, size float64, maxWidth float64) []string {
	if ttf, ok := r.fonts[f.ID]; ok {
		return font.LineBreak(ttf, text, size, maxWidth)
	}
	if std, ok := font.NewStandard14Font(f.ID); ok {
		return font.LineBreak(std, text, size, maxWidth)
	}
	return approximateBreak(text, size, maxWidth)
}

// approximateBreak performs rough line breaking without font metrics.
func approximateBreak(text string, fontSize, maxWidth float64) []string {
	avgCharWidth := fontSize * 0.5
	if avgCharWidth <= 0 {
		return []string{text}
	}
	charsPerLine := int(maxWidth / avgCharWidth)
	if charsPerLine <= 0 {
		charsPerLine = 1
	}

	var lines []string
	for _, paragraph := range strings.Split(text, "\n") {
		words := strings.Fields(paragraph)
		if len(words) == 0 {
			lines = append(lines, "")
			continue
		}

		current := words[0]
		for _, word := range words[1:] {
			if runeLen(current)+1+runeLen(word) > charsPerLine {
				lines = append(lines, current)
				current = word
			} else {
				current += " " + word
			}
		}
		lines = append(lines, current)
	}
	return lines
}

func runeLen(s string) int {
	n := 0
	for range s {
		n++
	}
	return n
}
