package builder_test

import (
	"testing"

	"github.com/gpdf-dev/gpdf/document"
	"github.com/gpdf-dev/gpdf/pdf"
	"github.com/gpdf-dev/gpdf/_examples/testutil"
	"github.com/gpdf-dev/gpdf/template"
)

func TestExample_18_Invoice(t *testing.T) {
	doc := template.New(
		template.WithPageSize(document.A4),
		template.WithMargins(document.UniformEdges(document.Mm(20))),
		template.WithMetadata(document.DocumentMetadata{
			Title:  "Invoice #INV-2026-001",
			Author: "ACME Corporation",
		}),
	)

	page := doc.AddPage()

	// Company header
	page.AutoRow(func(r *template.RowBuilder) {
		r.Col(6, func(c *template.ColBuilder) {
			c.Text("ACME Corporation", template.FontSize(24), template.Bold(),
				template.TextColor(pdf.RGBHex(0x1A237E)))
			c.Text("123 Business Street")
			c.Text("Suite 100")
			c.Text("San Francisco, CA 94105")
		})
		r.Col(6, func(c *template.ColBuilder) {
			c.Text("INVOICE", template.FontSize(28), template.Bold(), template.AlignRight(),
				template.TextColor(pdf.RGBHex(0x1A237E)))
			c.Spacer(document.Mm(3))
			c.Text("#INV-2026-001", template.AlignRight(), template.FontSize(12))
			c.Text("Date: March 1, 2026", template.AlignRight())
			c.Text("Due: March 31, 2026", template.AlignRight())
		})
	})

	page.AutoRow(func(r *template.RowBuilder) {
		r.Col(12, func(c *template.ColBuilder) {
			c.Spacer(document.Mm(5))
			c.Line(template.LineColor(pdf.RGBHex(0x1A237E)), template.LineThickness(document.Pt(2)))
			c.Spacer(document.Mm(5))
		})
	})

	// Bill to
	page.AutoRow(func(r *template.RowBuilder) {
		r.Col(6, func(c *template.ColBuilder) {
			c.Text("Bill To:", template.Bold(), template.TextColor(pdf.Gray(0.4)))
			c.Spacer(document.Mm(2))
			c.Text("John Smith", template.Bold())
			c.Text("Tech Solutions Inc.")
			c.Text("456 Client Avenue")
			c.Text("New York, NY 10001")
		})
		r.Col(6, func(c *template.ColBuilder) {
			c.Text("Payment Info:", template.Bold(), template.TextColor(pdf.Gray(0.4)))
			c.Spacer(document.Mm(2))
			c.Text("Bank: First National Bank")
			c.Text("Account: 1234-5678-9012")
			c.Text("Routing: 021000021")
		})
	})

	page.AutoRow(func(r *template.RowBuilder) {
		r.Col(12, func(c *template.ColBuilder) {
			c.Spacer(document.Mm(10))
		})
	})

	// Items table
	headerBlue := pdf.RGBHex(0x1A237E)
	stripeGray := pdf.RGBHex(0xF5F5F5)

	page.AutoRow(func(r *template.RowBuilder) {
		r.Col(12, func(c *template.ColBuilder) {
			c.Table(
				[]string{"Description", "Qty", "Unit Price", "Amount"},
				[][]string{
					{"Web Development - Frontend", "40 hrs", "$150.00", "$6,000.00"},
					{"Web Development - Backend", "60 hrs", "$150.00", "$9,000.00"},
					{"UI/UX Design", "20 hrs", "$120.00", "$2,400.00"},
					{"Database Design", "15 hrs", "$130.00", "$1,950.00"},
					{"QA Testing", "25 hrs", "$100.00", "$2,500.00"},
					{"Project Management", "10 hrs", "$140.00", "$1,400.00"},
				},
				template.ColumnWidths(40, 15, 20, 25),
				template.TableHeaderStyle(
					template.TextColor(pdf.White),
					template.BgColor(headerBlue),
				),
				template.TableStripe(stripeGray),
			)
		})
	})

	page.AutoRow(func(r *template.RowBuilder) {
		r.Col(12, func(c *template.ColBuilder) {
			c.Spacer(document.Mm(5))
		})
	})

	// Totals
	page.AutoRow(func(r *template.RowBuilder) {
		r.Col(8, func(c *template.ColBuilder) {
			// empty left side
		})
		r.Col(4, func(c *template.ColBuilder) {
			c.Text("Subtotal:    $23,250.00", template.AlignRight())
			c.Text("Tax (10%):    $2,325.00", template.AlignRight())
			c.Spacer(document.Mm(2))
			c.Line(template.LineThickness(document.Pt(1)))
			c.Spacer(document.Mm(2))
			c.Text("Total:       $25,575.00", template.AlignRight(),
				template.Bold(), template.FontSize(14))
		})
	})

	page.AutoRow(func(r *template.RowBuilder) {
		r.Col(12, func(c *template.ColBuilder) {
			c.Spacer(document.Mm(15))
			c.Line(template.LineColor(pdf.Gray(0.8)))
			c.Spacer(document.Mm(3))
			c.Text("Thank you for your business!", template.AlignCenter(),
				template.Italic(), template.TextColor(pdf.Gray(0.5)))
		})
	})

	testutil.GeneratePDF(t, "18_invoice.pdf", doc)
}
