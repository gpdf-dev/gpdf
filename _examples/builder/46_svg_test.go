package builder_test

import (
	"os"
	"testing"

	"github.com/gpdf-dev/gpdf/_examples/testutil"
	"github.com/gpdf-dev/gpdf/document"
	"github.com/gpdf-dev/gpdf/template"
)

func readSVG(t *testing.T, name string) []byte {
	t.Helper()
	data, err := os.ReadFile("../testdata/svg/" + name)
	if err != nil {
		t.Fatal(err)
	}
	return data
}

func TestExample_46_SVG(t *testing.T) {
	logo := readSVG(t, "logo.svg")
	icons := readSVG(t, "icons.svg")
	chart := readSVG(t, "chart.svg")

	doc := template.New(
		template.WithPageSize(document.A4),
		template.WithMargins(document.UniformEdges(document.Mm(20))),
	)

	page := doc.AddPage()
	page.AutoRow(func(r *template.RowBuilder) {
		r.Col(12, func(c *template.ColBuilder) {
			c.Text("SVG Images", template.FontSize(20), template.Bold())
			c.Text("Embedded as vector graphics: sharp at any zoom level.")
			c.Spacer(document.Mm(6))
		})
	})

	page.AutoRow(func(r *template.RowBuilder) {
		r.Col(4, func(c *template.ColBuilder) {
			c.Image(logo, template.FitWidth(document.Mm(40)))
		})
		r.Col(8, func(c *template.ColBuilder) {
			c.Image(chart, template.FitWidth(document.Mm(100)))
		})
	})

	page.AutoRow(func(r *template.RowBuilder) {
		r.Col(12, func(c *template.ColBuilder) {
			c.Spacer(document.Mm(8))
			c.Image(icons, template.FitWidth(document.Mm(60)))
			c.Spacer(document.Mm(4))
			c.Image(icons, template.FitWidth(document.Mm(150)))
		})
	})

	testutil.GeneratePDFSharedGolden(t, "46_svg.pdf", doc)
}
