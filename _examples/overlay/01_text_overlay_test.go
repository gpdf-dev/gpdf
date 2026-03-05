package overlay_test

import (
	"fmt"
	"testing"

	gpdf "github.com/gpdf-dev/gpdf"
	"github.com/gpdf-dev/gpdf/_examples/testutil"
	"github.com/gpdf-dev/gpdf/document"
	"github.com/gpdf-dev/gpdf/pdf"
	"github.com/gpdf-dev/gpdf/template"
)

// generateSourcePDF creates a multi-page PDF to use as a base for overlay tests.
func generateSourcePDF(t *testing.T, numPages int) []byte {
	t.Helper()
	doc := template.New(
		template.WithPageSize(document.A4),
		template.WithMargins(document.UniformEdges(document.Mm(20))),
		template.WithMetadata(document.DocumentMetadata{
			Title:  "Source Document",
			Author: "gpdf",
		}),
	)

	for i := 1; i <= numPages; i++ {
		page := doc.AddPage()
		page.AutoRow(func(r *template.RowBuilder) {
			r.Col(12, func(c *template.ColBuilder) {
				c.Text(fmt.Sprintf("Page %d - Original Content", i),
					template.FontSize(18), template.Bold())
			})
		})
		page.AutoRow(func(r *template.RowBuilder) {
			r.Col(12, func(c *template.ColBuilder) {
				c.Spacer(document.Mm(5))
				c.Line()
				c.Spacer(document.Mm(5))
			})
		})
		for j := 0; j < 5; j++ {
			page.AutoRow(func(r *template.RowBuilder) {
				r.Col(12, func(c *template.ColBuilder) {
					c.Text("Lorem ipsum dolor sit amet, consectetur adipiscing elit. "+
						"Sed do eiusmod tempor incididunt ut labore et dolore magna aliqua.")
				})
			})
		}
	}

	data, err := doc.Generate()
	if err != nil {
		t.Fatalf("generate source PDF: %v", err)
	}
	return data
}

func TestExample_Overlay_01_TextOverlay(t *testing.T) {
	source := generateSourcePDF(t, 1)

	doc, err := gpdf.Open(source)
	if err != nil {
		t.Fatalf("gpdf.Open: %v", err)
	}

	// Add a "DRAFT" watermark using absolute positioning.
	err = doc.Overlay(0, func(p *template.PageBuilder) {
		p.Absolute(document.Mm(40), document.Mm(120), func(c *template.ColBuilder) {
			c.Text("DRAFT",
				template.FontSize(72),
				template.TextColor(pdf.Color{R: 0.9, G: 0.9, B: 0.9, A: 1, Space: pdf.ColorSpaceRGB}),
			)
		})
	})
	if err != nil {
		t.Fatalf("Overlay: %v", err)
	}

	result, err := doc.Save()
	if err != nil {
		t.Fatalf("Save: %v", err)
	}

	testutil.AssertValidPDF(t, result)
	testutil.WritePDF(t, "01_text_overlay.pdf", result)
}
