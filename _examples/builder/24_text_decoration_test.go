package builder_test

import (
	"testing"

	"github.com/gpdf-dev/gpdf/document"
	"github.com/gpdf-dev/gpdf/pdf"
	"github.com/gpdf-dev/gpdf/_examples/testutil"
	"github.com/gpdf-dev/gpdf/template"
)

func TestExample_24_TextDecoration(t *testing.T) {
	doc := template.New(
		template.WithPageSize(document.A4),
		template.WithMargins(document.UniformEdges(document.Mm(20))),
	)

	page := doc.AddPage()
	page.AutoRow(func(r *template.RowBuilder) {
		r.Col(12, func(c *template.ColBuilder) {
			c.Text("Text Decoration Demo", template.FontSize(20), template.Bold())
			c.Spacer(document.Mm(8))

			c.Text("Normal text without decoration")
			c.Spacer(document.Mm(4))

			c.Text("Underlined text for emphasis", template.Underline())
			c.Spacer(document.Mm(4))

			c.Text("Strikethrough text for deletions", template.Strikethrough())
			c.Spacer(document.Mm(4))

			c.Text("Combined underline and strikethrough",
				template.Underline(), template.Strikethrough())
			c.Spacer(document.Mm(4))

			c.Text("Colored underlined text",
				template.Underline(),
				template.TextColor(pdf.RGBHex(0x1565C0)),
				template.FontSize(14))
			c.Spacer(document.Mm(4))

			c.Text("Bold underlined heading",
				template.Bold(), template.Underline(), template.FontSize(16))
		})
	})

	testutil.GeneratePDF(t, "24_text_decoration.pdf", doc)
}
