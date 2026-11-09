package builder_test

import (
	"testing"

	"github.com/gpdf-dev/gpdf/_examples/testutil"
	"github.com/gpdf-dev/gpdf/document"
	"github.com/gpdf-dev/gpdf/pdf"
	"github.com/gpdf-dev/gpdf/template"
)

func TestExample_41_Box(t *testing.T) {
	doc := template.New(
		template.WithPageSize(document.A4),
		template.WithMargins(document.UniformEdges(document.Mm(20))),
	)

	border := template.Border(
		template.BorderWidth(document.Pt(1)),
		template.BorderColor(pdf.RGBHex(0x1A237E)),
	)

	page := doc.AddPage()
	page.AutoRow(func(r *template.RowBuilder) {
		r.Col(12, func(c *template.ColBuilder) {
			c.Text("Box Demo", template.FontSize(20), template.Bold())
			c.Spacer(document.Mm(6))
		})
	})

	page.AutoRow(func(r *template.RowBuilder) {
		r.Col(3, func(c *template.ColBuilder) {
			c.Box(func(c *template.ColBuilder) {
				c.Text("LOGO", template.Bold(), template.AlignCenter())
			},
				template.WithBoxBackground(pdf.RGBHex(0xE8EAF6)),
				template.WithBoxHeight(document.Mm(20)),
				template.WithBoxPadding(document.UniformEdges(document.Mm(3))),
			)
		})
		r.Col(5, func(c *template.ColBuilder) {
			c.Box(func(c *template.ColBuilder) {
				c.Text("line 1")
				c.Text("line 2")
				c.Text("line 3")
			}, template.WithBoxPadding(document.UniformEdges(document.Mm(2))))
		})
		r.Col(4, func(c *template.ColBuilder) {
			c.Box(func(c *template.ColBuilder) {
				c.Text("line 1", template.AlignRight())
				c.Text("line 2", template.AlignRight())
				c.Text("line 3", template.AlignRight())
			},
				template.WithBoxBorder(border),
				template.WithBoxPadding(document.Edges{
					Top: document.Mm(2), Right: document.Mm(3),
					Bottom: document.Mm(2), Left: document.Mm(3),
				}),
			)
		})
	})

	page.AutoRow(func(r *template.RowBuilder) {
		r.Col(12, func(c *template.ColBuilder) {
			c.Spacer(document.Mm(6))
			c.Box(func(c *template.ColBuilder) {
				c.Text("Outer box", template.Bold())
				c.Box(func(c *template.ColBuilder) {
					c.Text("Nested box with its own background")
				},
					template.WithBoxBackground(pdf.RGBHex(0xFFF3E0)),
					template.WithBoxPadding(document.UniformEdges(document.Mm(3))),
				)
			},
				template.WithBoxBorder(border),
				template.WithBoxPadding(document.UniformEdges(document.Mm(4))),
			)
		})
	})

	testutil.GeneratePDFSharedGolden(t, "41_box.pdf", doc)
}
