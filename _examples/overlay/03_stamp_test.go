package overlay_test

import (
	"testing"

	gpdf "github.com/gpdf-dev/gpdf"
	"github.com/gpdf-dev/gpdf/_examples/testutil"
	"github.com/gpdf-dev/gpdf/document"
	"github.com/gpdf-dev/gpdf/pdf"
	"github.com/gpdf-dev/gpdf/template"
)

func TestExample_Overlay_03_Stamp(t *testing.T) {
	source := generateSourcePDF(t, 1)

	doc, err := gpdf.Open(source)
	if err != nil {
		t.Fatalf("gpdf.Open: %v", err)
	}

	// Add an "APPROVED" stamp in red at the top-right corner.
	err = doc.Overlay(0, func(p *template.PageBuilder) {
		// Red stamp text.
		p.Absolute(document.Mm(130), document.Mm(15), func(c *template.ColBuilder) {
			c.Text("APPROVED",
				template.FontSize(28),
				template.Bold(),
				template.TextColor(pdf.RGB(0.8, 0, 0)),
			)
		})

		// Date below the stamp.
		p.Absolute(document.Mm(130), document.Mm(25), func(c *template.ColBuilder) {
			c.Text("2026-03-05",
				template.FontSize(10),
				template.TextColor(pdf.RGB(0.5, 0.5, 0.5)),
			)
		})
	})
	if err != nil {
		t.Fatalf("Overlay: %v", err)
	}

	result, err := doc.Save()
	if err != nil {
		t.Fatalf("Save: %v", err)
	}

	testutil.AssertValidPDF(t, result)
	testutil.WritePDF(t, "03_stamp.pdf", result)
}
