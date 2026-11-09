package builder_test

import (
	"testing"

	"github.com/gpdf-dev/gpdf/_examples/testutil"
	"github.com/gpdf-dev/gpdf/document"
	"github.com/gpdf-dev/gpdf/pdf"
	"github.com/gpdf-dev/gpdf/template"
)

func TestExample_40_RichText(t *testing.T) {
	doc := template.New(
		template.WithPageSize(document.A4),
		template.WithMargins(document.UniformEdges(document.Mm(20))),
	)

	page := doc.AddPage()
	page.AutoRow(func(r *template.RowBuilder) {
		r.Col(12, func(c *template.ColBuilder) {
			c.Text("Rich Text Demo", template.FontSize(20), template.Bold())
			c.Spacer(document.Mm(6))

			c.RichText(func(rt *template.RichTextBuilder) {
				rt.Span("Welcome to ", template.FontSize(12))
				rt.Span("gpdf Engine", template.FontSize(14), template.Bold(), template.TextColor(pdf.RGBHex(0x007BFF)))
				rt.Span("! This paragraph mixes inline styles.", template.FontSize(12), template.Italic())
			})
			c.Spacer(document.Mm(4))

			c.RichText(func(rt *template.RichTextBuilder) {
				rt.Span("Status: ")
				rt.Span("PAID", template.Bold(), template.TextColor(pdf.RGBHex(0x2E7D32)))
				rt.Span(" — ")
				rt.Span("old price", template.Strikethrough())
				rt.Span(" ")
				rt.Span("new price", template.Underline())
			}, template.AlignCenter())
		})
	})

	testutil.GeneratePDFSharedGolden(t, "40_rich_text.pdf", doc)
}
