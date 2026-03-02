package htmlpdf

import (
	"bytes"
	"fmt"
	"io"
	"os"

	"github.com/gpdf-dev/gpdf/document"
	"github.com/gpdf-dev/gpdf/document/layout"
	"github.com/gpdf-dev/gpdf/document/render"
	"github.com/gpdf-dev/gpdf/html"
	"github.com/gpdf-dev/gpdf/pdf"
	"github.com/gpdf-dev/gpdf/pdf/font"
)

// Config holds the configuration for HTML→PDF conversion.
type Config struct {
	PageSize    document.Size
	Margins     document.Edges
	Fonts       map[string][]byte // family name → TTF data
	DefaultFont string
	DefaultSize float64
	ExtraCSS    string // additional CSS stylesheet
	BaseURL     string // for resolving relative paths
}

// defaultStyle returns the default document style from the config.
func (c *Config) defaultStyle() document.Style {
	s := document.DefaultStyle()
	if c.DefaultFont != "" {
		s.FontFamily = c.DefaultFont
	}
	if c.DefaultSize > 0 {
		s.FontSize = c.DefaultSize
	}
	s.LineHeight = s.FontSize * 1.2
	return s
}

// Option configures the HTML→PDF conversion.
type Option func(*Config)

// WithPageSize sets the page dimensions.
func WithPageSize(size document.Size) Option {
	return func(c *Config) { c.PageSize = size }
}

// WithMargins sets the page margins.
func WithMargins(margins document.Edges) Option {
	return func(c *Config) { c.Margins = margins }
}

// WithFont registers a TrueType font by family name.
func WithFont(family string, data []byte) Option {
	return func(c *Config) {
		if c.Fonts == nil {
			c.Fonts = make(map[string][]byte)
		}
		c.Fonts[family] = data
	}
}

// WithDefaultFont sets the default font family and size.
func WithDefaultFont(family string, size float64) Option {
	return func(c *Config) {
		c.DefaultFont = family
		c.DefaultSize = size
	}
}

// WithStylesheet adds additional CSS to be applied after document styles.
func WithStylesheet(cssText string) Option {
	return func(c *Config) { c.ExtraCSS = cssText }
}

// WithBaseURL sets the base URL for resolving relative resource paths.
func WithBaseURL(url string) Option {
	return func(c *Config) { c.BaseURL = url }
}

// Result holds the conversion result.
type Result struct {
	doc          *document.Document
	fontResolver *htmlFontResolver
	config       *Config
}

// Write writes the PDF to the given writer.
func (r *Result) Write(w io.Writer) error {
	return r.render(w)
}

