package font

import (
	"unicode"
)

// Metrics holds font metric information in font design units.
type Metrics struct {
	UnitsPerEm  int
	Ascender    int // in font units (positive, above baseline)
	Descender   int // in font units (negative, below baseline)
	LineGap     int
	CapHeight   int
	XHeight     int
	ItalicAngle float64
}

// Font is the interface for fonts used by higher layers of the PDF library.
// Implementations must provide glyph metrics, encoding, and subsetting.
type Font interface {
	// Name returns the PostScript name of the font.
	Name() string

	// Metrics returns the font's metric information.
	Metrics() Metrics

	// GlyphWidth returns the advance width of the glyph for the given rune,
	// in font design units. The bool indicates whether the glyph was found.
	GlyphWidth(r rune) (int, bool)

	// Encode encodes a text string into bytes suitable for a PDF content stream.
	Encode(text string) []byte

	// Subset creates a subsetted font containing only the given runes.
	// Returns the subsetted font file data.
	Subset(runes []rune) ([]byte, error)
}

// MeasureString calculates the total advance width of text rendered at the
// given fontSize using font f. The result is in the same units as fontSize
// (typically points).
func MeasureString(f Font, text string, fontSize float64) float64 {
	m := f.Metrics()
	if m.UnitsPerEm == 0 {
		return 0
	}

	var totalWidth int
	for _, r := range text {
		w, ok := f.GlyphWidth(r)
		if !ok {
			// Fall back to space width or a default.
			w, _ = f.GlyphWidth(' ')
		}
		totalWidth += w
	}

	return float64(totalWidth) * fontSize / float64(m.UnitsPerEm)
}

// LineBreak splits text into lines that fit within maxWidth when rendered
// at the given fontSize using font f. It handles:
//   - Word wrapping at space boundaries for Latin text
//   - Character-level breaking for CJK characters
//   - Forced breaks on newline characters
func LineBreak(f Font, text string, fontSize, maxWidth float64) []string {
	if len(text) == 0 {
		return []string{""}
	}

	var lines []string
	runes := []rune(text)
	lineStart := 0

	for lineStart < len(runes) {
		// Find how many runes fit on this line.
		lineEnd := lineStart
		lastBreakPoint := -1
		var lineWidth float64

		for lineEnd < len(runes) {
			r := runes[lineEnd]

			// Handle explicit newlines.
			if r == '\n' {
				lines = append(lines, string(runes[lineStart:lineEnd]))
				lineStart = lineEnd + 1
				goto nextLine
			}

			w, ok := f.GlyphWidth(r)
			if !ok {
				w, _ = f.GlyphWidth(' ')
			}
			charWidth := float64(w) * fontSize / float64(f.Metrics().UnitsPerEm)

			if lineWidth+charWidth > maxWidth && lineEnd > lineStart {
				// This character doesn't fit.
				if lastBreakPoint > lineStart {
					bp := adjustBreakForKinsoku(runes, lastBreakPoint, lineStart)
					if bp > lineStart {
						lines = append(lines, string(runes[lineStart:bp]))
						lineStart = bp
						// Skip a space at the break point.
						if lineStart < len(runes) && runes[lineStart] == ' ' {
							lineStart++
						}
					} else {
						lines = append(lines, string(runes[lineStart:lineEnd]))
						lineStart = lineEnd
					}
				} else {
					// No space found; break at current position
					// (character-level break for CJK or long words).
					bp := adjustBreakForKinsoku(runes, lineEnd, lineStart)
					if bp > lineStart {
						lines = append(lines, string(runes[lineStart:bp]))
						lineStart = bp
					} else {
						lines = append(lines, string(runes[lineStart:lineEnd]))
						lineStart = lineEnd
					}
				}
				goto nextLine
			}

			lineWidth += charWidth

			// Record break opportunities.
			if r == ' ' {
				lastBreakPoint = lineEnd
			} else if isCJK(r) {
				// CJK characters can break after any character.
				lastBreakPoint = lineEnd + 1
			}

			lineEnd++
		}

		// Remaining text forms the last line.
		lines = append(lines, string(runes[lineStart:lineEnd]))
		break

	nextLine:
	}

	return lines
}

// kinsokuStart contains characters prohibited at the start of a line
// (行頭禁則). Breaking before any of these characters is suppressed.
var kinsokuStart = map[rune]bool{
	'）': true, '」': true, '』': true, '】': true,
	'〉': true, '》': true, '、': true, '。': true, '．': true,
	'，': true, '！': true, '？': true, '：': true, '；': true,
	'ー': true, '…': true, '‥': true, '｝': true, '〕': true,
	')': true, ']': true, '}': true, '!': true, '?': true,
	',': true, '.': true, ':': true, ';': true, '・': true,
	'゛': true, '゜': true, 'ゝ': true, 'ゞ': true,
	'ヽ': true, 'ヾ': true, '々': true,
}

// kinsokuEnd contains characters prohibited at the end of a line
// (行末禁則). Breaking after any of these characters is suppressed.
var kinsokuEnd = map[rune]bool{
	'（': true, '「': true, '『': true, '【': true,
	'〈': true, '《': true, '｛': true, '〔': true,
	'(': true, '[': true, '{': true,
}

// adjustBreakForKinsoku adjusts a candidate break position to satisfy
// kinsoku rules. It returns the adjusted position which may be earlier
// than breakPos but never earlier than lineStart.
func adjustBreakForKinsoku(runes []rune, breakPos, lineStart int) int {
	pos := breakPos
	// Move back if the character after the break is a kinsoku-start char
	// (would appear at line start) or the character at the break is a
	// kinsoku-end char (would appear at line end).
	for pos > lineStart {
		afterBreak := -1
		if pos < len(runes) {
			afterBreak = int(runes[pos])
		}
		// Skip space at the break point to look at the real next char.
		if afterBreak == ' ' && pos+1 < len(runes) {
			afterBreak = int(runes[pos+1])
		}
		atBreak := int(runes[pos-1])

		startViolation := afterBreak >= 0 && kinsokuStart[rune(afterBreak)]
		endViolation := kinsokuEnd[rune(atBreak)]

		if !startViolation && !endViolation {
			break
		}
		pos--
	}
	return pos
}

// isCJK returns true if the rune is a CJK ideograph or common CJK punctuation.
func isCJK(r rune) bool {
	return unicode.Is(unicode.Han, r) ||
		unicode.Is(unicode.Hangul, r) ||
		unicode.Is(unicode.Hiragana, r) ||
		unicode.Is(unicode.Katakana, r) ||
		(r >= 0x3000 && r <= 0x303F) || // CJK symbols and punctuation
		(r >= 0xFF00 && r <= 0xFFEF) // fullwidth forms
}
