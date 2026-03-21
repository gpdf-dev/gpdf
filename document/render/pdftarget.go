package render

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"image/png"
	"sort"
	"strings"

	"github.com/gpdf-dev/gpdf/document"
	"github.com/gpdf-dev/gpdf/document/layout"
	"github.com/gpdf-dev/gpdf/pdf"
	"github.com/gpdf-dev/gpdf/pdf/font"
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
	ttFonts     map[string]*font.TrueTypeFont // font family -> TrueType font object
	ttFontData  map[string][]byte             // font family -> raw TTF data
	pageContent []byte                        // accumulated content stream for the current page
	pageWidth   float64                       // current page width for MediaBox
	pageHeight  float64                       // current page height for Y-coordinate conversion
}

// NewPDFRenderer creates a PDFRenderer that writes to the given pdf.Writer.
func NewPDFRenderer(w *pdf.Writer) *PDFRenderer {
	return &PDFRenderer{
		writer:     w,
		fontMap:    make(map[string]string),
		fontRefs:   make(map[string]pdf.ObjectRef),
		imageMap:   make(map[string]string),
		imageRefs:  make(map[string]pdf.ObjectRef),
		ttFonts:    make(map[string]*font.TrueTypeFont),
		ttFontData: make(map[string][]byte),
	}
}

// RegisterTTFont registers a TrueType font for CJK-aware rendering.
// The font object is used for encoding text as glyph IDs, and the raw data
// is used for embedding a subsetted font file in the PDF.
func (r *PDFRenderer) RegisterTTFont(family string, ttf *font.TrueTypeFont, rawData []byte) {
	r.ttFonts[family] = ttf
	r.ttFontData[family] = rawData
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
	fontName, ttFontName, simulateBold, simulateItalic := r.resolveTextFont(style)

	fontResName, err := r.ensureFont(fontName)
	if err != nil {
		return err
	}

	pdfY := r.pageHeight - pos.Y - fontSize

	var buf strings.Builder
	buf.WriteString(style.Color.FillColorCmd())
	buf.WriteByte('\n')

	r.writeTextBoldSetup(&buf, simulateBold, style, fontSize)
	r.writeTextBegin(&buf, fontResName, fontSize, simulateBold, simulateItalic, pos.X, pdfY)
	r.writeTextSpacing(&buf, style)

	if !simulateItalic {
		fmt.Fprintf(&buf, "%g %g Td\n", pos.X, pdfY)
	}
	if ttf, ok := r.ttFonts[ttFontName]; ok {
		fmt.Fprintf(&buf, "<%s> Tj\n", hex.EncodeToString(ttf.Encode(text)))
	} else {
		fmt.Fprintf(&buf, "(%s) Tj\n", escapeStringPDF(text))
	}

	r.writeTextEnd(&buf, style, simulateBold)
	r.pageContent = append(r.pageContent, buf.String()...)
	return nil
}

func (r *PDFRenderer) resolveTextFont(style document.Style) (fontName, ttFontName string, simulateBold, simulateItalic bool) {
	fontName = resolvePDFFontName(style.FontFamily, style.FontWeight, style.FontStyle)
	ttFontName = fontName
	if _, ok := r.ttFonts[fontName]; !ok {
		if _, ok := r.ttFonts[style.FontFamily]; ok {
			ttFontName = style.FontFamily
			fontName = style.FontFamily
			simulateBold = style.FontWeight >= document.WeightBold
			simulateItalic = style.FontStyle == document.StyleItalic
		}
	}
	return
}

func (r *PDFRenderer) writeTextBoldSetup(buf *strings.Builder, simulateBold bool, style document.Style, fontSize float64) {
	if simulateBold {
		buf.WriteString(style.Color.StrokeColorCmd())
		buf.WriteByte('\n')
		fmt.Fprintf(buf, "%g w\n", fontSize*0.03)
	}
}

