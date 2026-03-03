package builder_test

import (
	"testing"

	"github.com/gpdf-dev/gpdf/_examples/testutil"
	"github.com/gpdf-dev/gpdf/document"
	"github.com/gpdf-dev/gpdf/pdf"
	"github.com/gpdf-dev/gpdf/qrcode"
	"github.com/gpdf-dev/gpdf/template"
)

func TestExample_29_QRBarcode_Invoice(t *testing.T) {
	doc := template.New(
		template.WithPageSize(document.A4),
		template.WithMargins(document.UniformEdges(document.Mm(20))),
		template.WithMetadata(document.DocumentMetadata{
			Title:  "Invoice with QR/Barcode",
			Author: "ACME Corporation",
		}),
	)

	page := doc.AddPage()

	// Header
	page.AutoRow(func(r *template.RowBuilder) {
		r.Col(8, func(c *template.ColBuilder) {
			c.Text("ACME Corporation", template.FontSize(22), template.Bold(),
				template.TextColor(pdf.RGBHex(0x1A237E)))
			c.Text("Invoice #INV-2026-042", template.FontSize(12),
				template.TextColor(pdf.Gray(0.4)))
		})
		r.Col(4, func(c *template.ColBuilder) {
			c.QRCode("https://pay.acme.com/inv/2026-042",
				template.QRSize(document.Mm(30)),
				template.QRErrorCorrection(qrcode.LevelH))
		})
	})

	page.AutoRow(func(r *template.RowBuilder) {
		r.Col(12, func(c *template.ColBuilder) {
			c.Spacer(document.Mm(3))
			c.Line(template.LineColor(pdf.RGBHex(0x1A237E)),
				template.LineThickness(document.Pt(2)))
			c.Spacer(document.Mm(5))
		})
	})

	// Bill to
	page.AutoRow(func(r *template.RowBuilder) {
		r.Col(6, func(c *template.ColBuilder) {
			c.Text("Bill To:", template.Bold(), template.TextColor(pdf.Gray(0.4)))
			c.Spacer(document.Mm(2))
			c.Text("Jane Doe", template.Bold())
			c.Text("Tech Solutions Inc.")
			c.Text("456 Client Avenue")
			c.Text("New York, NY 10001")
		})
		r.Col(6, func(c *template.ColBuilder) {
			c.Text("Date: March 1, 2026", template.AlignRight())
			c.Text("Due: March 31, 2026", template.AlignRight())
		})
	})

	page.AutoRow(func(r *template.RowBuilder) {
		r.Col(12, func(c *template.ColBuilder) {
			c.Spacer(document.Mm(8))
		})
	})

	// Items table
	page.AutoRow(func(r *template.RowBuilder) {
		r.Col(12, func(c *template.ColBuilder) {
			c.Table(
				[]string{"Description", "Qty", "Price", "Amount"},
				[][]string{
					{"Web Development", "40 hrs", "$150.00", "$6,000.00"},
					{"UI/UX Design", "20 hrs", "$120.00", "$2,400.00"},
					{"QA Testing", "15 hrs", "$100.00", "$1,500.00"},
				},
				template.ColumnWidths(40, 15, 20, 25),
				template.TableHeaderStyle(
					template.TextColor(pdf.White),
					template.BgColor(pdf.RGBHex(0x1A237E)),
				),
				template.TableStripe(pdf.RGBHex(0xF5F5F5)),
			)
		})
	})

	page.AutoRow(func(r *template.RowBuilder) {
		r.Col(8, func(c *template.ColBuilder) {})
		r.Col(4, func(c *template.ColBuilder) {
			c.Spacer(document.Mm(3))
			c.Text("Total:  $9,900.00", template.AlignRight(),
				template.Bold(), template.FontSize(14))
		})
	})

	page.AutoRow(func(r *template.RowBuilder) {
		r.Col(12, func(c *template.ColBuilder) {
			c.Spacer(document.Mm(10))
			c.Line(template.LineColor(pdf.Gray(0.8)))
			c.Spacer(document.Mm(5))
		})
	})

	// Barcode at bottom for scanning
	page.AutoRow(func(r *template.RowBuilder) {
		r.Col(12, func(c *template.ColBuilder) {
			c.Text("Order Reference:", template.FontSize(9),
				template.TextColor(pdf.Gray(0.4)))
			c.Spacer(document.Mm(2))
			c.Barcode("INV-2026-042", template.BarcodeWidth(document.Mm(100)),
				template.BarcodeHeight(document.Mm(15)))
			c.Spacer(document.Mm(5))
			c.Text("Scan QR code to pay online", template.AlignCenter(),
				template.Italic(), template.FontSize(9),
				template.TextColor(pdf.Gray(0.5)))
		})
	})

	testutil.GeneratePDFSharedGolden(t, "29_qr_barcode_invoice.pdf", doc)
}
