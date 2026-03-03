package builder_test

import (
	"testing"

	"github.com/gpdf-dev/gpdf/document"
	"github.com/gpdf-dev/gpdf/_examples/testutil"
	"github.com/gpdf-dev/gpdf/template"
)

func TestExample_22_LetterSpacing(t *testing.T) {
	doc := template.New(
		template.WithPageSize(document.A4),
		template.WithMargins(document.UniformEdges(document.Mm(20))),
	)

	page := doc.AddPage()
	page.AutoRow(func(r *template.RowBuilder) {
		r.Col(12, func(c *template.ColBuilder) {
			c.Text("Letter Spacing Demo", template.FontSize(20), template.Bold())
			c.Spacer(document.Mm(8))

			c.Text("Normal spacing (0pt)")
			c.Spacer(document.Mm(3))

			c.Text("Letter spacing 1pt", template.LetterSpacing(1))
			c.Spacer(document.Mm(3))

			c.Text("Letter spacing 3pt", template.LetterSpacing(3))
			c.Spacer(document.Mm(3))

			c.Text("WIDE HEADER", template.FontSize(16), template.Bold(),
				template.LetterSpacing(5))
			c.Spacer(document.Mm(3))

			c.Text("Tight spacing -0.5pt", template.LetterSpacing(-0.5))
		})
	})

	testutil.GeneratePDF(t, "22_letter_spacing.pdf", doc)
}
