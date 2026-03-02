package layout

import (
	"strings"

	"github.com/gpdf-dev/gpdf/document"
)

// FlowLayout handles inline text layout with word wrapping. It breaks
// text content into lines that fit within the available width and
// calculates the vertical extent of the resulting text block.
type FlowLayout struct{}

// LayoutText breaks the given text into wrapped lines and returns a
// Result describing the bounds and per-line placed nodes. If the text
// does not fit vertically, the remaining text is returned as overflow.
func (fl *FlowLayout) LayoutText(text string, style document.Style, constraints Constraints) Result {
	if text == "" {
		return Result{
			Bounds: document.Rectangle{Width: constraints.AvailableWidth, Height: 0},
		}
	}

	fontSize := style.FontSize
	if fontSize <= 0 {
		fontSize = 12
	}

	lineHeight := style.LineHeight
	if lineHeight <= 0 {
		lineHeight = 1.2
	}

	lineSpacing := fontSize * lineHeight

	// Resolve text indent for the first line.
	indent := style.TextIndent.Resolve(constraints.AvailableWidth, fontSize)

	// Break text into lines, using a narrower first line when indented.
	lines := breakTextLinesIndented(text, style, fontSize, indent, constraints)

	var placed []PlacedNode
	cursorY := 0.0

	for i, line := range lines {
		// Check whether this line fits in the remaining vertical space.
		if cursorY+lineSpacing > constraints.AvailableHeight {
			return overflowResult(lines[i:], style, constraints.AvailableWidth, cursorY, placed)
		}

		lineWidth := measureLineWidth(line, style, fontSize, constraints)
		lineIndent := 0.0
		if i == 0 {
			lineIndent = indent
		}
		pn := placeLine(line, lineWidth, style, i == len(lines)-1, constraints.AvailableWidth, lineSpacing, cursorY)
		if lineIndent != 0 {
			pn.Position.X += lineIndent
		}
		placed = append(placed, pn)
		cursorY += lineSpacing
	}

	return Result{
		Bounds: document.Rectangle{
			Width:  constraints.AvailableWidth,
			Height: cursorY,
		},
		Children: placed,
	}
}

// breakTextLinesIndented splits text into wrapped lines, using a narrower
// width for the first line when indent > 0.
func breakTextLinesIndented(text string, style document.Style, fontSize, indent float64, constraints Constraints) []string {
	firstWidth := constraints.AvailableWidth - indent
	if firstWidth < 0 {
		firstWidth = 0
	}

	// When there is no indent, use the full width for all lines.
	if indent == 0 {
		return breakTextLines(text, style, fontSize, constraints.AvailableWidth, constraints)
	}

	// Break first line with narrower width, then remaining text with full width.
	firstLines := breakTextLines(text, style, fontSize, firstWidth, constraints)
	if len(firstLines) <= 1 {
		return firstLines
	}

	// The first element is the first line; re-break the rest at full width.
	rest := joinLines(firstLines[1:])
	remaining := breakTextLines(rest, style, fontSize, constraints.AvailableWidth, constraints)
	return append(firstLines[:1], remaining...)
}

// breakTextLines splits text into wrapped lines using the font resolver or
// a simple approximation at the given maxWidth.
func breakTextLines(text string, style document.Style, fontSize, maxWidth float64, constraints Constraints) []string {
	if constraints.FontResolver != nil {
		italic := style.FontStyle == document.StyleItalic
		font := constraints.FontResolver.Resolve(style.FontFamily, style.FontWeight, italic)
		return constraints.FontResolver.LineBreak(font, text, fontSize, maxWidth)
	}
	return approximateLineBreak(text, fontSize, maxWidth)
}

// measureLineWidth returns the rendered width of a line, including letter
// spacing between characters.
func measureLineWidth(line string, style document.Style, fontSize float64, constraints Constraints) float64 {
	if constraints.FontResolver == nil {
		return constraints.AvailableWidth
	}
	italic := style.FontStyle == document.StyleItalic
	font := constraints.FontResolver.Resolve(style.FontFamily, style.FontWeight, italic)
	w := constraints.FontResolver.MeasureString(font, line, fontSize)
	if style.LetterSpacing != 0 {
		charCount := len([]rune(line))
		if charCount > 1 {
			w += style.LetterSpacing * float64(charCount-1)
		}
	}
	return w
}

