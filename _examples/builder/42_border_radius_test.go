package builder_test

import (
	"image/color"
	"testing"

	"github.com/gpdf-dev/gpdf/_examples/testutil"
	"github.com/gpdf-dev/gpdf/document"
	"github.com/gpdf-dev/gpdf/pdf"
	"github.com/gpdf-dev/gpdf/template"
)

func TestExample_42_BorderRadius(t *testing.T) {
	banner := testutil.TestImagePNG(t, 200, 100, color.RGBA{R: 66, G: 133, B: 244, A: 255})
	avatar := testutil.TestImagePNG(t, 100, 100, color.RGBA{R: 52, G: 168, B: 83, A: 255})

	doc := template.New(
		template.WithPageSize(document.A4),
		template.WithMargins(document.UniformEdges(document.Mm(20))),
	)

	page := doc.AddPage()
	page.AutoRow(func(r *template.RowBuilder) {
		r.Col(12, func(c *template.ColBuilder) {
			c.Text("Border Radius", template.FontSize(20), template.Bold())
			c.Spacer(document.Mm(6))
		})
	})

	page.AutoRow(func(r *template.RowBuilder) {
		r.Col(4, func(c *template.ColBuilder) {
			c.Box(func(c *template.ColBuilder) {
				c.Text("Filled card")
			},
				template.WithBoxBackground(pdf.RGBHex(0xE8EAF6)),
				template.WithBoxPadding(document.UniformEdges(document.Mm(4))),
				template.WithBoxRadius(document.Mm(4)),
			)
		})
		r.Col(4, func(c *template.ColBuilder) {
			c.Box(func(c *template.ColBuilder) {
				c.Text("Outlined card")
			},
				template.WithBoxBorder(template.Border(
					template.BorderWidth(document.Pt(1)),
					template.BorderColor(pdf.RGBHex(0x1A237E)),
				)),
				template.WithBoxPadding(document.UniformEdges(document.Mm(4))),
				template.WithBoxRadius(document.Mm(3)),
			)
		})
		r.Col(4, func(c *template.ColBuilder) {
			c.Text("Pill label",
				template.BgColor(pdf.RGBHex(0xFFE0B2)),
				template.TextPadding(document.UniformEdges(document.Mm(2))),
				template.WithTextRadius(document.Pt(8)),
				template.AlignCenter(),
			)
		})
	})

	page.AutoRow(func(r *template.RowBuilder) {
		r.Col(12, func(c *template.ColBuilder) {
			c.Spacer(document.Mm(8))
		})
	})

	page.AutoRow(func(r *template.RowBuilder) {
		r.Col(6, func(c *template.ColBuilder) {
			c.Image(banner, template.FitWidth(document.Mm(60)), template.WithImageRadius(document.Mm(5)))
		})
		r.Col(6, func(c *template.ColBuilder) {
			c.Image(avatar, template.FitWidth(document.Mm(30)), template.WithImageRadius(document.Mm(15)))
		})
	})

	testutil.GeneratePDFSharedGolden(t, "42_border_radius.pdf", doc)
}
