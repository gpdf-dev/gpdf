package template

import (
	"encoding/base64"
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/gpdf-dev/gpdf/barcode"
	"github.com/gpdf-dev/gpdf/document"
	"github.com/gpdf-dev/gpdf/pdf"
	"github.com/gpdf-dev/gpdf/qrcode"
)

// ---------------------------------------------------------------------------
// JSON Schema types
// ---------------------------------------------------------------------------

// Schema represents the top-level JSON structure for declarative PDF
// document definition. It maps directly to the builder API, allowing
// documents to be defined as JSON instead of Go code.
//
// Example JSON:
//
//	{
//	  "page": { "size": "A4", "margins": "15mm" },
//	  "body": [
//	    { "row": { "cols": [
//	      { "span": 12, "text": "Hello", "style": { "size": 24, "bold": true } }
//	    ]}}
//	  ]
//	}
type Schema struct {
	Page     SchemaPage       `json:"page"`
	Metadata *SchemaMeta      `json:"metadata,omitempty"`
	Header   []SchemaRow      `json:"header,omitempty"`
	Footer   []SchemaRow      `json:"footer,omitempty"`
	Body     []SchemaRow      `json:"body,omitempty"`
	Pages    []SchemaPageBody `json:"pages,omitempty"` // multiple explicit pages
	Absolute []SchemaAbsolute `json:"absolute,omitempty"`
}

// SchemaPageBody defines the body content for a single page.
type SchemaPageBody struct {
	Body     []SchemaRow      `json:"body"`
	Absolute []SchemaAbsolute `json:"absolute,omitempty"`
}

// SchemaPage defines page-level settings.
//
// Size accepts the ISO 216 / ISO 269 names "A0"-"A10", "B0"-"B10",
// "C0"-"C10", "DL", plus "Letter" and "Legal" (case-insensitive).
// Orientation "landscape" swaps the page to landscape; "portrait" (the
// default) keeps the size as is.
type SchemaPage struct {
	Size        string            `json:"size"`                  // e.g. "A4", "A6", "Letter"
	Orientation string            `json:"orientation,omitempty"` // "portrait" (default) or "landscape"
	Margins     string            `json:"margins,omitempty"`     // e.g., "15mm", "20pt"
	Background  *SchemaBackground `json:"background,omitempty"`  // drawn beneath every page
}

// SchemaBackground defines a full-page background drawn beneath the
// content of every page, equivalent to [WithBackgroundColor] and
// [WithBackgroundImage].
//
//	"background": {"image": "letterhead.png", "fit": "cover", "opacity": 0.3}
type SchemaBackground struct {
	Color   string  `json:"color,omitempty"`   // "#RRGGBB" or named
	Image   string  `json:"image,omitempty"`   // base64, data URI, or file path
	Fit     string  `json:"fit,omitempty"`     // "cover" (default)|"contain"|"stretch"|"original"
	Opacity float64 `json:"opacity,omitempty"` // 0.0–1.0; 0 means opaque
}

// SchemaMeta defines document metadata.
type SchemaMeta struct {
	Title   string `json:"title,omitempty"`
	Author  string `json:"author,omitempty"`
	Subject string `json:"subject,omitempty"`
	Creator string `json:"creator,omitempty"`
}

// SchemaRow wraps a single row definition.
type SchemaRow struct {
	Row SchemaRowDef `json:"row"`
}

// SchemaRowDef defines the height and columns of a row.
type SchemaRowDef struct {
	Height string      `json:"height,omitempty"` // "auto" or dimension e.g. "12mm"
	Cols   []SchemaCol `json:"cols"`
}

// SchemaCol defines a grid column with its span and content.
// Content can be specified as direct shorthand fields (text, image, etc.)
// or via the Elements array for multiple elements in one column.
type SchemaCol struct {
	Span int `json:"span"`

	// Shorthand: single element per column.
	Text    string         `json:"text,omitempty"`
	Image   *SchemaImage   `json:"image,omitempty"`
	Table   *SchemaTable   `json:"table,omitempty"`
	List    *SchemaList    `json:"list,omitempty"`
	Line    *SchemaLine    `json:"line,omitempty"`
	Spacer  string         `json:"spacer,omitempty"` // dimension string
	QRCode  *SchemaQRCode  `json:"qrcode,omitempty"`
	Barcode *SchemaBarcode `json:"barcode,omitempty"`
	Style   *SchemaStyle   `json:"style,omitempty"` // applies to text shorthand

	RichText *SchemaRichText `json:"richText,omitempty"`
	Box      *SchemaBox      `json:"box,omitempty"`

	// Multiple elements in one column.
	Elements []SchemaElement `json:"elements,omitempty"`
}

// SchemaElement defines a single content element within a column.
type SchemaElement struct {
	// Type selects the element kind: "text", "link", "richText", "image",
	// "table", "list", "line", "spacer", "qrcode", "barcode", "box",
	// "pageNumber", "totalPages".
	//
	// "link" is text that links to Style.Link and is drawn blue and
	// underlined by default, equivalent to [ColBuilder.Link].
	Type    string       `json:"type"`
	Content string       `json:"content,omitempty"` // for text
	Style   *SchemaStyle `json:"style,omitempty"`

	Image   *SchemaImage   `json:"image,omitempty"`
	Table   *SchemaTable   `json:"table,omitempty"`
	List    *SchemaList    `json:"list,omitempty"`
	Line    *SchemaLine    `json:"line,omitempty"`
	Height  string         `json:"height,omitempty"` // for spacer
	QRCode  *SchemaQRCode  `json:"qrcode,omitempty"`
	Barcode *SchemaBarcode `json:"barcode,omitempty"`

	RichText *SchemaRichText `json:"richText,omitempty"`
	Box      *SchemaBox      `json:"box,omitempty"`
}

