// Package font provides TrueType font parsing, glyph metrics, text
// measurement, line breaking, and font subsetting for PDF generation.
//
// # Font Interface
//
// The [Font] interface defines the contract for fonts used by the document
// model and PDF writer. Currently the only implementation is [TrueTypeFont],
// which supports the full Unicode range including CJK characters.
//
// # Parsing
//
// Use [ParseTrueType] to load a font from raw TrueType (.ttf) binary data:
//
//	data, _ := os.ReadFile("NotoSansJP-Regular.ttf")
//	ttf, err := font.ParseTrueType(data)
//	name := ttf.Name()       // PostScript name
//	m := ttf.Metrics()       // ascender, descender, cap height, etc.
//
// # Text Measurement
//
// [MeasureString] calculates the rendered width of a string at a given font
// size. [LineBreak] splits text into lines that fit a maximum width:
//
//	width := font.MeasureString(ttf, "Hello", 12)
//	lines := font.LineBreak(ttf, "Long text...", 12, 200)
//
// # Subsetting
//
// For PDF embedding, fonts should be subsetted to include only the glyphs
// actually used. Call [TrueTypeFont.Subset] with the set of runes to produce
// a minimal font file:
//
//	subsetData, err := ttf.Subset(usedRunes)
package font
