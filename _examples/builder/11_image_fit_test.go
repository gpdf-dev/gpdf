package builder_test

import (
	"image/color"
	"testing"

	"github.com/gpdf-dev/gpdf/_examples/testutil"
	"github.com/gpdf-dev/gpdf/document"
	"github.com/gpdf-dev/gpdf/template"
)

func TestExample_11_ImageFit(t *testing.T) {
	doc := template.New(
		template.WithPageSize(document.A4),
		template.WithMargins(document.UniformEdges(document.Mm(20))),
	)

	page := doc.AddPage()

	page.AutoRow(func(r *template.RowBuilder) {
		r.Col(12, func(c *template.ColBuilder) {
			c.Text("Image Fit Options", template.FontSize(18), template.Bold())
			c.Spacer(document.Mm(5))
		})
	})

	imgData := testutil.TestImagePNG(t, 300, 200, color.RGBA{R: 100, G: 149, B: 237, A: 255})

	// FitWidth
	page.AutoRow(func(r *template.RowBuilder) {
		r.Col(12, func(c *template.ColBuilder) {
			c.Text("FitWidth(80mm):")
			c.Spacer(document.Mm(2))
			c.Image(imgData, template.FitWidth(document.Mm(80)))
			c.Spacer(document.Mm(5))
		})
	})

	// FitHeight
	page.AutoRow(func(r *template.RowBuilder) {
		r.Col(12, func(c *template.ColBuilder) {
			c.Text("FitHeight(30mm):")
			c.Spacer(document.Mm(2))
			c.Image(imgData, template.FitHeight(document.Mm(30)))
			c.Spacer(document.Mm(5))
		})
	})

	// Default (no fit options)
	page.AutoRow(func(r *template.RowBuilder) {
		r.Col(12, func(c *template.ColBuilder) {
			c.Text("Default (no fit options):")
			c.Spacer(document.Mm(2))
			c.Image(imgData)
		})
	})

	testutil.GeneratePDFSharedGolden(t, "11_image_fit.pdf", doc)
}
