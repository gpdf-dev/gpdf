package builder_test

import (
	"testing"

	"github.com/gpdf-dev/gpdf/document"
	"github.com/gpdf-dev/gpdf/pdf"
	"github.com/gpdf-dev/gpdf/_examples/testutil"
	"github.com/gpdf-dev/gpdf/template"
)

func TestExample_02_TextStyling(t *testing.T) {
	doc := template.New(
		template.WithPageSize(document.A4),
		template.WithMargins(document.UniformEdges(document.Mm(20))),
	)

	page := doc.AddPage()

	// Title
	page.AutoRow(func(r *template.RowBuilder) {
		r.Col(12, func(c *template.ColBuilder) {
			c.Text("Text Styling Examples", template.FontSize(20), template.Bold())
			c.Spacer(document.Mm(5))
		})
	})

	// Font sizes
	page.AutoRow(func(r *template.RowBuilder) {
		r.Col(12, func(c *template.ColBuilder) {
			c.Text("Font Size 8pt", template.FontSize(8))
			c.Text("Font Size 12pt (default)", template.FontSize(12))
			c.Text("Font Size 18pt", template.FontSize(18))
			c.Text("Font Size 24pt", template.FontSize(24))
			c.Text("Font Size 36pt", template.FontSize(36))
			c.Spacer(document.Mm(5))
		})
	})

	// Font weight and style
	page.AutoRow(func(r *template.RowBuilder) {
		r.Col(12, func(c *template.ColBuilder) {
			c.Text("Normal text")
			c.Text("Bold text", template.Bold())
			c.Text("Italic text", template.Italic())
			c.Text("Bold + Italic text", template.Bold(), template.Italic())
			c.Spacer(document.Mm(5))
		})
	})

	// Text colors
	page.AutoRow(func(r *template.RowBuilder) {
		r.Col(12, func(c *template.ColBuilder) {
			c.Text("Red text", template.TextColor(pdf.Red))
			c.Text("Green text", template.TextColor(pdf.Green))
			c.Text("Blue text", template.TextColor(pdf.Blue))
			c.Text("Custom color (orange)", template.TextColor(pdf.RGB(1.0, 0.5, 0.0)))
			c.Text("Hex color (#336699)", template.TextColor(pdf.RGBHex(0x336699)))
			c.Spacer(document.Mm(5))
		})
	})

	// Background colors
	page.AutoRow(func(r *template.RowBuilder) {
		r.Col(12, func(c *template.ColBuilder) {
			c.Text("Yellow background", template.BgColor(pdf.Yellow))
			c.Text("Cyan background", template.BgColor(pdf.Cyan))
			c.Text("White text on dark background",
				template.TextColor(pdf.White),
				template.BgColor(pdf.RGBHex(0x333333)),
			)
			c.Spacer(document.Mm(5))
		})
	})

	// Text alignment
	page.AutoRow(func(r *template.RowBuilder) {
		r.Col(12, func(c *template.ColBuilder) {
			c.Text("Left aligned (default)", template.AlignLeft())
			c.Text("Center aligned", template.AlignCenter())
			c.Text("Right aligned", template.AlignRight())
		})
	})

	testutil.GeneratePDFSharedGolden(t, "02_text_styling.pdf", doc)
}
