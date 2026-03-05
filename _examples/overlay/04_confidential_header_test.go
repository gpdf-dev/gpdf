package overlay_test

import (
	"testing"

	gpdf "github.com/gpdf-dev/gpdf"
	"github.com/gpdf-dev/gpdf/_examples/testutil"
	"github.com/gpdf-dev/gpdf/document"
	"github.com/gpdf-dev/gpdf/pdf"
	"github.com/gpdf-dev/gpdf/template"
)

func TestExample_Overlay_04_ConfidentialHeader(t *testing.T) {
	source := generateSourcePDF(t, 3)

	doc, err := gpdf.Open(source)
	if err != nil {
		t.Fatalf("gpdf.Open: %v", err)
	}

	// Add a red "CONFIDENTIAL" banner on every page.
	err = doc.EachPage(func(pageIndex int, p *template.PageBuilder) {
		// Red background bar at the top (page origin).
		p.Absolute(document.Mm(0), document.Mm(0), func(c *template.ColBuilder) {
			c.Line(
				template.LineColor(pdf.RGB(0.8, 0, 0)),
				template.LineThickness(document.Mm(8)),
			)
		}, template.AbsoluteOriginPage(), template.AbsoluteWidth(document.Mm(210)))

		// White text on the red bar.
		p.Absolute(document.Mm(70), document.Mm(1), func(c *template.ColBuilder) {
			c.Text("CONFIDENTIAL",
				template.FontSize(14),
				template.Bold(),
				template.TextColor(pdf.RGB(1, 1, 1)),
			)
		}, template.AbsoluteOriginPage())
	})
	if err != nil {
		t.Fatalf("EachPage: %v", err)
	}

	result, err := doc.Save()
	if err != nil {
		t.Fatalf("Save: %v", err)
	}

	testutil.AssertValidPDF(t, result)
	testutil.WritePDF(t, "04_confidential_header.pdf", result)
}
