package builder_test

import (
	"image/color"
	"testing"

	"github.com/gpdf-dev/gpdf/_examples/testutil"
	"github.com/gpdf-dev/gpdf/document"
	"github.com/gpdf-dev/gpdf/template"
)

// TestExample_34_ImageMinSize verifies that an image with MinDisplayHeight
// overflows to the next page when the remaining space would force it below
// the configured minimum, instead of being shrunk to fit.
func TestExample_34_ImageMinSize(t *testing.T) {
	doc := template.New(
		template.WithPageSize(document.A4),
		template.WithMargins(document.UniformEdges(document.Mm(20))),
	)

	page := doc.AddPage()

	page.AutoRow(func(r *template.RowBuilder) {
		r.Col(12, func(c *template.ColBuilder) {
			c.Text("Image Minimum Display Size", template.FontSize(18), template.Bold())
			c.Spacer(document.Mm(5))
		})
	})

	// Fill most of the page so the image's natural height cannot fit.
	for i := 0; i < 23; i++ {
		page.AutoRow(func(r *template.RowBuilder) {
			r.Col(12, func(c *template.ColBuilder) {
				c.Text("Lorem ipsum dolor sit amet, consectetur adipiscing elit. " +
					"Sed do eiusmod tempor incididunt ut labore et dolore magna aliqua.")
			})
		})
	}

	// A tall image (300x500 px → 0.6 aspect ratio). Displayed at 80mm width
	// its natural height is ~133mm. The remaining space on page 1 is far
	// smaller than the 100mm minimum, so the image must move to page 2.
	imgData := testutil.TestImagePNG(t, 300, 500, color.RGBA{R: 100, G: 149, B: 237, A: 255})

	page.AutoRow(func(r *template.RowBuilder) {
		r.Col(12, func(c *template.ColBuilder) {
			c.Image(imgData,
				template.FitWidth(document.Mm(80)),
				template.MinDisplayHeight(document.Mm(100)),
			)
		})
	})

	testutil.GeneratePDFSharedGolden(t, "34_image_min_size.pdf", doc)
}
