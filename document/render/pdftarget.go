package render

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"image/png"
	"strings"

	"github.com/gpdf-dev/gpdf/document"
	"github.com/gpdf-dev/gpdf/document/layout"
	"github.com/gpdf-dev/gpdf/pdf"
)

// PDFRenderer renders laid-out document nodes to a PDF file through a
// pdf.Writer. It translates high-level rendering commands into PDF content
// stream operators.
type PDFRenderer struct {
	writer      *pdf.Writer
	fontMap     map[string]string // font family key -> PDF resource name (e.g., "F1")
	fontRefs    map[string]pdf.ObjectRef
	imageMap    map[string]string // image content hash -> PDF resource name (e.g., "Im1")
	imageRefs   map[string]pdf.ObjectRef
	pageContent []byte  // accumulated content stream for the current page
	pageWidth   float64 // current page width for MediaBox
	pageHeight  float64 // current page height for Y-coordinate conversion
}

// NewPDFRenderer creates a PDFRenderer that writes to the given pdf.Writer.
func NewPDFRenderer(w *pdf.Writer) *PDFRenderer {
	return &PDFRenderer{
		writer:    w,
		fontMap:   make(map[string]string),
		fontRefs:  make(map[string]pdf.ObjectRef),
		imageMap:  make(map[string]string),
		imageRefs: make(map[string]pdf.ObjectRef),
	}
}

// BeginDocument sets the document metadata on the underlying PDF writer.
func (r *PDFRenderer) BeginDocument(info document.DocumentMetadata) error {
	r.writer.SetInfo(pdf.DocumentInfo{
		Title:    info.Title,
		Author:   info.Author,
		Subject:  info.Subject,
		Creator:  info.Creator,
		Producer: info.Producer,
	})
	return nil
}

// BeginPage starts a new page. The page height is recorded so that
// layout-coordinate Y values (origin at top-left) can be converted to
// PDF coordinates (origin at bottom-left).
func (r *PDFRenderer) BeginPage(size document.Size) error {
	r.pageContent = nil
	r.pageWidth = size.Width
	r.pageHeight = size.Height
	return nil
}

// EndPage writes the accumulated content stream as a PDF stream object,
// builds the page's resource dictionary, and adds the page to the writer.
func (r *PDFRenderer) EndPage() error {
	// Write the content stream.
	contentRef := r.writer.AllocObject()

	streamContent := r.pageContent
	contentStream := pdf.Stream{
		Dict:    pdf.Dict{},
		Content: streamContent,
	}
	if err := r.writer.WriteObject(contentRef, contentStream); err != nil {
		return fmt.Errorf("render: failed to write page content stream: %w", err)
	}

	// Build resource dictionary.
	resources := pdf.ResourceDict{}

	if len(r.fontRefs) > 0 {
		fontDict := make(pdf.Dict, len(r.fontRefs))
		for family, ref := range r.fontRefs {
			resName := r.fontMap[family]
			fontDict[pdf.Name(resName)] = ref
		}
		resources.Font = fontDict
	}

	if len(r.imageRefs) > 0 {
		xobjDict := make(pdf.Dict, len(r.imageRefs))
		for key, ref := range r.imageRefs {
			resName := r.imageMap[key]
			xobjDict[pdf.Name(resName)] = ref
		}
		resources.XObject = xobjDict
	}

	page := pdf.PageObject{
		MediaBox: pdf.Rectangle{
			LLX: 0,
			LLY: 0,
			URX: r.pageWidth,
			URY: r.pageHeight,
		},
		Resources: resources,
		Contents:  []pdf.ObjectRef{contentRef},
	}

	if err := r.writer.AddPage(page); err != nil {
		return fmt.Errorf("render: failed to add page: %w", err)
	}

	return nil
}

