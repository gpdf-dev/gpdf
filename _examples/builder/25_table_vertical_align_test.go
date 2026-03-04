package builder_test

import (
	"testing"

	"github.com/gpdf-dev/gpdf/_examples/testutil"
	"github.com/gpdf-dev/gpdf/document"
	"github.com/gpdf-dev/gpdf/pdf"
	"github.com/gpdf-dev/gpdf/template"
)

func TestExample_25_TableVerticalAlign(t *testing.T) {
	doc := template.New(
		template.WithPageSize(document.A4),
		template.WithMargins(document.UniformEdges(document.Mm(20))),
	)

	page := doc.AddPage()
	page.AutoRow(func(r *template.RowBuilder) {
		r.Col(12, func(c *template.ColBuilder) {
			c.Text("Table Vertical Align Demo", template.FontSize(20), template.Bold())
			c.Spacer(document.Mm(8))

			// Default (top) alignment
			c.Text("Default (Top) Alignment:", template.Bold())
			c.Spacer(document.Mm(3))
			c.Table(
				[]string{"Short", "Tall Cell"},
				[][]string{
					{"A", "This cell has\nmuch more content\nthat spans\nmultiple lines"},
					{"B", "Another tall\ncell with\nlong text"},
				},
				template.TableHeaderStyle(
					template.BgColor(pdf.RGBHex(0x1565C0)),
					template.TextColor(pdf.White),
				),
			)
			c.Spacer(document.Mm(8))

			// Middle alignment
			c.Text("Middle Alignment:", template.Bold())
			c.Spacer(document.Mm(3))
			c.Table(
				[]string{"Short", "Tall Cell"},
				[][]string{
					{"A", "This cell has\nmuch more content\nthat spans\nmultiple lines"},
					{"B", "Another tall\ncell with\nlong text"},
				},
				template.TableHeaderStyle(
					template.BgColor(pdf.RGBHex(0x2E7D32)),
					template.TextColor(pdf.White),
				),
				template.TableCellVAlign(document.VAlignMiddle),
			)
			c.Spacer(document.Mm(8))

			// Bottom alignment
			c.Text("Bottom Alignment:", template.Bold())
			c.Spacer(document.Mm(3))
			c.Table(
				[]string{"Short", "Tall Cell"},
				[][]string{
					{"A", "This cell has\nmuch more content\nthat spans\nmultiple lines"},
					{"B", "Another tall\ncell with\nlong text"},
				},
				template.TableHeaderStyle(
					template.BgColor(pdf.RGBHex(0xE65100)),
					template.TextColor(pdf.White),
				),
				template.TableCellVAlign(document.VAlignBottom),
			)
		})
	})

	testutil.GeneratePDFSharedGolden(t, "25_table_vertical_align.pdf", doc)
}
