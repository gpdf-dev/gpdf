package examples_test

import (
	"testing"

	"github.com/gpdf-dev/gpdf/document"
	"github.com/gpdf-dev/gpdf/pdf"
	"github.com/gpdf-dev/gpdf/template"
)

func TestExample_05_Line(t *testing.T) {
	doc := template.New(
		template.WithPageSize(document.A4),
		template.WithMargins(document.UniformEdges(document.Mm(20))),
	)

	page := doc.AddPage()

	page.AutoRow(func(r *template.RowBuilder) {
		r.Col(12, func(c *template.ColBuilder) {
			c.Text("Line / Horizontal Rule Examples", template.FontSize(18), template.Bold())
			c.Spacer(document.Mm(5))
		})
	})

	// Default line
	page.AutoRow(func(r *template.RowBuilder) {
		r.Col(12, func(c *template.ColBuilder) {
			c.Text("Default line (gray, 1pt):")
			c.Line()
			c.Spacer(document.Mm(5))
		})
	})

	// Colored lines
	page.AutoRow(func(r *template.RowBuilder) {
		r.Col(12, func(c *template.ColBuilder) {
			c.Text("Red line:")
			c.Line(template.LineColor(pdf.Red))
			c.Spacer(document.Mm(3))
			c.Text("Blue line:")
			c.Line(template.LineColor(pdf.Blue))
			c.Spacer(document.Mm(3))
			c.Text("Green line:")
			c.Line(template.LineColor(pdf.Green))
			c.Spacer(document.Mm(5))
		})
	})

	// Thick lines
	page.AutoRow(func(r *template.RowBuilder) {
		r.Col(12, func(c *template.ColBuilder) {
			c.Text("Thin line (0.5pt):")
			c.Line(template.LineThickness(document.Pt(0.5)))
			c.Spacer(document.Mm(3))
			c.Text("Medium line (2pt):")
			c.Line(template.LineThickness(document.Pt(2)))
			c.Spacer(document.Mm(3))
			c.Text("Thick line (5pt):")
			c.Line(template.LineThickness(document.Pt(5)))
			c.Spacer(document.Mm(5))
		})
	})

	// Combined: color + thickness
	page.AutoRow(func(r *template.RowBuilder) {
		r.Col(12, func(c *template.ColBuilder) {
			c.Text("Thick red line (3pt):")
			c.Line(template.LineColor(pdf.Red), template.LineThickness(document.Pt(3)))
			c.Spacer(document.Mm(3))
			c.Text("Thick blue line (4pt):")
			c.Line(template.LineColor(pdf.Blue), template.LineThickness(document.Pt(4)))
		})
	})

	generatePDF(t, "05_line.pdf", doc)
}
