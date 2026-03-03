package builder_test

import (
	"testing"

	"github.com/gpdf-dev/gpdf/document"
	"github.com/gpdf-dev/gpdf/pdf"
	"github.com/gpdf-dev/gpdf/_examples/testutil"
	"github.com/gpdf-dev/gpdf/template"
)

func TestExample_08_TableStyled(t *testing.T) {
	doc := template.New(
		template.WithPageSize(document.A4),
		template.WithMargins(document.UniformEdges(document.Mm(15))),
	)

	page := doc.AddPage()

	page.AutoRow(func(r *template.RowBuilder) {
		r.Col(12, func(c *template.ColBuilder) {
			c.Text("Styled Table", template.FontSize(18), template.Bold())
			c.Spacer(document.Mm(5))
		})
	})

	darkBlue := pdf.RGBHex(0x1A237E)
	lightGray := pdf.RGBHex(0xF5F5F5)

	page.AutoRow(func(r *template.RowBuilder) {
		r.Col(12, func(c *template.ColBuilder) {
			c.Table(
				[]string{"Product", "Category", "Qty", "Unit Price", "Total"},
				[][]string{
					{"Laptop Pro 15", "Electronics", "2", "$1,299.00", "$2,598.00"},
					{"Wireless Mouse", "Accessories", "10", "$29.99", "$299.90"},
					{"USB-C Hub", "Accessories", "5", "$49.99", "$249.95"},
					{"Monitor 27\"", "Electronics", "3", "$399.00", "$1,197.00"},
					{"Keyboard", "Accessories", "10", "$79.99", "$799.90"},
					{"Webcam HD", "Electronics", "4", "$89.99", "$359.96"},
				},
				template.ColumnWidths(30, 20, 10, 20, 20),
				template.TableHeaderStyle(
					template.TextColor(pdf.White),
					template.BgColor(darkBlue),
				),
				template.TableStripe(lightGray),
			)
		})
	})

	testutil.GeneratePDFSharedGolden(t, "08_table_styled.pdf", doc)
}