// RenderText draws text at the given layout position. The Y coordinate is
// converted from top-left origin to PDF's bottom-left origin.
func (r *PDFRenderer) RenderText(text string, pos document.Point, style document.Style) error {
	if text == "" {
		return nil
	}

	fontSize := style.FontSize
	if fontSize <= 0 {
		fontSize = 12
	}

	// Resolve the PDF font name including weight/style variants.
	fontName := resolvePDFFontName(style.FontFamily, style.FontWeight, style.FontStyle)

	// Ensure the font is registered.
	fontResName, err := r.ensureFont(fontName)
	if err != nil {
		return err
	}

	// Convert layout Y (top-left origin) to PDF Y (bottom-left origin).
	// The text baseline is approximately at pos.Y + fontSize (ascender).
	pdfY := r.pageHeight - pos.Y - fontSize

	var buf strings.Builder

	// Set fill color for text.
	buf.WriteString(style.Color.FillColorCmd())
	buf.WriteByte('\n')

	// Begin text block.
	buf.WriteString("BT\n")
	fmt.Fprintf(&buf, "/%s %g Tf\n", fontResName, fontSize)
	if style.WordSpacing != 0 {
		fmt.Fprintf(&buf, "%g Tw\n", style.WordSpacing)
	}
	if style.LetterSpacing != 0 {
		fmt.Fprintf(&buf, "%g Tc\n", style.LetterSpacing)
	}
	fmt.Fprintf(&buf, "%g %g Td\n", pos.X, pdfY)
	fmt.Fprintf(&buf, "(%s) Tj\n", escapeStringPDF(text))
	if style.LetterSpacing != 0 {
		buf.WriteString("0 Tc\n")
	}
	if style.WordSpacing != 0 {
		buf.WriteString("0 Tw\n")
	}
	buf.WriteString("ET\n")

	r.pageContent = append(r.pageContent, buf.String()...)
	return nil
}

// RenderRect draws a rectangle with optional fill and stroke.
func (r *PDFRenderer) RenderRect(rect document.Rectangle, style RectStyle) error {
	// Convert layout Y to PDF Y.
	pdfY := r.pageHeight - rect.Y - rect.Height

	var buf strings.Builder

	// Save graphics state.
	buf.WriteString("q\n")

	hasFill := style.FillColor != nil
	hasStroke := style.StrokeColor != nil

	if hasFill {
		buf.WriteString(style.FillColor.FillColorCmd())
		buf.WriteByte('\n')
	}

	if hasStroke {
		buf.WriteString(style.StrokeColor.StrokeColorCmd())
		buf.WriteByte('\n')
		if style.StrokeWidth > 0 {
			fmt.Fprintf(&buf, "%g w\n", style.StrokeWidth)
		}
	}

	// Draw rectangle path: x y width height re
	fmt.Fprintf(&buf, "%g %g %g %g re\n", rect.X, pdfY, rect.Width, rect.Height)

	// Fill and/or stroke.
	switch {
	case hasFill && hasStroke:
		buf.WriteString("B\n") // fill and stroke
	case hasFill:
		buf.WriteString("f\n") // fill only
	case hasStroke:
		buf.WriteString("S\n") // stroke only
	default:
		buf.WriteString("n\n") // no-op path
	}

	// Restore graphics state.
	buf.WriteString("Q\n")

	r.pageContent = append(r.pageContent, buf.String()...)
	return nil
}

// RenderImage draws an image at the given position and size using the
// PDF cm (concat matrix) and Do operators.
func (r *PDFRenderer) RenderImage(src document.ImageSource, pos document.Point, size document.Size) error {
	// Register the image via a content hash.
	imgKey := imageKey(src.Data)
	imgResName, err := r.ensureImage(imgKey, src)
	if err != nil {
		return err
	}

	// Convert layout Y to PDF Y. The image is placed with its
	// bottom-left at (pos.X, pdfY).
	pdfY := r.pageHeight - pos.Y - size.Height

	var buf strings.Builder

	// Save graphics state.
	buf.WriteString("q\n")

	// Apply transformation matrix: scale and translate.
	// The cm operator takes [a b c d e f] where the matrix maps
	// (1,1) image space to (width, height) in user space at (x, y).
	fmt.Fprintf(&buf, "%g 0 0 %g %g %g cm\n", size.Width, size.Height, pos.X, pdfY)

	// Paint the image XObject.
	fmt.Fprintf(&buf, "/%s Do\n", imgResName)

	// Restore graphics state.
	buf.WriteString("Q\n")

	r.pageContent = append(r.pageContent, buf.String()...)
	return nil
}

