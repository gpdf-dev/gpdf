package font

import "unicode"

// Direction is the base (paragraph) direction passed to a [Shaper].
type Direction int

const (
	// DirectionAuto lets the shaper pick the base direction from the first
	// strong character of the text (Unicode Bidi rules P2 and P3).
	DirectionAuto Direction = iota
	// DirectionLTR forces a left-to-right base direction.
	DirectionLTR
	// DirectionRTL forces a right-to-left base direction.
	DirectionRTL
)

// ShapeFace is one font a [Shaper] may draw with. Faces are passed in
// priority order: the first face is the requested font, the others are
// fallbacks for characters the first face does not cover.
type ShapeFace struct {
	// Key is the family name the font was registered under. It is stable
	// for the lifetime of the document and may be used as a cache key
	// together with Data.
	Key string
	// Data is the raw TrueType font file.
	Data []byte
}

// Glyph is one positioned glyph produced by a [Shaper]. All distances are
// in the font design units of the face the glyph belongs to.
type Glyph struct {
	// ID is the glyph index in the font.
	ID uint16
	// XAdvance is the distance the pen moves after this glyph.
	XAdvance int
	// XOffset and YOffset move this glyph from the pen position without
	// moving the pen (mark positioning, for example). YOffset is positive
	// upwards.
	XOffset, YOffset int
	// Text is the source text this glyph represents, used to make the
	// PDF text searchable and copyable. When several glyphs come from the
	// same characters (a cluster), only one of them carries the text and
	// the others leave it empty. A ligature glyph carries all the
	// characters it replaces.
	Text string
}

// GlyphRun is a sequence of glyphs drawn with a single face.
type GlyphRun struct {
	// Face is an index into the faces passed to [Shaper.Shape].
	Face int
	// Glyphs are in visual order, left to right.
	Glyphs []Glyph
}

// Shaper turns one line of text into positioned glyphs. It is the hook for
// complex scripts (Arabic, Hebrew, Indic scripts, ...) that need contextual
// glyph selection, ligatures, mark positioning or bidirectional reordering,
// none of which gpdf does on its own. gpdf stays free of dependencies: an
// implementation usually wraps a shaping engine such as HarfBuzz.
//
// Shape receives a single line in logical (typing) order and returns its
// runs in visual order, left to right, as they will be drawn. Line breaking
// stays with gpdf, which measures candidate lines with Shape before
// drawing them.
//
// Implementations must be safe for concurrent use.
type Shaper interface {
	Shape(text string, faces []ShapeFace, dir Direction) []GlyphRun
}

// variantSuffixes are the suffixes gpdf appends to a family name to find
// its bold and italic faces, longest first.
var variantSuffixes = []string{"-BoldItalic", "-Bold", "-Italic"}

// FallbackKeys returns the registration keys of the fallback faces to use
// after the resolved primary face key. For each fallback family the same
// variant as the primary (e.g. "-Bold") is used when registered, otherwise
// the base family. Families that are not registered (has returns false)
// and duplicates of the primary are skipped.
func FallbackKeys(primary string, fallbacks []string, has func(key string) bool) []string {
	suffix := ""
	for _, s := range variantSuffixes {
		if len(primary) > len(s) && primary[len(primary)-len(s):] == s {
			suffix = s
			break
		}
	}
	seen := map[string]bool{primary: true}
	var keys []string
	for _, family := range fallbacks {
		key := family + suffix
		if suffix == "" || !has(key) {
			key = family
		}
		if seen[key] || !has(key) {
			continue
		}
		seen[key] = true
		keys = append(keys, key)
	}
	return keys
}

// ShapedWidth returns the advance width in points of shaped runs drawn at
// fontSize. upem returns the units per em of the face at the given index.
func ShapedWidth(runs []GlyphRun, upem func(face int) int, fontSize float64) float64 {
	var w float64
	for _, run := range runs {
		u := upem(run.Face)
		if u <= 0 {
			continue
		}
		var units int
		for _, g := range run.Glyphs {
			units += g.XAdvance
		}
		w += float64(units) * fontSize / float64(u)
	}
	return w
}

// WrapWords splits text into lines no wider than maxWidth, breaking at
// spaces and at explicit newlines. measure returns the width of a piece of
// text. Lines are built in logical order, before any bidi reordering, as
// the Unicode Bidi Algorithm requires. A word wider than maxWidth is cut
// between characters, never before a combining mark.
func WrapWords(text string, maxWidth float64, measure func(string) float64) []string {
	var lines []string
	space := measure(" ")
	for _, paragraph := range splitLines(text) {
		words := splitSpaces(paragraph)
		if len(words) == 0 {
			lines = append(lines, "")
			continue
		}
		var line string
		var lineW float64
		for _, word := range words {
			w := measure(word)
			switch {
			case line == "" && w <= maxWidth:
				line, lineW = word, w
			case line != "" && lineW+space+w <= maxWidth:
				line += " " + word
				lineW += space + w
			default:
				if line != "" {
					lines = append(lines, line)
				}
				line, lineW = word, w
				if w > maxWidth {
					pieces := cutWord(word, maxWidth, measure)
					lines = append(lines, pieces[:len(pieces)-1]...)
					line = pieces[len(pieces)-1]
					lineW = measure(line)
				}
			}
		}
		lines = append(lines, line)
	}
	return lines
}

func splitLines(text string) []string {
	var out []string
	start := 0
	for i := 0; i < len(text); i++ {
		if text[i] == '\n' {
			out = append(out, text[start:i])
			start = i + 1
		}
	}
	return append(out, text[start:])
}

func splitSpaces(s string) []string {
	var out []string
	start := -1
	for i, r := range s {
		if r == ' ' || r == '\t' {
			if start >= 0 {
				out = append(out, s[start:i])
				start = -1
			}
		} else if start < 0 {
			start = i
		}
	}
	if start >= 0 {
		out = append(out, s[start:])
	}
	return out
}

// cutWord splits a word into pieces that each fit maxWidth (a piece of one
// character may still be wider). Combining marks stay with their base.
func cutWord(word string, maxWidth float64, measure func(string) float64) []string {
	runes := []rune(word)
	var pieces []string
	start := 0
	for start < len(runes) {
		end := start + 1
		for end < len(runes) && isMark(runes[end]) {
			end++
		}
		for end < len(runes) {
			next := end + 1
			for next < len(runes) && isMark(runes[next]) {
				next++
			}
			if measure(string(runes[start:next])) > maxWidth {
				break
			}
			end = next
		}
		pieces = append(pieces, string(runes[start:end]))
		start = end
	}
	return pieces
}

func isMark(r rune) bool {
	return unicode.In(r, unicode.Mn, unicode.Me) || r == 0x200D // ZWJ
}
