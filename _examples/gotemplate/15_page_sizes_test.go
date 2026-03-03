package gotemplate_test

import (
	"testing"

	"github.com/gpdf-dev/gpdf/_examples/testutil"
	"github.com/gpdf-dev/gpdf/template"
)

func TestTmpl_15_PageSizes(t *testing.T) {
	sizes := []struct {
		name string
		size string
		file string
	}{
		{"A4 (210mm x 297mm)", "A4", "15a_pagesize_a4.pdf"},
		{"A3 (297mm x 420mm)", "A3", "15b_pagesize_a3.pdf"},
		{"Letter (8.5in x 11in)", "Letter", "15c_pagesize_letter.pdf"},
		{"Legal (8.5in x 14in)", "Legal", "15d_pagesize_legal.pdf"},
	}
	for _, s := range sizes {
		t.Run(s.name, func(t *testing.T) {
			schema := []byte(`{
				"page": {"size": "` + s.size + `", "margins": "20mm"},
				"body": [
					{"row": {"cols": [
						{"span": 12, "elements": [
							{"type": "text", "content": "{{.SizeTitle}}", "style": {"size": 20, "bold": true}},
							{"type": "spacer", "height": "10mm"},
							{"type": "text", "content": "{{.Description}}"}
						]}
					]}}
				]
			}`)
			data := map[string]any{
				"SizeTitle":   "Page Size: " + s.name,
				"Description": "This page demonstrates the " + s.name + " page format.",
			}
			doc, err := template.FromJSON(schema, data)
			if err != nil {
				t.Fatalf("FromJSON error: %v", err)
			}
			testutil.GeneratePDFSharedGolden(t, s.file, doc)
		})
	}
}
