package merge_test

import (
	"fmt"
	"testing"

	gpdf "github.com/gpdf-dev/gpdf"
	"github.com/gpdf-dev/gpdf/_examples/testutil"
	"github.com/gpdf-dev/gpdf/document"
	"github.com/gpdf-dev/gpdf/template"
)

// generateDoc creates a PDF with the given number of pages and a title prefix.
func generateDoc(t *testing.T, numPages int, prefix string) []byte {
	t.Helper()
	doc := template.New(
		template.WithPageSize(document.A4),
		template.WithMargins(document.UniformEdges(document.Mm(20))),
		template.WithMetadata(document.DocumentMetadata{
			Title:  prefix,
			Author: "gpdf",
		}),
	)

	for i := 1; i <= numPages; i++ {
		page := doc.AddPage()
		page.AutoRow(func(r *template.RowBuilder) {
			r.Col(12, func(c *template.ColBuilder) {
				c.Text(fmt.Sprintf("%s — Page %d", prefix, i),
					template.FontSize(18), template.Bold())
			})
		})
		page.AutoRow(func(r *template.RowBuilder) {
			r.Col(12, func(c *template.ColBuilder) {
				c.Spacer(document.Mm(5))
				c.Line()
				c.Spacer(document.Mm(5))
			})
		})
		for j := 0; j < 3; j++ {
			page.AutoRow(func(r *template.RowBuilder) {
				r.Col(12, func(c *template.ColBuilder) {
					c.Text("Lorem ipsum dolor sit amet, consectetur adipiscing elit. " +
						"Sed do eiusmod tempor incididunt ut labore et dolore magna aliqua.",
					)
				})
			})
		}
	}

	data, err := doc.Generate()
	if err != nil {
		t.Fatalf("generate %s: %v", prefix, err)
	}
	return data
}

func TestExample_Merge_01_Basic(t *testing.T) {
	// Generate two separate PDFs.
	cover := generateDoc(t, 1, "Cover Page")
	body := generateDoc(t, 3, "Report Body")

	// Merge them into a single document.
	merged, err := gpdf.Merge([]gpdf.Source{
		{Data: cover},
		{Data: body},
	})
	if err != nil {
		t.Fatalf("Merge: %v", err)
	}

	testutil.AssertValidPDF(t, merged)
	testutil.WritePDF(t, "01_basic_merge.pdf", merged)
}