// Bytes returns the PDF as a byte slice.
func (r *Result) Bytes() ([]byte, error) {
	var buf bytes.Buffer
	if err := r.render(&buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (r *Result) render(w io.Writer) error {
	pdfWriter := pdf.NewWriter(w)
	pdfWriter.SetCompression(true)

	// Register fonts on the PDF writer
	for family, data := range r.config.Fonts {
		if _, _, err := pdfWriter.RegisterFont(family, data); err != nil {
			return fmt.Errorf("htmlpdf: register font %q: %w", family, err)
		}
	}

	// Run layout
	paginator := layout.NewPaginator(r.config.PageSize, r.config.Margins, r.fontResolver)
	pageLayouts := paginator.Paginate(r.doc)

	// Resolve page number placeholders
	layout.ResolvePageNumbers(pageLayouts)

	// Render
	pdfRenderer := render.NewPDFRenderer(pdfWriter)
	return pdfRenderer.RenderDocument(pageLayouts, document.DocumentMetadata{})
}

// FromHTML converts an HTML string to a PDF Result.
func FromHTML(htmlStr string, opts ...Option) (*Result, error) {
	config := defaultConfig()
	for _, opt := range opts {
		opt(config)
	}

	// Parse HTML
	root, err := html.ParseString(htmlStr)
	if err != nil {
		return nil, fmt.Errorf("htmlpdf: parse HTML: %w", err)
	}

	// Convert to document model
	conv := newConverter(config, root)
	doc := conv.convert(root)

	// Build font resolver
	fr, err := newHTMLFontResolver(config)
	if err != nil {
		return nil, fmt.Errorf("htmlpdf: font resolver: %w", err)
	}

	return &Result{
		doc:          doc,
		fontResolver: fr,
		config:       config,
	}, nil
}

// FromHTMLFile converts an HTML file to a PDF Result.
func FromHTMLFile(path string, opts ...Option) (*Result, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("htmlpdf: read file: %w", err)
	}
	return FromHTML(string(data), opts...)
}

// defaultConfig returns the default configuration.
func defaultConfig() *Config {
	return &Config{
		PageSize: document.A4,
		Margins:  document.UniformEdges(document.Mm(20)),
		Fonts:    make(map[string][]byte),
	}
}

// ─── Font resolver ────────────────────────────────────────────────

// htmlFontResolver implements layout.FontResolver for HTML→PDF conversion.
type htmlFontResolver struct {
	fonts   map[string]*font.TrueTypeFont
	metrics map[string]layout.FontMetrics
}

func newHTMLFontResolver(config *Config) (*htmlFontResolver, error) {
	fr := &htmlFontResolver{
		fonts:   make(map[string]*font.TrueTypeFont),
		metrics: make(map[string]layout.FontMetrics),
	}

	for family, data := range config.Fonts {
		ttf, err := font.ParseTrueType(data)
		if err != nil {
			return nil, fmt.Errorf("parse font %q: %w", family, err)
		}
		fr.fonts[family] = ttf
		fr.metrics[family] = extractMetrics(ttf)
	}

	return fr, nil
}

func extractMetrics(ttf *font.TrueTypeFont) layout.FontMetrics {
	m := ttf.Metrics()
	unitsPerEm := float64(m.UnitsPerEm)
	if unitsPerEm == 0 {
		unitsPerEm = 1000
	}
	return layout.FontMetrics{
		Ascender:   float64(m.Ascender) / unitsPerEm,
		Descender:  float64(m.Descender) / unitsPerEm,
		LineHeight: float64(m.Ascender-m.Descender) / unitsPerEm,
		CapHeight:  float64(m.CapHeight) / unitsPerEm,
	}
}

func (fr *htmlFontResolver) Resolve(family string, weight document.FontWeight, italic bool) layout.ResolvedFont {
	// Try exact family match
	if _, ok := fr.fonts[family]; ok {
		return layout.ResolvedFont{
			ID:      family,
			Metrics: fr.metrics[family],
		}
	}
	// Fall back to first available font
	for name := range fr.fonts {
		return layout.ResolvedFont{
			ID:      name,
			Metrics: fr.metrics[name],
		}
	}
	// No fonts available — return a placeholder
	return layout.ResolvedFont{
		ID: "default",
		Metrics: layout.FontMetrics{
			Ascender:   0.8,
			Descender:  -0.2,
			LineHeight: 1.0,
			CapHeight:  0.7,
		},
	}
}

func (fr *htmlFontResolver) MeasureString(f layout.ResolvedFont, text string, size float64) float64 {
	ttf, ok := fr.fonts[f.ID]
	if !ok {
		// Approximate: 0.5 * fontSize per character
		return float64(len([]rune(text))) * size * 0.5
	}
	return font.MeasureString(ttf, text, size)
}

func (fr *htmlFontResolver) LineBreak(f layout.ResolvedFont, text string, size float64, maxWidth float64) []string {
	ttf, ok := fr.fonts[f.ID]
	if !ok {
		return simpleLineBreak(text, size, maxWidth)
	}
	return font.LineBreak(ttf, text, size, maxWidth)
}

// simpleLineBreak is a fallback word-wrap without font metrics.
func simpleLineBreak(text string, size float64, maxWidth float64) []string {
	if maxWidth <= 0 {
		return []string{text}
	}
	charWidth := size * 0.5
	charsPerLine := int(maxWidth / charWidth)
	if charsPerLine < 1 {
		charsPerLine = 1
	}

	runes := []rune(text)
	var lines []string
	for len(runes) > 0 {
		end := charsPerLine
		if end > len(runes) {
			end = len(runes)
		}
		lines = append(lines, string(runes[:end]))
		runes = runes[end:]
	}
	return lines
}
