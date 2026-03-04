package builder_test

import (
	"testing"

	"github.com/gpdf-dev/gpdf/_examples/testutil"
	"github.com/gpdf-dev/gpdf/document"
	"github.com/gpdf-dev/gpdf/pdf"
	"github.com/gpdf-dev/gpdf/template"
)

func TestExample_31_AbsolutePosition(t *testing.T) {
	doc := template.New(
		template.WithPageSize(document.A4),
		template.WithMargins(document.UniformEdges(document.Mm(20))),
		template.WithMetadata(document.DocumentMetadata{
			Title:  "Absolute Positioning",
			Author: "gpdf",
		}),
	)

	page := doc.AddPage()

	// Title
	page.AutoRow(func(r *template.RowBuilder) {
		r.Col(12, func(c *template.ColBuilder) {
			c.Text("Absolute Positioning", template.FontSize(18), template.Bold())
		})
	})

	// Description
	page.AutoRow(func(r *template.RowBuilder) {
		r.Col(12, func(c *template.ColBuilder) {
			c.Text("This is normal flow content. The absolute elements below are overlaid on top.")
		})
	})

	// Spacer
	page.AutoRow(func(r *template.RowBuilder) {
		r.Col(12, func(c *template.ColBuilder) {
			c.Spacer(document.Mm(5))
		})
	})

	// Table
	page.AutoRow(func(r *template.RowBuilder) {
		r.Col(12, func(c *template.ColBuilder) {
			c.Table(
				[]string{"Product", "Quantity", "Unit Price"},
				[][]string{
					{"Alpha", "10", "$5.00"},
					{"Beta", "5", "$12.00"},
					{"Gamma", "20", "$3.50"},
				},
			)
		})
	})

	// Absolute: CONFIDENTIAL stamp (content area origin)
	page.Absolute(document.Mm(130), document.Mm(20), func(c *template.ColBuilder) {
		c.Text("CONFIDENTIAL",
			template.FontSize(12),
			template.Bold(),
			template.TextColor(pdf.RGBHex(0xCC3333)),
		)
	}, template.AbsoluteWidth(document.Mm(40)))

	// Absolute: DRAFT watermark (page origin)
	page.Absolute(document.Mm(30), document.Mm(150), func(c *template.ColBuilder) {
		c.Text("DRAFT",
			template.FontSize(60),
			template.TextColor(pdf.RGBHex(0xE5E5E5)),
		)
	}, template.AbsoluteOriginPage())

	// Absolute: QR code
	page.Absolute(document.Mm(0), document.Mm(230), func(c *template.ColBuilder) {
		c.QRCode("https://gpdf.dev", template.QRSize(document.Mm(20)))
	})

	testutil.GeneratePDFSharedGolden(t, "31_absolute_position.pdf", doc)
}
