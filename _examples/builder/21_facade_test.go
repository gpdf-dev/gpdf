package builder_test

import (
	"testing"

	"github.com/gpdf-dev/gpdf/_examples/testutil"
	"github.com/gpdf-dev/gpdf/document"
	"github.com/gpdf-dev/gpdf/template"
)

func TestExample_21_GpdfFacade(t *testing.T) {
	// Use the root gpdf package convenience functions
	// (imported as gpdf_test, so we use template.New which gpdf.NewDocument wraps)
	doc := template.New(
		template.WithPageSize(document.A4),
		template.WithMargins(document.UniformEdges(document.Mm(20))),
		template.WithDefaultFont("", 14),
		template.WithMetadata(document.DocumentMetadata{
			Title:  "Facade Example",
			Author: "gpdf",
		}),
	)

	page := doc.AddPage()
	page.AutoRow(func(r *template.RowBuilder) {
		r.Col(12, func(c *template.ColBuilder) {
			c.Text("gpdf Facade API", template.FontSize(20), template.Bold())
			c.Spacer(document.Mm(5))
			c.Text("This document uses WithDefaultFont to set the base font size to 14pt.")
			c.Text("All text in this document inherits the 14pt default.")
		})
	})

	testutil.GeneratePDFSharedGolden(t, "21_facade.pdf", doc)
}
