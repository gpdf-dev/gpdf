package builder_test

import (
	"testing"

	"github.com/gpdf-dev/gpdf/_examples/testutil"
	"github.com/gpdf-dev/gpdf/document"
	"github.com/gpdf-dev/gpdf/template"
)

// A table that starts with less room than one row at the bottom of a page
// moves to the next page instead of drawing its rows on top of each other.
func TestExample_47_TablePageBottom(t *testing.T) {
	doc := template.New(
		template.WithPageSize(document.A4),
		template.WithMargins(document.UniformEdges(document.Mm(15))),
	)

	page := doc.AddPage()

	// The A4 body is 267mm tall; leave 1mm for the table.
	page.AutoRow(func(r *template.RowBuilder) {
		r.Col(12, func(c *template.ColBuilder) {
			c.Spacer(document.Mm(266))
		})
	})

	page.AutoRow(func(r *template.RowBuilder) {
		r.Col(12, func(c *template.ColBuilder) {
			c.Table(
				[]string{"Name", "Age", "City"},
				[][]string{
					{"Alice", "30", "Tokyo"},
					{"Bob", "25", "New York"},
					{"Charlie", "35", "London"},
				},
			)
		})
	})

	testutil.GeneratePDFSharedGolden(t, "47_table_page_bottom.pdf", doc)
}