// SchemaRichText defines a paragraph made of inline fragments with
// individual styles, equivalent to [ColBuilder.RichText].
//
// Style is the paragraph-level (block) style; it also serves as the base
// style of every fragment. Each fragment's Style is applied on top of it.
//
//	"richText": {
//	  "style": {"align": "left"},
//	  "fragments": [
//	    {"text": "Welcome to "},
//	    {"text": "gpdf", "style": {"bold": true, "color": "#007BFF"}}
//	  ]
//	}
type SchemaRichText struct {
	Style     *SchemaStyle             `json:"style,omitempty"`
	Fragments []SchemaRichTextFragment `json:"fragments"`
}

// SchemaRichTextFragment is a single inline span within a rich text
// paragraph.
type SchemaRichTextFragment struct {
	Text  string       `json:"text"`
	Style *SchemaStyle `json:"style,omitempty"`
}

// SchemaBox defines a styled container that stacks its elements
// vertically, equivalent to [ColBuilder.Box]. Boxes can be nested to build
// layouts that the flat row/column grid cannot express on its own.
//
// Padding / Margin are uniform values; Paddings / Margins override them
// with per-edge values in CSS order [top, right, bottom, left].
type SchemaBox struct {
	Width      string          `json:"width,omitempty"`
	Height     string          `json:"height,omitempty"`
	Padding    string          `json:"padding,omitempty"`
	Paddings   []string        `json:"paddings,omitempty"`
	Margin     string          `json:"margin,omitempty"`
	Margins    []string        `json:"margins,omitempty"`
	Background string          `json:"background,omitempty"`
	Border     *SchemaBorder   `json:"border,omitempty"`
	Elements   []SchemaElement `json:"elements,omitempty"`

	BorderRadius string  `json:"borderRadius,omitempty"` // e.g. "6pt"
	Opacity      float64 `json:"opacity,omitempty"`      // 0.0–1.0; 0 means opaque
	Link         string  `json:"link,omitempty"`         // URI the whole box links to
}

// SchemaStyle defines text styling properties.
//
// Padding is the uniform padding inside the text element (e.g. "10mm").
// Paddings overrides Padding with per-edge values in CSS order
// [top, right, bottom, left]. The padded area is included in the
// element's height and is filled by Background when set.
type SchemaStyle struct {
	Size          float64  `json:"size,omitempty"`
	Bold          bool     `json:"bold,omitempty"`
	Italic        bool     `json:"italic,omitempty"`
	Align         string   `json:"align,omitempty"`      // "left", "center", "right"
	Color         string   `json:"color,omitempty"`      // "#RRGGBB" or named
	Background    string   `json:"background,omitempty"` // "#RRGGBB" or named
	FontFamily    string   `json:"fontFamily,omitempty"`
	Underline     bool     `json:"underline,omitempty"`
	Strikethrough bool     `json:"strikethrough,omitempty"`
	LetterSpacing float64  `json:"letterSpacing,omitempty"`
	TextIndent    string   `json:"textIndent,omitempty"`   // e.g. "24pt", "10mm"
	Padding       string   `json:"padding,omitempty"`      // uniform, e.g. "10mm"
	Paddings      []string `json:"paddings,omitempty"`     // per-edge [top, right, bottom, left]
	BorderRadius  string   `json:"borderRadius,omitempty"` // rounds the background, e.g. "4pt"
	Opacity       float64  `json:"opacity,omitempty"`      // 0.0–1.0; 0 means opaque
	Link          string   `json:"link,omitempty"`         // URI the text links to
}

// SchemaImage defines an image element.
type SchemaImage struct {
	Src        string        `json:"src"`                  // base64, data URI, file path, or inline SVG markup
	Width      string        `json:"width,omitempty"`      // dimension
	Height     string        `json:"height,omitempty"`     // dimension
	MinWidth   string        `json:"minWidth,omitempty"`   // minimum display width; overflow to next page when violated
	MinHeight  string        `json:"minHeight,omitempty"`  // minimum display height; overflow to next page when violated
	Fit        string        `json:"fit,omitempty"`        // "contain"|"cover"|"stretch"|"original"
	Align      string        `json:"align,omitempty"`      // "left"|"center"|"right"
	Border     *SchemaBorder `json:"border,omitempty"`     // optional border around the image
	Background string        `json:"background,omitempty"` // "#RRGGBB" or named, drawn behind the image

	BorderRadius string  `json:"borderRadius,omitempty"` // clips the image to rounded corners
	Opacity      float64 `json:"opacity,omitempty"`      // 0.0–1.0; 0 means opaque
	Link         string  `json:"link,omitempty"`         // URI the image links to
}

// SchemaTable defines a table element.
type SchemaTable struct {
	Header         []string      `json:"header"`
	Rows           [][]string    `json:"rows"`
	ColumnWidths   []float64     `json:"columnWidths,omitempty"`
	ColumnAlign    []string      `json:"columnAlign,omitempty"` // per-column horizontal alignment: "left", "center", "right"
	HeaderStyle    *SchemaStyle  `json:"headerStyle,omitempty"`
	StripeColor    string        `json:"stripeColor,omitempty"`
	CellVAlign     string        `json:"cellVAlign,omitempty"` // "top", "middle", "bottom"
	Border         *SchemaBorder `json:"border,omitempty"`
	CellBorder     *SchemaBorder `json:"cellBorder,omitempty"` // applies to every header + body cell
	BorderCollapse *bool         `json:"borderCollapse,omitempty"`
	Background     string        `json:"background,omitempty"` // "#RRGGBB" or named
}

// SchemaBorder defines border styling for table/image/box elements.
//
// Width is the uniform edge width (e.g. "1pt", "2mm"). Use Widths for per-edge
// values in CSS order [top, right, bottom, left]. If both are set, Widths wins.
// Style accepts "solid" (default), "dashed", "dotted", or "none".
type SchemaBorder struct {
	Width  string   `json:"width,omitempty"`
	Widths []string `json:"widths,omitempty"`
	Color  string   `json:"color,omitempty"`
	Style  string   `json:"style,omitempty"`
}

// SchemaList defines a list element.
type SchemaList struct {
	Type  string   `json:"type,omitempty"` // "ordered" or "unordered" (default)
	Items []string `json:"items"`
}

