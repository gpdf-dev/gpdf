package builder_test

import (
	"image/color"
	"testing"

	"github.com/gpdf-dev/gpdf/_examples/testutil"
	"github.com/gpdf-dev/gpdf/document"
	"github.com/gpdf-dev/gpdf/pdf"
	"github.com/gpdf-dev/gpdf/template"
)

func TestExample_44_Links(t *testing.T) {
	logo := testutil.TestImagePNG(t, 120, 40, color.RGBA{R: 26, G: 35, B: 126, A: 255})

	doc := template.New(
		template.WithPageSize(document.A4),
		template.WithMargins(document.UniformEdges(document.Mm(20))),
	)

	page := doc.AddPage()
	page.AutoRow(func(r *template.RowBuilder) {
		r.Col(12, func(c *template.ColBuilder) {
			c.Text("Links", template.FontSize(20), template.Bold())
			c.Spacer(document.Mm(6))
			c.Link("Visit gpdf.dev", "https://gpdf.dev")
			c.Spacer(document.Mm(3))
			c.RichText(func(rt *template.RichTextBuilder) {
				rt.Span("Read the ")
				rt.Span("documentation",
					template.WithTextLink("https://gpdf.dev/docs"),
					template.TextColor(pdf.RGBHex(0x0645AD)),
					template.Underline(),
				)
				rt.Span(" for every option.")
			})
			c.Spacer(document.Mm(3))
			c.Text("Plain-looking text can link too.", template.WithTextLink("https://github.com/gpdf-dev/gpdf"))
			c.Spacer(document.Mm(6))
		})
	})

	page.AutoRow(func(r *template.RowBuilder) {
		r.Col(6, func(c *template.ColBuilder) {
			c.Image(logo, template.FitWidth(document.Mm(40)), template.WithImageLink("https://gpdf.dev"))
		})
		r.Col(6, func(c *template.ColBuilder) {
			c.Box(func(c *template.ColBuilder) {
				c.Text("The whole card is clickable")
			},
				template.WithBoxBackground(pdf.RGBHex(0xE8EAF6)),
				template.WithBoxPadding(document.UniformEdges(document.Mm(4))),
				template.WithBoxLink("https://gpdf.dev/blog"),
			)
		})
	})

	testutil.GeneratePDFSharedGolden(t, "44_links.pdf", doc)
}
