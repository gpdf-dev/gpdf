package builder_test

import (
	"fmt"
	"testing"

	"github.com/gpdf-dev/gpdf/_examples/testutil"
	"github.com/gpdf-dev/gpdf/document"
	"github.com/gpdf-dev/gpdf/pdf"
	"github.com/gpdf-dev/gpdf/template"
)

// TestExample_33_MultiPageTable verifies whether a large table automatically
// splits across multiple pages when its content exceeds the available space.
func TestExample_33_MultiPageTable(t *testing.T) {
	doc := template.New(
		template.WithPageSize(document.A4),
		template.WithMargins(document.UniformEdges(document.Mm(15))),
	)

	// Header/footer to verify they repeat on every page.
	doc.Header(func(p *template.PageBuilder) {
		p.AutoRow(func(r *template.RowBuilder) {
			r.Col(6, func(c *template.ColBuilder) {
				c.Text("Multi-Page Table Test", template.Bold(), template.FontSize(9))
			})
			r.Col(6, func(c *template.ColBuilder) {
				c.Text("Header", template.AlignRight(), template.FontSize(9),
					template.TextColor(pdf.Gray(0.5)))
			})
		})
		p.AutoRow(func(r *template.RowBuilder) {
			r.Col(12, func(c *template.ColBuilder) {
				c.Line()
				c.Spacer(document.Mm(3))
			})
		})
	})

	doc.Footer(func(p *template.PageBuilder) {
		p.AutoRow(func(r *template.RowBuilder) {
			r.Col(12, func(c *template.ColBuilder) {
				c.Spacer(document.Mm(3))
				c.Line()
			})
		})
		p.AutoRow(func(r *template.RowBuilder) {
			r.Col(12, func(c *template.ColBuilder) {
				c.Text("Footer - Page Break Test",
					template.AlignCenter(), template.FontSize(8), template.TextColor(pdf.Gray(0.5)))
			})
		})
	})

	// Generate 100 rows — far more than a single A4 page can hold.
	rows := make([][]string, 100)
	for i := range rows {
		rows[i] = []string{
			fmt.Sprintf("%d", i+1),
			fmt.Sprintf("Product-%04d", i+1),
			fmt.Sprintf("Category %c", 'A'+rune(i%5)),
			fmt.Sprintf("$%d.%02d", 10+i*3, (i*7)%100),
			fmt.Sprintf("%d", 100+i*10),
		}
	}

	page := doc.AddPage()

	page.AutoRow(func(r *template.RowBuilder) {
		r.Col(12, func(c *template.ColBuilder) {
			c.Text("Large Table — Auto Page Break Test", template.FontSize(16), template.Bold())
			c.Spacer(document.Mm(5))
		})
	})

	page.AutoRow(func(r *template.RowBuilder) {
		r.Col(12, func(c *template.ColBuilder) {
			c.Table(
				[]string{"#", "Product", "Category", "Price", "Stock"},
				rows,
				template.ColumnWidths(8, 30, 22, 20, 20),
				template.TableHeaderStyle(
					template.TextColor(pdf.White),
					template.BgColor(pdf.RGBHex(0x1A237E)),
				),
				template.TableStripe(pdf.RGBHex(0xF5F5F5)),
			)
		})
	})

	data, err := doc.Generate()
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}
	testutil.AssertValidPDF(t, data)
	testutil.WritePDF(t, "33_multi_page_table.pdf", data)

	// Log the PDF size so we can check it produced meaningful output.
	t.Logf("Generated PDF: %d bytes", len(data))
}