// placeLine creates a PlacedNode for a single line of text, applying
// justification and alignment as needed.
func placeLine(line string, lineWidth float64, style document.Style, isLastLine bool, availWidth, lineSpacing, cursorY float64) PlacedNode {
	lineStyle := style

	// For justified text, distribute extra space between words.
	// The last line of a paragraph uses left alignment (standard typographic convention).
	if style.TextAlign == document.AlignJustify && !isLastLine {
		spaces := strings.Count(line, " ")
		if spaces > 0 {
			lineStyle.WordSpacing = (availWidth - lineWidth) / float64(spaces)
		}
	}

	xPos := alignTextX(style.TextAlign, lineWidth, availWidth)

	nodeWidth := lineWidth
	if style.TextAlign == document.AlignJustify && !isLastLine && strings.Count(line, " ") > 0 {
		nodeWidth = availWidth
	}

	return PlacedNode{
		Node:     &document.Text{Content: line, TextStyle: lineStyle},
		Position: document.Point{X: xPos, Y: cursorY},
		Size:     document.Size{Width: nodeWidth, Height: lineSpacing},
	}
}

// overflowResult builds a Result for text that does not fit vertically.
func overflowResult(remaining []string, style document.Style, availWidth, cursorY float64, placed []PlacedNode) Result {
	overflowText := joinLines(remaining)
	overflow := &document.Text{
		Content:   overflowText,
		TextStyle: style,
	}
	return Result{
		Bounds: document.Rectangle{
			Width:  availWidth,
			Height: cursorY,
		},
		Children: placed,
		Overflow: overflow,
	}
}

// alignTextX returns the X offset for a line of text given the alignment
// mode, measured line width, and available container width.
func alignTextX(align document.TextAlign, lineWidth, availableWidth float64) float64 {
	switch align {
	case document.AlignCenter:
		return (availableWidth - lineWidth) / 2
	case document.AlignRight:
		return availableWidth - lineWidth
	default: // Left and Justify start at x=0
		return 0
	}
}

// approximateLineBreak performs a rough line break when no font resolver
// is available. It estimates character width as 0.5 * fontSize and wraps
// at word boundaries.
func approximateLineBreak(text string, fontSize, maxWidth float64) []string {
	avgCharWidth := fontSize * 0.5
	if avgCharWidth <= 0 {
		return []string{text}
	}
	charsPerLine := int(maxWidth / avgCharWidth)
	if charsPerLine <= 0 {
		charsPerLine = 1
	}

	var lines []string
	runes := []rune(text)
	start := 0
	for start < len(runes) {
		end := start + charsPerLine
		if end >= len(runes) {
			lines = append(lines, string(runes[start:]))
			break
		}
		// Try to break at a space.
		breakAt := end
		for breakAt > start && runes[breakAt] != ' ' {
			breakAt--
		}
		if breakAt == start {
			breakAt = end // force break mid-word
		}
		lines = append(lines, string(runes[start:breakAt]))
		start = breakAt
		// Skip the space at the break point.
		if start < len(runes) && runes[start] == ' ' {
			start++
		}
	}
	return lines
}

// ---------------------------------------------------------------------------
// RichText layout
// ---------------------------------------------------------------------------

// textRun is a word-level unit for inline layout. Each run carries its
// measured width and resolved style.
type textRun struct {
	text      string
	style     document.Style
	width     float64
	fontSize  float64
	isSpace   bool
	isNewline bool
}

