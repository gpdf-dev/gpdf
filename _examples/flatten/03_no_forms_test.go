package flatten_test

import (
	"testing"

	gpdf "github.com/gpdf-dev/gpdf"
	"github.com/gpdf-dev/gpdf/_examples/testutil"
	"github.com/gpdf-dev/gpdf/document"
	"github.com/gpdf-dev/gpdf/template"
)

func TestExample_Flatten_03_NoForms(t *testing.T) {
	// Flattening a PDF with no forms should be a safe no-op.
	doc := template.New(
		template.WithPageSize(document.A4),
		template.WithMargins(document.UniformEdges(document.Mm(20))),
	)
	page := doc.AddPage()
	page.AutoRow(func(r *template.RowBuilder) {
		r.Col(12, func(c *template.ColBuilder) {
			c.Text("This PDF has no forms.", template.FontSize(16), template.Bold())
		})
	})

	source, err := doc.Generate()
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}

	existing, err := gpdf.Open(source)
	if err != nil {
		t.Fatalf("gpdf.Open: %v", err)
	}

	// Should succeed without error (no-op).
	if err := existing.FlattenForms(); err != nil {
		t.Fatalf("FlattenForms: %v", err)
	}

	result, err := existing.Save()
	if err != nil {
		t.Fatalf("Save: %v", err)
	}

	testutil.AssertValidPDF(t, result)
	testutil.WritePDF(t, "03_no_forms.pdf", result)
}