func (r *PDFRenderer) writeTextBegin(buf *strings.Builder, fontResName string, fontSize float64, simulateBold, simulateItalic bool, x, y float64) {
	buf.WriteString("BT\n")
	fmt.Fprintf(buf, "/%s %g Tf\n", fontResName, fontSize)
	if simulateBold {
		buf.WriteString("2 Tr\n")
	}
	if simulateItalic {
		fmt.Fprintf(buf, "1 0 0.2126 1 %g %g Tm\n", x, y)
	}
}

func (r *PDFRenderer) writeTextSpacing(buf *strings.Builder, style document.Style) {
	if style.WordSpacing != 0 {
		fmt.Fprintf(buf, "%g Tw\n", style.WordSpacing)
	}
	if style.LetterSpacing != 0 {
		fmt.Fprintf(buf, "%g Tc\n", style.LetterSpacing)
	}
}

func (r *PDFRenderer) writeTextEnd(buf *strings.Builder, style document.Style, simulateBold bool) {
	if style.LetterSpacing != 0 {
		buf.WriteString("0 Tc\n")
	}
	if style.WordSpacing != 0 {
		buf.WriteString("0 Tw\n")
	}
	if simulateBold {
		buf.WriteString("0 Tr\n")
	}
	buf.WriteString("ET\n")
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

	// Render background if present.
	if style.Background != nil {
		if err := r.RenderRect(document.Rectangle{
			X:      absX,
			Y:      absY,
			Width:  pn.Size.Width,
			Height: pn.Size.Height,
		}, RectStyle{FillColor: style.Background}); err != nil {
			return err
		}
	}

	// Render borders if any.
	if err := r.renderBorders(pn, style, absX, absY); err != nil {
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
func (r *PDFRenderer) renderBorders(pn layout.PlacedNode, style document.Style, absX, absY float64) error {
	border := style.Border
	x := absX
	y := absY
	w := pn.Size.Width
	h := pn.Size.Height

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

	// Check if this is a TrueType font registered via RegisterTTFont.
	if ttf, ok := r.ttFonts[family]; ok {
		resName, ref := r.writer.ReserveFontRef(family)
		r.fontMap[family] = resName
		r.fontRefs[family] = ref

		// Register a beforeClose hook to write the Type0/CIDFont structure
		// after all text has been encoded (so we know which glyphs to subset).
		rawData := r.ttFontData[family]
		r.writer.OnBeforeClose(func(pw *pdf.Writer) error {
			return r.writeType0Font(pw, family, ref, ttf, rawData)
		})
		return resName, nil
	}

	// Standard font: register directly with the writer.
	resName, ref, err := r.writer.RegisterFont(family, nil)
	if err != nil {
		return "", fmt.Errorf("render: failed to register font %q: %w", family, err)
	}

	r.fontMap[family] = resName
	r.fontRefs[family] = ref
	return resName, nil
}

// writeType0Font writes the complete Type0 composite font structure required
// for CJK and other non-WinAnsi text. The structure is:
//
//	Type0 Font → DescendantFonts → CIDFont (CIDFontType2)
//	                                 ├── FontDescriptor → FontFile2 (subsetted TTF)
//	                                 ├── DW (default width)
//	                                 └── W (per-glyph widths)
//	           → ToUnicode CMap stream
func (r *PDFRenderer) writeType0Font(pw *pdf.Writer, family string, fontRef pdf.ObjectRef, ttf *font.TrueTypeFont, rawData []byte) error {
	metrics := ttf.Metrics()

	// Subset the font to include only used glyphs.
	subsetData := r.subsetFontData(ttf, rawData)

	// Write FontFile2 (embedded font stream).
	fontFileRef, err := writeCompressedStream(pw, subsetData, pdf.Dict{
		pdf.Name("Length1"): pdf.Integer(len(subsetData)),
	})
	if err != nil {
		return err
	}

	// Write FontDescriptor.
	descRef, err := r.writeFontDescriptor(pw, family, metrics, fontFileRef)
	if err != nil {
		return err
	}

	// Build widths and write CIDFont + ToUnicode + Type0.
	runeToGID := ttf.RuneToGID()
	wArray := buildGlyphWidthArray(ttf, runeToGID, metrics.UnitsPerEm)

	dw := 1000
	if spaceW, ok := ttf.GlyphWidth(' '); ok {
		dw = spaceW * 1000 / metrics.UnitsPerEm
	}

	cidFontRef, err := r.writeCIDFont(pw, family, descRef, dw, wArray)
	if err != nil {
		return err
	}

	toUnicodeRef, err := writeToUnicodeCMap(pw, runeToGID)
	if err != nil {
		return err
	}

	return pw.WriteObject(fontRef, pdf.Dict{
		pdf.Name("Type"):            pdf.Name("Font"),
		pdf.Name("Subtype"):         pdf.Name("Type0"),
		pdf.Name("BaseFont"):        pdf.Name(family),
		pdf.Name("Encoding"):        pdf.Name("Identity-H"),
		pdf.Name("DescendantFonts"): pdf.Array{cidFontRef},
		pdf.Name("ToUnicode"):       toUnicodeRef,
	})
}

func (r *PDFRenderer) subsetFontData(ttf *font.TrueTypeFont, rawData []byte) []byte {
	usedRunes := ttf.UsedRunes()
	runes := make([]rune, 0, len(usedRunes))
	for r := range usedRunes {
		runes = append(runes, r)
	}
	sort.Slice(runes, func(i, j int) bool { return runes[i] < runes[j] })
	subsetData, err := ttf.Subset(runes)
	if err != nil {
		return rawData
	}
	return subsetData
}

func writeCompressedStream(pw *pdf.Writer, data []byte, extraDict pdf.Dict) (pdf.ObjectRef, error) {
	ref := pw.AllocObject()
	dict := extraDict
	content := data
	if compressed, err := pdf.CompressFlate(data); err == nil {
		dict[pdf.Name("Filter")] = pdf.Name("FlateDecode")
		content = compressed
	}
	return ref, pw.WriteObject(ref, pdf.Stream{Dict: dict, Content: content})
}

func (r *PDFRenderer) writeFontDescriptor(pw *pdf.Writer, family string, metrics font.Metrics, fontFileRef pdf.ObjectRef) (pdf.ObjectRef, error) {
	descRef := pw.AllocObject()
	return descRef, pw.WriteObject(descRef, pdf.Dict{
		pdf.Name("Type"):        pdf.Name("FontDescriptor"),
		pdf.Name("FontName"):    pdf.Name(family),
		pdf.Name("Flags"):       pdf.Integer(4), // Symbolic
		pdf.Name("ItalicAngle"): pdf.Real(metrics.ItalicAngle),
		pdf.Name("Ascent"):      pdf.Integer(metrics.Ascender * 1000 / metrics.UnitsPerEm),
		pdf.Name("Descent"):     pdf.Integer(metrics.Descender * 1000 / metrics.UnitsPerEm),
		pdf.Name("CapHeight"):   pdf.Integer(metrics.CapHeight * 1000 / metrics.UnitsPerEm),
		pdf.Name("FontFile2"):   fontFileRef,
		pdf.Name("FontBBox"):    pdf.Rectangle{LLX: 0, LLY: float64(metrics.Descender * 1000 / metrics.UnitsPerEm), URX: 1000, URY: float64(metrics.Ascender * 1000 / metrics.UnitsPerEm)},
	})
}

type gidWidth struct {
	gid   uint16
	width int
}

func buildGlyphWidthArray(ttf *font.TrueTypeFont, runeToGID map[rune]uint16, unitsPerEm int) pdf.Array {
	var gidWidths []gidWidth
	for r, gid := range runeToGID {
		w, ok := ttf.GlyphWidth(r)
		if !ok {
			continue
		}
		gidWidths = append(gidWidths, gidWidth{gid: gid, width: w * 1000 / unitsPerEm})
	}
	sort.Slice(gidWidths, func(i, j int) bool {
		return gidWidths[i].gid < gidWidths[j].gid
	})

	var wArray pdf.Array
	for i := 0; i < len(gidWidths); {
		startGID := gidWidths[i].gid
		widths := pdf.Array{pdf.Integer(gidWidths[i].width)}
		j := i + 1
		for j < len(gidWidths) && gidWidths[j].gid == gidWidths[j-1].gid+1 {
			widths = append(widths, pdf.Integer(gidWidths[j].width))
			j++
		}
		wArray = append(wArray, pdf.Integer(startGID), widths)
		i = j
	}
	return wArray
}

func (r *PDFRenderer) writeCIDFont(pw *pdf.Writer, family string, descRef pdf.ObjectRef, dw int, wArray pdf.Array) (pdf.ObjectRef, error) {
	cidFontRef := pw.AllocObject()
	cidFontDict := pdf.Dict{
		pdf.Name("Type"):     pdf.Name("Font"),
		pdf.Name("Subtype"):  pdf.Name("CIDFontType2"),
		pdf.Name("BaseFont"): pdf.Name(family),
		pdf.Name("CIDSystemInfo"): pdf.Dict{
			pdf.Name("Registry"):   pdf.LiteralString("Adobe"),
			pdf.Name("Ordering"):   pdf.LiteralString("Identity"),
			pdf.Name("Supplement"): pdf.Integer(0),
		},
		pdf.Name("FontDescriptor"): descRef,
		pdf.Name("DW"):             pdf.Integer(dw),
		pdf.Name("CIDToGIDMap"):    pdf.Name("Identity"),
	}
	if len(wArray) > 0 {
		cidFontDict[pdf.Name("W")] = wArray
	}
	return cidFontRef, pw.WriteObject(cidFontRef, cidFontDict)
}

func writeToUnicodeCMap(pw *pdf.Writer, runeToGID map[rune]uint16) (pdf.ObjectRef, error) {
	data := font.GenerateToUnicodeCMap(runeToGID)
	return writeCompressedStream(pw, data, pdf.Dict{})
}

// ensureImage ensures an image is registered and returns its resource name.
// It handles format-specific processing: JPEG data is stored with DCTDecode,
// PNG data is decoded to raw RGB pixels and stored with FlateDecode.
// For PNGs with transparency, an SMask image object is created for the alpha channel.
func (r *PDFRenderer) ensureImage(key string, src document.ImageSource) (string, error) {
	if resName, ok := r.imageMap[key]; ok {
		return resName, nil
	}

	var data []byte
	var smaskData []byte
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
		raw, alpha, pw, ph, err := decodePNGToRaw(src.Data)
		if err != nil {
			return "", fmt.Errorf("render: failed to decode PNG: %w", err)
		}
		data = raw
		smaskData = alpha
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

	resName, ref, err := r.writer.RegisterImage(key, data, w, h, colorSpace, filter, smaskData)
	if err != nil {
		return "", fmt.Errorf("render: failed to register image: %w", err)
	}

	r.imageMap[key] = resName
	r.imageRefs[key] = ref
	return resName, nil
}

// decodePNGToRaw decodes PNG binary data into raw RGB byte data and an
// optional alpha channel. If the image is fully opaque, alpha is nil.
func decodePNGToRaw(data []byte) (rgb []byte, alpha []byte, w, h int, err error) {
	img, err := png.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, nil, 0, 0, err
	}
	bounds := img.Bounds()
	w = bounds.Dx()
	h = bounds.Dy()

	rgb = make([]byte, w*h*3)
	alphaData := make([]byte, w*h)
	rgbIdx := 0
	alphaIdx := 0
	hasAlpha := false

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			r, g, b, a := img.At(x, y).RGBA()
			rgb[rgbIdx] = byte(r >> 8)
			rgb[rgbIdx+1] = byte(g >> 8)
			rgb[rgbIdx+2] = byte(b >> 8)
			rgbIdx += 3
			ab := byte(a >> 8)
			alphaData[alphaIdx] = ab
			alphaIdx++
			if ab != 0xFF {
				hasAlpha = true
			}
		}
	}

	if hasAlpha {
		return rgb, alphaData, w, h, nil
	}
	return rgb, nil, w, h, nil
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
