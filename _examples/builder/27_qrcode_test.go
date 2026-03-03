package builder_test

import (
	"testing"

	"github.com/gpdf-dev/gpdf/_examples/testutil"
	"github.com/gpdf-dev/gpdf/document"
	"github.com/gpdf-dev/gpdf/pdf"
	"github.com/gpdf-dev/gpdf/qrcode"
	"github.com/gpdf-dev/gpdf/template"
)

func TestExample_27_QRCode(t *testing.T) {
	doc := template.New(
		template.WithPageSize(document.A4),
		template.WithMargins(document.UniformEdges(document.Mm(20))),
		template.WithMetadata(document.DocumentMetadata{
			Title:  "QR Code Examples",
			Author: "gpdf",
		}),
	)

	page := doc.AddPage()

	page.AutoRow(func(r *template.RowBuilder) {
		r.Col(12, func(c *template.ColBuilder) {
			c.Text("QR Code Examples", template.FontSize(18), template.Bold())
			c.Spacer(document.Mm(5))
		})
	})

	// Basic QR code
	page.AutoRow(func(r *template.RowBuilder) {
		r.Col(12, func(c *template.ColBuilder) {
			c.Text("Basic QR code (URL):")
			c.Spacer(document.Mm(2))
			c.QRCode("https://gpdf.dev")
			c.Spacer(document.Mm(5))
		})
	})

	// QR codes with different sizes side by side
	page.AutoRow(func(r *template.RowBuilder) {
		r.Col(12, func(c *template.ColBuilder) {
			c.Text("QR codes with different sizes:")
			c.Spacer(document.Mm(2))
		})
	})

	page.AutoRow(func(r *template.RowBuilder) {
		r.Col(4, func(c *template.ColBuilder) {
			c.Text("20mm", template.FontSize(9), template.TextColor(pdf.Gray(0.4)))
			c.QRCode("https://gpdf.dev", template.QRSize(document.Mm(20)))
		})
		r.Col(4, func(c *template.ColBuilder) {
			c.Text("30mm", template.FontSize(9), template.TextColor(pdf.Gray(0.4)))
			c.QRCode("https://gpdf.dev", template.QRSize(document.Mm(30)))
		})
		r.Col(4, func(c *template.ColBuilder) {
			c.Text("40mm", template.FontSize(9), template.TextColor(pdf.Gray(0.4)))
			c.QRCode("https://gpdf.dev", template.QRSize(document.Mm(40)))
		})
	})

	page.AutoRow(func(r *template.RowBuilder) {
		r.Col(12, func(c *template.ColBuilder) {
			c.Spacer(document.Mm(5))
		})
	})

	// QR codes with different EC levels
	page.AutoRow(func(r *template.RowBuilder) {
		r.Col(12, func(c *template.ColBuilder) {
			c.Text("Error correction levels (L / M / Q / H):")
			c.Spacer(document.Mm(2))
		})
	})

	page.AutoRow(func(r *template.RowBuilder) {
		r.Col(3, func(c *template.ColBuilder) {
			c.Text("Level L", template.FontSize(9))
			c.QRCode("HELLO", template.QRSize(document.Mm(25)),
				template.QRErrorCorrection(qrcode.LevelL))
		})
		r.Col(3, func(c *template.ColBuilder) {
			c.Text("Level M", template.FontSize(9))
			c.QRCode("HELLO", template.QRSize(document.Mm(25)),
				template.QRErrorCorrection(qrcode.LevelM))
		})
		r.Col(3, func(c *template.ColBuilder) {
			c.Text("Level Q", template.FontSize(9))
			c.QRCode("HELLO", template.QRSize(document.Mm(25)),
				template.QRErrorCorrection(qrcode.LevelQ))
		})
		r.Col(3, func(c *template.ColBuilder) {
			c.Text("Level H", template.FontSize(9))
			c.QRCode("HELLO", template.QRSize(document.Mm(25)),
				template.QRErrorCorrection(qrcode.LevelH))
		})
	})

	page.AutoRow(func(r *template.RowBuilder) {
		r.Col(12, func(c *template.ColBuilder) {
			c.Spacer(document.Mm(5))
		})
	})

	// Japanese content
	page.AutoRow(func(r *template.RowBuilder) {
		r.Col(12, func(c *template.ColBuilder) {
			c.Text("Japanese content:")
			c.Spacer(document.Mm(2))
			c.QRCode("こんにちは世界", template.QRSize(document.Mm(30)))
		})
	})

	testutil.GeneratePDFSharedGolden(t, "27_qrcode.pdf", doc)
}