// SchemaLine defines a horizontal line element.
type SchemaLine struct {
	Color     string `json:"color,omitempty"`
	Thickness string `json:"thickness,omitempty"`
}

// SchemaQRCode defines a QR code element.
type SchemaQRCode struct {
	Data            string `json:"data"`
	Size            string `json:"size,omitempty"`
	MinSize         string `json:"minSize,omitempty"`         // minimum display size; overflow to next page when violated
	ErrorCorrection string `json:"errorCorrection,omitempty"` // "L", "M", "Q", "H"
}

// SchemaBarcode defines a barcode element.
type SchemaBarcode struct {
	Data   string `json:"data"`
	Width  string `json:"width,omitempty"`
	Height string `json:"height,omitempty"`
	Format string `json:"format,omitempty"` // "code128" (default)
}

// SchemaAbsolute defines an absolute-positioned element placed at
// fixed XY coordinates on the page.
type SchemaAbsolute struct {
	X        string          `json:"x"`
	Y        string          `json:"y"`
	Width    string          `json:"width,omitempty"`
	Height   string          `json:"height,omitempty"`
	Origin   string          `json:"origin,omitempty"` // "content" (default) or "page"
	Elements []SchemaElement `json:"elements"`
}

// ---------------------------------------------------------------------------
// Parsing utilities
// ---------------------------------------------------------------------------

// parseValue parses a dimension string like "15mm", "12pt", "auto" into
// a document.Value. A bare number defaults to points.
func parseValue(s string) (document.Value, error) {
	s = strings.TrimSpace(s)
	if s == "" || s == "auto" {
		return document.Auto, nil
	}

	type unitSuffix struct {
		suffix string
		unit   document.Unit
	}
	// Longer suffixes first to avoid "m" matching before "mm".
	suffixes := []unitSuffix{
		{"mm", document.UnitMm},
		{"cm", document.UnitCm},
		{"in", document.UnitIn},
		{"pt", document.UnitPt},
		{"em", document.UnitEm},
		{"%", document.UnitPct},
	}

	for _, u := range suffixes {
		if strings.HasSuffix(s, u.suffix) {
			numStr := strings.TrimSpace(strings.TrimSuffix(s, u.suffix))
			v, err := strconv.ParseFloat(numStr, 64)
			if err != nil {
				return document.Value{}, fmt.Errorf("invalid value %q: %w", s, err)
			}
			return document.Value{Amount: v, Unit: u.unit}, nil
		}
	}

	// Plain number defaults to pt.
	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return document.Value{}, fmt.Errorf("invalid value %q", s)
	}
	return document.Pt(v), nil
}

// parsePageSize converts a page size name to document.Size.
func parsePageSize(s string) (document.Size, error) {
	if size, ok := pageSizes[strings.ToLower(strings.TrimSpace(s))]; ok {
		return size, nil
	}
	return document.Size{}, fmt.Errorf("unknown page size: %q", s)
}

// schemaPageSize resolves the page size and orientation of a SchemaPage.
// An empty size defaults to A4.
func schemaPageSize(p SchemaPage) (document.Size, error) {
	size := document.A4
	if p.Size != "" {
		var err error
		if size, err = parsePageSize(p.Size); err != nil {
			return document.Size{}, err
		}
	}
	switch strings.ToLower(strings.TrimSpace(p.Orientation)) {
	case "", "portrait":
		return size, nil
	case "landscape":
		return size.Landscape(), nil
	default:
		return document.Size{}, fmt.Errorf("unknown page orientation: %q", p.Orientation)
	}
}

// pageSizes maps lowercase page-size names accepted by the JSON schema to
// their dimensions.
var pageSizes = map[string]document.Size{
	"a0":     document.A0,
	"a1":     document.A1,
	"a2":     document.A2,
	"a3":     document.A3,
	"a4":     document.A4,
	"a5":     document.A5,
	"a6":     document.A6,
	"a7":     document.A7,
	"a8":     document.A8,
	"a9":     document.A9,
	"a10":    document.A10,
	"b0":     document.B0,
	"b1":     document.B1,
	"b2":     document.B2,
	"b3":     document.B3,
	"b4":     document.B4,
	"b5":     document.B5,
	"b6":     document.B6,
	"b7":     document.B7,
	"b8":     document.B8,
	"b9":     document.B9,
	"b10":    document.B10,
	"c0":     document.C0,
	"c1":     document.C1,
	"c2":     document.C2,
	"c3":     document.C3,
	"c4":     document.C4,
	"c5":     document.C5,
	"c6":     document.C6,
	"c7":     document.C7,
	"c8":     document.C8,
	"c9":     document.C9,
	"c10":    document.C10,
	"dl":     document.DL,
	"letter": document.Letter,
	"legal":  document.Legal,
}

// namedColors maps color name strings to their pdf.Color values.
var namedColors = map[string]pdf.Color{
	"black":   pdf.Black,
	"white":   pdf.White,
	"red":     pdf.Red,
	"green":   pdf.Green,
	"blue":    pdf.Blue,
	"yellow":  pdf.Yellow,
	"cyan":    pdf.Cyan,
	"magenta": pdf.Magenta,
}

// parseColor parses a color string into pdf.Color.
// Supported formats: "#RRGGBB" hex, "rgb(r,g,b)", "gray(v)", or named colors.
func parseColor(s string) (pdf.Color, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return pdf.Black, nil
	}

	lower := strings.ToLower(s)
	if c, ok := namedColors[lower]; ok {
		return c, nil
	}

	// gray(N) format: grayscale color.
	if strings.HasPrefix(lower, "gray(") && strings.HasSuffix(lower, ")") {
		valStr := lower[5 : len(lower)-1]
		val, err := strconv.ParseFloat(valStr, 64)
		if err != nil {
			return pdf.Color{}, fmt.Errorf("invalid gray color %q: %w", s, err)
		}
		return pdf.Gray(val), nil
	}

	// rgb(r, g, b) format: float RGB color (0.0-1.0).
	if strings.HasPrefix(lower, "rgb(") && strings.HasSuffix(lower, ")") {
		return parseRGBColor(s, lower)
	}

	// Hex color: #RRGGBB.
	if strings.HasPrefix(s, "#") && len(s) == 7 {
		hex, err := strconv.ParseUint(s[1:], 16, 32)
		if err != nil {
			return pdf.Color{}, fmt.Errorf("invalid color %q: %w", s, err)
		}
		return pdf.RGBHex(uint32(hex)), nil
	}

	return pdf.Color{}, fmt.Errorf("unknown color: %q", s)
}

