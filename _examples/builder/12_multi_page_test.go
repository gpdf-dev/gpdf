package builder_test

import (
	"testing"

	"github.com/gpdf-dev/gpdf/document"
	"github.com/gpdf-dev/gpdf/_examples/testutil"
	"github.com/gpdf-dev/gpdf/template"
)

func TestExample_12_MultiPage(t *testing.T) {
	doc := template.New(
		template.WithPageSize(document.A4),
		template.WithMargins(document.UniformEdges(document.Mm(20))),
	)

	for i := 1; i <= 5; i++ {
		page := doc.AddPage()
		page.AutoRow(func(r *template.RowBuilder) {
			r.Col(12, func(c *template.ColBuilder) {
				c.Text("Multi-Page Document", template.FontSize(20), template.Bold())
				c.Spacer(document.Mm(5))
				c.Line()
				c.Spacer(document.Mm(10))
			})
		})

		// Fill the page with some content
		for j := 1; j <= 10; j++ {
			page.AutoRow(func(r *template.RowBuilder) {
				r.Col(12, func(c *template.ColBuilder) {
					c.Text("Lorem ipsum dolor sit amet, consectetur adipiscing elit. " +
						"Sed do eiusmod tempor incididunt ut labore et dolore magna aliqua.")
				})
			})
		}
	}

	testutil.GeneratePDFSharedGolden(t, "12_multi_page.pdf", doc)
}
