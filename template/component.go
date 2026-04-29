package template

import (
	"github.com/gpdf-dev/gpdf/barcode"
	"github.com/gpdf-dev/gpdf/document"
	"github.com/gpdf-dev/gpdf/pdf"
	"github.com/gpdf-dev/gpdf/qrcode"
)

// --- Text Options ---

// TextOption configures a Text element.
type TextOption func(*document.Style)

// FontSize sets the font size in points.
func FontSize(size float64) TextOption {
	return func(s *document.Style) { s.FontSize = size }
}

// Bold sets the font weight to bold.
func Bold() TextOption {
	return func(s *document.Style) { s.FontWeight = document.WeightBold }
}

// Italic sets the font style to italic.
func Italic() TextOption {
	return func(s *document.Style) { s.FontStyle = document.StyleItalic }
}

// TextColor sets the text foreground color.
func TextColor(c pdf.Color) TextOption {
	return func(s *document.Style) { s.Color = c }
}

// BgColor sets the background color.
func BgColor(c pdf.Color) TextOption {
	return func(s *document.Style) { s.Background = &c }
}

// AlignLeft sets left text alignment.
func AlignLeft() TextOption {
	return func(s *document.Style) { s.TextAlign = document.AlignLeft }
}

// AlignCenter sets center text alignment.
func AlignCenter() TextOption {
	return func(s *document.Style) { s.TextAlign = document.AlignCenter }
}

// AlignRight sets right text alignment.
func AlignRight() TextOption {
	return func(s *document.Style) { s.TextAlign = document.AlignRight }
}

// FontFamily sets the font family name.
func FontFamily(family string) TextOption {
	return func(s *document.Style) { s.FontFamily = family }
}

// LetterSpacing sets the extra space between characters in points.
func LetterSpacing(pts float64) TextOption {
	return func(s *document.Style) { s.LetterSpacing = pts }
}

// TextIndent sets the first-line indentation.
func TextIndent(v document.Value) TextOption {
	return func(s *document.Style) { s.TextIndent = v }
}

// Underline adds underline decoration to text.
func Underline() TextOption {
	return func(s *document.Style) { s.TextDecoration |= document.DecorationUnderline }
}

// Strikethrough adds strikethrough decoration to text.
func Strikethrough() TextOption {
	return func(s *document.Style) { s.TextDecoration |= document.DecorationStrikethrough }
}

// --- Image Options ---

// ImageOption configures an Image element.
type ImageOption func(*imageConfig)

type imageConfig struct {
	width      document.Value
	height     document.Value
	minWidth   document.Value
	minHeight  document.Value
	fitMode    document.ImageFitMode
	align      document.TextAlign
	border     *BorderSpec
	background *pdf.Color
}

// FitWidth sets the image to fit within the specified width.
func FitWidth(width document.Value) ImageOption {
	return func(cfg *imageConfig) {
		cfg.width = width
		cfg.fitMode = document.FitContain
	}
}

// FitHeight sets the image to fit within the specified height.
func FitHeight(height document.Value) ImageOption {
	return func(cfg *imageConfig) {
		cfg.height = height
		cfg.fitMode = document.FitContain
	}
}

// WithFitMode sets the image fit mode.
func WithFitMode(mode document.ImageFitMode) ImageOption {
	return func(cfg *imageConfig) {
		cfg.fitMode = mode
	}
}

// WithAlign sets the horizontal alignment of the image within its column.
func WithAlign(align document.TextAlign) ImageOption {
	return func(cfg *imageConfig) {
		cfg.align = align
	}
}

// MinDisplayWidth sets a minimum display width for the image. If the layout
// engine would need to shrink the image below this width to fit the remaining
// space, the image is moved to the next page instead.
func MinDisplayWidth(width document.Value) ImageOption {
	return func(cfg *imageConfig) {
		cfg.minWidth = width
	}
}

// MinDisplayHeight sets a minimum display height for the image. If the layout
// engine would need to shrink the image below this height to fit the remaining
// space, the image is moved to the next page instead.
func MinDisplayHeight(height document.Value) ImageOption {
	return func(cfg *imageConfig) {
		cfg.minHeight = height
	}
}

