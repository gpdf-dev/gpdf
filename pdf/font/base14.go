package font

// Base14Font represents one of the 14 standard PDF fonts with built-in
// glyph width metrics. These fonts are guaranteed to be available in all
// conforming PDF viewers without embedding.
type Base14Font struct {
	name       string
	widths     map[rune]int
	unitsPerEm int
	ascender   int
	descender  int
	capHeight  int
}

func (f *Base14Font) Name() string { return f.name }
func (f *Base14Font) Metrics() Metrics {
	return Metrics{
		UnitsPerEm: f.unitsPerEm,
		Ascender:   f.ascender,
		Descender:  f.descender,
		CapHeight:  f.capHeight,
	}
}

func (f *Base14Font) GlyphWidth(r rune) (int, bool) {
	w, ok := f.widths[r]
	return w, ok
}

func (f *Base14Font) Encode(text string) []byte       { return []byte(text) }
func (f *Base14Font) Subset(_ []rune) ([]byte, error) { return nil, nil }

// Helvetica returns the standard Helvetica font with AFM-derived metrics.
func Helvetica() *Base14Font {
	return &Base14Font{
		name:       "Helvetica",
		unitsPerEm: 1000,
		ascender:   718,
		descender:  -207,
		capHeight:  718,
		widths:     helveticaWidths,
	}
}

// HelveticaBold returns the standard Helvetica-Bold font with AFM-derived metrics.
func HelveticaBold() *Base14Font {
	return &Base14Font{
		name:       "Helvetica-Bold",
		unitsPerEm: 1000,
		ascender:   718,
		descender:  -207,
		capHeight:  718,
		widths:     helveticaBoldWidths,
	}
}

// HelveticaOblique returns the standard Helvetica-Oblique font with AFM-derived metrics.
func HelveticaOblique() *Base14Font {
	// Helvetica-Oblique has the same widths as Helvetica Regular.
	return &Base14Font{
		name:       "Helvetica-Oblique",
		unitsPerEm: 1000,
		ascender:   718,
		descender:  -207,
		capHeight:  718,
		widths:     helveticaWidths,
	}
}

// HelveticaBoldOblique returns the standard Helvetica-BoldOblique font with AFM-derived metrics.
func HelveticaBoldOblique() *Base14Font {
	// Helvetica-BoldOblique has the same widths as Helvetica-Bold.
	return &Base14Font{
		name:       "Helvetica-BoldOblique",
		unitsPerEm: 1000,
		ascender:   718,
		descender:  -207,
		capHeight:  718,
		widths:     helveticaBoldWidths,
	}
}

