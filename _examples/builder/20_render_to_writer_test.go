package builder_test

import (
	"bytes"
	"testing"

	"github.com/gpdf-dev/gpdf/_examples/testutil"
	"github.com/gpdf-dev/gpdf/document"
	"github.com/gpdf-dev/gpdf/template"
)

func TestExample_20_RenderToWriter(t *testing.T) {
	doc := template.New(
		template.WithPageSize(document.A4),
		template.WithMargins(document.UniformEdges(document.Mm(20))),
	)

	page := doc.AddPage()
	page.AutoRow(func(r *template.RowBuilder) {
		r.Col(12, func(c *template.ColBuilder) {
			c.Text("Render to io.Writer", template.FontSize(18), template.Bold())
			c.Spacer(document.Mm(5))
			c.Text("This PDF was rendered using doc.Render(w) instead of doc.Generate().")
		})
	})

	// Use Render instead of Generate
	var buf bytes.Buffer
	if err := doc.Render(&buf); err != nil {
		t.Fatalf("Render failed: %v", err)
	}

	data := buf.Bytes()
	testutil.AssertValidPDF(t, data)
	testutil.WritePDF(t, "20_render_to_writer.pdf", data)
	testutil.AssertMatchesSharedGolden(t, "20_render_to_writer.pdf", data)
}
