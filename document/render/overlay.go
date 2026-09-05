package render

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/gpdf-dev/gpdf/document"
	"github.com/gpdf-dev/gpdf/document/layout"
	"github.com/gpdf-dev/gpdf/pdf"
	"github.com/gpdf-dev/gpdf/pdf/font"
)

// OverlayResult holds the content stream and resource references
// produced by rendering overlay content for a single page.
type OverlayResult struct {
	Content   []byte   // raw PDF content stream operators
	Resources pdf.Dict // resource dictionary entries to merge

	// FontObjects contains font objects that need to be written.
	FontObjects map[string]fontObject
	// ImageObjects contains image objects that need to be written.
	ImageObjects map[string]imageObject
}

type fontObject struct {
	ResName string
	Family  string
	Data    []byte             // nil for standard (non-embedded) fonts
	TTF     *font.TrueTypeFont // nil for standard fonts
}

type imageObject struct {
	ResName    string
	Data       []byte
	SmaskData  []byte
	Width      int
	Height     int
	ColorSpace string
	Filter     string
}

// OverlayRenderer renders document nodes to a content stream byte slice,
// suitable for overlaying on an existing PDF page. Unlike PDFRenderer,
// it does not write to a pdf.Writer directly — instead it captures
// the content stream and tracks resource references.
type OverlayRenderer struct {
	content    []byte
	pageWidth  float64
	pageHeight float64

	fontMap     map[string]string // family -> resource name (F1, F2, ...)
	fontCount   int
	fontObjects map[string]fontObject

	imageMap     map[string]string // hash -> resource name (Im1, Im2, ...)
	imageCount   int
	imageObjects map[string]imageObject

	fontDataMap map[string][]byte
	fonts       map[string]*font.TrueTypeFont
}

// NewOverlayRenderer creates a renderer that captures overlay content.
// pageWidth and pageHeight define the coordinate system (for Y-flip).
// fontDataMap maps font families to their TTF data for registration.
func NewOverlayRenderer(pageWidth, pageHeight float64, fonts map[string]*font.TrueTypeFont, fontDataMap map[string][]byte) *OverlayRenderer {
	return &OverlayRenderer{
		pageWidth:    pageWidth,
		pageHeight:   pageHeight,
		fontMap:      make(map[string]string),
		fontObjects:  make(map[string]fontObject),
		imageMap:     make(map[string]string),
		imageObjects: make(map[string]imageObject),
		fontDataMap:  fontDataMap,
		fonts:        fonts,
	}
}

// RenderOverlay renders the given placed nodes and returns the result.
func (r *OverlayRenderer) RenderOverlay(nodes []layout.PlacedNode) (*OverlayResult, error) {
	r.content = nil

	if err := r.renderPlacedNodes(nodes, 0, 0); err != nil {
		return nil, err
	}

	// Build resource dict.
	resources := make(pdf.Dict)
	if len(r.fontMap) > 0 {
		fontDict := make(pdf.Dict)
		for family, resName := range r.fontMap {
			// Font refs will be resolved by the caller.
			fontDict[pdf.Name(resName)] = pdf.Name(family)
		}
		resources[pdf.Name("Font")] = fontDict
	}

	return &OverlayResult{
		Content:      r.content,
		Resources:    resources,
		FontObjects:  r.fontObjects,
		ImageObjects: r.imageObjects,
	}, nil
}

func (r *OverlayRenderer) renderPlacedNodes(nodes []layout.PlacedNode, offsetX, offsetY float64) error {
	for _, pn := range nodes {
		if err := r.renderPlacedNode(pn, offsetX, offsetY); err != nil {
			return err
		}
	}
	return nil
}

func (r *OverlayRenderer) renderPlacedNode(pn layout.PlacedNode, offsetX, offsetY float64) error {
	if pn.Node == nil {
		return nil
	}

	absX := pn.Position.X + offsetX
	absY := pn.Position.Y + offsetY
	style := pn.Node.Style()

	// Background.
	if style.Background != nil {
		r.renderRect(document.Rectangle{
			X: absX, Y: absY, Width: pn.Size.Width, Height: pn.Size.Height,
		}, RectStyle{FillColor: style.Background})
	}

	// Node-specific content.
	switch pn.Node.NodeType() {
	case document.NodeText:
		if len(pn.Children) == 0 {
			if textNode, ok := pn.Node.(*document.Text); ok {
				r.renderText(textNode.Content, document.Point{X: absX, Y: absY}, style)
			}
		}
	case document.NodeImage:
		if imgNode, ok := pn.Node.(*document.Image); ok {
			r.renderImage(imgNode.Source, document.Point{X: absX, Y: absY}, pn.Size)
		}
	}

	return r.renderPlacedNodes(pn.Children, absX, absY)
}

