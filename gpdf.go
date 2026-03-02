// Package gpdf is a pure Go, zero-dependency PDF generation library.
//
// gpdf provides a layered architecture for PDF creation:
//
//   - pdf (Layer 1): Low-level PDF primitives — objects, streams, fonts, images
//   - document (Layer 2): Document model — nodes, box model, layout engine, renderer
//   - template (Layer 3): High-level declarative API — builders, grids, components
//
// Most users should start with the template package for the simplest API:
//
//	doc := gpdf.NewDocument(
//	    gpdf.WithPageSize(document.A4),
//	    gpdf.WithMargins(document.UniformEdges(document.Mm(15))),
//	)
//	page := doc.AddPage()
//	page.AutoRow(func(r *template.RowBuilder) {
//	    r.Col(12, func(c *template.ColBuilder) {
//	        c.Text("Hello, World!", template.FontSize(24))
//	    })
//	})
//	data, err := doc.Generate()
package gpdf

import (
	gotemplate "text/template"

	"github.com/gpdf-dev/gpdf/document"
	"github.com/gpdf-dev/gpdf/htmlpdf"
	"github.com/gpdf-dev/gpdf/template"
)

// NewDocument creates a new PDF document builder with the given options.
// This is the primary entry point for creating PDFs using the high-level API.
func NewDocument(opts ...template.Option) *template.Document {
	return template.New(opts...)
}

// Re-export commonly used option functions for convenience.
var (
	// WithPageSize sets the page dimensions.
	WithPageSize = template.WithPageSize
	// WithMargins sets the page margins.
	WithMargins = template.WithMargins
	// WithFont registers a TrueType font.
	WithFont = template.WithFont
	// WithDefaultFont sets the default font family and size.
	WithDefaultFont = template.WithDefaultFont
	// WithMetadata sets document metadata (title, author, etc.).
	WithMetadata = template.WithMetadata
)

// Re-export commonly used page sizes for convenience.
var (
	A4     = document.A4
	A3     = document.A3
	Letter = document.Letter
	Legal  = document.Legal
)

// Re-export QR code option functions for convenience.
var (
	// QRSize sets the display size of a QR code (width = height).
	QRSize = template.QRSize
	// QRErrorCorrection sets the QR error correction level.
	QRErrorCorrection = template.QRErrorCorrection
	// QRScale sets the pixels per QR module.
	QRScale = template.QRScale
)

// Re-export barcode option functions for convenience.
var (
	// BarcodeWidth sets the display width of a barcode.
	BarcodeWidth = template.BarcodeWidth
	// BarcodeHeight sets the display height of a barcode.
	BarcodeHeight = template.BarcodeHeight
	// BarcodeFormat sets the barcode symbology.
	BarcodeFormat = template.BarcodeFormat
)

// Re-export JSON schema / Go template integration functions.
var (
	// FromJSON creates a Document from a JSON schema definition with
	// optional Go template data binding.
	FromJSON = template.FromJSON
	// FromTemplate creates a Document by executing a pre-parsed Go
	// template that produces JSON schema output.
	FromTemplate = template.FromTemplate
	// TemplateFuncMap returns helper functions (e.g., toJSON) for use
	// when parsing Go templates for FromTemplate.
	TemplateFuncMap = template.TemplateFuncMap
)

// Re-export reusable component constructors for convenience.
var (
	// NewInvoice creates a ready-to-generate invoice PDF from structured data.
	NewInvoice = template.Invoice
	// NewReport creates a ready-to-generate report PDF from structured data.
	NewReport = template.Report
	// NewLetter creates a ready-to-generate business letter PDF from structured data.
	NewLetter = template.Letter
)

// FromHTML converts an HTML string to PDF bytes.
// This is a convenience wrapper around htmlpdf.FromHTML.
func FromHTML(html string, opts ...htmlpdf.Option) ([]byte, error) {
	result, err := htmlpdf.FromHTML(html, opts...)
	if err != nil {
		return nil, err
	}
	return result.Bytes()
}

// Re-export HTML→PDF option functions for convenience.
var (
	// HTMLWithPageSize sets the page dimensions for HTML→PDF conversion.
	HTMLWithPageSize = htmlpdf.WithPageSize
	// HTMLWithMargins sets the page margins for HTML→PDF conversion.
	HTMLWithMargins = htmlpdf.WithMargins
	// HTMLWithFont registers a TrueType font for HTML→PDF conversion.
	HTMLWithFont = htmlpdf.WithFont
	// HTMLWithDefaultFont sets the default font for HTML→PDF conversion.
	HTMLWithDefaultFont = htmlpdf.WithDefaultFont
	// HTMLWithStylesheet adds additional CSS for HTML→PDF conversion.
	HTMLWithStylesheet = htmlpdf.WithStylesheet
	// HTMLWithBaseURL sets the base URL for resolving relative paths.
	HTMLWithBaseURL = htmlpdf.WithBaseURL
)

// NewDocumentFromJSON is an alias for FromJSON that creates a Document
// from a JSON schema, optionally resolving Go template expressions with data.
func NewDocumentFromJSON(schema []byte, data any, opts ...template.Option) (*template.Document, error) {
	return template.FromJSON(schema, data, opts...)
}

// NewDocumentFromTemplate creates a Document by executing a Go template
// that produces JSON schema output.
func NewDocumentFromTemplate(tmpl *gotemplate.Template, data any, opts ...template.Option) (*template.Document, error) {
	return template.FromTemplate(tmpl, data, opts...)
}