// WithImageBorder draws a border around the image using the given [BorderSpec].
// Build a spec with [Border] and [BorderWidth], [BorderColor], etc.
func WithImageBorder(spec BorderSpec) ImageOption {
	return func(cfg *imageConfig) {
		cfg.border = &spec
	}
}

// WithImageBackground fills the image's box with the given color before the
// image is drawn. Useful for transparent PNGs that need a solid backdrop.
func WithImageBackground(c pdf.Color) ImageOption {
	return func(cfg *imageConfig) {
		cfg.background = &c
	}
}

// --- Table Options ---

// TableOption configures a Table element.
type TableOption func(*tableConfig)

type tableConfig struct {
	headerBgColor   *pdf.Color
	headerTextColor *pdf.Color
	stripeColor     *pdf.Color
	columnWidths    []float64
	cellVAlign      document.VerticalAlign
	hasCellVAlign   bool
	border          *BorderSpec
	cellBorder      *BorderSpec
	background      *pdf.Color
	borderCollapse  bool
	hasCollapse     bool
}

// TableHeaderStyle sets the header background and text color.
func TableHeaderStyle(opts ...TextOption) TableOption {
	return func(cfg *tableConfig) {
		s := document.DefaultStyle()
		for _, opt := range opts {
			opt(&s)
		}
		if s.Background != nil {
			cfg.headerBgColor = s.Background
		}
		cfg.headerTextColor = &s.Color
	}
}

// TableStripe sets the background color for alternating rows.
func TableStripe(c pdf.Color) TableOption {
	return func(cfg *tableConfig) {
		cfg.stripeColor = &c
	}
}

// ColumnWidths sets column widths as percentages.
func ColumnWidths(widths ...float64) TableOption {
	return func(cfg *tableConfig) {
		cfg.columnWidths = widths
	}
}

// TableCellVAlign sets the vertical alignment for table body cells.
func TableCellVAlign(align document.VerticalAlign) TableOption {
	return func(cfg *tableConfig) {
		cfg.cellVAlign = align
		cfg.hasCellVAlign = true
	}
}

// WithTableBorder draws a border around the table using the given [BorderSpec].
func WithTableBorder(spec BorderSpec) TableOption {
	return func(cfg *tableConfig) {
		cfg.border = &spec
	}
}

// WithTableBorderCollapse merges adjacent cell borders into a single line,
// like CSS border-collapse: collapse. The default is separated borders.
func WithTableBorderCollapse(collapse bool) TableOption {
	return func(cfg *tableConfig) {
		cfg.borderCollapse = collapse
		cfg.hasCollapse = true
	}
}

// WithTableBackground fills the table's outer box with the given color.
func WithTableBackground(c pdf.Color) TableOption {
	return func(cfg *tableConfig) {
		cfg.background = &c
	}
}

// WithTableCellBorder draws the same border around every header and body
// cell, producing a uniform grid. Combine with [WithTableBorder] for an
// outer frame, or use alone for cell-only grid lines.
func WithTableCellBorder(spec BorderSpec) TableOption {
	return func(cfg *tableConfig) {
		cfg.cellBorder = &spec
	}
}

// --- Box Options ---

// BoxOption configures a styled Box container created with [ColBuilder.Box].
type BoxOption func(*boxConfig)

type boxConfig struct {
	border     *BorderSpec
	background *pdf.Color
	padding    document.Edges
	hasPadding bool
	margin     document.Edges
	hasMargin  bool
	width      document.Value
	height     document.Value
}

// WithBoxBorder draws a border around the box using the given [BorderSpec].
func WithBoxBorder(spec BorderSpec) BoxOption {
	return func(cfg *boxConfig) {
		cfg.border = &spec
	}
}

// WithBoxBackground fills the box with the given color.
func WithBoxBackground(c pdf.Color) BoxOption {
	return func(cfg *boxConfig) {
		cfg.background = &c
	}
}

// WithBoxPadding sets uniform or per-edge padding inside the box.
func WithBoxPadding(e document.Edges) BoxOption {
	return func(cfg *boxConfig) {
		cfg.padding = e
		cfg.hasPadding = true
	}
}

// WithBoxMargin sets uniform or per-edge margin around the box.
func WithBoxMargin(e document.Edges) BoxOption {
	return func(cfg *boxConfig) {
		cfg.margin = e
		cfg.hasMargin = true
	}
}

