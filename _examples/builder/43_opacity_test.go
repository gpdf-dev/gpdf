package builder_test

import (
	"image/color"
	"testing"

	"github.com/gpdf-dev/gpdf/_examples/testutil"
	"github.com/gpdf-dev/gpdf/document"
	"github.com/gpdf-dev/gpdf/pdf"
	"github.com/gpdf-dev/gpdf/template"
)

func TestExample_43_Opacity(t *testing.T) {
	photo := testutil.TestImagePNG(t, 200, 100, color.RGBA{R: 234, G: 67, B: 53, A: 255})

	doc := template.New(
		template.WithPageSize(document.A4),
		template.WithMargins(document.UniformEdges(document.Mm(20))),
	)

	page := doc.AddPage()
	page.AutoRow(func(r *template.RowBuilder) {
		r.Col(12, func(c *template.ColBuilder) {
			c.Text("Opacity", template.FontSize(20), template.Bold())
			c.Spacer(document.Mm(6))
		})
	})

	page.AutoRow(func(r *template.RowBuilder) {
		r.Col(6, func(c *template.ColBuilder) {
			c.Box(func(c *template.ColBuilder) {
				c.Text("Box at 30%", template.TextColor(pdf.White), template.Bold())
			},
				template.WithBoxBackground(pdf.RGBHex(0x1A237E)),
				template.WithBoxPadding(document.UniformEdges(document.Mm(4))),
				template.WithBoxOpacity(0.3),
			)
		})
		r.Col(6, func(c *template.ColBuilder) {
			c.Text("Text at 40%", template.FontSize(24), template.WithTextOpacity(0.4))
		})
	})

	page.AutoRow(func(r *template.RowBuilder) {
		r.Col(12, func(c *template.ColBuilder) {
			c.Spacer(document.Mm(6))
			c.Image(photo, template.FitWidth(document.Mm(50)), template.WithImageOpacity(0.5))
		})
	})

	page.Absolute(document.Mm(20), document.Mm(120), func(c *template.ColBuilder) {
		c.Text("DRAFT", template.FontSize(96), template.Bold(),
			template.TextColor(pdf.RGBHex(0xFF0000)), template.WithTextOpacity(0.15))
	})

	testutil.GeneratePDFSharedGolden(t, "43_opacity.pdf", doc)
}
