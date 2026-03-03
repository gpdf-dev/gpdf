package json_test

import (
	"testing"

	"github.com/gpdf-dev/gpdf/_examples/testutil"
	"github.com/gpdf-dev/gpdf/template"
)

func TestJSON_22_LetterSpacing(t *testing.T) {
	schema := []byte(`{
		"page": {"size": "A4", "margins": "20mm"},
		"body": [
			{"row": {"cols": [
				{"span": 12, "elements": [
					{"type": "text", "content": "Letter Spacing Demo", "style": {"size": 20, "bold": true}},
					{"type": "spacer", "height": "8mm"},
					{"type": "text", "content": "Normal spacing (0pt)"},
					{"type": "spacer", "height": "3mm"},
					{"type": "text", "content": "Letter spacing 1pt", "style": {"letterSpacing": 1}},
					{"type": "spacer", "height": "3mm"},
					{"type": "text", "content": "Letter spacing 3pt", "style": {"letterSpacing": 3}},
					{"type": "spacer", "height": "3mm"},
					{"type": "text", "content": "WIDE HEADER", "style": {"size": 16, "bold": true, "letterSpacing": 5}},
					{"type": "spacer", "height": "3mm"},
					{"type": "text", "content": "Tight spacing -0.5pt", "style": {"letterSpacing": -0.5}}
				]}
			]}}
		]
	}`)

	doc, err := template.FromJSON(schema, nil)
	if err != nil {
		t.Fatalf("FromJSON error: %v", err)
	}
	testutil.GeneratePDFSharedGolden(t, "22_letter_spacing.pdf", doc)
}
