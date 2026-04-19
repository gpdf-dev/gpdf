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
	width     document.Value
	height    document.Value
	minWidth  document.Value
	minHeight document.Value
	fitMode   document.ImageFitMode
	align     document.TextAlign
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
	ecLevel qrcode.ErrorCorrectionLevel
	scale   int
}

// QRSize sets the display size (width = height) of the QR code.
func QRSize(v document.Value) QRCodeOption {
	return func(cfg *qrCodeConfig) {
		cfg.size = v
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