// parseRGBColor parses an "rgb(r, g, b)" color string with float components (0.0-1.0).
func parseRGBColor(original, lower string) (pdf.Color, error) {
	inner := lower[4 : len(lower)-1]
	parts := strings.Split(inner, ",")
	if len(parts) != 3 {
		return pdf.Color{}, fmt.Errorf("invalid rgb color %q: expected 3 components", original)
	}
	r, err := strconv.ParseFloat(strings.TrimSpace(parts[0]), 64)
	if err != nil {
		return pdf.Color{}, fmt.Errorf("invalid rgb color %q: %w", original, err)
	}
	g, err := strconv.ParseFloat(strings.TrimSpace(parts[1]), 64)
	if err != nil {
		return pdf.Color{}, fmt.Errorf("invalid rgb color %q: %w", original, err)
	}
	b, err := strconv.ParseFloat(strings.TrimSpace(parts[2]), 64)
	if err != nil {
		return pdf.Color{}, fmt.Errorf("invalid rgb color %q: %w", original, err)
	}
	return pdf.RGB(r, g, b), nil
}

// parseAlignOption converts an alignment string to a TextOption.
func parseAlignOption(align string) TextOption {
	switch strings.ToLower(align) {
	case "center":
		return AlignCenter()
	case "right":
		return AlignRight()
	default:
		return AlignLeft()
	}
}

// applySchemaStyle converts a SchemaStyle to a slice of TextOption.
func applySchemaStyle(ss *SchemaStyle) []TextOption {
	if ss == nil {
		return nil
	}
	var opts []TextOption
	if ss.Size > 0 {
		opts = append(opts, FontSize(ss.Size))
	}
	if ss.Bold {
		opts = append(opts, Bold())
	}
	if ss.Italic {
		opts = append(opts, Italic())
	}
	if ss.Align != "" {
		opts = append(opts, parseAlignOption(ss.Align))
	}
	opts = appendColorOpt(opts, ss.Color, TextColor)
	opts = appendColorOpt(opts, ss.Background, BgColor)
	if ss.FontFamily != "" {
		opts = append(opts, FontFamily(ss.FontFamily))
	}
	if ss.Underline {
		opts = append(opts, Underline())
	}
	if ss.Strikethrough {
		opts = append(opts, Strikethrough())
	}
	if ss.LetterSpacing != 0 {
		opts = append(opts, LetterSpacing(ss.LetterSpacing))
	}
	if ss.TextIndent != "" {
		if v, err := parseValue(ss.TextIndent); err == nil {
			opts = append(opts, TextIndent(v))
		}
	}
	if e, ok := parseSchemaPadding(ss); ok {
		opts = append(opts, TextPadding(e))
	}
	return appendSchemaTextEffects(opts, ss)
}

// appendSchemaTextEffects appends the border-radius, opacity and link
// options of a SchemaStyle.
func appendSchemaTextEffects(opts []TextOption, ss *SchemaStyle) []TextOption {
	if v, ok := parseOptionalValue(ss.BorderRadius); ok {
		opts = append(opts, WithTextRadius(v))
	}
	if ss.Opacity > 0 {
		opts = append(opts, WithTextOpacity(ss.Opacity))
	}
	if ss.Link != "" {
		opts = append(opts, WithTextLink(ss.Link))
	}
	return opts
}

// parseOptionalValue parses a dimension string, reporting false when it is
// empty or invalid.
func parseOptionalValue(s string) (document.Value, bool) {
	if s == "" {
		return document.Value{}, false
	}
	v, err := parseValue(s)
	if err != nil {
		return document.Value{}, false
	}
	return v, true
}

// parseSchemaPadding resolves [SchemaStyle.Padding] / [SchemaStyle.Paddings]
// into [document.Edges]. Paddings, when present, overrides Padding and
// follows the CSS [top, right, bottom, left] order.
func parseSchemaPadding(ss *SchemaStyle) (document.Edges, bool) {
	return parseSchemaEdges(ss.Padding, ss.Paddings)
}

// parseSchemaEdges resolves a uniform value and an optional per-edge list
// (CSS [top, right, bottom, left] order) into [document.Edges]. The list,
// when present, takes precedence over the uniform value.
func parseSchemaEdges(uniform string, perEdge []string) (document.Edges, bool) {
	if len(perEdge) > 0 {
		edges, err := parseEdgeList(perEdge)
		if err != nil {
			return document.Edges{}, false
		}
		return edges, true
	}
	if uniform != "" {
		v, err := parseValue(uniform)
		if err != nil {
			return document.Edges{}, false
		}
		return document.UniformEdges(v), true
	}
	return document.Edges{}, false
}

// parseEdgeList converts a CSS-style 1/2/3/4 element shorthand into
// per-edge [document.Edges]. The mapping mirrors the CSS padding
// shorthand: [top right bottom left] for 4 values, [top horizontal bottom]
// for 3, [vertical horizontal] for 2, and [all] for 1.
func parseEdgeList(values []string) (document.Edges, error) {
	parsed := make([]document.Value, len(values))
	for i, s := range values {
		v, err := parseValue(s)
		if err != nil {
			return document.Edges{}, fmt.Errorf("edge %d: %w", i, err)
		}
		parsed[i] = v
	}
	switch len(parsed) {
	case 1:
		return document.UniformEdges(parsed[0]), nil
	case 2:
		return document.Edges{Top: parsed[0], Right: parsed[1], Bottom: parsed[0], Left: parsed[1]}, nil
	case 3:
		return document.Edges{Top: parsed[0], Right: parsed[1], Bottom: parsed[2], Left: parsed[1]}, nil
	case 4:
		return document.Edges{Top: parsed[0], Right: parsed[1], Bottom: parsed[2], Left: parsed[3]}, nil
	default:
		return document.Edges{}, fmt.Errorf("expected 1-4 values, got %d", len(parsed))
	}
}

