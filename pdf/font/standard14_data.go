package font

// Advance widths and metrics for the Adobe Core 14 fonts in 1000 em units.
//
// Values transcribed from the Adobe Font Metrics (AFM) files published by
// Adobe as part of the Core Font Program. Coverage is printable ASCII
// (0x20-0x7E); characters outside this range fall back to the space
// glyph's width — which matches how PDF viewers render unmapped runes for
// non-embedded Type1 fonts.

// printableASCIIWidths maps rune positions 0x20..0x7E (space through ~) to
// their advance widths. Helper to expand a compact [95]uint16 into the
// rune-keyed map stored on each font entry.
func printableASCIIWidths(ws [95]uint16) map[rune]int {
	m := make(map[rune]int, 95)
	for i, w := range ws {
		m[rune(0x20+i)] = int(w)
	}
	return m
}

// monospaceWidths returns a 95-entry table where every printable ASCII
// glyph has the same advance width — used for the Courier family.
func monospaceWidths(w uint16) [95]uint16 {
	var arr [95]uint16
	for i := range arr {
		arr[i] = w
	}
	return arr
}

var helveticaASCII = [95]uint16{
	// 0x20..0x27:   ! " # $ % & '
	278, 278, 355, 556, 556, 889, 667, 191,
	// 0x28..0x2F: ( ) * + , - . /
	333, 333, 389, 584, 278, 333, 278, 278,
	// 0x30..0x37: 0 1 2 3 4 5 6 7
	556, 556, 556, 556, 556, 556, 556, 556,
	// 0x38..0x3F: 8 9 : ; < = > ?
	556, 556, 278, 278, 584, 584, 584, 556,
	// 0x40..0x47: @ A B C D E F G
	1015, 667, 667, 722, 722, 667, 611, 778,
	// 0x48..0x4F: H I J K L M N O
	722, 278, 500, 667, 556, 833, 722, 778,
	// 0x50..0x57: P Q R S T U V W
	667, 778, 722, 667, 611, 722, 667, 944,
	// 0x58..0x5F: X Y Z [ \ ] ^ _
	667, 667, 611, 278, 278, 278, 469, 556,
	// 0x60..0x67: ` a b c d e f g
	333, 556, 556, 500, 556, 556, 278, 556,
	// 0x68..0x6F: h i j k l m n o
	556, 222, 222, 500, 222, 833, 556, 556,
	// 0x70..0x77: p q r s t u v w
	556, 556, 333, 500, 278, 556, 500, 722,
	// 0x78..0x7E: x y z { | } ~
	500, 500, 500, 334, 260, 334, 584,
}

var helveticaBoldASCII = [95]uint16{
	278, 333, 474, 556, 556, 889, 722, 238,
	333, 333, 389, 584, 278, 333, 278, 278,
	556, 556, 556, 556, 556, 556, 556, 556,
	556, 556, 333, 333, 584, 584, 584, 611,
	975, 722, 722, 722, 722, 667, 611, 778,
	722, 278, 556, 722, 611, 833, 722, 778,
	667, 778, 722, 667, 611, 722, 667, 944,
	667, 667, 611, 333, 278, 333, 584, 556,
	333, 556, 611, 556, 611, 556, 333, 611,
	611, 278, 278, 556, 278, 889, 611, 611,
	611, 611, 389, 556, 333, 611, 556, 778,
	556, 556, 500, 389, 280, 389, 584,
}

var timesRomanASCII = [95]uint16{
	250, 333, 408, 500, 500, 833, 778, 180,
	333, 333, 500, 564, 250, 333, 250, 278,
	500, 500, 500, 500, 500, 500, 500, 500,
	500, 500, 278, 278, 564, 564, 564, 444,
	921, 722, 667, 667, 722, 611, 556, 722,
	722, 333, 389, 722, 611, 889, 722, 722,
	556, 722, 667, 556, 611, 722, 722, 944,
	722, 722, 611, 333, 278, 333, 469, 500,
	333, 444, 500, 444, 500, 444, 333, 500,
	500, 278, 278, 500, 278, 778, 500, 500,
	500, 500, 333, 389, 278, 500, 500, 722,
	500, 500, 444, 480, 200, 480, 541,
}

