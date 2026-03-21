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
	"github.com/gpdf-dev/gpdf/encrypt"
	"github.com/gpdf-dev/gpdf/pdfa"
	"github.com/gpdf-dev/gpdf/pdf"
	"github.com/gpdf-dev/gpdf/signature"
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
	// WithWriterSetup registers a Writer configuration hook for extensions.
	WithWriterSetup = template.WithWriterSetup
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

// Re-export absolute positioning option functions for convenience.
var (
	// AbsoluteWidth sets the width constraint for absolute-positioned content.
	AbsoluteWidth = template.AbsoluteWidth
	// AbsoluteHeight sets the height constraint for absolute-positioned content.
	AbsoluteHeight = template.AbsoluteHeight
	// AbsoluteOriginPage sets the coordinate origin to the page corner.
	AbsoluteOriginPage = template.AbsoluteOriginPage
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

// WithPDFA returns a template.Option that enables PDF/A conformance.
func WithPDFA(opts ...pdfa.Option) template.Option {
	return template.WithWriterSetup(func(pw *pdf.Writer) {
		pdfa.Apply(pw, opts...)
	})
}

// WithEncryption returns a template.Option that enables AES-256 encryption.
func WithEncryption(opts ...encrypt.Option) template.Option {
	return template.WithWriterSetup(func(pw *pdf.Writer) {
		_ = encrypt.Apply(pw, opts...)
	})
}

// SignDocument adds a digital signature to a generated PDF.
// This is a post-processing step applied after document generation.
func SignDocument(pdfData []byte, signer signature.Signer, opts ...signature.Option) ([]byte, error) {
	return signature.Sign(pdfData, signer, opts...)
}

// Open creates an ExistingDocument from raw PDF data for reading and modifying.
// Use the returned document's Overlay method to add content on top of existing pages.
//
//	doc, err := gpdf.Open(pdfBytes, gpdf.WithFont("NotoSans", fontData))
//	doc.Overlay(0, func(p *template.PageBuilder) {
//	    p.AutoRow(func(r *template.RowBuilder) {
//	        r.Col(12, func(c *template.ColBuilder) {
//	            c.Text("APPROVED", template.FontSize(48))
//	        })
//	    })
//	})
//	result, err := doc.Save()
func Open(data []byte, opts ...template.Option) (*template.ExistingDocument, error) {
	return template.OpenExisting(data, opts...)
}

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
