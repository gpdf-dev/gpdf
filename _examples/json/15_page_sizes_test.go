package json_test

import (
	"fmt"
	"testing"

	"github.com/gpdf-dev/gpdf/_examples/testutil"
	"github.com/gpdf-dev/gpdf/template"
)

func TestJSON_15_PageSizes(t *testing.T) {
	sizes := []struct {
		name string
		size string
		desc string
		file string
	}{
		{"A4 (210mm x 297mm)", "A4", "A4 (210mm x 297mm)", "15a_pagesize_a4.pdf"},
		{"A3 (297mm x 420mm)", "A3", "A3 (297mm x 420mm)", "15b_pagesize_a3.pdf"},
		{"Letter (8.5in x 11in)", "Letter", "Letter (8.5in x 11in)", "15c_pagesize_letter.pdf"},
		{"Legal (8.5in x 14in)", "Legal", "Legal (8.5in x 14in)", "15d_pagesize_legal.pdf"},
	}
	for _, s := range sizes {
		t.Run(s.name, func(t *testing.T) {
			schema := []byte(fmt.Sprintf(`{
				"page": {"size": "%s", "margins": "20mm"},
				"body": [
					{"row": {"cols": [
						{"span": 12, "elements": [
							{"type": "text", "content": "Page Size: %s", "style": {"size": 20, "bold": true}},
							{"type": "spacer", "height": "10mm"},
							{"type": "text", "content": "This page demonstrates the %s page format."}
						]}
					]}}
				]
			}`, s.size, s.desc, s.desc))
			doc, err := template.FromJSON(schema, nil)
			if err != nil {
				t.Fatalf("FromJSON error: %v", err)
			}
			testutil.GeneratePDFSharedGolden(t, s.file, doc)
		})
	}
}