// RenderPath draws a path with optional fill, stroke, and dash pattern.
func (r *PDFRenderer) RenderPath(path document.Path, style PathStyle) error {
	var buf strings.Builder
	buf.WriteString("q\n")

	hasFill := style.FillColor != nil
	hasStroke := style.StrokeColor != nil

	writePathStyle(&buf, style, hasFill, hasStroke)
	r.writePathSegments(&buf, path)
	writePaintOp(&buf, hasFill, hasStroke)

	buf.WriteString("Q\n")
	r.pageContent = append(r.pageContent, buf.String()...)
	return nil
}

// writePathStyle emits PDF color, stroke-width, and dash-pattern operators.
func writePathStyle(buf *strings.Builder, style PathStyle, hasFill, hasStroke bool) {
	if hasFill {
		buf.WriteString(style.FillColor.FillColorCmd())
		buf.WriteByte('\n')
	}
	if hasStroke {
		buf.WriteString(style.StrokeColor.StrokeColorCmd())
		buf.WriteByte('\n')
		if style.StrokeWidth > 0 {
			fmt.Fprintf(buf, "%g w\n", style.StrokeWidth)
		}
	}
	if len(style.DashPattern) > 0 {
		buf.WriteByte('[')
		for i, d := range style.DashPattern {
			if i > 0 {
				buf.WriteByte(' ')
			}
			fmt.Fprintf(buf, "%g", d)
		}
		fmt.Fprintf(buf, "] %g d\n", style.DashPhase)
	}
}

// writePathSegments emits PDF path-construction operators (m, l, c, h)
// for each segment, converting Y coordinates from layout to PDF space.
func (r *PDFRenderer) writePathSegments(buf *strings.Builder, path document.Path) {
	for _, seg := range path.Segments {
		switch seg.Op {
		case document.PathMoveTo:
			if len(seg.Points) >= 1 {
				p := seg.Points[0]
				fmt.Fprintf(buf, "%g %g m\n", p.X, r.pageHeight-p.Y)
			}
		case document.PathLineTo:
			if len(seg.Points) >= 1 {
				p := seg.Points[0]
				fmt.Fprintf(buf, "%g %g l\n", p.X, r.pageHeight-p.Y)
			}
		case document.PathCurveTo:
			if len(seg.Points) >= 3 {
				c1, c2, ep := seg.Points[0], seg.Points[1], seg.Points[2]
				fmt.Fprintf(buf, "%g %g %g %g %g %g c\n",
					c1.X, r.pageHeight-c1.Y,
					c2.X, r.pageHeight-c2.Y,
					ep.X, r.pageHeight-ep.Y)
			}
		case document.PathClose:
			buf.WriteString("h\n")
		}
	}
}

// writePaintOp emits the appropriate PDF paint operator based on fill/stroke flags.
func writePaintOp(buf *strings.Builder, hasFill, hasStroke bool) {
	switch {
	case hasFill && hasStroke:
		buf.WriteString("B\n")
	case hasFill:
		buf.WriteString("f\n")
	case hasStroke:
		buf.WriteString("S\n")
	default:
		buf.WriteString("n\n")
	}
}

// RenderLine draws a straight line between two points by delegating to
// RenderPath with a MoveTo–LineTo path.
func (r *PDFRenderer) RenderLine(from, to document.Point, style LineStyle) error {
	path := document.Path{
		Segments: []document.PathSegment{
			{Op: document.PathMoveTo, Points: []document.Point{from}},
			{Op: document.PathLineTo, Points: []document.Point{to}},
		},
	}
	c := style.Color
	return r.RenderPath(path, PathStyle{
		StrokeColor: &c,
		StrokeWidth: style.Width,
		DashPattern: style.DashPattern,
		DashPhase:   style.DashPhase,
	})
}

