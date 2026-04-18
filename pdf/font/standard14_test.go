package font

import "testing"

// Widths are checked against values published in the Adobe Core 14 AFM
// files. Spot-checking across letters with distinct widths (narrow 'I',
// wide 'M', digits, punctuation) verifies the table transcription.
func TestStandard14Width_Helvetica(t *testing.T) {
	cases := []struct {
		r    rune
		want int
	}{
		{' ', 278}, {'I', 278}, {'M', 833}, {'W', 944},
		{'i', 222}, {'m', 833}, {'w', 722},
		{'0', 556}, {'1', 556}, {':', 278}, {',', 278}, {'-', 333},
	}
	for _, c := range cases {
		got, ok := Standard14Width(Helvetica, c.r)
		if !ok {
			t.Fatalf("Standard14Width(Helvetica, %q) not found", c.r)
		}
		if got != c.want {
			t.Errorf("Helvetica %q = %d, want %d", c.r, got, c.want)
		}
	}
}

func TestStandard14Width_HelveticaBold(t *testing.T) {
	// Spot-check individual glyph widths against Adobe AFM values.
	for _, c := range []struct {
		r    rune
		want int
	}{
		{'I', 278}, {'N', 722}, {'V', 667}, {'O', 778}, {'C', 722}, {'E', 667},
	} {
		got, _ := Standard14Width(HelveticaBold, c.r)
		if got != c.want {
			t.Errorf("HelveticaBold %q = %d, want %d", c.r, got, c.want)
		}
	}
	// Sum for all 7 glyphs in "INVOICE" must equal 4112 em units —
	// at 28pt that is 115.136pt, matching pdftotext measurement of
	// the INVOICE glyph run in golden 18_invoice.pdf.
	total := 0
	for _, r := range "INVOICE" {
		w, _ := Standard14Width(HelveticaBold, r)
		total += w
	}
	if total != 4112 {
		t.Errorf("sum of INVOICE glyphs = %d em units, want 4112", total)
	}
	wantAt28 := float64(total) * 28 / 1000
	if wantAt28 < 115.13 || wantAt28 > 115.14 {
		t.Errorf("INVOICE@28pt HelveticaBold = %.3f pt, want 115.136", wantAt28)
	}
}

func TestStandard14Width_CourierMonospace(t *testing.T) {
	// Every printable ASCII glyph must be 600 em units in Courier.
	for r := rune(0x20); r <= 0x7E; r++ {
		got, ok := Standard14Width(Courier, r)
		if !ok {
			t.Fatalf("Courier width missing for %q", r)
		}
		if got != 600 {
			t.Errorf("Courier %q = %d, want 600", r, got)
		}
	}
}

func TestStandard14Width_UnknownFontReturnsFalse(t *testing.T) {
	if _, ok := Standard14Width("NotAFont", 'A'); ok {
		t.Error("Standard14Width should return false for unknown font")
	}
}

func TestStandard14Width_UnmappedRuneFallsBackToSpace(t *testing.T) {
	want, _ := Standard14Width(Helvetica, ' ')
	got, ok := Standard14Width(Helvetica, '\u3042') // Japanese あ — outside ASCII
	if !ok {
		t.Fatal("expected fallback ok=true for unmapped rune")
	}
	if got != want {
		t.Errorf("unmapped rune width = %d, want space width %d", got, want)
	}
}

func TestIsStandard14(t *testing.T) {
	for _, name := range []string{
		Helvetica, HelveticaBold, HelveticaOblique, HelveticaBoldOblique,
		TimesRoman, TimesBold, TimesItalic, TimesBoldItalic,
		Courier, CourierBold, CourierOblique, CourierBoldOblique,
		Symbol, ZapfDingbats,
	} {
		if !IsStandard14(name) {
			t.Errorf("IsStandard14(%q) = false, want true", name)
		}
	}
	if IsStandard14("RandomFont") {
		t.Error("IsStandard14(\"RandomFont\") should be false")
	}
}

func TestStandard14Metrics_Helvetica(t *testing.T) {
	m, ok := Standard14Metrics(Helvetica)
	if !ok {
		t.Fatal("Helvetica metrics missing")
	}
	if m.UnitsPerEm != 1000 || m.Ascender != 718 || m.Descender != -207 {
		t.Errorf("Helvetica metrics %+v mismatch", m)
	}
}

func TestNewStandard14Font_MeasureStringMatchesAFM(t *testing.T) {
	// Exact values from pdftotext -bbox-layout applied to a PDF rendered
	// by a conformant viewer against Helvetica's AFM widths. If these drift,
	// right-alignment will no longer match visible text extents.
	cases := []struct {
		fontName string
		text     string
		size     float64
		want     float64
	}{
		{HelveticaBold, "INVOICE", 28, 115.136},
		{Helvetica, "Date: March 1, 2026", 12, 108.72},
		{Helvetica, "Due: March 31, 2026", 12, 112.056},
	}
	for _, c := range cases {
		f, ok := NewStandard14Font(c.fontName)
		if !ok {
			t.Fatalf("NewStandard14Font(%q) missing", c.fontName)
		}
		got := MeasureString(f, c.text, c.size)
		// Allow 0.01pt tolerance for floating-point arithmetic.
		if got < c.want-0.01 || got > c.want+0.01 {
			t.Errorf("MeasureString(%s, %q, %v) = %.3f, want %.3f",
				c.fontName, c.text, c.size, got, c.want)
		}
	}
}