// helveticaWidths contains glyph advance widths for Helvetica Regular,
// sourced from the standard Adobe AFM file. Units are per 1000 em.
var helveticaWidths = map[rune]int{
	' ': 278, '!': 278, '"': 355, '#': 556, '$': 556,
	'%': 889, '&': 667, '\'': 191, '(': 333, ')': 333,
	'*': 389, '+': 584, ',': 278, '-': 333, '.': 278, '/': 278,
	'0': 556, '1': 556, '2': 556, '3': 556, '4': 556,
	'5': 556, '6': 556, '7': 556, '8': 556, '9': 556,
	':': 278, ';': 278, '<': 584, '=': 584, '>': 584, '?': 556, '@': 1015,
	'A': 667, 'B': 667, 'C': 722, 'D': 722, 'E': 611,
	'F': 556, 'G': 778, 'H': 722, 'I': 278, 'J': 500,
	'K': 667, 'L': 556, 'M': 833, 'N': 722, 'O': 778,
	'P': 667, 'Q': 778, 'R': 722, 'S': 667, 'T': 611,
	'U': 722, 'V': 667, 'W': 944, 'X': 667, 'Y': 667, 'Z': 611,
	'[': 278, '\\': 278, ']': 278, '^': 469, '_': 556, '`': 333,
	'a': 556, 'b': 556, 'c': 500, 'd': 556, 'e': 556,
	'f': 278, 'g': 556, 'h': 556, 'i': 222, 'j': 222,
	'k': 500, 'l': 222, 'm': 833, 'n': 556, 'o': 556,
	'p': 556, 'q': 556, 'r': 333, 's': 500, 't': 278,
	'u': 556, 'v': 500, 'w': 722, 'x': 500, 'y': 500, 'z': 500,
	'{': 334, '|': 260, '}': 334, '~': 584,
	// Latin-1 Supplement (U+00A0–U+00FF)
	'\u00A0': 278, '\u00A1': 333, '\u00A2': 556, '\u00A3': 556,
	'\u00A4': 556, '\u00A5': 556, '\u00A6': 260, '\u00A7': 556,
	'\u00A8': 333, '\u00A9': 737, '\u00AA': 370, '\u00AB': 556,
	'\u00AC': 584, '\u00AD': 333, '\u00AE': 737, '\u00AF': 333,
	'\u00B0': 400, '\u00B1': 584, '\u00B2': 333, '\u00B3': 333,
	'\u00B4': 333, '\u00B5': 556, '\u00B6': 537, '\u00B7': 278,
	'\u00B8': 333, '\u00B9': 333, '\u00BA': 365, '\u00BB': 556,
	'\u00BC': 834, '\u00BD': 834, '\u00BE': 834, '\u00BF': 611,
	'\u00C0': 667, '\u00C1': 667, '\u00C2': 667, '\u00C3': 667,
	'\u00C4': 667, '\u00C5': 667, '\u00C6': 1000, '\u00C7': 722,
	'\u00C8': 611, '\u00C9': 611, '\u00CA': 611, '\u00CB': 611,
	'\u00CC': 278, '\u00CD': 278, '\u00CE': 278, '\u00CF': 278,
	'\u00D0': 722, '\u00D1': 722, '\u00D2': 778, '\u00D3': 778,
	'\u00D4': 778, '\u00D5': 778, '\u00D6': 778, '\u00D7': 584,
	'\u00D8': 778, '\u00D9': 722, '\u00DA': 722, '\u00DB': 722,
	'\u00DC': 722, '\u00DD': 667, '\u00DE': 667, '\u00DF': 611,
	'\u00E0': 556, '\u00E1': 556, '\u00E2': 556, '\u00E3': 556,
	'\u00E4': 556, '\u00E5': 556, '\u00E6': 889, '\u00E7': 500,
	'\u00E8': 556, '\u00E9': 556, '\u00EA': 556, '\u00EB': 556,
	'\u00EC': 278, '\u00ED': 278, '\u00EE': 278, '\u00EF': 278,
	'\u00F0': 556, '\u00F1': 556, '\u00F2': 556, '\u00F3': 556,
	'\u00F4': 556, '\u00F5': 556, '\u00F6': 556, '\u00F7': 584,
	'\u00F8': 611, '\u00F9': 556, '\u00FA': 556, '\u00FB': 556,
	'\u00FC': 556, '\u00FD': 500, '\u00FE': 556, '\u00FF': 500,
	// WinAnsiEncoding specials (0x80–0x9F range)
	'\u20AC': 556,  // € Euro sign
	'\u201A': 222,  // ‚
	'\u0192': 556,  // ƒ
	'\u201E': 333,  // „
	'\u2026': 1000, // …
	'\u2020': 556,  // †
	'\u2021': 556,  // ‡
	'\u02C6': 333,  // ˆ
	'\u2030': 1000, // ‰
	'\u0160': 667,  // Š
	'\u2039': 333,  // ‹
	'\u0152': 1000, // Œ
	'\u017D': 611,  // Ž
	'\u2018': 222,  // '
	'\u2019': 222,  // '
	'\u201C': 333,  // "
	'\u201D': 333,  // "
	'\u2022': 350,  // •
	'\u2013': 556,  // –
	'\u2014': 1000, // —
	'\u02DC': 333,  // ˜
	'\u2122': 1000, // ™
	'\u0161': 500,  // š
	'\u203A': 333,  // ›
	'\u0153': 944,  // œ
	'\u017E': 500,  // ž
	'\u0178': 667,  // Ÿ
}