// EndDocument finalizes the PDF by calling Close on the writer.
func (r *PDFRenderer) EndDocument() error {
	return r.writer.Close()
}

// RenderDocument renders a complete paginated document by iterating over
// the page layouts produced by the paginator. This is the primary entry
// point for rendering a fully laid-out document.
func (r *PDFRenderer) RenderDocument(pages []layout.PageLayout, info document.DocumentMetadata) error {
	if err := r.BeginDocument(info); err != nil {
		return err
	}

	for _, page := range pages {
		if err := r.BeginPage(page.Size); err != nil {
			return err
		}

		if err := r.renderPlacedNodes(page.Children, 0, 0); err != nil {
			return err
		}

		if err := r.EndPage(); err != nil {
			return err
		}
	}

	return r.EndDocument()
}

// renderPlacedNodes recursively renders a list of placed nodes.
// offsetX and offsetY accumulate parent positions so that children
// (whose positions are relative to their parent) are drawn at the
// correct absolute page coordinates.
func (r *PDFRenderer) renderPlacedNodes(nodes []layout.PlacedNode, offsetX, offsetY float64) error {
	for _, pn := range nodes {
		if err := r.renderPlacedNode(pn, offsetX, offsetY); err != nil {
			return err
		}
	}
	return nil
}

// renderPlacedNode renders a single placed node and its children.
// offsetX/offsetY is the accumulated absolute position of the parent.
func (r *PDFRenderer) renderPlacedNode(pn layout.PlacedNode, offsetX, offsetY float64) error {
	if pn.Node == nil {
		return nil
	}

	// Compute absolute position by adding parent offset.
	absX := pn.Position.X + offsetX
	absY := pn.Position.Y + offsetY

	style := pn.Node.Style()

	// Resolve margins so background and borders render at the border box,
	// not the margin box. The PlacedNode position/size includes margins.
	fontSize := style.FontSize
	if fontSize <= 0 {
		fontSize = 12
	}
	margin := style.Margin.Resolve(pn.Size.Width, pn.Size.Height, fontSize)
	borderX := absX + margin.Left
	borderY := absY + margin.Top
	borderW := pn.Size.Width - margin.Left - margin.Right
	borderH := pn.Size.Height - margin.Top - margin.Bottom

	// Render background if present (within border box).
	if style.Background != nil {
		if err := r.RenderRect(document.Rectangle{
			X:      borderX,
			Y:      borderY,
			Width:  borderW,
			Height: borderH,
		}, RectStyle{FillColor: style.Background}); err != nil {
			return err
		}
	}

	// Render borders if any (within border box).
	if err := r.renderBorders(pn, style, borderX, borderY, borderW, borderH); err != nil {
		return err
	}

	// Render node-specific content.
	switch pn.Node.NodeType() {
	case document.NodeText:
		// Only render text for leaf nodes (individual lines from FlowLayout).
		// Parent text nodes with children act as containers; their children
		// carry the actual per-line content.
		if len(pn.Children) == 0 {
			textNode, ok := pn.Node.(*document.Text)
			if ok {
				if err := r.RenderText(textNode.Content, document.Point{X: absX, Y: absY}, style); err != nil {
					return err
				}
				if err := r.renderTextDecoration(style, absX, absY, pn.Size.Width); err != nil {
					return err
				}
			}
		}
	case document.NodeImage:
		imgNode, ok := pn.Node.(*document.Image)
		if ok {
			if err := r.RenderImage(imgNode.Source, document.Point{X: absX, Y: absY}, pn.Size); err != nil {
				return err
			}
		}
	case document.NodeRichText:
		// RichText lines act as containers. Background is handled above.
		// Child Text nodes are rendered by the recursive call below.
	}

	// Render children with this node's absolute position as offset.
	return r.renderPlacedNodes(pn.Children, absX, absY)
}

