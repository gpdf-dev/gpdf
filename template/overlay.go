package template

import (
	"fmt"

	"github.com/gpdf-dev/gpdf/document"
	"github.com/gpdf-dev/gpdf/document/render"
	"github.com/gpdf-dev/gpdf/pdf"
	"github.com/gpdf-dev/gpdf/pdf/font"
)

// ExistingDocument represents an existing PDF that can be modified
// with overlay content using the same builder API as new documents.
type ExistingDocument struct {
	reader      *pdf.Reader
	modifier    *pdf.Modifier
	fonts       map[string]*font.TrueTypeFont
	fontDataMap map[string][]byte
	config      Config
}

// OpenExisting creates an ExistingDocument from raw PDF data.
func OpenExisting(data []byte, opts ...Option) (*ExistingDocument, error) {
	reader, err := pdf.NewReader(data)
	if err != nil {
		return nil, fmt.Errorf("gpdf: open existing PDF: %w", err)
	}

	cfg := Config{
		PageSize: document.A4,
		Margins:  document.UniformEdges(document.Pt(0)),
		FontSize: 12,
		rawFonts: make(map[string][]byte),
	}
	for _, opt := range opts {
		opt(&cfg)
	}

	doc := &ExistingDocument{
		reader:      reader,
		modifier:    pdf.NewModifier(reader),
		fonts:       make(map[string]*font.TrueTypeFont),
		fontDataMap: make(map[string][]byte),
		config:      cfg,
	}

	for family, data := range cfg.rawFonts {
		ttf, err := font.ParseTrueType(data)
		if err == nil {
			doc.fonts[family] = ttf
			doc.fontDataMap[family] = data
		}
	}

	return doc, nil
}

// PageCount returns the number of pages in the existing PDF.
func (d *ExistingDocument) PageCount() (int, error) {
	return d.reader.PageCount()
}

// Overlay adds content on top of the specified page using the builder API.
// The callback receives a PageBuilder for defining overlay content.
// Coordinates are relative to the page origin (top-left of the full page).
func (d *ExistingDocument) Overlay(pageIndex int, fn func(p *PageBuilder)) error {
	info, err := d.reader.Page(pageIndex)
	if err != nil {
		return err
	}

	// Use the actual page dimensions from the existing PDF.
	pageSize := document.Size{
		Width:  info.MediaBox.URX - info.MediaBox.LLX,
		Height: info.MediaBox.URY - info.MediaBox.LLY,
	}

	// Build document nodes from the builder.
	templateDoc := &Document{
		config:      d.config,
		fonts:       d.fonts,
		fontDataMap: d.fontDataMap,
	}
	templateDoc.fontResolver = newBuiltinFontResolver(d.fonts)

	pb := &PageBuilder{doc: templateDoc}
	fn(pb)
	nodes := pb.buildNodes()

	if len(nodes) == 0 {
		return nil
	}

	// Use zero margins for overlay (content relative to page edge).
	margins := document.UniformEdges(document.Pt(0))

	// Render overlay content.
	result, err := render.RenderOverlayContent(
		nodes,
		pageSize,
		margins,
		d.fonts,
		d.fontDataMap,
		templateDoc.fontResolver,
	)
	if err != nil {
		return fmt.Errorf("gpdf: render overlay: %w", err)
	}

	// Write overlay resources to modifier and get content + resource dict.
	content, resources, err := render.WriteOverlayToModifier(result, d.modifier)
	if err != nil {
		return fmt.Errorf("gpdf: write overlay: %w", err)
	}

	return d.modifier.OverlayPage(pageIndex, content, resources)
}

// EachPage applies the overlay function to every page.
// The callback receives the page index and a PageBuilder.
func (d *ExistingDocument) EachPage(fn func(pageIndex int, p *PageBuilder)) error {
	count, err := d.PageCount()
	if err != nil {
		return err
	}
	for i := 0; i < count; i++ {
		idx := i
		if err := d.Overlay(i, func(p *PageBuilder) {
			fn(idx, p)
		}); err != nil {
			return err
		}
	}
	return nil
}

// FlattenForms flattens AcroForm fields into page content streams,
// making form data part of the static page content and removing
// all interactive form elements. Returns nil if no forms are present.
func (d *ExistingDocument) FlattenForms() error {
	return d.modifier.FlattenForms()
}

// Save generates the modified PDF as a byte slice.
func (d *ExistingDocument) Save() ([]byte, error) {
	return d.modifier.Bytes()
}