// LayoutRichText lays out a RichText node, placing multiple styled
// fragments inline with word-wrapping. The resulting PlacedNode tree
// has one PlacedNode per line, each containing child PlacedNodes for
// the individual text runs.
func (fl *FlowLayout) LayoutRichText(rt *document.RichText, constraints Constraints) Result {
	if len(rt.Fragments) == 0 {
		return Result{
			Bounds: document.Rectangle{Width: constraints.AvailableWidth, Height: 0},
		}
	}

	runs := fragmentsToRuns(rt.Fragments, rt.BlockStyle, constraints)
	if len(runs) == 0 {
		return Result{
			Bounds: document.Rectangle{Width: constraints.AvailableWidth, Height: 0},
		}
	}

	indent := rt.BlockStyle.TextIndent.Resolve(constraints.AvailableWidth, effectiveFontSize(rt.BlockStyle.FontSize))

	lines := fillLines(runs, constraints.AvailableWidth, indent)

	var placed []PlacedNode
	cursorY := 0.0
	lineHeight := rt.BlockStyle.LineHeight
	if lineHeight <= 0 {
		lineHeight = 1.2
	}

	for i, line := range lines {
		maxFontSize := maxFontSizeInLine(line)
		lineSpacing := maxFontSize * lineHeight

		if cursorY+lineSpacing > constraints.AvailableHeight {
			// Build overflow from remaining lines.
			overflow := rebuildRichTextOverflow(lines[i:], rt.BlockStyle, rt.BreakPolicy)
			return Result{
				Bounds: document.Rectangle{
					Width:  constraints.AvailableWidth,
					Height: cursorY,
				},
				Children: placed,
				Overflow: overflow,
			}
		}

		lineIndent := 0.0
		if i == 0 {
			lineIndent = indent
		}
		pn := placeRichLine(line, rt.BlockStyle, constraints.AvailableWidth, lineSpacing, cursorY, lineIndent, i == len(lines)-1)
		placed = append(placed, pn)
		cursorY += lineSpacing
	}

	return Result{
		Bounds: document.Rectangle{
			Width:  constraints.AvailableWidth,
			Height: cursorY,
		},
		Children: placed,
	}
}

// fragmentsToRuns splits each RichTextFragment into word-level runs,
// inserting space runs between words. Each run's width is measured via
// the FontResolver.
func fragmentsToRuns(fragments []document.RichTextFragment, blockStyle document.Style, constraints Constraints) []textRun {
	var runs []textRun

	for _, frag := range fragments {
		if frag.Content == "" {
			continue
		}
		style := frag.FragmentStyle
		fontSize := effectiveFontSize(style.FontSize)
		if fontSize == 0 {
			fontSize = effectiveFontSize(blockStyle.FontSize)
		}
		style.FontSize = fontSize

		words := splitIntoWordsAndSpaces(frag.Content)
		for _, w := range words {
			if w == "\n" {
				runs = append(runs, textRun{
					text:      w,
					style:     style,
					fontSize:  fontSize,
					isNewline: true,
				})
				continue
			}
			isSpace := isAllSpaces(w)
			width := measureRunWidth(w, style, fontSize, constraints)
			runs = append(runs, textRun{
				text:     w,
				style:    style,
				width:    width,
				fontSize: fontSize,
				isSpace:  isSpace,
			})
		}
	}

	return runs
}

// splitIntoWordsAndSpaces splits text into alternating word and space runs.
// Newline characters are emitted as separate "\n" tokens so that the
// caller can insert forced line breaks (needed for <pre> and <br>).
// For example "Hello  world\nfoo" → ["Hello", "  ", "world", "\n", "foo"].
func splitIntoWordsAndSpaces(text string) []string {
	var parts []string
	runes := []rune(text)
	i := 0
	for i < len(runes) {
		switch runes[i] {
		case '\n':
			parts = append(parts, "\n")
			i++
		case ' ':
			j := i
			for j < len(runes) && runes[j] == ' ' {
				j++
			}
			parts = append(parts, string(runes[i:j]))
			i = j
		default:
			j := i
			for j < len(runes) && runes[j] != ' ' && runes[j] != '\n' {
				j++
			}
			parts = append(parts, string(runes[i:j]))
			i = j
		}
	}
	return parts
}