// renderTextDecoration draws underline, strikethrough, and/or overline
// lines for a text node when the style requests them.
func (r *PDFRenderer) renderTextDecoration(style document.Style, absX, absY, textWidth float64) error {
	if style.TextDecoration == document.DecorationNone {
		return nil
	}

	fontSize := style.FontSize
	if fontSize <= 0 {
		fontSize = 12
	}

	thickness := fontSize / 18

	if style.TextDecoration&document.DecorationUnderline != 0 {
		// Underline: slightly below the baseline.
		y := absY + fontSize + fontSize*0.1
		if err := r.RenderLine(
			document.Point{X: absX, Y: y},
			document.Point{X: absX + textWidth, Y: y},
			LineStyle{Color: style.Color, Width: thickness},
		); err != nil {
			return err
		}
	}

	if style.TextDecoration&document.DecorationStrikethrough != 0 {
		// Strikethrough: approximately at x-height center (≈ 0.35 * fontSize from top).
		y := absY + fontSize*0.65
		if err := r.RenderLine(
			document.Point{X: absX, Y: y},
			document.Point{X: absX + textWidth, Y: y},
			LineStyle{Color: style.Color, Width: thickness},
		); err != nil {
			return err
		}
	}

	if style.TextDecoration&document.DecorationOverline != 0 {
		// Overline: at the top of the text (ascender line).
		y := absY
		if err := r.RenderLine(
			document.Point{X: absX, Y: y},
			document.Point{X: absX + textWidth, Y: y},
			LineStyle{Color: style.Color, Width: thickness},
		); err != nil {
			return err
		}
	}

	return nil
}

// renderBorders draws the border sides of a placed node at the given
// absolute position (absX, absY).
func (r *PDFRenderer) renderBorders(pn layout.PlacedNode, style document.Style, x, y, w, h float64) error {
	border := style.Border

	if border.Top.Style != document.BorderNone {
		bw := border.Top.Width.Resolve(w, style.FontSize)
		if err := r.RenderRect(document.Rectangle{
			X: x, Y: y, Width: w, Height: bw,
		}, RectStyle{FillColor: &border.Top.Color}); err != nil {
			return err
		}
	}
	if border.Bottom.Style != document.BorderNone {
		bw := border.Bottom.Width.Resolve(w, style.FontSize)
		if err := r.RenderRect(document.Rectangle{
			X: x, Y: y + h - bw, Width: w, Height: bw,
		}, RectStyle{FillColor: &border.Bottom.Color}); err != nil {
			return err
		}
	}
	if border.Left.Style != document.BorderNone {
		bw := border.Left.Width.Resolve(w, style.FontSize)
		if err := r.RenderRect(document.Rectangle{
			X: x, Y: y, Width: bw, Height: h,
		}, RectStyle{FillColor: &border.Left.Color}); err != nil {
			return err
		}
	}
	if border.Right.Style != document.BorderNone {
		bw := border.Right.Width.Resolve(w, style.FontSize)
		if err := r.RenderRect(document.Rectangle{
			X: x + w - bw, Y: y, Width: bw, Height: h,
		}, RectStyle{FillColor: &border.Right.Color}); err != nil {
			return err
		}
	}
	return nil
}

// ensureFont ensures the font is registered with the PDF writer and
// returns its resource name (e.g., "F1").
func (r *PDFRenderer) ensureFont(family string) (string, error) {
	if family == "" {
		family = "Helvetica"
	}
	if resName, ok := r.fontMap[family]; ok {
		return resName, nil
	}

	// Register the font with the writer. For standard fonts, no font
	// data is needed.
	resName, ref, err := r.writer.RegisterFont(family, nil)
	if err != nil {
		return "", fmt.Errorf("render: failed to register font %q: %w", family, err)
	}

	r.fontMap[family] = resName
	r.fontRefs[family] = ref
	return resName, nil
}

