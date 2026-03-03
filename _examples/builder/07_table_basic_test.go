package builder_test

import (
	"testing"

	"github.com/gpdf-dev/gpdf/document"
	"github.com/gpdf-dev/gpdf/_examples/testutil"
	"github.com/gpdf-dev/gpdf/template"
)

func TestExample_07_TableBasic(t *testing.T) {
	doc := template.New(
		template.WithPageSize(document.A4),
		template.WithMargins(document.UniformEdges(document.Mm(15))),
	)

	page := doc.AddPage()

	page.AutoRow(func(r *template.RowBuilder) {
		r.Col(12, func(c *template.ColBuilder) {
			c.Text("Basic Table", template.FontSize(18), template.Bold())
			c.Spacer(document.Mm(5))
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
					{"Diana", "28", "Paris"},
				},
			)
		})
	})

	testutil.GeneratePDF(t, "07_table_basic.pdf", doc)
}