// isAllSpaces reports whether s consists entirely of space characters.
func isAllSpaces(s string) bool {
	for _, r := range s {
		if r != ' ' {
			return false
		}
	}
	return len(s) > 0
}

// measureRunWidth measures the width of a text run.
func measureRunWidth(text string, style document.Style, fontSize float64, constraints Constraints) float64 {
	if constraints.FontResolver == nil {
		return float64(len([]rune(text))) * fontSize * 0.5
	}
	italic := style.FontStyle == document.StyleItalic
	font := constraints.FontResolver.Resolve(style.FontFamily, style.FontWeight, italic)
	w := constraints.FontResolver.MeasureString(font, text, fontSize)
	if style.LetterSpacing != 0 {
		charCount := len([]rune(text))
		if charCount > 1 {
			w += style.LetterSpacing * float64(charCount-1)
		}
	}
	return w
}

// fillLines distributes runs into lines using a greedy algorithm.
// A line break is inserted at a space run when the next word would
// exceed the available width. Newline runs force an unconditional
// line break (used by <br> and <pre> blocks). The first line width
// is reduced by indent.
func fillLines(runs []textRun, availWidth, indent float64) [][]textRun {
	if len(runs) == 0 {
		return nil
	}

	var lines [][]textRun
	var currentLine []textRun
	lineWidth := 0.0
	lineNum := 0
	maxWidth := availWidth - indent
	afterForcedBreak := false

	for _, run := range runs {
		if run.isNewline {
			// Force a line break. Emit the current line (even if empty,
			// so that consecutive newlines produce blank lines).
			if len(currentLine) > 0 {
				currentLine = trimTrailingSpaces(currentLine)
			}
			lines = append(lines, currentLine)
			currentLine = nil
			lineWidth = 0
			lineNum++
			maxWidth = availWidth
			afterForcedBreak = true
			continue
		}

		if run.isSpace {
			// If adding this space would not exceed width, keep it.
			// Spaces at line start after a soft wrap are skipped, but
			// leading spaces after a forced break are preserved (e.g.
			// indentation inside <pre> blocks).
			if len(currentLine) == 0 {
				if afterForcedBreak {
					currentLine = append(currentLine, run)
					lineWidth += run.width
					afterForcedBreak = false
					continue
				}
				continue // skip leading spaces on a soft-wrapped line
			}
			afterForcedBreak = false
			if lineWidth+run.width <= maxWidth {
				currentLine = append(currentLine, run)
				lineWidth += run.width
			} else {
				// Space at line boundary: break here.
				lines = append(lines, currentLine)
				currentLine = nil
				lineWidth = 0
				lineNum++
				maxWidth = availWidth
			}
		} else {
			afterForcedBreak = false
			if len(currentLine) == 0 {
				// First word on line always fits.
				currentLine = append(currentLine, run)
				lineWidth += run.width
			} else if lineWidth+run.width <= maxWidth {
				currentLine = append(currentLine, run)
				lineWidth += run.width
			} else {
				// Word does not fit — break before it.
				// Trim trailing spaces from current line.
				currentLine = trimTrailingSpaces(currentLine)
				lines = append(lines, currentLine)
				currentLine = []textRun{run}
				lineWidth = run.width
				lineNum++
				maxWidth = availWidth
			}
		}
	}

	if len(currentLine) > 0 {
		currentLine = trimTrailingSpaces(currentLine)
		lines = append(lines, currentLine)
	}

	return lines
}

// trimTrailingSpaces removes space runs from the end of a line.
func trimTrailingSpaces(runs []textRun) []textRun {
	for len(runs) > 0 && runs[len(runs)-1].isSpace {
		runs = runs[:len(runs)-1]
	}
	return runs
}

