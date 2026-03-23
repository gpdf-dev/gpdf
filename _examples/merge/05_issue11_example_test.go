package merge_test

import (
	"testing"

	gpdf "github.com/gpdf-dev/gpdf"
	"github.com/gpdf-dev/gpdf/_examples/testutil"
	"github.com/gpdf-dev/gpdf/document"
	"github.com/gpdf-dev/gpdf/template"
)

func TestExample_Merge_05_Issue11(t *testing.T) {
	// 1. Read existing PDFs (simulated with generated docs)
	coverPage := generateDoc(t, 1, "Cover Page")
	attachment := generateDoc(t, 5, "Terms and Conditions")

	// 2. Generate a document with gpdf as normal
	doc := template.New(
		template.WithPageSize(document.A4),
		template.WithMargins(document.Edges{
			Top:    document.Pt(40),
			Left:   document.Pt(40),
			Right:  document.Pt(40),
			Bottom: document.Pt(60),
		}),
	)
	page := doc.AddPage()
	page.AutoRow(func(r *template.RowBuilder) {
		r.Col(12, func(c *template.ColBuilder) {
			c.Text("Policy Certificate", template.FontSize(18), template.Bold())
		})
	})
	page.AutoRow(func(r *template.RowBuilder) {
		r.Col(12, func(c *template.ColBuilder) {
			c.Text("Lorem ipsum dolor sit amet, consectetur adipiscing elit. " +
				"Sed do eiusmod tempor incididunt ut labore et dolore magna aliqua.",
			)
		})
	})
	generated, err := doc.Generate()
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}

	// 3. Merge: cover + generated + attachment
	merged, err := gpdf.Merge(
		[]gpdf.Source{
			{Data: coverPage},
			{Data: generated},
			{Data: attachment, Pages: gpdf.PageRange{From: 1, To: 3}}, // only first 3 pages
		},
		gpdf.WithMergeMetadata("Policy Bundle", "Example Ltd", ""),
	)
	if err != nil {
		t.Fatalf("Merge: %v", err)
	}

	testutil.AssertValidPDF(t, merged)
	testutil.WritePDF(t, "05_issue11_example.pdf", merged)
}
