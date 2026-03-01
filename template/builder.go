// Package template provides a declarative, builder-pattern API for
// constructing PDF documents. It sits on top of the document model
// (Layer 2) and provides high-level constructs such as grids, headers,
// footers, and reusable components.
package template

import (
	"bytes"
	"io"

	"github.com/gpdf-dev/gpdf/document"
	"github.com/gpdf-dev/gpdf/document/layout"
	"github.com/gpdf-dev/gpdf/document/render"
	"github.com/gpdf-dev/gpdf/internal/buildinfo"
	"github.com/gpdf-dev/gpdf/pdf"
	"github.com/gpdf-dev/gpdf/pdf/font"
)

// Document is the top-level builder for constructing PDF documents.
type Document struct {
	config       Config
	pages        []*PageBuilder
	headerFn     func(p *PageBuilder)
	footerFn     func(p *PageBuilder)
	fonts        map[string]*font.TrueTypeFont
	fontDataMap  map[string][]byte
	fontResolver *builtinFontResolver
}

// Config holds document-level configuration.
type Config struct {
	PageSize    document.Size
	Margins     document.Edges
	DefaultFont string
	FontSize    float64
	Metadata    document.DocumentMetadata
	rawFonts    map[string][]byte // font family -> raw TTF data
}

// Option configures a Document.
type Option func(*Config)

// WithPageSize sets the page size for the document.
func WithPageSize(size document.Size) Option {
	return func(c *Config) { c.PageSize = size }
}

// WithMargins sets the page margins for the document.
func WithMargins(margins document.Edges) Option {
	return func(c *Config) { c.Margins = margins }
}

// WithFont registers a TrueType font with the given family name.
func WithFont(family string, data []byte) Option {
	return func(c *Config) {
		if c.rawFonts == nil {
			c.rawFonts = make(map[string][]byte)
		}
		c.rawFonts[family] = data
	}
}

// WithDefaultFont sets the default font family and size.
func WithDefaultFont(family string, size float64) Option {
	return func(c *Config) {
		c.DefaultFont = family
		c.FontSize = size
	}
}

// WithMetadata sets the document metadata (title, author, etc.).
func WithMetadata(info document.DocumentMetadata) Option {
	return func(c *Config) { c.Metadata = info }
}

// New creates a new Document builder with the given options.
func New(opts ...Option) *Document {
	cfg := Config{
		PageSize: document.A4,
		Margins:  document.UniformEdges(document.Mm(20)),
		FontSize: 12,
		rawFonts: make(map[string][]byte),
	}

	for _, opt := range opts {
		opt(&cfg)
	}

	doc := &Document{
		config:      cfg,
		fonts:       make(map[string]*font.TrueTypeFont),
		fontDataMap: make(map[string][]byte),
	}

	// Parse registered fonts.
	for family, data := range cfg.rawFonts {
		ttf, err := font.ParseTrueType(data)
		if err == nil {
			doc.fonts[family] = ttf
			doc.fontDataMap[family] = data
		}
	}

	doc.fontResolver = newBuiltinFontResolver(doc.fonts)

	return doc
}

// AddPage adds a new page to the document and returns its builder.
func (d *Document) AddPage() *PageBuilder {
	pb := &PageBuilder{
		doc: d,
	}
	d.pages = append(d.pages, pb)
	return pb
}

// Header registers a function that builds header content. The function
// is called for every page to produce consistent headers.
func (d *Document) Header(fn func(p *PageBuilder)) {
	d.headerFn = fn
}

// Footer registers a function that builds footer content. The function
// is called for every page to produce consistent footers.
func (d *Document) Footer(fn func(p *PageBuilder)) {
	d.footerFn = fn
}

// Generate produces the PDF as a byte slice.
func (d *Document) Generate() ([]byte, error) {
	var buf bytes.Buffer
	if err := d.Render(&buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// Render writes the PDF to the given writer.
func (d *Document) Render(w io.Writer) error {
	// 1. Build the document model from the builder state (body only).
	doc := d.buildDocument()

	// 2. Build header/footer nodes separately.
	headerNodes := d.buildSection(d.headerFn)
	footerNodes := d.buildSection(d.footerFn)

	// 3. Paginate using the layout engine.
	paginator := layout.NewPaginator(d.config.PageSize, d.config.Margins, d.fontResolver)
	paginator.SetHeaderFooter(headerNodes, footerNodes)
	pages := paginator.Paginate(doc)

	// 4. Replace page-number placeholders.
	layout.ResolvePageNumbers(pages)

	// 5. Render to PDF.
	pw := pdf.NewWriter(w)

	// Register fonts with the PDF writer.
	for family, data := range d.fontDataMap {
		if _, _, err := pw.RegisterFont(family, data); err != nil {
			return err
		}
	}

	renderer := render.NewPDFRenderer(pw)
	return renderer.RenderDocument(pages, doc.Metadata)
}

// buildSection builds document nodes from a header or footer function.
func (d *Document) buildSection(fn func(p *PageBuilder)) []document.DocumentNode {
	if fn == nil {
		return nil
	}
	pb := &PageBuilder{doc: d}
	fn(pb)
	return pb.buildNodes()
}

// buildDocument converts the builder state into a document.Document tree.
func (d *Document) buildDocument() *document.Document {
	doc := &document.Document{
		Metadata: d.config.Metadata,
		DefaultStyle: document.Style{
			FontFamily: d.config.DefaultFont,
			FontSize:   d.config.FontSize,
			FontWeight: document.WeightNormal,
			Color:      pdf.Black,
			TextAlign:  document.AlignLeft,
			LineHeight: 1.2,
		},
	}

	if d.config.Metadata.Producer == "" {
		doc.Metadata.Producer = "gpdf/" + buildinfo.Version
	}

	for _, pb := range d.pages {
		page := &document.Page{
			Size:    d.config.PageSize,
			Margins: d.config.Margins,
		}

		// Body content only — header and footer are handled by the
		// paginator so they appear on overflow pages as well.
		page.Content = append(page.Content, pb.buildNodes()...)

		doc.Pages = append(doc.Pages, page)
	}

	return doc
}
