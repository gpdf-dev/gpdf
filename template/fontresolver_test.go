package template

import (
	"math"
	"testing"

	"github.com/gpdf-dev/gpdf/document"
	"github.com/gpdf-dev/gpdf/pdf/font"
)

// Without any TTF registered, MeasureString must still return AFM-accurate
// widths for Adobe 14 font names so that right/center alignment lines up
// with the glyphs PDF viewers actually draw.
func TestFontResolver_MeasureString_Standard14Fallback(t *testing.T) {
	r := newBuiltinFontResolver(nil)

	cases := []struct {
		family string
		weight document.FontWeight
		italic bool
		text   string
		size   float64
		want   float64
	}{
		{"Helvetica", document.WeightBold, false, "INVOICE", 28, 115.136},
		{"Helvetica", document.WeightNormal, false, "Date: March 1, 2026", 12, 108.72},
		{"Helvetica", document.WeightNormal, false, "Due: March 31, 2026", 12, 112.056},
	}
	for _, c := range cases {
		f := r.Resolve(c.family, c.weight, c.italic)
		got := r.MeasureString(f, c.text, c.size)
		if math.Abs(got-c.want) > 0.01 {
			t.Errorf("MeasureString(%s, weight=%v, %q, %v) = %.3f, want %.3f",
				c.family, c.weight, c.text, c.size, got, c.want)
		}
	}
}

// TTF-registered fonts continue to use the TTF measurements — Standard14
// fallback must not override a registered font.
func TestFontResolver_MeasureString_UnknownFontFallsBackToApproximation(t *testing.T) {
	r := newBuiltinFontResolver(nil)
	f := r.Resolve("NonStandardFont", document.WeightNormal, false)

	got := r.MeasureString(f, "hello", 12)
	want := float64(len("hello")) * 12 * 0.5
	if got != want {
		t.Errorf("unknown font fallback = %.3f, want %.3f (approximation)", got, want)
	}
}

// Resolve should return AFM metrics for Standard 14 font names so line
// heights match the viewer's rendering, not the 0.8/-0.2/1.2/0.7 fallback.
func TestFontResolver_Resolve_Standard14Metrics(t *testing.T) {
	r := newBuiltinFontResolver(nil)
	f := r.Resolve("Helvetica", document.WeightNormal, false)

	if f.ID != font.Helvetica {
		t.Fatalf("resolved font ID = %q, want %q", f.ID, font.Helvetica)
	}
	// Adobe Helvetica AFM: Ascender 718, Descender -207 in 1000 em units.
	// Scaled to em-relative units (divide by 1000).
	if math.Abs(f.Metrics.Ascender-0.718) > 0.001 {
		t.Errorf("Ascender = %.4f, want 0.718", f.Metrics.Ascender)
	}
	if math.Abs(f.Metrics.Descender-(-0.207)) > 0.001 {
		t.Errorf("Descender = %.4f, want -0.207", f.Metrics.Descender)
	}
}

// Italic Helvetica must resolve to the same metrics as Helvetica-Oblique
// (buildFontVariantID emits "-Italic"; Standard14 data aliases it).
func TestFontResolver_Resolve_HelveticaItalicAliasedToOblique(t *testing.T) {
	r := newBuiltinFontResolver(nil)
	f := r.Resolve("Helvetica", document.WeightNormal, true)

	// The ID reflects the generic variant name used internally, but
	// metrics must come from the AFM table via the alias.
	if f.Metrics.Ascender == 0.8 {
		t.Error("italic Helvetica returned the 0.8 fallback Ascender instead of AFM-derived value")
	}
	got := r.MeasureString(f, "INVOICE", 28)
	// Helvetica-Oblique shares widths with Helvetica (obliquing preserves
	// advance widths): I+N+V+O+I+C+E = 278+722+667+778+278+722+667 = 4112
	// em units; at 28pt → 115.136pt.
	if math.Abs(got-115.136) > 0.01 {
		t.Errorf("italic Helvetica INVOICE@28 = %.3f, want 115.136 (alias miss?)", got)
	}
}
