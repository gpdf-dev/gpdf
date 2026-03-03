package builder_test

import (
	"image/color"
	"testing"

	"github.com/gpdf-dev/gpdf/document"
	"github.com/gpdf-dev/gpdf/_examples/testutil"
	"github.com/gpdf-dev/gpdf/template"
)

func TestExample_10_Image(t *testing.T) {
	doc := template.New(
		template.WithPageSize(document.A4),
		template.WithMargins(document.UniformEdges(document.Mm(20))),
	)

	page := doc.AddPage()

	page.AutoRow(func(r *template.RowBuilder) {
		r.Col(12, func(c *template.ColBuilder) {
			c.Text("Image Examples", template.FontSize(18), template.Bold())
			c.Spacer(document.Mm(5))
		})
	})

	// Create test images
	pngData := testutil.TestImagePNG(t, 200, 100, color.RGBA{R: 66, G: 133, B: 244, A: 255})
	jpegData := testutil.TestImageJPEG(t, 200, 100, color.RGBA{R: 234, G: 67, B: 53, A: 255})

	// PNG image
	page.AutoRow(func(r *template.RowBuilder) {
		r.Col(12, func(c *template.ColBuilder) {
			c.Text("PNG image (blue):")
			c.Spacer(document.Mm(2))
			c.Image(pngData)
			c.Spacer(document.Mm(5))
		})
	})

	// JPEG image
	page.AutoRow(func(r *template.RowBuilder) {
		r.Col(12, func(c *template.ColBuilder) {
			c.Text("JPEG image (red):")
			c.Spacer(document.Mm(2))
			c.Image(jpegData)
			c.Spacer(document.Mm(5))
		})
	})

	// Images in columns
	page.AutoRow(func(r *template.RowBuilder) {
		r.Col(12, func(c *template.ColBuilder) {
			c.Text("Images side by side in grid columns:")
			c.Spacer(document.Mm(2))
		})
	})

	greenImg := testutil.TestImagePNG(t, 150, 80, color.RGBA{R: 52, G: 168, B: 83, A: 255})
	yellowImg := testutil.TestImagePNG(t, 150, 80, color.RGBA{R: 251, G: 188, B: 4, A: 255})

	page.AutoRow(func(r *template.RowBuilder) {
		r.Col(6, func(c *template.ColBuilder) {
			c.Text("Green PNG")
			c.Image(greenImg)
		})
		r.Col(6, func(c *template.ColBuilder) {
			c.Text("Yellow PNG")
			c.Image(yellowImg)
		})
	})

	testutil.GeneratePDFSharedGolden(t, "10_image.pdf", doc)
}
