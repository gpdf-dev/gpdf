package builder_test

import (
	"testing"

	"github.com/gpdf-dev/gpdf/document"
	"github.com/gpdf-dev/gpdf/pdf"
	"github.com/gpdf-dev/gpdf/_examples/testutil"
	"github.com/gpdf-dev/gpdf/template"
)

func TestExample_03_GridLayout(t *testing.T) {
	doc := template.New(
		template.WithPageSize(document.A4),
		template.WithMargins(document.UniformEdges(document.Mm(15))),
	)

	page := doc.AddPage()

	// Title
	page.AutoRow(func(r *template.RowBuilder) {
		r.Col(12, func(c *template.ColBuilder) {
			c.Text("12-Column Grid Layout", template.FontSize(18), template.Bold())
			c.Spacer(document.Mm(5))
		})
	})

	// Full width (12 columns)
	page.AutoRow(func(r *template.RowBuilder) {
		r.Col(12, func(c *template.ColBuilder) {
			c.Text("Col 12 (full width)", template.BgColor(pdf.RGBHex(0xE3F2FD)))
		})
	})

	page.AutoRow(func(r *template.RowBuilder) {
		r.Col(12, func(c *template.ColBuilder) {
			c.Spacer(document.Mm(3))
		})
	})

	// Two columns (6 + 6)
	page.AutoRow(func(r *template.RowBuilder) {
		r.Col(6, func(c *template.ColBuilder) {
			c.Text("Col 6 (left)", template.BgColor(pdf.RGBHex(0xE8F5E9)))
		})
		r.Col(6, func(c *template.ColBuilder) {
			c.Text("Col 6 (right)", template.BgColor(pdf.RGBHex(0xFFF3E0)))
		})
	})

	page.AutoRow(func(r *template.RowBuilder) {
		r.Col(12, func(c *template.ColBuilder) {
			c.Spacer(document.Mm(3))
		})
	})

	// Three columns (4 + 4 + 4)
	page.AutoRow(func(r *template.RowBuilder) {
		r.Col(4, func(c *template.ColBuilder) {
			c.Text("Col 4", template.BgColor(pdf.RGBHex(0xFCE4EC)))
		})
		r.Col(4, func(c *template.ColBuilder) {
			c.Text("Col 4", template.BgColor(pdf.RGBHex(0xF3E5F5)))
		})
		r.Col(4, func(c *template.ColBuilder) {
			c.Text("Col 4", template.BgColor(pdf.RGBHex(0xE8EAF6)))
		})
	})

	page.AutoRow(func(r *template.RowBuilder) {
		r.Col(12, func(c *template.ColBuilder) {
			c.Spacer(document.Mm(3))
		})
	})

	// Four columns (3 + 3 + 3 + 3)
	page.AutoRow(func(r *template.RowBuilder) {
		r.Col(3, func(c *template.ColBuilder) {
			c.Text("Col 3", template.BgColor(pdf.RGBHex(0xE0F7FA)))
		})
		r.Col(3, func(c *template.ColBuilder) {
			c.Text("Col 3", template.BgColor(pdf.RGBHex(0xE0F2F1)))
		})
		r.Col(3, func(c *template.ColBuilder) {
			c.Text("Col 3", template.BgColor(pdf.RGBHex(0xFFF9C4)))
		})
		r.Col(3, func(c *template.ColBuilder) {
			c.Text("Col 3", template.BgColor(pdf.RGBHex(0xFFECB3)))
		})
	})

	page.AutoRow(func(r *template.RowBuilder) {
		r.Col(12, func(c *template.ColBuilder) {
			c.Spacer(document.Mm(3))
		})
	})

	// Asymmetric layout (3 + 9)
	page.AutoRow(func(r *template.RowBuilder) {
		r.Col(3, func(c *template.ColBuilder) {
			c.Text("Sidebar (3)", template.BgColor(pdf.RGBHex(0xD7CCC8)))
		})
		r.Col(9, func(c *template.ColBuilder) {
			c.Text("Main content (9)", template.BgColor(pdf.RGBHex(0xF5F5F5)))
		})
	})

	page.AutoRow(func(r *template.RowBuilder) {
		r.Col(12, func(c *template.ColBuilder) {
			c.Spacer(document.Mm(3))
		})
	})

	// Asymmetric layout (8 + 4)
	page.AutoRow(func(r *template.RowBuilder) {
		r.Col(8, func(c *template.ColBuilder) {
			c.Text("Article area (8)", template.BgColor(pdf.RGBHex(0xE1F5FE)))
		})
		r.Col(4, func(c *template.ColBuilder) {
			c.Text("Side panel (4)", template.BgColor(pdf.RGBHex(0xFBE9E7)))
		})
	})

	page.AutoRow(func(r *template.RowBuilder) {
		r.Col(12, func(c *template.ColBuilder) {
			c.Spacer(document.Mm(3))
		})
	})

	// Multiple content in columns
	page.AutoRow(func(r *template.RowBuilder) {
		r.Col(6, func(c *template.ColBuilder) {
			c.Text("Left column - line 1")
			c.Text("Left column - line 2")
			c.Text("Left column - line 3")
		})
		r.Col(6, func(c *template.ColBuilder) {
			c.Text("Right column - line 1")
			c.Text("Right column - line 2")
			c.Text("Right column - line 3")
		})
	})

	testutil.GeneratePDFSharedGolden(t, "03_grid_layout.pdf", doc)
}