// placeRichLine creates a PlacedNode for a single line of mixed-style text.
// It handles baseline alignment and text alignment (left/center/right/justify).
func placeRichLine(runs []textRun, blockStyle document.Style, availWidth, lineSpacing, cursorY, indent float64, isLastLine bool) PlacedNode {
	lineContentWidth := lineRunsWidth(runs)
	effectiveAvail := availWidth - indent

	// Calculate X offsets for text alignment.
	baseX := indent
	var extraSpacePerGap float64

	if blockStyle.TextAlign == document.AlignJustify && !isLastLine {
		spaceCount := countSpaceRuns(runs)
		if spaceCount > 0 {
			extraSpacePerGap = (effectiveAvail - lineContentWidth) / float64(spaceCount)
		}
	} else {
		baseX += alignTextX(blockStyle.TextAlign, lineContentWidth, effectiveAvail)
	}

	// Find max font size for baseline calculation.
	maxFS := maxFontSizeInLine(runs)

	// Place each run as a child PlacedNode.
	var children []PlacedNode
	cursorX := baseX

	for _, run := range runs {
		if run.isSpace {
			spaceW := run.width
			if blockStyle.TextAlign == document.AlignJustify && !isLastLine {
				spaceW += extraSpacePerGap
			}
			cursorX += spaceW
			continue
		}

		// Baseline alignment: all runs share the same baseline.
		// Y offset within the line = (maxFS - runFS) to align baselines.
		yOffset := maxFS - run.fontSize
		if yOffset < 0 {
			yOffset = 0
		}

		runNode := &document.Text{
			Content:   run.text,
			TextStyle: run.style,
		}

		children = append(children, PlacedNode{
			Node:     runNode,
			Position: document.Point{X: cursorX, Y: yOffset},
			Size:     document.Size{Width: run.width, Height: lineSpacing},
		})

		cursorX += run.width
	}

	// The line container uses a minimal RichText node for NodeType dispatch.
	nodeWidth := availWidth
	lineNode := &document.RichText{BlockStyle: blockStyle}

	return PlacedNode{
		Node:     lineNode,
		Position: document.Point{X: 0, Y: cursorY},
		Size:     document.Size{Width: nodeWidth, Height: lineSpacing},
		Children: children,
	}
}

// lineRunsWidth computes the total width of all runs in a line.
func lineRunsWidth(runs []textRun) float64 {
	w := 0.0
	for _, r := range runs {
		w += r.width
	}
	return w
}

// countSpaceRuns counts the number of space runs in a line (for justify).
func countSpaceRuns(runs []textRun) int {
	n := 0
	for _, r := range runs {
		if r.isSpace {
			n++
		}
	}
	return n
}

// maxFontSizeInLine returns the largest font size among all runs.
func maxFontSizeInLine(runs []textRun) float64 {
	maxFS := 0.0
	for _, r := range runs {
		if r.fontSize > maxFS {
			maxFS = r.fontSize
		}
	}
	if maxFS == 0 {
		maxFS = 12
	}
	return maxFS
}

// effectiveFontSize returns fontSize if positive, otherwise 12.
func effectiveFontSize(fontSize float64) float64 {
	if fontSize <= 0 {
		return 12
	}
	return fontSize
}

// rebuildRichTextOverflow reconstructs a RichText node from remaining
// lines of runs for overflow to the next page.
func rebuildRichTextOverflow(lines [][]textRun, blockStyle document.Style, bp document.BreakPolicy) *document.RichText {
	var fragments []document.RichTextFragment
	for i, line := range lines {
		if i > 0 {
			// Add a space between lines that were originally part of the same text flow.
			fragments = append(fragments, document.RichTextFragment{
				Content:       " ",
				FragmentStyle: line[0].style,
			})
		}
		for _, run := range line {
			fragments = append(fragments, document.RichTextFragment{
				Content:       run.text,
				FragmentStyle: run.style,
			})
		}
	}
	return &document.RichText{
		Fragments:   fragments,
		BlockStyle:  blockStyle,
		BreakPolicy: bp,
	}
}

// joinLines concatenates lines with a single space separator, restoring
// the original text for overflow handling.
func joinLines(lines []string) string {
	if len(lines) == 0 {
		return ""
	}
	result := lines[0]
	for _, l := range lines[1:] {
		result += " " + l
	}
	return result
}
