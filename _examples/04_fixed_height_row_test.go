package examples_test

import (
	"testing"

	"github.com/gpdf-dev/gpdf/document"
	"github.com/gpdf-dev/gpdf/pdf"
	"github.com/gpdf-dev/gpdf/template"
)

func TestExample_04_FixedHeightRow(t *testing.T) {
	doc := template.New(
		template.WithPageSize(document.A4),
		template.WithMargins(document.UniformEdges(document.Mm(20))),
	)

	page := doc.AddPage()

	page.AutoRow(func(r *template.RowBuilder) {
		r.Col(12, func(c *template.ColBuilder) {
			c.Text("Fixed-Height Row Examples", template.FontSize(18), template.Bold())
			c.Spacer(document.Mm(5))
		})
	})

	// Fixed height row: 30mm
	page.Row(document.Mm(30), func(r *template.RowBuilder) {
		r.Col(12, func(c *template.ColBuilder) {
			c.Text("This row is 30mm tall", template.BgColor(pdf.RGBHex(0xE3F2FD)))
		})
	})

	page.AutoRow(func(r *template.RowBuilder) {
		r.Col(12, func(c *template.ColBuilder) {
			c.Spacer(document.Mm(3))
		})
	})

	// Fixed height row: 50mm
	page.Row(document.Mm(50), func(r *template.RowBuilder) {
		r.Col(6, func(c *template.ColBuilder) {
			c.Text("Left: 50mm row", template.BgColor(pdf.RGBHex(0xE8F5E9)))
		})
		r.Col(6, func(c *template.ColBuilder) {
			c.Text("Right: 50mm row", template.BgColor(pdf.RGBHex(0xFFF3E0)))
		})
	})

	page.AutoRow(func(r *template.RowBuilder) {
		r.Col(12, func(c *template.ColBuilder) {
			c.Spacer(document.Mm(3))
		})
	})

	// Auto-height row for comparison
	page.AutoRow(func(r *template.RowBuilder) {
		r.Col(12, func(c *template.ColBuilder) {
			c.Text("This row has auto height (fits content)", template.BgColor(pdf.RGBHex(0xFCE4EC)))
		})
	})

	generatePDF(t, "04_fixed_height_row.pdf", doc)
}
