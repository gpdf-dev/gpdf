package gotemplate_test

import (
	"testing"

	"github.com/gpdf-dev/gpdf/_examples/testutil"
	"github.com/gpdf-dev/gpdf/template"
)

func TestTmpl_22_LetterSpacing(t *testing.T) {
	schema := []byte(`{
		"page": {"size": "A4", "margins": "20mm"},
		"body": [
			{"row": {"cols": [
				{"span": 12, "elements": [
					{"type": "text", "content": "{{.Title}}", "style": {"size": 20, "bold": true}},
					{"type": "spacer", "height": "8mm"},
					{"type": "text", "content": "{{.Normal}}"},
					{"type": "spacer", "height": "3mm"},
					{"type": "text", "content": "{{.Spacing1}}", "style": {"letterSpacing": 1}},
					{"type": "spacer", "height": "3mm"},
					{"type": "text", "content": "{{.Spacing3}}", "style": {"letterSpacing": 3}},
					{"type": "spacer", "height": "3mm"},
					{"type": "text", "content": "{{.WideHeader}}", "style": {"size": 16, "bold": true, "letterSpacing": 5}},
					{"type": "spacer", "height": "3mm"},
					{"type": "text", "content": "{{.Tight}}", "style": {"letterSpacing": -0.5}}
				]}
			]}}
		]
	}`)

	data := map[string]any{
		"Title":      "Letter Spacing Demo",
		"Normal":     "Normal spacing (0pt)",
		"Spacing1":   "Letter spacing 1pt",
		"Spacing3":   "Letter spacing 3pt",
		"WideHeader": "WIDE HEADER",
		"Tight":      "Tight spacing -0.5pt",
	}

	doc, err := template.FromJSON(schema, data)
	if err != nil {
		t.Fatalf("FromJSON error: %v", err)
	}
	testutil.GeneratePDFSharedGolden(t, "22_letter_spacing.pdf", doc)
}