// appendColorOpt parses a color string and appends the resulting option if valid.
func appendColorOpt(opts []TextOption, s string, fn func(pdf.Color) TextOption) []TextOption {
	if s == "" {
		return opts
	}
	if c, err := parseColor(s); err == nil {
		opts = append(opts, fn(c))
	}
	return opts
}

// parseBorderLineStyle converts a string into a [document.BorderStyle].
// Empty string returns BorderSolid (matches the [Border] default).
func parseBorderLineStyle(s string) (document.BorderStyle, bool) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "", "solid":
		return document.BorderSolid, true
	case "dashed":
		return document.BorderDashed, true
	case "dotted":
		return document.BorderDotted, true
	case "none":
		return document.BorderNone, true
	default:
		return document.BorderSolid, false
	}
}

// parseSchemaBorder turns a SchemaBorder into a [BorderSpec]. Returns ok=false
// when the schema is nil or the width/widths cannot be parsed.
func parseSchemaBorder(b *SchemaBorder) (BorderSpec, bool) {
	if b == nil {
		return BorderSpec{}, false
	}
	var bopts []BorderOption
	switch {
	case len(b.Widths) == 4:
		var edges [4]document.Value
		for i, w := range b.Widths {
			v, err := parseValue(w)
			if err != nil {
				return BorderSpec{}, false
			}
			edges[i] = v
		}
		bopts = append(bopts, BorderWidths(edges[0], edges[1], edges[2], edges[3]))
	case b.Width != "":
		v, err := parseValue(b.Width)
		if err != nil {
			return BorderSpec{}, false
		}
		bopts = append(bopts, BorderWidth(v))
	}
	if b.Color != "" {
		if c, err := parseColor(b.Color); err == nil {
			bopts = append(bopts, BorderColor(c))
		}
	}
	if b.Style != "" {
		if st, ok := parseBorderLineStyle(b.Style); ok {
			bopts = append(bopts, BorderLine(st))
		}
	}
	return Border(bopts...), true
}

// decodeBase64Image decodes a base64-encoded image string.
// Supports both raw base64 and data URI format (data:image/...;base64,...).
func decodeBase64Image(s string) ([]byte, error) {
	if strings.HasPrefix(s, "data:") {
		idx := strings.Index(s, ",")
		if idx < 0 {
			return nil, fmt.Errorf("invalid data URI")
		}
		header, payload := s[:idx], s[idx+1:]
		if !strings.Contains(header, ";base64") {
			// Non-base64 data URI, e.g. "data:image/svg+xml;utf8,<svg …>".
			text, err := url.PathUnescape(payload)
			if err != nil {
				return nil, fmt.Errorf("invalid data URI: %w", err)
			}
			return []byte(text), nil
		}
		s = payload
	}
	return base64.StdEncoding.DecodeString(s)
}

// loadImageData resolves the image source string to raw image bytes.
// It supports data URIs (base64 or URL-encoded), file:// URIs, file paths,
// raw base64 strings, and inline SVG markup (a string starting with "<").
func loadImageData(src string) ([]byte, error) {
	// inline SVG markup
	if trimmed := strings.TrimSpace(src); strings.HasPrefix(trimmed, "<") {
		return []byte(trimmed), nil
	}
	// data URI
	if strings.HasPrefix(src, "data:") {
		return decodeBase64Image(src)
	}
	// file URI
	if strings.HasPrefix(src, "file://") {
		return os.ReadFile(strings.TrimPrefix(src, "file://"))
	}
	// relative file path (unambiguous)
	if strings.HasPrefix(src, "./") || strings.HasPrefix(src, "../") {
		return os.ReadFile(src)
	}
	// Windows drive letter (e.g., "C:\...")
	if len(src) >= 3 && src[1] == ':' && (src[2] == '/' || src[2] == '\\') {
		return os.ReadFile(src)
	}
	// For absolute paths starting with /, try file first, then base64.
	// This handles JPEG base64 strings that start with "/9j/...".
	if strings.HasPrefix(src, "/") {
		if data, err := os.ReadFile(src); err == nil {
			return data, nil
		}
		// Not a valid file path, try base64.
		return decodeBase64Image(src)
	}
	// fallback: raw base64
	return decodeBase64Image(src)
}

// isFilePath returns true if the string looks like a file system path
// rather than a base64-encoded string.
func isFilePath(s string) bool {
	if strings.HasPrefix(s, "/") || strings.HasPrefix(s, "./") || strings.HasPrefix(s, "../") {
		return true
	}
	// Windows drive letter (e.g., "C:\...")
	if len(s) >= 3 && s[1] == ':' && (s[2] == '/' || s[2] == '\\') {
		return true
	}
	return false
}

// ---------------------------------------------------------------------------
// Schema → Document builder
// ---------------------------------------------------------------------------

