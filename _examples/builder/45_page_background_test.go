package builder_test

import (
	"image/color"
	"testing"

	"github.com/gpdf-dev/gpdf/_examples/testutil"
	"github.com/gpdf-dev/gpdf/document"
	"github.com/gpdf-dev/gpdf/pdf"
	"github.com/gpdf-dev/gpdf/template"
)

func TestExample_45_PageBackground(t *testing.T) {
	letterhead := testutil.TestImagePNG(t, 300, 100, color.RGBA{R: 255, G: 213, B: 79, A: 255})

	doc := template.New(
		template.WithPageSize(document.A4),
		template.WithMargins(document.UniformEdges(document.Mm(20))),
		template.WithBackgroundColor(pdf.RGBHex(0xFFFDF5)),
		template.WithBackgroundImage(letterhead,
			template.BackgroundFit(document.FitContain),
			template.BackgroundOpacity(0.25),
		),
	)

	for _, title := range []string{"Page Background (page 1)", "Page Background (page 2)"} {
		page := doc.AddPage()
		page.AutoRow(func(r *template.RowBuilder) {
			r.Col(12, func(c *template.ColBuilder) {
				c.Text(title, template.FontSize(20), template.Bold())
				c.Spacer(document.Mm(4))
				c.Text("The background is drawn beneath the content of every page, ignoring margins.")
			})
		})
	}

	testutil.GeneratePDFSharedGolden(t, "45_page_background.pdf", doc)
}
