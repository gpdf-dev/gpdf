package builder_test

import (
	"testing"

	"github.com/gpdf-dev/gpdf/_examples/testutil"
	"github.com/gpdf-dev/gpdf/document"
	"github.com/gpdf-dev/gpdf/template"
)

func TestExample_23_TextIndent(t *testing.T) {
	doc := template.New(
		template.WithPageSize(document.A4),
		template.WithMargins(document.UniformEdges(document.Mm(20))),
	)

	page := doc.AddPage()
	page.AutoRow(func(r *template.RowBuilder) {
		r.Col(12, func(c *template.ColBuilder) {
			c.Text("Text Indent Demo", template.FontSize(20), template.Bold())
			c.Spacer(document.Mm(8))

			c.Text("This paragraph has a 24pt first-line indent. "+
				"The first line starts further to the right, while subsequent "+
				"lines wrap at the normal left margin. This is commonly used "+
				"in book typography to indicate new paragraphs.",
				template.TextIndent(document.Pt(24)))
			c.Spacer(document.Mm(5))

			c.Text("This paragraph uses a larger 48pt indent for a more dramatic "+
				"effect. The indentation makes it easy to distinguish where a "+
				"new paragraph begins without adding extra vertical space.",
				template.TextIndent(document.Pt(48)))
			c.Spacer(document.Mm(5))

			c.Text("No indent on this paragraph for comparison. " +
				"Standard left-aligned text without any first-line indentation " +
				"starts flush with the left margin.")
		})
	})

	testutil.GeneratePDFSharedGolden(t, "23_text_indent.pdf", doc)
}