// buildFromSchema constructs a Document from a parsed Schema.
func buildFromSchema(schema *Schema, opts []Option) (*Document, error) {
	var docOpts []Option

	if schema.Page.Size != "" || schema.Page.Orientation != "" {
		size, err := schemaPageSize(schema.Page)
		if err != nil {
			return nil, err
		}
		docOpts = append(docOpts, WithPageSize(size))
	}

	if schema.Page.Margins != "" {
		v, err := parseValue(schema.Page.Margins)
		if err != nil {
			return nil, fmt.Errorf("invalid margins: %w", err)
		}
		docOpts = append(docOpts, WithMargins(document.UniformEdges(v)))
	}

	bgOpts, err := schemaBackgroundOptions(schema.Page.Background)
	if err != nil {
		return nil, err
	}
	docOpts = append(docOpts, bgOpts...)

	if schema.Metadata != nil {
		docOpts = append(docOpts, WithMetadata(document.DocumentMetadata{
			Title:   schema.Metadata.Title,
			Author:  schema.Metadata.Author,
			Subject: schema.Metadata.Subject,
			Creator: schema.Metadata.Creator,
		}))
	}

	// User-provided options override schema-level settings.
	docOpts = append(docOpts, opts...)

	doc := New(docOpts...)

	if len(schema.Header) > 0 {
		rows := schema.Header // capture for closure
		doc.Header(func(p *PageBuilder) {
			buildSchemaRows(p, rows)
		})
	}

	if len(schema.Footer) > 0 {
		rows := schema.Footer // capture for closure
		doc.Footer(func(p *PageBuilder) {
			buildSchemaRows(p, rows)
		})
	}

	if len(schema.Body) > 0 {
		page := doc.AddPage()
		buildSchemaRows(page, schema.Body)
		buildSchemaAbsolutes(page, schema.Absolute)
	}

	for _, p := range schema.Pages {
		page := doc.AddPage()
		buildSchemaRows(page, p.Body)
		buildSchemaAbsolutes(page, p.Absolute)
	}

	return doc, nil
}

// schemaBackgroundOptions converts a SchemaBackground into document
// options. Unlike element-level fields, an unreadable background image or
// invalid color is reported as an error because it affects every page.
func schemaBackgroundOptions(bg *SchemaBackground) ([]Option, error) {
	if bg == nil {
		return nil, nil
	}
	var opts []Option
	if bg.Color != "" {
		col, err := parseColor(bg.Color)
		if err != nil {
			return nil, fmt.Errorf("invalid page background color: %w", err)
		}
		opts = append(opts, WithBackgroundColor(col))
	}
	var bgOpts []BackgroundOption
	if bg.Fit != "" {
		mode, ok := parseFitMode(bg.Fit)
		if !ok {
			return nil, fmt.Errorf("invalid page background fit: %q", bg.Fit)
		}
		bgOpts = append(bgOpts, BackgroundFit(mode))
	}
	if bg.Opacity > 0 {
		bgOpts = append(bgOpts, BackgroundOpacity(bg.Opacity))
	}
	if bg.Image != "" {
		data, err := loadImageData(bg.Image)
		if err != nil {
			return nil, fmt.Errorf("invalid page background image: %w", err)
		}
		opts = append(opts, WithBackgroundImage(data, bgOpts...))
	} else if bg.Opacity > 0 {
		// Opacity also applies to a color-only background.
		opts = append(opts, func(c *Config) { BackgroundOpacity(bg.Opacity)(c.ensureBackground()) })
	}
	return opts, nil
}

// buildSchemaRows adds rows from the schema to a PageBuilder.
func buildSchemaRows(p *PageBuilder, rows []SchemaRow) {
	for _, sr := range rows {
		cols := sr.Row.Cols
		height := sr.Row.Height

		if height == "" || height == "auto" {
			p.AutoRow(func(r *RowBuilder) {
				buildSchemaCols(r, cols)
			})
		} else {
			v, err := parseValue(height)
			if err != nil {
				continue // skip rows with invalid height
			}
			p.Row(v, func(r *RowBuilder) {
				buildSchemaCols(r, cols)
			})
		}
	}
}

// buildSchemaCols adds columns from the schema to a RowBuilder.
func buildSchemaCols(r *RowBuilder, cols []SchemaCol) {
	for _, sc := range cols {
		col := sc // capture for closure
		r.Col(col.Span, func(c *ColBuilder) {
			buildSchemaColContent(c, col)
		})
	}
}

// buildSchemaColContent adds content to a column from its schema definition.
func buildSchemaColContent(c *ColBuilder, col SchemaCol) {
	// If elements array is provided, use it.
	if len(col.Elements) > 0 {
		for _, elem := range col.Elements {
			buildSchemaElement(c, elem)
		}
		return
	}

	// Shorthand: direct properties on the column.
	if col.Text != "" {
		opts := applySchemaStyle(col.Style)
		c.Text(col.Text, opts...)
	}
	if col.Image != nil {
		buildSchemaImage(c, col.Image)
	}
	if col.Table != nil {
		buildSchemaTable(c, col.Table)
	}
	if col.List != nil {
		buildSchemaList(c, col.List)
	}
	if col.Line != nil {
		buildSchemaLine(c, col.Line)
	}
	if col.Spacer != "" {
		if v, err := parseValue(col.Spacer); err == nil {
			c.Spacer(v)
		}
	}
	if col.QRCode != nil {
		buildSchemaQRCode(c, col.QRCode)
	}
	if col.Barcode != nil {
		buildSchemaBarcode(c, col.Barcode)
	}
	buildSchemaRichText(c, col.RichText)
	buildSchemaBox(c, col.Box)
}

// buildSchemaElement adds a single element to a ColBuilder.
func buildSchemaElement(c *ColBuilder, elem SchemaElement) {
	switch strings.ToLower(elem.Type) {
	case "text", "link":
		buildSchemaText(c, elem)
	case "image":
		buildSchemaImage(c, elem.Image)
	case "table":
		buildSchemaTable(c, elem.Table)
	case "list":
		buildSchemaList(c, elem.List)
	case "line":
		buildSchemaLine(c, elem.Line)
	case "spacer":
		if v, err := parseValue(elem.Height); err == nil {
			c.Spacer(v)
		}
	case "qrcode":
		buildSchemaQRCode(c, elem.QRCode)
	case "barcode":
		buildSchemaBarcode(c, elem.Barcode)
	case "richtext":
		buildSchemaRichText(c, elem.RichText)
	case "box":
		buildSchemaBox(c, elem.Box)
	case "pagenumber":
		opts := applySchemaStyle(elem.Style)
		c.PageNumber(opts...)
	case "totalpages":
		opts := applySchemaStyle(elem.Style)
		c.TotalPages(opts...)
	}
}