// WithBoxWidth sets an explicit width for the box.
func WithBoxWidth(v document.Value) BoxOption {
	return func(cfg *boxConfig) { cfg.width = v }
}

// WithBoxHeight sets an explicit height for the box.
func WithBoxHeight(v document.Value) BoxOption {
	return func(cfg *boxConfig) { cfg.height = v }
}

// --- Text border ---

// WithTextBorder draws a border around a text paragraph. Background color is
// available via the existing [BgColor] option.
func WithTextBorder(spec BorderSpec) TextOption {
	return func(s *document.Style) {
		s.Border = spec.toEdges()
	}
}

// --- List Options ---

// ListOption configures a List element.
type ListOption func(*listConfig)

type listConfig struct {
	indent float64
}

// ListIndent sets the indentation width for list markers.
func ListIndent(v document.Value) ListOption {
	return func(cfg *listConfig) {
		cfg.indent = v.Resolve(0, 12)
	}
}

// --- Line Options ---

// LineOption configures a Line element.
type LineOption func(*lineConfig)

type lineConfig struct {
	color     pdf.Color
	thickness document.Value
}

// LineColor sets the line color.
func LineColor(c pdf.Color) LineOption {
	return func(cfg *lineConfig) {
		cfg.color = c
	}
}

// LineThickness sets the line thickness.
func LineThickness(v document.Value) LineOption {
	return func(cfg *lineConfig) {
		cfg.thickness = v
	}
}

// --- QR Code Options ---

// QRCodeOption configures a QR code element.
type QRCodeOption func(*qrCodeConfig)

type qrCodeConfig struct {
	size    document.Value
	minSize document.Value
	ecLevel qrcode.ErrorCorrectionLevel
	scale   int
}

// QRSize sets the display size (width = height) of the QR code.
func QRSize(v document.Value) QRCodeOption {
	return func(cfg *qrCodeConfig) {
		cfg.size = v
	}
}

// QRMinSize sets a minimum display size (width = height) for the QR code.
// When the layout would shrink the QR code below this value it is moved to
// the next page instead, preserving scannability.
func QRMinSize(v document.Value) QRCodeOption {
	return func(cfg *qrCodeConfig) {
		cfg.minSize = v
	}
}

// QRErrorCorrection sets the error correction level (L/M/Q/H).
func QRErrorCorrection(level qrcode.ErrorCorrectionLevel) QRCodeOption {
	return func(cfg *qrCodeConfig) {
		cfg.ecLevel = level
	}
}

// QRScale sets the number of pixels per QR module.
func QRScale(s int) QRCodeOption {
	return func(cfg *qrCodeConfig) {
		cfg.scale = s
	}
}

// --- Barcode Options ---

// BarcodeOption configures a barcode element.
type BarcodeOption func(*barcodeConfig)

type barcodeConfig struct {
	width  document.Value
	height document.Value
	format barcode.Format
}

// BarcodeWidth sets the display width of the barcode.
func BarcodeWidth(v document.Value) BarcodeOption {
	return func(cfg *barcodeConfig) {
		cfg.width = v
	}
}

// BarcodeHeight sets the display height of the barcode.
func BarcodeHeight(v document.Value) BarcodeOption {
	return func(cfg *barcodeConfig) {
		cfg.height = v
	}
}

// BarcodeFormat sets the barcode symbology.
func BarcodeFormat(f barcode.Format) BarcodeOption {
	return func(cfg *barcodeConfig) {
		cfg.format = f
	}
}

// --- Absolute Positioning Options ---

// AbsoluteOption configures an absolute-positioned element.
type AbsoluteOption func(*absoluteConfig)

type absoluteConfig struct {
	width  document.Value
	height document.Value
	origin document.PositionOrigin
}

// AbsoluteWidth sets the width constraint for the absolute-positioned content.
func AbsoluteWidth(w document.Value) AbsoluteOption {
	return func(c *absoluteConfig) { c.width = w }
}

// AbsoluteHeight sets the height constraint for the absolute-positioned content.
func AbsoluteHeight(h document.Value) AbsoluteOption {
	return func(c *absoluteConfig) { c.height = h }
}

// AbsoluteOriginPage sets the coordinate origin to the page corner
// (ignoring margins).
func AbsoluteOriginPage() AbsoluteOption {
	return func(c *absoluteConfig) { c.origin = document.OriginPage }
}