var timesBoldASCII = [95]uint16{
	250, 333, 555, 500, 500, 1000, 833, 278,
	333, 333, 500, 570, 250, 333, 250, 278,
	500, 500, 500, 500, 500, 500, 500, 500,
	500, 500, 333, 333, 570, 570, 570, 500,
	930, 722, 667, 722, 722, 667, 611, 778,
	778, 389, 500, 778, 667, 944, 722, 778,
	611, 778, 722, 556, 667, 722, 722, 1000,
	722, 722, 667, 333, 278, 333, 581, 500,
	333, 500, 556, 444, 556, 444, 333, 500,
	556, 278, 333, 556, 278, 833, 556, 500,
	556, 556, 444, 389, 333, 556, 500, 722,
	500, 500, 444, 394, 220, 394, 520,
}

var timesItalicASCII = [95]uint16{
	250, 333, 420, 500, 500, 833, 778, 214,
	333, 333, 500, 675, 250, 333, 250, 278,
	500, 500, 500, 500, 500, 500, 500, 500,
	500, 500, 333, 333, 675, 675, 675, 500,
	920, 611, 611, 667, 722, 611, 611, 722,
	722, 333, 444, 667, 556, 833, 667, 722,
	611, 722, 611, 500, 556, 722, 611, 833,
	611, 556, 556, 389, 278, 389, 422, 500,
	333, 500, 500, 444, 500, 444, 278, 500,
	500, 278, 278, 444, 278, 722, 500, 500,
	500, 500, 389, 389, 278, 500, 444, 667,
	444, 444, 389, 400, 275, 400, 541,
}

var timesBoldItalicASCII = [95]uint16{
	250, 389, 555, 500, 500, 833, 778, 278,
	333, 333, 500, 570, 250, 333, 250, 278,
	500, 500, 500, 500, 500, 500, 500, 500,
	500, 500, 333, 333, 570, 570, 570, 500,
	832, 667, 667, 667, 722, 667, 667, 722,
	778, 389, 500, 667, 611, 889, 722, 722,
	611, 722, 667, 556, 611, 722, 667, 889,
	667, 611, 611, 333, 278, 333, 570, 500,
	333, 500, 500, 444, 500, 444, 333, 500,
	556, 278, 278, 500, 278, 778, 556, 500,
	500, 500, 389, 389, 278, 556, 444, 667,
	500, 444, 389, 348, 220, 348, 570,
}

// Symbol and ZapfDingbats use specialized encodings that don't map to
// printable ASCII in the usual way. For layout measurement we return a
// plausible average width — users generally don't set body text in these
// fonts, and any mismatch only affects alignment within those contexts.
var symbolFallbackWidth = uint16(500)
var zapfDingbatsFallbackWidth = uint16(788)