// buildSchemaText adds a "text" or "link" element to a ColBuilder. A link
// takes its URI from Style.Link.
func buildSchemaText(c *ColBuilder, elem SchemaElement) {
	opts := applySchemaStyle(elem.Style)
	if strings.EqualFold(elem.Type, "link") && elem.Style != nil && elem.Style.Link != "" {
		c.Link(elem.Content, elem.Style.Link, opts...)
		return
	}
	c.Text(elem.Content, opts...)
}

// buildSchemaRichText adds a rich text paragraph to a ColBuilder.
func buildSchemaRichText(c *ColBuilder, rt *SchemaRichText) {
	if rt == nil || len(rt.Fragments) == 0 {
		return
	}
	c.RichText(func(b *RichTextBuilder) {
		for _, f := range rt.Fragments {
			b.Span(f.Text, applySchemaStyle(f.Style)...)
		}
	}, applySchemaStyle(rt.Style)...)
}

// buildSchemaBox adds a styled container with nested elements to a
// ColBuilder.
func buildSchemaBox(c *ColBuilder, sb *SchemaBox) {
	if sb == nil {
		return
	}
	c.Box(func(inner *ColBuilder) {
		for _, elem := range sb.Elements {
			buildSchemaElement(inner, elem)
		}
	}, schemaBoxOptions(sb)...)
}

// schemaBoxOptions converts the styling fields of a SchemaBox into
// BoxOptions. Invalid values are skipped, matching the rest of the schema.
func schemaBoxOptions(sb *SchemaBox) []BoxOption {
	var opts []BoxOption
	if v, err := parseValue(sb.Width); sb.Width != "" && err == nil {
		opts = append(opts, WithBoxWidth(v))
	}
	if v, err := parseValue(sb.Height); sb.Height != "" && err == nil {
		opts = append(opts, WithBoxHeight(v))
	}
	if e, ok := parseSchemaEdges(sb.Padding, sb.Paddings); ok {
		opts = append(opts, WithBoxPadding(e))
	}
	if e, ok := parseSchemaEdges(sb.Margin, sb.Margins); ok {
		opts = append(opts, WithBoxMargin(e))
	}
	if sb.Background != "" {
		if col, err := parseColor(sb.Background); err == nil {
			opts = append(opts, WithBoxBackground(col))
		}
	}
	if spec, ok := parseSchemaBorder(sb.Border); ok {
		opts = append(opts, WithBoxBorder(spec))
	}
	if v, ok := parseOptionalValue(sb.BorderRadius); ok {
		opts = append(opts, WithBoxRadius(v))
	}
	if sb.Opacity > 0 {
		opts = append(opts, WithBoxOpacity(sb.Opacity))
	}
	if sb.Link != "" {
		opts = append(opts, WithBoxLink(sb.Link))
	}
	return opts
}

func buildSchemaImage(c *ColBuilder, img *SchemaImage) {
	if img == nil {
		return
	}
	data, err := loadImageData(img.Src)
	if err != nil {
		return // silently skip, consistent with builder API pattern
	}
	opts := schemaImageDimensionOpts(img)
	if img.Fit != "" {
		if mode, ok := parseFitMode(img.Fit); ok {
			opts = append(opts, WithFitMode(mode))
		}
	}
	if img.Align != "" {
		if align, ok := parseImageAlign(img.Align); ok {
			opts = append(opts, WithAlign(align))
		}
	}
	if spec, ok := parseSchemaBorder(img.Border); ok {
		opts = append(opts, WithImageBorder(spec))
	}
	if img.Background != "" {
		if c2, err := parseColor(img.Background); err == nil {
			opts = append(opts, WithImageBackground(c2))
		}
	}
	c.Image(data, append(opts, schemaImageEffectOpts(img)...)...)
}

// schemaImageEffectOpts converts the border-radius, opacity and link fields
// of a SchemaImage into ImageOptions.
func schemaImageEffectOpts(img *SchemaImage) []ImageOption {
	var opts []ImageOption
	if v, ok := parseOptionalValue(img.BorderRadius); ok {
		opts = append(opts, WithImageRadius(v))
	}
	if img.Opacity > 0 {
		opts = append(opts, WithImageOpacity(img.Opacity))
	}
	if img.Link != "" {
		opts = append(opts, WithImageLink(img.Link))
	}
	return opts
}

// schemaImageDimensionOpts converts the dimension fields of a SchemaImage
// (width, height, minWidth, minHeight) into the matching ImageOption slice.
func schemaImageDimensionOpts(img *SchemaImage) []ImageOption {
	dims := []struct {
		raw string
		fn  func(document.Value) ImageOption
	}{
		{img.Width, FitWidth},
		{img.Height, FitHeight},
		{img.MinWidth, MinDisplayWidth},
		{img.MinHeight, MinDisplayHeight},
	}
	var opts []ImageOption
	for _, d := range dims {
		if d.raw == "" {
			continue
		}
		v, err := parseValue(d.raw)
		if err != nil {
			continue
		}
		opts = append(opts, d.fn(v))
	}
	return opts
}

// parseFitMode converts a fit mode string to an ImageFitMode constant.
func parseFitMode(s string) (document.ImageFitMode, bool) {
	switch strings.ToLower(s) {
	case "contain":
		return document.FitContain, true
	case "cover":
		return document.FitCover, true
	case "stretch":
		return document.FitStretch, true
	case "original":
		return document.FitOriginal, true
	default:
		return document.FitContain, false
	}
}

// parseVerticalAlign converts a vertical alignment string to a VerticalAlign constant.
func parseVerticalAlign(s string) (document.VerticalAlign, bool) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "top":
		return document.VAlignTop, true
	case "middle":
		return document.VAlignMiddle, true
	case "bottom":
		return document.VAlignBottom, true
	default:
		return document.VAlignTop, false
	}
}

// parseImageAlign converts an alignment string to a TextAlign constant.
func parseImageAlign(s string) (document.TextAlign, bool) {
	switch strings.ToLower(s) {
	case "left":
		return document.AlignLeft, true
	case "center":
		return document.AlignCenter, true
	case "right":
		return document.AlignRight, true
	default:
		return document.AlignLeft, false
	}
}

