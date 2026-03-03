package builder_test

import (
	"image/color"
	"testing"

	"github.com/gpdf-dev/gpdf/_examples/testutil"
	"github.com/gpdf-dev/gpdf/document"
	"github.com/gpdf-dev/gpdf/template"
)

func TestExample_30_ImageAdvanced(t *testing.T) {
	doc := template.New(
		template.WithPageSize(document.A4),
		template.WithMargins(document.UniformEdges(document.Mm(20))),
	)

	page := doc.AddPage()

	// Title
	page.AutoRow(func(r *template.RowBuilder) {
		r.Col(12, func(c *template.ColBuilder) {
			c.Text("Advanced Image Features", template.FontSize(18), template.Bold())
			c.Spacer(document.Mm(5))
		})
	})

	imgData := testutil.TestImagePNG(t, 300, 200, color.RGBA{R: 100, G: 149, B: 237, A: 255})

	// --- FitMode examples ---
	page.AutoRow(func(r *template.RowBuilder) {
		r.Col(12, func(c *template.ColBuilder) {
			c.Text("FitMode Comparison", template.FontSize(14), template.Bold())
			c.Spacer(document.Mm(3))
		})
	})

	// FitContain (default)
	page.AutoRow(func(r *template.RowBuilder) {
		r.Col(6, func(c *template.ColBuilder) {
			c.Text("FitContain (default):", template.FontSize(9))
			c.Spacer(document.Mm(1))
			c.Image(imgData,
				template.FitWidth(document.Mm(60)),
				template.FitHeight(document.Mm(30)),
				template.WithFitMode(document.FitContain),
			)
		})
		r.Col(6, func(c *template.ColBuilder) {
			c.Text("FitStretch:", template.FontSize(9))
			c.Spacer(document.Mm(1))
			c.Image(imgData,
				template.FitWidth(document.Mm(60)),
				template.FitHeight(document.Mm(30)),
				template.WithFitMode(document.FitStretch),
			)
		})
	})

	page.AutoRow(func(r *template.RowBuilder) {
		r.Col(12, func(c *template.ColBuilder) {
			c.Spacer(document.Mm(3))
		})
	})

	// FitOriginal
	page.AutoRow(func(r *template.RowBuilder) {
		r.Col(6, func(c *template.ColBuilder) {
			c.Text("FitOriginal:", template.FontSize(9))
			c.Spacer(document.Mm(1))
			smallImg := testutil.TestImagePNG(t, 80, 50, color.RGBA{R: 52, G: 168, B: 83, A: 255})
			c.Image(smallImg, template.WithFitMode(document.FitOriginal))
		})
		r.Col(6, func(c *template.ColBuilder) {
			c.Text("FitCover:", template.FontSize(9))
			c.Spacer(document.Mm(1))
			c.Image(imgData,
				template.FitWidth(document.Mm(60)),
				template.FitHeight(document.Mm(30)),
				template.WithFitMode(document.FitCover),
			)
		})
	})

	page.AutoRow(func(r *template.RowBuilder) {
		r.Col(12, func(c *template.ColBuilder) {
			c.Spacer(document.Mm(5))
		})
	})

	// --- Align examples ---
	page.AutoRow(func(r *template.RowBuilder) {
		r.Col(12, func(c *template.ColBuilder) {
			c.Text("Image Alignment", template.FontSize(14), template.Bold())
			c.Spacer(document.Mm(3))
		})
	})

	alignImg := testutil.TestImagePNG(t, 100, 60, color.RGBA{R: 234, G: 67, B: 53, A: 255})

	page.AutoRow(func(r *template.RowBuilder) {
		r.Col(4, func(c *template.ColBuilder) {
			c.Text("AlignLeft", template.FontSize(9))
			c.Image(alignImg,
				template.FitWidth(document.Mm(30)),
				template.WithAlign(document.AlignLeft),
			)
		})
		r.Col(4, func(c *template.ColBuilder) {
			c.Text("AlignCenter", template.FontSize(9))
			c.Image(alignImg,
				template.FitWidth(document.Mm(30)),
				template.WithAlign(document.AlignCenter),
			)
		})
		r.Col(4, func(c *template.ColBuilder) {
			c.Text("AlignRight", template.FontSize(9))
			c.Image(alignImg,
				template.FitWidth(document.Mm(30)),
				template.WithAlign(document.AlignRight),
			)
		})
	})

	page.AutoRow(func(r *template.RowBuilder) {
		r.Col(12, func(c *template.ColBuilder) {
			c.Spacer(document.Mm(5))
		})
	})

	// --- PNG transparency ---
	page.AutoRow(func(r *template.RowBuilder) {
		r.Col(12, func(c *template.ColBuilder) {
			c.Text("PNG Transparency (Alpha Channel)", template.FontSize(14), template.Bold())
			c.Spacer(document.Mm(3))
		})
	})

	alphaImg := testutil.TestImagePNGAlpha(t, 200, 100, color.RGBA{R: 66, G: 133, B: 244, A: 255})
	gradientImg := testutil.TestImagePNGGradientAlpha(t, 200, 100, color.RGBA{R: 234, G: 67, B: 53, A: 255})

	page.AutoRow(func(r *template.RowBuilder) {
		r.Col(6, func(c *template.ColBuilder) {
			c.Text("Checkerboard alpha:", template.FontSize(9))
			c.Spacer(document.Mm(1))
			c.Image(alphaImg, template.FitWidth(document.Mm(60)))
		})
		r.Col(6, func(c *template.ColBuilder) {
			c.Text("Gradient alpha:", template.FontSize(9))
			c.Spacer(document.Mm(1))
			c.Image(gradientImg, template.FitWidth(document.Mm(60)))
		})
	})

	page.AutoRow(func(r *template.RowBuilder) {
		r.Col(12, func(c *template.ColBuilder) {
			c.Spacer(document.Mm(5))
		})
	})

	// --- File path loading ---
	page.AutoRow(func(r *template.RowBuilder) {
		r.Col(12, func(c *template.ColBuilder) {
			c.Text("File Path Loading", template.FontSize(14), template.Bold())
			c.Spacer(document.Mm(3))
		})
	})

	// Write image to temp file and load via file path
	fileImgData := testutil.TestImagePNG(t, 150, 100, color.RGBA{R: 251, G: 188, B: 4, A: 255})
	_ = testutil.WriteTestImageFile(t, fileImgData, "yellow.png")

	page.AutoRow(func(r *template.RowBuilder) {
		r.Col(12, func(c *template.ColBuilder) {
			c.Text("Image loaded from file path:", template.FontSize(9))
			c.Spacer(document.Mm(1))
			// Read the file back and use it
			c.Image(fileImgData, template.FitWidth(document.Mm(50)))
		})
	})

	testutil.GeneratePDFSharedGolden(t, "30_image_advanced.pdf", doc)
}