// standard14Data holds per-font metrics and ASCII width tables.
// Helvetica-Oblique / Helvetica-BoldOblique share widths with their upright
// counterparts (obliquing does not alter advance widths). Courier-* are all
// 600 em units wide (monospace).
var standard14Data = map[string]standard14FontEntry{
	Helvetica: {
		metrics: Metrics{UnitsPerEm: 1000, Ascender: 718, Descender: -207, CapHeight: 718, XHeight: 523, LineGap: 0},
		widths:  printableASCIIWidths(helveticaASCII),
	},
	HelveticaBold: {
		metrics: Metrics{UnitsPerEm: 1000, Ascender: 718, Descender: -207, CapHeight: 718, XHeight: 532, LineGap: 0},
		widths:  printableASCIIWidths(helveticaBoldASCII),
	},
	HelveticaOblique: {
		metrics: Metrics{UnitsPerEm: 1000, Ascender: 718, Descender: -207, CapHeight: 718, XHeight: 523, LineGap: 0, ItalicAngle: -12},
		widths:  printableASCIIWidths(helveticaASCII),
	},
	HelveticaBoldOblique: {
		metrics: Metrics{UnitsPerEm: 1000, Ascender: 718, Descender: -207, CapHeight: 718, XHeight: 532, LineGap: 0, ItalicAngle: -12},
		widths:  printableASCIIWidths(helveticaBoldASCII),
	},
	TimesRoman: {
		metrics: Metrics{UnitsPerEm: 1000, Ascender: 683, Descender: -217, CapHeight: 662, XHeight: 450, LineGap: 0},
		widths:  printableASCIIWidths(timesRomanASCII),
	},
	TimesBold: {
		metrics: Metrics{UnitsPerEm: 1000, Ascender: 683, Descender: -217, CapHeight: 676, XHeight: 461, LineGap: 0},
		widths:  printableASCIIWidths(timesBoldASCII),
	},
	TimesItalic: {
		metrics: Metrics{UnitsPerEm: 1000, Ascender: 683, Descender: -217, CapHeight: 653, XHeight: 441, LineGap: 0, ItalicAngle: -15.5},
		widths:  printableASCIIWidths(timesItalicASCII),
	},
	TimesBoldItalic: {
		metrics: Metrics{UnitsPerEm: 1000, Ascender: 683, Descender: -217, CapHeight: 669, XHeight: 462, LineGap: 0, ItalicAngle: -15},
		widths:  printableASCIIWidths(timesBoldItalicASCII),
	},
	Courier: {
		metrics: Metrics{UnitsPerEm: 1000, Ascender: 629, Descender: -157, CapHeight: 562, XHeight: 426, LineGap: 0},
		widths:  printableASCIIWidths(monospaceWidths(600)),
	},
	CourierBold: {
		metrics: Metrics{UnitsPerEm: 1000, Ascender: 629, Descender: -157, CapHeight: 562, XHeight: 439, LineGap: 0},
		widths:  printableASCIIWidths(monospaceWidths(600)),
	},
	CourierOblique: {
		metrics: Metrics{UnitsPerEm: 1000, Ascender: 629, Descender: -157, CapHeight: 562, XHeight: 426, LineGap: 0, ItalicAngle: -12},
		widths:  printableASCIIWidths(monospaceWidths(600)),
	},
	CourierBoldOblique: {
		metrics: Metrics{UnitsPerEm: 1000, Ascender: 629, Descender: -157, CapHeight: 562, XHeight: 439, LineGap: 0, ItalicAngle: -12},
		widths:  printableASCIIWidths(monospaceWidths(600)),
	},
	Symbol: {
		metrics: Metrics{UnitsPerEm: 1000, Ascender: 1010, Descender: -293, CapHeight: 729, XHeight: 525, LineGap: 0},
		widths:  printableASCIIWidths(monospaceWidths(symbolFallbackWidth)),
	},
	ZapfDingbats: {
		metrics: Metrics{UnitsPerEm: 1000, Ascender: 820, Descender: -143, CapHeight: 820, XHeight: 456, LineGap: 0},
		widths:  printableASCIIWidths(monospaceWidths(zapfDingbatsFallbackWidth)),
	},
}

// Variant aliases: template.buildFontVariantID emits generic "-Italic"
// suffixes regardless of family, but the Adobe Core 14 canonical names use
// "-Oblique" for Helvetica/Courier (and "-Roman" as the base form of
// Times). Map those variants so layout lookups find the same metrics the
// PDF renderer writes out via base14Variants in pdftarget.go.
func init() {
	for alias, canonical := range map[string]string{
		"Helvetica-Italic":     HelveticaOblique,
		"Helvetica-BoldItalic": HelveticaBoldOblique,
		"Courier-Italic":       CourierOblique,
		"Courier-BoldItalic":   CourierBoldOblique,
		"Times":                TimesRoman,
	} {
		if _, exists := standard14Data[alias]; !exists {
			standard14Data[alias] = standard14Data[canonical]
		}
	}
}
