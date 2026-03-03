package builder_test

import (
	"testing"

	"github.com/gpdf-dev/gpdf/_examples/testutil"
	"github.com/gpdf-dev/gpdf/document"
	"github.com/gpdf-dev/gpdf/template"
)

func TestExample_06_Spacer(t *testing.T) {
	doc := template.New(
		template.WithPageSize(document.A4),
		template.WithMargins(document.UniformEdges(document.Mm(20))),
	)

	page := doc.AddPage()

	page.AutoRow(func(r *template.RowBuilder) {
		r.Col(12, func(c *template.ColBuilder) {
			c.Text("Spacer Examples", template.FontSize(18), template.Bold())
		})
	})

	page.AutoRow(func(r *template.RowBuilder) {
		r.Col(12, func(c *template.ColBuilder) {
			c.Text("Text before 5mm spacer")
			c.Spacer(document.Mm(5))
			c.Text("Text after 5mm spacer")
		})
	})

	page.AutoRow(func(r *template.RowBuilder) {
		r.Col(12, func(c *template.ColBuilder) {
			c.Text("Text before 15mm spacer")
			c.Spacer(document.Mm(15))
			c.Text("Text after 15mm spacer")
		})
	})

	page.AutoRow(func(r *template.RowBuilder) {
		r.Col(12, func(c *template.ColBuilder) {
			c.Text("Text before 30mm spacer")
			c.Spacer(document.Mm(30))
			c.Text("Text after 30mm spacer")
		})
	})

	testutil.GeneratePDFSharedGolden(t, "06_spacer.pdf", doc)
}