func (r *OverlayRenderer) renderText(text string, pos document.Point, style document.Style) {
	if text == "" {
		return
	}

	fontSize := style.FontSize
	if fontSize <= 0 {
		fontSize = 12
	}

	fontName, ttf := r.resolveTextFont(style)
	resName := r.ensureFont(fontName)

	pdfY := r.pageHeight - pos.Y - fontSize

	var buf strings.Builder
	buf.WriteString(style.Color.FillColorCmd())
	buf.WriteByte('\n')
	buf.WriteString("BT\n")
	fmt.Fprintf(&buf, "/%s %g Tf\n", resName, fontSize)
	if style.WordSpacing != 0 {
		fmt.Fprintf(&buf, "%g Tw\n", style.WordSpacing)
	}
	if style.LetterSpacing != 0 {
		fmt.Fprintf(&buf, "%g Tc\n", style.LetterSpacing)
	}
	fmt.Fprintf(&buf, "%g %g Td\n", pos.X, pdfY)
	if ttf != nil {
		// Embedded TrueType fonts are written as Type0/Identity-H composite
		// fonts, so text is encoded as big-endian glyph IDs. Emitting the raw
		// UTF-8 bytes here instead makes viewers render every non-ASCII
		// character as "?" (issue #37).
		fmt.Fprintf(&buf, "<%s> Tj\n", hex.EncodeToString(ttf.Encode(text)))
	} else {
		fmt.Fprintf(&buf, "(%s) Tj\n", escapeStringPDF(text))
	}
	if style.LetterSpacing != 0 {
		buf.WriteString("0 Tc\n")
	}
	if style.WordSpacing != 0 {
		buf.WriteString("0 Tw\n")
	}
	buf.WriteString("ET\n")

	r.content = append(r.content, buf.String()...)
}

func (r *OverlayRenderer) renderRect(rect document.Rectangle, style RectStyle) {
	pdfY := r.pageHeight - rect.Y - rect.Height

	var buf strings.Builder
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

	fmt.Fprintf(&buf, "%g %g %g %g re\n", rect.X, pdfY, rect.Width, rect.Height)

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

	buf.WriteString("Q\n")
	r.content = append(r.content, buf.String()...)
}

func (r *OverlayRenderer) renderImage(src document.ImageSource, pos document.Point, size document.Size) {
	imgKey := overlayImageKey(src.Data)
	resName := r.ensureImage(imgKey, src)

	pdfY := r.pageHeight - pos.Y - size.Height

	var buf strings.Builder
	buf.WriteString("q\n")
	fmt.Fprintf(&buf, "%g 0 0 %g %g %g cm\n", size.Width, size.Height, pos.X, pdfY)
	fmt.Fprintf(&buf, "/%s Do\n", resName)
	buf.WriteString("Q\n")

	r.content = append(r.content, buf.String()...)
}

// resolveTextFont picks the PDF font name for a style and returns the
// registered TrueType font behind it, if any. Like the main renderer it falls
// back from the variant name (e.g. "MyFont-Bold") to the base family so that
// a single registered face still gets used — and embedded — for bold/italic
// text instead of silently degrading to Helvetica.
func (r *OverlayRenderer) resolveTextFont(style document.Style) (string, *font.TrueTypeFont) {
	fontName := resolvePDFFontName(style.FontFamily, style.FontWeight, style.FontStyle)
	if ttf, ok := r.fonts[fontName]; ok {
		return fontName, ttf
	}
	if ttf, ok := r.fonts[style.FontFamily]; ok {
		return style.FontFamily, ttf
	}
	return fontName, nil
}

func (r *OverlayRenderer) ensureFont(family string) string {
	if family == "" {
		family = "Helvetica"
	}
	if resName, ok := r.fontMap[family]; ok {
		return resName
	}

	r.fontCount++
	resName := fmt.Sprintf("OvF%d", r.fontCount)
	r.fontMap[family] = resName

	ttf := r.fonts[family]

	var data []byte
	if r.fontDataMap != nil {
		data = r.fontDataMap[family]
	}
	if len(data) == 0 && ttf != nil {
		data = ttf.Data()
	}

	r.fontObjects[family] = fontObject{
		ResName: resName,
		Family:  family,
		Data:    data,
		TTF:     ttf,
	}

	return resName
}

