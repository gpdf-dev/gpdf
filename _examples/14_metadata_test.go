package examples_test

import (
	"testing"

	"github.com/gpdf-dev/gpdf/document"
	"github.com/gpdf-dev/gpdf/template"
)

func TestExample_14_Metadata(t *testing.T) {
	doc := template.New(
		template.WithPageSize(document.A4),
		template.WithMargins(document.UniformEdges(document.Mm(20))),
		template.WithMetadata(document.DocumentMetadata{
			Title:   "Annual Report 2026",
			Author:  "gpdf Library",
			Subject: "Example of document metadata",
			Creator: "gpdf example_test.go",
		}),
	)

	page := doc.AddPage()
	page.AutoRow(func(r *template.RowBuilder) {
		r.Col(12, func(c *template.ColBuilder) {
			c.Text("Document with Metadata", template.FontSize(20), template.Bold())
			c.Spacer(document.Mm(5))
			c.Text("This PDF has the following metadata set:")
			c.Spacer(document.Mm(3))
			c.Text("Title: Annual Report 2026")
			c.Text("Author: gpdf Library")
			c.Text("Subject: Example of document metadata")
			c.Text("Creator: gpdf example_test.go")
			c.Text("Producer: gpdf (set automatically)")
			c.Spacer(document.Mm(5))
			c.Text("Open the PDF properties in your viewer to verify.", template.Italic())
		})
	})

	generatePDF(t, "14_metadata.pdf", doc)
}