// parseSchemaColumnAligns converts per-column alignment strings into TextAlign
// values. Unrecognized entries fall back to AlignLeft. Returns nil for an
// empty input so callers can skip applying the option entirely.
func parseSchemaColumnAligns(s []string) []document.TextAlign {
	if len(s) == 0 {
		return nil
	}
	aligns := make([]document.TextAlign, len(s))
	for i, a := range s {
		if align, ok := parseImageAlign(a); ok {
			aligns[i] = align
		}
	}
	return aligns
}

func buildSchemaTable(c *ColBuilder, tbl *SchemaTable) {
	if tbl == nil {
		return
	}
	var opts []TableOption
	if len(tbl.ColumnWidths) > 0 {
		opts = append(opts, ColumnWidths(tbl.ColumnWidths...))
	}
	if tbl.HeaderStyle != nil {
		textOpts := applySchemaStyle(tbl.HeaderStyle)
		opts = append(opts, TableHeaderStyle(textOpts...))
	}
	if tbl.StripeColor != "" {
		if clr, err := parseColor(tbl.StripeColor); err == nil {
			opts = append(opts, TableStripe(clr))
		}
	}
	if tbl.CellVAlign != "" {
		if align, ok := parseVerticalAlign(tbl.CellVAlign); ok {
			opts = append(opts, TableCellVAlign(align))
		}
	}
	if aligns := parseSchemaColumnAligns(tbl.ColumnAlign); len(aligns) > 0 {
		opts = append(opts, ColumnAlign(aligns...))
	}
	if spec, ok := parseSchemaBorder(tbl.Border); ok {
		opts = append(opts, WithTableBorder(spec))
	}
	if spec, ok := parseSchemaBorder(tbl.CellBorder); ok {
		opts = append(opts, WithTableCellBorder(spec))
	}
	if tbl.BorderCollapse != nil {
		opts = append(opts, WithTableBorderCollapse(*tbl.BorderCollapse))
	}
	if tbl.Background != "" {
		if c2, err := parseColor(tbl.Background); err == nil {
			opts = append(opts, WithTableBackground(c2))
		}
	}
	c.Table(tbl.Header, tbl.Rows, opts...)
}

func buildSchemaList(c *ColBuilder, lst *SchemaList) {
	if lst == nil {
		return
	}
	if strings.ToLower(lst.Type) == "ordered" {
		c.OrderedList(lst.Items)
	} else {
		c.List(lst.Items)
	}
}

func buildSchemaLine(c *ColBuilder, ln *SchemaLine) {
	if ln == nil {
		c.Line()
		return
	}
	var opts []LineOption
	if ln.Color != "" {
		if clr, err := parseColor(ln.Color); err == nil {
			opts = append(opts, LineColor(clr))
		}
	}
	if ln.Thickness != "" {
		if v, err := parseValue(ln.Thickness); err == nil {
			opts = append(opts, LineThickness(v))
		}
	}
	c.Line(opts...)
}

func buildSchemaQRCode(c *ColBuilder, qr *SchemaQRCode) {
	if qr == nil {
		return
	}
	var opts []QRCodeOption
	if qr.Size != "" {
		if v, err := parseValue(qr.Size); err == nil {
			opts = append(opts, QRSize(v))
		}
	}
	if qr.MinSize != "" {
		if v, err := parseValue(qr.MinSize); err == nil {
			opts = append(opts, QRMinSize(v))
		}
	}
	if qr.ErrorCorrection != "" {
		if level, ok := parseQRErrorCorrection(qr.ErrorCorrection); ok {
			opts = append(opts, QRErrorCorrection(level))
		}
	}
	c.QRCode(qr.Data, opts...)
}

// parseQRErrorCorrection converts an error correction string to a qrcode level.
func parseQRErrorCorrection(s string) (qrcode.ErrorCorrectionLevel, bool) {
	switch strings.ToUpper(strings.TrimSpace(s)) {
	case "L":
		return qrcode.LevelL, true
	case "M":
		return qrcode.LevelM, true
	case "Q":
		return qrcode.LevelQ, true
	case "H":
		return qrcode.LevelH, true
	default:
		return qrcode.LevelM, false
	}
}

func buildSchemaBarcode(c *ColBuilder, bc *SchemaBarcode) {
	if bc == nil {
		return
	}
	var opts []BarcodeOption
	if bc.Width != "" {
		if v, err := parseValue(bc.Width); err == nil {
			opts = append(opts, BarcodeWidth(v))
		}
	}
	if bc.Height != "" {
		if v, err := parseValue(bc.Height); err == nil {
			opts = append(opts, BarcodeHeight(v))
		}
	}
	if bc.Format != "" {
		switch strings.ToLower(bc.Format) {
		case "code128":
			opts = append(opts, BarcodeFormat(barcode.Code128))
		}
	}
	c.Barcode(bc.Data, opts...)
}

// buildSchemaAbsolutes adds absolute-positioned elements to a page.
func buildSchemaAbsolutes(p *PageBuilder, absolutes []SchemaAbsolute) {
	for _, abs := range absolutes {
		xVal, err := parseValue(abs.X)
		if err != nil {
			continue
		}
		yVal, err := parseValue(abs.Y)
		if err != nil {
			continue
		}

		var opts []AbsoluteOption
		if abs.Width != "" {
			if v, err := parseValue(abs.Width); err == nil {
				opts = append(opts, AbsoluteWidth(v))
			}
		}
		if abs.Height != "" {
			if v, err := parseValue(abs.Height); err == nil {
				opts = append(opts, AbsoluteHeight(v))
			}
		}
		if strings.ToLower(abs.Origin) == "page" {
			opts = append(opts, AbsoluteOriginPage())
		}

		elements := abs.Elements // capture for closure
		p.Absolute(xVal, yVal, func(c *ColBuilder) {
			for _, elem := range elements {
				buildSchemaElement(c, elem)
			}
		}, opts...)
	}
}