// ensureImage ensures an image is registered and returns its resource name.
// It handles format-specific processing: JPEG data is stored with DCTDecode,
// PNG data is decoded to raw RGB pixels and stored with FlateDecode.
func (r *PDFRenderer) ensureImage(key string, src document.ImageSource) (string, error) {
	if resName, ok := r.imageMap[key]; ok {
		return resName, nil
	}

	var data []byte
	var w, h int
	var colorSpace, filter string

	switch src.Format {
	case document.ImageJPEG:
		data = src.Data
		w = src.Width
		h = src.Height
		colorSpace = "DeviceRGB"
		filter = "DCTDecode"
	case document.ImagePNG:
		raw, pw, ph, err := decodePNGToRaw(src.Data)
		if err != nil {
			return "", fmt.Errorf("render: failed to decode PNG: %w", err)
		}
		data = raw
		w = pw
		h = ph
		colorSpace = "DeviceRGB"
		filter = ""
	default:
		data = src.Data
		w = src.Width
		h = src.Height
		colorSpace = "DeviceRGB"
		filter = ""
	}

	resName, ref, err := r.writer.RegisterImage(key, data, w, h, colorSpace, filter)
	if err != nil {
		return "", fmt.Errorf("render: failed to register image: %w", err)
	}

	r.imageMap[key] = resName
	r.imageRefs[key] = ref
	return resName, nil
}

// decodePNGToRaw decodes PNG binary data into raw RGB byte data.
func decodePNGToRaw(data []byte) ([]byte, int, int, error) {
	img, err := png.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, 0, 0, err
	}
	bounds := img.Bounds()
	w := bounds.Dx()
	h := bounds.Dy()

	raw := make([]byte, w*h*3)
	idx := 0
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			r, g, b, _ := img.At(x, y).RGBA()
			raw[idx] = byte(r >> 8)
			raw[idx+1] = byte(g >> 8)
			raw[idx+2] = byte(b >> 8)
			idx += 3
		}
	}
	return raw, w, h, nil
}

// imageKey returns a hex-encoded SHA-256 hash of the image data, used
// as a deduplication key for registered images.
func imageKey(data []byte) string {
	h := sha256.Sum256(data)
	return fmt.Sprintf("%x", h[:8])
}

// fontVariantKey identifies a (bold, italic) combination for font variant lookup.
type fontVariantKey struct {
	bold, italic bool
}

// base14Variants maps standard PDF Base-14 font families to their
// bold/italic variant names.
var base14Variants = map[string]map[fontVariantKey]string{
	"Helvetica": {
		{true, false}: "Helvetica-Bold",
		{false, true}: "Helvetica-Oblique",
		{true, true}:  "Helvetica-BoldOblique",
	},
	"Times-Roman": {
		{true, false}: "Times-Bold",
		{false, true}: "Times-Italic",
		{true, true}:  "Times-BoldItalic",
	},
	"Courier": {
		{true, false}: "Courier-Bold",
		{false, true}: "Courier-Oblique",
		{true, true}:  "Courier-BoldOblique",
	},
}

// resolvePDFFontName maps a font family, weight, and style to the correct
// PDF font name. For the standard PDF Base-14 fonts it returns the canonical
// variant name (e.g. "Helvetica-Bold"); for other families it appends
// "-Bold" / "-Italic" / "-BoldItalic" suffixes.
func resolvePDFFontName(family string, weight document.FontWeight, fontStyle document.FontStyle) string {
	if family == "" {
		family = "Helvetica"
	}

	bold := weight >= document.WeightBold
	italic := fontStyle == document.StyleItalic

	if !bold && !italic {
		return family
	}

	key := fontVariantKey{bold, italic}

	// Standard PDF Base-14 font families use specific variant names.
	if variants, ok := base14Variants[family]; ok {
		if name, ok := variants[key]; ok {
			return name
		}
	}

	// For non-standard fonts, append conventional suffixes.
	suffix := "-Italic"
	if bold {
		suffix = "-Bold"
	}
	if bold && italic {
		suffix = "-BoldItalic"
	}
	return family + suffix
}

