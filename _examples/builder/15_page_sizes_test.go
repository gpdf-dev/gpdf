package builder_test

import (
	"testing"

	"github.com/gpdf-dev/gpdf/_examples/testutil"
	"github.com/gpdf-dev/gpdf/document"
	"github.com/gpdf-dev/gpdf/template"
)

func TestExample_15_PageSizes(t *testing.T) {
	sizes := []struct {
		name string
		size document.Size
		file string
	}{
		{"A4 (210mm x 297mm)", document.A4, "15a_pagesize_a4.pdf"},
		{"A3 (297mm x 420mm)", document.A3, "15b_pagesize_a3.pdf"},
		{"Letter (8.5in x 11in)", document.Letter, "15c_pagesize_letter.pdf"},
		{"Legal (8.5in x 14in)", document.Legal, "15d_pagesize_legal.pdf"},
	}

	for _, s := range sizes {
		t.Run(s.name, func(t *testing.T) {
			doc := template.New(
				template.WithPageSize(s.size),
				template.WithMargins(document.UniformEdges(document.Mm(20))),
			)

			page := doc.AddPage()
			page.AutoRow(func(r *template.RowBuilder) {
				r.Col(12, func(c *template.ColBuilder) {
					c.Text("Page Size: "+s.name, template.FontSize(20), template.Bold())
					c.Spacer(document.Mm(10))
					c.Text("This page demonstrates the " + s.name + " page format.")
				})
			})

			testutil.GeneratePDFSharedGolden(t, s.file, doc)
		})
	}
}
