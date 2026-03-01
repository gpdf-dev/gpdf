package examples_test

import (
	"testing"

	"github.com/gpdf-dev/gpdf/document"
	"github.com/gpdf-dev/gpdf/pdf"
	"github.com/gpdf-dev/gpdf/template"
)

func TestExample_17_Colors(t *testing.T) {
	doc := template.New(
		template.WithPageSize(document.A4),
		template.WithMargins(document.UniformEdges(document.Mm(20))),
	)

	page := doc.AddPage()

	page.AutoRow(func(r *template.RowBuilder) {
		r.Col(12, func(c *template.ColBuilder) {
			c.Text("Color System Examples", template.FontSize(18), template.Bold())
			c.Spacer(document.Mm(5))
		})
	})

	// Predefined colors
	page.AutoRow(func(r *template.RowBuilder) {
		r.Col(12, func(c *template.ColBuilder) {
			c.Text("Predefined Colors:", template.Bold())
			c.Text("Red", template.TextColor(pdf.Red))
			c.Text("Green", template.TextColor(pdf.Green))
			c.Text("Blue", template.TextColor(pdf.Blue))
			c.Text("Yellow", template.TextColor(pdf.Yellow))
			c.Text("Cyan", template.TextColor(pdf.Cyan))
			c.Text("Magenta", template.TextColor(pdf.Magenta))
			c.Spacer(document.Mm(5))
		})
	})

	// RGB colors (0.0-1.0)
	page.AutoRow(func(r *template.RowBuilder) {
		r.Col(12, func(c *template.ColBuilder) {
			c.Text("RGB Colors (float):", template.Bold())
			c.Text("RGB(1.0, 0.5, 0.0) - Orange", template.TextColor(pdf.RGB(1.0, 0.5, 0.0)))
			c.Text("RGB(0.5, 0.0, 0.5) - Purple", template.TextColor(pdf.RGB(0.5, 0.0, 0.5)))
			c.Text("RGB(0.0, 0.5, 0.5) - Teal", template.TextColor(pdf.RGB(0.0, 0.5, 0.5)))
			c.Spacer(document.Mm(5))
		})
	})

	// Hex colors
	page.AutoRow(func(r *template.RowBuilder) {
		r.Col(12, func(c *template.ColBuilder) {
			c.Text("Hex Colors:", template.Bold())
			c.Text("#FF6B6B - Coral", template.TextColor(pdf.RGBHex(0xFF6B6B)))
			c.Text("#4ECDC4 - Turquoise", template.TextColor(pdf.RGBHex(0x4ECDC4)))
			c.Text("#45B7D1 - Sky Blue", template.TextColor(pdf.RGBHex(0x45B7D1)))
			c.Text("#96CEB4 - Sage", template.TextColor(pdf.RGBHex(0x96CEB4)))
			c.Spacer(document.Mm(5))
		})
	})

	// Grayscale
	page.AutoRow(func(r *template.RowBuilder) {
		r.Col(12, func(c *template.ColBuilder) {
			c.Text("Grayscale:", template.Bold())
			c.Text("Gray(0.0) - Black", template.TextColor(pdf.Gray(0.0)))
			c.Text("Gray(0.3) - Dark gray", template.TextColor(pdf.Gray(0.3)))
			c.Text("Gray(0.5) - Medium gray", template.TextColor(pdf.Gray(0.5)))
			c.Text("Gray(0.7) - Light gray", template.TextColor(pdf.Gray(0.7)))
			c.Spacer(document.Mm(5))
		})
	})

	// Background color swatches
	page.AutoRow(func(r *template.RowBuilder) {
		r.Col(12, func(c *template.ColBuilder) {
			c.Text("Background Color Swatches:", template.Bold())
		})
	})

	page.AutoRow(func(r *template.RowBuilder) {
		r.Col(3, func(c *template.ColBuilder) {
			c.Text(" Red ", template.TextColor(pdf.White), template.BgColor(pdf.Red))
		})
		r.Col(3, func(c *template.ColBuilder) {
			c.Text(" Green ", template.TextColor(pdf.White), template.BgColor(pdf.Green))
		})
		r.Col(3, func(c *template.ColBuilder) {
			c.Text(" Blue ", template.TextColor(pdf.White), template.BgColor(pdf.Blue))
		})
		r.Col(3, func(c *template.ColBuilder) {
			c.Text(" Yellow ", template.BgColor(pdf.Yellow))
		})
	})

	generatePDF(t, "17_colors.pdf", doc)
}
