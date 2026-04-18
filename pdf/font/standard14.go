package font

// Adobe Standard 14 font names. PDF viewers render these without any
// embedded font data, using their own built-in metrics. gpdf writes text
// with these fonts as Type1 font entries (see pdf/writer.go:RegisterFont);
// layout measurement must match the AFM widths that viewers use so that
// right/center alignment lands at the same coordinate the PDF viewer
// actually draws the glyphs at.
const (
	Helvetica            = "Helvetica"
	HelveticaBold        = "Helvetica-Bold"
	HelveticaOblique     = "Helvetica-Oblique"
	HelveticaBoldOblique = "Helvetica-BoldOblique"
	TimesRoman           = "Times-Roman"
	TimesBold            = "Times-Bold"
	TimesItalic          = "Times-Italic"
	TimesBoldItalic      = "Times-BoldItalic"
	Courier              = "Courier"
	CourierBold          = "Courier-Bold"
	CourierOblique       = "Courier-Oblique"
	CourierBoldOblique   = "Courier-BoldOblique"
	Symbol               = "Symbol"
	ZapfDingbats         = "ZapfDingbats"
)

// Standard14EmUnits is the design unit size for all Adobe 14 fonts.
const Standard14EmUnits = 1000

// IsStandard14 reports whether name is one of the Adobe 14 standard fonts.
func IsStandard14(name string) bool {
	_, ok := standard14Data[name]
	return ok
}

// Standard14Metrics returns design-unit metrics for a Standard 14 font,
// or zero Metrics and false if the name is not a Standard 14 font.
func Standard14Metrics(name string) (Metrics, bool) {
	d, ok := standard14Data[name]
	if !ok {
		return Metrics{}, false
	}
	return d.metrics, true
}

// Standard14Width returns the advance width of r in the named font, in
// design units (1000 per em). Returns the space glyph's width when r is
// not present in the font's width table, matching the fallback most PDF
// viewers apply when asked to measure characters outside the font's
// encoding.
func Standard14Width(name string, r rune) (int, bool) {
	d, ok := standard14Data[name]
	if !ok {
		return 0, false
	}
	if w, ok := d.widths[r]; ok {
		return w, true
	}
	return d.widths[' '], true
}

// standard14FontEntry carries the metrics and per-rune advance widths for
// one Adobe 14 font.
type standard14FontEntry struct {
	metrics Metrics
	widths  map[rune]int
}

// standard14Font implements Font for layout measurement of Adobe 14 fonts
// for which no TTF is registered. Encode/Subset are not meaningful for
// non-embedded Type1 fonts and return empty results.
type standard14Font struct {
	name    string
	metrics Metrics
	widths  map[rune]int
}

// NewStandard14Font returns a Font backed by the AFM width tables for the
// named Adobe 14 font. Returns (nil, false) if name is not a Standard 14
// font. The returned Font is safe for layout measurement (Metrics and
// GlyphWidth) but Encode returns the raw UTF-8 bytes and Subset is a
// no-op — non-embedded Type1 fonts do not require subsetting.
func NewStandard14Font(name string) (Font, bool) {
	d, ok := standard14Data[name]
	if !ok {
		return nil, false
	}
	return &standard14Font{name: name, metrics: d.metrics, widths: d.widths}, true
}

func (f *standard14Font) Name() string     { return f.name }
func (f *standard14Font) Metrics() Metrics { return f.metrics }

func (f *standard14Font) GlyphWidth(r rune) (int, bool) {
	if w, ok := f.widths[r]; ok {
		return w, true
	}
	return 0, false
}

func (f *standard14Font) Encode(text string) []byte {
	return []byte(text)
}

func (f *standard14Font) Subset(_ []rune) ([]byte, error) {
	return nil, nil
}
