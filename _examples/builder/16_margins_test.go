package builder_test

import (
	"testing"

	"github.com/gpdf-dev/gpdf/_examples/testutil"
	"github.com/gpdf-dev/gpdf/document"
	"github.com/gpdf-dev/gpdf/template"
)

func TestExample_16_Margins(t *testing.T) {
	doc := template.New(
		template.WithPageSize(document.A4),
		template.WithMargins(document.Edges{
			Top:    document.Mm(10),
			Right:  document.Mm(40),
			Bottom: document.Mm(10),
			Left:   document.Mm(40),
		}),
	)

	page := doc.AddPage()
	page.AutoRow(func(r *template.RowBuilder) {
		r.Col(12, func(c *template.ColBuilder) {
			c.Text("Custom Margins", template.FontSize(18), template.Bold())
			c.Spacer(document.Mm(5))
			c.Text("This page has asymmetric margins: 10mm top/bottom, 40mm left/right. " +
				"The wide side margins create a narrower text area, similar to a book layout.")
			c.Spacer(document.Mm(5))
			c.Line()
			c.Spacer(document.Mm(5))
			c.Text("Lorem ipsum dolor sit amet, consectetur adipiscing elit. " +
				"Sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. " +
				"Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris.")
		})
	})

	testutil.GeneratePDFSharedGolden(t, "16_margins.pdf", doc)
}