func (r *OverlayRenderer) ensureImage(key string, src document.ImageSource) string {
	if resName, ok := r.imageMap[key]; ok {
		return resName
	}

	r.imageCount++
	resName := fmt.Sprintf("OvIm%d", r.imageCount)
	r.imageMap[key] = resName

	var data, smaskData []byte
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
		if err == nil {
			data = raw
			smaskData = alpha
			w = pw
			h = ph
		}
		colorSpace = "DeviceRGB"
	default:
		data = src.Data
		w = src.Width
		h = src.Height
		colorSpace = "DeviceRGB"
	}

	r.imageObjects[key] = imageObject{
		ResName:    resName,
		Data:       data,
		SmaskData:  smaskData,
		Width:      w,
		Height:     h,
		ColorSpace: colorSpace,
		Filter:     filter,
	}

	return resName
}

func overlayImageKey(data []byte) string {
	h := sha256.Sum256(data)
	return fmt.Sprintf("ov_%x", h[:8])
}

// RenderOverlayContent is a convenience function that takes document nodes,
// runs them through layout for a single page, then renders to overlay content.
func RenderOverlayContent(
	nodes []document.DocumentNode,
	pageSize document.Size,
	margins document.Edges,
	fonts map[string]*font.TrueTypeFont,
	fontDataMap map[string][]byte,
	fontResolver layout.FontResolver,
) (*OverlayResult, error) {
	// Create a single page document for layout.
	page := &document.Page{
		Size:    pageSize,
		Margins: margins,
		Content: nodes,
	}
	doc := &document.Document{
		Pages: []*document.Page{page},
		DefaultStyle: document.Style{
			FontFamily: "Helvetica",
			FontSize:   12,
			LineHeight: 1.2,
		},
	}

	// Run layout.
	paginator := layout.NewPaginator(pageSize, margins, fontResolver)
	layouts := paginator.Paginate(doc)
	layout.ResolvePageNumbers(layouts)

	if len(layouts) == 0 {
		return &OverlayResult{}, nil
	}

	// Render the first page's content.
	renderer := NewOverlayRenderer(pageSize.Width, pageSize.Height, fonts, fontDataMap)
	return renderer.RenderOverlay(layouts[0].Children)
}

// modifierSink adapts a [pdf.Modifier] to the objectSink interface used by
// the shared composite-font writer in pdftarget.go.
type modifierSink struct{ m *pdf.Modifier }

func (s modifierSink) AllocObject() pdf.ObjectRef { return s.m.AllocObject() }

func (s modifierSink) WriteObject(ref pdf.ObjectRef, obj pdf.Object) error {
	s.m.SetObject(ref, obj)
	return nil
}

// OverlayFontRegistry assigns one PDF font object per font family across all
// overlay operations on a document. Object references are handed out eagerly
// so pages can reference them, while the font objects themselves are written
// by Flush — after every page has been rendered, so the embedded subset covers
// all glyphs used anywhere in the document and is embedded exactly once.
type OverlayFontRegistry struct {
	refs  map[string]pdf.ObjectRef
	fonts map[string]fontObject
}

// NewOverlayFontRegistry creates an empty registry.
func NewOverlayFontRegistry() *OverlayFontRegistry {
	return &OverlayFontRegistry{
		refs:  make(map[string]pdf.ObjectRef),
		fonts: make(map[string]fontObject),
	}
}

// ref returns the object reference for a font, allocating it on first use.
func (reg *OverlayFontRegistry) ref(m *pdf.Modifier, fo fontObject) pdf.ObjectRef {
	if ref, ok := reg.refs[fo.Family]; ok {
		return ref
	}
	ref := m.AllocObject()
	reg.refs[fo.Family] = ref
	reg.fonts[fo.Family] = fo
	return ref
}

// Flush writes the font objects for every family registered so far.
func (reg *OverlayFontRegistry) Flush(m *pdf.Modifier) error {
	sink := modifierSink{m: m}
	for family, fo := range reg.fonts {
		ref := reg.refs[family]
		if fo.TTF != nil {
			if err := writeType0Font(sink, family, ref, fo.TTF, fo.Data); err != nil {
				return fmt.Errorf("write overlay font %q: %w", family, err)
			}
			continue
		}
		// Standard Type1 font — no embedded data.
		m.SetObject(ref, pdf.Dict{
			pdf.Name("Type"):     pdf.Name("Font"),
			pdf.Name("Subtype"):  pdf.Name("Type1"),
			pdf.Name("BaseFont"): pdf.Name(family),
		})
	}
	clear(reg.fonts)
	return nil
}