// escapeStringPDF converts a UTF-8 Go string to a WinAnsiEncoding byte
// sequence suitable for a PDF literal string, escaping special characters.
// Characters outside the WinAnsiEncoding repertoire are replaced with '?'.
func escapeStringPDF(s string) string {
	var buf strings.Builder
	buf.Grow(len(s))
	for _, r := range s {
		b := runeToWinAnsi(r)
		switch b {
		case '(':
			buf.WriteString(`\(`)
		case ')':
			buf.WriteString(`\)`)
		case '\\':
			buf.WriteString(`\\`)
		case '\r':
			buf.WriteString(`\r`)
		case '\n':
			buf.WriteString(`\n`)
		default:
			buf.WriteByte(b)
		}
	}
	return buf.String()
}

// runeToWinAnsi maps a Unicode rune to its WinAnsiEncoding byte value.
// WinAnsiEncoding is the default encoding for Standard 14 (Type1) and
// simple TrueType fonts in PDF. It matches ISO 8859-1 (Latin-1) for
// 0x20–0x7E and 0xA0–0xFF, with additional characters in 0x80–0x9F.
// Returns '?' for runes not representable in WinAnsiEncoding.
func runeToWinAnsi(r rune) byte {
	// ASCII printable + common controls.
	if r >= 0x20 && r <= 0x7E {
		return byte(r)
	}
	// Latin-1 Supplement (0xA0–0xFF): direct mapping.
	if r >= 0xA0 && r <= 0xFF {
		return byte(r)
	}
	// Control characters used by PDF.
	switch r {
	case '\n', '\r', '\t':
		return byte(r)
	}
	// WinAnsiEncoding characters in the 0x80–0x9F range.
	if b, ok := winAnsiSpecial[r]; ok {
		return b
	}
	return '?'
}

// winAnsiSpecial maps Unicode code points to WinAnsiEncoding byte positions
// for the 0x80–0x9F range, which differs from ISO 8859-1.
var winAnsiSpecial = map[rune]byte{
	'\u20AC': 0x80, // € Euro sign
	'\u201A': 0x82, // ‚ Single low-9 quotation mark
	'\u0192': 0x83, // ƒ Latin small letter f with hook
	'\u201E': 0x84, // „ Double low-9 quotation mark
	'\u2026': 0x85, // … Horizontal ellipsis
	'\u2020': 0x86, // † Dagger
	'\u2021': 0x87, // ‡ Double dagger
	'\u02C6': 0x88, // ˆ Modifier letter circumflex accent
	'\u2030': 0x89, // ‰ Per mille sign
	'\u0160': 0x8A, // Š Latin capital letter S with caron
	'\u2039': 0x8B, // ‹ Single left-pointing angle quotation mark
	'\u0152': 0x8C, // Œ Latin capital ligature OE
	'\u017D': 0x8E, // Ž Latin capital letter Z with caron
	'\u2018': 0x91, // ' Left single quotation mark
	'\u2019': 0x92, // ' Right single quotation mark
	'\u201C': 0x93, // " Left double quotation mark
	'\u201D': 0x94, // " Right double quotation mark
	'\u2022': 0x95, // • Bullet
	'\u2013': 0x96, // – En dash
	'\u2014': 0x97, // — Em dash
	'\u02DC': 0x98, // ˜ Small tilde
	'\u2122': 0x99, // ™ Trade mark sign
	'\u0161': 0x9A, // š Latin small letter s with caron
	'\u203A': 0x9B, // › Single right-pointing angle quotation mark
	'\u0153': 0x9C, // œ Latin small ligature oe
	'\u017E': 0x9E, // ž Latin small letter z with caron
	'\u0178': 0x9F, // Ÿ Latin capital letter Y with diaeresis
}
