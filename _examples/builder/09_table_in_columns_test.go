package builder_test

import (
	"testing"

	"github.com/gpdf-dev/gpdf/document"
	"github.com/gpdf-dev/gpdf/_examples/testutil"
	"github.com/gpdf-dev/gpdf/template"
)

func TestExample_09_TableInColumns(t *testing.T) {
	doc := template.New(
		template.WithPageSize(document.A4),
		template.WithMargins(document.UniformEdges(document.Mm(15))),
	)

	page := doc.AddPage()

	page.AutoRow(func(r *template.RowBuilder) {
		r.Col(12, func(c *template.ColBuilder) {
			c.Text("Tables in Grid Columns", template.FontSize(18), template.Bold())
			c.Spacer(document.Mm(5))
		})
	})

	page.AutoRow(func(r *template.RowBuilder) {
		r.Col(6, func(c *template.ColBuilder) {
			c.Text("Team A", template.Bold())
			c.Spacer(document.Mm(2))
			c.Table(
				[]string{"Player", "Score"},
				[][]string{
					{"Alice", "95"},
					{"Bob", "87"},
					{"Charlie", "92"},
				},
				template.ColumnWidths(60, 40),
			)
		})
		r.Col(6, func(c *template.ColBuilder) {
			c.Text("Team B", template.Bold())
			c.Spacer(document.Mm(2))
			c.Table(
				[]string{"Player", "Score"},
				[][]string{
					{"Diana", "91"},
					{"Eve", "88"},
					{"Frank", "85"},
				},
				template.ColumnWidths(60, 40),
			)
		})
	})

	testutil.GeneratePDF(t, "09_table_in_columns.pdf", doc)
}