// WriteOverlayToModifier registers overlay fonts and images with the modifier,
// and returns the content bytes and a resource dict with the correct ObjectRefs.
// Fonts are embedded immediately; use [WriteOverlayToModifierWithFonts] with a
// shared [OverlayFontRegistry] when overlaying several pages of one document so
// that each font is embedded only once.
func WriteOverlayToModifier(result *OverlayResult, m *pdf.Modifier) ([]byte, *pdf.Dict, error) {
	reg := NewOverlayFontRegistry()
	content, resources, err := WriteOverlayToModifierWithFonts(result, m, reg)
	if err != nil {
		return nil, nil, err
	}
	if err := reg.Flush(m); err != nil {
		return nil, nil, err
	}
	return content, resources, nil
}

// WriteOverlayToModifierWithFonts is [WriteOverlayToModifier] with an explicit
// font registry. The caller is responsible for calling reg.Flush before the
// modified PDF is written out.
func WriteOverlayToModifierWithFonts(result *OverlayResult, m *pdf.Modifier, reg *OverlayFontRegistry) ([]byte, *pdf.Dict, error) {
	if result == nil || len(result.Content) == 0 {
		return nil, nil, nil
	}

	resources := make(pdf.Dict)

	// Register fonts.
	if len(result.FontObjects) > 0 {
		fontDict := make(pdf.Dict)
		for _, fo := range result.FontObjects {
			fontDict[pdf.Name(fo.ResName)] = reg.ref(m, fo)
		}
		resources[pdf.Name("Font")] = fontDict
	}

	// Register images.
	if len(result.ImageObjects) > 0 {
		xobjDict := make(pdf.Dict)
		for _, io := range result.ImageObjects {
			var smaskRef pdf.ObjectRef
			if len(io.SmaskData) > 0 {
				smaskRef = m.AllocObject()
				smaskContent, err := pdf.CompressFlate(io.SmaskData)
				if err != nil {
					return nil, nil, fmt.Errorf("compress smask: %w", err)
				}
				m.SetObject(smaskRef, pdf.Stream{
					Dict: pdf.Dict{
						pdf.Name("Type"):             pdf.Name("XObject"),
						pdf.Name("Subtype"):          pdf.Name("Image"),
						pdf.Name("Width"):            pdf.Integer(io.Width),
						pdf.Name("Height"):           pdf.Integer(io.Height),
						pdf.Name("ColorSpace"):       pdf.Name("DeviceGray"),
						pdf.Name("BitsPerComponent"): pdf.Integer(8),
						pdf.Name("Filter"):           pdf.Name("FlateDecode"),
					},
					Content: smaskContent,
				})
			}

			imgRef := m.AllocObject()
			imgDict := pdf.Dict{
				pdf.Name("Type"):             pdf.Name("XObject"),
				pdf.Name("Subtype"):          pdf.Name("Image"),
				pdf.Name("Width"):            pdf.Integer(io.Width),
				pdf.Name("Height"):           pdf.Integer(io.Height),
				pdf.Name("ColorSpace"):       pdf.Name(io.ColorSpace),
				pdf.Name("BitsPerComponent"): pdf.Integer(8),
			}

			if smaskRef.Number > 0 {
				imgDict[pdf.Name("SMask")] = smaskRef
			}

			content := io.Data
			switch io.Filter {
			case "DCTDecode":
				imgDict[pdf.Name("Filter")] = pdf.Name("DCTDecode")
			default:
				compressed, err := pdf.CompressFlate(content)
				if err != nil {
					return nil, nil, fmt.Errorf("compress image: %w", err)
				}
				imgDict[pdf.Name("Filter")] = pdf.Name("FlateDecode")
				content = compressed
			}

			m.SetObject(imgRef, pdf.Stream{
				Dict:    imgDict,
				Content: content,
			})

			xobjDict[pdf.Name(io.ResName)] = imgRef
		}
		resources[pdf.Name("XObject")] = xobjDict
	}

	if len(resources) == 0 {
		return result.Content, nil, nil
	}

	return result.Content, &resources, nil
}