// helveticaBoldWidths contains glyph advance widths for Helvetica-Bold,
// sourced from the standard Adobe AFM file. Units are per 1000 em.
var helveticaBoldWidths = map[rune]int{
	' ': 278, '!': 333, '"': 474, '#': 556, '$': 556,
	'%': 889, '&': 722, '\'': 238, '(': 333, ')': 333,
	'*': 389, '+': 584, ',': 278, '-': 333, '.': 278, '/': 278,
	'0': 556, '1': 556, '2': 556, '3': 556, '4': 556,
	'5': 556, '6': 556, '7': 556, '8': 556, '9': 556,
	':': 333, ';': 333, '<': 584, '=': 584, '>': 584, '?': 611, '@': 975,
	'A': 722, 'B': 722, 'C': 722, 'D': 722, 'E': 667,
	'F': 611, 'G': 778, 'H': 722, 'I': 278, 'J': 556,
	'K': 722, 'L': 611, 'M': 833, 'N': 722, 'O': 778,
	'P': 667, 'Q': 778, 'R': 722, 'S': 667, 'T': 611,
	'U': 722, 'V': 667, 'W': 944, 'X': 667, 'Y': 667, 'Z': 611,
	'[': 333, '\\': 278, ']': 333, '^': 584, '_': 556, '`': 333,
	'a': 556, 'b': 611, 'c': 556, 'd': 611, 'e': 556,
	'f': 333, 'g': 611, 'h': 611, 'i': 278, 'j': 278,
	'k': 556, 'l': 278, 'm': 889, 'n': 611, 'o': 611,
	'p': 611, 'q': 611, 'r': 389, 's': 556, 't': 333,
	'u': 611, 'v': 556, 'w': 778, 'x': 556, 'y': 556, 'z': 500,
	'{': 389, '|': 280, '}': 389, '~': 584,
	// Latin-1 Supplement (U+00A0–U+00FF)
	'\u00A0': 278, '\u00A1': 333, '\u00A2': 556, '\u00A3': 556,
	'\u00A4': 556, '\u00A5': 556, '\u00A6': 280, '\u00A7': 556,
	'\u00A8': 333, '\u00A9': 737, '\u00AA': 370, '\u00AB': 556,
	'\u00AC': 584, '\u00AD': 333, '\u00AE': 737, '\u00AF': 333,
	'\u00B0': 400, '\u00B1': 584, '\u00B2': 333, '\u00B3': 333,
	'\u00B4': 333, '\u00B5': 611, '\u00B6': 556, '\u00B7': 278,
	'\u00B8': 333, '\u00B9': 333, '\u00BA': 365, '\u00BB': 556,
	'\u00BC': 834, '\u00BD': 834, '\u00BE': 834, '\u00BF': 611,
	'\u00C0': 722, '\u00C1': 722, '\u00C2': 722, '\u00C3': 722,
	'\u00C4': 722, '\u00C5': 722, '\u00C6': 1000, '\u00C7': 722,
	'\u00C8': 667, '\u00C9': 667, '\u00CA': 667, '\u00CB': 667,
	'\u00CC': 278, '\u00CD': 278, '\u00CE': 278, '\u00CF': 278,
	'\u00D0': 722, '\u00D1': 722, '\u00D2': 778, '\u00D3': 778,
	'\u00D4': 778, '\u00D5': 778, '\u00D6': 778, '\u00D7': 584,
	'\u00D8': 778, '\u00D9': 722, '\u00DA': 722, '\u00DB': 722,
	'\u00DC': 722, '\u00DD': 667, '\u00DE': 667, '\u00DF': 611,
	'\u00E0': 556, '\u00E1': 556, '\u00E2': 556, '\u00E3': 556,
	'\u00E4': 556, '\u00E5': 556, '\u00E6': 889, '\u00E7': 556,
	'\u00E8': 556, '\u00E9': 556, '\u00EA': 556, '\u00EB': 556,
	'\u00EC': 278, '\u00ED': 278, '\u00EE': 278, '\u00EF': 278,
	'\u00F0': 611, '\u00F1': 611, '\u00F2': 611, '\u00F3': 611,
	'\u00F4': 611, '\u00F5': 611, '\u00F6': 611, '\u00F7': 584,
	'\u00F8': 611, '\u00F9': 611, '\u00FA': 611, '\u00FB': 611,
	'\u00FC': 611, '\u00FD': 556, '\u00FE': 611, '\u00FF': 556,
	// WinAnsiEncoding specials (0x80–0x9F range)
	'\u20AC': 556,  // € Euro sign
	'\u201A': 278,  // ‚
	'\u0192': 556,  // ƒ
	'\u201E': 500,  // „
	'\u2026': 1000, // …
	'\u2020': 556,  // †
	'\u2021': 556,  // ‡
	'\u02C6': 333,  // ˆ
	'\u2030': 1000, // ‰
	'\u0160': 667,  // Š
	'\u2039': 333,  // ‹
	'\u0152': 1000, // Œ
	'\u017D': 611,  // Ž
	'\u2018': 278,  // '
	'\u2019': 278,  // '
	'\u201C': 500,  // "
	'\u201D': 500,  // "
	'\u2022': 350,  // •
	'\u2013': 556,  // –
	'\u2014': 1000, // —
	'\u02DC': 333,  // ˜
	'\u2122': 1000, // ™
	'\u0161': 556,  // š
	'\u203A': 333,  // ›
	'\u0153': 944,  // œ
	'\u017E': 500,  // ž
	'\u0178': 667,  // Ÿ
}
